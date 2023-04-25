package cmd

import (
	"fmt"

	"github.com/nhost/cli/v2/controller"
	"github.com/nhost/cli/v2/nhostclient"
	"github.com/nhost/cli/v2/system"
	"github.com/spf13/cobra"
)

func secretsListCmd() *cobra.Command {
	return &cobra.Command{ //nolint:exhaustruct
		Use:   "list",
		Short: "List all secrets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			projsf, err := system.GetNhostProjectInfoFile()
			if err != nil {
				return fmt.Errorf("failed to get project's file: %w", err)
			}
			defer projsf.Close()

			cl := nhostclient.New(cmd.Flag(flagDomain).Value.String())
			ctrl := controller.New(cmd, cl, GetNhostCredentials)

			return ctrl.SecretsList(cmd.Context(), projsf) //nolint:wrapcheck
		},
	}
}
