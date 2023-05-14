package dev

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/cmd/config"
	"github.com/nhost/cli/v2/dockercompose"
	"github.com/urfave/cli/v2"
)

const (
	flagHTTPPort     = "http-port"
	flagDisableTLS   = "disable-tls"
	flagPostgresPort = "postgres-port"
	flagProjectName  = "project-name"
)

const (
	defaultHTTPPort     = 443
	defaultPostgresPort = 5432
)

func CommandUp() *cli.Command {
	return &cli.Command{ //nolint:exhaustruct
		Name:    "up",
		Aliases: []string{},
		Usage:   "Start local development environment",
		Action:  commandUp,
		Flags: []cli.Flag{
			&cli.UintFlag{ //nolint:exhaustruct
				Name:    flagHTTPPort,
				Usage:   "HTTP port to listen on",
				Value:   defaultHTTPPort,
				EnvVars: []string{"NHOST_HTTP_PORT"},
			},
			&cli.BoolFlag{ //nolint:exhaustruct
				Name:    flagDisableTLS,
				Usage:   "Disable TLS",
				Value:   false,
				EnvVars: []string{"NHOST_DISABLE_TLS"},
			},
			&cli.UintFlag{ //nolint:exhaustruct
				Name:    flagPostgresPort,
				Usage:   "Postgres port to listen on",
				Value:   defaultPostgresPort,
				EnvVars: []string{"NHOST_POSTGRES_PORT"},
			},
			&cli.StringFlag{ //nolint:exhaustruct
				Name:    flagProjectName,
				Usage:   "Project name",
				Value:   "nhost",
				EnvVars: []string{"NHOST_PROJECT_NAME"},
			},
		},
	}
}

func commandUp(cCtx *cli.Context) error {
	ce := clienv.New(cCtx)

	if !clienv.PathExists(ce.Path.NhostToml()) {
		return fmt.Errorf( //nolint:goerr113
			"no nhost project found, please run `nhost init`",
		)
	}
	if !clienv.PathExists(ce.Path.Secrets()) {
		return fmt.Errorf("no secrets found, please run `nhost init`") //nolint:goerr113
	}

	return Up(
		cCtx.Context,
		ce,
		cCtx.String(flagProjectName),
		cCtx.Uint(flagHTTPPort),
		!cCtx.Bool(flagDisableTLS),
		cCtx.Uint(flagPostgresPort),
	)
}

func up(
	ctx context.Context,
	ce *clienv.CliEnv,
	dc *dockercompose.DockerCompose,
	projectName string,
	httpPort uint,
	useTLS bool,
	postgresPort uint,
) error {
	ctx, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-sigChan
		cancel()
	}()

	cfg, err := config.Validate(ce)
	if err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	ce.Infoln("Setting up Nhost development environment...")
	composeFile, err := dockercompose.ComposeFileFromConfig(
		cfg, projectName, httpPort, useTLS, postgresPort,
		ce.Path.DataFolder(), ce.Path.NhostFolder(), ce.Path.DotNhostFolder(), ce.Path.FunctionsFolder(),
	)
	if err != nil {
		return fmt.Errorf("failed to generate docker-compose.yaml: %w", err)
	}
	if err := dc.WriteComposeFile(composeFile); err != nil {
		return fmt.Errorf("failed to write docker-compose.yaml: %w", err)
	}

	ce.Infoln("Starting Nhost development environment...")

	if err = dc.Start(ctx); err != nil {
		return fmt.Errorf("failed to start Nhost development environment: %w", err)
	}

	ce.Infoln("Applying migrations...")
	if err = dc.ApplyMigrations(ctx); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	ce.Infoln("Applying metadata...")
	if err = dc.ApplyMetadata(ctx); err != nil {
		return fmt.Errorf("failed to apply metadata: %w", err)
	}

	ce.Infoln("Nhost development environment started.")

	printInfo(ce, httpPort, useTLS)

	ce.Println("")
	ce.Println("Run `nhost dev up` to reload the development environment")
	ce.Println("Run `nhost dev down` to stop the development environment")
	ce.Println("Run `nhost dev logs` to watch the logs")
	return nil
}

func url(service string, port uint, useTLS bool) string {
	if useTLS && port == 443 {
		return fmt.Sprintf("https://local.%s.nhost.run", service)
	} else if !useTLS && port == 80 {
		return fmt.Sprintf("http://local.%s.nhost.run", service)
	}

	protocol := "http"
	if useTLS {
		protocol = "https"
	}
	return fmt.Sprintf("%s://local.%s.nhost.run:%d", protocol, service, port)
}

func printInfo(ce *clienv.CliEnv, port uint, useTLS bool) {
	ce.Println("URLs:")
	ce.Println("- Postgres:             postgres://postgres:postgres@localhost:5432/postgres")
	ce.Println("- Hasura:               %s", url("hasura", port, useTLS))
	ce.Println("- GraphQL:              %s", url("graphql", port, useTLS))
	ce.Println("- Auth:                 %s", url("auth", port, useTLS))
	ce.Println("- Storage:              %s", url("storage", port, useTLS))
	ce.Println("- Functions:            %s", url("functions", port, useTLS))
	ce.Println("- Dashboard:            %s", url("dashboard", port, useTLS))
	ce.Println("")
	ce.Println("SDK Configuration:")
	ce.Println(" Subdomain:             local")
	ce.Println(" Region:                (empty)")
}

func Up(
	ctx context.Context,
	ce *clienv.CliEnv,
	projectName string,
	httpPort uint,
	useTLS bool,
	postgresPort uint,
) error {
	dc := dockercompose.New(ce.Path.DockerCompose(), projectName)

	if err := up(
		ctx, ce, dc, projectName, httpPort, useTLS, postgresPort,
	); err != nil {
		ce.Warnln(err.Error())

		ce.PromptMessage("Do you want to stop Nhost development environment it? [y/N] ")
		resp, err := ce.PromptInput(false)
		if err != nil {
			ce.Warnln("failed to read input: %s", err)
			return nil
		}
		if resp != "y" {
			return nil
		}

		ce.Infoln("Stopping Nhost development environment...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		if err := dc.Stop(ctx); err != nil { //nolint:contextcheck
			ce.Warnln("failed to stop Nhost development environment: %s", err)
		}

		return err //nolint:wrapcheck
	}

	return nil
}
