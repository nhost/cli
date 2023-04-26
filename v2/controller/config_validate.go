package controller

import (
	"context"
	"fmt"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/be/services/mimir/schema"
	"github.com/nhost/be/services/mimir/schema/appconfig"
	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/project"
	"github.com/nhost/cli/v2/tui"
)

func respToSecrets(env []*graphql.GetSecrets_AppSecrets) model.Secrets {
	secrets := make(model.Secrets, len(env))
	for i, s := range env {
		secrets[i] = &model.ConfigEnvironmentVariable{
			Name:  s.Name,
			Value: s.Value,
		}
	}
	return secrets
}

func ConfigValidate(p Printer) error {
	cfg, err := project.ConfigFromDisk()
	if err != nil {
		return err //nolint:wrapcheck
	}

	secrets, err := project.SecretsFromDisk()
	if err != nil {
		return fmt.Errorf("failed to parse secrets: %w", err)
	}

	schema, err := schema.New()
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	_, err = appconfig.Config(schema, cfg, secrets)
	if err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	p.Println(tui.Info("Config is valid!"))

	return nil
}

func ConfigValidateRemote(
	ctx context.Context,
	p Printer,
	cl NhostClient,
) error {
	cfg, err := project.ConfigFromDisk()
	if err != nil {
		return err //nolint:wrapcheck
	}

	schema, err := schema.New()
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	proj, err := project.InfoFromDisk()
	if err != nil {
		return err //nolint:wrapcheck
	}

	session, err := GetNhostSession(ctx, cl)
	if err != nil {
		return err
	}

	p.Println(tui.Info("Getting secrets..."))
	secrets, err := cl.GetSecrets(
		ctx,
		proj.ID,
		graphql.WithAccessToken(session.Session.AccessToken),
	)
	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	_, err = appconfig.Config(schema, cfg, respToSecrets(secrets.GetAppSecrets()))
	if err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	p.Println(tui.Info("Config is valid!"))

	return nil
}
