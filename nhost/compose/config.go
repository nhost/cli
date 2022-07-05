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

const (
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
	// --

	// data directory names
	dataDirDb      = "db"
	dataDirMailhog = "mailhog"
	dataDirMinio   = "minio"
	// --

	// default docker images
	svcPostgresDefaultImage      = "nhost/postgres:12-v0.0.6"
	svcAuthDefaultImage          = "nhost/hasura-auth:0.6.3"
	svcStorageDefaultImage       = "nhost/hasura-storage:0.2.1"
	svcFunctionsDefaultImage     = "nhost/functions" // TODO: build docker image
	svcMinioDefaultImage         = "minio/minio:RELEASE.2021-09-24T00-24-24Z"
	svcMailhogDefaultImage       = "mailhog/mailhog"
	svcHasuraDefaultImage        = "hasura/graphql-engine:v2.2.0"
	svcHasuraConsoleDefaultImage = "nhost/hasura:v2.8.1"
	svcTraefikDefaultImage       = "traefik:v2.8"
	// --

	// environment variables

	// envs prefixes
	envPrefixAuth    = "AUTH"
	envPrefixStorage = "STORAGE"

	// minio
	envMinioRootUser     = "MINIO_ROOT_USER"
	envMinioRootPassword = "MINIO_ROOT_PASSWORD"

	// auth
	envAuthSmtpHost   = "AUTH_SMTP_HOST"
	envAuthSmtpPort   = "AUTH_SMTP_PORT"
	envAuthSmtpUser   = "AUTH_SMTP_USER"
	envAuthSmtpPass   = "AUTH_SMTP_PASS"
	envAuthSmtpSecure = "AUTH_SMTP_SECURE"
	envAuthSmtpSender = "AUTH_SMTP_SENDER"

	// postgres
	envPostgresPassword = "POSTGRES_PASSWORD"
	envPostgresDb       = "POSTGRES_DB"
	envPostgresUser     = "POSTGRES_USER"
	envPostgresData     = "PGDATA"

	// default values for environment variables
	envPostgresDbDefaultValue        = "postgres"
	envPostgresUserDefaultValue      = "postgres"
	envPostgresPasswordDefaultValue  = "postgres"
	envPostgresDataDefaultValue      = "/var/lib/postgresql/data/pgdata"
	envMinioRootUserDefaultValue     = "minioaccesskey123123"
	envMinioRootPasswordDefaultValue = "minioaccesskey123123"

	// --

	// default ports
	serverDefaultPort      = 1337
	svcPostgresDefaultPort = 5432
	svcHasuraDefaultPort   = 8080
	// --
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

func (c Config) serviceDockerExposePort(svcName string, fallbackPort uint32) uint32 {
	if svcConf, ok := c.nhostConfig.Services[svcName]; ok {
		if svcConf.Port != 0 {
			return uint32(svcConf.Port)
		}
	}

	return fallbackPort
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
		c.minioService(),
		c.storageService(),
		c.functionsService(),
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

func (c Config) postgresConnectionString() string {
	postgresEnv := c.postgresServiceEnvs()
	user := postgresEnv[envPostgresUser]
	password := postgresEnv[envPostgresPassword]
	db := postgresEnv[envPostgresDb]

	return fmt.Sprintf("postgres://%s:%s@postgres:5432/%s", user, password, db)
}

func (c Config) mailhogServiceEnvs() env {
	authEnv := c.authServiceEnvs()

	e := env{
		"SMTP_HOST":   authEnv[envAuthSmtpHost],
		"SMTP_PORT":   authEnv[envAuthSmtpPort],
		"SMTP_PASS":   authEnv[envAuthSmtpPass],
		"SMTP_USER":   authEnv[envAuthSmtpUser],
		"SMTP_SECURE": authEnv[envAuthSmtpSecure],
		"SMTP_SENDER": authEnv[envAuthSmtpSender],
	}

	e.merge(c.serviceConfigEnvs(svcMailhog))
	return e
}

func (c Config) runMailhogService() bool {
	if conf, ok := c.nhostConfig.Services[svcMailhog]; ok {
		if conf.NoContainer {
			return false
		}
	}

	authEnv := c.authServiceEnvs()

	return authEnv[envAuthSmtpHost] == svcMailhog
}

func (c Config) mailhogService() *types.ServiceConfig {
	if !c.runMailhogService() {
		return nil
	}

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
				Source: c.hostDataDirectory(dataDirMailhog),
				Target: "/maildir",
			},
		},
	}
}

func (c Config) minioServiceEnvs() env {
	e := env{
		envMinioRootUser:     envMinioRootUserDefaultValue,
		envMinioRootPassword: envMinioRootPasswordDefaultValue,
	}
	e.merge(c.serviceConfigEnvs(svcMinio))
	return e
}

func (c Config) runMinioService() bool {
	if conf, ok := c.nhostConfig.Services[svcMinio]; ok {
		if conf.NoContainer {
			return false
		}
	}

	return true
}

func (c Config) minioService() *types.ServiceConfig {
	if !c.runMinioService() {
		return nil
	}

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
				Source: c.hostDataDirectory(dataDirMinio),
				Target: "/data",
			},
		},
	}
}

