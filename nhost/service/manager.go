package service

import (
	"context"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/nhost/cli/nhost"
	"github.com/nhost/cli/nhost/compose"
	"github.com/nhost/cli/util"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"sync"
	"time"
)

type Ports struct {
	Console  int
	GraphQL  int
	Postgres int
}

const (
	retryCount = 3
)

type Manager interface {
	SyncExec(ctx context.Context, f func(ctx context.Context) error) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	StopSvc(ctx context.Context, svc string) error
	SetGitBranch(string)
	Restart(ctx context.Context) error
	IsServiceHealthy(ctx context.Context, svc string) (bool, error)
	IsStackReady(ctx context.Context) (bool, error)
}

func NewDockerComposeManager(c *nhost.Configuration, gitBranch string, projectName string, logger logrus.FieldLogger, status *util.Status, debug bool) *dockerComposeManager {
	if gitBranch == "" {
		gitBranch = "main"
	}

	return &dockerComposeManager{debug: debug, branch: gitBranch, projectName: projectName, nhostConfig: c, composeConfig: compose.NewConfig(c, gitBranch, projectName), l: logger, status: status}
}

type dockerComposeManager struct {
	sync.Mutex
	debug         bool
	branch        string
	projectName   string
	nhostConfig   *nhost.Configuration
	composeConfig *compose.Config
	status        *util.Status
	l             logrus.FieldLogger
}

func (m *dockerComposeManager) SyncExec(ctx context.Context, f func(ctx context.Context) error) error {
	m.Lock()
	defer m.Unlock()

	return f(ctx)
}

func (m *dockerComposeManager) SetGitBranch(gitBranch string) {
	if m.branch == gitBranch {
		return
	}

	m.branch = gitBranch
	m.composeConfig = compose.NewConfig(m.nhostConfig, gitBranch, m.projectName)
}

func (m *dockerComposeManager) Start(ctx context.Context) error {
	ds := compose.DataStreams{}
	if m.debug {
		ds.Stdout = os.Stdout
		ds.Stderr = os.Stderr
	}

	m.status.Executing("Starting nhost app...")
	m.l.Debug("Starting docker compose")
	cmd, err := compose.WrapperCmd(ctx, []string{"up", "-d"}, m.composeConfig, ds)
	if err != nil {
		m.status.Error("Failed to start nhost app")
		m.l.WithError(err).Debug("Failed to start docker compose")
		return err
	}

	err = cmd.Run()
	if err != nil {
		m.status.Error("Failed to start nhost app")
		m.l.WithError(err).Debug("Failed to start docker compose")
		return err
	}

	err = m.waitForGraphqlEngine(ctx, time.Millisecond*100, time.Minute*2)
	if err != nil {
		m.status.Error("Timed out waiting for graphql-engine service to be ready")
		m.l.WithError(err).Debug("Timed out waiting for graphql-engine service to be ready")
		return err
	}

	// migrations
	{
		files, err := os.ReadDir(nhost.MIGRATIONS_DIR)
		if err != nil {
			return err
		}

		if len(files) > 0 {
			err = m.applyMigrations(ctx, ds)
			if err != nil {
				m.status.Error("Failed to apply migrations")
				m.l.WithError(err).Debug("Failed to apply migrations")
				return err
			}
		}
	}

	// metadata
	{
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		metaFiles, err := os.ReadDir(nhost.METADATA_DIR)
		if err != nil {
			return err
		}

		if len(metaFiles) == 0 {
			err = m.exportMetadata(ctx, ds)
			if err != nil {
				m.status.Error("Failed to export metadata")
				m.l.WithError(err).Debug("Failed to export metadata")
				return err
			}
		}

		err = m.applyMetadata(ctx, ds)
		if err != nil {
			m.status.Error("Failed to apply metadata")
			m.l.WithError(err).Debug("Failed to apply metadata")
			return err
		}
	}

	err = m.restartAuthStorageContainers(ctx, ds)

	// export metadata again
	err = m.exportMetadata(ctx, ds)
	if err != nil {
		m.status.Error("Failed to export metadata")
		m.l.WithError(err).Debug("Failed to export metadata")
		return err
	}

	m.status.Info("Ready to use")
	return nil
}

func (m *dockerComposeManager) Stop(ctx context.Context) error {
	m.l.Debug("Stopping docker compose")
	cmd, err := compose.WrapperCmd(ctx, []string{"stop"}, m.composeConfig, compose.DataStreams{})
	if err != nil {
		m.l.WithError(err).Debug("Failed to stop docker compose")
		return err
	}

	return cmd.Run()
}

func (m *dockerComposeManager) StopSvc(ctx context.Context, svc string) error {
	m.status.Executing(fmt.Sprintf("Stopping service %s", svc))
	m.l.Debugf("Stopping %s service", svc)
	cmd, err := compose.WrapperCmd(ctx, []string{"stop", svc}, m.composeConfig, compose.DataStreams{})
	if err != nil {
		m.l.WithError(err).Debugf("Failed to stop %s service", svc)
		return err
	}

	return cmd.Run()
}

