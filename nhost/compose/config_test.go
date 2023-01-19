package compose

import (
	"github.com/compose-spec/compose-go/types"
	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/config"
	"github.com/nhost/cli/internal/ports"
	"github.com/stretchr/testify/assert"
	"testing"
)

func defaultConfig(t *testing.T) *config.Config {
	t.Helper()
	conf, err := config.DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	return conf
}

func defaultPorts(t *testing.T) *ports.Ports {
	t.Helper()
	return ports.NewPorts(
		ports.DefaultProxyPort,
		ports.DefaultDBPort,
		ports.DefaultGraphQLPort,
		ports.DefaultHasuraConsolePort,
		ports.DefaultHasuraConsoleApiPort,
		ports.DefaultSMTPPort,
		ports.DefaultS3MinioPort,
		ports.DefaultMailhogPort,
		ports.DefaultDashboardPort,
	)
}

func TestConfig_dashboardService(t *testing.T) {
	assert := assert.New(t)

	c := &Config{
		dotenv: []string{"FOO=BAR", "BAR=BAZ"},
		ports: ports.NewPorts(
			1, 2, 3, 4, 5, 6, 7, 8, 9,
		),
	}

	svc := c.dashboardService()
	assert.Equal("dashboard", svc.Name)
	assert.Equal("nhost/dashboard:0.9.9", svc.Image)
	assert.Equal([]types.ServicePortConfig{
		{
			Mode:      "ingress",
			Target:    3000,
			Published: "9",
			Protocol:  "tcp",
		},
	}, svc.Ports)
	assert.Equal(types.NewMappingWithEquals([]string{
		"FOO=BAR",
		"BAR=BAZ",
		"NEXT_PUBLIC_NHOST_LOCAL_BACKEND_PORT=1",
		"NEXT_PUBLIC_NHOST_HASURA_PORT=4",
		"NEXT_PUBLIC_NHOST_MIGRATIONS_PORT=5",
		"NEXT_PUBLIC_NHOST_PLATFORM=false",
		"NEXT_PUBLIC_ENV=dev",
		"NEXT_TELEMETRY_DISABLED=1",
	}), svc.Environment)
}

func TestConfig_addLocaldevExtraHost(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	c := &Config{}
	svc := &types.ServiceConfig{}
	c.addLocaldevExtraHost(svc)

	assert.Equal(svc.ExtraHosts["host.docker.internal"], "host-gateway")
}

func TestConfig_hasuraServiceEnvs(t *testing.T) {
	t.Parallel()

	a := assert.New(t)

	tests := []struct {
		name     string
		makeConf func() *config.Config
		want     env
	}{
		{
			name:     "with default config",
			makeConf: func() *config.Config { return defaultConfig(t) },
			want: env{
				"HASURA_GRAPHQL_DATABASE_URL":              "postgres://nhost_hasura@postgres:5432/postgres",
				"HASURA_GRAPHQL_JWT_SECRET":                "{\"type\":\"HS256\", \"key\": \"0f987876650b4a085e64594fae9219e7781b17506bec02489ad061fba8cb22db\"}",
				"HASURA_GRAPHQL_ADMIN_SECRET":              "nhost-admin-secret",
				"NHOST_ADMIN_SECRET":                       "nhost-admin-secret",
				"NHOST_BACKEND_URL":                        "http://traefik:1337",
				"NHOST_SUBDOMAIN":                          "localhost",
				"NHOST_REGION":                             "",
				"NHOST_HASURA_URL":                         "http://traefik:1337",
				"NHOST_GRAPHQL_URL":                        "http://traefik:1337/v1/graphql",
				"NHOST_AUTH_URL":                           "http://traefik:1337/v1/auth",
				"NHOST_STORAGE_URL":                        "http://traefik:1337/v1/storage",
				"NHOST_FUNCTIONS_URL":                      "http://traefik:1337/v1/functions",
				"HASURA_GRAPHQL_UNAUTHORIZED_ROLE":         "public",
				"HASURA_GRAPHQL_DEV_MODE":                  "true",
				"HASURA_GRAPHQL_LOG_LEVEL":                 "debug",
				"HASURA_GRAPHQL_ENABLE_CONSOLE":            "false",
				"HASURA_GRAPHQL_MIGRATIONS_SERVER_TIMEOUT": "20",
				"HASURA_GRAPHQL_NO_OF_RETRIES":             "20",
				"HASURA_GRAPHQL_ENABLE_TELEMETRY":          "false",
				"NHOST_WEBHOOK_SECRET":                     "nhost-webhook-secret",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			c := Config{
				conf:  tt.makeConf(),
				ports: defaultPorts(t),
			}
			a.Equal(tt.want, c.hasuraServiceEnvs())
		})
	}
}

