package dockercompose //nolint:testpackage

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nhost/be/services/mimir/model"
)

func TestRun(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		cfg      func() *model.ConfigRunServiceConfig
		useTlS   bool
		expected func() *Service
	}{
		{
			name: "success",
			cfg: func() *model.ConfigRunServiceConfig {
				return &model.ConfigRunServiceConfig{
					Name: "service-name",
					Image: &model.ConfigRunServiceImage{
						Image: "image:tag",
					},
					Command: []string{"asd", "qwe"},
					Environment: []*model.ConfigEnvironmentVariable{
						{
							Name:  "ENV",
							Value: "value",
						},
						{
							Name:  "ENV2",
							Value: "value2",
						},
					},
					Ports: []*model.ConfigRunServicePort{
						{
							Port:    80,
							Type:    "http",
							Publish: ptr(true),
							Ingresses: []*model.ConfigIngress{
								{
									Fqdn: []string{"svc.domain.com"},
								},
							},
						},
						{
							Port:      3000,
							Type:      "tcp",
							Publish:   ptr(false),
							Ingresses: nil,
						},
						{
							Port:      3000,
							Type:      "udp",
							Publish:   ptr(false),
							Ingresses: nil,
						},
					},
					Resources: &model.ConfigRunServiceResources{
						Compute: &model.ConfigComputeResources{
							Cpu:    250,
							Memory: 512,
						},
						Storage: []*model.ConfigRunServiceResourcesStorage{
							{
								Name:     "asd",
								Capacity: 1,
								Path:     "/path/to/asd",
							},
						},
						Replicas: 1,
					},
					HealthCheck: &model.ConfigHealthCheck{
						Port:                80,
						InitialDelaySeconds: ptr(10),
						ProbePeriodSeconds:  ptr(10),
					},
				}
			},
			useTlS: false,
			expected: func() *Service {
				return &Service{
					Image:       "image:tag",
					DependsOn:   map[string]DependsOn{},
					EntryPoint:  []string{"asd", "qwe"},
					Command:     []string{},
					Environment: map[string]string{"ENV": "value", "ENV2": "value2"},
					ExtraHosts: []string{
						"host.docker.internal:host-gateway",
						"local.auth.nhost.run:host-gateway",
						"local.db.nhost.run:host-gateway",
						"local.functions.nhost.run:host-gateway",
						"local.graphql.nhost.run:host-gateway",
						"local.hasura.nhost.run:host-gateway",
						"local.storage.nhost.run:host-gateway",
					},
					Labels: map[string]string{},
					Ports: []Port{
						{Mode: "ingress", Target: 80, Published: "80", Protocol: "tcp"},
					},
					Restart:     "always",
					WorkingDir:  nil,
					HealthCheck: nil,
					Volumes: []Volume{
						{
							Type:     "volume",
							Source:   "branch-run-service-name-asd",
							Target:   "/path/to/asd",
							ReadOnly: ptr(false),
						},
					},
				}
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			got := run(tc.cfg(), "branch")
			if diff := cmp.Diff(tc.expected(), got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
