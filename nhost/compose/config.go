package compose

import (
	"encoding/json"
	"fmt"
	"github.com/compose-spec/compose-go/types"
	"github.com/nhost/cli/nhost"
	"github.com/nhost/cli/util"
	"path/filepath"
	"time"
)

// TODO: allow to set custom postgres user/password
const (
	postgresDefaultPassword = "postgres"

	// docker compose service names
	svcPostgres      = "postgres"
	svcAuth          = "auth"
	svcStorage       = "storage"
	svcFunctions     = "functions"
	svcMinio         = "minio"
	svcMailhog       = "mailhog"
	svcHasura        = "hasura"
	svcHasuraConsole = "hasura-console"
	svcTraefik       = "traefik"
	svcGraphqlEngine = "graphql-engine"

	// default docker images
	svcPostgresDefaultImage      = "nhost/postgres:12-v0.0.6"
	svcAuthDefaultImage          = "nhost/hasura-auth:0.6.3"
	svcStorageDefaultImage       = "nhost/hasura-storage:0.2.1"
	svcFunctionsDefaultImage     = "nhost/functions"
	svcMinioDefaultImage         = "minio/minio:RELEASE.2021-09-24T00-24-24Z"
	svcMailhogDefaultImage       = "mailhog/mailhog"
	svcHasuraDefaultImage        = "hasura/graphql-engine:v2.2.0"
	svcHasuraConsoleDefaultImage = "nhost/hasura:v2.8.1"
	svcTraefikDefaultImage       = "traefik:v2.8"

	// environment variables

	// envs prefixes
	envPrefixAuth    = "AUTH"
	envPrefixStorage = "STORAGE"

	// postgres
	envPostgresPassword = "POSTGRES_PASSWORD"
	envPostgresDb       = "POSTGRES_DB"
	envPostgresUser     = "POSTGRES_USER"
	envPostgresData     = "PGDATA"

	//

	// default values for environment variables
	envPostgresDbDefaultValue       = "postgres"
	envPostgresUserDefaultValue     = "postgres"
	envPostgresPasswordDefaultValue = "postgres"
	envPostgresDataDefaultValue     = "/var/lib/postgresql/data/pgdata"
)

type Config struct {
	nhostConfig        *nhost.Configuration // nhost configuration to read custom values from, not used atm
	gitBranch          string               // git branch name, used as a namespace for postgres data mounted from host
	composeConfig      *types.Config
	composeProjectName string
	dotenv             []string // environment variables from .env file
}

func NewConfig(conf *nhost.Configuration, env []string, gitBranch, projectName string) *Config {
	return &Config{nhostConfig: conf, dotenv: env, gitBranch: gitBranch, composeProjectName: projectName}
}

func (c Config) serviceDockerImage(svcName, dockerImageFallback string) string {
	if svcConf, ok := c.nhostConfig.Services[svcName]; ok {
		if svcConf.Image != "" {
			return svcConf.Image
		}
	}

	return dockerImageFallback
}

// serviceConfigEnvs returns environment variables from "services".$name."environment" section in yaml config
func (c *Config) serviceConfigEnvs(svc string) env {
	e := env{}

	if svcConf, ok := c.nhostConfig.Services[svc]; ok {
		e.mergeWithServiceEnv(svcConf.Environment)
	}

	return e
}

func (c *Config) build() *types.Config {
	config := &types.Config{}

	// build services, they may be nil
	services := []*types.ServiceConfig{
		c.traefikService(),
		c.postgresService(),
		c.hasuraService(),
		c.hasuraConsoleService(),
		c.authService(),
		c.storageService(),
		c.functionsService(),
		c.minioService(),
		c.mailhogService(),
	}

	// loop over services and filter out nils, i.e. services that are not enabled
	for _, service := range services {
		if service != nil {
			config.Services = append(config.Services, *service)
		}
	}

	// set volumes
	config.Volumes = types.Volumes{
		"functions_node_modules": types.VolumeConfig{},
	}

	c.composeConfig = config

	return config
}

func (c Config) hostDataDirectory(path string) string {
	return filepath.Join("data", path)
}

func (c Config) hostDataDirectoryBranchScoped(path string) string {
	return filepath.Join("data", path, c.gitBranch)
}

