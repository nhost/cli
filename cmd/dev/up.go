package dev

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/clienv"
	"github.com/nhost/cli/cmd/config"
	"github.com/nhost/cli/cmd/run"
	"github.com/nhost/cli/dockercompose"
	"github.com/nhost/cli/project/env"
	"github.com/urfave/cli/v2"
)

func deptr[T any](t *T) T { //nolint:ireturn
	if t == nil {
		return *new(T)
	}
	return *t
}

const (
	flagHTTPPort           = "http-port"
	flagDisableTLS         = "disable-tls"
	flagPostgresPort       = "postgres-port"
	flagApplySeeds         = "apply-seeds"
	flagAuthPort           = "auth-port"
	flagStoragePort        = "storage-port"
	flagsFunctionsPort     = "functions-port"
	flagsHasuraPort        = "hasura-port"
	flagsHasuraConsolePort = "hasura-console-port"
	flagDashboardVersion   = "dashboard-version"
	flagConfigserverImage  = "configserver-image"
	flagRunService         = "run-service"
)

const (
	defaultHTTPPort     = 443
	defaultPostgresPort = 5432
)

func CommandUp() *cli.Command { //nolint:funlen
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
			&cli.BoolFlag{ //nolint:exhaustruct
				Name:    flagApplySeeds,
				Usage:   "Apply seeds. If the .nhost folder does not exist, seeds will be applied regardless of this flag",
				Value:   false,
				EnvVars: []string{"NHOST_APPLY_SEEDS"},
			},
			&cli.UintFlag{ //nolint:exhaustruct
				Name:  flagAuthPort,
				Usage: "If specified, expose auth on this port. Not recommended",
				Value: 0,
			},
			&cli.UintFlag{ //nolint:exhaustruct
				Name:  flagStoragePort,
				Usage: "If specified, expose storage on this port. Not recommended",
				Value: 0,
			},
			&cli.UintFlag{ //nolint:exhaustruct
				Name:  flagsFunctionsPort,
				Usage: "If specified, expose functions on this port. Not recommended",
				Value: 0,
			},
			&cli.UintFlag{ //nolint:exhaustruct
				Name:  flagsHasuraPort,
				Usage: "If specified, expose hasura on this port. Not recommended",
				Value: 0,
			},
			&cli.UintFlag{ //nolint:exhaustruct
				Name:  flagsHasuraConsolePort,
				Usage: "If specified, expose hasura console on this port. Not recommended",
				Value: 0,
			},
			&cli.StringFlag{ //nolint:exhaustruct
				Name:    flagDashboardVersion,
				Usage:   "Dashboard version to use",
				Value:   "nhost/dashboard:1.13.0",
				EnvVars: []string{"NHOST_DASHBOARD_VERSION"},
			},
			&cli.StringFlag{ //nolint:exhaustruct
				Name:    flagConfigserverImage,
				Hidden:  true,
				Value:   "",
				EnvVars: []string{"NHOST_CONFIGSERVER_IMAGE"},
			},
			&cli.StringSliceFlag{ //nolint:exhaustruct
				Name:    flagRunService,
				Usage:   "Run service to add to the development environment. Can be passed multiple times. Comma-separated values are also accepted. Format: /path/to/run-service.toml[:overlay_name]", //nolint:lll
				EnvVars: []string{"NHOST_RUN_SERVICE"},
			},
		},
	}
}

func commandUp(cCtx *cli.Context) error {
	ce := clienv.FromCLI(cCtx)

	// projname to be root directory

	if !clienv.PathExists(ce.Path.NhostToml()) {
		return fmt.Errorf( //nolint:goerr113
			"no nhost project found, please run `nhost init` or `nhost config pull`",
		)
	}
	if !clienv.PathExists(ce.Path.Secrets()) {
		return fmt.Errorf( //nolint:goerr113
			"no secrets found, please run `nhost init` or `nhost config pull`",
		)
	}

	configserverImage := cCtx.String(flagConfigserverImage)
	if configserverImage == "" {
		configserverImage = "nhost/cli:" + cCtx.App.Version
	}

	applySeeds := cCtx.Bool(flagApplySeeds) || !clienv.PathExists(ce.Path.DotNhostFolder())
	return Up(
		cCtx.Context,
		ce,
		cCtx.Uint(flagHTTPPort),
		!cCtx.Bool(flagDisableTLS),
		cCtx.Uint(flagPostgresPort),
		applySeeds,
		dockercompose.ExposePorts{
			Auth:      cCtx.Uint(flagAuthPort),
			Storage:   cCtx.Uint(flagStoragePort),
			Graphql:   cCtx.Uint(flagsHasuraPort),
			Console:   cCtx.Uint(flagsHasuraConsolePort),
			Functions: cCtx.Uint(flagsFunctionsPort),
		},
		cCtx.String(flagDashboardVersion),
		configserverImage,
		cCtx.StringSlice(flagRunService),
	)
}

