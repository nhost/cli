package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/nhost/cli/v2/nhostclient"
	"github.com/nhost/cli/v2/tui"
)

func Login(
	ctx context.Context,
	printer func(i ...any),
	authWr io.Writer,
	email string,
	password string,
	cl *nhostclient.Client,
) error {
	printer(tui.Info("Authenticating\n"))
	loginResp, err := cl.Login(ctx, email, password)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	printer(tui.Info("Successfully logged in, creating PAT\n"))
	patRes, err := cl.CreatePAT(ctx, loginResp.Session.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to create PAT: %w", err)
	}
	printer(tui.Info("Successfully created PAT\n"))
	printer(tui.Info("Storing PAT for future user\n"))

	b, err := json.Marshal(patRes)
	if err != nil {
		return fmt.Errorf("failed to marshal PAT: %w", err)
	}

	if _, err := authWr.Write(b); err != nil {
		return fmt.Errorf("failed to write PAT to file: %w", err)
	}

	return nil
}
