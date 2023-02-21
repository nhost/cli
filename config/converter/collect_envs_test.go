package converter

import (
	"github.com/nhost/cli/nhost"
	"github.com/nhost/cli/nhost/envvars"
	"github.com/stretchr/testify/assert"
	"testing"
)

func legacyDefaultConfig(t *testing.T) nhost.Configuration {
	t.Helper()
	return nhost.GenerateConfig(nhost.App{})
}

func Test_collectLegacyConfigServiceEnv(t *testing.T) {
	assert := assert.New(t)
	legacyConfig := legacyDefaultConfig(t)
	assert.Equal(envvars.Env{
		"HASURA_GRAPHQL_ENABLE_REMOTE_SCHEMA_PERMISSIONS": "true",
	}, collectLegacyConfigServiceEnv(legacyConfig.Services["hasura"].Environment))
}

func Test_flattenEnvs(t *testing.T) {
	legacyConfig := legacyDefaultConfig(t)

	type args struct {
		envs   map[any]any
		prefix string
	}
	tests := []struct {
		name string
		args args
		want envvars.Env
	}{
		{
			name: "auth envs",
			args: args{
				envs:   legacyConfig.Auth,
				prefix: "AUTH",
			},
			want: envvars.Env{
				"AUTH_ACCESS_CONTROL_EMAIL_ALLOWED_EMAILS":        "",
				"AUTH_ACCESS_CONTROL_EMAIL_ALLOWED_EMAIL_DOMAINS": "",
				"AUTH_ACCESS_CONTROL_EMAIL_BLOCKED_EMAILS":        "",
				"AUTH_ACCESS_CONTROL_EMAIL_BLOCKED_EMAIL_DOMAINS": "",
				"AUTH_ACCESS_CONTROL_URL_ALLOWED_REDIRECT_URLS":   "",
				"AUTH_ANONYMOUS_USERS_ENABLED":                    "false",
				"AUTH_CLIENT_URL":                                 "http://localhost:3000",
				"AUTH_DISABLE_NEW_USERS":                          "false",
				"AUTH_EMAIL_ENABLED":                              "false",
				"AUTH_EMAIL_PASSWORDLESS_ENABLED":                 "false",
				"AUTH_EMAIL_SIGNIN_EMAIL_VERIFIED_REQUIRED":       "true",
				"AUTH_EMAIL_TEMPLATE_FETCH_URL":                   "",
				"AUTH_GRAVATAR_DEFAULT":                           "",
				"AUTH_GRAVATAR_ENABLED":                           "true",
				"AUTH_GRAVATAR_RATING":                            "",
				"AUTH_LOCALE_ALLOWED":                             "en",
				"AUTH_LOCALE_DEFAULT":                             "en",
				"AUTH_PASSWORD_HIBP_ENABLED":                      "false",
				"AUTH_PASSWORD_MIN_LENGTH":                        "3",
				"AUTH_PROVIDER_APPLE_CLIENT_ID":                   "",
				"AUTH_PROVIDER_APPLE_ENABLED":                     "false",
				"AUTH_PROVIDER_APPLE_KEY_ID":                      "",
				"AUTH_PROVIDER_APPLE_PRIVATE_KEY":                 "",
				"AUTH_PROVIDER_APPLE_SCOPE":                       "name,email",
				"AUTH_PROVIDER_APPLE_TEAM_ID":                     "",
				"AUTH_PROVIDER_BITBUCKET_CLIENT_ID":               "",
				"AUTH_PROVIDER_BITBUCKET_CLIENT_SECRET":           "",
				"AUTH_PROVIDER_BITBUCKET_ENABLED":                 "false",
				"AUTH_PROVIDER_FACEBOOK_CLIENT_ID":                "",
				"AUTH_PROVIDER_FACEBOOK_CLIENT_SECRET":            "",
				"AUTH_PROVIDER_FACEBOOK_ENABLED":                  "false",
				"AUTH_PROVIDER_FACEBOOK_SCOPE":                    "email,photos,displayName",
				"AUTH_PROVIDER_GITHUB_CLIENT_ID":                  "",
				"AUTH_PROVIDER_GITHUB_CLIENT_SECRET":              "",
				"AUTH_PROVIDER_GITHUB_ENABLED":                    "false",
				"AUTH_PROVIDER_GITHUB_SCOPE":                      "user:email",
				"AUTH_PROVIDER_GITHUB_TOKEN_URL":                  "",
				"AUTH_PROVIDER_GITHUB_USER_PROFILE_URL":           "",
				"AUTH_PROVIDER_GITLAB_BASE_URL":                   "",
				"AUTH_PROVIDER_GITLAB_CLIENT_ID":                  "",
				"AUTH_PROVIDER_GITLAB_CLIENT_SECRET":              "",
				"AUTH_PROVIDER_GITLAB_ENABLED":                    "false",
				"AUTH_PROVIDER_GITLAB_SCOPE":                      "read_user",
				"AUTH_PROVIDER_GOOGLE_CLIENT_ID":                  "",
				"AUTH_PROVIDER_GOOGLE_CLIENT_SECRET":              "",
				"AUTH_PROVIDER_GOOGLE_ENABLED":                    "false",
				"AUTH_PROVIDER_GOOGLE_SCOPE":                      "email,profile",
				"AUTH_PROVIDER_LINKEDIN_CLIENT_ID":                "",
				"AUTH_PROVIDER_LINKEDIN_CLIENT_SECRET":            "",
				"AUTH_PROVIDER_LINKEDIN_ENABLED":                  "false",
				"AUTH_PROVIDER_LINKEDIN_SCOPE":                    "r_emailaddress,r_liteprofile",
				"AUTH_PROVIDER_SPOTIFY_CLIENT_ID":                 "",
				"AUTH_PROVIDER_SPOTIFY_CLIENT_SECRET":             "",
				"AUTH_PROVIDER_SPOTIFY_ENABLED":                   "false",
				"AUTH_PROVIDER_SPOTIFY_SCOPE":                     "user-read-email,user-read-private",
				"AUTH_PROVIDER_STRAVA_CLIENT_ID":                  "",
				"AUTH_PROVIDER_STRAVA_CLIENT_SECRET":              "",
				"AUTH_PROVIDER_STRAVA_ENABLED":                    "false",
				"AUTH_PROVIDER_TWILIO_ACCOUNT_SID":                "",
				"AUTH_PROVIDER_TWILIO_AUTH_TOKEN":                 "",
				"AUTH_PROVIDER_TWILIO_ENABLED":                    "false",
				"AUTH_PROVIDER_TWILIO_MESSAGING_SERVICE_ID":       "",
				"AUTH_PROVIDER_TWITTER_CONSUMER_KEY":              "",
				"AUTH_PROVIDER_TWITTER_CONSUMER_SECRET":           "",
				"AUTH_PROVIDER_TWITTER_ENABLED":                   "false",
				"AUTH_PROVIDER_WINDOWS_LIVE_CLIENT_ID":            "",
				"AUTH_PROVIDER_WINDOWS_LIVE_CLIENT_SECRET":        "",
				"AUTH_PROVIDER_WINDOWS_LIVE_ENABLED":              "false",
				"AUTH_PROVIDER_WINDOWS_LIVE_SCOPE":                "wl.basic,wl.emails,wl.contacts_emails",
				"AUTH_SMS_ENABLED":                                "false",
				"AUTH_SMS_PASSWORDLESS_ENABLED":                   "false",
				"AUTH_SMS_PROVIDER_TWILIO_FROM":                   "",
				"AUTH_SMS_PROVIDER_TWILIO_ACCOUNT_SID":            "",
				"AUTH_SMS_PROVIDER_TWILIO_AUTH_TOKEN":             "",
				"AUTH_SMS_PROVIDER_TWILIO_MESSAGING_SERVICE_ID":   "",
				"AUTH_SMTP_HOST":                                  "mailhog",
				"AUTH_SMTP_PORT":                                  "1025",
				"AUTH_SMTP_METHOD":                                "",
				"AUTH_SMTP_PASS":                                  "password",
				"AUTH_SMTP_SECURE":                                "false",
				"AUTH_SMTP_SENDER":                                "hasura-auth@example.com",
				"AUTH_SMTP_USER":                                  "user",
				"AUTH_TOKEN_ACCESS_EXPIRES_IN":                    "900",
				"AUTH_TOKEN_REFRESH_EXPIRES_IN":                   "43200",
				"AUTH_USER_ALLOWED_ROLES":                         "user,me",
				"AUTH_USER_DEFAULT_ALLOWED_ROLES":                 "user,me",
				"AUTH_USER_DEFAULT_ROLE":                          "user",
				"AUTH_USER_MFA_ENABLED":                           "false",
				"AUTH_USER_MFA_ISSUER":                            "nhost",
			},
		},
		{
			name: "storage envs",
			args: args{
				envs:   legacyConfig.Storage,
				prefix: "STORAGE",
			},
			want: envvars.Env{
				"STORAGE_FORCE_DOWNLOAD_FOR_CONTENT_TYPES": "text/html,application/javascript",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, flattenEnvs(tt.args.envs, tt.args.prefix), "flattenEnvs(%v, %v)", tt.args.envs, tt.args.prefix)
		})
	}
}
