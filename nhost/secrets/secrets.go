package secrets

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/nhost/be/services/mimir/model"
	"strings"
)

func Interpolate(envs []*model.ConfigEnvironmentVariable, secrets []byte) ([]*model.ConfigEnvironmentVariable, error) {
	out := make([]*model.ConfigEnvironmentVariable, len(envs))
	secretVars, err := parseSecrets(secrets)
	if err != nil {
		return nil, err
	}

	for i, env := range envs {
		out[i] = &model.ConfigEnvironmentVariable{
			Name:  env.Name,
			Value: env.Value,
		}
		for _, secret := range secretVars {
			out[i].Value = strings.ReplaceAll(out[i].Value, fmt.Sprintf("{{ secrets.%s }}", secret.name), secret.value)
		}
	}

	return out, nil
}

type variable struct {
	name  string
	value string
}

// Parse the secrets file with KEY=VALUE and return a list of variables.
func parseSecrets(secrets []byte) ([]variable, error) {
	var vars []variable

	secrets = bytes.TrimSpace(secrets)
	if len(secrets) == 0 {
		return vars, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(secrets))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid secret: %s", line)
		}

		vars = append(vars, variable{name: strings.TrimSpace(parts[0]), value: strings.TrimSpace(parts[1])})
	}

	return vars, nil
}
