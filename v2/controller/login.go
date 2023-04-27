package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nhost/cli/v2/controller/workflows"
	"github.com/nhost/cli/v2/system"
	"github.com/nhost/cli/v2/tui"
)

func Login(
	ctx context.Context,
	p Printer,
	cl NhostClient,
	email string,
	password string,
) error {
	p.Println(tui.Info("Authenticating"))
	loginResp, err := cl.Login(ctx, email, password)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	p.Println(tui.Info("Successfully logged in, creating PAT"))
	creds, err := cl.CreatePAT(ctx, loginResp.Session.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to create PAT: %w", err)
	}
	p.Println(tui.Info("Successfully created PAT"))
	p.Println(tui.Info("Storing PAT for future user"))

	if err := workflows.MarshalFile(creds, system.PathAuthFile(), json.Marshal); err != nil {
		return fmt.Errorf("failed to write PAT to file: %w", err)
	}

	return nil
}