func (c Config) functionsServiceEnvs() env {
	e := env{"NHOST_BACKEND_URL": c.envValueNhostBackendUrl()}
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
	minioEnv := c.minioServiceEnvs()
	s3Endpoint := "http://minio:8484" // TODO: port

	if minioConf, ok := c.nhostConfig.Services[svcMinio]; ok {
		if minioConf.NoContainer {
			s3Endpoint = minioConf.Address
		}
	}

	e := env{
		"BIND":                        ":8576",                 // TODO: port
		"PUBLIC_URL":                  "http://localhost:8576", // TODO: port
		"POSTGRES_MIGRATIONS":         "1",
		"HASURA_METADATA":             "1",
		"HASURA_ENDPOINT":             c.envValueHasuraEndpoint(),
		"HASURA_GRAPHQL_ADMIN_SECRET": util.ADMIN_SECRET,
		"S3_ACCESS_KEY":               minioEnv[envMinioRootUser],
		"S3_SECRET_KEY":               minioEnv[envMinioRootPassword],
		"S3_ENDPOINT":                 s3Endpoint,
		"S3_BUCKET":                   "nhost",
		"HASURA_GRAPHQL_JWT_SECRET":   c.envValueHasuraGraphqlJwtSecret(),
		"NHOST_JWT_SECRET":            c.envValueHasuraGraphqlJwtSecret(),
		"NHOST_ADMIN_SECRET":          util.ADMIN_SECRET,
		"NHOST_WEBHOOK_SECRET":        util.WEBHOOK_SECRET,
		"POSTGRES_MIGRATIONS_SOURCE":  fmt.Sprintf("%s?sslmode=disable", c.postgresConnectionString()),
		"NHOST_BACKEND_URL":           c.envValueNhostBackendUrl(),
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
		Expose:      []string{"8000"}, // TODO
	}
}

func (c Config) authServiceEnvs() env {
	hasuraPort := c.serviceDockerExposePort(svcHasura, svcHasuraDefaultPort)

	e := env{
		"AUTH_HOST":                   "0.0.0.0",
		"HASURA_GRAPHQL_DATABASE_URL": c.postgresConnectionString(),
		"HASURA_GRAPHQL_GRAPHQL_URL":  fmt.Sprintf("http://graphql-engine:%d/v1/graphql", hasuraPort),
		"HASURA_GRAPHQL_JWT_SECRET":   c.envValueHasuraGraphqlJwtSecret(),
		"HASURA_GRAPHQL_ADMIN_SECRET": util.ADMIN_SECRET,
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
		//Expose:      []string{"4000"}, // TODO: is it needed?
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
				Source: filepath.Join(util.WORKING_DIR, ".nhost/custom"),
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
	return "http://localhost:1337" // TODO: port
}

func (c Config) envValueHasuraGraphqlJwtSecret() string {
	return fmt.Sprintf(`{"type":"HS256", "key": "%s"}`, util.JWT_KEY)
}

func (c Config) envValueHasuraEndpoint() string {
	hasuraPort := c.serviceDockerExposePort(svcHasura, svcHasuraDefaultPort)
	return fmt.Sprintf("http://graphql-engine:%d/v1", hasuraPort)
}

func (c Config) hasuraServiceEnvs() env {
	e := env{
		"HASURA_GRAPHQL_DATABASE_URL":              c.postgresConnectionString(),
		"HASURA_GRAPHQL_JWT_SECRET":                c.envValueHasuraGraphqlJwtSecret(),
		"HASURA_GRAPHQL_ADMIN_SECRET":              util.ADMIN_SECRET,
		"NHOST_ADMIN_SECRET":                       util.ADMIN_SECRET,
		"NHOST_BACKEND_URL":                        c.envValueNhostBackendUrl(),
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

	port := c.serviceDockerExposePort(svcHasura, svcHasuraDefaultPort)

	return &types.ServiceConfig{
		Name:        svcGraphqlEngine,
		Image:       c.serviceDockerImage(svcHasura, svcHasuraDefaultImage),
		Environment: c.hasuraServiceEnvs().dockerServiceConfigEnv(),
		//Expose:      []string{"8080"},
		Labels: labels,
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    svcHasuraDefaultPort,
				Published: fmt.Sprint(port),
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
	hasuraPort := c.serviceDockerExposePort(svcHasura, svcHasuraDefaultPort)

	return env{
		"HASURA_GRAPHQL_DATABASE_URL":              c.postgresConnectionString(),
		"HASURA_GRAPHQL_JWT_SECRET":                c.envValueHasuraGraphqlJwtSecret(),
		"HASURA_GRAPHQL_ADMIN_SECRET":              util.ADMIN_SECRET,
		"HASURA_GRAPHQL_ENDPOINT":                  fmt.Sprintf("http://127.0.0.1:%d", hasuraPort),
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
		"traefik.http.services.hasura-console.loadbalancer.server.port": "9695", // TODO: port
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
				Target:    9695, // TODO: port
				Published: "9695",
				Protocol:  "tcp",
			},
			{
				Mode:      "ingress",
				Target:    9693, // TODO: port
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
	e := env{
		envPostgresData:     envPostgresDataDefaultValue,
		envPostgresUser:     envPostgresUserDefaultValue,
		envPostgresPassword: envPostgresPasswordDefaultValue,
		envPostgresDb:       envPostgresDbDefaultValue,
	}

	e.merge(c.serviceConfigEnvs(svcPostgres))

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
	port := c.serviceDockerExposePort(svcPostgres, svcPostgresDefaultPort)

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
				Source: c.hostDataDirectoryBranchScoped(dataDirDb),
				Target: envPostgresDataDefaultValue,
			},
		},
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    port,
				Published: fmt.Sprint(port),
				Protocol:  "tcp",
			},
		},
	}
}

func (c Config) serverPort() uint32 {
	return serverDefaultPort
}

func (c Config) traefikService() *types.ServiceConfig {
	port := c.serverPort()
	return &types.ServiceConfig{
		Name:    svcTraefik,
		Image:   c.serviceDockerImage(svcTraefik, svcTraefikDefaultImage),
		Restart: types.RestartPolicyAlways,
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				Target:    serverDefaultPort,
				Published: fmt.Sprint(port),
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
			fmt.Sprintf("--entrypoints.web.address=:%d", serverDefaultPort),
		},
	}
}