func TestConfig_authServiceEnvs(t *testing.T) {
	t.Parallel()

	a := assert.New(t)

	tests := []struct {
		name       string
		makeConf   func() *config.Config
		assertFunc func(a *assert.Assertions, actual env)
	}{
		{
			name:     "with default config",
			makeConf: func() *config.Config { return defaultConfig(t) },
			assertFunc: func(a *assert.Assertions, actual env) {
				a.Equal(env{
					"HASURA_GRAPHQL_DATABASE_URL":               "postgres://nhost_auth_admin@postgres:5432/postgres",
					"AUTH_ACCESS_CONTROL_ALLOWED_REDIRECT_URLS": "",
					"AUTH_WEBAUTHN_ENABLED":                     "false",
					"AUTH_EMAIL_SIGNIN_EMAIL_VERIFIED_REQUIRED": "true",
					"AUTH_ACCESS_CONTROL_BLOCKED_EMAIL_DOMAINS": "",
					"AUTH_SMTP_PORT":                            "1025",
					"AUTH_PASSWORD_HIBP_ENABLED":                "false",
					"AUTH_SMS_PROVIDER":                         "",
					"NHOST_WEBHOOK_SECRET":                      "nhost-webhook-secret",
					"AUTH_CLIENT_URL":                           "http://localhost:3000",
					"AUTH_ANONYMOUS_USERS_ENABLED":              "false",
					"AUTH_SMTP_SENDER":                          "hasura-auth@example.com",
					"HASURA_GRAPHQL_GRAPHQL_URL":                "http://graphql:8080/v1/graphql",
					"AUTH_MFA_TOTP_ISSUER":                      "",
					"AUTH_SMTP_USER":                            "user",
					"AUTH_USER_DEFAULT_ALLOWED_ROLES":           "user,me",
					"AUTH_WEBAUTHN_ATTESTATION_TIMEOUT":         "60000",
					"AUTH_ACCESS_CONTROL_ALLOWED_EMAIL_DOMAINS": "",
					"AUTH_SMS_TWILIO_ACCOUNT_SID":               "",
					"AUTH_GRAVATAR_RATING":                      "g",
					"AUTH_SMTP_HOST":                            "mailhog",
					"HASURA_GRAPHQL_ADMIN_SECRET":               "nhost-admin-secret",
					"HASURA_GRAPHQL_JWT_SECRET":                 "{\"type\":\"HS256\", \"key\": \"0f987876650b4a085e64594fae9219e7781b17506bec02489ad061fba8cb22db\"}",
					"AUTH_HOST":                                 "0.0.0.0",
					"NHOST_ADMIN_SECRET":                        "nhost-admin-secret",
					"AUTH_EMAIL_PASSWORDLESS_ENABLED":           "false",
					"AUTH_SMS_TWILIO_AUTH_TOKEN":                "",
					"AUTH_WEBAUTHN_RP_ORIGINS":                  "",
					"AUTH_MFA_ENABLED":                          "false",
					"AUTH_GRAVATAR_DEFAULT":                     "blank",
					"AUTH_WEBAUTHN_RP_NAME":                     "",
					"AUTH_GRAVATAR_ENABLED":                     "true",
					"AUTH_ACCESS_CONTROL_ALLOWED_EMAILS":        "",
					"AUTH_JWT_CUSTOM_CLAIMS":                    "",
					"AUTH_ACCESS_CONTROL_BLOCKED_EMAILS":        "",
					"AUTH_REFRESH_TOKEN_EXPIRES_IN":             "43200",
					"AUTH_SMS_PASSWORDLESS_ENABLED":             "false",
					"AUTH_ACCESS_TOKEN_EXPIRES_IN":              "900",
					"AUTH_SMS_TWILIO_MESSAGING_SERVICE_ID":      "",
					"AUTH_LOCALE_DEFAULT":                       "en",
					"AUTH_CONCEAL_ERRORS":                       "",
					"AUTH_EMAIL_TEMPLATE_FETCH_URL":             "",
					"AUTH_USER_DEFAULT_ROLE":                    "user",
					"AUTH_SMTP_PASS":                            "password",
					"AUTH_SERVER_URL":                           "http://localhost:1337/v1/auth",
					"AUTH_DISABLE_NEW_USERS":                    "false",
					"AUTH_SMTP_SECURE":                          "false",
					"AUTH_PASSWORD_MIN_LENGTH":                  "9",
					"AUTH_LOCALE_ALLOWED_LOCALES":               "en",
				}, actual)
			},
		},
		{
			name: "when sms provider is set to non-twilio",
			makeConf: func() *config.Config {
				c := defaultConfig(t)
				c.Provider().Sms = &model.ConfigSms{
					Provider:           "foo",
					AccountSid:         "123",
					AuthToken:          "456",
					MessagingServiceId: "789",
				}
				return c
			},
			assertFunc: func(a *assert.Assertions, actual env) {
				a.Equal("", actual["AUTH_SMS_TWILIO_ACCOUNT_SID"])
				a.Equal("", actual["AUTH_SMS_TWILIO_AUTH_TOKEN"])
				a.Equal("", actual["AUTH_SMS_TWILIO_MESSAGING_SERVICE_ID"])
			},
		},
		{
			name: "when sms provider is set to twilio regardless of case",
			makeConf: func() *config.Config {
				c := defaultConfig(t)
				c.Provider().Sms = &model.ConfigSms{
					Provider:           "tWiLiO",
					AccountSid:         "123",
					AuthToken:          "456",
					MessagingServiceId: "789",
				}
				return c
			},
			assertFunc: func(a *assert.Assertions, actual env) {
				a.Equal("123", actual["AUTH_SMS_TWILIO_ACCOUNT_SID"])
				a.Equal("456", actual["AUTH_SMS_TWILIO_AUTH_TOKEN"])
				a.Equal("789", actual["AUTH_SMS_TWILIO_MESSAGING_SERVICE_ID"])
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			c := Config{
				conf:  tt.makeConf(),
				ports: defaultPorts(t),
			}
			tt.assertFunc(a, c.authServiceEnvs())
		})
	}
}

