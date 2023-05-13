package swmanagement

import (
	"fmt"
	"os"

	"github.com/nhost/cli/cmd"
	"github.com/urfave/cli/v2"
)

const (
	forceFlag = "force"
)

func CommandUninstall() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:     "uninstall",
		Aliases:  []string{},
		Usage:    "Remove the installed CLI from system permanently",
		Category: category,
		Action:   commandUninstall,
		Flags: []cli.Flag{
			&cli.BoolFlag{ //nolint:exhaustruct
				Name:        forceFlag,
				Usage:       "Force uninstall without confirmation",
				EnvVars:     []string{"NHOST_FORCE_UNINSTALL"},
				DefaultText: "false",
			},
		},
	}
}

func commandUninstall(cCtx *cli.Context) error {
	app := cmd.NewApplication(cCtx)

	path, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find installed CLI: %w", err)
	}

	if cCtx.App.Version == devVersion || cCtx.App.Version == "" {
		// we fake it in dev mode
		path = "/tmp/nhost"
	}

	app.Infoln("Found Nhost cli in %s", path)

	if !cCtx.Bool(forceFlag) {
		app.PromptMessage("Are you sure you want to uninstall Nhost CLI? [y/N] ")
		resp, err := app.PromptInput(false)
		if err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}

		if resp != "y" && resp != "Y" {
			return nil
		}
	}

	app.Infoln("Uninstalling Nhost CLI...")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove CLI: %w", err)
	}

	return nil
}
