package swmanagement

import (
	"fmt"
	"runtime"

	"github.com/nhost/cli/cmd"
	"github.com/nhost/cli/v2/software"
	"github.com/urfave/cli/v2"
)

func CommandVersion() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:     "version",
		Aliases:  []string{},
		Usage:    "Show the current version of Nhost CLI you have installed",
		Category: category,
		Action:   commandVersion,
	}
}

func commandVersion(cCtx *cli.Context) error {
	app := cmd.NewApplication(cCtx)

	app.Infoln("Nhost CLI %s for %s-%s", cCtx.App.Version, runtime.GOOS, runtime.GOARCH)

	mgr := software.NewManager()
	releases, err := mgr.GetReleases(cCtx.Context)
	if err != nil {
		return fmt.Errorf("failed to get releases: %w", err)
	}

	latest := releases[0]
	if latest.TagName == cCtx.App.Version {
		return nil
	}

	app.Warnln("A new version of Nhost CLI is available: %s", latest.TagName)
	app.Println("You can upgrade by running `nhost upgrade`")

	if cCtx.App.Version == devVersion || cCtx.App.Version == "" {
		return nil
	}

	app.Println("Changes since your current version:")
	for _, release := range releases {
		if release.Prerelease {
			continue
		}
		app.Infoln("%s", release.TagName)
		app.Println(release.Body)
	}

	return nil
}
