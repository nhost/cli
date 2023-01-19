package compose

import (
	"fmt"
	"github.com/nhost/cli/config"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"strings"
	"time"

	"github.com/nhost/cli/internal/ports"

	"github.com/compose-spec/compose-go/types"
	"github.com/nhost/cli/nhost"
)

const (
	// hosts
	hostDockerInternal = "host.docker.internal"
	hostGateway        = "host-gateway"

	// docker compose service names
	SvcPostgres  = "postgres"
	SvcAuth      = "auth"
	SvcStorage   = "storage"
	SvcFunctions = "functions"
	SvcMinio     = "minio"
	SvcMailhog   = "mailhog"
	SvcHasura    = "hasura"
	SvcTraefik   = "traefik"
	SvcGraphql   = "graphql"
	SvcDashboard = "dashboard"
	// --

	// container ports
	graphqlPort = 8080
	dbPort      = 5432
	proxyPort   = 1337
	// --

	// default docker images
	svcDashboardDefaultImage = "nhost/dashboard:0.9.9"
	svcPostgresDefaultImage  = "nhost/postgres:14.5-20230104-1"
	svcAuthDefaultImage      = "nhost/hasura-auth:0.17.0"
	svcStorageDefaultImage   = "nhost/hasura-storage:0.3.0"
	svcFunctionsDefaultImage = "nhost/functions:0.1.8"
	svcMinioDefaultImage     = "minio/minio:RELEASE.2022-07-08T00-05-23Z"
	svcMailhogDefaultImage   = "mailhog/mailhog"
	svcHasuraDefaultImage    = "hasura/graphql-engine:v2.15.2"
	svcTraefikDefaultImage   = "traefik:v2.8"
	// --

	// volume names
	volFunctionsNodeModules = "functions_node_modules"
	volRootNodeModules      = "root_node_modules"
	// --

	// environment variables

	// dashboard
	// envDashboardNextPublicNhost
	envDashboardNextPublicNhostLocalBackendPort = "NEXT_PUBLIC_NHOST_LOCAL_BACKEND_PORT"
	envDashboardNextPublicNhostHasuraPort       = "NEXT_PUBLIC_NHOST_HASURA_PORT"
	envDashboardNextPublicNhostMigrationsPort   = "NEXT_PUBLIC_NHOST_MIGRATIONS_PORT"
	envDashboardNextPublicNhostPlatform         = "NEXT_PUBLIC_NHOST_PLATFORM"
	envDashboardNextPublicEnv                   = "NEXT_PUBLIC_ENV"
	envDashboardNextTelemetryDisabled           = "NEXT_TELEMETRY_DISABLED"

	// dashboard envs values
	envDashboardNextPublicNhostPlatformValue = "false"
	envDashboardNextPublicEnvValue           = "dev"
	envDashboardNextTelemetryDisabledValue   = "1"

	// minio
	envMinioRootUser     = "MINIO_ROOT_USER"
	envMinioRootPassword = "MINIO_ROOT_PASSWORD"

	// postgres
	envPostgresPassword = "POSTGRES_PASSWORD"
	envPostgresDb       = "POSTGRES_DB"
	envPostgresUser     = "POSTGRES_USER"
	envPostgresData     = "PGDATA"

	// default values for environment variables
	envPostgresDbDefaultValue       = "postgres"
	envPostgresUserDefaultValue     = "postgres"
	envPostgresPasswordDefaultValue = "postgres"
	envPostgresDataDefaultValue     = "/var/lib/postgresql/data/pgdata"

	// --
)

type Config struct {
	conf               *config.Config
	gitBranch          string // git branch name, used as a namespace for postgres data mounted from host
	composeConfig      *types.Config
	composeProjectName string
	dotenv             []string // environment variables from .env file
	ports              *ports.Ports
}

// HasuraCliVersion extracts version from Hasura CLI docker image. That allows us to keep the same version of Hasura CLI
// both in the docker image and in the hasura-cli on the host
func HasuraCliVersion() (string, error) {
	s := strings.SplitN(svcHasuraDefaultImage, ":", 2)
	if len(s) != 2 {
		return "", fmt.Errorf("invalid hasura cli version: %s", svcHasuraDefaultImage)
	}

	return s[1], nil
}

