package project

import (
	"fmt"
	"io"

	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/system"
)

func UnmarshalProjectInfo(r io.Reader) (*graphql.GetWorkspacesApps_Workspaces_Apps, error) {
	var project *graphql.GetWorkspacesApps_Workspaces_Apps
	err := system.UnmarshalJSON(r, &project)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal project information: %w", err)
	}
	return project, nil
}

func MarshalProjectInfo(project *graphql.GetWorkspacesApps_Workspaces_Apps, w io.Writer) error {
	return system.MarshalJSON(project, w) //nolint:wrapcheck
}
