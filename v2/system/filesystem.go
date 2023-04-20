package system

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetStateHome() string {
	var path string
	if os.Getenv("XDG_STATE_HOME") != "" {
		path = filepath.Join(os.Getenv("XDG_STATE_HOME"), "nhost")
	} else {
		path = filepath.Join(os.Getenv("HOME"), ".nhost", "state")
	}

	if err := os.MkdirAll(path, 0o755); err != nil { //nolint:gomnd
		return ""
	}

	return path
}

func GetStateAuthFile() (*os.File, error) {
	f, err := os.OpenFile(filepath.Join(GetStateHome(), "auth.json"), os.O_RDWR|os.O_CREATE, 0o600) //nolint:gomnd
	if err != nil {
		return nil, fmt.Errorf("failed to open auth file: %w", err)
	}
	return f, nil
}
