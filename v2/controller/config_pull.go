package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/v2/nhostclient/credentials"
	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/project"
	"github.com/nhost/cli/v2/system"
	"github.com/nhost/cli/v2/tui"
)

const (
	DefaultHasuraGraphqlAdminSecret = "nhost-admin-secret" //nolint:gosec
	DefaultGraphqlJWTSecret         = "0f987876650b4a085e64594fae9219e7781b17506bec02489ad061fba8cb22db"
	DefaultNhostWebhookSecret       = "nhost-webhook-secret" //nolint:gosec
)

func configPull(
	ctx context.Context,
	p Printer,
	cl NhostClient,
	proj *graphql.GetWorkspacesApps_Workspaces_Apps,
	session credentials.Session,
) error {
	p.Println(tui.Info("Pulling config from Nhost..."))
	cfg, err := cl.GetConfigRawJSON(ctx, proj.ID, graphql.WithAccessToken(session.Session.AccessToken))
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	var v model.ConfigConfig
	if err := system.UnmarshalJSON(strings.NewReader(cfg.ConfigRawJSON), &v); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	tomlf, err := system.GetConfigFile()
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer tomlf.Close()

	if err := system.MarshalTOML(v, tomlf); err != nil {
		return fmt.Errorf("failed to save nhost.toml: %w", err)
	}

	p.Println(tui.Info("Getting secrets list from Nhost..."))
	resp, err := cl.GetSecrets(ctx, proj.ID, graphql.WithAccessToken(session.Session.AccessToken))
	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	if err := project.SecretsToDisk(respToSecrets(resp.GetAppSecrets())); err != nil {
		return fmt.Errorf("failed to save nhost.toml: %w", err)
	}

	p.Println(tui.Info("Adding .secrets to .gitignore..."))
	if err := system.AddToGitignore("\n.secrets\n"); err != nil {
		return fmt.Errorf("failed to add .secrets to .gitignore: %w", err)
	}

	p.Println(tui.Info("Success!"))
	p.Println(tui.Warn("- Review `nhost/nhost.toml` and make sure there are no secrets before you commit it to git."))
	p.Println(tui.Warn("- Review `.secrets` file and set your development secrets"))
	p.Println(tui.Warn("- Review `.secrets` was added to .gitignore"))

	return nil
}

func ConfigPull(
	ctx context.Context,
	p Printer,
	cl NhostClient,
) error {
	proj, err := project.InfoFromDisk()
	if err != nil {
		return err //nolint:wrapcheck
	}

	session, err := GetNhostSession(ctx, cl)
	if err != nil {
		return err
	}

	return configPull(ctx, p, cl, proj, session)
}
