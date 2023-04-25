package controller

import (
	"context"
	"fmt"
	"io"

	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/project"
)

func (c *Controller) SecretsList(
	ctx context.Context,
	projectf io.Reader,
) error {
	proj, err := project.UnmarshalProjectInfo(projectf)
	if err != nil {
		return err //nolint:wrapcheck
	}

	session, err := c.GetNhostSession(ctx)
	if err != nil {
		return err
	}

	secrets, err := c.cl.GetSecrets(
		ctx,
		proj.ID,
		graphql.WithAccessToken(session.Session.AccessToken),
	)
	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	for _, secret := range secrets.GetAppSecrets() {
		c.p.Println(secret.Name)
	}

	return nil
}
