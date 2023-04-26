package controller

import (
	"context"
	"fmt"

	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/project"
)

func SecretsList(
	ctx context.Context,
	p Printer,
	cl NhostClient,
) error {
	proj, err := project.InfoFromDisk()
	if err != nil {
		return err //nolint:wrapcheck
	}

	session, err := GetNhostSession(ctx, cl)
	if err != nil {
		return err
	}

	secrets, err := cl.GetSecrets(
		ctx,
		proj.ID,
		graphql.WithAccessToken(session.Session.AccessToken),
	)
	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	for _, secret := range secrets.GetAppSecrets() {
		p.Println(secret.Name)
	}

	return nil
}
