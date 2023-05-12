package dockercompose

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nhost/be/services/mimir/model"
)

const (
	authPort      = 4000
	dashboardPort = 3000
	storagePort   = 8000
	functionsPort = 3000
	hasuraPort    = 8080
	consolePort   = 9695
	migratePort   = 1337
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

func traefik(projectName string, port uint) *Service {
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
		},
		WorkingDir: nil,
	}
}

func auth(cfg *model.ConfigConfig, useTLS bool, nhostFolder string) *Service { //nolint:funlen
	envVars := map[string]string{
		"AUTH_ACCESS_CONTROL_ALLOWED_EMAIL_DOMAINS": "",
		"AUTH_ACCESS_CONTROL_ALLOWED_EMAILS":        "",
		"AUTH_ACCESS_CONTROL_ALLOWED_REDIRECT_URLS": "",
		"AUTH_ACCESS_CONTROL_BLOCKED_EMAIL_DOMAINS": "",
		"AUTH_ACCESS_CONTROL_BLOCKED_EMAILS":        "",
		"AUTH_ACCESS_TOKEN_EXPIRES_IN":              "901",
		"AUTH_ANONYMOUS_USERS_ENABLED":              "false",
		"AUTH_CLIENT_URL":                           "",
		"AUTH_DISABLE_NEW_USERS":                    "false",
		"AUTH_EMAIL_PASSWORDLESS_ENABLED":           "true",
		"AUTH_EMAIL_SIGNIN_EMAIL_VERIFIED_REQUIRED": "false",
		"AUTH_GRAVATAR_DEFAULT":                     "blank",
		"AUTH_GRAVATAR_ENABLED":                     "true",
		"AUTH_GRAVATAR_RATING":                      "g",
		"AUTH_HOST":                                 "0.0.0.0",
		"AUTH_JWT_CUSTOM_CLAIMS":                    "{}",
		"AUTH_LOCALE_ALLOWED_LOCALES":               "en",
		"AUTH_LOCALE_DEFAULT":                       "en",
		"AUTH_MFA_ENABLED":                          "false",
		"AUTH_MFA_TOTP_ISSUER":                      "nhost",
		"AUTH_PASSWORD_HIBP_ENABLED":                "false",
		"AUTH_PASSWORD_MIN_LENGTH":                  "8",
		"AUTH_REFRESH_TOKEN_EXPIRES_IN":             "2592000",
		"AUTH_SERVER_URL":                           "http://auth:4000",
		"AUTH_SMS_PASSWORDLESS_ENABLED":             "false",
		"AUTH_SMS_PROVIDER":                         "",
		"AUTH_SMS_TWILIO_ACCOUNT_SID":               "",
		"AUTH_SMS_TWILIO_AUTH_TOKEN":                "",
		"AUTH_SMS_TWILIO_MESSAGING_SERVICE_ID":      "",
		"AUTH_SMTP_AUTH_METHOD":                     "LOGIN",
		"AUTH_SMTP_HOST":                            "",
		"AUTH_SMTP_PASS":                            "",
		"AUTH_SMTP_PORT":                            "465",
		"AUTH_SMTP_SECURE":                          "true",
		"AUTH_SMTP_SENDER":                          "",
		"AUTH_SMTP_USER":                            "apikey",
		"AUTH_USER_DEFAULT_ALLOWED_ROLES":           "user",
		"AUTH_USER_DEFAULT_ROLE":                    "user",
		"AUTH_WEBAUTHN_ATTESTATION_TIMEOUT":         "0",
		"AUTH_WEBAUTHN_ENABLED":                     "false",
		"AUTH_WEBAUTHN_RP_NAME":                     "VSM NFT Staging",
		"AUTH_WEBAUTHN_RP_ORIGINS":                  "",
		"HASURA_GRAPHQL_ADMIN_SECRET":               cfg.Hasura.AdminSecret,
		"HASURA_GRAPHQL_DATABASE_URL":               "postgres://nhost_auth_admin@local.db.nhost.run:5432/postgres",
		"HASURA_GRAPHQL_GRAPHQL_URL":                "http://graphql:8080/v1/graphql",
		"HASURA_GRAPHQL_JWT_SECRET":                 `{"type":"HS256", "key": "0ac0b96af3f247ddbe95c380e70533a0cd6131b3e276ec91cd9971b6c9d2867fa18beb181ab777c450a7080acb742b32fbaeb1e5c782097553a1fa363775f1ba"}`, //nolint:lll
		"NHOST_ADMIN_SECRET":                        cfg.Hasura.AdminSecret,
		"NHOST_AUTH_URL":                            "https://local.auth.nhost.run/v1",
		"NHOST_FUNCTIONS_URL":                       "https://local.functions.nhost.run/v1",
		"NHOST_GRAPHQL_URL":                         "http://local.graphql.nhost.run/v1",
		"NHOST_HASURA_URL":                          "https://local.hasura.nhost.run/console",
		"NHOST_JWT_SECRET":                          `{"type":"HS256", "key": "1ac0b96af3f247ddbe95c380e70533a0cd6131b3e276ec91cd9971b6c9d2867fa18beb181ab777c450a7080acb742b32fbaeb1e5c782097553a1fa363775f1bb"}`, //nolint:lll
		"NHOST_REGION":                              "",
		"NHOST_STORAGE_URL":                         "https://local.storage.nhost.run/v1",
		"NHOST_SUBDOMAIN":                           "local",
		"NHOST_WEBHOOK_SECRET":                      "472680a19ef9332aad4be3390cfaffb1",
	}
	for _, envVar := range cfg.GetGlobal().GetEnvironment() {
		envVars[envVar.GetName()] = envVar.GetValue()
	}
	return &Service{
		Image: fmt.Sprintf("nhost/hasura-auth:%s", *cfg.Auth.Version),
		DependsOn: map[string]DependsOn{
			"graphql": {
				Condition: "service_healthy",
			},
			"postgres": {
				Condition: "service_healthy",
			},
		},
		EntryPoint:  nil,
		Command:     nil,
		Environment: envVars,
		ExtraHosts:  extraHosts(),
		HealthCheck: &HealthCheck{
			Test:        []string{"CMD", "wget", "--spider", "-S", "http://localhost:4000/healthz"},
			Interval:    "5s",
			StartPeriod: "60s",
		},
		Labels: Ingresses{
			{
				Name: "auth",
				TLS:  useTLS,
				Rule: "Host(`local.auth.nhost.run`) && PathPrefix(`/v1`)",
				Port: authPort,
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
				Source: fmt.Sprintf("%s/emails", nhostFolder),
				Target: "/app/email-templates",
			},
		},
		WorkingDir: nil,
	}
}

