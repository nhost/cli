package run

import (
	"encoding/json"
	"fmt"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/nhostclient/graphql"
	"github.com/urfave/cli/v2"
)

func CommandConfigReplace() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:    "config-replace",
		Aliases: []string{},
		Usage:   "Replace service configuration",
		Action:  commandConfigReplace,
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
			&cli.StringFlag{ //nolint:exhaustruct
				Name:     flagServiceID,
				Usage:    "Service ID to update",
				Required: true,
				EnvVars:  []string{"NHOST_RUN_SERVICE_ID"},
			},
		},
	}
}

func transform[T, V any](t *T) (*V, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}

	var v V
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return &v, nil
}

func commandConfigReplace(cCtx *cli.Context) error {
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

	if err := ValidateRemote(cCtx.Context, ce, session, cfg, proj.ID); err != nil {
		return err
	}

	cl := ce.GetNhostClient()
	replaceConfig, err := transform[model.ConfigRunServiceConfig, graphql.ConfigRunServiceConfigInsertInput](cfg)
	if err != nil {
		return fmt.Errorf("failed to transform configuration into replace input: %w", err)
	}

	if _, err := cl.ReplaceRunServiceConfig(
		cCtx.Context,
		proj.ID,
		cCtx.String(flagServiceID),
		*replaceConfig,
		graphql.WithAccessToken(session.Session.AccessToken),
	); err != nil {
		return fmt.Errorf("failed to replace service config: %w", err)
	}

	ce.Infoln("Service configuration replaced")

	return nil
}
