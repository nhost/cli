package cmd

import (
	"fmt"

	"github.com/nhost/cli/v2/controller"
	"github.com/nhost/cli/v2/nhostclient"
	"github.com/nhost/cli/v2/system"
	"github.com/spf13/cobra"
)

func secretsCreateCmd() *cobra.Command {
	return &cobra.Command{ //nolint:exhaustruct
		Use:   "create SECRET_NAME SECRET_VALUE",
		Short: "Create a new secret",
		Args:  cobra.ExactArgs(2), //nolint:gomnd
		RunE: func(cmd *cobra.Command, args []string) error {
			projsf, err := system.GetNhostProjectInfoFile()
			if err != nil {
				return fmt.Errorf("failed to get project's file: %w", err)
			}
			defer projsf.Close()

			cl := nhostclient.New(cmd.Flag(flagDomain).Value.String())
			ctrl := controller.New(cmd, cl, GetNhostCredentials)

			return ctrl.SecretsCreate(cmd.Context(), projsf, args[0], args[1]) //nolint:wrapcheck
		},
	}
}
