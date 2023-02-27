package converter

import (
	"fmt"
	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/config"
	"github.com/nhost/cli/internal/generichelper"
	"github.com/nhost/cli/nhost"
	"github.com/nhost/cli/nhost/envvars"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

func Convert(logger logrus.FieldLogger, legacyConfig *nhost.Configuration) (*model.ConfigConfig, error) {
	newConfig, err := config.DefaultConfig()
	if err != nil {
		return nil, err
	}

	authConf := newConfig.Auth
	userConf := authConf.GetUser()
	methodConf := authConf.GetMethod()
	oauthConf := methodConf.GetOauth()
	providerConf := newConfig.Provider
	mfaConf := authConf.GetTotp()
	sessionConf := authConf.GetSession()

	{
		// Convert global environment variables
		globalConfigEnvs := collectLegacyServiceEnvs(legacyConfig)

		var globalEnvs []*model.ConfigEnvironmentVariable

		for key, value := range globalConfigEnvs {
			globalEnvs = append(globalEnvs, &model.ConfigEnvironmentVariable{
				Name:  key,
				Value: value,
			})
		}

		newConfig.Global.Environment = globalEnvs
	}

	// Fill out new config with legacy config values by checking all possible environment variables
	serviceConfigs := collectLegacyServiceConfigEnvs(legacyConfig)

	for key, value := range serviceConfigs {
		switch key {
		case "AUTH_ACCESS_CONTROL_EMAIL_ALLOWED_EMAILS", "AUTH_ACCESS_CONTROL_ALLOWED_EMAILS":
			if len(value) > 0 {
				emails := strings.Split(value, ",")
				for i, email := range emails {
					emails[i] = strings.TrimSpace(email)
				}
				userConf.Email.Allowed = emails
			}
		case "AUTH_ACCESS_CONTROL_EMAIL_ALLOWED_EMAIL_DOMAINS", "AUTH_ACCESS_CONTROL_ALLOWED_EMAIL_DOMAINS":
			if len(value) > 0 {
				domains := strings.Split(value, ",")
				for i, domain := range domains {
					domains[i] = strings.TrimSpace(domain)
				}
				userConf.EmailDomains.Allowed = domains
			}

		case "AUTH_ACCESS_CONTROL_EMAIL_BLOCKED_EMAILS", "AUTH_ACCESS_CONTROL_BLOCKED_EMAILS":
			if len(value) > 0 {
				emails := strings.Split(value, ",")
				for i, email := range emails {
					emails[i] = strings.TrimSpace(email)
				}
				userConf.Email.Blocked = emails
			}

		case "AUTH_ACCESS_CONTROL_EMAIL_BLOCKED_EMAIL_DOMAINS", "AUTH_ACCESS_CONTROL_BLOCKED_EMAIL_DOMAINS":
			if len(value) > 0 {
				domains := strings.Split(value, ",")
				for i, domain := range domains {
					domains[i] = strings.TrimSpace(domain)
				}
			}

		case "AUTH_ACCESS_CONTROL_URL_ALLOWED_REDIRECT_URLS", "AUTH_ACCESS_CONTROL_ALLOWED_REDIRECT_URLS":
			notSupportedEnv(logger, key)

		case "AUTH_ANONYMOUS_USERS_ENABLED":
			if value != "" {
				authConf.GetMethod().GetAnonymous().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_CLIENT_URL":
			if value != "" {
				authConf.GetRedirections().ClientUrl = generichelper.Pointerify(value)
			}

		case "AUTH_DISABLE_NEW_USERS":
			if value != "" {
				authConf.GetSignUp().Enabled = generichelper.Pointerify(strings.ToLower(value) != "true")
			}

		case "AUTH_EMAIL_ENABLED":
			notSupportedEnv(logger, key)

		case "AUTH_EMAIL_PASSWORDLESS_ENABLED":
			if value != "" {
				authConf.GetMethod().GetEmailPasswordless().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_EMAIL_SIGNIN_EMAIL_VERIFIED_REQUIRED":
			if value != "" {
				authConf.GetMethod().GetEmailPassword().EmailVerificationRequired = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_EMAIL_TEMPLATE_FETCH_URL":
			notSupportedEnv(logger, key)

		case "AUTH_GRAVATAR_DEFAULT":
			if value != "" {
				authConf.GetUser().GetGravatar().Default = generichelper.Pointerify(value)
			}

		case "AUTH_GRAVATAR_ENABLED":
			if value != "" {
				authConf.GetUser().GetGravatar().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_GRAVATAR_RATING":
			if value != "" {
				authConf.GetUser().GetGravatar().Rating = generichelper.Pointerify(value)
			}

		case "AUTH_LOCALE_ALLOWED", "AUTH_LOCALE_ALLOWED_LOCALES":
			if value != "" {
				locales := strings.Split(value, ",")
				for i, locale := range locales {
					locales[i] = strings.TrimSpace(locale)
				}
				authConf.GetUser().GetLocale().Allowed = locales
			}

		case "AUTH_LOCALE_DEFAULT":
			if value != "" {
				authConf.GetUser().GetLocale().Default = generichelper.Pointerify(value)
			}

		case "AUTH_PASSWORD_HIBP_ENABLED":
			if value != "" {
				authConf.GetMethod().GetEmailPassword().HibpEnabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_PASSWORD_MIN_LENGTH":
			if value != "" {
				// convert string to uint8

				uint8Val, err := strconv.ParseUint(value, 10, 8)
				if err != nil {
					return nil, fmt.Errorf("failed to convert %s to uint8 in %s env: %w", value, key, err)
				}

				authConf.GetMethod().GetEmailPassword().PasswordMinLength = generichelper.Pointerify(uint8(uint8Val))
			}

		case "AUTH_PROVIDER_APPLE_CLIENT_ID":
			if value != "" {
				oauthConf.GetApple().ClientId = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_APPLE_ENABLED":
			if value != "" {
				oauthConf.GetApple().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_PROVIDER_APPLE_KEY_ID":
			if value != "" {
				oauthConf.GetApple().KeyId = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_APPLE_PRIVATE_KEY":
			if value != "" {
				oauthConf.GetApple().PrivateKey = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_APPLE_SCOPE":
			if value != "" {
				scopes := strings.Split(value, ",")
				for i, scope := range scopes {
					scopes[i] = strings.TrimSpace(scope)
				}
				oauthConf.GetApple().Scope = scopes
			}

		case "AUTH_PROVIDER_APPLE_TEAM_ID":
			if value != "" {
				oauthConf.GetApple().TeamId = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_BITBUCKET_CLIENT_ID":
			if value != "" {
				oauthConf.GetBitbucket().ClientId = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_BITBUCKET_CLIENT_SECRET":
			if value != "" {
				oauthConf.GetBitbucket().ClientSecret = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_BITBUCKET_ENABLED":
			if value != "" {
				oauthConf.GetBitbucket().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_PROVIDER_FACEBOOK_CLIENT_ID":
			if value != "" {
				oauthConf.GetFacebook().ClientId = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_FACEBOOK_CLIENT_SECRET":
			if value != "" {
				oauthConf.GetFacebook().ClientSecret = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_FACEBOOK_ENABLED":
			if value != "" {
				oauthConf.GetFacebook().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_PROVIDER_FACEBOOK_SCOPE":
			if value != "" {
				scope := strings.Split(value, ",")
				for i, v := range scope {
					scope[i] = strings.TrimSpace(v)
				}
				oauthConf.GetFacebook().Scope = scope
			}

		case "AUTH_PROVIDER_GITHUB_CLIENT_ID":
			if value != "" {
				oauthConf.GetGithub().ClientId = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_GITHUB_CLIENT_SECRET":
			if value != "" {
				oauthConf.GetGithub().ClientSecret = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_GITHUB_ENABLED":
			if value != "" {
				oauthConf.GetGithub().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_PROVIDER_GITHUB_SCOPE":
			if value != "" {
				scope := strings.Split(value, ",")
				for i, v := range scope {
					scope[i] = strings.TrimSpace(v)
				}
				oauthConf.GetGithub().Scope = scope
			}

		case "AUTH_PROVIDER_GITHUB_TOKEN_URL":
			notSupportedEnv(logger, key)

		case "AUTH_PROVIDER_GITHUB_USER_PROFILE_URL":
			notSupportedEnv(logger, key)

		case "AUTH_PROVIDER_GITLAB_BASE_URL":
			notSupportedEnv(logger, key)

		case "AUTH_PROVIDER_GITLAB_CLIENT_ID":
			if value != "" {
				oauthConf.GetGitlab().ClientId = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_GITLAB_CLIENT_SECRET":
			if value != "" {
				oauthConf.GetGitlab().ClientSecret = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_GITLAB_ENABLED":
			if value != "" {
				oauthConf.GetGitlab().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_PROVIDER_GITLAB_SCOPE":
			if value != "" {
				scope := strings.Split(value, ",")
				for i, v := range scope {
					scope[i] = strings.TrimSpace(v)
				}
				oauthConf.GetGitlab().Scope = scope
			}

		case "AUTH_PROVIDER_GOOGLE_CLIENT_ID":
			if value != "" {
				oauthConf.GetGoogle().ClientId = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_GOOGLE_CLIENT_SECRET":
			if value != "" {
				oauthConf.GetGoogle().ClientSecret = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_GOOGLE_ENABLED":
			if value != "" {
				oauthConf.GetGoogle().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_PROVIDER_GOOGLE_SCOPE":
			if value != "" {
				scope := strings.Split(value, ",")
				for i, v := range scope {
					scope[i] = strings.TrimSpace(v)
				}
				oauthConf.GetGoogle().Scope = scope
			}

		case "AUTH_PROVIDER_LINKEDIN_CLIENT_ID":
			if value != "" {
				oauthConf.GetLinkedin().ClientId = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_LINKEDIN_CLIENT_SECRET":
			if value != "" {
				oauthConf.GetLinkedin().ClientSecret = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_LINKEDIN_ENABLED":
			if value != "" {
				oauthConf.GetLinkedin().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_PROVIDER_LINKEDIN_SCOPE":
			if value != "" {
				scope := strings.Split(value, ",")
				for i, v := range scope {
					scope[i] = strings.TrimSpace(v)
				}
				oauthConf.GetLinkedin().Scope = scope
			}

		case "AUTH_PROVIDER_SPOTIFY_CLIENT_ID":
			if value != "" {
				oauthConf.GetSpotify().ClientId = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_SPOTIFY_CLIENT_SECRET":
			if value != "" {
				oauthConf.GetSpotify().ClientSecret = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_SPOTIFY_ENABLED":
			if value != "" {
				oauthConf.GetSpotify().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_PROVIDER_SPOTIFY_SCOPE":
			if value != "" {
				scope := strings.Split(value, ",")
				for i, v := range scope {
					scope[i] = strings.TrimSpace(v)
				}
				oauthConf.GetSpotify().Scope = scope
			}

		case "AUTH_PROVIDER_STRAVA_CLIENT_ID":
			if value != "" {
				oauthConf.GetStrava().ClientId = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_STRAVA_CLIENT_SECRET":
			if value != "" {
				oauthConf.GetStrava().ClientSecret = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_STRAVA_ENABLED":
			if value != "" {
				oauthConf.GetStrava().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_PROVIDER_TWILIO_ACCOUNT_SID", "AUTH_SMS_PROVIDER_TWILIO_ACCOUNT_SID":
			if value != "" {
				providerConf.GetSms().AccountSid = value
			}

		case "AUTH_PROVIDER_TWILIO_AUTH_TOKEN", "AUTH_SMS_PROVIDER_TWILIO_AUTH_TOKEN":
			if value != "" {
				providerConf.GetSms().AuthToken = value
			}

		case "AUTH_PROVIDER_TWILIO_ENABLED":
			notSupportedEnv(logger, key)

		case "AUTH_PROVIDER_TWILIO_MESSAGING_SERVICE_ID", "AUTH_SMS_TWILIO_MESSAGING_SERVICE_ID", "AUTH_SMS_PROVIDER_TWILIO_MESSAGING_SERVICE_ID":
			if value != "" {
				providerConf.GetSms().MessagingServiceId = value
			}

		case "AUTH_PROVIDER_TWITTER_CONSUMER_KEY":
			if value != "" {
				oauthConf.GetTwitter().ConsumerKey = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_TWITTER_CONSUMER_SECRET":
			if value != "" {
				oauthConf.GetTwitter().ConsumerSecret = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_TWITTER_ENABLED":
			if value != "" {
				oauthConf.GetTwitter().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_PROVIDER_WINDOWS_LIVE_CLIENT_ID":
			if value != "" {
				oauthConf.GetWindowslive().ClientId = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_WINDOWS_LIVE_CLIENT_SECRET":
			if value != "" {
				oauthConf.GetWindowslive().ClientSecret = generichelper.Pointerify(value)
			}

		case "AUTH_PROVIDER_WINDOWS_LIVE_ENABLED":
			if value != "" {
				oauthConf.GetWindowslive().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_PROVIDER_WINDOWS_LIVE_SCOPE":
			if value != "" {
				scope := strings.Split(value, ",")
				for i, v := range scope {
					scope[i] = strings.TrimSpace(v)
				}
				oauthConf.GetWindowslive().Scope = scope
			}

		case "AUTH_SMS_ENABLED":
			notSupportedEnv(logger, key)

		case "AUTH_SMS_PASSWORDLESS_ENABLED":
			if value != "" {
				methodConf.GetEmailPasswordless().Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_SMS_PROVIDER_TWILIO_FROM":
			notSupportedEnv(logger, key)

		case "AUTH_SMTP_HOST":
			if value != "" {
				providerConf.GetSmtp().Host = value
			}

		case "AUTH_SMTP_METHOD":
			if value != "" {
				providerConf.GetSmtp().Method = value
			}

		case "AUTH_SMTP_PASS":
			if value != "" {
				providerConf.GetSmtp().Password = value
			}

		case "AUTH_SMTP_PORT":
			if value != "" {
				// parse value to uint16
				uint16Val, err := strconv.ParseUint(value, 10, 16)
				if err != nil {
					return nil, fmt.Errorf("failed to convert %s to uint16 in %s env: %w", value, key, err)
				}

				providerConf.GetSmtp().Port = uint16(uint16Val)
			}

		case "AUTH_SMTP_SECURE":
			if value != "" {
				providerConf.GetSmtp().Secure = strings.ToLower(value) == "true"
			}

		case "AUTH_SMTP_SENDER":
			if value != "" {
				providerConf.GetSmtp().Sender = value
			}

		case "AUTH_SMTP_USER":
			if value != "" {
				providerConf.GetSmtp().User = value
			}

		case "AUTH_TOKEN_ACCESS_EXPIRES_IN", "AUTH_ACCESS_TOKEN_EXPIRES_IN":
			if value != "" {
				uint32Val, err := strconv.ParseUint(value, 10, 32)
				if err != nil {
					return nil, fmt.Errorf("failed to convert %s to uint32 in %s env: %w", value, key, err)
				}

				sessionConf.GetAccessToken().ExpiresIn = generichelper.Pointerify(uint32(uint32Val))
			}

		case "AUTH_TOKEN_REFRESH_EXPIRES_IN":
			if value != "" {
				uint32Val, err := strconv.ParseUint(value, 10, 32)
				if err != nil {
					return nil, fmt.Errorf("failed to convert %s to uint32 in %s env: %w", value, key, err)
				}

				sessionConf.GetRefreshToken().ExpiresIn = generichelper.Pointerify(uint32(uint32Val))
			}

		case "AUTH_USER_ALLOWED_ROLES":
			notSupportedEnv(logger, key)

		case "AUTH_USER_DEFAULT_ALLOWED_ROLES":
			if value != "" {
				allowedRoles := strings.Split(value, ",")
				for i, v := range allowedRoles {
					allowedRoles[i] = strings.TrimSpace(v)
				}
				userConf.GetRoles().Allowed = allowedRoles
			}

		case "AUTH_USER_DEFAULT_ROLE":
			if value != "" {
				userConf.GetRoles().Default = generichelper.Pointerify(value)
			}

		case "AUTH_USER_MFA_ENABLED", "AUTH_MFA_ENABLED":
			if value != "" {
				mfaConf.Enabled = generichelper.Pointerify(strings.ToLower(value) == "true")
			}

		case "AUTH_USER_MFA_ISSUER", "AUTH_MFA_TOTP_ISSUER":
			if value != "" {
				mfaConf.Issuer = generichelper.Pointerify(value)
			}

		case "HASURA_GRAPHQL_ENABLE_REMOTE_SCHEMA_PERMISSIONS":
			notSupportedEnv(logger, key)

		case "POSTGRES_PASSWORD":
			notSupportedEnv(logger, key)

		case "POSTGRES_USER":
			notSupportedEnv(logger, key)

		case "STORAGE_FORCE_DOWNLOAD_FOR_CONTENT_TYPES":
			notSupportedEnv(logger, key)
		default:
			logger.Infof("[NOTICE] unknown environment variable '%s'='%s'", key, value)
		}
	}

	return newConfig, nil
}

func collectLegacyServiceEnvs(legacyConfig *nhost.Configuration) envvars.Env {
	env := envvars.New()

	// postgres
	if postgresConf := legacyConfig.Services["postgres"]; postgresConf != nil {
		env.Merge(collectLegacyConfigServiceEnvs(postgresConf.Environment))
	}

	// hasura
	if hasuraConf := legacyConfig.Services["hasura"]; hasuraConf != nil {
		env.Merge(collectLegacyConfigServiceEnvs(hasuraConf.Environment))
	}

	// auth
	if authSvcConf := legacyConfig.Services["auth"]; authSvcConf != nil {
		env.Merge(collectLegacyConfigServiceEnvs(authSvcConf.Environment))
	}

	// storage
	if storageSvcConf := legacyConfig.Services["storage"]; storageSvcConf != nil {
		env.Merge(collectLegacyConfigServiceEnvs(storageSvcConf.Environment))
	}

	return env
}

func collectLegacyServiceConfigEnvs(legacyConfig *nhost.Configuration) envvars.Env {
	env := envvars.New()

	env.Merge(flattenEnvs(legacyConfig.Auth, "AUTH"))
	env.Merge(flattenEnvs(legacyConfig.Storage, "STORAGE"))

	return env
}

func notSupportedEnv(logger logrus.FieldLogger, name string) {
	logger.Infof("[NOTICE] environment variable '%s' can't be converted", name)
}

func collectLegacyConfigServiceEnvs(legacyEnv map[string]any) envvars.Env {
	env := envvars.New()

	if legacyEnv == nil {
		return env
	}

	for k, v := range legacyEnv {
		env[strings.ToUpper(k)] = fmt.Sprintf("%v", v)
	}

	return env
}
