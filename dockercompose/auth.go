package dockercompose

import (
	"fmt"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/be/services/mimir/schema/appconfig"
)

func auth(cfg *model.ConfigConfig, useTLS bool, nhostFolder string) (*Service, error) { //nolint:funlen
	envars, err := appconfig.HasuraAuthEnv(
		cfg,
		"http://graphql:8080/v1/graphql",
		"http://auth:4000",
		"postgres://nhost_auth_admin@postgres:5432/postgres",
		&model.ConfigSmtp{
			User:     "apiKey",
			Password: "",
			Sender:   "",
			Host:     "",
			Port:     465, //nolint:gomnd
			Secure:   true,
			Method:   "LOGIN",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get hasura env vars: %w", err)
	}

	env := make(map[string]string, len(envars))
	for _, v := range envars {
		env[v.Name] = v.Value
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
		Environment: env,
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
	}, nil
}
