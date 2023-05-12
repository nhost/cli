package controller

import (
	"context"
	"fmt"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/be/services/mimir/schema"
	"github.com/nhost/be/services/mimir/schema/appconfig"
	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/project/env"
	"github.com/nhost/cli/v2/system"
	"github.com/nhost/cli/v2/tui"
	"github.com/pelletier/go-toml/v2"
)

func ConfigValidate(p Printer, fs *system.PathStructure) (*model.ConfigConfig, error) {
	p.Println(tui.Info("Verifying configuration..."))

	cfg := &model.ConfigConfig{} //nolint:exhaustruct
	if err := UnmarshalFile(fs.NhostToml(), cfg, toml.Unmarshal); err != nil {
		return nil, err
	}

	var secrets model.Secrets
	if err := UnmarshalFile(fs.Secrets(), &secrets, env.Unmarshal); err != nil {
		return nil, fmt.Errorf("failed to parse secrets: %w", err)
	}

	schema, err := schema.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	cfg, err = appconfig.Config(schema, cfg, secrets)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	p.Println(tui.Info("Configuration is valid!"))

	return cfg, nil
}

func ConfigValidateRemote(
	ctx context.Context,
	p Printer,
	cl NhostClient,
	fs *system.PathStructure,
) error {
	var cfg *model.ConfigConfig
	if err := UnmarshalFile(fs.NhostToml(), cfg, toml.Unmarshal); err != nil {
		return err
	}

	schema, err := schema.New()
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	proj, err := GetAppInfo(ctx, p, cl, fs)
	if err != nil {
		return err
	}

	session, err := LoadSession(ctx, p, cl, fs)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
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

	_, err = appconfig.Config(schema, cfg, respToSecrets(secrets.GetAppSecrets(), false))
	if err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	p.Println(tui.Info("Config is valid!"))

	return nil
}
