package dockercompose //nolint:testpackage

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nhost/be/services/mimir/model"
)

func expectedAuth() *Service {
	//nolint:lll
	return &Service{
		Image:   "nhost/hasura-auth:0.13.1",
		Command: nil,
		DependsOn: map[string]DependsOn{
			"graphql":  {Condition: "service_healthy"},
			"postgres": {Condition: "service_healthy"},
		},
		EntryPoint: nil,
		Environment: map[string]string{
			"AUTH_ACCESS_CONTROL_ALLOWED_EMAILS":        "asd@asd.com",
			"AUTH_ACCESS_CONTROL_ALLOWED_EMAIL_DOMAINS": "asd.com",
			"AUTH_ACCESS_CONTROL_ALLOWED_REDIRECT_URLS": "http://localhost:3000",
			"AUTH_ACCESS_CONTROL_BLOCKED_EMAILS":        "qwe@wqe.com",
			"AUTH_ACCESS_CONTROL_BLOCKED_EMAIL_DOMAINS": "qwe.com",
			"AUTH_ACCESS_TOKEN_EXPIRES_IN":              "900",
			"AUTH_ANONYMOUS_USERS_ENABLED":              "true",
			"AUTH_CLIENT_URL":                           "http://localhost:3000",
			"AUTH_DISABLE_NEW_USERS":                    "false",
			"AUTH_EMAIL_PASSWORDLESS_ENABLED":           "true",
			"AUTH_EMAIL_SIGNIN_EMAIL_VERIFIED_REQUIRED": "true",
			"AUTH_GRAVATAR_DEFAULT":                     "gravatarDefault",
			"AUTH_GRAVATAR_ENABLED":                     "true",
			"AUTH_GRAVATAR_RATING":                      "gravatarRating",
			"AUTH_HOST":                                 "0.0.0.0",
			"AUTH_JWT_CUSTOM_CLAIMS":                    `{"customClaimKey":"customClaimValue"}`,
			"AUTH_LOCALE_ALLOWED_LOCALES":               "en,se,ca,es",
			"AUTH_LOCALE_DEFAULT":                       "en",
			"AUTH_MFA_ENABLED":                          "true",
			"AUTH_MFA_TOTP_ISSUER":                      "totpIssuer",
			"AUTH_PASSWORD_HIBP_ENABLED":                "true",
			"AUTH_PASSWORD_MIN_LENGTH":                  "12",
			"AUTH_PORT":                                 "4000",
			"AUTH_PROVIDER_APPLE_CLIENT_ID":             "appleClientId",
			"AUTH_PROVIDER_APPLE_ENABLED":               "true",
			"AUTH_PROVIDER_APPLE_KEY_ID":                "appleKeyId",
			"AUTH_PROVIDER_APPLE_PRIVATE_KEY":           "applePrivateKey",
			"AUTH_PROVIDER_APPLE_TEAM_ID":               "appleTeamId",
			"AUTH_PROVIDER_AZUREAD_CLIENT_ID":           "azureadClientId",
			"AUTH_PROVIDER_AZUREAD_CLIENT_SECRET":       "azureadClientSecret",
			"AUTH_PROVIDER_AZUREAD_ENABLED":             "true",
			"AUTH_PROVIDER_AZUREAD_TENANT":              "azureadTenant",
			"AUTH_PROVIDER_DISCORD_CLIENT_ID":           "discordClientId",
			"AUTH_PROVIDER_DISCORD_CLIENT_SECRET":       "discordClientSecret",
			"AUTH_PROVIDER_DISCORD_ENABLED":             "true",
			"AUTH_PROVIDER_DISCORD_SCOPE":               "identify,email",
			"AUTH_PROVIDER_FACEBOOK_CLIENT_ID":          "facebookClientId",
			"AUTH_PROVIDER_FACEBOOK_CLIENT_SECRET":      "facebookClientSecret",
			"AUTH_PROVIDER_FACEBOOK_ENABLED":            "true",
			"AUTH_PROVIDER_FACEBOOK_SCOPE":              "email",
			"AUTH_PROVIDER_GITHUB_CLIENT_ID":            "githubClientId",
			"AUTH_PROVIDER_GITHUB_CLIENT_SECRET":        "githubClientSecret",
			"AUTH_PROVIDER_GITHUB_ENABLED":              "true",
			"AUTH_PROVIDER_GOOGLE_CLIENT_ID":            "googleClientId",
			"AUTH_PROVIDER_GOOGLE_CLIENT_SECRET":        "googleClientSecret",
			"AUTH_PROVIDER_GOOGLE_ENABLED":              "true",
			"AUTH_PROVIDER_GOOGLE_SCOPE":                "openid,profile,email",
			"AUTH_PROVIDER_LINKEDIN_CLIENT_ID":          "linkedinClientId",
			"AUTH_PROVIDER_LINKEDIN_CLIENT_SECRET":      "linkedinClientSecret",
			"AUTH_PROVIDER_LINKEDIN_ENABLED":            "true",
			"AUTH_PROVIDER_LINKEDIN_SCOPE":              "r_liteprofile,r_emailaddress",
			"AUTH_PROVIDER_SPOTIFY_CLIENT_ID":           "spotifyClientId",
			"AUTH_PROVIDER_SPOTIFY_CLIENT_SECRET":       "spotifyClientSecret",
			"AUTH_PROVIDER_SPOTIFY_ENABLED":             "true",
			"AUTH_PROVIDER_SPOTIFY_SCOPE":               "user-read-email",
			"AUTH_PROVIDER_TWITCH_CLIENT_ID":            "twitchClientId",
			"AUTH_PROVIDER_TWITCH_CLIENT_SECRET":        "twitchClientSecret",
			"AUTH_PROVIDER_TWITCH_ENABLED":              "true",
			"AUTH_PROVIDER_TWITCH_SCOPE":                "user:email",
			"AUTH_PROVIDER_TWITTER_CONSUMER_KEY":        "twitterConsumerKey",
			"AUTH_PROVIDER_TWITTER_CONSUMER_SECRET":     "twitterConsumerSecret",
			"AUTH_PROVIDER_TWITTER_ENABLED":             "true",
			"AUTH_PROVIDER_WINDOWS_LIVE_CLIENT_ID":      "windowsliveClientId",
			"AUTH_PROVIDER_WINDOWS_LIVE_CLIENT_SECRET":  "windowsliveClientSecret",
			"AUTH_PROVIDER_WINDOWS_LIVE_ENABLED":        "true",
			"AUTH_PROVIDER_WINDOWS_LIVE_SCOPE":          "wl.emails",
			"AUTH_PROVIDER_WORKOS_CLIENT_ID":            "workosClientId",
			"AUTH_PROVIDER_WORKOS_CLIENT_SECRET":        "workosClientSecret",
			"AUTH_PROVIDER_WORKOS_DEFAULT_CONNECTION":   "workosConnection",
			"AUTH_PROVIDER_WORKOS_DEFAULT_ORGANIZATION": "workosOrganization",
			"AUTH_PROVIDER_WORKOS_ENABLED":              "true",
			"AUTH_REFRESH_TOKEN_EXPIRES_IN":             "99",
			"AUTH_SERVER_URL":                           "http://local.auth.nhost.run:1336/v1",
			"AUTH_SMS_PASSWORDLESS_ENABLED":             "true",
			"AUTH_SMS_PROVIDER":                         "twilio",
			"AUTH_SMS_TWILIO_ACCOUNT_SID":               "smsAccountSid",
			"AUTH_SMS_TWILIO_AUTH_TOKEN":                "smsAuthToken",
			"AUTH_SMS_TWILIO_MESSAGING_SERVICE_ID":      "smsMessagingServiceId",
			"AUTH_SMTP_AUTH_METHOD":                     "LOGIN",
			"AUTH_SMTP_HOST":                            "mailhog",
			"AUTH_SMTP_PASS":                            "password",
			"AUTH_SMTP_PORT":                            "1025",
			"AUTH_SMTP_SECURE":                          "false",
			"AUTH_SMTP_SENDER":                          "hasura-auth@example.com",
			"AUTH_SMTP_USER":                            "user",
			"AUTH_USER_DEFAULT_ALLOWED_ROLES":           "user,admin",
			"AUTH_USER_DEFAULT_ROLE":                    "user",
			"AUTH_WEBAUTHN_ATTESTATION_TIMEOUT":         "60000",
			"AUTH_WEBAUTHN_ENABLED":                     "true",
			"AUTH_WEBAUTHN_RP_NAME":                     "webauthnRelyingPartyName",
			"AUTH_WEBAUTHN_RP_ORIGINS":                  "http://localhost:3000",
			"ENV1":                                      "VALUE1",
			"ENV2":                                      "VALUE2",
			"HASURA_GRAPHQL_ADMIN_SECRET":               "adminSecret",
			"HASURA_GRAPHQL_DATABASE_URL":               "postgres://nhost_auth_admin@postgres:5432/local",
			"HASURA_GRAPHQL_GRAPHQL_URL":                "http://graphql:8080/v1/graphql",
			"HASURA_GRAPHQL_JWT_SECRET":                 `{"claims_map":{"x-hasura-allowed-roles":{"path":"$.roles"},"x-hasura-default-role":"viewer","x-hasura-org-id":{"default":"public","path":"$.org"},"x-hasura-user-id":{"path":"$.sub"}},"key":"jwtSecretKey","type":"HS256"}`,
		},
		ExtraHosts: []string{
			"host.docker.internal:host-gateway", "local.auth.nhost.run:host-gateway",
			"local.db.nhost.run:host-gateway", "local.functions.nhost.run:host-gateway",
			"local.graphql.nhost.run:host-gateway", "local.hasura.nhost.run:host-gateway",
			"local.storage.nhost.run:host-gateway",
		},
		HealthCheck: &HealthCheck{
			Test:        []string{"CMD", "wget", "--spider", "-S", "http://localhost:4000/healthz"},
			Interval:    "5s",
			StartPeriod: "60s",
		},
		Labels: map[string]string{
			"traefik.enable": "true",
			"traefik.http.middlewares.replace-auth.replacepathregex.regex":       "/v1(/|$)(.*)",
			"traefik.http.middlewares.replace-auth.replacepathregex.replacement": "/$2",
			"traefik.http.routers.auth.entrypoints":                              "web",
			"traefik.http.routers.auth.middlewares":                              "replace-auth",
			"traefik.http.routers.auth.rule":                                     "Host(`local.auth.nhost.run`) && PathPrefix(`/v1`)",
			"traefik.http.routers.auth.service":                                  "auth",
			"traefik.http.routers.auth.tls":                                      "false",
			"traefik.http.services.auth.loadbalancer.server.port":                "4000",
		},
		Ports:   []Port{},
		Restart: "always",
		Volumes: []Volume{
			{Type: "bind", Source: "/tmp/nhost/emails", Target: "/app/email-templates"},
		},
		WorkingDir: nil,
	}
}

func TestAuth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		cfg      func() *model.ConfigConfig
		useTlS   bool
		expected func() *Service
	}{
		{
			name:     "success",
			cfg:      getConfig,
			expected: expectedAuth,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			got, err := auth(tc.cfg(), 1336, tc.useTlS, "/tmp/nhost")
			if err != nil {
				t.Errorf("got error: %v", err)
			}

			if diff := cmp.Diff(tc.expected(), got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
