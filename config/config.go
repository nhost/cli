package config

import (
	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/be/services/mimir/schema"
	"github.com/nhost/cli/internal/generichelper"
	"github.com/nhost/cli/internal/ports"
	"github.com/nhost/cli/util"
)

func DefaultConfig() (*model.ConfigConfig, error) {
	s, err := schema.New()
	if err != nil {
		return nil, err
	}

	c := &model.ConfigConfig{
		Auth:     defaultAuthConfig(),
		Provider: defaultProviderConfig(),
		Hasura:   defaultHasuraConfig(),
	}

	c, err = s.Fill(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func defaultAuthConfig() *model.ConfigAuth {
	return &model.ConfigAuth{
		Redirections: &model.ConfigAuthRedirections{
			ClientUrl: generichelper.Pointerify("http://localhost:3000"),
		},
	}
}

func defaultProviderConfig() *model.ConfigProvider {
	return &model.ConfigProvider{
		Smtp: &model.ConfigSmtp{
			User:     "user",
			Password: "password",
			Sender:   "hasura-auth@example.com",
			Host:     "mailhog",
			Port:     uint16(ports.DefaultSMTPPort),
			Secure:   false,
			Method:   "PLAIN",
		},
	}
}

func defaultHasuraConfig() *model.ConfigHasura {
	return &model.ConfigHasura{
		AdminSecret:   util.ADMIN_SECRET,
		WebhookSecret: util.WEBHOOK_SECRET,
		JwtSecrets: []*model.ConfigJWTSecret{
			{
				Type: generichelper.Pointerify("HS256"),
				Key:  generichelper.Pointerify(util.JWT_KEY),
			},
		},
	}
}
