package dockercompose

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/ssl"
)

const (
	authPort      = 4000
	mailhogPort   = 8025
	dashboardPort = 3000
	storagePort   = 5000
	functionsPort = 3000
	hasuraPort    = 8080
	consolePort   = 9695
	postgresPort  = 5432
)

const (
	minimumHasuraVerson = "v2.18.0"
)

func ptr[T any](v T) *T {
	return &v
}

type ComposeFile struct {
	Version  string              `yaml:"version"`
	Services map[string]*Service `yaml:"services"`
	Volumes  map[string]struct{} `yaml:"volumes"`
}

//nolint:tagliatelle
type Service struct {
	Image       string               `yaml:"image"`
	DependsOn   map[string]DependsOn `yaml:"depends_on,omitempty"`
	EntryPoint  []string             `yaml:"entrypoint,omitempty"`
	Command     []string             `yaml:"command,omitempty"`
	Environment map[string]string    `yaml:"environment,omitempty"`
	ExtraHosts  []string             `yaml:"extra_hosts"`
	HealthCheck *HealthCheck         `yaml:"healthcheck,omitempty"`
	Labels      map[string]string    `yaml:"labels,omitempty"`
	Ports       []Port               `yaml:"ports,omitempty"`
	Restart     string               `yaml:"restart"`
	Volumes     []Volume             `yaml:"volumes,omitempty"`
	WorkingDir  *string              `yaml:"working_dir,omitempty"`
}

type DependsOn struct {
	Condition string `yaml:"condition"`
}

//nolint:tagliatelle
type HealthCheck struct {
	Test        []string `yaml:"test"`
	Interval    string   `yaml:"interval"`
	StartPeriod string   `yaml:"start_period"`
}

type Port struct {
	Mode      string `yaml:"mode"`
	Target    uint   `yaml:"target"`
	Published string `yaml:"published"`
	Protocol  string `yaml:"protocol"`
}

//nolint:tagliatelle
type Volume struct {
	Type     string `yaml:"type"`
	Source   string `yaml:"source"`
	Target   string `yaml:"target"`
	ReadOnly *bool  `yaml:"read_only,omitempty"`
}

func extraHosts() []string {
	return []string{
		"host.docker.internal:host-gateway",
		"local.auth.nhost.run:host-gateway",
		"local.db.nhost.run:host-gateway",
		"local.functions.nhost.run:host-gateway",
		"local.graphql.nhost.run:host-gateway",
		"local.hasura.nhost.run:host-gateway",
		"local.storage.nhost.run:host-gateway",
	}
}

const traefikConfig = `
# v1
# DO NOT EDIT THIS FILE
tls:
  certificates:
    - certFile: /opt/traefik/certs/local.crt
      keyFile: /opt/traefik/certs/local.key
log:
  level: DEBUG
accessLog: {}
`

func trafikFiles(dotnhostfolder string) error {
	if err := os.MkdirAll(filepath.Join(dotnhostfolder, "traefik", "certs"), 0o755); err != nil { //nolint:gomnd
		return fmt.Errorf("failed to create traefik folder: %w", err)
	}

	f1, err := os.OpenFile(
		filepath.Join(dotnhostfolder, "traefik", "certs", "local.crt"),
		os.O_TRUNC|os.O_CREATE|os.O_WRONLY,
		0o644, //nolint:gomnd
	)
	if err != nil {
		return fmt.Errorf("failed to open local.crt: %w", err)
	}
	defer f1.Close()

	if _, err := f1.Write(ssl.CertFile); err != nil {
		return fmt.Errorf("failed to write local.crt: %w", err)
	}

	f2, err := os.OpenFile(
		filepath.Join(dotnhostfolder, "traefik", "certs", "local.key"),
		os.O_TRUNC|os.O_CREATE|os.O_WRONLY,
		0o644, //nolint:gomnd
	)
	if err != nil {
		return fmt.Errorf("failed to open local.key: %w", err)
	}
	defer f2.Close()

	if _, err := f2.Write(ssl.KeyFile); err != nil {
		return fmt.Errorf("failed to write local.cert: %w", err)
	}

	f3, err := os.OpenFile(
		filepath.Join(dotnhostfolder, "traefik", "traefik.yaml"),
		os.O_TRUNC|os.O_CREATE|os.O_WRONLY,
		0o644, //nolint:gomnd
	)
	if err != nil {
		return fmt.Errorf("failed to open traefik.yaml: %w", err)
	}
	defer f3.Close()

	if _, err := f3.Write([]byte(traefikConfig)); err != nil {
		return fmt.Errorf("failed to write traefik.yaml: %w", err)
	}

	return nil
}

