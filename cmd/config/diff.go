package config

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/nhostclient/credentials"
	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v2"
)

func CommandDiff() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:        "diff",
		Aliases:     []string{},
		Usage:       "Shows difference between local config and remote config (applies overlays automatically for the subdomain, if any)", //nolint:lll
		Description: "Note that this command will always use the local secrets, even if you specify subdomain",
		Action:      commandDiff,
		Flags: []cli.Flag{
			&cli.StringFlag{ //nolint:exhaustruct
				Name:    flagSubdomain,
				Usage:   "Diff config for a specific subdomain, defaults to linked project",
				EnvVars: []string{"NHOST_SUBDOMAIN"},
			},
		},
	}
}

func getRemoteConfig(
	ctx context.Context,
	ce *clienv.CliEnv,
	subdomain string,
	session credentials.Session,
) (*model.ConfigConfig, error) {
	apps, err := ce.GetAppInfo(ctx, subdomain)
	if err != nil {
		return nil, fmt.Errorf("error getting app info: %w", err)
	}

	cfg, err := pullConfigFromSubdomain(ctx, ce, apps.ID, session)
	if err != nil {
		return nil, fmt.Errorf("remote: %w", err)
	}

	return cfg, nil
}

func commandDiff(c *cli.Context) error {
	ce := clienv.FromCLI(c)

	subdomain := c.String(flagSubdomain)
	if subdomain == "" {
		app, err := ce.GetAppInfo(c.Context, "")
		if err != nil {
			return err //nolint:wrapcheck
		}
		subdomain = app.GetSubdomain()
	}

	session, err := ce.LoadSession(c.Context)
	if err != nil {
		return fmt.Errorf("error getting session: %w", err)
	}

	cfg, err := getRemoteConfig(c.Context, ce, subdomain, session)
	if err != nil {
		return err
	}
	remoteb, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	cfg, err = Validate(ce, subdomain, false)
	if err != nil {
		return fmt.Errorf("local: %w", err)
	}
	localb, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	d, err := diffContent(localb, remoteb)
	if err != nil {
		return err
	}
	if len(d) == 0 {
		ce.Println("No differences found")
		return nil
	}

	ce.Println(string(d))

	return nil
}

func savetodisk(path string, r io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func diffContent(local, remote []byte) ([]byte, error) {
	tmpdir, err := os.MkdirTemp("", "nhost-diff")
	if err != nil {
		return nil, fmt.Errorf("error creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpdir)

	localFile := fmt.Sprintf("%s/local.toml", tmpdir)
	if err := savetodisk(localFile, bytes.NewReader(local)); err != nil {
		return nil, err
	}

	remoteFile := fmt.Sprintf("%s/remote.toml", tmpdir)
	if err := savetodisk(remoteFile, bytes.NewReader(remote)); err != nil {
		return nil, err
	}

	cmd := exec.Command("diff", "-u", localFile, remoteFile)
	b, _ := cmd.CombinedOutput()
	return b, nil
}
