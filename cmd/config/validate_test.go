package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/cmd/config"
)

func ptr[T any](t T) *T {
	return &t
}

func expectedConfig() *model.ConfigConfig {
	//nolint:exhaustruct
	return &model.ConfigConfig{
		Global: &model.ConfigGlobal{
			Environment: []*model.ConfigEnvironmentVariable{
				{Name: "ENVIRONMENT", Value: "development"},
				{Name: "FUNCTION_LOG_LEVEL", Value: "debug"},
			},
		},
		Hasura: &model.ConfigHasura{
			Version: ptr("v2.25.0-ce"),
			JwtSecrets: []*model.ConfigJWTSecret{
				{
					Type: ptr("HS256"),
					Key:  ptr("0f987876650b4a085e64594fae9219e7781b17506bec02489ad061fba8cb22db"),
				},
			},
			AdminSecret:   "nhost-admin-secret",
			WebhookSecret: "nhost-webhook-secret",
			Settings: &model.ConfigHasuraSettings{
				CorsDomain:                    []string{"*"},
				DevMode:                       ptr(true),
				EnableAllowList:               ptr(false),
				EnableConsole:                 ptr(true),
				EnableRemoteSchemaPermissions: new(bool),
				EnabledAPIs: []string{
					"metadata",
					"graphql",
					"pgdump",
					"config",
				},
			},
			Logs:   &model.ConfigHasuraLogs{Level: ptr("warn")},
			Events: &model.ConfigHasuraEvents{HttpPoolSize: ptr(uint32(100))},
		},
		Functions: &model.ConfigFunctions{Node: &model.ConfigFunctionsNode{Version: ptr(16)}},
		Auth: &model.ConfigAuth{
			Version: ptr("0.20.0"),
			Redirections: &model.ConfigAuthRedirections{
				ClientUrl:   ptr("http://localhost:3000"),
				AllowedUrls: []string{},
			},
			SignUp: &model.ConfigAuthSignUp{Enabled: ptr(true)},
			User: &model.ConfigAuthUser{
				Roles: &model.ConfigAuthUserRoles{
					Default: ptr("user"),
					Allowed: []string{"user", "me"},
				},
				Locale: &model.ConfigAuthUserLocale{
					Default: ptr("en"),
					Allowed: []string{"en"},
				},
				Gravatar: &model.ConfigAuthUserGravatar{
					Enabled: ptr(true),
					Default: ptr("blank"),
					Rating:  ptr("g"),
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
					ExpiresIn:    ptr(uint32(900)),
					CustomClaims: []*model.ConfigAuthsessionaccessTokenCustomClaims{},
				},
				RefreshToken: &model.ConfigAuthSessionRefreshToken{
					ExpiresIn: ptr(uint32(43200)),
				},
			},
			Method: &model.ConfigAuthMethod{
				Anonymous: &model.ConfigAuthMethodAnonymous{
					Enabled: ptr(false),
				},
				EmailPasswordless: &model.ConfigAuthMethodEmailPasswordless{
					Enabled: ptr(false),
				},
				EmailPassword: &model.ConfigAuthMethodEmailPassword{
					HibpEnabled:               ptr(false),
					EmailVerificationRequired: ptr(true),
					PasswordMinLength:         ptr(uint8(9)),
				},
				SmsPasswordless: &model.ConfigAuthMethodSmsPasswordless{
					Enabled: ptr(false),
				},
				Oauth: &model.ConfigAuthMethodOauth{
					Apple: &model.ConfigAuthMethodOauthApple{
						Enabled: ptr(false),
					},
					Azuread: &model.ConfigAuthMethodOauthAzuread{
						Enabled: ptr(false),
						Tenant:  ptr("common"),
					},
					Bitbucket: &model.ConfigStandardOauthProvider{
						Enabled: ptr(false),
					},
					Discord: &model.ConfigStandardOauthProviderWithScope{
						Enabled: ptr(false),
					},
					Facebook: &model.ConfigStandardOauthProviderWithScope{
						Enabled: ptr(false),
					},
					Github: &model.ConfigStandardOauthProviderWithScope{
						Enabled: ptr(false),
					},
					Gitlab: &model.ConfigStandardOauthProviderWithScope{
						Enabled: ptr(false),
					},
					Google: &model.ConfigStandardOauthProviderWithScope{
						Enabled: ptr(false),
					},
					Linkedin: &model.ConfigStandardOauthProviderWithScope{
						Enabled: ptr(false),
					},
					Spotify: &model.ConfigStandardOauthProviderWithScope{
						Enabled: ptr(false),
					},
					Strava: &model.ConfigStandardOauthProviderWithScope{
						Enabled: ptr(false),
					},
					Twitch: &model.ConfigStandardOauthProviderWithScope{
						Enabled: ptr(false),
					},
					Twitter: &model.ConfigAuthMethodOauthTwitter{
						Enabled: ptr(false),
					},
					Windowslive: &model.ConfigStandardOauthProviderWithScope{
						Enabled: ptr(false),
					},
					Workos: &model.ConfigAuthMethodOauthWorkos{
						Enabled: ptr(false),
					},
				},
				Webauthn: &model.ConfigAuthMethodWebauthn{
					Enabled:      ptr(false),
					RelyingParty: nil,
					Attestation: &model.ConfigAuthMethodWebauthnAttestation{
						Timeout: ptr(uint32(60000)),
					},
				},
			},
			Totp: &model.ConfigAuthTotp{Enabled: ptr(false)},
		},
		Postgres: &model.ConfigPostgres{Version: ptr("14.6-20230406-2")},
		Provider: &model.ConfigProvider{},
		Storage:  &model.ConfigStorage{Version: ptr("0.3.4")},
		Observability: &model.ConfigObservability{
			Grafana: &model.ConfigGrafana{AdminPassword: "grafana-admin-password"},
		},
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		path         string
		expected     func() *model.ConfigConfig
		applyPatches bool
	}{
		{
			name:         "applypatches",
			path:         "success",
			expected:     expectedConfig,
			applyPatches: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			ce := clienv.New(
				os.Stdout,
				os.Stderr,
				clienv.NewPathStructure(
					".",
					filepath.Join("testdata", "validate", tc.path),
					filepath.Join("testdata", "validate", tc.path, ".nhost"),
					filepath.Join("testdata", "validate", tc.path, ".nhost", "data"),
					filepath.Join("testdata", "validate", tc.path, "nhost"),
				),
				"fakedomain",
				"",
			)

			cfg, err := config.Validate(ce, "local")
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expected(), cfg); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
