package project

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/creack/pty"
	"github.com/hashicorp/go-getter"
	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/cmd/config"
	"github.com/nhost/cli/v2/nhostclient/graphql"
	"github.com/nhost/cli/v2/project"
	"github.com/nhost/cli/v2/project/env"
	"github.com/nhost/cli/v2/system"
	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

const (
	flagRemote = "remote"
)

const hasuraMetadataVersion = 3

func CommandInit() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:    "init",
		Aliases: []string{},
		Usage:   "Initialize a new Nhost project",
		Action:  commandInit,
		Flags: []cli.Flag{
			&cli.BoolFlag{ //nolint:exhaustruct
				Name:    flagRemote,
				Usage:   "Initialize pulling configuration, migrations and metadata from the linked project",
				Value:   false,
				EnvVars: []string{"NHOST_REMOTE"},
			},
		},
	}
}

func commandInit(cCtx *cli.Context) error {
	ce := clienv.New(cCtx)

	if clienv.PathExists(ce.Path.NhostFolder()) {
		return fmt.Errorf("nhost folder already exists") //nolint:goerr113
	}

	if err := os.MkdirAll(ce.Path.NhostFolder(), 0o755); err != nil { //nolint:gomnd
		return fmt.Errorf("failed to create nhost folder: %w", err)
	}

	ce.Infoln("Initializing Nhost project")
	if err := Init(cCtx.Context, ce); err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}
	ce.Infoln("Successfully initialized Nhost project, run `nhost dev` to start development")
	return nil
}

func Init(ctx context.Context, ce *clienv.CliEnv) error {
	config, err := project.DefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to create default config: %w", err)
	}
	if err := clienv.MarshalFile(config, ce.Path.NhostToml(), toml.Marshal); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	secrets := project.DefaultSecrets()
	if err := clienv.MarshalFile(secrets, ce.Path.Secrets(), env.Marshal); err != nil {
		return fmt.Errorf("failed to save secrets: %w", err)
	}

	return initInit(ctx, ce.Path)
}

func initInit(
	ctx context.Context, fs *clienv.PathStructure,
) error {
	hasuraConf := map[string]any{"version": hasuraMetadataVersion}
	if err := clienv.MarshalFile(hasuraConf, fs.HasuraConfig(), yaml.Marshal); err != nil {
		return fmt.Errorf("failed to save hasura config: %w", err)
	}

	if err := initFolders(fs); err != nil {
		return err
	}

	gitingoref, err := os.OpenFile(".gitignore", os.O_RDWR|os.O_CREATE, 0o600) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("failed to open .gitignore file: %w", err)
	}
	defer gitingoref.Close()

	if err := system.AddToGitignore(fs.Secrets()); err != nil {
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

func initFolders(fs *clienv.PathStructure) error {
	folders := []string{
		fs.DotNhostFolder(),
		fs.FunctionsFolder(),
		filepath.Join(fs.NhostFolder(), "migrations"),
		filepath.Join(fs.NhostFolder(), "metadata"),
		filepath.Join(fs.NhostFolder(), "seeds"),
		filepath.Join(fs.NhostFolder(), "emails"),
	}
	for _, f := range folders {
		if err := os.MkdirAll(f, 0o755); err != nil { //nolint:gomnd
			return fmt.Errorf("failed to create folder %s: %w", f, err)
		}
	}

	return nil
}

func InitRemote(
	ctx context.Context,
	ce *clienv.CliEnv,
) error {
	proj, err := ce.GetAppInfo()
	if err != nil {
		return fmt.Errorf("failed to get app info: %w", err)
	}

	session, err := ce.LoadSession(ctx)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	cfg, err := config.Pull(ctx, ce, proj, session)
	if err != nil {
		return fmt.Errorf("failed to pull config: %w", err)
	}

	if err := initInit(ctx, ce.Path); err != nil {
		return err
	}

	cl := ce.GetNhostClient()
	hasuraAdminSecret, err := cl.GetHasuraAdminSecret(
		ctx, proj.ID, graphql.WithAccessToken(session.Session.AccessToken),
	)
	if err != nil {
		return fmt.Errorf("failed to get hasura admin secret: %w", err)
	}

	hasuraEndpoint := fmt.Sprintf(
		"https://%s.hasura.%s.%s", proj.Subdomain, proj.Region.AwsName, ce.Domain(),
	)

	ce.Infoln("Creating postgres migration")
	if err := createPostgresMigration(
		ctx,
		ce.Path.NhostFolder(),
		*cfg.Hasura.Version,
		hasuraEndpoint,
		hasuraAdminSecret.App.Config.Hasura.AdminSecret,
		"public",
	); err != nil {
		return fmt.Errorf("failed to create postgres migration: %w", err)
	}

	ce.Infoln("Downloading metadata")
	if err := createMetada(
		ctx, ce.Path.NhostFolder(),
		*cfg.Hasura.Version,
		hasuraEndpoint,
		hasuraAdminSecret.App.Config.Hasura.AdminSecret,
	); err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}

	ce.Infoln("Project initialized successfully!")
	return nil
}

func createPostgresMigration(
	ctx context.Context, nhostfolder, hasuraVersion, hasuraEndpoint, adminSecret, schema string,
) error {
	cmd := exec.CommandContext( //nolint:gosec
		ctx,
		"docker", "run",
		"-v", fmt.Sprintf("%s:/app", nhostfolder),
		"-e", "HASURA_GRAPHQL_ENABLE_TELEMETRY=false",
		"-w", "/app",
		"-it", "--rm",
		"--entrypoint", "hasura-cli",
		fmt.Sprintf("hasura/graphql-engine:%s.cli-migrations-v3", hasuraVersion),
		"--endpoint", hasuraEndpoint,
		"--admin-secret", adminSecret,
		"migrate", "create", "init", "--from-server", "--schema", schema,
		"--database-name", "default",
		"--skip-update-check",
		"--log-level", "ERROR",
	)

	f, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(os.Stdout, f); err != nil {
		return fmt.Errorf("failed to copy output: %w", err)
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
		"-e", "HASURA_GRAPHQL_ENABLE_TELEMETRY=false",
		"-w", "/app",
		"-it", "--rm",
		"--entrypoint", "hasura-cli",
		fmt.Sprintf("hasura/graphql-engine:%s.cli-migrations-v3", hasuraVersion),
		"--endpoint", hasuraEndpoint,
		"--admin-secret", adminSecret,
		"metadata", "export",
		"--skip-update-check",
		"--log-level", "ERROR",
	)

	f, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(os.Stdout, f); err != nil {
		return fmt.Errorf("failed to copy output: %w", err)
	}

	return nil
}