func (c *Config) BuildJSON() ([]byte, error) {
	return json.MarshalIndent(c.build(), "", "  ")
}

func (c Config) postgresPasswordEnvValueWithDefaultValue() string {
	return fmt.Sprintf("${POSTGRES_PASSWORD:-%s}", postgresDefaultPassword)
}

func (c Config) postgresConnectionString() string {
	return fmt.Sprintf("postgres://postgres:%s@postgres:5432/postgres", c.postgresPasswordEnvValueWithDefaultValue())
}

func (c Config) mailhogServiceEnvs() env {
	e := env{
		"SMTP_HOST":   "${AUTH_SMTP_HOST:-mailhog}",
		"SMTP_PORT":   "${AUTH_SMTP_PORT:-1025}",
		"SMTP_PASS":   "${AUTH_SMTP_PASS:-password}",
		"SMTP_USER":   "${AUTH_SMTP_USER:-user}",
		"SMTP_SECURE": "${AUTH_SMTP_SECURE:-false}",
		"SMTP_SENDER": "${AUTH_SMTP_SENDER:-hbp@hbp.com}",
	}
	e.merge(c.serviceConfigEnvs(svcMailhog))
	return e
}

func (c Config) mailhogService() *types.ServiceConfig {
	return &types.ServiceConfig{
		Name:        svcMailhog,
		Environment: c.mailhogServiceEnvs().dockerServiceConfigEnv(),
		Restart:     types.RestartPolicyAlways,
		Image:       c.serviceDockerImage(svcMailhog, svcMailhogDefaultImage),
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    1025,
				Published: "1025",
				Protocol:  "tcp",
			},
		},
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   types.VolumeTypeBind,
				Source: c.hostDataDirectory("mailhog"),
				Target: "/maildir",
			},
		},
	}
}

func (c Config) minioServiceEnvs() env {
	e := env{
		"MINIO_ROOT_USER":     "minioaccesskey123123", // TODO: creds
		"MINIO_ROOT_PASSWORD": "minioaccesskey123123",
	}
	e.merge(c.serviceConfigEnvs(svcMinio))
	return e
}

func (c Config) minioService() *types.ServiceConfig {
	return &types.ServiceConfig{
		Name:        svcMinio,
		Environment: c.minioServiceEnvs().dockerServiceConfigEnv(),
		Restart:     types.RestartPolicyAlways,
		Image:       c.serviceDockerImage(svcMinio, svcMinioDefaultImage),
		Entrypoint:  []string{"sh"},
		Command:     []string{"-c", "mkdir -p /data/nhost && /opt/bin/minio server --address :8484 /data"}, // TODO: port
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    8484, // TODO: port
				Published: "8484",
				Protocol:  "tcp",
			},
		},
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   types.VolumeTypeBind,
				Source: c.hostDataDirectory("minio"),
				Target: "/data",
			},
		},
	}
}

func (c Config) functionsServiceEnvs() env {
	e := env{"NHOST_BACKEND_URL": "http://localhost:1337"}
	e.mergeWithSlice(c.dotenv)
	return e
}

func (c Config) functionsService() *types.ServiceConfig {
	labels := map[string]string{
		"traefik.enable": "true",
		"traefik.http.middlewares.strip-functions.stripprefix.prefixes": "/v1/functions",
		"traefik.http.routers.functions.rule":                           "Host(`localhost`) && PathPrefix(`/v1/functions`)",
		"traefik.http.routers.functions.middlewares":                    "strip-functions@docker",
		"traefik.http.routers.functions.entrypoints":                    "web",
	}

	return &types.ServiceConfig{
		Name:        svcFunctions,
		Image:       c.serviceDockerImage(svcFunctions, svcFunctionsDefaultImage), // TODO: build, push & pin version
		Labels:      labels,
		Restart:     types.RestartPolicyAlways,
		Expose:      []string{"3000"},
		Environment: c.functionsServiceEnvs().dockerServiceConfigEnv(),
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   types.VolumeTypeBind,
				Source: "..",
				Target: "/opt/project",
			},
			{
				Type:   types.VolumeTypeVolume,
				Source: "functions_node_modules",
				Target: "/opt/project/node_modules",
			},
			{
				Type:   types.VolumeTypeVolume,
				Target: "/opt/project/data/",
			},
			{
				Type:   types.VolumeTypeVolume,
				Target: "/opt/project/initdb.d/",
			},
		},
	}
}

