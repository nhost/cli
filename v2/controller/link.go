package controller

import (
	"context"
	"fmt"
	"io"

	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/project"
	"github.com/nhost/cli/v2/tui"
)

func getApp(
	workspaces []*graphql.GetWorkspacesApps_Workspaces,
	idx string,
) (*graphql.GetWorkspacesApps_Workspaces_Apps, error) {
	x := 1
	var app *graphql.GetWorkspacesApps_Workspaces_Apps
OUTER:
	for _, ws := range workspaces {
		for _, a := range ws.GetApps() {
			if fmt.Sprintf("%d", x) == idx {
				a := a
				app = a
				break OUTER
			}
			x++
		}
	}

	if app == nil {
		return nil, fmt.Errorf("invalid input") //nolint:goerr113
	}

	return app, nil
}

func confirmApp(app *graphql.GetWorkspacesApps_Workspaces_Apps, p Printer) error {
	p.Print(tui.PromptMessage("Enter project subdomain to confirm: "))
	confirm, err := tui.PromptInput(false)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	if confirm != app.Subdomain {
		return fmt.Errorf("input doesn't match the subdomain") //nolint:goerr113
	}

	return nil
}

func (c *Controller) Link(ctx context.Context, projectf io.Writer) error {
	session, err := c.GetNhostSession(ctx)
	if err != nil {
		return fmt.Errorf("failed to get nhost session: %w", err)
	}

	workspaces, err := c.cl.GetWorkspacesApps(
		ctx,
		graphql.WithAccessToken(session.Session.AccessToken),
	)
	if err != nil {
		return fmt.Errorf("failed to get workspaces: %w", err)
	}

	if len(workspaces.GetWorkspaces()) == 0 {
		return fmt.Errorf("no workspaces found") //nolint:goerr113
	}

	if err := c.list(workspaces.GetWorkspaces()); err != nil {
		return err
	}

	c.p.Print(tui.PromptMessage("Select # the workspace to link: "))
	idx, err := tui.PromptInput(false)
	if err != nil {
		return fmt.Errorf("failed to read workspace: %w", err)
	}

	app, err := getApp(workspaces.GetWorkspaces(), idx)
	if err != nil {
		return err
	}

	if err := confirmApp(app, c.p); err != nil {
		return err
	}

	if err := project.MarshalProjectInfo(app, projectf); err != nil {
		return fmt.Errorf("failed to marshal project information: %w", err)
	}

	return nil
}