func NewConfig(conf *config.Config, p *ports.Ports, env []string, gitBranch, projectName string) *Config {
	return &Config{conf: conf, ports: p, dotenv: env, gitBranch: gitBranch, composeProjectName: projectName}
}

func (c Config) addLocaldevExtraHost(svc *types.ServiceConfig) *types.ServiceConfig {
	svc.ExtraHosts = map[string]string{
		hostDockerInternal: hostGateway, // for Linux
	}
	return svc
}

func (c Config) serviceDockerImage(svcName, dockerImageFallback string) string {
	// TODO: decide how to set custom docker image in the new config
	//if svcConf, ok := c.nhostConfig.Services[svcName]; ok && svcConf != nil {
	//	if svcConf.Image != "" {
	//		return svcConf.Image
	//	}
	//}

	return dockerImageFallback
}

func (c *Config) build() *types.Config {
	config := &types.Config{}

	// build services, they may be nil
	services := []*types.ServiceConfig{
		c.traefikService(),
		c.dashboardService(),
		c.postgresService(),
		c.hasuraService(),
		c.authService(),
		c.minioService(),
		c.storageService(),
		c.functionsService(),
		c.mailhogService(),
	}

	// set volumes
	config.Volumes = types.Volumes{
		volFunctionsNodeModules: types.VolumeConfig{},
		volRootNodeModules:      types.VolumeConfig{},
	}

	// loop over services and filter out nils, i.e. services that are not enabled
	for _, service := range services {
		if service != nil {
			config.Services = append(config.Services, *c.addLocaldevExtraHost(service))
		}
	}

	c.composeConfig = config

	return config
}

func (c *Config) BuildYAML() ([]byte, error) {
	return yaml.Marshal(c.build())
}

func (c Config) connectionStringForUser(user string) string {
	postgresEnv := c.postgresServiceEnvs()
	db := postgresEnv[envPostgresDb]

	return fmt.Sprintf("postgres://%s@%s:%d/%s", user, SvcPostgres, dbPort, db)
}

func (c Config) PublicHasuraConnectionString() string {
	return fmt.Sprintf("http://localhost:%d/v1/graphql", c.ports.Proxy())
}

func (c Config) PublicAuthConnectionString() string {
	return fmt.Sprintf("http://localhost:%d/v1/auth", c.ports.Proxy())
}

func (c Config) PublicStorageConnectionString() string {
	return fmt.Sprintf("http://localhost:%d/v1/storage", c.ports.Proxy())
}

func (c Config) PublicFunctionsConnectionString() string {
	return fmt.Sprintf("http://localhost:%d/v1/functions", c.ports.Proxy())
}

func (c Config) PublicPostgresConnectionString() string {
	postgresEnv := c.postgresServiceEnvs()
	user := postgresEnv[envPostgresUser]
	password := postgresEnv[envPostgresPassword]
	db := postgresEnv[envPostgresDb]

	return fmt.Sprintf("postgres://%s:%s@localhost:%d/%s", user, password, c.ports.DB(), db)
}

func (c Config) DashboardURL() string {
	return fmt.Sprintf("http://localhost:%d", c.ports.Dashboard())
}

func (c Config) mailhogServiceEnvs() env {
	smtpConf := c.conf.Provider().GetSmtp()

	return env{
		"SMTP_HOST":   smtpConf.GetHost(),
		"SMTP_PORT":   fmt.Sprint(smtpConf.GetPort()),
		"SMTP_PASS":   smtpConf.GetPassword(),
		"SMTP_USER":   smtpConf.GetUser(),
		"SMTP_SECURE": fmt.Sprint(smtpConf.GetSecure()),
		"SMTP_SENDER": smtpConf.GetSender(),
	}.merge(
		// global environment variables from config
		c.globalEnvs(),
	)
}

func (c Config) dashboardServiceEnvs() env {
	e := env{
		envDashboardNextPublicNhostLocalBackendPort: fmt.Sprint(c.ports.Proxy()),
		envDashboardNextPublicNhostHasuraPort:       fmt.Sprint(c.ports.HasuraConsole()),
		envDashboardNextPublicNhostMigrationsPort:   fmt.Sprint(c.ports.HasuraConsoleAPI()),
		envDashboardNextPublicNhostPlatform:         envDashboardNextPublicNhostPlatformValue,
		envDashboardNextPublicEnv:                   envDashboardNextPublicEnvValue,
		envDashboardNextTelemetryDisabled:           envDashboardNextTelemetryDisabledValue,
	}

	e.mergeWithSlice(c.dotenv)
	return e
}

