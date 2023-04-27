package workflows

import (
	"fmt"

	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/tui"
)

func list(p Printer, workspaces []*graphql.GetWorkspacesApps_Workspaces) error {
	if len(workspaces) == 0 {
		return fmt.Errorf("no workspaces found") //nolint:goerr113
	}

	num := tui.Column{
		Header: "#",
		Rows:   make([]string, 0),
	}
	subdomain := tui.Column{
		Header: "Subdomain",
		Rows:   make([]string, 0),
	}
	project := tui.Column{
		Header: "Project",
		Rows:   make([]string, 0),
	}
	workspace := tui.Column{
		Header: "Workspace",
		Rows:   make([]string, 0),
	}
	region := tui.Column{
		Header: "Region",
		Rows:   make([]string, 0),
	}

	for _, ws := range workspaces {
		for _, app := range ws.Apps {
			num.Rows = append(num.Rows, fmt.Sprintf("%d", len(num.Rows)+1))
			subdomain.Rows = append(subdomain.Rows, app.Subdomain)
			project.Rows = append(project.Rows, app.Name)
			workspace.Rows = append(workspace.Rows, ws.Name)
			region.Rows = append(region.Rows, app.Region.AwsName)
		}
	}

	p.Println(tui.Table(num, subdomain, project, workspace, region))

	return nil
}
