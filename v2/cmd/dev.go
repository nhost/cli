package cmd

import (
	"fmt"

	"github.com/nhost/cli/v2/controller"
	"github.com/nhost/cli/v2/system"
	"github.com/spf13/cobra"
)

func devCmd() *cobra.Command {
	return &cobra.Command{ //nolint:exhaustruct
		Use:        "dev",
		Aliases:    []string{"up"},
		SuggestFor: []string{"list", "init"},
		Short:      "Start local development environment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !system.PathExists(system.PathConfig()) {
				return fmt.Errorf("no nhost project found in current directory, please run `nhost init`") //nolint:goerr113
			}
			if !system.PathExists(system.PathSecrets()) {
				return fmt.Errorf("no secrets found in current directory, please run `nhost init`") //nolint:goerr113
			}
			httpPort, err := cmd.Flags().GetUint(flagHTTPPort)
			if err != nil {
				return fmt.Errorf("failed to parse https port: %w", err)
			}

			disableTLS, err := cmd.Flags().GetBool(flagDisableTLS)
			if err != nil {
				return fmt.Errorf("failed to parse use-tls: %w", err)
			}

			postgresPort, err := cmd.Flags().GetUint(flagPostgresPort)
			if err != nil {
				return fmt.Errorf("failed to parse postgres port: %w", err)
			}

			projecName, err := cmd.Flags().GetString(flagProjectName)
			if err != nil {
				return fmt.Errorf("failed to parse project name: %w", err)
			}

			_, dataFolder, nhostFolder, functionsFolder, err := getFolders(cmd)
			if err != nil {
				return err
			}

			return controller.Dev( //nolint:wrapcheck
				cmd.Context(),
				cmd,
				projecName,
				httpPort,
				!disableTLS,
				postgresPort,
				dataFolder,
				nhostFolder,
				functionsFolder,
			)
		},
	}
}