func traefik(projectName string, port uint, dotnhostfolder string) (*Service, error) {
	if err := trafikFiles(dotnhostfolder); err != nil {
		return nil, fmt.Errorf("failed to create traefik files: %w", err)
	}

	return &Service{
		Image:      "traefik:v2.8",
		DependsOn:  nil,
		EntryPoint: nil,
		Command: []string{
			"--api.insecure=true",
			"--providers.docker=true",
			"--providers.file.directory=/opt/traefik",
			"--providers.file.watch=true",
			"--providers.docker.exposedbydefault=false",
			fmt.Sprintf(
				"--providers.docker.constraints=Label(`com.docker.compose.project`,`%s`)",
				projectName,
			),
			fmt.Sprintf("--entrypoints.web.address=:%d", port),
		},
		Environment: nil,
		ExtraHosts:  extraHosts(),
		HealthCheck: nil,
		Labels:      nil,
		Ports: []Port{
			{
				Mode:      "ingress",
				Target:    port,
				Published: fmt.Sprintf("%d", port),
				Protocol:  "tcp",
			},
		},
		Restart: "always",
		Volumes: []Volume{
			{
				Type:     "bind",
				Source:   "/var/run/docker.sock",
				Target:   "/var/run/docker.sock",
				ReadOnly: ptr(true),
			},
			{
				Type:     "bind",
				Source:   filepath.Join(dotnhostfolder, "traefik"),
				Target:   "/opt/traefik",
				ReadOnly: ptr(true),
			},
		},
		WorkingDir: nil,
	}, nil
}

func minio(dataFolder string) (*Service, error) {
	if err := os.MkdirAll(fmt.Sprintf("%s/minio", dataFolder), 0o755); err != nil { //nolint:gomnd
		return nil, fmt.Errorf("failed to create minio data folder: %w", err)
	}
	return &Service{
		Image:      "minio/minio:RELEASE.2022-07-08T00-05-23Z",
		DependsOn:  nil,
		EntryPoint: []string{"/bin/sh"},
		Command: []string{
			"-c", "mkdir -p /data/nhost && /opt/bin/minio server --address :9000 /data",
		},
		Environment: map[string]string{
			"MINIO_ROOT_PASSWORD": "minioaccesskey123123",
			"MINIO_ROOT_USER":     "minioaccesskey123123",
		},
		ExtraHosts:  extraHosts(),
		Ports:       nil,
		Restart:     "always",
		HealthCheck: nil,
		Labels:      nil,
		Volumes: []Volume{
			{
				Type:   "bind",
				Source: fmt.Sprintf("%s/minio", dataFolder),
				Target: "/data",
			},
		},
		WorkingDir: nil,
	}, nil
}

