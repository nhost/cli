package dockercredentials

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nhost/cli/clienv"
	"github.com/urfave/cli/v2"
)

const (
	flagDomain     = "domain"
	flagAppBaseURL = "app-base-url"
)

func CommandGet() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:    "get",
		Aliases: []string{},
		Usage:   "Get credentials for the logged in user",
		Hidden:  true,
		Flags: []cli.Flag{
			&cli.StringFlag{ //nolint:exhaustruct
				Name:    flagDomain,
				Usage:   "Nhost domain",
				EnvVars: []string{"NHOST_DOMAIN"},
				Value:   "nhost.run",
				Hidden:  true,
			},
			&cli.StringFlag{ //nolint:exhaustruct
				Name:    flagAppBaseURL,
				Usage:   "Nhost app base URL",
				EnvVars: []string{"NHOST_APP_BASE_URL"},
				Value:   "https://app.nhost.io",
				Hidden:  true,
			},
		},
		Action: actionGet,
	}
}

func getToken(ctx context.Context, domain, appBaseURL string) (string, error) {
	ce := clienv.New(
		os.Stdout,
		os.Stderr,
		&clienv.PathStructure{},
		domain,
		appBaseURL,
		"unneeded",
		"unneeded",
	)
	session, err := ce.LoadSession(ctx)
	if err != nil {
		return "", err //nolint:wrapcheck
	}

	return session.Session.AccessToken, nil
}

//nolint:tagliatelle
type response struct {
	ServerURL string `json:"ServerURL"`
	Username  string `json:"Username"`
	Secret    string `json:"Secret"`
}

func actionGet(c *cli.Context) error {
	scanner := bufio.NewScanner(c.App.Reader)
	var input string
	for scanner.Scan() {
		input += scanner.Text()
	}
	token, err := getToken(c.Context, c.String(flagDomain), c.String(flagAppBaseURL))
	if err != nil {
		return err
	}

	b, err := json.Marshal(response{
		ServerURL: input,
		Username:  "nhost",
		Secret:    token,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	if _, err = c.App.Writer.Write(b); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}
