package system

import (
	"fmt"
	"os"
	"path/filepath"
)

func PathAuthFile() string {
	return filepath.Join(GetStateHome(), "auth.json")
}

func PathNhost() string {
	return "nhost"
}

func PathHasura() string {
	return filepath.Join(PathNhost(), "config.yaml")
}

func PathConfig() string {
	return filepath.Join(PathNhost(), "nhost.toml")
}

func PathSecrets() string {
	return ".secrets"
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func PathDotNhost() string {
	return ".nhost"
}

func GetHasuraFile() (*os.File, error) {
	f, err := os.OpenFile(PathHasura(), os.O_RDWR|os.O_CREATE, 0o600) //nolint:gomnd
	if err != nil {
		return nil, fmt.Errorf("failed to open hasura config file: %w", err)
	}
	return f, nil
}

func GetConfigFile() (*os.File, error) {
	f, err := os.OpenFile(PathConfig(), os.O_RDWR|os.O_CREATE, 0o600) //nolint:gomnd
	if err != nil {
		return nil, fmt.Errorf("failed to open project file: %w", err)
	}
	return f, nil
}

func GetSecretsFile() (*os.File, error) {
	f, err := os.OpenFile(PathSecrets(), os.O_RDWR|os.O_CREATE, 0o600) //nolint:gomnd
	if err != nil {
		return nil, fmt.Errorf("failed to open secrets file: %w", err)
	}
	return f, nil
}

func GetNhostProjectInfoFile() (*os.File, error) {
	f, err := os.OpenFile(
		filepath.Join(PathDotNhost(), "project.json"), os.O_RDWR|os.O_CREATE, 0o600, //nolint:gomnd
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

	return path
}

func GetStateAuthFile() (*os.File, error) {
	f, err := os.OpenFile(PathAuthFile(), os.O_RDWR|os.O_CREATE, 0o600) //nolint:gomnd
	if err != nil {
		return nil, fmt.Errorf("failed to open auth file: %w", err)
	}
	return f, nil
}
