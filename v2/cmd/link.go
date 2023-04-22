package cmd

import (
	"fmt"

	"github.com/nhost/cli/v2/controller"
	"github.com/nhost/cli/v2/nhostclient"
	"github.com/nhost/cli/v2/system"
	"github.com/spf13/cobra"
)

func linkCmd() *cobra.Command {
	return &cobra.Command{ //nolint:exhaustruct
		Use:        "link",
		SuggestFor: []string{"init"},
		Short:      "Link local app to a remote one",
		Long:       `Connect your already hosted Nhost app to local environment and start development or testings.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cl := nhostclient.New(cmd.Flag(flagDomain).Value.String())
			ctrl := controller.New(cmd, cl, GetNhostCredentials)

			f, err := system.GetNhostProjectFile()
			if err != nil {
				return fmt.Errorf("failed to get config app file: %w", err)
			}

			return ctrl.Link(cmd.Context(), f) //nolint:wrapcheck
		},
	}
}
