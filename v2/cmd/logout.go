package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/nhost/cli/v2/controller"
	"github.com/nhost/cli/v2/nhostclient"
	"github.com/nhost/cli/v2/system"
	"github.com/nhost/cli/v2/tui"
	"github.com/spf13/cobra"
)

func logoutCmd() *cobra.Command {
	return &cobra.Command{ //nolint:exhaustruct
		Use:        "logout",
		SuggestFor: []string{"login"},
		Short:      "Log out your Nhost account",
		RunE: func(cmd *cobra.Command, _ []string) error {
			f, err := system.GetStateAuthFile()
			if err != nil {
				return fmt.Errorf("failed to get auth file: %w", err)
			}
			defer f.Close()

			cl := nhostclient.New(cmd.Flag(flagDomain).Value.String())
			ctrl := controller.New(cmd, cl, GetNhostCredentials)

			err = ctrl.Logout(cmd.Context())
			switch {
			case errors.Is(err, system.ErrNoContent):
				return nil
			case err != nil:
				cmd.Print(tui.Warn("%s\n", err.Error()))
			}
			cmd.Print(tui.Info("Deleting PAT from local storage\n"))
			if err := os.Remove(f.Name()); err != nil {
				cmd.Print(tui.Warn("failed to remove auth file: %s\n", err.Error()))
			}

			return nil
		},
	}
}
