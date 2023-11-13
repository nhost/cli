package dockercompose

import (
	"fmt"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/be/services/mimir/schema/appconfig"
)

func ai(
	cfg *model.ConfigConfig,
) *Service {
	envars := appconfig.AIEnv(
		cfg,
		"http://graphql:8080/v1/graphql",
		"postgres://postgres@postgres:5432/local?sslmode=disable",
	)

	env := make(map[string]string, len(envars))
	for _, v := range envars {
		env[v.Name] = v.Value
	}

	return &Service{
		Image: fmt.Sprintf("nhost/graphite:%s", *cfg.GetAi().GetVersion()),
		DependsOn: map[string]DependsOn{
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
		Environment: env,
		ExtraHosts:  extraHosts(),
		Labels:      nil,
		Ports:       nil,
		Restart:     "always",
		HealthCheck: nil,
		Volumes:     nil,
		WorkingDir:  nil,
	}
}
