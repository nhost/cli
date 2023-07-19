package run

import (
	"context"
	"fmt"
	"os"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/be/services/mimir/schema"
	"github.com/nhost/be/services/mimir/schema/appconfig"
	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/nhostclient/credentials"
	"github.com/nhost/cli/nhostclient/graphql"
	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v2"
)

const (
	flagConfig    = "config"
	flagSubdomain = "subdomain"
)

func CommandConfigValidate() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:    "config-validate",
		Aliases: []string{},
		Usage:   "Validates service configuration after resolving secrets (only validation against cloud project supported)",
		Action:  commandConfigValidate,
		Flags: []cli.Flag{
			&cli.StringFlag{ //nolint:exhaustruct
				Name:     flagConfig,
				Aliases:  []string{},
				Usage:    "Service configuration file",
				Required: true,
				EnvVars:  []string{"NHOST_RUN_SERVICE_CONFIG"},
			},
			&cli.StringFlag{ //nolint:exhaustruct
				Name:     flagSubdomain,
				Usage:    "Validate this subdomain's configuration. Defaults to linked project",
				Required: true,
				EnvVars:  []string{"NHOST_SUBDOMAIN"},
			},
		},
	}
}

func respToSecrets(env []*graphql.GetSecrets_AppSecrets) model.Secrets {
	secrets := make(model.Secrets, len(env))
	for i, s := range env {
		secrets[i] = &model.ConfigEnvironmentVariable{
			Name:  s.Name,
			Value: s.Value,
		}
	}
	return secrets
}

func loadConfig(
	path string,
) (*model.ConfigRunServiceConfig, error) {
	cfg := &model.ConfigRunServiceConfig{} //nolint:exhaustruct

	r, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("failed to open file: %w", err)
	}
	defer r.Close()

	decoder := toml.NewDecoder(r)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

func ValidateRemote(
	ctx context.Context,
	ce *clienv.CliEnv,
	session credentials.Session,
	cfg *model.ConfigRunServiceConfig,
	appID string,
) error {
	schema, err := schema.New()
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	ce.Infoln("Getting secrets...")
	cl := ce.GetNhostClient()
	secretsResp, err := cl.GetSecrets(
		ctx,
		appID,
		graphql.WithAccessToken(session.Session.AccessToken),
	)
	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	secrets := respToSecrets(secretsResp.GetAppSecrets())
	_, err = appconfig.SecretsResolver(cfg, secrets, schema.FillRunServiceConfig)
	if err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	ce.Infoln("Config is valid!")

	return nil
}

func commandConfigValidate(cCtx *cli.Context) error {
	cfg, err := loadConfig(cCtx.String(flagConfig))
	if err != nil {
		return err
	}

	ce := clienv.FromCLI(cCtx)
	proj, err := ce.GetAppInfo(cCtx.Context, cCtx.String(flagSubdomain))
	if err != nil {
		return fmt.Errorf("failed to get app info: %w", err)
	}

	session, err := ce.LoadSession(cCtx.Context)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	return ValidateRemote(
		cCtx.Context,
		ce,
		session,
		cfg,
		proj.ID,
	)
}