func dashboard(cfg *model.ConfigConfig, httpPort uint, useTLS bool) *Service {
	return &Service{
		Image:      "nhost/dashboard:0.16.6",
		DependsOn:  nil,
		EntryPoint: nil,
		Command:    nil,
		Environment: map[string]string{
			"NEXT_PUBLIC_NHOST_ADMIN_SECRET": cfg.Hasura.AdminSecret,
			"NEXT_PUBLIC_NHOST_AUTH_URL":     URL("auth", httpPort, useTLS) + "/v1",
			"NEXT_PUBLIC_NHOST_FUNCTIONS_URL": URL(
				"functions",
				httpPort,
				useTLS,
			) + "/v1",
			"NEXT_PUBLIC_NHOST_GRAPHQL_URL":    URL("graphql", httpPort, useTLS) + "/v1",
			"NEXT_PUBLIC_NHOST_HASURA_API_URL": URL("hasura", httpPort, useTLS),
			"NEXT_PUBLIC_NHOST_HASURA_CONSOLE_URL": URL(
				"hasura",
				httpPort,
				useTLS,
			) + "/console",
			"NEXT_PUBLIC_NHOST_HASURA_MIGRATIONS_API_URL": URL("hasura", httpPort, useTLS),
			"NEXT_PUBLIC_NHOST_STORAGE_URL":               URL("storage", httpPort, useTLS) + "/v1",
		},
		ExtraHosts:  extraHosts(),
		HealthCheck: nil,
		Labels: Ingresses{
			{
				Name:    "dashboard",
				TLS:     useTLS,
				Rule:    "Host(`local.dashboard.nhost.run`)",
				Port:    dashboardPort,
				Rewrite: nil,
			},
		}.Labels(),
		Ports:      []Port{},
		Restart:    "",
		Volumes:    []Volume{},
		WorkingDir: new(string),
	}
}

func functions( //nolint:funlen
	cfg *model.ConfigConfig,
	useTLS bool,
	rootFolder string,
) *Service {
	envVars := map[string]string{
		"HASURA_GRAPHQL_ADMIN_SECRET": cfg.Hasura.AdminSecret,
		"HASURA_GRAPHQL_DATABASE_URL": "postgres://nhost_auth_admin@local.db.nhost.run:5432/local",
		"HASURA_GRAPHQL_GRAPHQL_URL":  "http://graphql:8080/v1/graphql",
		"HASURA_GRAPHQL_JWT_SECRET":   `{"type":"HS256", "key": "0ac0b96af3f247ddbe95c380e70533a0cd6131b3e276ec91cd9971b6c9d2867fa18beb181ab777c450a7080acb742b32fbaeb1e5c782097553a1fa363775f1ba"}`, //nolint:lll
		"NHOST_ADMIN_SECRET":          cfg.Hasura.AdminSecret,
		"NHOST_AUTH_URL":              "https://local.auth.nhost.run/v1",
		"NHOST_FUNCTIONS_URL":         "https://local.functions.nhost.run/v1",
		"NHOST_GRAPHQL_URL":           "http://local.graphql.nhost.run/v1",
		"NHOST_HASURA_URL":            "https://local.hasura.nhost.run/console",
		"NHOST_JWT_SECRET":            `{"type":"HS256", "key": "1ac0b96af3f247ddbe95c380e70533a0cd6131b3e276ec91cd9971b6c9d2867fa18beb181ab777c450a7080acb742b32fbaeb1e5c782097553a1fa363775f1bb"}`, //nolint:lll
		"NHOST_REGION":                "",
		"NHOST_STORAGE_URL":           "https://local.storage.nhost.run/v1",
		"NHOST_SUBDOMAIN":             "local",
		"NHOST_WEBHOOK_SECRET":        "472680a19ef9332aad4be3390cfaffb1",
	}
	for _, envVar := range cfg.GetGlobal().GetEnvironment() {
		envVars[envVar.GetName()] = envVar.GetValue()
	}
	return &Service{
		Image:       "nhost/functions:0.1.8",
		DependsOn:   nil,
		EntryPoint:  nil,
		Command:     nil,
		Environment: envVars,
		ExtraHosts:  extraHosts(),
		HealthCheck: &HealthCheck{
			Test:        []string{"CMD", "wget", "--spider", "-S", "http://localhost:3000/healthz"},
			Interval:    "5s",
			StartPeriod: "60s",
		},
		Labels: Ingresses{
			{
				Name: "functions",
				TLS:  useTLS,
				Rule: "Host(`local.functions.nhost.run`) && PathPrefix(`/v1`)",
				Port: functionsPort,
				Rewrite: &Rewrite{
					Regex:       "/v1(/|$)(.*)",
					Replacement: "/$2",
				},
			},
		}.Labels(),
		Ports:   []Port{},
		Restart: "always",
		Volumes: []Volume{
			{
				Type:   "bind",
				Source: rootFolder,
				Target: "/opt/project",
			},
			{
				Type:   "volume",
				Source: "root_node_modules",
				Target: "/opt/project/node_modules",
			},
			{
				Type:   "volume",
				Source: "functions_node_modules",
				Target: "/opt/project/functions/node_modules",
			},
		},
		WorkingDir: nil,
	}
}