func migrations(
	ctx context.Context,
	ce *clienv.CliEnv,
	dc *dockercompose.DockerCompose,
	applySeeds bool,
) error {
	if clienv.PathExists(filepath.Join(ce.Path.NhostFolder(), "migrations", "default")) {
		ce.Infoln("Applying migrations...")
		if err := dc.ApplyMigrations(ctx); err != nil {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
	} else {
		ce.Warnln("No migrations found, make sure this is intentional or it could lead to unexpected behavior")
	}

	if clienv.PathExists(filepath.Join(ce.Path.NhostFolder(), "metadata", "version.yaml")) {
		ce.Infoln("Applying metadata...")
		if err := dc.ApplyMetadata(ctx); err != nil {
			return fmt.Errorf("failed to apply metadata: %w", err)
		}
	} else {
		ce.Warnln("No metadata found, make sure this is intentional or it could lead to unexpected behavior")
	}

	if applySeeds {
		if clienv.PathExists(filepath.Join(ce.Path.NhostFolder(), "seeds", "default")) {
			ce.Infoln("Applying seeds...")
			if err := dc.ApplySeeds(ctx); err != nil {
				return fmt.Errorf("failed to apply seeds: %w", err)
			}
		}
	}

	return nil
}

func restart(
	ctx context.Context,
	ce *clienv.CliEnv,
	dc *dockercompose.DockerCompose,
	composeFile *dockercompose.ComposeFile,
) error {
	ce.Infoln("Restarting services to reapply metadata if needed...")

	args := []string{"restart", "storage", "functions"}
	if _, ok := composeFile.Services["auth"]; ok {
		args = append(args, "auth")
	}
	if err := dc.Wrapper(ctx, args...); err != nil {
		return fmt.Errorf("failed to restart services: %w", err)
	}

	return nil
}

func reload(
	ctx context.Context,
	ce *clienv.CliEnv,
	dc *dockercompose.DockerCompose,
) error {
	ce.Infoln("Reapplying metadata...")
	if err := dc.ReloadMetadata(ctx); err != nil {
		return fmt.Errorf("failed to reapply metadata: %w", err)
	}

	return nil
}

func parseRunServiceConfigFlag(value string) (string, string, error) {
	parts := strings.Split(value, ":")
	switch len(parts) {
	case 1:
		return parts[0], "", nil
	case 2: //nolint:gomnd
		return parts[0], parts[1], nil
	default:
		return "", "", fmt.Errorf( //nolint:goerr113
			"invalid run service format, must be /path/to/config.toml:overlay_name, got %s",
			value,
		)
	}
}

func processRunServices(
	ce *clienv.CliEnv,
	runServices []string,
	secrets model.Secrets,
) ([]*dockercompose.RunService, error) {
	r := make([]*dockercompose.RunService, 0, len(runServices))
	for _, runService := range runServices {
		cfgPath, overlayName, err := parseRunServiceConfigFlag(runService)
		if err != nil {
			return nil, err
		}

		cfg, err := run.Validate(ce, cfgPath, overlayName, secrets, false)
		if err != nil {
			return nil, fmt.Errorf("failed to validate run service %s: %w", cfgPath, err)
		}

		r = append(r, &dockercompose.RunService{
			Path:   cfgPath,
			Config: cfg,
		})
	}

	return r, nil
}

func up( //nolint:funlen,cyclop
	ctx context.Context,
	ce *clienv.CliEnv,
	dc *dockercompose.DockerCompose,
	httpPort uint,
	useTLS bool,
	postgresPort uint,
	applySeeds bool,
	ports dockercompose.ExposePorts,
	dashboardVersion string,
	configserverImage string,
	runServices []string,
) error {
	ctx, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-sigChan
		cancel()
	}()

	var secrets model.Secrets
	if err := clienv.UnmarshalFile(ce.Path.Secrets(), &secrets, env.Unmarshal); err != nil {
		return fmt.Errorf(
			"failed to parse secrets, make sure secret values are between quotes: %w",
			err,
		)
	}

	cfg, err := config.Validate(ce, "local", secrets)
	if err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	runServicesCfg, err := processRunServices(ce, runServices, secrets)
	if err != nil {
		return err
	}

	ce.Infoln("Setting up Nhost development environment...")
	composeFile, err := dockercompose.ComposeFileFromConfig(
		cfg,
		ce.ProjectName(),
		httpPort,
		useTLS,
		postgresPort,
		ce.Path.DataFolder(),
		ce.Path.NhostFolder(),
		ce.Path.DotNhostFolder(),
		ce.Path.Root(),
		ports,
		ce.Branch(),
		dashboardVersion,
		configserverImage,
		runServicesCfg...,
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

	if err := migrations(ctx, ce, dc, applySeeds); err != nil {
		return err
	}

	if err := restart(ctx, ce, dc, composeFile); err != nil {
		return err
	}

	docker := dockercompose.NewDocker()
	ce.Infoln("Downloading metadata...")
	if err := docker.HasuraWrapper(
		ctx, ce.Path.NhostFolder(),
		*cfg.Hasura.Version,
		"metadata", "export",
		"--skip-update-check",
		"--log-level", "ERROR",
		"--endpoint", dockercompose.URL("hasura", httpPort, useTLS),
		"--admin-secret", cfg.Hasura.AdminSecret,
	); err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}

	if err := reload(ctx, ce, dc); err != nil {
		return err
	}

	ce.Infoln("Nhost development environment started.")
	printInfo(httpPort, postgresPort, useTLS, runServicesCfg)
	return nil
}

