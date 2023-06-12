package clienv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/nhost/cli/nhostclient/graphql"
)

func getRemoteAppInfo(
	ctx context.Context,
	ce *CliEnv,
	subdomain string,
) (*graphql.GetWorkspacesApps_Workspaces_Apps, error) {
	session, err := ce.LoadSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	cl := ce.GetNhostClient()
	workspaces, err := cl.GetWorkspacesApps(
		ctx,
		graphql.WithAccessToken(session.Session.AccessToken),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}

	for _, workspace := range workspaces.Workspaces {
		for _, app := range workspace.Apps {
			if app.Subdomain == subdomain {
				return app, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to find app with subdomain: %s", subdomain) //nolint:goerr113
}

func (ce *CliEnv) GetAppInfo(
	ctx context.Context,
	subdomain string,
) (*graphql.GetWorkspacesApps_Workspaces_Apps, error) {
	if subdomain != "" {
		return getRemoteAppInfo(ctx, ce, subdomain)
	}

	var project *graphql.GetWorkspacesApps_Workspaces_Apps
	if err := UnmarshalFile(ce.Path.ProjectFile(), &project, json.Unmarshal); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			project, err = ce.Link(ctx)
			if err != nil {
				return nil, err
			}
		} else {
			ce.Warnln("Failed to find linked project: %v", err)
			ce.Infoln("Please run `nhost link` to link a project first")
			return nil, err
		}
	}

	return project, nil
}
