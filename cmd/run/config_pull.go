package run

import (
	"encoding/json"
	"fmt"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/nhostclient/graphql"
	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v2"
)

func CommandConfigPull() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:    "config-pull",
		Aliases: []string{},
		Usage:   "Download service configuration",
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
		Action: commandConfigPull,
	}
}

func commandConfigPull(cCtx *cli.Context) error {
	ce := clienv.FromCLI(cCtx)
	proj, err := ce.GetAppInfo(cCtx.Context, cCtx.String(flagSubdomain))
	if err != nil {
		return fmt.Errorf("failed to get app info: %w", err)
	}

	session, err := ce.LoadSession(cCtx.Context)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	cl := ce.GetNhostClient()

	resp, err := cl.GetRunServiceConfigRawJSON(
		cCtx.Context,
		proj.ID,
		cCtx.String(flagServiceID),
		false,
		graphql.WithAccessToken(session.Session.AccessToken),
	)
	if err != nil {
		return fmt.Errorf("failed to get service config: %w", err)
	}

	var v model.ConfigRunServiceConfig
	if err := json.Unmarshal([]byte(resp.RunServiceConfigRawJSON), &v); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := clienv.MarshalFile(v, cCtx.String(flagConfig), toml.Marshal); err != nil {
		return fmt.Errorf("failed to save config to file: %w", err)
	}

	return nil
}