func (m *dockerComposeManager) Restart(ctx context.Context) error {
	m.l.Debug("Stopping postgres service")
	cmd, err := compose.WrapperCmd(ctx, []string{"stop", "postgres"}, m.composeConfig, compose.DataStreams{})
	if err != nil {
		return fmt.Errorf("failed to stop postgres service: %w", err)
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to stop postgres service: %w", err)
	}

	err = m.Stop(ctx)
	if err != nil {
		m.l.WithError(err).Debug("Failed to restart docker compose")
		return err
	}

	return m.Start(ctx)
}

func (m *dockerComposeManager) IsServiceHealthy(ctx context.Context, svc string) (bool, error) {
	return true, nil
}

func (m *dockerComposeManager) IsStackReady(ctx context.Context) (bool, error) {
	return true, nil
}

func (m *dockerComposeManager) restartAuthStorageContainers(ctx context.Context, ds compose.DataStreams) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.l.Debug("Restarting auth and storage containers")
	c, err := compose.WrapperCmd(ctx, []string{"restart", "auth", "storage"}, m.composeConfig, ds)
	if err != nil {
		return fmt.Errorf("failed to restart auth and storage containers: %w", err)
	}

	return c.Run()
}

func (m *dockerComposeManager) applyMigrations(ctx context.Context, ds compose.DataStreams) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.status.Executing("Applying migrations...")
	err := retry.Do(func() error {
		m.l.Debug("Applying migrations")
		migrate, err := compose.WrapperCmd(
			ctx,
			[]string{"exec", "hasura-console", "hasura", "migrate", "apply", "--database-name", "default", "--disable-interactive", "--skip-update-check"},
			m.composeConfig,
			ds,
		)
		if err != nil {
			return fmt.Errorf("Failed to apply migrations: %w", err)
		}

		err = migrate.Run()
		if err != nil {
			return fmt.Errorf("Failed to apply migrations: %w", err)
		}

		return nil
	}, retry.Attempts(retryCount), retry.OnRetry(func(n uint, err error) {
		m.l.Debugf("Retrying migration apply: attempt %d\n", n)
	}))

	return err
}

func (m *dockerComposeManager) exportMetadata(ctx context.Context, ds compose.DataStreams) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.status.Executing("Exporting metadata...")
	err := retry.Do(func() error {
		m.l.Debug("Exporting metadata")
		export, err := compose.WrapperCmd(
			ctx,
			[]string{"exec", "hasura-console", "hasura", "--skip-update-check", "metadata", "export"},
			m.composeConfig,
			ds,
		)
		if err != nil {
			return fmt.Errorf("failed to export metadata: %w", err)
		}

		err = export.Run()
		if err != nil {
			return fmt.Errorf("failed to export metadata: %w", err)
		}

		return nil
	}, retry.Attempts(retryCount), retry.OnRetry(func(n uint, err error) {
		m.l.Debugf("Retrying metadata export: attempt %d\n", n)
	}))

	return err
}

func (m *dockerComposeManager) applyMetadata(ctx context.Context, ds compose.DataStreams) error {
	m.status.Executing("Applying metadata...")
	err := retry.Do(func() error {
		m.l.Debug("Applying metadata")
		export, err := compose.WrapperCmd(
			ctx,
			[]string{"exec", "hasura-console", "hasura", "--skip-update-check", "metadata", "apply"},
			m.composeConfig,
			ds,
		)
		if err != nil {
			return fmt.Errorf("failed to apply metadata: %w", err)
		}

		err = export.Run()
		if err != nil {
			return fmt.Errorf("failed to apply metadata: %w", err)
		}

		return nil
	}, retry.Attempts(retryCount), retry.OnRetry(func(n uint, err error) {
		m.l.Debugf("Retrying metadata apply: attempt %d\n", n)
	}))

	return err
}

func (m *dockerComposeManager) hasuraHealthcheck() (bool, error) {
	// curl /healthz and check for 200
	resp, err := http.Get("http://localhost:8080/healthz")
	if err != nil {
		return false, err
	}

	return resp.StatusCode == 200, nil
}

func (m *dockerComposeManager) waitForGraphqlEngine(ctx context.Context, interval time.Duration, timeout time.Duration) error {
	m.status.Executing("Waiting for graphql-engine service to be ready...")
	m.l.Debug("Waiting for graphql-engine service to be ready")

	t := time.After(timeout)

	ticker := time.NewTicker(interval)

	for range ticker.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t:
			return fmt.Errorf("timeout: graphql-engine not ready, please run the command again")
		default:
			if ok, err := m.hasuraHealthcheck(); err == nil && ok {
				return nil
			}
		}
	}

	return nil
}
