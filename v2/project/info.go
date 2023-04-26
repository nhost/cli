package project

import (
	"fmt"

	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/system"
)

func InfoFromDisk() (*graphql.GetWorkspacesApps_Workspaces_Apps, error) {
	f, err := system.GetNhostProjectInfoFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get project info file: %w", err)
	}
	defer f.Close()

	var project *graphql.GetWorkspacesApps_Workspaces_Apps
	if err := system.UnmarshalJSON(f, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project information: %w", err)
	}
	return project, nil
}

func InfoToDisk(project *graphql.GetWorkspacesApps_Workspaces_Apps) error {
	f, err := system.GetNhostProjectInfoFile()
	if err != nil {
		return fmt.Errorf("failed to get project info file: %w", err)
	}
	defer f.Close()

	return system.MarshalJSON(project, f) //nolint:wrapcheck
}
