package cmd

import (
	"fmt"

	"github.com/nhost/cli/v2/cmd/workflows"
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
		// PreRun: func(cmd *cobra.Command, args []string) {
		// 	//  if user is already logged in, ask to logout
		// 	if _, err := getUser(nhost.AUTH_PATH); err == nil {
		// 		status.Fatal(ErrLoggedIn)
		// 	}
		// },
		RunE: func(cmd *cobra.Command, _ []string) error {
			var err error

			email := cmd.Flag(flagEmail).Value.String()
			if email == "" {
				email, err = tui.UserInput(cmd.OutOrStdout(), "email", false)
				if err != nil {
					return fmt.Errorf("failed to read email: %w", err)
				}
			}

			password := cmd.Flag(flagPassword).Value.String()
			if password == "" {
				password, err = tui.UserInput(cmd.OutOrStdout(), "password", true)
				fmt.Fprintln(cmd.OutOrStdout())
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

			return workflows.Login(cmd.Context(), cmd.Print, f, email, password, cl) //nolint:wrapcheck
		},
	}
}