func (c Config) dashboardService() *types.ServiceConfig {
	return &types.ServiceConfig{
		Name:  SvcDashboard,
		Image: c.serviceDockerImage(SvcDashboard, svcDashboardDefaultImage),
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    3000,
				Published: fmt.Sprint(c.ports.Dashboard()),
				Protocol:  "tcp",
			},
		},
		Environment: c.dashboardServiceEnvs().dockerServiceConfigEnv(),
	}
}

func (c Config) mailhogService() *types.ServiceConfig {
	authEnv := c.authServiceEnvs()
	if authEnv["AUTH_SMTP_HOST"] != SvcMailhog {
		return nil
	}

	return &types.ServiceConfig{
		Name:        SvcMailhog,
		Environment: c.mailhogServiceEnvs().dockerServiceConfigEnv(),
		Restart:     types.RestartPolicyAlways,
		Image:       c.serviceDockerImage(SvcMailhog, svcMailhogDefaultImage),
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    1025,
				Published: fmt.Sprint(c.ports.SMTP()),
				Protocol:  "tcp",
			},
			{
				Mode:      "ingress",
				Target:    8025,
				Published: fmt.Sprint(c.ports.Mailhog()),
				Protocol:  "tcp",
			},
		},
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   types.VolumeTypeBind,
				Source: MailHogDataDirGiBranchScopedPath(c.gitBranch),
				Target: "/maildir",
			},
		},
	}
}

func (c Config) minioServiceEnvs() env {
	return env{
		envMinioRootUser:     nhost.MINIO_USER,
		envMinioRootPassword: nhost.MINIO_PASSWORD,
	}.merge(
		// global environment variables from config
		c.globalEnvs(),
	)
}

func (c Config) minioService() *types.ServiceConfig {
	return &types.ServiceConfig{
		Name:        SvcMinio,
		Environment: c.minioServiceEnvs().dockerServiceConfigEnv(),
		Restart:     types.RestartPolicyAlways,
		Image:       c.serviceDockerImage(SvcMinio, svcMinioDefaultImage),
		Command:     []string{"server", "/data", "--address", "0.0.0.0:9000", "--console-address", "0.0.0.0:8484"},
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    9000,
				Published: fmt.Sprint(c.ports.MinioS3()),
				Protocol:  "tcp",
			},
			{
				Mode:     "ingress",
				Target:   8484,
				Protocol: "tcp",
			},
		},
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   types.VolumeTypeBind,
				Source: MinioDataDirGitBranchScopedPath(c.gitBranch),
				Target: "/data",
			},
		},
	}
}

func (c Config) traefikServiceUrl(svc string) string {
	return fmt.Sprintf("http://%s:%d/v1/%s", SvcTraefik, proxyPort, svc)
}

func (c Config) functionsServiceEnvs() env {
	hasuraConf := c.conf.Hasura()

	return env{
		"NHOST_BACKEND_URL":    c.envValueNhostBackendUrl(),
		"NHOST_SUBDOMAIN":      "localhost",
		"NHOST_REGION":         "",
		"NHOST_HASURA_URL":     fmt.Sprintf("http://%s:%d", SvcTraefik, proxyPort),
		"NHOST_GRAPHQL_URL":    c.traefikServiceUrl(SvcGraphql),
		"NHOST_AUTH_URL":       c.traefikServiceUrl(SvcAuth),
		"NHOST_STORAGE_URL":    c.traefikServiceUrl(SvcStorage),
		"NHOST_FUNCTIONS_URL":  c.traefikServiceUrl(SvcFunctions),
		"NHOST_ADMIN_SECRET":   hasuraConf.GetAdminSecret(),
		"NHOST_WEBHOOK_SECRET": hasuraConf.GetWebhookSecret(),
		"NHOST_JWT_SECRET":     c.graphqlJwtSecret(),
	}.merge(
		// global environment variables from config
		c.globalEnvs(),
	)
}

