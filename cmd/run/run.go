package run

import "github.com/urfave/cli/v2"

const (
	flagSubdomain = "subdomain"
)

func Command() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:    "run",
		Aliases: []string{},
		Usage:   "Perform operations on Nhost Run",
		Subcommands: []*cli.Command{
			CommandConfigShow(),
			CommandConfigDeploy(),
			CommandConfigEdit(),
			CommandConfigEditImage(),
			CommandConfigPull(),
			CommandConfigValidate(),
			CommandConfigExample(),
		},
	}
}