func (c Config) storageServiceEnvs() env {
	e := env{
		"BIND":                        "8576",                  // TODO: randomize port
		"PUBLIC_URL":                  "http://localhost:8576", // TODO: port
		"POSTGRES_MIGRATIONS":         "1",
		"HASURA_METADATA":             "1",
		"HASURA_ENDPOINT":             "http://graphql-engine:8080/v1",
		"HASURA_GRAPHQL_ADMIN_SECRET": util.ADMIN_SECRET,
		"S3_ACCESS_KEY":               "minioaccesskey123123",
		"S3_SECRET_KEY":               "minioaccesskey123123",
		"S3_ENDPOINT":                 "http://minio:8484",
		"S3_BUCKET":                   "nhost",
		"HASURA_GRAPHQL_JWT_SECRET":   fmt.Sprintf(`{"type":"HS256", "key": "%s"}`, util.JWT_KEY),
		"NHOST_JWT_SECRET":            fmt.Sprintf(`{"type":"HS256", "key": "%s"}`, util.JWT_KEY),
		"NHOST_ADMIN_SECRET":          util.ADMIN_SECRET,
		"NHOST_WEBHOOK_SECRET":        util.WEBHOOK_SECRET,
		"POSTGRES_MIGRATIONS_SOURCE":  fmt.Sprintf("%s?sslmode=disable", c.postgresConnectionString()),
		"NHOST_BACKEND_URL":           "http://localhost:1337",
	}

	e.merge(c.serviceConfigEnvs(svcStorage))
	e.mergeWithConfigEnv(c.nhostConfig.Storage, envPrefixStorage)

	return e
}

func (c Config) storageService() *types.ServiceConfig {
	labels := map[string]string{
		"traefik.enable":                           "true",
		"traefik.http.routers.storage.rule":        "Host(`localhost`) && PathPrefix(`/v1/storage`)",
		"traefik.http.routers.storage.entrypoints": "web",
		// Rewrite the path so it matches with the new storage API path introduced in hasura-storage 0.2
		"traefik.http.middlewares.strip-suffix.replacepathregex.regex":       "^/v1/storage/(.*)",
		"traefik.http.middlewares.strip-suffix.replacepathregex.replacement": "/v1/$$1",
		"traefik.http.routers.storage.middlewares":                           "strip-suffix@docker",
	}

	return &types.ServiceConfig{
		Name:        svcStorage,
		Restart:     types.RestartPolicyAlways,
		Image:       c.serviceDockerImage(svcStorage, svcStorageDefaultImage),
		Environment: c.storageServiceEnvs().dockerServiceConfigEnv(),
		Labels:      labels,
		Command:     []string{"serve"},
		Expose:      []string{"8000"},
	}
}

func (c Config) authServiceEnvs() env {
	e := env{
		"AUTH_HOST":                   "0.0.0.0",
		"HASURA_GRAPHQL_DATABASE_URL": c.postgresConnectionString(),
		"HASURA_GRAPHQL_GRAPHQL_URL":  "http://graphql-engine:8080/v1/graphql",
		"HASURA_GRAPHQL_JWT_SECRET":   fmt.Sprintf(`{"type":"HS256", "key": "%s"}`, util.JWT_KEY),
		"HASURA_GRAPHQL_ADMIN_SECRET": util.ADMIN_SECRET,
		"AUTH_CLIENT_URL":             "${AUTH_CLIENT_URL:-http://localhost:3000}",
		"AUTH_SMTP_HOST":              "mailhog",
		"AUTH_SMTP_PORT":              "1025",
		"AUTH_SMTP_USER":              "user",
		"AUTH_SMTP_PASS":              "password",
		"AUTH_SMTP_SENDER":            "mail@example.com",
		"NHOST_ADMIN_SECRET":          util.ADMIN_SECRET,
		"NHOST_WEBHOOK_SECRET":        util.WEBHOOK_SECRET,
	}

	e.merge(c.serviceConfigEnvs(svcAuth))
	e.mergeWithConfigEnv(c.nhostConfig.Auth, envPrefixAuth)

	return e
}