func (c Config) functionsServiceHealthcheck(interval, startPeriod time.Duration) *types.HealthCheckConfig {
	i := types.Duration(interval)
	s := types.Duration(startPeriod)
	return &types.HealthCheckConfig{
		Test:        []string{"CMD-SHELL", "wget http://localhost:3000/healthz -q -O - > /dev/null 2>&1"},
		Interval:    &i,
		StartPeriod: &s,
	}
}

func (c Config) functionsService() *types.ServiceConfig {
	labels := map[string]string{
		"traefik.enable": "true",
		"traefik.http.middlewares.strip-functions.stripprefix.prefixes":                "/v1/functions",
		"traefik.http.middlewares.functions-cors.headers.accessControlAllowOriginList": "*",
		"traefik.http.middlewares.functions-cors.headers.accessControlAllowHeaders":    "origin,Accept,Authorization,Content-Type",
		"traefik.http.routers.functions.rule":                                          "PathPrefix(`/v1/functions`)",
		"traefik.http.routers.functions.middlewares":                                   "functions-cors@docker,strip-functions@docker",
		"traefik.http.routers.functions.entrypoints":                                   "web",
	}

	return &types.ServiceConfig{
		Name:        SvcFunctions,
		Image:       c.serviceDockerImage(SvcFunctions, svcFunctionsDefaultImage),
		Labels:      labels,
		Restart:     types.RestartPolicyAlways,
		Expose:      []string{"3000"},
		Environment: c.functionsServiceEnvs().dockerServiceConfigEnv(),
		HealthCheck: c.functionsServiceHealthcheck(time.Second*1, time.Minute*30), // 30 minutes is the maximum allowed time for a "functions" service to start, see more below
		// Probe failure during that period will not be counted towards the maximum number of retries
		// However, if a health check succeeds during the start period, the container is considered started and all
		// consecutive failures will be counted towards the maximum number of retries.
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   types.VolumeTypeBind,
				Source: "..",
				Target: "/opt/project",
			},
			{
				Type:   types.VolumeTypeVolume,
				Source: volRootNodeModules,
				Target: "/opt/project/node_modules",
			},
			{
				Type:   types.VolumeTypeVolume,
				Source: volFunctionsNodeModules,
				Target: "/opt/project/functions/node_modules",
			},
		},
	}
}

func (c Config) graphqlJwtSecret() string {
	hasuraConf := c.conf.Hasura()
	var graphqlJwtSecret string

	if len(hasuraConf.GetJwtSecrets()) > 0 { // TODO: what if we have more than one secret?
		graphqlJwtSecret = fmt.Sprintf(`{"type":"%s", "key": "%s"}`, hasuraConf.JwtSecrets[0].Type, hasuraConf.JwtSecrets[0].Key)
	}

	return graphqlJwtSecret
}

func (c Config) twilioSettings() (accountSid, authToken, messagingServiceId string) {
	providerConf := c.conf.Provider()

	if strings.ToLower(providerConf.GetSms().GetProvider()) == "twilio" {
		accountSid = providerConf.Sms.AccountSid
		authToken = providerConf.Sms.AuthToken
		messagingServiceId = providerConf.Sms.MessagingServiceId
	}

	return
}

func (c Config) storageServiceEnvs() env {
	minioEnv := c.minioServiceEnvs()
	hasuraConf := c.conf.Hasura()

	return env{
		"DEBUG":                       "true",
		"BIND":                        ":8576",
		"PUBLIC_URL":                  fmt.Sprintf("http://localhost:%d", c.ports.Proxy()),
		"API_ROOT_PREFIX":             "/v1/storage",
		"POSTGRES_MIGRATIONS":         "1",
		"HASURA_METADATA":             "1",
		"HASURA_ENDPOINT":             c.hasuraEndpoint(),
		"HASURA_GRAPHQL_ADMIN_SECRET": hasuraConf.GetAdminSecret(),
		"S3_ACCESS_KEY":               minioEnv[envMinioRootUser],
		"S3_SECRET_KEY":               minioEnv[envMinioRootPassword],
		//"S3_ENDPOINT":                 "", // TODO: ?
		"S3_BUCKET":                  "nhost",
		"HASURA_GRAPHQL_JWT_SECRET":  c.graphqlJwtSecret(),
		"NHOST_JWT_SECRET":           c.graphqlJwtSecret(),
		"NHOST_ADMIN_SECRET":         hasuraConf.GetAdminSecret(),
		"NHOST_WEBHOOK_SECRET":       hasuraConf.GetWebhookSecret(),
		"POSTGRES_MIGRATIONS_SOURCE": fmt.Sprintf("%s?sslmode=disable", c.connectionStringForUser("nhost_storage_admin")),
	}.merge(
		// global environment variables from config
		c.globalEnvs(),
	)
}

