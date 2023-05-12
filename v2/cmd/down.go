package cmd

import (
	"fmt"

	"github.com/nhost/cli/v2/controller"
	"github.com/nhost/cli/v2/system"
	"github.com/spf13/cobra"
)

func downCmd() *cobra.Command {
	return &cobra.Command{ //nolint:exhaustruct
		Use:        "down",
		SuggestFor: []string{"list", "init"},
		Short:      "Stop local development environment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !system.PathExists(system.PathConfig()) {
				return fmt.Errorf("no nhost project found in current directory, please run `nhost init`") //nolint:goerr113
			}
			if !system.PathExists(system.PathSecrets()) {
				return fmt.Errorf("no secrets found in current directory, please run `nhost init`") //nolint:goerr113
			}

			projecName, err := cmd.Flags().GetString(flagProjectName)
			if err != nil {
				return fmt.Errorf("failed to parse project name: %w", err)
			}

			return controller.Down(cmd.Context(), cmd, projecName) //nolint:wrapcheck
		},
	}
}