func (c Config) authService() *types.ServiceConfig {
	labels := map[string]string{
		"traefik.enable": "true",
		"traefik.http.middlewares.strip-auth.stripprefix.prefixes": "/v1/auth",
		"traefik.http.routers.auth.rule":                           "Host(`localhost`) && PathPrefix(`/v1/auth`)",
		"traefik.http.routers.auth.middlewares":                    "strip-auth@docker",
		"traefik.http.routers.auth.entrypoints":                    "web",
	}

	return &types.ServiceConfig{
		Name:        svcAuth,
		Image:       c.serviceDockerImage(svcAuth, svcAuthDefaultImage),
		Environment: c.authServiceEnvs().dockerServiceConfigEnv(),
		Labels:      labels,
		Expose:      []string{"4000"},
		DependsOn: map[string]types.ServiceDependency{
			svcPostgres: {
				Condition: types.ServiceConditionHealthy,
			},
			svcGraphqlEngine: {
				Condition: types.ServiceConditionStarted,
			},
		},
		Restart: types.RestartPolicyAlways,
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   types.VolumeTypeBind,
				Source: nhost.EMAILS_DIR,
				Target: "/app/email-templates",
			},
		},
	}
}

func (c Config) hasuraServiceEnvs() env {
	e := env{
		"HASURA_GRAPHQL_DATABASE_URL":              c.postgresConnectionString(),
		"HASURA_GRAPHQL_JWT_SECRET":                fmt.Sprintf(`{"type":"HS256", "key": "%s"}`, util.JWT_KEY),
		"HASURA_GRAPHQL_ADMIN_SECRET":              util.ADMIN_SECRET,
		"NHOST_ADMIN_SECRET":                       util.ADMIN_SECRET,
		"NHOST_BACKEND_URL":                        "http://localhost:1337",
		"HASURA_GRAPHQL_UNAUTHORIZED_ROLE":         "public",
		"HASURA_GRAPHQL_DEV_MODE":                  "true",
		"HASURA_GRAPHQL_LOG_LEVEL":                 "debug",
		"HASURA_GRAPHQL_ENABLE_CONSOLE":            "false",
		"HASURA_GRAPHQL_MIGRATIONS_SERVER_TIMEOUT": "20",
		"HASURA_GRAPHQL_NO_OF_RETRIES":             "20",
		"HASURA_GRAPHQL_ENABLE_TELEMETRY":          "false",
		"NHOST_WEBHOOK_SECRET":                     util.WEBHOOK_SECRET,
	}

	e.mergeWithSlice(c.dotenv)
	e.merge(c.serviceConfigEnvs(svcHasura))

	return e
}

func (c Config) hasuraService() *types.ServiceConfig {
	labels := map[string]string{
		"traefik.enable": "true",
	}

	return &types.ServiceConfig{
		Name:        svcGraphqlEngine,
		Image:       c.serviceDockerImage(svcHasura, svcHasuraDefaultImage),
		Environment: c.hasuraServiceEnvs().dockerServiceConfigEnv(),
		//Expose:      []string{"8080"},
		Labels: labels,
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    8080,
				Published: "8080",
				Protocol:  "tcp",
			},
		},
		DependsOn: map[string]types.ServiceDependency{
			svcPostgres: {
				Condition: types.ServiceConditionHealthy,
			},
		},
		Restart: types.RestartPolicyAlways,
	}
}

func (c Config) hasuraConsoleServiceEnvs() env {
	return env{
		"HASURA_GRAPHQL_DATABASE_URL":              c.postgresConnectionString(),
		"HASURA_GRAPHQL_JWT_SECRET":                fmt.Sprintf(`{"type":"HS256", "key": "%s"}`, util.JWT_KEY),
		"HASURA_GRAPHQL_ADMIN_SECRET":              util.ADMIN_SECRET,
		"HASURA_GRAPHQL_ENDPOINT":                  "http://127.0.0.1:8080",
		"HASURA_GRAPHQL_UNAUTHORIZED_ROLE":         "public",
		"HASURA_GRAPHQL_DEV_MODE":                  "true",
		"HASURA_GRAPHQL_LOG_LEVEL":                 "debug",
		"HASURA_GRAPHQL_ENABLE_CONSOLE":            "false",
		"HASURA_RUN_CONSOLE":                       "true",
		"HASURA_GRAPHQL_MIGRATIONS_SERVER_TIMEOUT": "20",
		"HASURA_GRAPHQL_NO_OF_RETRIES":             "20",
		"HASURA_GRAPHQL_ENABLE_TELEMETRY":          "false",
	}
}