func printInfo(
	httpPort, postgresPort uint,
	useTLS bool,
	runServices []*dockercompose.RunService,
) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0) //nolint:gomnd
	fmt.Fprintf(w, "URLs:\n")
	fmt.Fprintf(w,
		"- Postgres:\t\tpostgres://postgres:postgres@localhost:%d/local\n",
		postgresPort,
	)
	fmt.Fprintf(w, "- Hasura:\t\t%s\n", dockercompose.URL("hasura", httpPort, useTLS))
	fmt.Fprintf(w, "- GraphQL:\t\t%s\n", dockercompose.URL("graphql", httpPort, useTLS))
	fmt.Fprintf(w, "- Auth:\t\t%s\n", dockercompose.URL("auth", httpPort, useTLS))
	fmt.Fprintf(w, "- Storage:\t\t%s\n", dockercompose.URL("storage", httpPort, useTLS))
	fmt.Fprintf(w, "- Functions:\t\t%s\n", dockercompose.URL("functions", httpPort, useTLS))
	fmt.Fprintf(w, "- Dashboard:\t\t%s\n", dockercompose.URL("dashboard", httpPort, useTLS))
	fmt.Fprintf(w, "- Mailhog:\t\t%s\n", dockercompose.URL("mailhog", httpPort, useTLS))

	for _, svc := range runServices {
		for _, port := range svc.Config.GetPorts() {
			if deptr(port.GetPublish()) {
				fmt.Fprintf(
					w,
					"- run-%s:\t\tFrom laptop:\t%s://localhost:%d\n",
					svc.Config.Name,
					port.GetType(),
					port.GetPort(),
				)
				fmt.Fprintf(
					w,
					"\t\tFrom services:\t%s://run-%s:%d\n",
					port.GetType(),
					svc.Config.Name,
					port.GetPort(),
				)
			}
		}
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "SDK Configuration:\n")
	fmt.Fprintf(w, " Subdomain:\tlocal\n")
	fmt.Fprintf(w, " Region:\t(empty)\n")
	fmt.Fprintf(w, "")
	fmt.Fprintf(w, "Run `nhost up` to reload the development environment\n")
	fmt.Fprintf(w, "Run `nhost down` to stop the development environment\n")
	fmt.Fprintf(w, "Run `nhost logs` to watch the logs\n")

	w.Flush()
}

func Up(
	ctx context.Context,
	ce *clienv.CliEnv,
	httpPort uint,
	useTLS bool,
	postgresPort uint,
	applySeeds bool,
	ports dockercompose.ExposePorts,
	dashboardVersion string,
	configserverImage string,
	runServices []string,
) error {
	dc := dockercompose.New(ce.Path.WorkingDir(), ce.Path.DockerCompose(), ce.ProjectName())

	if err := up(
		ctx, ce, dc, httpPort, useTLS, postgresPort, applySeeds, ports, dashboardVersion, configserverImage, runServices,
	); err != nil {
		ce.Warnln(err.Error())

		ce.PromptMessage("Do you want to stop Nhost's development environment? [y/N] ")
		resp, err := ce.PromptInput(false)
		if err != nil {
			ce.Warnln("failed to read input: %s", err)
			return nil
		}
		if resp != "y" && resp != "Y" {
			return nil
		}

		ce.Infoln("Stopping Nhost development environment...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		if err := dc.Stop(ctx, false); err != nil { //nolint:contextcheck
			ce.Warnln("failed to stop Nhost development environment: %s", err)
		}

		return err //nolint:wrapcheck
	}

	return nil
}
