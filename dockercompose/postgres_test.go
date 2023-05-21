package dockercompose //nolint:testpackage

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nhost/be/services/mimir/model"
)

func expectedPostgres() *Service {
	return &Service{
		Image: "nhost/postgres:14.5-20220831-1",
		Command: []string{
			"postgres", "-c", "config_file=/etc/postgresql.conf", "-c",
			"hba_file=/etc/pg_hba_local.conf",
		},
		DependsOn:  nil,
		EntryPoint: nil,
		Environment: map[string]string{
			"PGDATA":            "/var/lib/postgresql/data/pgdata",
			"POSTGRES_DB":       "local",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_USER":     "postgres",
		},
		ExtraHosts: []string{
			"host.docker.internal:host-gateway", "local.auth.nhost.run:host-gateway",
			"local.db.nhost.run:host-gateway", "local.functions.nhost.run:host-gateway",
			"local.graphql.nhost.run:host-gateway", "local.hasura.nhost.run:host-gateway",
			"local.storage.nhost.run:host-gateway",
		},
		HealthCheck: &HealthCheck{
			Test:        []string{"CMD-SHELL", "pg_isready -U postgres", "-d", "postgres", "-q"},
			Interval:    "5s",
			StartPeriod: "60s",
		},
		Labels:  nil,
		Ports:   []Port{{Mode: "ingress", Target: 5432, Published: "5432", Protocol: "tcp"}},
		Restart: "always",
		Volumes: []Volume{
			{
				Type:   "bind",
				Source: "/tmp/data/db/pgdata",
				Target: "/var/lib/postgresql/data/pgdata",
			},
			{
				Type:   "bind",
				Source: "/tmp/data/db/pg_hba_local.conf",
				Target: "/etc/pg_hba_local.conf",
			},
		},
		WorkingDir: nil,
	}
}

func TestPostgres(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		cfg      func() *model.ConfigConfig
		expected func() *Service
	}{
		{
			name:     "success",
			cfg:      getConfig,
			expected: expectedPostgres,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc

			got, err := postgres(tc.cfg(), 5432, "/tmp/data")
			if err != nil {
				t.Errorf("got error: %v", err)
			}

			if diff := cmp.Diff(tc.expected(), got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
