package credentials

import (
	"fmt"
	"io"

	"github.com/nhost/cli/v2/system"
)

type Credentials struct {
	PersonalAccessToken string `json:"personalAccessToken"`
}

func FromReader(r io.Reader) (Credentials, error) {
	var creds Credentials
	err := system.UnmarshalJSON(r, &creds)
	if err != nil {
		return Credentials{}, fmt.Errorf("failed to unmarshal auth file: %w", err)
	}
	return creds, nil
}
