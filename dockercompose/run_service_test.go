package dockercompose_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/dockercompose"
)

func ptr[T any](v T) *T {
	return &v
}

func TestRunService(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		cfg             *model.ConfigRunServiceConfig
		expectedService *dockercompose.Service
		dataFolder      string
	}{
		{
			name: "success",
			cfg: &model.ConfigRunServiceConfig{
				Name: "httpbin",
				Image: &model.ConfigRunServiceImage{
					Image: "docker.io/kong/httpbin:0.1.0",
				},
				Command: []string{
					"gunicorn", "-b", "0.0.0.0:8080", "httpbin:app", "-k", "gevent",
				},
				Environment: []*model.ConfigEnvironmentVariable{
					{
						Name:  "FOO",
						Value: "bar",
					},
				},
				Ports: []*model.ConfigRunServicePort{
					{
						Port:    8080,
						Type:    "http",
						Publish: ptr(true),
					},
					{
						Port:    8088,
						Type:    "http",
						Publish: ptr(true),
					},
					{
						Port:    8081,
						Type:    "tcp",
						Publish: ptr(false),
					},
				},
				Resources: &model.ConfigRunServiceResources{
					Compute: &model.ConfigRunServiceResourcesCompute{
						Cpu:    1000,
						Memory: 1024,
					},
					Storage: []*model.ConfigRunServiceResourcesStorage{
						{
							Name:     "storage",
							Path:     "/storage",
							Capacity: 1,
						},
					},
					Replicas: 1,
				},
			},
			expectedService: &dockercompose.Service{
				Image:      "docker.io/kong/httpbin:0.1.0",
				DependsOn:  map[string]dockercompose.DependsOn{},
				EntryPoint: []string{},
				Command: []string{
					"gunicorn", "-b", "0.0.0.0:8080", "httpbin:app", "-k", "gevent",
				},
				Environment: map[string]string{
					"FOO":                 "bar",
					"NHOST_SUBDOMAIN":     "local",
					"NHOST_REGION":        "",
					"NHOST_AUTH_URL":      "http://auth:4000",
					"NHOST_FUNCTIONS_URL": "http://functions:3000",
					"NHOST_GRAPHQL_URL":   "http://graphql:8080/v1/graphql",
					"NHOST_HASURA_URL":    "http://hasura:8080",
					"NHOST_STORAGE_URL":   "http://storage:5000/v1",
					"NHOST_POSTGRES_HOST": "postgres:5432",
				},
				ExtraHosts:  []string{},
				HealthCheck: nil,
				Labels: map[string]string{
					"traefik.enable": "true",
					"traefik.http.routers.httpbin-8080.entrypoints":               "web",
					"traefik.http.routers.httpbin-8080.rule":                      "Host(`local-httpbin-8080.svc.nhost.run`)",
					"traefik.http.routers.httpbin-8080.service":                   "httpbin-8080",
					"traefik.http.routers.httpbin-8080.tls":                       "true",
					"traefik.http.routers.httpbin-8088.entrypoints":               "web",
					"traefik.http.routers.httpbin-8088.rule":                      "Host(`local-httpbin-8088.svc.nhost.run`)",
					"traefik.http.routers.httpbin-8088.service":                   "httpbin-8088",
					"traefik.http.routers.httpbin-8088.tls":                       "true",
					"traefik.http.services.httpbin-8080.loadbalancer.server.port": "8080",
					"traefik.http.services.httpbin-8088.loadbalancer.server.port": "8088",
				},
				Ports:   []dockercompose.Port{},
				Restart: "always",
				Volumes: []dockercompose.Volume{
					{
						Type:     "bind",
						Source:   ".nhost/data/branch/httpbin/storage",
						Target:   "/storage",
						ReadOnly: ptr(false),
					},
				},
				WorkingDir: nil,
			},
			dataFolder: ".nhost/data/branch",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			service, err := dockercompose.RunService(
				tc.cfg,
				tc.dataFolder,
				true,
				false,
			)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.expectedService, service); diff != "" {
				t.Errorf("unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}
