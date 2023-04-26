package project

import (
	"fmt"

	"github.com/nhost/cli/v2/nhostclient/credentials"
	"github.com/nhost/cli/v2/system"
)

func AuthFromDisk() (credentials.Credentials, error) {
	f, err := system.GetStateAuthFile()
	if err != nil {
		return credentials.Credentials{}, fmt.Errorf("failed to get auth file: %w", err)
	}
	defer f.Close()

	creds, err := credentials.FromReader(f)
	if err != nil {
		return credentials.Credentials{}, fmt.Errorf("failed to get credentials: %w", err)
	}
	return creds, nil
}

func AuthToDisk(creds credentials.Credentials) error {
	f, err := system.GetStateAuthFile()
	if err != nil {
		return fmt.Errorf("failed to get auth file: %w", err)
	}
	defer f.Close()

	if err := system.MarshalJSON(creds, f); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}
	return nil
}