func postgres(
	cfg *model.ConfigConfig,
	port uint,
	dataFolder string,
) (*Service, error) { //nolint:funlen
	if err := os.MkdirAll(fmt.Sprintf("%s/db/pgdata", dataFolder), 0o755); err != nil { //nolint:gomnd
		return nil, fmt.Errorf("failed to create postgres data folder: %w", err)
	}

	f, err := os.Create(fmt.Sprintf("%s/db/pg_hba_local.conf", dataFolder))
	if err != nil {
		return nil, fmt.Errorf("failed to create pg_hba_local.conf: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(
		"local all all trust\nhost all all all trust\n", //nolint:dupword
	); err != nil {
		return nil, fmt.Errorf("failed to write to pg_hba_local.conf: %w", err)
	}

	return &Service{
		Image:      fmt.Sprintf("nhost/postgres:%s", *cfg.GetPostgres().GetVersion()),
		DependsOn:  nil,
		EntryPoint: nil,
		Command: []string{
			"postgres",
			"-c", "config_file=/etc/postgresql.conf",
			"-c", "hba_file=/etc/pg_hba_local.conf",
		},
		Environment: map[string]string{
			"PGDATA":            "/var/lib/postgresql/data/pgdata",
			"POSTGRES_DB":       "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_USER":     "postgres",
			"NHOST_SUBDOMAIN":   "local",
		},
		ExtraHosts: extraHosts(),
		HealthCheck: &HealthCheck{
			Test: []string{
				"CMD-SHELL", "pg_isready -U postgres", "-d", "postgres", "-q",
			},
			Interval:    "5s",
			StartPeriod: "60s",
		},
		Labels: nil,
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
				Type:   "bind",
				Source: fmt.Sprintf("%s/db/pgdata", dataFolder),
				Target: "/var/lib/postgresql/data/pgdata",
			},
			{
				Type:   "bind",
				Source: fmt.Sprintf("%s/db/pg_hba_local.conf", dataFolder),
				Target: "/etc/pg_hba_local.conf",
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

func storage(cfg *model.ConfigConfig, useTLS bool) *Service {
	return &Service{
		Image: fmt.Sprintf("nhost/hasura-storage:%s", *cfg.GetStorage().GetVersion()),
		DependsOn: map[string]DependsOn{
			"minio": {
				Condition: "service_started",
			},
			"graphql": {
				Condition: "service_healthy",
			},
			"postgres": {
				Condition: "service_healthy",
			},
		},
		EntryPoint: nil,
		Command: []string{
			"serve",
		},
		Environment: map[string]string{
			"BIND":                        ":8000",
			"DEBUG":                       "true",
			"HASURA_ENDPOINT":             "http://graphql:8080/v1",
			"HASURA_GRAPHQL_ADMIN_SECRET": cfg.GetHasura().GetAdminSecret(),
			"HASURA_METADATA":             "1",
			"S3_ACCESS_KEY":               "minioaccesskey123123",
			"S3_BUCKET":                   "nhost",
			"S3_ENDPOINT":                 "http://minio:9000",
			"S3_SECRET_KEY":               "minioaccesskey123123",
			"POSTGRES_MIGRATIONS":         "1",
			"POSTGRES_MIGRATIONS_SOURCE":  "postgres://nhost_storage_admin@local.db.nhost.run:5432/postgres?sslmode=disable",
			"PUBLIC_URL":                  "https://local.storage.nhost.run",
		},
		ExtraHosts: extraHosts(),
		Labels: Ingresses{
			{
				Name:    "storage",
				TLS:     useTLS,
				Rule:    "PathPrefix(`/v1`) && Host(`local.storage.nhost.run`)",
				Port:    storagePort,
				Rewrite: nil,
			},
		}.Labels(),
		Ports:       nil,
		Restart:     "always",
		HealthCheck: nil,
		Volumes:     nil,
		WorkingDir:  nil,
	}
}

func graphql(cfg *model.ConfigConfig, useTLS bool) (*Service, error) { //nolint:funlen
	jwtSecret, err := json.Marshal(cfg.GetHasura().GetJwtSecrets()[0])
	if err != nil {
		return nil, fmt.Errorf("problem marshalling JWT secret: %w", err)
	}
	envVars := map[string]string{
		"HASURA_GRAPHQL_CORS_DOMAIN":               "*",
		"HASURA_GRAPHQL_ADMIN_SECRET":              cfg.GetHasura().GetAdminSecret(),
		"HASURA_GRAPHQL_DATABASE_URL":              "postgres://nhost_hasura@local.db.nhost.run:5432/postgres",
		"HASURA_GRAPHQL_DEV_MODE":                  "true",
		"HASURA_GRAPHQL_ENABLE_CONSOLE":            "true",
		"HASURA_GRAPHQL_ENABLE_TELEMETRY":          "false",
		"HASURA_GRAPHQL_EVENTS_HTTP_POOL_SIZE":     "100",
		"HASURA_GRAPHQL_JWT_SECRET":                string(jwtSecret),
		"HASURA_GRAPHQL_LOG_LEVEL":                 "info",
		"HASURA_GRAPHQL_MIGRATIONS_SERVER_TIMEOUT": "20",
		"HASURA_GRAPHQL_NO_OF_RETRIES":             "20",
		"HASURA_GRAPHQL_UNAUTHORIZED_ROLE":         "public",
		"NHOST_ADMIN_SECRET":                       cfg.GetHasura().GetAdminSecret(),
		"NHOST_AUTH_URL":                           "https://local.auth.nhost.run/v1",
		"NHOST_BACKEND_URL":                        "http://traefik:1337",
		"NHOST_FUNCTIONS_URL":                      "https://local.functions.nhost.run/v1",
		"NHOST_GRAPHQL_URL":                        "https://local.graphql.nhost.run/v1",
		"NHOST_HASURA_URL":                         "https://local.hasura.nhost.run/console",
		"NHOST_JWT_SECRET":                         string(jwtSecret),
		"NHOST_REGION":                             "",
		"NHOST_STORAGE_URL":                        "https://local.storage.nhost.run/v1",
		"NHOST_SUBDOMAIN":                          "local",
		"NHOST_WEBHOOK_SECRET":                     cfg.GetHasura().GetWebhookSecret(),
	}
	for _, envVar := range cfg.GetGlobal().GetEnvironment() {
		envVars[envVar.GetName()] = envVar.GetValue()
	}

	return &Service{
		Image: fmt.Sprintf("hasura/graphql-engine:%s", *cfg.GetHasura().GetVersion()),
		DependsOn: map[string]DependsOn{
			"postgres": {
				Condition: "service_healthy",
			},
		},
		EntryPoint:  nil,
		Command:     nil,
		Environment: envVars,
		ExtraHosts:  extraHosts(),
		HealthCheck: &HealthCheck{
			Test: []string{
				"CMD-SHELL",
				"curl http://localhost:8080/healthz > /dev/null 2>&1",
			},
			Interval:    "5s",
			StartPeriod: "60s",
		},
		Labels: Ingresses{
			{
				Name: "graphql",
				TLS:  useTLS,
				Rule: "PathPrefix(`/v1`) && Host(`local.graphql.nhost.run`)",
				Port: hasuraPort,
				Rewrite: &Rewrite{
					Regex:       "/v1(/|$)(.*)",
					Replacement: "/v1/graphql$2",
				},
			},
			{
				Name:    "hasurav1",
				TLS:     useTLS,
				Rule:    "PathPrefix(`/v1`) && Host(`local.hasura.nhost.run`)",
				Port:    hasuraPort,
				Rewrite: nil,
			},
			{
				Name:    "hasurav2",
				TLS:     useTLS,
				Rule:    "PathPrefix(`/v2`) && Host(`local.hasura.nhost.run`)",
				Port:    hasuraPort,
				Rewrite: nil,
			},
		}.Labels(),
		Ports:      nil,
		Restart:    "always",
		Volumes:    nil,
		WorkingDir: nil,
	}, nil
}

func console(
	cfg *model.ConfigConfig,
	port uint,
	useTLS bool,
	nhostFolder string,
) (*Service, error) { //nolint:funlen
	graphql, err := graphql(cfg, useTLS)
	if err != nil {
		return nil, err
	}

	graphql.Image += ".cli-migrations-v3"
	graphql.DependsOn = map[string]DependsOn{
		"graphql": {
			Condition: "service_healthy",
		},
	}

	hgeURL := fmt.Sprintf("http://local.hasura.nhost.run:%d", port)
	if useTLS {
		hgeURL = fmt.Sprintf("https://local.hasura.nhost.run:%d", port)
	}
	graphql.Command = []string{
		"bash", "-c",
		fmt.Sprintf(`
		hasura-cli \
		console \
		--no-browser \
		--endpoint http://graphql:8080 \
		--address 0.0.0.0 \
		--console-port 9695 \
		--api-port %d \
		--api-host http://local.hasura.nhost.run \
		--console-hge-endpoint %s \
        `, port, hgeURL),
	}

	graphql.Volumes = []Volume{
		{
			Type:     "bind",
			Source:   nhostFolder,
			Target:   "/app",
			ReadOnly: new(bool),
		},
	}

	graphql.ExtraHosts = []string{
		"host.docker.internal:host-gateway",
		"local.auth.nhost.run:host-gateway",
		"local.db.nhost.run:host-gateway",
		"local.functions.nhost.run:host-gateway",
		"local.graphql.nhost.run:host-gateway",
		"local.hasura.nhost.run:0.0.0.0",
		"local.storage.nhost.run:host-gateway",
	}

	graphql.WorkingDir = ptr("/app")

	graphql.Labels = Ingresses{
		{
			Name:    "console",
			TLS:     useTLS,
			Rule:    "Host(`local.hasura.nhost.run`)",
			Port:    consolePort,
			Rewrite: nil,
		},
		{
			Name:    "migrate",
			TLS:     useTLS,
			Rule:    "PathPrefix(`/apis/`) && Host(`local.hasura.nhost.run`)",
			Port:    migratePort,
			Rewrite: nil,
		},
	}.Labels()

	graphql.HealthCheck = &HealthCheck{
		Test: []string{
			"CMD-SHELL",
			"timeout 1s bash -c ':> /dev/tcp/127.0.0.1/9695' || exit 1",
		},
		Interval:    "5s",
		StartPeriod: "60s",
	}

	return graphql, nil
}

func dashboard(cfg *model.ConfigConfig, useTLS bool) *Service {
	return &Service{
		Image:      "nhost/dashboard:0.14.1",
		DependsOn:  nil,
		EntryPoint: nil,
		Command:    nil,
		Environment: map[string]string{
			"NEXT_PUBLIC_NHOST_ADMIN_SECRET":              cfg.Hasura.AdminSecret,
			"NEXT_PUBLIC_NHOST_AUTH_URL":                  "http://local.auth.nhost.run:1337/v1",
			"NEXT_PUBLIC_NHOST_FUNCTIONS_URL":             "http://local.functions.nhost.run:1337/v1",
			"NEXT_PUBLIC_NHOST_GRAPHQL_URL":               "http://local.graphql.nhost.run:1337/v1",
			"NEXT_PUBLIC_NHOST_HASURA_API_URL":            "http://local.hasura.nhost.run:1337",
			"NEXT_PUBLIC_NHOST_HASURA_CONSOLE_URL":        "http://local.hasura.nhost.run:1337/console",
			"NEXT_PUBLIC_NHOST_HASURA_MIGRATIONS_API_URL": "http://local.hasura.nhost.run:1337",
			"NEXT_PUBLIC_NHOST_STORAGE_URL":               "http://local.storage.nhost.run:1337/v1",
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

func functions(
	cfg *model.ConfigConfig,
	useTLS bool,
	functionsFolder string,
) *Service { //nolint:funlen
	envVars := map[string]string{
		"HASURA_GRAPHQL_ADMIN_SECRET": cfg.Hasura.AdminSecret,
		"HASURA_GRAPHQL_DATABASE_URL": "postgres://nhost_auth_admin@local.db.nhost.run:5432/postgres",
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
				Source: functionsFolder,
				Target: "/opt/project/functions",
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
		WorkingDir: new(string),
	}
}

func ComposeFileFromConfig(
	cfg *model.ConfigConfig,
	projectName string,
	port uint,
	useTLS bool,
	postgresPort uint,
	dataFolder string,
	nhostFolder string,
	functionsFolder string,
) (*ComposeFile, error) {
	minio, err := minio(dataFolder)
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

	console, err := console(cfg, port, useTLS, nhostFolder)
	if err != nil {
		return nil, err
	}

	c := &ComposeFile{
		Version: "3.8",
		Services: map[string]*Service{
			"auth":      auth(cfg, useTLS, nhostFolder),
			"console":   console,
			"dashboard": dashboard(cfg, useTLS),
			"functions": functions(cfg, useTLS, functionsFolder),
			"graphql":   graphql,
			"minio":     minio,
			"postgres":  postgres,
			"storage":   storage(cfg, useTLS),
			"traefik":   traefik(projectName, port),
		},
		Volumes: map[string]struct{}{
			"functions_node_modules": {},
			"root_node_modules":      {},
		},
	}
	return c, nil
}