func TestConfig_postgresServiceEnvs(t *testing.T) {
	t.Parallel()

	a := assert.New(t)

	tests := []struct {
		name       string
		makeConf   func() *config.Config
		assertFunc func(a *assert.Assertions, actual env)
	}{
		{
			name:     "with default config",
			makeConf: func() *config.Config { return defaultConfig(t) },
			assertFunc: func(a *assert.Assertions, actual env) {
				a.Equal(env{
					"PGDATA":            "/var/lib/postgresql/data/pgdata",
					"POSTGRES_USER":     "postgres",
					"POSTGRES_PASSWORD": "postgres",
					"POSTGRES_DB":       "postgres",
				}, actual)
			},
		},
		{
			name: "with custom config (global envs)",
			makeConf: func() *config.Config {
				c := defaultConfig(t)
				c.Global().Environment = []*model.ConfigEnvironmentVariable{
					{Name: "PGDATA", Value: "/custom/folder"},
				}
				return c
			},
			assertFunc: func(a *assert.Assertions, actual env) {
				a.Equal("/custom/folder", actual["PGDATA"])
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			c := Config{
				conf:  tt.makeConf(),
				ports: defaultPorts(t),
			}
			tt.assertFunc(a, c.postgresServiceEnvs())
		})
	}
}

