package controller

import (
	"context"
	"fmt"
	"io"

	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/project"
	"github.com/nhost/cli/v2/tui"
)

func (c *Controller) SecretsDelete(
	ctx context.Context,
	projectf io.Reader,
	name string,
) error {
	proj, err := project.UnmarshalProjectInfo(projectf)
	if err != nil {
		return err //nolint:wrapcheck
	}

	session, err := c.GetNhostSession(ctx)
	if err != nil {
		return err
	}

	if _, err := c.cl.DeleteSecret(
		ctx,
		proj.ID,
		name,
		graphql.WithAccessToken(session.Session.AccessToken),
	); err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	c.p.Println(tui.Info("Secret deleted successfully!"))

	return nil
}
