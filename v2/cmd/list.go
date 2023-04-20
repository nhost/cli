package cmd

import (
	"fmt"

	"github.com/nhost/cli/v2/cmd/workflows"
	"github.com/nhost/cli/v2/nhostclient"
	"github.com/nhost/cli/v2/tui"
	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {
	return &cobra.Command{ //nolint:exhaustruct
		Use:        "list",
		Aliases:    []string{"ls"},
		SuggestFor: []string{"init"},
		Short:      "List remote apps",
		Long: `Fetch the list of remote personal and team apps
for the logged in user from Nhost console.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cl, session, err := workflows.GetNhostClient(cmd.Context(), cmd.Flag(flagDomain).Value.String())
			if err != nil {
				return err //nolint:wrapcheck
			}
			workspaces, err := cl.Graphql.GetWorkspacesApps(
				cmd.Context(),
				nhostclient.WithAccessToken(session.Session.AccessToken),
			)
			if err != nil {
				return fmt.Errorf("failed to get workspaces: %w", err)
			}

			if len(workspaces.GetWorkspaces()) == 0 {
				return fmt.Errorf("no workspaces found") //nolint:goerr113
			}

			workspace := tui.Column{
				Header: "Workspace",
				Rows:   make([]string, 0),
			}
			project := tui.Column{
				Header: "Project",
				Rows:   make([]string, 0),
			}

			for _, ws := range workspaces.Workspaces {
				for _, app := range ws.Apps {
					workspace.Rows = append(workspace.Rows, ws.Name)
					project.Rows = append(project.Rows, app.Name)
				}
			}

			if _, err = fmt.Fprint(cmd.OutOrStdout(), tui.Table(workspace, project)); err != nil {
				return fmt.Errorf("failed to print table: %w", err)
			}

			return nil
		},
	}
}
