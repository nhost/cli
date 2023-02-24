package config

import (
	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/be/services/mimir/schema"
	"github.com/nhost/cli/internal/generichelper"
	"github.com/nhost/cli/internal/ports"
	"github.com/nhost/cli/util"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	c *model.ConfigConfig
}

func NewConfig(c *model.ConfigConfig) *Config {
	return &Config{
		c: c,
	}
}

func (c Config) Validate() error {
	s, err := schema.New()
	if err != nil {
		return err
	}
	return s.ValidateConfig(c.c)

}

func (c Config) Global() *model.ConfigGlobal {
	return c.c.Global
}

func (c Config) Auth() *model.ConfigAuth {
	return c.c.Auth
}

func (c Config) Provider() *model.ConfigProvider {
	return c.c.Provider
}

func (c Config) Hasura() *model.ConfigHasura {
	return c.c.Hasura
}

func (c Config) Functions() *model.ConfigFunctions {
	return c.c.Functions
}

func (c Config) Storage() *model.ConfigStorage {
	return c.c.Storage
}

func (c Config) Marshal() ([]byte, error) {
	return toml.Marshal(c.c)
}

func DefaultConfig() (*Config, error) {
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

	return &Config{c: c}, nil
}

func defaultAuthConfig() *model.ConfigAuth {
	return &model.ConfigAuth{
		Redirections: &model.ConfigAuthRedirections{
			ClientUrl: generichelper.Pointerify("http://localhost:3000"),
		},
		Method: &model.ConfigAuthMethod{
			Oauth: &model.ConfigAuthMethodOauth{
				Apple: &model.ConfigAuthMethodOauthApple{
					Enabled: generichelper.Pointerify(false),
					Scope:   []string{"name", "email"},
				},
				Facebook: &model.ConfigStandardOauthProviderWithScope{
					Scope: []string{"email", "photos", "displayName"},
				},
				Linkedin: &model.ConfigStandardOauthProviderWithScope{
					Scope: []string{"r_emailaddress", "r_liteprofile"},
				},
				Google: &model.ConfigStandardOauthProviderWithScope{
					Scope: []string{"email", "profile"},
				},
				Gitlab: &model.ConfigStandardOauthProviderWithScope{
					Scope: []string{"read_user"},
				},
				Github: &model.ConfigStandardOauthProviderWithScope{
					Scope: []string{"user:email"},
				},
				Windowslive: &model.ConfigStandardOauthProviderWithScope{
					Scope: []string{"wl.basic", "wl.emails", "wl.contacts_emails"},
				},
				Spotify: &model.ConfigStandardOauthProviderWithScope{
					Scope: []string{"user-read-email", "user-read-private"},
				},
			},
		},
		SignUp: &model.ConfigAuthSignUp{},
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
