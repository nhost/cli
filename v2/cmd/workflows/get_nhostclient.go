package workflows

import (
	"context"
	"errors"
	"fmt"

	"github.com/nhost/cli/v2/nhostclient"
	"github.com/nhost/cli/v2/system"
)

func GetNhostClient(ctx context.Context, domain string) (*nhostclient.Client, *nhostclient.LoginResponse, error) {
	authFile, err := system.GetStateAuthFile()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get auth file: %w", err)
	}
	defer authFile.Close()

	creds, err := system.GetCredentials(authFile)
	switch {
	case errors.Is(err, system.ErrAuthFileEmpty):
		return nil,
			nil,
			fmt.Errorf("no credentials found in local storage, please log in with `nhost login`") //nolint:goerr113
	case err != nil:
		return nil, nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	cl := nhostclient.New(domain)
	session, err := cl.LoginPAT(ctx, creds.PersonalAccessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to login with PAT: %w", err)
	}

	return cl, &session, nil
}
