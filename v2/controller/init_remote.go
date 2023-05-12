package controller

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/creack/pty"
	"github.com/nhost/cli/v2/nhostclient/graphql"
)

func InitRemote(
	ctx context.Context,
	p Printer,
	cl NhostClient,
	nhostFolder string,
	domain string,
) error {
	proj, err := GetAppInfo(ctx, p, cl)
	if err != nil {
		return err
	}

	session, err := LoadSession(ctx, p, cl)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	cfg, err := configPull(ctx, p, cl, proj, session)
	if err != nil {
		return err
	}

	if err := initInit(ctx); err != nil {
		return err
	}

	hasuraAdminSecret, err := cl.GetHasuraAdminSecret(
		ctx, proj.ID, graphql.WithAccessToken(session.Session.AccessToken),
	)
	if err != nil {
		return fmt.Errorf("failed to get hasura admin secret: %w", err)
	}

	hasuraEndpoint := fmt.Sprintf(
		"https://%s.hasura.%s.%s", proj.Subdomain, proj.Region.AwsName, domain,
	)

	f, err := os.OpenFile(
		filepath.Join(nhostFolder, "config.yaml"), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644, //nolint:gomnd
	)
	if err != nil {
		return fmt.Errorf("failed to open config.yaml: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("version: 3\n"); err != nil {
		return fmt.Errorf("failed to write version.yaml: %w", err)
	}

	if err := createPostgresMigration(
		ctx, nhostFolder, *cfg.Hasura.Version, hasuraEndpoint, hasuraAdminSecret.App.Config.Hasura.AdminSecret, "public",
	); err != nil {
		return fmt.Errorf("failed to create postgres migration: %w", err)
	}

	if err := createMetada(
		ctx, nhostFolder, *cfg.Hasura.Version, hasuraEndpoint, hasuraAdminSecret.App.Config.Hasura.AdminSecret,
	); err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}
	return nil
}

func createPostgresMigration(
	ctx context.Context, nhostfolder, hasuraVersion, hasuraEndpoint, adminSecret, schema string,
) error {
	cmd := exec.CommandContext( //nolint:gosec
		ctx,
		"docker", "run",
		"-v", fmt.Sprintf("%s:/app", nhostfolder),
		"-w", "/app",
		"-it", "--rm",
		"--entrypoint", "hasura-cli",
		fmt.Sprintf("hasura/graphql-engine:%s.cli-migrations-v3", hasuraVersion),
		"--endpoint", hasuraEndpoint,
		"--admin-secret", adminSecret,
		"migrate", "create", "init", "--from-server", "--schema", schema,
		"--database-name", "default",
	)

	f, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(os.Stdout, f); err != nil {
		return fmt.Errorf("failed to copy pty output: %w", err)
	}
	return nil
}

func createMetada(
	ctx context.Context, nhostfolder, hasuraVersion, hasuraEndpoint, adminSecret string,
) error {
	cmd := exec.CommandContext( //nolint:gosec
		ctx,
		"docker", "run",
		"-v", fmt.Sprintf("%s:/app", nhostfolder),
		"-w", "/app",
		"-it", "--rm",
		"--entrypoint", "hasura-cli",
		fmt.Sprintf("hasura/graphql-engine:%s.cli-migrations-v3", hasuraVersion),
		"--endpoint", hasuraEndpoint,
		"--admin-secret", adminSecret,
		"metadata", "export",
	)

	f, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(os.Stdout, f); err != nil {
		return fmt.Errorf("failed to copy pty output: %w", err)
	}
	return nil
}
