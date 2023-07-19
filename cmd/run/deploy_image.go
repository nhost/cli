package run

import (
	"fmt"

	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/nhostclient/graphql"
	"github.com/urfave/cli/v2"
)

const (
	flagAppID     = "app-id"
	flagServiceID = "service-id"
	flagImage     = "image"
)

func ptr[T any](v T) *T { return &v }

func CommandUpdateImage() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:    "update-image",
		Aliases: []string{},
		Usage:   "Update service image leaving rest of parameters intact",
		Action:  commandUpdateImage,
		Flags: []cli.Flag{
			&cli.StringFlag{ //nolint:exhaustruct
				Name:     flagAppID,
				Usage:    "App ID the service belongs to",
				Required: true,
			},
			&cli.StringFlag{ //nolint:exhaustruct
				Name:     flagServiceID,
				Usage:    "Service ID to update",
				Required: true,
			},
			&cli.StringFlag{ //nolint:exhaustruct
				Name:     flagImage,
				Usage:    "Image to set",
				Required: true,
			},
		},
	}
}

func commandUpdateImage(cCtx *cli.Context) error {
	ce := clienv.FromCLI(cCtx)
	session, err := ce.LoadSession(cCtx.Context)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	nhostcl := ce.GetNhostClient()
	if _, err := nhostcl.UpdateRunServiceConfig(
		cCtx.Context,
		cCtx.String(flagAppID),
		cCtx.String(flagServiceID),
		graphql.ConfigRunServiceConfigUpdateInput{ //nolint:exhaustruct
			Image: &graphql.ConfigRunServiceImageUpdateInput{
				Image: ptr(cCtx.String(flagImage)),
			},
		},
		graphql.WithAccessToken(session.Session.AccessToken),
	); err != nil {
		return fmt.Errorf("failed to update image: %w", err)
	}

	return nil
}