func (c Config) storageService() *types.ServiceConfig {
	labels := map[string]string{
		"traefik.enable":                           "true",
		"traefik.http.routers.storage.rule":        "PathPrefix(`/v1/storage`)",
		"traefik.http.routers.storage.entrypoints": "web",
	}

	return &types.ServiceConfig{
		Name:        SvcStorage,
		Restart:     types.RestartPolicyAlways,
		Image:       c.serviceDockerImage(SvcStorage, svcStorageDefaultImage),
		Environment: c.storageServiceEnvs().dockerServiceConfigEnv(),
		Labels:      labels,
		Command:     []string{"serve"},
		Expose:      []string{"8576"},
	}
}

func (c Config) globalEnvs() env {
	o := env{}
	globalEnvs := c.conf.Global().Environment
	for _, e := range globalEnvs {
		o[e.Name] = e.Value
	}
	return o
}

func (c Config) authServiceEnvs() env {
	authConf := c.conf.Auth()
	hasuraConf := c.conf.Hasura()
	providerConf := c.conf.Provider()

	twilioAccountSid, twilioAuthToken, twilioMessagingServiceId := c.twilioSettings()

	return env{
		// default environment variables
		"AUTH_HOST":                   "0.0.0.0",
		"HASURA_GRAPHQL_DATABASE_URL": c.connectionStringForUser("nhost_auth_admin"),
		"HASURA_GRAPHQL_GRAPHQL_URL":  fmt.Sprintf("%s/graphql", c.hasuraEndpoint()),
		"AUTH_SERVER_URL":             c.PublicAuthConnectionString(),
	}.merge(
		// global environment variables from config
		c.globalEnvs(),
	).merge(env{
		// environment variables from config
		"NHOST_ADMIN_SECRET":          hasuraConf.GetAdminSecret(),
		"NHOST_WEBHOOK_SECRET":        hasuraConf.GetWebhookSecret(),
		"HASURA_GRAPHQL_JWT_SECRET":   c.graphqlJwtSecret(),
		"HASURA_GRAPHQL_ADMIN_SECRET": hasuraConf.GetAdminSecret(),
		//"AUTH_PORT":                   "", // TODO: ???
		"AUTH_SMTP_PASS":   providerConf.GetSmtp().GetPassword(),
		"AUTH_SMTP_HOST":   providerConf.GetSmtp().GetHost(),
		"AUTH_SMTP_USER":   providerConf.GetSmtp().GetUser(),
		"AUTH_SMTP_SENDER": providerConf.GetSmtp().GetSender(),
		//"AUTH_SMTP_AUTH_METHOD":                     "", // TODO: ???
		"AUTH_SMTP_PORT":    fmt.Sprint(providerConf.GetSmtp().GetPort()),
		"AUTH_SMTP_SECURE":  fmt.Sprint(providerConf.GetSmtp().GetSecure()),
		"AUTH_SMS_PROVIDER": providerConf.GetSms().GetProvider(),
		//"AUTH_SMS_TEST_PHONE_NUMBERS":               "", // TODO: ???
		"AUTH_SMS_TWILIO_ACCOUNT_SID":               twilioAccountSid,
		"AUTH_SMS_TWILIO_AUTH_TOKEN":                twilioAuthToken,
		"AUTH_SMS_TWILIO_MESSAGING_SERVICE_ID":      twilioMessagingServiceId,
		"AUTH_GRAVATAR_ENABLED":                     fmt.Sprint(authConf.GetUser().GetGravatar().GetEnabled()),
		"AUTH_GRAVATAR_DEFAULT":                     authConf.GetUser().GetGravatar().GetDefault(),
		"AUTH_GRAVATAR_RATING":                      authConf.GetUser().GetGravatar().GetRating(),
		"AUTH_CLIENT_URL":                           authConf.GetRedirections().GetClientUrl(),
		"AUTH_WEBAUTHN_ENABLED":                     fmt.Sprint(authConf.GetMethod().GetWebauthn().GetEnabled()),
		"AUTH_WEBAUTHN_RP_NAME":                     authConf.GetMethod().GetWebauthn().GetRelyingParty().GetName(),
		"AUTH_WEBAUTHN_RP_ORIGINS":                  strings.Join(authConf.GetMethod().GetWebauthn().GetRelyingParty().GetOrigins(), ","),
		"AUTH_WEBAUTHN_ATTESTATION_TIMEOUT":         fmt.Sprint(authConf.GetMethod().GetWebauthn().GetAttestation().GetTimeout()),
		"AUTH_ANONYMOUS_USERS_ENABLED":              fmt.Sprint(authConf.GetMethod().GetAnonymous().GetEnabled()),
		"AUTH_DISABLE_NEW_USERS":                    fmt.Sprint(!authConf.GetSignUp().GetEnabled()),
		"AUTH_ACCESS_CONTROL_ALLOWED_EMAILS":        strings.Join(authConf.GetUser().GetEmail().GetAllowed(), ","),
		"AUTH_ACCESS_CONTROL_ALLOWED_EMAIL_DOMAINS": strings.Join(authConf.GetUser().GetEmailDomains().GetAllowed(), ","),
		"AUTH_ACCESS_CONTROL_BLOCKED_EMAILS":        strings.Join(authConf.GetUser().GetEmail().GetBlocked(), ","),
		"AUTH_ACCESS_CONTROL_BLOCKED_EMAIL_DOMAINS": strings.Join(authConf.GetUser().GetEmailDomains().GetBlocked(), ","),
		"AUTH_PASSWORD_MIN_LENGTH":                  fmt.Sprint(authConf.GetMethod().GetEmailPassword().GetPasswordMinLength()),
		"AUTH_PASSWORD_HIBP_ENABLED":                fmt.Sprint(authConf.GetMethod().GetEmailPassword().GetHibpEnabled()),
		"AUTH_USER_DEFAULT_ROLE":                    authConf.GetUser().GetRoles().GetDefault(),
		"AUTH_USER_DEFAULT_ALLOWED_ROLES":           strings.Join(authConf.GetUser().GetRoles().GetAllowed(), ","),
		"AUTH_LOCALE_DEFAULT":                       authConf.GetUser().GetLocale().GetDefault(),
		"AUTH_LOCALE_ALLOWED_LOCALES":               strings.Join(authConf.GetUser().GetLocale().GetAllowed(), ","),
		"AUTH_EMAIL_PASSWORDLESS_ENABLED":           fmt.Sprint(authConf.GetMethod().GetEmailPasswordless().GetEnabled()),
		"AUTH_SMS_PASSWORDLESS_ENABLED":             fmt.Sprint(authConf.GetMethod().GetSmsPasswordless().GetEnabled()),
		"AUTH_EMAIL_SIGNIN_EMAIL_VERIFIED_REQUIRED": fmt.Sprint(authConf.GetMethod().GetEmailPassword().GetEmailVerificationRequired()),
		"AUTH_ACCESS_CONTROL_ALLOWED_REDIRECT_URLS": strings.Join(authConf.GetRedirections().GetAllowedUrls(), ","),
		"AUTH_MFA_ENABLED":                          fmt.Sprint(authConf.GetTotp().GetEnabled()),
		"AUTH_MFA_TOTP_ISSUER":                      authConf.GetTotp().GetIssuer(),
		"AUTH_ACCESS_TOKEN_EXPIRES_IN":              fmt.Sprint(authConf.GetSession().GetAccessToken().GetExpiresIn()),
		"AUTH_REFRESH_TOKEN_EXPIRES_IN":             fmt.Sprint(authConf.GetSession().GetRefreshToken().GetExpiresIn()),
		"AUTH_EMAIL_TEMPLATE_FETCH_URL":             "", // TODO: deprecated?
		"AUTH_JWT_CUSTOM_CLAIMS":                    "", // TODO: ???
		"AUTH_CONCEAL_ERRORS":                       "", // TODO: ???
	})
}

