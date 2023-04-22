package system

import (
	"fmt"
	"io"
	"strings"
)

func AddToGitignore(g io.ReadWriter, l string) error {
	b, err := io.ReadAll(g)
	if err != nil {
		return fmt.Errorf("failed to read gitignore: %w", err)
	}

	if strings.Contains(string(b), l) {
		return nil
	}

	if _, err := g.Write([]byte(l)); err != nil {
		return fmt.Errorf("failed to write gitignore: %w", err)
	}

	return nil
}
