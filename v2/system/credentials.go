package system

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/nhost/cli/v2/nhostclient"
)

var ErrAuthFileEmpty = fmt.Errorf("auth file is empty")

func GetCredentials(r io.Reader) (nhostclient.CreatePATResponse, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nhostclient.CreatePATResponse{}, fmt.Errorf("failed to read auth file: %w", err)
	}

	if len(b) == 0 {
		return nhostclient.CreatePATResponse{}, ErrAuthFileEmpty
	}

	credentials := nhostclient.CreatePATResponse{} //nolint:exhaustruct
	if err := json.Unmarshal(b, &credentials); err != nil {
		return nhostclient.CreatePATResponse{}, fmt.Errorf("failed to unmarshal auth file: %w", err)
	}

	return credentials, nil
}
