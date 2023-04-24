package project

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/nhost/be/services/mimir/model"
)

type InvalidSecretError struct {
	line int
}

func (e *InvalidSecretError) Error() string {
	return fmt.Sprintf("invalid secret on line %d", e.line)
}

func UnmarshalSecrets(r io.Reader) (model.Secrets, error) {
	secrets := model.Secrets{}

	scanner := bufio.NewScanner(r)
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

func MarshalSecrets(secrets model.Secrets, w io.Writer) error {
	for _, v := range secrets {
		if _, err := fmt.Fprintf(w, "%s=%s", v.Name, v.Value); err != nil {
			return fmt.Errorf("failed to write env: %w", err)
		}
	}
	return nil
}
