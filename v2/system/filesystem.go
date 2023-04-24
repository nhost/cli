package system

import (
	"fmt"
	"os"
	"path/filepath"
)

func PathConfig() string {
	return filepath.Join("nhost", "nhost.toml")
}

func PathSecretsFile() string {
	return ".secrets"
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func GetNhostFolder() string {
	const path = ".nhost"

	if err := os.MkdirAll(path, 0o755); err != nil { //nolint:gomnd
		return ""
	}

	return path
}

func GetConfigFile() (*os.File, error) {
	f, err := os.OpenFile(PathConfig(), os.O_RDWR|os.O_CREATE, 0o600) //nolint:gomnd
	if err != nil {
		return nil, fmt.Errorf("failed to open project file: %w", err)
	}
	return f, nil
}

func GetSecretsFile() (*os.File, error) {
	f, err := os.OpenFile(PathSecretsFile(), os.O_RDWR|os.O_CREATE, 0o600) //nolint:gomnd
	if err != nil {
		return nil, fmt.Errorf("failed to open project file: %w", err)
	}
	return f, nil
}

func GetNhostProjectInfoFile() (*os.File, error) {
	f, err := os.OpenFile(
		filepath.Join(GetNhostFolder(), "project.json"), os.O_RDWR|os.O_CREATE, 0o600, //nolint:gomnd
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open project file: %w", err)
	}
	return f, nil
}

func GetGitignoreFile() (*os.File, error) {
	f, err := os.OpenFile(".gitignore", os.O_RDWR|os.O_CREATE, 0o600) //nolint:gomnd
	if err != nil {
		return nil, fmt.Errorf("failed to open .gitignore file: %w", err)
	}
	return f, nil
}

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
	f, err := os.OpenFile(
		filepath.Join(GetStateHome(), "auth.json"), os.O_RDWR|os.O_CREATE, 0o600, //nolint:gomnd
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open auth file: %w", err)
	}
	return f, nil
}
