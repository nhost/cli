package controller

import (
	"context"
	"fmt"
	"io"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/be/services/mimir/schema"
	"github.com/nhost/be/services/mimir/schema/appconfig"
	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/project"
	"github.com/nhost/cli/v2/system"
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

func (c *Controller) ConfigValidate(
	tomlf io.Reader,
	secretsf io.Reader,
) error {
	var v any
	if err := system.UnmarshalTOML(tomlf, &v); err != nil {
		return fmt.Errorf("failed to parse config.toml: %w", err)
	}

	schema, err := schema.New()
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	cfg, err := schema.Fill(v)
	if err != nil {
		return fmt.Errorf("failed to apply config to the schema: %w", err)
	}

	secrets, err := project.UnmarshalSecrets(secretsf)
	if err != nil {
		return fmt.Errorf("failed to parse secrets: %w", err)
	}

	_, err = appconfig.Config(schema, cfg, secrets)
	if err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	c.p.Println(tui.Info("Config is valid!"))

	return nil
}

func (c *Controller) ConfigValidateRemote(
	ctx context.Context,
	tomlf io.Reader,
	projectf io.Reader,
) error {
	var v any
	if err := system.UnmarshalTOML(tomlf, &v); err != nil {
		return fmt.Errorf("failed to parse config.toml: %w", err)
	}

	schema, err := schema.New()
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	cfg, err := schema.Fill(v)
	if err != nil {
		return fmt.Errorf("failed to apply config to the schema: %w", err)
	}

	proj, err := project.UnmarshalProjectInfo(projectf)
	if err != nil {
		return err //nolint:wrapcheck
	}

	session, err := c.GetNhostSession(ctx)
	if err != nil {
		return err
	}

	c.p.Println(tui.Info("Getting secrets..."))
	secrets, err := c.cl.GetSecrets(
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

	c.p.Println(tui.Info("Config is valid!"))

	return nil
}
