package configserver_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/nhost/be/services/mimir/graph"
	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/cmd/configserver"
)

const rawConfig = `[hasura]
adminSecret = 'hasuraAdminSecret'
webhookSecret = 'webhookSecret'

[[hasura.jwtSecrets]]
type = 'HS256'
key = 'asdasdasdasd'

[observability]
[observability.grafana]
adminPassword = 'asdasd'
`

const rawSecrets = `someSecret = 'asdasd'
`

func ptr[T any](v T) *T {
	return &v
}

func newApp() *graph.App {
	return &graph.App{
		Config: &model.ConfigConfig{
			Global: nil,
			Hasura: &model.ConfigHasura{ //nolint:exhaustruct
				AdminSecret:   "hasuraAdminSecret",
				WebhookSecret: "webhookSecret",
				JwtSecrets: []*model.ConfigJWTSecret{
					{
						Type: ptr("HS256"),
						Key:  ptr("asdasdasdasd"),
					},
				},
			},
			Functions: nil,
			Auth:      nil,
			Postgres:  nil,
			Provider:  nil,
			Storage:   nil,
			Ai:        nil,
			Observability: &model.ConfigObservability{
				Grafana: &model.ConfigGrafana{
					AdminPassword: "asdasd",
				},
			},
		},
		SystemConfig: nil,
		Secrets: []*model.ConfigEnvironmentVariable{
			{
				Name:  "someSecret",
				Value: "asdasd",
			},
		},
		Services: nil,
		AppID:    "00000000-0000-0000-0000-000000000000",
	}
}

func TestLocalGetApps(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		configf  io.ReadWriter
		secretsf io.ReadWriter
		expected []*graph.App
	}{
		{
			name:     "works",
			configf:  bytes.NewBufferString(rawConfig),
			secretsf: bytes.NewBufferString(rawSecrets),
			expected: []*graph.App{newApp()},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			st := configserver.NewLocal(tc.configf, tc.secretsf)
			got, err := st.GetApps(tc.configf, tc.secretsf)
			if err != nil {
				t.Errorf("GetApps() got error: %v", err)
			}

			cmpOpts := cmpopts.IgnoreUnexported(graph.App{}) //nolint:exhaustruct

			if diff := cmp.Diff(tc.expected, got, cmpOpts); diff != "" {
				t.Errorf("GetApps() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLocalUpdateConfig(t *testing.T) { //nolint:dupl
	t.Parallel()

	cases := []struct {
		name     string
		configf  io.ReadWriter
		secretsf io.ReadWriter
		newApp   *graph.App
		expected string
	}{
		{
			name:     "works",
			configf:  bytes.NewBufferString(rawConfig),
			secretsf: bytes.NewBufferString(rawSecrets),
			newApp:   newApp(),
			expected: rawConfig,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			st := configserver.NewLocal(tc.configf, tc.secretsf)

			if err := st.UpdateConfig(
				context.Background(),
				nil,
				tc.newApp,
				nil,
			); err != nil {
				t.Errorf("UpdateConfig() got error: %v", err)
			}

			b, err := io.ReadAll(tc.configf)
			if err != nil {
				t.Errorf("failed to read config file: %v", err)
			}

			if diff := cmp.Diff(tc.expected, string(b)); diff != "" {
				t.Errorf("UpdateConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLocalUpdateSecrets(t *testing.T) { //nolint:dupl
	t.Parallel()

	cases := []struct {
		name     string
		configf  io.ReadWriter
		secretsf io.ReadWriter
		newApp   *graph.App
		expected string
	}{
		{
			name:     "works",
			configf:  bytes.NewBufferString(rawConfig),
			secretsf: bytes.NewBufferString(rawSecrets),
			newApp:   newApp(),
			expected: rawSecrets,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			st := configserver.NewLocal(tc.configf, tc.secretsf)

			if err := st.UpdateSecrets(
				context.Background(),
				nil,
				tc.newApp,
				nil,
			); err != nil {
				t.Errorf("UpdateSecrets() got error: %v", err)
			}

			b, err := io.ReadAll(tc.secretsf)
			if err != nil {
				t.Errorf("failed to read config file: %v", err)
			}

			if diff := cmp.Diff(tc.expected, string(b)); diff != "" {
				t.Errorf("UpdateSecrets() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
