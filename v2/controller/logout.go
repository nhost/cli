package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/nhost/cli/v2/system"
	"github.com/nhost/cli/v2/tui"
)

func (c *Controller) Logout(
	ctx context.Context,
) error {
	c.p.Println(tui.Info("Retrieving credentials from local storage"))

	creds, err := c.credsFunc()
	switch {
	case errors.Is(err, system.ErrNoContent):
		c.p.Println(tui.Info("No credentials found in local storage"))
		return err
	case err != nil:
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	c.p.Println(tui.Info("Getting an access token"))
	loginResp, err := c.cl.LoginPAT(
		ctx,
		creds.PersonalAccessToken,
	)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	c.p.Println(tui.Info("Invalidating PAT"))
	if err := c.cl.Logout(
		ctx,
		creds.PersonalAccessToken,
		loginResp.Session.AccessToken,
	); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}