func mailhog(dataFolder string, useTLS bool) (*Service, error) {
	mailhogDataFolder := filepath.Join(dataFolder, "mailhog")
	if err := os.MkdirAll(mailhogDataFolder, 0o755); err != nil { //nolint:gomnd
		return nil, fmt.Errorf("failed to create mailhog folder: %w", err)
	}

	return &Service{
		Image:      "jcalonso/mailhog:v1.0.1",
		DependsOn:  nil,
		EntryPoint: []string{},
		Command:    []string{},
		Environment: map[string]string{
			"SMTP_HOST":   "mailhog",
			"SMTP_PASS":   "password",
			"SMTP_PORT":   "1025",
			"SMTP_SECURE": "false",
			"SMTP_SENDER": "hasura-auth@example.com",
			"SMTP_USER":   "user",
		},
		ExtraHosts:  extraHosts(),
		HealthCheck: nil,
		Labels: Ingresses{
			{
				Name: "mailhog",
				TLS:  useTLS,
				Rule: "Host(`local.mailhog.nhost.run`)",
				Port: mailhogPort,
				// Rewrite: &Rewrite{
				// 	Regex:       "/mailhog(/|$)(.*)",
				// 	Replacement: "/$2",
				// },
			},
		}.Labels(),
		Ports:   nil,
		Restart: "always",
		Volumes: []Volume{
			{
				Type:     "bind",
				Source:   mailhogDataFolder,
				Target:   "/maildir",
				ReadOnly: ptr(false),
			},
		},
		WorkingDir: nil,
	}, nil
}

func ComposeFileFromConfig( //nolint:funlen
	cfg *model.ConfigConfig,
	projectName string,
	httpPort uint,
	useTLS bool,
	postgresPort uint,
	dataFolder string,
	nhostFolder string,
	dotNhostFolder string,
	rootFolder string,
) (*ComposeFile, error) {
	minio, err := minio(dataFolder)
	if err != nil {
		return nil, err
	}

	auth, err := auth(cfg, httpPort, useTLS, nhostFolder)
	if err != nil {
		return nil, err
	}

	storage, err := storage(cfg, useTLS, httpPort)
	if err != nil {
		return nil, err
	}

	postgres, err := postgres(cfg, postgresPort, dataFolder)
	if err != nil {
		return nil, err
	}

	graphql, err := graphql(cfg, useTLS)
	if err != nil {
		return nil, err
	}

	console, err := console(cfg, httpPort, useTLS, nhostFolder)
	if err != nil {
		return nil, err
	}

	traefik, err := traefik(projectName, httpPort, dotNhostFolder)
	if err != nil {
		return nil, err
	}

	mailhog, err := mailhog(dataFolder, useTLS)
	if err != nil {
		return nil, err
	}

	c := &ComposeFile{
		Version: "3.8",
		Services: map[string]*Service{
			"auth":      auth,
			"console":   console,
			"dashboard": dashboard(cfg, httpPort, useTLS),
			"functions": functions(cfg, useTLS, rootFolder),
			"graphql":   graphql,
			"minio":     minio,
			"postgres":  postgres,
			"storage":   storage,
			"mailhog":   mailhog,
			"traefik":   traefik,
		},
		Volumes: map[string]struct{}{
			"functions_node_modules": {},
			"root_node_modules":      {},
		},
	}
	return c, nil
}
