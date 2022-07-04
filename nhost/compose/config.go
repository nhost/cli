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
const postgresDefaultPassword = "postgres"

type Config struct {
	nhostConfig        *nhost.Configuration // nhost configuration to read custom values from, not used atm
	gitBranch          string               // git branch name, used as a namespace for postgres data mounted from host
	composeConfig      *types.Config
	composeProjectName string
}

func NewConfig(conf *nhost.Configuration, gitBranch, projectName string) *Config {
	return &Config{nhostConfig: conf, gitBranch: gitBranch, composeProjectName: projectName}
}

func (c *Config) build() *types.Config {
	config := &types.Config{}

	// set services
	config.Services = []types.ServiceConfig{
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

func (c Config) mailhogService() types.ServiceConfig {
	envs := types.NewMappingWithEquals([]string{
		"SMTP_HOST=${AUTH_SMTP_HOST:-mailhog}",
		"SMTP_PORT=${AUTH_SMTP_PORT:-1025}",
		"SMTP_PASS=${AUTH_SMTP_PASS:-password}",
		"SMTP_USER=${AUTH_SMTP_USER:-user}",
		"SMTP_SECURE=${AUTH_SMTP_SECURE:-false}",
		"SMTP_SENDER=${AUTH_SMTP_SENDER:-hbp@hbp.com}",
	})

	return types.ServiceConfig{
		Name:        "mailhog",
		Environment: envs,
		Restart:     types.RestartPolicyAlways,
		Image:       "mailhog/mailhog",
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

func (c Config) minioService() types.ServiceConfig {
	envs := types.NewMappingWithEquals([]string{
		"MINIO_ROOT_USER=minioaccesskey123123", // TODO: creds
		"MINIO_ROOT_PASSWORD=minioaccesskey123123",
	})

	return types.ServiceConfig{
		Name:        "minio",
		Environment: envs,
		Restart:     types.RestartPolicyAlways,
		Image:       "minio/minio:RELEASE.2021-09-24T00-24-24Z",
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

func (c Config) functionsService() types.ServiceConfig {
	labels := map[string]string{
		"traefik.enable": "true",
		"traefik.http.middlewares.strip-functions.stripprefix.prefixes": "/v1/functions",
		"traefik.http.routers.functions.rule":                           "Host(`localhost`) && PathPrefix(`/v1/functions`)",
		"traefik.http.routers.functions.middlewares":                    "strip-functions@docker",
		"traefik.http.routers.functions.entrypoints":                    "web",
	}

	return types.ServiceConfig{
		Name:    "functions",
		Image:   "nhost/functions", // TODO: build, push & pin version
		Labels:  labels,
		Restart: types.RestartPolicyAlways,
		Expose:  []string{"3000"},
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

func (c Config) storageService() types.ServiceConfig {
	envs := types.NewMappingWithEquals([]string{
		"BIND:8576",                        // TODO: randomize port
		"PUBLIC_URL=http://localhost:8576", // TODO: port
		"POSTGRES_MIGRATIONS=1",
		"HASURA_METADATA=1",
		"HASURA_ENDPOINT=http://graphql-engine:8080/v1",
		fmt.Sprintf("HASURA_GRAPHQL_ADMIN_SECRET=%s", util.ADMIN_SECRET),
		"S3_ACCESS_KEY=minioaccesskey123123",
		"S3_SECRET_KEY=minioaccesskey123123",
		"S3_ENDPOINT=http://minio:8484",
		"S3_BUCKET=nhost",
		"POSTGRES_MIGRATIONS=1",
		fmt.Sprintf("POSTGRES_MIGRATIONS_SOURCE=%s?sslmode=disable", c.postgresConnectionString()),
	})

	labels := map[string]string{
		"traefik.enable":                           "true",
		"traefik.http.routers.storage.rule":        "Host(`localhost`) && PathPrefix(`/v1/storage`)",
		"traefik.http.routers.storage.entrypoints": "web",
		// Rewrite the path so it matches with the new storage API path introduced in hasura-storage 0.2
		"traefik.http.middlewares.strip-suffix.replacepathregex.regex":       "^/v1/storage/(.*)",
		"traefik.http.middlewares.strip-suffix.replacepathregex.replacement": "/v1/$$1",
		"traefik.http.routers.storage.middlewares":                           "strip-suffix@docker",
	}

	return types.ServiceConfig{
		Name:        "storage",
		Restart:     types.RestartPolicyAlways,
		Image:       "nhost/hasura-storage:0.2.1",
		Environment: envs,
		Labels:      labels,
		Command:     []string{"serve"},
		Expose:      []string{"8000"},
	}
}

func (c Config) authService() types.ServiceConfig {
	envs := types.NewMappingWithEquals([]string{
		"AUTH_HOST=0.0.0.0",
		fmt.Sprintf("HASURA_GRAPHQL_DATABASE_URL=%s", c.postgresConnectionString()),
		"HASURA_GRAPHQL_GRAPHQL_URL=http://graphql-engine:8080/v1/graphql",
		fmt.Sprintf("HASURA_GRAPHQL_JWT_SECRET=%s", fmt.Sprintf(`{"type":"HS256", "key": "%s"}`, util.JWT_KEY)),
		fmt.Sprintf("HASURA_GRAPHQL_ADMIN_SECRET=%s", util.ADMIN_SECRET),
		"AUTH_CLIENT_URL=${AUTH_CLIENT_URL:-http://localhost:3000}",
		"AUTH_SMTP_HOST=mailhog",
		"AUTH_SMTP_PORT=1025",
		"AUTH_SMTP_USER=user",
		"AUTH_SMTP_PASS=password",
		"AUTH_SMTP_SENDER=mail@example.com",
	})

	labels := map[string]string{
		"traefik.enable": "true",
		"traefik.http.middlewares.strip-auth.stripprefix.prefixes": "/v1/auth",
		"traefik.http.routers.auth.rule":                           "Host(`localhost`) && PathPrefix(`/v1/auth`)",
		"traefik.http.routers.auth.middlewares":                    "strip-auth@docker",
		"traefik.http.routers.auth.entrypoints":                    "web",
	}

	return types.ServiceConfig{
		Name:        "auth",
		Image:       "nhost/hasura-auth:0.6.3",
		Environment: envs,
		Labels:      labels,
		Expose:      []string{"4000"},
		DependsOn: map[string]types.ServiceDependency{
			"postgres": {
				Condition: types.ServiceConditionHealthy,
			},
			"graphql-engine": {
				Condition: types.ServiceConditionStarted,
			},
		},
		Restart: types.RestartPolicyAlways,
	}
}

func (c Config) hasuraService() types.ServiceConfig {
	// TODO: add envs from .env.development
	// TODO: check whether we need ALL envs from util.RuntimeVars
	envs := types.NewMappingWithEquals([]string{
		fmt.Sprintf("HASURA_GRAPHQL_DATABASE_URL=%s", c.postgresConnectionString()),
		fmt.Sprintf("HASURA_GRAPHQL_JWT_SECRET=%s", fmt.Sprintf(`{"type":"HS256", "key": "%s"}`, util.JWT_KEY)),
		fmt.Sprintf("HASURA_GRAPHQL_ADMIN_SECRET=%s", util.ADMIN_SECRET),
		"HASURA_GRAPHQL_UNAUTHORIZED_ROLE=public",
		"HASURA_GRAPHQL_DEV_MODE=true",
		"HASURA_GRAPHQL_LOG_LEVEL=debug",
		"HASURA_GRAPHQL_ENABLE_CONSOLE=false",
		"HASURA_GRAPHQL_MIGRATIONS_SERVER_TIMEOUT=20",
		"HASURA_GRAPHQL_NO_OF_RETRIES=20",
		"HASURA_GRAPHQL_ENABLE_TELEMETRY=false",
	})

	//labels := map[string]string{
	//	"traefik.enable":                          "true",
	//	"traefik.http.routers.hasura.rule":        "Host(`localhost`) && PathPrefix(`/`)",
	//	"traefik.http.routers.hasura.entrypoints": "web",
	//}

	labels := map[string]string{
		"traefik.enable": "true",
	}

	return types.ServiceConfig{
		Name:        "graphql-engine",
		Image:       "hasura/graphql-engine:v2.2.0",
		Environment: envs,
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
			"postgres": {
				Condition: types.ServiceConditionHealthy,
			},
		},
		Restart: types.RestartPolicyAlways,
	}
}

func (c Config) hasuraConsoleService() types.ServiceConfig {
	envs := types.NewMappingWithEquals([]string{
		fmt.Sprintf("HASURA_GRAPHQL_DATABASE_URL=%s", c.postgresConnectionString()),
		fmt.Sprintf("HASURA_GRAPHQL_JWT_SECRET=%s", fmt.Sprintf(`{"type":"HS256", "key": "%s"}`, util.JWT_KEY)),
		fmt.Sprintf("HASURA_GRAPHQL_ADMIN_SECRET=%s", util.ADMIN_SECRET),
		"HASURA_GRAPHQL_ENDPOINT=http://127.0.0.1:8080",
		"HASURA_GRAPHQL_UNAUTHORIZED_ROLE=public",
		"HASURA_GRAPHQL_DEV_MODE=true",
		"HASURA_GRAPHQL_LOG_LEVEL=debug",
		"HASURA_GRAPHQL_ENABLE_CONSOLE=false",
		"HASURA_RUN_CONSOLE=true",
		"HASURA_GRAPHQL_MIGRATIONS_SERVER_TIMEOUT=20",
		"HASURA_GRAPHQL_NO_OF_RETRIES=20",
		"HASURA_GRAPHQL_ENABLE_TELEMETRY=false",
	})

	labels := map[string]string{
		"traefik.enable": "true",
		"traefik.http.services.hasura-console.loadbalancer.server.port": "9695",
		"traefik.http.routers.hasura-console.rule":                      "Host(`localhost`) && PathPrefix(`/`)",
		"traefik.http.routers.hasura-console.entrypoints":               "web",
	}

	return types.ServiceConfig{
		Name:        "hasura-console",
		Image:       "nhost/hasura:v2.8.1",
		Environment: envs,
		Labels:      labels,
		DependsOn: map[string]types.ServiceDependency{
			"postgres": {
				Condition: types.ServiceConditionHealthy,
			},
			"graphql-engine": {
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

func (c Config) postgresService() types.ServiceConfig {
	envs := types.NewMappingWithEquals([]string{
		fmt.Sprintf("POSTGRES_PASSWORD=%s", c.postgresPasswordEnvValueWithDefaultValue()),
		"POSTGRES_DB=postgres",
		"PGDATA=/var/lib/postgresql/data/pgdata",
	})

	// healthcheck
	interval := types.Duration(time.Second * 3)
	startPeriod := types.Duration(time.Minute * 2)

	return types.ServiceConfig{
		Name:        "postgres",
		Image:       "nhost/postgres:12-v0.0.6",
		Restart:     types.RestartPolicyAlways,
		Environment: envs,
		HealthCheck: &types.HealthCheckConfig{
			Test:        []string{"CMD-SHELL", "pg_isready -U postgres -d postgres -q"},
			Interval:    &interval,
			StartPeriod: &startPeriod,
		},
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   types.VolumeTypeBind,
				Source: c.hostDataDirectoryBranchScoped("db"),
				Target: "/var/lib/postgresql/data/pgdata",
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

func (c Config) traefikService() types.ServiceConfig {
	return types.ServiceConfig{
		Name:    "traefik",
		Image:   "traefik:v2.8",
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
