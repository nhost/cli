package main

import (
	"errors"
	"log"
	"os"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/nhost/cli/cmd/swmanagement"
	"github.com/urfave/cli/v2"
)

var Version string

func main() {
	app := &cli.App{ //nolint: exhaustruct
		Name:                 "nhost",
		EnableBashCompletion: true,
		Version:              Version,
		Description:          "Nhost CLI tool",
		Commands: []*cli.Command{
			swmanagement.CommandUninstall(),
			swmanagement.CommandUpgrade(),
			swmanagement.CommandVersion(),
		},
		Metadata: map[string]any{
			"Author":  "Nhost",
			"LICENSE": "MIT",
		},
		Flags: []cli.Flag{},
	}

	if err := app.Run(os.Args); err != nil {
		var graphqlErr *clientv2.ErrorResponse

		switch {
		case errors.As(err, &graphqlErr):
			log.Fatal(graphqlErr.GqlErrors)
		case err != nil:
			log.Fatal(err)
		}
	}
}
