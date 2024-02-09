package dockercompose

import "fmt"

func configserver(image, rootPath, nhostPath string, useTLS bool) *Service {
	return &Service{
		Image:      image,
		DependsOn:  map[string]DependsOn{},
		EntryPoint: []string{},
		Command: []string{
			"configserver",
		},
		Environment: map[string]string{},
		ExtraHosts:  []string{},
		HealthCheck: nil,
		Labels: Ingresses{
			{
				Name:    "configserver",
				TLS:     useTLS,
				Rule:    "Host(`local.dashboard.nhost.run`) && PathPrefix(`/v1/configserver`)",
				Port:    configserverPort,
				Rewrite: nil,
			},
		}.Labels(),
		Ports:   []Port{},
		Restart: "always",
		Volumes: []Volume{
			{
				Type:     "bind",
				Source:   fmt.Sprintf("%s/nhost.toml", nhostPath),
				Target:   "/tmp/config.toml",
				ReadOnly: ptr(false),
			},
			{
				Type:     "bind",
				Source:   fmt.Sprintf("%s/.secrets", rootPath),
				Target:   "/tmp/secrets.toml",
				ReadOnly: ptr(false),
			},
		},
		WorkingDir: nil,
	}
}
