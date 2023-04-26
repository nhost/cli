package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/nhost/cli/v2/project"
	"github.com/nhost/cli/v2/system"
	"github.com/nhost/cli/v2/tui"
)

func Logout(
	ctx context.Context,
	p Printer,
	cl NhostClient,
) error {
	p.Println(tui.Info("Retrieving credentials from local storage"))

	creds, err := project.AuthFromDisk()
	switch {
	case errors.Is(err, system.ErrNoContent):
		p.Println(tui.Info("No credentials found in local storage"))
		return err
	case err != nil:
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	p.Println(tui.Info("Getting an access token"))
	loginResp, err := cl.LoginPAT(
		ctx,
		creds.PersonalAccessToken,
	)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	p.Println(tui.Info("Invalidating PAT"))
	if err := cl.Logout(
		ctx,
		creds.PersonalAccessToken,
		loginResp.Session.AccessToken,
	); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}
