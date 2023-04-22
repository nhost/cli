package cmd

import (
	"fmt"

	"github.com/nhost/cli/v2/controller"
	"github.com/nhost/cli/v2/nhostclient"
	"github.com/nhost/cli/v2/system"
	"github.com/nhost/cli/v2/tui"
	"github.com/spf13/cobra"
)

const (
	flagEmail    = "email"
	flagPassword = "password"
)

// loginCmd represents the login command.
func logincCmd() *cobra.Command {
	return &cobra.Command{ //nolint:exhaustruct
		Use:        "login",
		SuggestFor: []string{"logout"},
		Short:      "Log in to your Nhost account",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var err error

			email := cmd.Flag(flagEmail).Value.String()
			if email == "" {
				cmd.Println(tui.PromptMessage("email: "))
				email, err = tui.PromptInput(false)
				if err != nil {
					return fmt.Errorf("failed to read email: %w", err)
				}
			}

			password := cmd.Flag(flagPassword).Value.String()
			if password == "" {
				cmd.Println(tui.PromptMessage("password: "))
				password, err = tui.PromptInput(false)
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
			}

			f, err := system.GetStateAuthFile()
			if err != nil {
				return fmt.Errorf("failed to get auth file: %w", err)
			}
			defer f.Close()

			cl := nhostclient.New(cmd.Flag(flagDomain).Value.String())
			ctrl := controller.New(cmd, cl, GetNhostCredentials)

			return ctrl.Login(cmd.Context(), f, email, password) //nolint:wrapcheck
		},
	}
}