func (c Config) authServiceHealthcheck(interval, startPeriod time.Duration) *types.HealthCheckConfig {
	i := types.Duration(interval)
	s := types.Duration(startPeriod)
	return &types.HealthCheckConfig{
		Test:        []string{"CMD-SHELL", "wget http://localhost:4000/healthz -q -O - > /dev/null 2>&1"},
		Interval:    &i,
		StartPeriod: &s,
	}
}

func (c Config) authService() *types.ServiceConfig {
	labels := map[string]string{
		"traefik.enable": "true",
		"traefik.http.middlewares.strip-auth.stripprefix.prefixes": "/v1/auth",
		"traefik.http.routers.auth.rule":                           "PathPrefix(`/v1/auth`)",
		"traefik.http.routers.auth.middlewares":                    "strip-auth@docker",
		"traefik.http.routers.auth.entrypoints":                    "web",
	}

	return &types.ServiceConfig{
		Name:        SvcAuth,
		Image:       c.serviceDockerImage(SvcAuth, svcAuthDefaultImage),
		Environment: c.authServiceEnvs().dockerServiceConfigEnv(),
		Labels:      labels,
		Expose:      []string{"4000"},
		DependsOn: map[string]types.ServiceDependency{
			SvcPostgres: {
				Condition: types.ServiceConditionHealthy,
			},
			SvcGraphql: {
				Condition: types.ServiceConditionStarted,
			},
		},
		Restart:     types.RestartPolicyAlways,
		HealthCheck: c.authServiceHealthcheck(time.Second*3, time.Minute*5),
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   types.VolumeTypeBind,
				Source: filepath.Join(nhost.DOT_NHOST_DIR, "custom"),
				Target: "/app/custom",
			},
			{
				Type:   types.VolumeTypeBind,
				Source: nhost.EMAILS_DIR,
				Target: "/app/email-templates",
			},
		},
	}
}

