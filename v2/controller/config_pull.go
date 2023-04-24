package controller

import (
	"context"
	"fmt"
	"io"
	"strings"

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

func (c *Controller) ConfigPull(
	ctx context.Context,
	projectf io.Reader,
	tomlf io.Writer,
	secretsf io.Writer,
	gitignoref io.ReadWriter,
) error {
	proj, err := project.UnmarshalProjectInfo(projectf)
	if err != nil {
		return err //nolint:wrapcheck
	}

	session, err := c.GetNhostSession(ctx)
	if err != nil {
		return err
	}

	c.p.Println(tui.Info("Pulling config from Nhost..."))
	cfg, err := c.cl.GetConfigRawJSON(ctx, proj.ID, graphql.WithAccessToken(session.Session.AccessToken))
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	var v any
	if err := system.UnmarshalJSON(strings.NewReader(cfg.ConfigRawJSON), &v); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := system.MarshalTOML(v, tomlf); err != nil {
		return fmt.Errorf("failed to save nhost.toml: %w", err)
	}

	c.p.Println(tui.Info("Getting secrets list from Nhost..."))
	resp, err := c.cl.GetSecrets(ctx, proj.ID, graphql.WithAccessToken(session.Session.AccessToken))
	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}

	if err := project.MarshalSecrets(respToSecrets(resp.GetAppSecrets()), secretsf); err != nil {
		return fmt.Errorf("failed to save nhost.toml: %w", err)
	}

	c.p.Println(tui.Info("Adding .secrets to .gitignore..."))
	if err := system.AddToGitignore(gitignoref, "\n.secrets\n"); err != nil {
		return fmt.Errorf("failed to add .secrets to .gitignore: %w", err)
	}

	c.p.Println(tui.Info("Success!"))
	c.p.Println(tui.Warn("- Review `nhost/nhost.toml` and make sure there are no secrets before you commit it to git."))
	c.p.Println(tui.Warn("- Review `.secrets` file and set your development secrets"))
	c.p.Println(tui.Warn("- Review `.secrets` was added to .gitignore"))

	return nil
}
