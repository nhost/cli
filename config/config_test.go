package config_test

import (
	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/config"
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
		Name:        "Nhost application",
	}

	expectedAuthConfig := &model.ConfigAuth{
		Version: "0.17.0",
		Redirections: &model.ConfigAuthRedirections{
			ClientUrl:   "http://localhost:3000",
			AllowedUrls: []string{},
		},
		SignUp: &model.ConfigAuthSignUp{Enabled: true},
		User: &model.ConfigAuthUser{
			Roles: &model.ConfigAuthUserRoles{
				Default: "user",
				Allowed: []string{"user", "me"},
			},
			Locale: &model.ConfigAuthUserLocale{
				Default: "en",
				Allowed: []string{"en"},
			},
			Gravatar: &model.ConfigAuthUserGravatar{
				Enabled: true,
				Default: "blank",
				Rating:  "g",
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
				ExpiresIn:    900,
				CustomClaims: []*model.ConfigAuthsessionaccessTokenCustomClaims{},
			},
			RefreshToken: &model.ConfigAuthSessionRefreshToken{
				ExpiresIn: 43200,
			},
		},
		Method: &model.ConfigAuthMethod{
			Anonymous: &model.ConfigAuthMethodAnonymous{
				Enabled: false,
			},
			EmailPasswordless: &model.ConfigAuthMethodEmailPasswordless{
				Enabled: false,
			},
			EmailPassword: &model.ConfigAuthMethodEmailPassword{
				HibpEnabled:               false,
				EmailVerificationRequired: true,
				PasswordMinLength:         9,
			},
			SmsPasswordless: &model.ConfigAuthMethodSmsPasswordless{
				Enabled: false,
			},
			Oauth: &model.ConfigAuthMethodOauth{
				Apple: &model.ConfigAuthMethodOauthApple{
					Enabled: false,
				},
				Azuread: &model.ConfigAuthMethodOauthAzuread{
					Enabled: false,
					Tenant:  "common",
				},
				Bitbucket: &model.ConfigStandardOauthProvider{
					Enabled: false,
				},
				Discord: &model.ConfigStandardOauthProvider{
					Enabled: false,
				},
				Facebook: &model.ConfigStandardOauthProvider{
					Enabled: false,
					Scope:   []string{"email", "photos", "displayName"},
				},
				Github: &model.ConfigStandardOauthProvider{
					Enabled: false,
				},
				Gitlab: &model.ConfigStandardOauthProvider{
					Enabled: false,
				},
				Google: &model.ConfigStandardOauthProvider{
					Enabled: false,
				},
				Linkedin: &model.ConfigStandardOauthProvider{
					Enabled: false,
					Scope:   []string{"r_emailaddress", "r_liteprofile"},
				},
				Spotify: &model.ConfigStandardOauthProvider{
					Enabled: false,
				},
				Strava: &model.ConfigStandardOauthProvider{
					Enabled: false,
				},
				Twitch: &model.ConfigStandardOauthProvider{
					Enabled: false,
				},
				Twitter: &model.ConfigAuthMethodOauthTwitter{
					Enabled: false,
				},
				Windowslive: &model.ConfigStandardOauthProvider{
					Enabled: false,
				},
				Workos: &model.ConfigAuthMethodOauthWorkos{
					Enabled: false,
				},
			},
			Webauthn: &model.ConfigAuthMethodWebauthn{
				Enabled: false,
				Attestation: &model.ConfigAuthMethodWebauthnAttestation{
					Timeout: 60000,
				},
			},
		},
		Totp: &model.ConfigAuthTotp{
			Enabled: false,
		},
	}

	expectedProviderConfig := &model.ConfigProvider{
		Smtp: &model.ConfigSmtp{
			User:     "user",
			Password: "password",
			Sender:   "hasura-auth@example.com",
			Host:     "mailhog",
			Port:     ports.DefaultSMTPPort,
			Secure:   false,
			Method:   "PLAIN",
		},
	}

	expectedHasuraConfig := &model.ConfigHasura{
		Version: "v2.15.2",
		Settings: &model.ConfigHasuraSettings{
			EnableRemoteSchemaPermissions: false,
		},
		AdminSecret:   util.ADMIN_SECRET,
		WebhookSecret: util.WEBHOOK_SECRET,
		JwtSecrets: []*model.ConfigJWTSecret{
			{
				Type: "HS256",
				Key:  util.JWT_KEY,
			},
		},
	}

	expectedFunctionsConfig := &model.ConfigFunctions{
		Node: &model.ConfigFunctionsNode{
			Version: 16,
		},
	}

	expectedStorageConfig := &model.ConfigStorage{
		Version: "0.3.1",
	}

	assert.Equal(expectedGlobalConfig, defaultConfig.Global())
	assert.Equal(expectedAuthConfig, defaultConfig.Auth())
	assert.Equal(expectedProviderConfig, defaultConfig.Provider())
	assert.Equal(expectedHasuraConfig, defaultConfig.Hasura())
	assert.Equal(expectedFunctionsConfig, defaultConfig.Functions())
	assert.Equal(expectedStorageConfig, defaultConfig.Storage())
}
