package system

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v2"
)

var ErrNoContent = fmt.Errorf("no content")

func Marshal(v any, w io.Writer, fn func(any) ([]byte, error)) error {
	if f, ok := w.(*os.File); ok {
		if err := f.Truncate(0); err != nil {
			return fmt.Errorf("error truncating file: %w", err)
		}
	}

	b, err := fn(v)
	if err != nil {
		return fmt.Errorf("error marshalling object: %w", err)
	}

	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("error writing marshalled object: %w", err)
	}

	return nil
}

func MarshalTOML(v any, w io.Writer) error {
	return Marshal(v, w, toml.Marshal)
}

func MarshalJSON(v any, w io.Writer) error {
	return Marshal(v, w, json.Marshal)
}

func MarshalYAML(v any, w io.Writer) error {
	return Marshal(v, w, yaml.Marshal)
}

func Unmarshal(r io.Reader, v any, f func([]byte, any) error) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read contents of reader: %w", err)
	}

	if err := f(b, v); err != nil {
		return fmt.Errorf("failed to unmarshal object: %w", err)
	}

	return nil
}

func UnmarshalJSON(r io.Reader, v any) error {
	return Unmarshal(r, v, json.Unmarshal)
}

func UnmarshalTOML(r io.Reader, v any) error {
	return Unmarshal(r, v, toml.Unmarshal)
}