func TestConfig_mailhogServiceEnvs(t *testing.T) {
	t.Parallel()

	a := assert.New(t)

	tests := []struct {
		name       string
		makeConf   func() *config.Config
		assertFunc func(a *assert.Assertions, actual env)
	}{
		{
			name:     "with default config",
			makeConf: func() *config.Config { return defaultConfig(t) },
			assertFunc: func(a *assert.Assertions, actual env) {
				a.Equal(env{
					"SMTP_HOST":   "mailhog",
					"SMTP_PORT":   "1025",
					"SMTP_PASS":   "password",
					"SMTP_USER":   "user",
					"SMTP_SECURE": "false",
					"SMTP_SENDER": "hasura-auth@example.com",
				}, actual)
			},
		},
		{
			name: "with custom config (global envs)",
			makeConf: func() *config.Config {
				c := defaultConfig(t)
				c.Global().Environment = []*model.ConfigEnvironmentVariable{
					{Name: "SMTP_HOST", Value: "my-hostname"},
					{Name: "SMTP_USER", Value: "my-user"},
				}
				return c
			},
			assertFunc: func(a *assert.Assertions, actual env) {
				a.Equal("my-hostname", actual["SMTP_HOST"])
				a.Equal("my-user", actual["SMTP_USER"])
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			c := Config{
				conf:  tt.makeConf(),
				ports: defaultPorts(t),
			}
			tt.assertFunc(a, c.mailhogServiceEnvs())
		})
	}
}

func TestConfig_minioServiceEnvs(t *testing.T) {
	t.Parallel()

	a := assert.New(t)

	tests := []struct {
		name       string
		makeConf   func() *config.Config
		assertFunc func(a *assert.Assertions, actual env)
	}{
		{
			name:     "with default config",
			makeConf: func() *config.Config { return defaultConfig(t) },
			assertFunc: func(a *assert.Assertions, actual env) {
				a.Equal(env{
					"MINIO_ROOT_USER":     "minioaccesskey123123",
					"MINIO_ROOT_PASSWORD": "minioaccesskey123123",
				}, actual)
			},
		},
		{
			name: "with custom config (global envs)",
			makeConf: func() *config.Config {
				c := defaultConfig(t)
				c.Global().Environment = []*model.ConfigEnvironmentVariable{
					{Name: "MINIO_ROOT_PASSWORD", Value: "passwd"},
				}
				return c
			},
			assertFunc: func(a *assert.Assertions, actual env) {
				a.Equal("passwd", actual["MINIO_ROOT_PASSWORD"])
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			c := Config{
				conf:  tt.makeConf(),
				ports: defaultPorts(t),
			}
			tt.assertFunc(a, c.minioServiceEnvs())
		})
	}
}

func TestConfig_functionsServiceEnvs(t *testing.T) {
	t.Parallel()

	a := assert.New(t)

	tests := []struct {
		name       string
		makeConf   func() *config.Config
		assertFunc func(a *assert.Assertions, actual env)
	}{
		{
			name:     "with default config",
			makeConf: func() *config.Config { return defaultConfig(t) },
			assertFunc: func(a *assert.Assertions, actual env) {
				a.Equal(env{
					"NHOST_BACKEND_URL":    "http://traefik:1337",
					"NHOST_SUBDOMAIN":      "localhost",
					"NHOST_REGION":         "",
					"NHOST_HASURA_URL":     "http://traefik:1337",
					"NHOST_GRAPHQL_URL":    "http://traefik:1337/v1/graphql",
					"NHOST_AUTH_URL":       "http://traefik:1337/v1/auth",
					"NHOST_STORAGE_URL":    "http://traefik:1337/v1/storage",
					"NHOST_FUNCTIONS_URL":  "http://traefik:1337/v1/functions",
					"NHOST_ADMIN_SECRET":   "nhost-admin-secret",
					"NHOST_WEBHOOK_SECRET": "nhost-webhook-secret",
					"NHOST_JWT_SECRET":     "{\"type\":\"HS256\", \"key\": \"0f987876650b4a085e64594fae9219e7781b17506bec02489ad061fba8cb22db\"}",
				}, actual)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			c := Config{
				conf:  tt.makeConf(),
				ports: defaultPorts(t),
			}
			tt.assertFunc(a, c.functionsServiceEnvs())
		})
	}
}

