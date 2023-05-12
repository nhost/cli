package dockercompose

import "fmt"

type Ingresses []Ingress

func (i Ingresses) Labels() map[string]string {
	labels := make(map[string]string)
	labels["traefik.enable"] = "true"
	for _, ingress := range i {
		for k, v := range ingress.Labels() {
			labels[k] = v
		}
	}
	return labels
}

type Ingress struct {
	Name    string
	TLS     bool
	Rule    string
	Port    uint
	Rewrite *Rewrite
}

type Rewrite struct {
	Regex       string
	Replacement string
}

func (i Ingress) Labels() map[string]string {
	labels := map[string]string{
		fmt.Sprintf("traefik.http.routers.%s.entrypoints", i.Name): "web",
		fmt.Sprintf("traefik.http.routers.%s.rule", i.Name):        i.Rule,
		fmt.Sprintf("traefik.http.routers.%s.service", i.Name):     i.Name,
		fmt.Sprintf("traefik.http.routers.%s.tls", i.Name): fmt.Sprintf(
			"%t",
			i.TLS,
		),
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", i.Name): fmt.Sprintf(
			"%d",
			i.Port,
		),
	}

	if i.Rewrite != nil {
		labels[fmt.Sprintf("traefik.http.middlewares.replace-%s.replacepathregex.regex", i.Name)] = i.Rewrite.Regex
		//nolint:lll
		labels[fmt.Sprintf("traefik.http.middlewares.replace-%s.replacepathregex.replacement", i.Name)] = i.Rewrite.Replacement
		labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", i.Name)] = fmt.Sprintf(
			"replace-%s",
			i.Name,
		)
	}

	return labels
}
