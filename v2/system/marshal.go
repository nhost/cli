package system

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/pelletier/go-toml/v2"
)

var ErrNoContent = fmt.Errorf("no content")

func UnmarshalJSON(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read auth file: %w", err)
	}

	if len(b) == 0 {
		return ErrNoContent
	}

	if err := json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("failed to unmarshal auth file: %w", err)
	}

	return nil
}

func MarshalTOML(v any, w io.Writer) error {
	b, err := toml.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal toml: %w", err)
	}

	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("failed to write toml: %w", err)
	}

	return nil
}

func MarshalEnv(v map[string]string, w io.Writer) error {
	for k, v := range v {
		if _, err := fmt.Fprintf(w, "%s=%s", k, v); err != nil {
			return fmt.Errorf("failed to write env: %w", err)
		}
	}
	return nil
}