func (c Config) envValueNhostBackendUrl() string {
	return fmt.Sprintf("http://traefik:%d", proxyPort)
}

func (c Config) hasuraEndpoint() string {
	return fmt.Sprintf("http://%s:%d/v1", SvcGraphql, graphqlPort)
}

func (c Config) hasuraServiceEnvs() env {
	hasuraConf := c.conf.Hasura()

	return env{
		"HASURA_GRAPHQL_DATABASE_URL":              c.connectionStringForUser("nhost_hasura"),
		"HASURA_GRAPHQL_JWT_SECRET":                c.graphqlJwtSecret(),
		"HASURA_GRAPHQL_ADMIN_SECRET":              hasuraConf.GetAdminSecret(),
		"NHOST_ADMIN_SECRET":                       hasuraConf.GetAdminSecret(),
		"NHOST_BACKEND_URL":                        c.envValueNhostBackendUrl(),
		"NHOST_SUBDOMAIN":                          "localhost",
		"NHOST_REGION":                             "",
		"NHOST_HASURA_URL":                         fmt.Sprintf("http://%s:%d", SvcTraefik, proxyPort),
		"NHOST_GRAPHQL_URL":                        c.traefikServiceUrl(SvcGraphql),
		"NHOST_AUTH_URL":                           c.traefikServiceUrl(SvcAuth),
		"NHOST_STORAGE_URL":                        c.traefikServiceUrl(SvcStorage),
		"NHOST_FUNCTIONS_URL":                      c.traefikServiceUrl(SvcFunctions),
		"HASURA_GRAPHQL_UNAUTHORIZED_ROLE":         "public",
		"HASURA_GRAPHQL_DEV_MODE":                  "true",
		"HASURA_GRAPHQL_LOG_LEVEL":                 "debug",
		"HASURA_GRAPHQL_ENABLE_CONSOLE":            "false",
		"HASURA_GRAPHQL_MIGRATIONS_SERVER_TIMEOUT": "20",
		"HASURA_GRAPHQL_NO_OF_RETRIES":             "20",
		"HASURA_GRAPHQL_ENABLE_TELEMETRY":          "false",
		"NHOST_WEBHOOK_SECRET":                     hasuraConf.GetWebhookSecret(),
	}.merge(
		// global environment variables from config
		c.globalEnvs(),
	)
}

