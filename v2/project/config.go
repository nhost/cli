package project

import (
	"fmt"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/be/services/mimir/schema"
	"github.com/nhost/cli/v2/system"
)

func ConfigToDisk(config *model.ConfigConfig) error {
	f, err := system.GetConfigFile()
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer f.Close()

	if err := system.MarshalTOML(config, f); err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return nil
}

func ConfigFromDisk() (*model.ConfigConfig, error) {
	f, err := system.GetConfigFile()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	defer f.Close()

	config := &model.ConfigConfig{}
	if err = system.UnmarshalTOML(f, config); err != nil {
		return nil, fmt.Errorf("failed to parse config.toml: %w", err)
	}
	return config, err
}

func DefaultConfig() (*model.ConfigConfig, error) {
	s, err := schema.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	c := &model.ConfigConfig{ //nolint:exhaustruct
		Hasura: &model.ConfigHasura{ //nolint:exhaustruct
			AdminSecret:   "{{ secrets.HASURA_GRAPHQL_ADMIN_SECRET }}",
			WebhookSecret: "{{ secrets.NHOST_WEBHOOK_SECRET }}",
			JwtSecrets: []*model.ConfigJWTSecret{
				{
					Type: ptr("HS256"),
					Key:  ptr("{{ secrets.HASURA_GRAPHQL_JWT_SECRET }}"),
				},
			},
		},
	}

	if c, err = s.Fill(c); err != nil {
		return nil, fmt.Errorf("failed to fill config: %w", err)
	}

	if err = s.ValidateConfig(c); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return c, nil
}
