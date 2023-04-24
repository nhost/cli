package cmd

import (
	"fmt"

	"github.com/nhost/cli/v2/controller"
	"github.com/nhost/cli/v2/nhostclient"
	"github.com/nhost/cli/v2/system"
	"github.com/spf13/cobra"
)

func configValidateCmd() *cobra.Command {
	return &cobra.Command{ //nolint:exhaustruct
		Use:  "validate",
		Long: `Validate configuration`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			validateRemote, err := cmd.Flags().GetBool(flagRemote)
			if err != nil {
				return fmt.Errorf("failed to get local flag: %w", err)
			}

			cl := nhostclient.New(cmd.Flag(flagDomain).Value.String())
			ctrl := controller.New(cmd, cl, GetNhostCredentials)

			tomlf, err := system.GetConfigFile()
			if err != nil {
				return fmt.Errorf("failed to get config app file: %w", err)
			}
			defer tomlf.Close()

			if validateRemote {
				projsf, err := system.GetNhostProjectInfoFile()
				if err != nil {
					return fmt.Errorf("failed to get project's file: %w", err)
				}
				defer projsf.Close()

				return ctrl.ConfigValidateRemote(cmd.Context(), tomlf, projsf) //nolint:wrapcheck
			}

			secretsf, err := system.GetSecretsFile()
			if err != nil {
				return fmt.Errorf("failed to get config app file: %w", err)
			}
			defer secretsf.Close()

			return ctrl.ConfigValidate(tomlf, secretsf) //nolint:wrapcheck
		},
	}
}
