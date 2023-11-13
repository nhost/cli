package dockercompose //nolint:testpackage

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nhost/be/services/mimir/model"
)

func expectedAI() *Service {
	return &Service{ //nolint:exhaustruct
		Image: "nhost/graphite:0.2.5",
		DependsOn: map[string]DependsOn{
			"graphql":  {Condition: "service_healthy"},
			"postgres": {Condition: "service_healthy"},
		},
		Command: []string{"serve"},
		Environment: map[string]string{
			"ENV1":                        "VALUE1",
			"ENV2":                        "VALUE2",
			"GRAPHITE_BASE_URL":           "http://ai:8090",
			"GRAPHITE_WEBHOOK_SECRET":     "webhookSecret",
			"HASURA_GRAPHQL_ADMIN_SECRET": "adminSecret",
			"NHOST_GRAPHQL_URL":           "http://graphql:8080/v1/graphql",
			"OPENAI_API_KEY":              "openaiApiKey",
			"OPENAI_ORG":                  "my-org",
			"POSTGRES_CONNECTION":         "postgres://postgres@postgres:5432/local?sslmode=disable",
			"SYNCH_PERIOD":                "10m",
		},
		ExtraHosts: []string{
			"host.docker.internal:host-gateway",
			"local.auth.nhost.run:host-gateway",
			"local.db.nhost.run:host-gateway",
			"local.functions.nhost.run:host-gateway",
			"local.graphql.nhost.run:host-gateway",
			"local.hasura.nhost.run:host-gateway",
			"local.storage.nhost.run:host-gateway",
		},
		Restart: "always",
	}
}

func TestAI(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		cfg      func() *model.ConfigConfig
		useTlS   bool
		expected func() *Service
	}{
		{
			name:     "success",
			cfg:      getConfig,
			expected: expectedAI,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			got := ai(tc.cfg())
			if diff := cmp.Diff(tc.expected(), got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
