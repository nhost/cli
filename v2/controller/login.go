package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/nhost/cli/v2/tui"
)

func (c *Controller) Login(
	ctx context.Context,
	authWr io.Writer,
	email string,
	password string,
) error {
	c.p.Println(tui.Info("Authenticating"))
	loginResp, err := c.cl.Login(ctx, email, password)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	c.p.Println(tui.Info("Successfully logged in, creating PAT"))
	patRes, err := c.cl.CreatePAT(ctx, loginResp.Session.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to create PAT: %w", err)
	}
	c.p.Println(tui.Info("Successfully created PAT"))
	c.p.Println(tui.Info("Storing PAT for future user"))

	b, err := json.Marshal(patRes)
	if err != nil {
		return fmt.Errorf("failed to marshal PAT: %w", err)
	}

	if _, err := authWr.Write(b); err != nil {
		return fmt.Errorf("failed to write PAT to file: %w", err)
	}

	return nil
}
