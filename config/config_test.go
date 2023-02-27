package config_test

import (
	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/config"
	"github.com/nhost/cli/internal/generichelper"
	"github.com/nhost/cli/internal/ports"
	"github.com/nhost/cli/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	defaultConfig, err := config.DefaultConfig()
	assert.NoError(err)

	expectedGlobalConfig := &model.ConfigGlobal{
		Environment: []*model.ConfigEnvironmentVariable{},
	}

	expectedAuthConfig := &model.ConfigAuth{
		Version: generichelper.Pointerify("0.19.0"),
		Redirections: &model.ConfigAuthRedirections{
			ClientUrl:   generichelper.Pointerify("http://localhost:3000"),
			AllowedUrls: []string{},
		},
		SignUp: &model.ConfigAuthSignUp{Enabled: generichelper.Pointerify(true)},
		User: &model.ConfigAuthUser{
			Roles: &model.ConfigAuthUserRoles{
				Default: generichelper.Pointerify("user"),
				Allowed: []string{"user", "me"},
			},
			Locale: &model.ConfigAuthUserLocale{
				Default: generichelper.Pointerify("en"),
				Allowed: []string{"en"},
			},
			Gravatar: &model.ConfigAuthUserGravatar{
				Enabled: generichelper.Pointerify(true),
				Default: generichelper.Pointerify("blank"),
				Rating:  generichelper.Pointerify("g"),
			},
			Email: &model.ConfigAuthUserEmail{
				Allowed: []string{},
				Blocked: []string{},
			},
			EmailDomains: &model.ConfigAuthUserEmailDomains{
				Allowed: []string{},
				Blocked: []string{},
			},
		},
		Session: &model.ConfigAuthSession{
			AccessToken: &model.ConfigAuthSessionAccessToken{
				ExpiresIn:    generichelper.Pointerify(uint32(900)),
				CustomClaims: []*model.ConfigAuthsessionaccessTokenCustomClaims{},
			},
			RefreshToken: &model.ConfigAuthSessionRefreshToken{
				ExpiresIn: generichelper.Pointerify(uint32(43200)),
			},
		},
		Method: &model.ConfigAuthMethod{
			Anonymous: &model.ConfigAuthMethodAnonymous{
				Enabled: generichelper.Pointerify(false),
			},
			EmailPasswordless: &model.ConfigAuthMethodEmailPasswordless{
				Enabled: generichelper.Pointerify(false),
			},
			EmailPassword: &model.ConfigAuthMethodEmailPassword{
				HibpEnabled:               generichelper.Pointerify(false),
				EmailVerificationRequired: generichelper.Pointerify(true),
				PasswordMinLength:         generichelper.Pointerify(uint8(9)),
			},
			SmsPasswordless: &model.ConfigAuthMethodSmsPasswordless{
				Enabled: generichelper.Pointerify(false),
			},
			Oauth: &model.ConfigAuthMethodOauth{
				Apple: &model.ConfigAuthMethodOauthApple{
					Enabled: generichelper.Pointerify(false),
				},
				Azuread: &model.ConfigAuthMethodOauthAzuread{
					Enabled: generichelper.Pointerify(false),
					Tenant:  generichelper.Pointerify("common"),
				},
				Bitbucket: &model.ConfigStandardOauthProvider{
					Enabled: generichelper.Pointerify(false),
				},
				Discord: &model.ConfigStandardOauthProviderWithScope{
					Enabled: generichelper.Pointerify(false),
				},
				Facebook: &model.ConfigStandardOauthProviderWithScope{
					Enabled: generichelper.Pointerify(false),
				},
				Github: &model.ConfigStandardOauthProviderWithScope{
					Enabled: generichelper.Pointerify(false),
				},
				Gitlab: &model.ConfigStandardOauthProviderWithScope{
					Enabled: generichelper.Pointerify(false),
				},
				Google: &model.ConfigStandardOauthProviderWithScope{
					Enabled: generichelper.Pointerify(false),
				},
				Linkedin: &model.ConfigStandardOauthProviderWithScope{
					Enabled: generichelper.Pointerify(false),
				},
				Spotify: &model.ConfigStandardOauthProviderWithScope{
					Enabled: generichelper.Pointerify(false),
				},
				Strava: &model.ConfigStandardOauthProviderWithScope{
					Enabled: generichelper.Pointerify(false),
				},
				Twitch: &model.ConfigStandardOauthProviderWithScope{
					Enabled: generichelper.Pointerify(false),
				},
				Twitter: &model.ConfigAuthMethodOauthTwitter{
					Enabled: generichelper.Pointerify(false),
				},
				Windowslive: &model.ConfigStandardOauthProviderWithScope{
					Enabled: generichelper.Pointerify(false),
				},
				Workos: &model.ConfigAuthMethodOauthWorkos{
					Enabled: generichelper.Pointerify(false),
				},
			},
			Webauthn: &model.ConfigAuthMethodWebauthn{
				Enabled: generichelper.Pointerify(false),
				Attestation: &model.ConfigAuthMethodWebauthnAttestation{
					Timeout: generichelper.Pointerify(uint32(60000)),
				},
			},
		},
		Totp: &model.ConfigAuthTotp{
			Enabled: generichelper.Pointerify(false),
		},
	}

	expectedProviderConfig := &model.ConfigProvider{
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

	expectedHasuraConfig := &model.ConfigHasura{
		Version: generichelper.Pointerify("v2.15.2"),
		Settings: &model.ConfigHasuraSettings{
			EnableRemoteSchemaPermissions: generichelper.Pointerify(false),
		},
		AdminSecret:   util.ADMIN_SECRET,
		WebhookSecret: util.WEBHOOK_SECRET,
		JwtSecrets: []*model.ConfigJWTSecret{
			{
				Type: generichelper.Pointerify("HS256"),
				Key:  generichelper.Pointerify(util.JWT_KEY),
			},
		},
	}

	expectedFunctionsConfig := &model.ConfigFunctions{
		Node: &model.ConfigFunctionsNode{
			Version: generichelper.Pointerify(16),
		},
	}

	expectedStorageConfig := &model.ConfigStorage{
		Version: generichelper.Pointerify("0.3.3"),
	}

	assert.Equal(expectedGlobalConfig, defaultConfig.Global)
	assert.Equal(expectedAuthConfig, defaultConfig.Auth)
	assert.Equal(expectedProviderConfig, defaultConfig.Provider)
	assert.Equal(expectedHasuraConfig, defaultConfig.Hasura)
	assert.Equal(expectedFunctionsConfig, defaultConfig.Functions)
	assert.Equal(expectedStorageConfig, defaultConfig.Storage)
}
