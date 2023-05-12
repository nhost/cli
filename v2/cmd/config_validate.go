package cmd

import (
	"fmt"

	"github.com/nhost/cli/v2/controller"
	"github.com/nhost/cli/v2/nhostclient"
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

			if validateRemote {
				cl := nhostclient.New(cmd.Flag(flagDomain).Value.String())
				return controller.ConfigValidateRemote(cmd.Context(), cmd, cl) //nolint:wrapcheck
			}

			_, err = controller.ConfigValidate(cmd)
			return err //nolint:wrapcheck
		},
	}
}
