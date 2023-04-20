package workflows

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nhost/cli/v2/nhostclient"
	"github.com/nhost/cli/v2/system"
	"github.com/nhost/cli/v2/tui"
)

func Logout(
	ctx context.Context,
	printer func(i ...any),
	authr io.Reader,
	cl *nhostclient.Client,
) error {
	printer(tui.Info("Retrieving credentials from local storage\n"))
	credentials, err := system.GetCredentials(authr)
	switch {
	case errors.Is(err, system.ErrAuthFileEmpty):
		printer(tui.Info("No credentials found in local storage\n"))
		return err //nolint:wrapcheck
	case err != nil:
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	printer(tui.Info("Getting an access token\n"))
	loginResp, err := cl.LoginPAT(
		ctx,
		credentials.PersonalAccessToken,
	)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	printer(tui.Info("Invalidating PAT\n"))
	if err := cl.Logout(
		ctx,
		credentials.PersonalAccessToken,
		loginResp.Session.AccessToken,
	); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}
