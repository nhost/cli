package compose

import (
	"fmt"
	"github.com/compose-spec/compose-go/types"
)

type env map[string]string

func (e env) merge(otherEnv ...env) env {
	for _, e2 := range otherEnv {
		for k, v := range e2 {
			e[k] = v
		}
	}

	return e
}

func (e env) dockerServiceConfigEnv() types.MappingWithEquals {
	out := []string{}

	for k, v := range e {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}

	return types.NewMappingWithEquals(out)
}