func TestConfig_storageServiceEnvs(t *testing.T) {
	t.Parallel()

	a := assert.New(t)

	tests := []struct {
		name       string
		makeConf   func() *config.Config
		assertFunc func(a *assert.Assertions, actual env)
	}{
		{
			name:     "with default config",
			makeConf: func() *config.Config { return defaultConfig(t) },
			assertFunc: func(a *assert.Assertions, actual env) {
				a.Equal(env{
					"DEBUG":                       "true",
					"BIND":                        ":8576",
					"PUBLIC_URL":                  "http://localhost:1337",
					"API_ROOT_PREFIX":             "/v1/storage",
					"POSTGRES_MIGRATIONS":         "1",
					"HASURA_METADATA":             "1",
					"HASURA_ENDPOINT":             "http://graphql:8080/v1",
					"HASURA_GRAPHQL_ADMIN_SECRET": "nhost-admin-secret",
					"S3_ACCESS_KEY":               "minioaccesskey123123",
					"S3_SECRET_KEY":               "minioaccesskey123123",
					"S3_BUCKET":                   "nhost",
					"HASURA_GRAPHQL_JWT_SECRET":   "{\"type\":\"HS256\", \"key\": \"0f987876650b4a085e64594fae9219e7781b17506bec02489ad061fba8cb22db\"}",
					"NHOST_JWT_SECRET":            "{\"type\":\"HS256\", \"key\": \"0f987876650b4a085e64594fae9219e7781b17506bec02489ad061fba8cb22db\"}",
					"NHOST_ADMIN_SECRET":          "nhost-admin-secret",
					"NHOST_WEBHOOK_SECRET":        "nhost-webhook-secret",
					"POSTGRES_MIGRATIONS_SOURCE":  "postgres://nhost_storage_admin@postgres:5432/postgres?sslmode=disable",
				}, actual)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			c := Config{
				conf:  tt.makeConf(),
				ports: defaultPorts(t),
			}
			tt.assertFunc(a, c.storageServiceEnvs())
		})
	}
}

func TestConfig_graphqlJwtSecret(t *testing.T) {
	t.Parallel()

	a := assert.New(t)

	tests := []struct {
		name     string
		makeConf func() *config.Config
		want     string
	}{
		{
			name:     "with default config",
			makeConf: func() *config.Config { return defaultConfig(t) },
			want:     "{\"type\":\"HS256\", \"key\": \"0f987876650b4a085e64594fae9219e7781b17506bec02489ad061fba8cb22db\"}",
		},
		{
			name: "without jwt secrets",
			makeConf: func() *config.Config {
				conf := defaultConfig(t)
				conf.Hasura().JwtSecrets = nil
				return conf
			},
			want: "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			c := Config{
				conf:  tt.makeConf(),
				ports: defaultPorts(t),
			}
			a.Equal(tt.want, c.graphqlJwtSecret())
		})
	}
}

func TestConfig_twilioSettings(t *testing.T) {
	t.Parallel()

	a := assert.New(t)

	tests := []struct {
		name                   string
		makeConf               func() *config.Config
		wantAccountID          string
		wantAuthToken          string
		wantMessagingServiceID string
	}{
		{
			name:                   "with default config",
			makeConf:               func() *config.Config { return defaultConfig(t) },
			wantAccountID:          "",
			wantAuthToken:          "",
			wantMessagingServiceID: "",
		},
		{
			name: "when sms provider is not set",
			makeConf: func() *config.Config {
				c := defaultConfig(t)
				c.Provider().Sms = nil
				return c
			},
			wantAccountID:          "",
			wantAuthToken:          "",
			wantMessagingServiceID: "",
		},
		{
			name: "when sms provider is set to non-twilio",
			makeConf: func() *config.Config {
				c := defaultConfig(t)
				c.Provider().Sms = &model.ConfigSms{
					Provider:           "foo",
					AccountSid:         "123",
					AuthToken:          "456",
					MessagingServiceId: "789",
				}
				return c
			},
			wantAccountID:          "",
			wantAuthToken:          "",
			wantMessagingServiceID: "",
		},
		{
			name: "when sms provider is set to twilio regardless of case",
			makeConf: func() *config.Config {
				c := defaultConfig(t)
				c.Provider().Sms = &model.ConfigSms{
					Provider:           "tWiLiO",
					AccountSid:         "123",
					AuthToken:          "456",
					MessagingServiceId: "789",
				}
				return c
			},
			wantAccountID:          "123",
			wantAuthToken:          "456",
			wantMessagingServiceID: "789",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt := tt
			c := Config{
				conf:  tt.makeConf(),
				ports: defaultPorts(t),
			}
			accountID, authToken, messagingServiceID := c.twilioSettings()
			a.Equal(tt.wantAccountID, accountID)
			a.Equal(tt.wantAuthToken, authToken)
			a.Equal(tt.wantMessagingServiceID, messagingServiceID)
		})
	}
}
