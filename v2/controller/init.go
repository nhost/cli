package controller

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-getter"
	"github.com/nhost/cli/v2/project"
	"github.com/nhost/cli/v2/system"
)

const hasuraVersion = 3

func (c *Controller) Init(
	ctx context.Context,
	configf io.Writer,
	secretsf io.Writer,
	gitignoref io.ReadWriter,
	hasuraConfigf io.Writer,
) error {
	config, err := project.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to create default config: %w", err)
	}
	if err := project.MarshalConfig(config, configf); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	secrets := project.DefaultSecrets()
	if err := project.MarshalSecrets(secrets, secretsf); err != nil {
		return fmt.Errorf("failed to save secrets: %w", err)
	}

	hasuraConf := map[string]any{"version": hasuraVersion}
	if err := system.MarshalYAML(hasuraConf, hasuraConfigf); err != nil {
		return fmt.Errorf("failed to save hasura config: %w", err)
	}

	folders := []string{
		system.PathDotNhost(),
		filepath.Join(system.PathNhost(), "migrations"),
		filepath.Join(system.PathNhost(), "metadata"),
		filepath.Join(system.PathNhost(), "seeds"),
		filepath.Join(system.PathNhost(), "emails"),
	}

	for _, f := range folders {
		if err := os.MkdirAll(f, 0o755); err != nil { //nolint:gomnd
			return fmt.Errorf("failed to create folder %s: %w", f, err)
		}
	}

	// add .nhost, .secrets node_modules
	if err := system.AddToGitignore(gitignoref, system.PathSecretsFile()); err != nil {
		return fmt.Errorf("failed to add secrets to .gitignore: %w", err)
	}

	getclient := &getter.Client{ //nolint:exhaustruct
		Ctx:  ctx,
		Src:  "github.com/nhost/hasura-auth/email-templates",
		Dst:  "nhost/emails",
		Mode: getter.ClientModeAny,
		Detectors: []getter.Detector{
			&getter.GitHubDetector{},
		},
	}

	if err := getclient.Get(); err != nil {
		return fmt.Errorf("failed to download email templates: %w", err)
	}

	return nil
}
