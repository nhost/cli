package swmanagement

import (
	"fmt"
	"os"
	"runtime"

	"github.com/nhost/cli/cmd"
	"github.com/nhost/cli/v2/software"
	"github.com/urfave/cli/v2"
)

func CommandUpgrade() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:     "upgrade",
		Aliases:  []string{},
		Usage:    "Upgrade the CLI to the latest version",
		Category: category,
		Action:   commandUpgrade,
	}
}

func commandUpgrade(cCtx *cli.Context) error {
	app := cmd.NewApplication(cCtx)

	mgr := software.NewManager()
	releases, err := mgr.GetReleases(cCtx.Context)
	if err != nil {
		return fmt.Errorf("failed to get releases: %w", err)
	}

	latest := releases[0]
	if latest.TagName == cCtx.App.Version {
		app.Infoln("You have the latest version. Hurray!")
		return nil
	}

	app.Infoln("Upgrading to %s...", latest.TagName)

	want := fmt.Sprintf("cli-%s-%s-%s.tar.gz", latest.TagName, runtime.GOOS, runtime.GOARCH)
	var url string
	for _, asset := range latest.Assets {
		if asset.Name == want {
			url = asset.BrowserDownloadURL
		}
	}

	if url == "" {
		return fmt.Errorf("failed to find asset for %s", want) //nolint:goerr113
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "nhost-cli-")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err := mgr.DownloadAsset(cCtx.Context, url, tmpFile); err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}

	return install(cCtx, app, tmpFile.Name())
}

func install(cCtx *cli.Context, app *cmd.Application, tmpFile string) error {
	curBin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find installed CLI: %w", err)
	}

	if cCtx.App.Version == devVersion || cCtx.App.Version == "" {
		// we are in dev mode, we fake curBin for testing
		curBin = "/tmp/nhost"
	}

	app.Infoln("Copying to %s...", curBin)
	if err := os.Rename(tmpFile, curBin); err != nil {
		return fmt.Errorf("failed to rename %s to %s: %w", tmpFile, curBin, err)
	}

	app.Infoln("Setting permissions...")
	if err := os.Chmod(curBin, 0o755); err != nil { //nolint:gomnd
		return fmt.Errorf("failed to set permissions on %s: %w", curBin, err)
	}

	return nil
}
