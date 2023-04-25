package cmd

import (
	"fmt"
	"os"

	"github.com/nhost/cli/v2/controller"
	"github.com/nhost/cli/v2/system"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	return &cobra.Command{ //nolint:exhaustruct
		Use:   "init",
		Short: "Initialize current directory as Nhost app",
		Long: `Initialize current working directory as an Nhost application.

Without specifying --remote flag, only a blank Nhost app will be initialized.

Specifying --remote flag will initialize a local app from app.nhost.io
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if system.PathExists(system.PathNhost()) {
				return fmt.Errorf("nhost folder already exists in this directory") //nolint:goerr113
			}

			if err := os.MkdirAll(system.PathNhost(), 0o755); err != nil { //nolint:gomnd
				return fmt.Errorf("failed to create nhost folder: %w", err)
			}

			if err := os.MkdirAll(system.PathDotNhost(), 0o755); err != nil { //nolint:gomnd
				return fmt.Errorf("failed to create .nhost folder: %w", err)
			}

			ctrl := controller.New(cmd, nil, GetNhostCredentials)

			tomlf, err := system.GetConfigFile()
			if err != nil {
				return fmt.Errorf("failed to get config app file: %w", err)
			}
			defer tomlf.Close()

			secretsf, err := system.GetSecretsFile()
			if err != nil {
				return fmt.Errorf("failed to get config app file: %w", err)
			}
			defer secretsf.Close()

			gitignoref, err := system.GetGitignoreFile()
			if err != nil {
				return fmt.Errorf("failed to get config app file: %w", err)
			}
			defer gitignoref.Close()

			hasuraConfigF, err := system.GetHasuraConfigFile()
			if err != nil {
				return fmt.Errorf("failed to get config app file: %w", err)
			}
			defer hasuraConfigF.Close()

			return ctrl.Init(cmd.Context(), tomlf, secretsf, gitignoref, hasuraConfigF) //nolint:wrapcheck
		},
	}
}
