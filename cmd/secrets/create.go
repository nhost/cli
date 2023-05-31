package secrets //nolint:dupl

import (
	"fmt"

	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/nhostclient/graphql"
	"github.com/urfave/cli/v2"
)

func CommandCreate() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:      "create",
		ArgsUsage: "NAME VALUE",
		Aliases:   []string{},
		Usage:     "Create secret in the cloud environment",
		Action:    commandCreate,
		Flags:     []cli.Flag{},
	}
}

func commandCreate(cCtx *cli.Context) error {
	if cCtx.NArg() != 2 { //nolint:gomnd
		return fmt.Errorf("invalid number of arguments") //nolint:goerr113
	}

	ce := clienv.FromCLI(cCtx)
	proj, err := ce.GetAppInfo(cCtx.Context)
	if err != nil {
		return fmt.Errorf("failed to get app info: %w", err)
	}

	session, err := ce.LoadSession(cCtx.Context)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	cl := ce.GetNhostClient()
	if _, err := cl.CreateSecret(
		cCtx.Context,
		proj.ID,
		cCtx.Args().Get(0),
		cCtx.Args().Get(1),
		graphql.WithAccessToken(session.Session.AccessToken),
	); err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	ce.Infoln("Secret created successfully!")

	return nil
}