func (c Config) hasuraService() *types.ServiceConfig {
	labels := map[string]string{
		"traefik.enable":                          "true",
		"traefik.http.routers.hasura.rule":        "PathPrefix(`/v1/graphql`, `/v2/query`, `/v1/metadata`, `/v1/config`)",
		"traefik.http.routers.hasura.entrypoints": "web",
	}

	return &types.ServiceConfig{
		Name:        SvcGraphql,
		Image:       c.serviceDockerImage(SvcHasura, svcHasuraDefaultImage),
		Environment: c.hasuraServiceEnvs().dockerServiceConfigEnv(),
		Labels:      labels,
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    graphqlPort,
				Published: fmt.Sprint(c.ports.GraphQL()),
				Protocol:  "tcp",
			},
		},
		DependsOn: map[string]types.ServiceDependency{
			SvcPostgres: {
				Condition: types.ServiceConditionHealthy,
			},
		},
		Restart: types.RestartPolicyAlways,
	}
}

func (c Config) postgresServiceEnvs() env {
	return env{
		// default environment variables
		envPostgresData:     envPostgresDataDefaultValue,
		envPostgresUser:     envPostgresUserDefaultValue,
		envPostgresPassword: envPostgresPasswordDefaultValue,
		envPostgresDb:       envPostgresDbDefaultValue,
	}.merge(
		// global environment variables from config
		c.globalEnvs(),
	)
}

func (c Config) postgresServiceHealthcheck(interval, startPeriod time.Duration) *types.HealthCheckConfig {
	i := types.Duration(interval)
	s := types.Duration(startPeriod)

	e := c.postgresServiceEnvs()
	pgUser := e[envPostgresUser]
	pgDb := e[envPostgresDb]

	return &types.HealthCheckConfig{
		Test:        []string{"CMD-SHELL", fmt.Sprintf("pg_isready -U %s -d %s -q", pgUser, pgDb)},
		Interval:    &i,
		StartPeriod: &s,
	}
}

func (c Config) postgresService() *types.ServiceConfig {
	return &types.ServiceConfig{
		Name: SvcPostgres,
		// keep in mind that the provided postgres image should create schemas and triggers like in https://github.com/nhost/postgres/blob/ea53451b6df9f4b10ce515a2cefbd9ddfdfadb25/v12/db/0001-create-schema.sql
		Image:       c.serviceDockerImage(SvcPostgres, svcPostgresDefaultImage),
		Restart:     types.RestartPolicyAlways,
		Environment: c.postgresServiceEnvs().dockerServiceConfigEnv(),
		HealthCheck: c.postgresServiceHealthcheck(time.Second*3, time.Minute*2),
		Command: []string{
			"postgres",
			"-c", "config_file=/etc/postgresql.conf",
			"-c", "hba_file=/etc/pg_hba_local.conf",
		},
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   types.VolumeTypeBind,
				Source: DbDataDirGitBranchScopedPath(c.gitBranch, dataDirPgdata),
				Target: envPostgresDataDefaultValue,
			},
			{
				Type:   types.VolumeTypeBind,
				Source: DbDataDirGitBranchScopedPath(c.gitBranch, "pg_hba_local.conf"),
				Target: "/etc/pg_hba_local.conf",
			},
		},
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    dbPort,
				Published: fmt.Sprint(c.ports.DB()),
				Protocol:  "tcp",
			},
		},
	}
}

func (c Config) traefikService() *types.ServiceConfig {
	return &types.ServiceConfig{
		Name:    SvcTraefik,
		Image:   c.serviceDockerImage(SvcTraefik, svcTraefikDefaultImage),
		Restart: types.RestartPolicyAlways,
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    proxyPort,
				Published: fmt.Sprint(c.ports.Proxy()),
				Protocol:  "tcp",
			},
			{
				Mode:     "ingress",
				Target:   8080,
				Protocol: "tcp",
			},
		},
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:     types.VolumeTypeBind,
				Source:   "/var/run/docker.sock",
				Target:   "/var/run/docker.sock",
				ReadOnly: true,
			},
		},
		Command: []string{
			"--api.insecure=true",
			"--providers.docker=true",
			"--providers.docker.exposedbydefault=false",
			fmt.Sprintf("--providers.docker.constraints=Label(`com.docker.compose.project`,`%s`)", c.composeProjectName),
			fmt.Sprintf("--entrypoints.web.address=:%d", proxyPort),
		},
	}
}
