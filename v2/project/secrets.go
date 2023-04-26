package project

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/v2/system"
)

func DefaultSecrets() model.Secrets {
	return model.Secrets{
		{
			Name:  "HASURA_GRAPHQL_ADMIN_SECRET",
			Value: "nhost-admin-secret",
		},
		{
			Name:  "HASURA_GRAPHQL_JWT_SECRET",
			Value: "0f987876650b4a085e64594fae9219e7781b17506bec02489ad061fba8cb22db",
		},
		{
			Name:  "NHOST_WEBHOOK_SECRET",
			Value: "nhost-webhook-secret",
		},
	}
}

type InvalidSecretError struct {
	line int
}

func (e *InvalidSecretError) Error() string {
	return fmt.Sprintf("invalid secret on line %d", e.line)
}

func SecretsFromDisk() (model.Secrets, error) {
	secrets := model.Secrets{}

	f, err := system.GetSecretsFile()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	i := 1
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Split(line, "#")[0]
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2) //nolint:gomnd
		if len(parts) != 2 {                  //nolint:gomnd
			return nil, &InvalidSecretError{i}
		}

		secrets = append(
			secrets,
			&model.ConfigEnvironmentVariable{
				Name:  strings.TrimSpace(parts[0]),
				Value: strings.TrimSpace(parts[1]),
			},
		)
		i++
	}

	return secrets, nil
}

func SecretsToDisk(secrets model.Secrets) error {
	f, err := system.GetSecretsFile()
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer f.Close()

	for _, v := range secrets {
		if _, err := fmt.Fprintf(f, "%s=%s\n", v.Name, v.Value); err != nil {
			return fmt.Errorf("failed to write env: %w", err)
		}
	}
	return nil
}
