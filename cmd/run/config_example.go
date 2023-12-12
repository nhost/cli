package run

import (
	"fmt"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/be/services/mimir/schema"
	"github.com/nhost/cli/clienv"
	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v2"
)

func ptr[T any](v T) *T {
	return &v
}

func CommandConfigExample() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:    "config-example",
		Aliases: []string{},
		Usage:   "Shows an example config file",
		Action:  commandConfigExample,
		Flags:   []cli.Flag{},
	}
}

func commandConfigExample(cCtx *cli.Context) error { //nolint:funlen
	ce := clienv.FromCLI(cCtx)

	//nolint:gomnd
	cfg := &model.ConfigRunServiceConfig{
		Name: "my-run-service",
		Image: &model.ConfigRunServiceImage{
			Image: "docker.io/org/img:latest",
		},
		Command: []string{
			"start",
		},
		Environment: []*model.ConfigEnvironmentVariable{
			{
				Name:  "ENV_VAR1",
				Value: "value1",
			},
			{
				Name:  "ENV_VAR2",
				Value: "value2",
			},
		},
		Ports: []*model.ConfigRunServicePort{
			{
				Port:    8080,
				Type:    "http",
				Publish: ptr(true),
				Ingresses: []*model.ConfigIngress{
					{
						Fqdn: []string{"my-run-service.acme.com"},
					},
				},
			},
		},
		Resources: &model.ConfigRunServiceResources{
			Compute: &model.ConfigRunServiceResourcesCompute{
				Cpu:    125,
				Memory: 256,
			},
			Storage: []*model.ConfigRunServiceResourcesStorage{
				{
					Name:     "my-storage",
					Capacity: 1,
					Path:     "/var/lib/my-storage",
				},
			},
			Replicas: 1,
		},
		HealthCheck: &model.ConfigHealthCheck{
			Port:                8080,
			InitialDelaySeconds: ptr(10),
			ProbePeriodSeconds:  ptr(20),
		},
	}

	sch, err := schema.New()
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	cfg, err = sch.FillRunServiceConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	b, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	ce.Println(string(b))

	return nil
}