func (c Config) hasuraConsoleService() *types.ServiceConfig {
	labels := map[string]string{
		"traefik.enable": "true",
		"traefik.http.services.hasura-console.loadbalancer.server.port": "9695",
		"traefik.http.routers.hasura-console.rule":                      "Host(`localhost`) && PathPrefix(`/`)",
		"traefik.http.routers.hasura-console.entrypoints":               "web",
	}

	return &types.ServiceConfig{
		Name:        svcHasuraConsole,
		Image:       c.serviceDockerImage(svcHasuraConsole, svcHasuraConsoleDefaultImage),
		Environment: c.hasuraConsoleServiceEnvs().dockerServiceConfigEnv(),
		Labels:      labels,
		DependsOn: map[string]types.ServiceDependency{
			svcPostgres: {
				Condition: types.ServiceConditionHealthy,
			},
			svcGraphqlEngine: {
				Condition: types.ServiceConditionStarted,
			},
		},
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    9695,
				Published: "9695",
				Protocol:  "tcp",
			},
			{
				Mode:      "ingress",
				Target:    9693,
				Published: "9693",
				Protocol:  "tcp",
			},
		},
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   types.VolumeTypeBind,
				Source: "../nhost",
				Target: "/usr/src/hasura",
			},
		},
		Restart: types.RestartPolicyAlways,
	}
}

func (c Config) postgresServiceEnvs() env {
	e := env{envPostgresData: envPostgresDataDefaultValue}

	e.merge(c.serviceConfigEnvs(svcPostgres))

	// set defaults
	if e[envPostgresUser] == "" {
		e[envPostgresUser] = envPostgresUserDefaultValue
	}

	if e[envPostgresPassword] == "" {
		e[envPostgresPassword] = envPostgresPasswordDefaultValue
	}

	if e[envPostgresDb] == "" {
		e[envPostgresDb] = envPostgresDbDefaultValue
	}

	return e
}

func (c Config) postgresServiceHealthcheck(interval, startPeriod time.Duration) *types.HealthCheckConfig {
	i := types.Duration(interval)
	s := types.Duration(startPeriod)
	return &types.HealthCheckConfig{
		Test:        []string{"CMD-SHELL", "pg_isready -U postgres -d postgres -q"}, // TODO: don't hardcode postgres user and db name
		Interval:    &i,
		StartPeriod: &s,
	}
}

func (c Config) postgresService() *types.ServiceConfig {
	return &types.ServiceConfig{
		Name: svcPostgres,
		// keep in mind that the provided postgres image should create schemas and triggers like in https://github.com/nhost/postgres/blob/ea53451b6df9f4b10ce515a2cefbd9ddfdfadb25/v12/db/0001-create-schema.sql
		Image:       c.serviceDockerImage(svcPostgres, svcPostgresDefaultImage),
		Restart:     types.RestartPolicyAlways,
		Environment: c.postgresServiceEnvs().dockerServiceConfigEnv(),
		HealthCheck: c.postgresServiceHealthcheck(time.Second*3, time.Minute*2),
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   types.VolumeTypeBind,
				Source: c.hostDataDirectoryBranchScoped("db"),
				Target: envPostgresDataDefaultValue,
			},
		},
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    5432,
				Published: "5432",
				Protocol:  "tcp",
			},
		},
	}
}

func (c Config) traefikService() *types.ServiceConfig {
	return &types.ServiceConfig{
		Name:    svcTraefik,
		Image:   c.serviceDockerImage(svcTraefik, svcTraefikDefaultImage),
		Restart: types.RestartPolicyAlways,
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    1337,
				Published: "1337",
				Protocol:  "tcp",
			},
			{
				Mode:      "ingress",
				Target:    8080,
				Published: "9090",
				Protocol:  "tcp",
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
			"--entrypoints.web.address=:1337",
		},
	}
}
