package run

import (
	"fmt"

	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/dockercompose"
	"github.com/urfave/cli/v2"
)

func CommandUp() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:    "up",
		Aliases: []string{},
		Usage:   "Starts the service locally",
		Action:  commandUp,
		Flags: []cli.Flag{
			&cli.StringFlag{ //nolint:exhaustruct
				Name:    flagConfig,
				Aliases: []string{},
				Usage:   "Service configuration file",
				Value:   "nhost-service.toml",
				EnvVars: []string{"NHOST_RUN_SERVICE_CONFIG"},
			},
		},
	}
}

func commandUp(cCtx *cli.Context) error {
	cfg, err := loadConfig(cCtx.String(flagConfig))
	if err != nil {
		return err
	}

	ce := clienv.FromCLI(cCtx)

	// TODO: overlays
	cfg, err = ValidateAndResolve(
		ce,
		cfg,
	)
	if err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	composeFile, err := dockercompose.ComposeFileForRunServiceStandalone(
		cfg, ce.ProjectName(), 443, true, ce.Path.DataFolder(), ce.Path.DotNhostFolder(),
	)

	dc := dockercompose.New(ce.Path.WorkingDir(), ce.Path.DockerCompose(), ce.ProjectName())
	if err := dc.WriteComposeFile(composeFile); err != nil {
		return fmt.Errorf("failed to write docker-compose.yaml: %w", err)
	}

	if err = dc.Start(cCtx.Context); err != nil {
		return fmt.Errorf("failed to start Nhost development environment: %w", err)
	}

	return err
}
