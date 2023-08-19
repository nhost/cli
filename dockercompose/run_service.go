package dockercompose

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nhost/be/services/mimir/model"
)

func getIngresses(
	cfg *model.ConfigRunServiceConfig,
	useTLS bool,
) Ingresses {
	ingresses := make(Ingresses, 0, len(cfg.Ports))
	for _, p := range cfg.Ports {
		if *p.GetPublish() {
			name := fmt.Sprintf("%s-%d", cfg.Name, p.Port)
			ingresses = append(ingresses, Ingress{
				Name:    name,
				TLS:     useTLS,
				Rule:    fmt.Sprintf("Host(`local-%s.svc.nhost.run`)", name),
				Port:    uint(p.Port),
				Rewrite: nil,
			})
		}
	}

	return ingresses
}

func getEnv(cfg *model.ConfigRunServiceConfig) map[string]string {
	env := map[string]string{}
	for _, e := range cfg.Environment {
		env[e.Name] = e.Value
	}

	env["NHOST_AUTH_URL"] = "http://auth:4000"
	env["NHOST_FUNCTIONS_URL"] = "http://functions:3000"
	env["NHOST_GRAPHQL_URL"] = "http://graphql:8080/v1/graphql"
	env["NHOST_HASURA_URL"] = "http://hasura:8080"
	env["NHOST_POSTGRES_HOST"] = "postgres:5432"
	env["NHOST_REGION"] = ""
	env["NHOST_SUBDOMAIN"] = "local"
	env["NHOST_STORAGE_URL"] = "http://storage:5000/v1"

	return env
}

func getVolumes(
	cfg *model.ConfigRunServiceConfig, dataFolder string, createFolders bool,
) ([]Volume, error) {
	volumes := make([]Volume, 0, len(cfg.GetResources().GetStorage()))
	for _, v := range cfg.GetResources().GetStorage() {
		path := filepath.Join(dataFolder, cfg.Name, v.Name)
		if createFolders {
			if err := os.MkdirAll(path, 0o755); err != nil { //nolint:gomnd
				return nil, fmt.Errorf("failed to create folder %s: %w", path, err)
			}
		}
		volumes = append(volumes, Volume{
			Type:     "bind",
			Source:   path,
			Target:   v.GetPath(),
			ReadOnly: ptr(false),
		})
	}

	return volumes, nil
}

func RunService(
	cfg *model.ConfigRunServiceConfig,
	dataFolder string,
	useTLS bool,
	createFolders bool,
) (*Service, error) {
	volumes, err := getVolumes(cfg, dataFolder, createFolders)
	if err != nil {
		return nil, err
	}

	return &Service{
		Image:       cfg.Image.Image,
		DependsOn:   map[string]DependsOn{},
		EntryPoint:  []string{},
		Command:     cfg.Command,
		Environment: getEnv(cfg),
		ExtraHosts:  []string{},
		HealthCheck: nil,
		Labels:      getIngresses(cfg, useTLS).Labels(),
		Ports:       []Port{},
		Restart:     "always",
		Volumes:     volumes,
		WorkingDir:  nil,
	}, nil
}

func ComposeFileForRunServiceStandalone(
	cfg *model.ConfigRunServiceConfig,
	projectName string,
	httpPort uint,
	useTLS bool,
	dataFolder string,
	dotNhostFolder string,
) (*ComposeFile, error) {
	traefik, err := traefik(projectName, httpPort, dotNhostFolder)
	if err != nil {
		return nil, err
	}

	service, err := RunService(cfg, dataFolder, useTLS, true)
	if err != nil {
		return nil, err
	}

	c := &ComposeFile{
		Version: "3.8",
		Services: map[string]*Service{
			"traefik": traefik,
			cfg.Name:  service,
		},
		Volumes: map[string]struct{}{},
	}
	return c, nil
}
