package dockercompose

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/ssl"
)

const (
	authPort         = 4000
	mailhogPort      = 8025
	dashboardPort    = 3000
	storagePort      = 5000
	functionsPort    = 3000
	hasuraPort       = 8080
	consolePort      = 9695
	postgresPort     = 5432
	configserverPort = 8088
)

const (
	minimumHasuraVerson = "v2.18.0"
)

func rootNodeModules(branch string) string {
	return sanitizeBranch(branch) + "-root_node_modules"
}

func functionsNodeModules(branch string) string {
	return sanitizeBranch(branch) + "-functions_node_modules"
}

func ptr[T any](v T) *T {
	return &v
}

func ports(host, container uint) []Port {
	if host == 0 {
		return nil
	}
	return []Port{
		{
			Mode:      "ingress",
			Target:    container,
			Published: strconv.FormatUint(uint64(host), 10),
			Protocol:  "tcp",
		},
	}
}

type ComposeFile struct {
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
	Timeout     string   `yaml:"timeout"`
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

func extraHosts(subdomain string) []string {
	return []string{
		"host.docker.internal:host-gateway",
		subdomain + ".auth.local.nhost.run:host-gateway",
		subdomain + ".db.local.nhost.run:host-gateway",
		subdomain + ".functions.local.nhost.run:host-gateway",
		subdomain + ".graphql.local.nhost.run:host-gateway",
		subdomain + ".hasura.local.nhost.run:host-gateway",
		subdomain + ".storage.local.nhost.run:host-gateway",
		// below entries shouldn't be needed unless
		// users are hardcoding these subdomains
		// adding out of an abundance of caution
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
    - certFile: /opt/traefik/certs/sub.crt
      keyFile: /opt/traefik/certs/sub.key
log:
  level: DEBUG
accessLog: {}
`

func dumpCert(
	cert []byte,
	key []byte,
	dstName string,
	dotnhostfolder string,
) error {
	f1, err := os.OpenFile(
		filepath.Join(dotnhostfolder, "traefik", "certs", dstName+".crt"),
		os.O_TRUNC|os.O_CREATE|os.O_WRONLY,
		0o644, //nolint:mnd
	)
	if err != nil {
		return fmt.Errorf("failed to open local.crt: %w", err)
	}
	defer f1.Close()

	if _, err := f1.Write(cert); err != nil {
		return fmt.Errorf("failed to write local.crt: %w", err)
	}

	f2, err := os.OpenFile(
		filepath.Join(dotnhostfolder, "traefik", "certs", dstName+".key"),
		os.O_TRUNC|os.O_CREATE|os.O_WRONLY,
		0o644, //nolint:mnd
	)
	if err != nil {
		return fmt.Errorf("failed to open local.key: %w", err)
	}
	defer f2.Close()

	if _, err := f2.Write(key); err != nil {
		return fmt.Errorf("failed to write local.cert: %w", err)
	}

	return nil
}

func trafikFiles(dotnhostfolder string) error {
	if err := os.MkdirAll(filepath.Join(dotnhostfolder, "traefik", "certs"), 0o755); err != nil { //nolint:mnd
		return fmt.Errorf("failed to create traefik folder: %w", err)
	}

	if err := dumpCert(ssl.LocalCertFile, ssl.LocalKeyFile, "local", dotnhostfolder); err != nil {
		return fmt.Errorf("failed to dump local cert: %w", err)
	}

	if err := dumpCert(ssl.SubCertFile, ssl.SubKeyFile, "sub", dotnhostfolder); err != nil {
		return fmt.Errorf("failed to dump sub cert: %w", err)
	}

	f, err := os.OpenFile(
		filepath.Join(dotnhostfolder, "traefik", "traefik.yaml"),
		os.O_TRUNC|os.O_CREATE|os.O_WRONLY,
		0o644, //nolint:mnd
	)
	if err != nil {
		return fmt.Errorf("failed to open traefik.yaml: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(traefikConfig); err != nil {
		return fmt.Errorf("failed to write traefik.yaml: %w", err)
	}

	return nil
}

func traefik(subdomain, projectName string, port uint, dotnhostfolder string) (*Service, error) {
	if err := trafikFiles(dotnhostfolder); err != nil {
		return nil, fmt.Errorf("failed to create traefik files: %w", err)
	}

	return &Service{
		Image:      "traefik:v3.1",
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
		ExtraHosts:  extraHosts(subdomain),
		HealthCheck: nil,
		Labels:      nil,
		Ports: []Port{
			{
				Mode:      "ingress",
				Target:    port,
				Published: strconv.FormatUint(uint64(port), 10),
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

func minio(subdomain, dataFolder string) (*Service, error) {
	if err := os.MkdirAll(dataFolder+"/minio", 0o755); err != nil { //nolint:mnd
		return nil, fmt.Errorf("failed to create minio data folder: %w", err)
	}
	return &Service{
		Image:      "minio/minio:RELEASE.2025-02-28T09-55-16Z",
		DependsOn:  nil,
		EntryPoint: []string{"/bin/sh"},
		Command: []string{
			"-c", "mkdir -p /data/nhost && /opt/bin/minio server --address :9000 /data",
		},
		Environment: map[string]string{
			"MINIO_ROOT_PASSWORD": "minioaccesskey123123",
			"MINIO_ROOT_USER":     "minioaccesskey123123",
		},
		ExtraHosts:  extraHosts(subdomain),
		Ports:       nil,
		Restart:     "always",
		HealthCheck: nil,
		Labels:      nil,
		Volumes: []Volume{
			{
				Type:     "bind",
				Source:   dataFolder + "/minio",
				Target:   "/data",
				ReadOnly: nil,
			},
		},
		WorkingDir: nil,
	}, nil
}

func dashboard(
	cfg *model.ConfigConfig,
	subdomain string,
	dashboardVersion string,
	httpPort uint,
	useTLS bool,
) *Service {
	return &Service{
		Image:      dashboardVersion,
		DependsOn:  nil,
		EntryPoint: nil,
		Command:    nil,
		Environment: map[string]string{
			"NEXT_PUBLIC_ENV":                "dev",
			"NEXT_PUBLIC_NHOST_PLATFORM":     "false",
			"NEXT_PUBLIC_NHOST_ADMIN_SECRET": cfg.Hasura.AdminSecret,
			"NEXT_PUBLIC_NHOST_AUTH_URL": URL(
				subdomain, "auth", httpPort, useTLS) + "/v1",
			"NEXT_PUBLIC_NHOST_CONFIGSERVER_URL": URL(
				subdomain, "dashboard", httpPort, useTLS,
			) + "/v1/configserver/graphql",
			"NEXT_PUBLIC_NHOST_FUNCTIONS_URL": URL(
				subdomain, "functions", httpPort, useTLS,
			) + "/v1",
			"NEXT_PUBLIC_NHOST_GRAPHQL_URL": URL(
				subdomain, "graphql", httpPort, useTLS) + "/v1",
			"NEXT_PUBLIC_NHOST_HASURA_API_URL": URL(subdomain, "hasura", httpPort, useTLS),
			"NEXT_PUBLIC_NHOST_HASURA_CONSOLE_URL": URL(
				subdomain, "hasura", httpPort, useTLS,
			) + "/console",
			"NEXT_PUBLIC_NHOST_HASURA_MIGRATIONS_API_URL": URL(
				subdomain, "hasura", httpPort, useTLS),
			"NEXT_PUBLIC_NHOST_STORAGE_URL": URL(
				subdomain, "storage", httpPort, useTLS) + "/v1",
		},
		ExtraHosts:  extraHosts(subdomain),
		HealthCheck: nil,
		Labels: Ingresses{
			{
				Name:    "dashboard",
				TLS:     useTLS,
				Rule:    traefikHostMatch("dashboard"),
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
	subdomain string,
	httpPort uint,
	useTLS bool,
	rootFolder string,
	jwtSecret string,
	port uint,
	branch string,
) *Service {
	envVars := map[string]string{
		"HASURA_GRAPHQL_ADMIN_SECRET": cfg.Hasura.AdminSecret,
		"HASURA_GRAPHQL_DATABASE_URL": "postgres://nhost_auth_admin@local.db.nhost.run:5432/local",
		"HASURA_GRAPHQL_GRAPHQL_URL":  "http://graphql:8080/v1/graphql",
		"HASURA_GRAPHQL_JWT_SECRET":   jwtSecret,
		"NHOST_ADMIN_SECRET":          cfg.Hasura.AdminSecret,
		"NHOST_AUTH_URL":              URL(subdomain, "auth", httpPort, useTLS) + "/v1",
		"NHOST_FUNCTIONS_URL":         URL(subdomain, "functions", httpPort, useTLS) + "/v1",
		"NHOST_GRAPHQL_URL":           URL(subdomain, "graphql", httpPort, useTLS) + "/v1",
		"NHOST_HASURA_URL":            URL(subdomain, "hasura", httpPort, useTLS) + "/console",
		"NHOST_STORAGE_URL":           URL(subdomain, "storage", httpPort, useTLS) + "/v1",
		"NHOST_JWT_SECRET":            jwtSecret,
		"NHOST_REGION":                "local",
		"NHOST_SUBDOMAIN":             subdomain,
		"NHOST_WEBHOOK_SECRET":        cfg.Hasura.WebhookSecret,
		"GRAPHITE_WEBHOOK_SECRET":     cfg.GetAi().GetWebhookSecret(),
	}
	for _, envVar := range cfg.GetGlobal().GetEnvironment() {
		envVars[envVar.GetName()] = envVar.GetValue()
	}

	return &Service{
		Image:       fmt.Sprintf("nhost/functions:%d-1.4.0", *cfg.GetFunctions().GetNode().Version),
		DependsOn:   nil,
		EntryPoint:  nil,
		Command:     nil,
		Environment: envVars,
		ExtraHosts:  extraHosts(subdomain),
		HealthCheck: &HealthCheck{
			Test:        []string{"CMD", "wget", "--spider", "-S", "http://localhost:3000/healthz"},
			Interval:    "5s",
			Timeout:     "600s",
			StartPeriod: "600s",
		},
		Labels: Ingresses{
			{
				Name: "functions",
				TLS:  useTLS,
				Rule: traefikHostMatch("functions") + "&& PathPrefix(`/v1`)",
				Port: functionsPort,
				Rewrite: &Rewrite{
					Regex:       "/v1(/|$$)(.*)",
					Replacement: "/$$2",
				},
			},
		}.Labels(),
		Ports:   ports(port, functionsPort),
		Restart: "always",
		Volumes: []Volume{
			{
				Type:     "bind",
				Source:   rootFolder,
				Target:   "/opt/project",
				ReadOnly: ptr(false),
			},
			{
				Type:     "volume",
				Source:   rootNodeModules(branch),
				Target:   "/opt/project/node_modules",
				ReadOnly: ptr(false),
			},
			{
				Type:     "volume",
				Source:   functionsNodeModules(branch),
				Target:   "/opt/project/functions/node_modules",
				ReadOnly: ptr(false),
			},
		},
		WorkingDir: nil,
	}
}

func mailhog(subdomain, dataFolder string, useTLS bool) (*Service, error) {
	mailhogDataFolder := filepath.Join(dataFolder, "mailhog")
	if err := os.MkdirAll(mailhogDataFolder, 0o755); err != nil { //nolint:mnd
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
		ExtraHosts:  extraHosts(subdomain),
		HealthCheck: nil,
		Labels: Ingresses{
			{
				Name:    "mailhog",
				TLS:     useTLS,
				Rule:    traefikHostMatch("mailhog"),
				Port:    mailhogPort,
				Rewrite: nil,
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

type ExposePorts struct {
	Auth      uint
	Storage   uint
	Graphql   uint
	Console   uint
	Functions uint
}

func sanitizeBranch(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	return strings.ToLower(re.ReplaceAllString(name, ""))
}

func IsJWTSecretCompatibleWithHasuraAuth( //nolint:cyclop
	jwtSecret *model.ConfigJWTSecret,
) bool {
	if jwtSecret != nil && jwtSecret.Type != nil && *jwtSecret.Type != "" && jwtSecret.Key != nil &&
		*jwtSecret.Key != "" {
		return *jwtSecret.Type == "HS256" || *jwtSecret.Type == "HS384" ||
			*jwtSecret.Type == "HS512" || *jwtSecret.Type == "RS256" ||
			*jwtSecret.Type == "RS384" || *jwtSecret.Type == "RS512"
	}
	return false
}

func getServices( //nolint: funlen,cyclop
	cfg *model.ConfigConfig,
	subdomain string,
	projectName string,
	httpPort uint,
	useTLS bool,
	postgresPort uint,
	dataFolder string,
	nhostFolder string,
	dotNhostFolder string,
	rootFolder string,
	ports ExposePorts,
	branch string,
	dashboardVersion string,
	configserviceImage string,
	startFunctions bool,
	runServices ...*RunService,
) (map[string]*Service, error) {
	minio, err := minio(subdomain, dataFolder)
	if err != nil {
		return nil, err
	}

	storage, err := storage(cfg, subdomain, useTLS, httpPort, ports.Storage)
	if err != nil {
		return nil, err
	}

	pgVolumeName := "pgdata_" + sanitizeBranch(branch)
	postgres, err := postgres(cfg, subdomain, postgresPort, dataFolder, pgVolumeName)
	if err != nil {
		return nil, err
	}

	graphql, err := graphql(cfg, subdomain, useTLS, httpPort, ports.Graphql)
	if err != nil {
		return nil, err
	}
	jwtSecret := graphql.Environment["HASURA_GRAPHQL_JWT_SECRET"]

	console, err := console(cfg, subdomain, httpPort, useTLS, nhostFolder, ports.Console)
	if err != nil {
		return nil, err
	}

	traefik, err := traefik(subdomain, projectName, httpPort, dotNhostFolder)
	if err != nil {
		return nil, err
	}

	mailhog, err := mailhog(subdomain, dataFolder, useTLS)
	if err != nil {
		return nil, err
	}

	services := map[string]*Service{
		"console":   console,
		"dashboard": dashboard(cfg, subdomain, dashboardVersion, httpPort, useTLS),
		"graphql":   graphql,
		"minio":     minio,
		"postgres":  postgres,
		"storage":   storage,
		"mailhog":   mailhog,
		"traefik":   traefik,
		"configserver": configserver(
			configserviceImage,
			rootFolder,
			nhostFolder,
			useTLS,
			runServices...),
	}

	if startFunctions {
		services["functions"] = functions(
			cfg,
			subdomain,
			httpPort,
			useTLS,
			rootFolder,
			jwtSecret,
			ports.Functions,
			branch,
		)
	}

	if len(cfg.GetHasura().GetJwtSecrets()) > 0 &&
		IsJWTSecretCompatibleWithHasuraAuth(cfg.GetHasura().GetJwtSecrets()[0]) &&
		cfg.GetHasura().GetAuthHook() == nil {
		auth, err := auth(cfg, subdomain, httpPort, useTLS, nhostFolder, ports.Auth)
		if err != nil {
			return nil, err
		}
		services["auth"] = auth

		if cfg.Ai != nil {
			services["ai"] = ai(cfg, subdomain)
		}
	}

	for _, runService := range runServices {
		services["run-"+runService.Config.Name] = run(runService.Config, subdomain, branch)
	}

	return services, nil
}

type RunService struct {
	Config *model.ConfigRunServiceConfig
	Path   string
}

func mountCACertificates(
	path string,
	services map[string]*Service,
) {
	for _, service := range services {
		service.Volumes = append(service.Volumes, Volume{
			Type:     "bind",
			Source:   path,
			Target:   "/etc/ssl/certs/ca-certificates.crt",
			ReadOnly: ptr(true),
		})
	}
}

func ComposeFileFromConfig( //nolint:funlen
	cfg *model.ConfigConfig,
	subdomain string,
	projectName string,
	httpPort uint,
	useTLS bool,
	postgresPort uint,
	dataFolder string,
	nhostFolder string,
	dotNhostFolder string,
	rootFolder string,
	ports ExposePorts,
	branch string,
	dashboardVersion string,
	configserverImage string,
	startFunctions bool,
	caCertificatesPath string,
	runServices ...*RunService,
) (*ComposeFile, error) {
	services, err := getServices(
		cfg,
		subdomain,
		projectName,
		httpPort,
		useTLS,
		postgresPort,
		dataFolder,
		nhostFolder,
		dotNhostFolder,
		rootFolder,
		ports,
		branch,
		dashboardVersion,
		configserverImage,
		startFunctions,
		runServices...,
	)
	if err != nil {
		return nil, err
	}

	pgVolumeName := "pgdata_" + sanitizeBranch(branch)
	volumes := map[string]struct{}{
		rootNodeModules(branch): {},
		pgVolumeName:            {},
	}

	if startFunctions {
		volumes[functionsNodeModules(branch)] = struct{}{}
	}

	for _, runService := range runServices {
		for _, s := range runService.Config.GetResources().GetStorage() {
			volumes[runVolumeName(runService.Config.Name, s.GetName(), branch)] = struct{}{}
		}
	}

	if caCertificatesPath != "" {
		mountCACertificates(caCertificatesPath, services)
	}

	return &ComposeFile{
		Services: services,
		Volumes:  volumes,
	}, nil
}
