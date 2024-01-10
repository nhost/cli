package clienv

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/nhost/cli/nhostclient/credentials"
)

func savePAT(
	ce *CliEnv,
	session credentials.Credentials,
) error {
	dir := filepath.Dir(ce.Path.AuthFile())
	if !PathExists(dir) {
		if err := os.MkdirAll(dir, 0o755); err != nil { //nolint:gomnd
			return fmt.Errorf("failed to create dir: %w", err)
		}
	}

	if err := MarshalFile(session, ce.Path.AuthFile(), json.Marshal); err != nil {
		return fmt.Errorf("failed to write PAT to file: %w", err)
	}

	return nil
}

func signinHandler(ch chan<- string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ch <- r.URL.Query().Get("refreshToken")
		fmt.Fprintf(w, "You may now close this window.")
	}
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	if err := exec.Command(cmd, args...).Start(); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	return nil
}

func (ce *CliEnv) Login(
	ctx context.Context,
	pat string,
) (credentials.Credentials, error) {
	if pat != "" {
		session := credentials.Credentials{
			ID:                  "",
			PersonalAccessToken: pat,
		}
		if err := savePAT(ce, session); err != nil {
			return credentials.Credentials{}, err
		}
		return session, nil
	}

	refreshToken := make(chan string)
	http.HandleFunc("/signin", signinHandler(refreshToken))
	go func() {
		if err := http.ListenAndServe(":8099", nil); err != nil { //nolint:gosec
			log.Fatal(err)
		}
	}()

	signinPage := fmt.Sprintf("%s/signin?redirectTo=http://localhost:8099/signin", ce.AppBaseURL())
	ce.Infoln("Opening browser to sign-in")
	if err := openBrowser(signinPage); err != nil {
		return credentials.Credentials{}, err
	}
	ce.Infoln("Waiting for sign-in to complete")

	refreshTokenValue := <-refreshToken

	cl := ce.GetNhostClient()
	refreshTokenResp, err := cl.RefreshToken(ctx, refreshTokenValue)
	if err != nil {
		return credentials.Credentials{}, fmt.Errorf("failed to get access token: %w", err)
	}
	ce.Infoln("Successfully logged in, creating PAT")
	session, err := cl.CreatePAT(ctx, refreshTokenResp.AccessToken)
	if err != nil {
		return credentials.Credentials{}, fmt.Errorf("failed to create PAT: %w", err)
	}
	ce.Infoln("Successfully created PAT")
	ce.Infoln("Storing PAT for future user")

	if err := savePAT(ce, session); err != nil {
		return credentials.Credentials{}, err
	}

	return session, nil
}
