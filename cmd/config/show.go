package config

import (
	"fmt"

	"github.com/nhost/cli/clienv"
	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v2"
)

const (
	flagSkipPatches = "skip-patches"
	flagOverlay     = "overlay"
)

func CommandShow() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:        "show",
		Aliases:     []string{},
		Usage:       "Shows configuration after applying jsonpatches and resolving secrets",
		Description: "Note that this command will always use the local secrets, even if you specify the overlay for a cloud project.", //nolint:lll
		Action:      commandShow,
		Flags: []cli.Flag{
			&cli.BoolFlag{ //nolint:exhaustruct
				Name:    flagSkipPatches,
				Usage:   "Skip applying jsonpatches",
				Value:   false,
				EnvVars: []string{"NHOST_SKIP_PATCHES"},
			},
			&cli.StringFlag{ //nolint:exhaustruct
				Name:    flagOverlay,
				Usage:   "Overaly to use",
				Value:   "local",
				EnvVars: []string{"NHOST_OVERLAY"},
			},
		},
	}
}

func commandShow(c *cli.Context) error {
	ce := clienv.FromCLI(c)

	cfg, err := Validate(ce, !c.Bool(flagSkipPatches), c.String(flagOverlay))
	if err != nil {
		return err
	}

	b, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	ce.Println(string(b))
	return nil
}
