package dockercompose

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"gopkg.in/yaml.v2"
)

type DockerCompose struct {
	filepath    string
	projectName string
}

func New(filepath, projectName string) *DockerCompose {
	return &DockerCompose{
		filepath:    filepath,
		projectName: projectName,
	}
}

func (dc *DockerCompose) WriteComposeFile(composeFile *ComposeFile) error {
	f, err := os.Create(dc.filepath)
	if err != nil {
		return fmt.Errorf("failed to create docker-compose file: %w", err)
	}
	defer f.Close()

	b, err := yaml.Marshal(composeFile)
	if err != nil {
		return fmt.Errorf("failed to marshal docker-compose file: %w", err)
	}

	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("failed to write docker-compose file: %w", err)
	}

	return nil
}

func (dc *DockerCompose) Start(ctx context.Context) error {
	cmd := exec.CommandContext( //nolint:gosec
		ctx,
		"docker-compose",
		"-f", dc.filepath,
		"-p", dc.projectName,
		"up",
		"-d", "--wait",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start docker-compose: %w", err)
	}
	return nil
}

func (dc *DockerCompose) Stop(ctx context.Context) error {
	cmd := exec.CommandContext( //nolint:gosec
		ctx,
		"docker-compose",
		"-f", dc.filepath,
		"-p", dc.projectName,
		"down",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop docker-compose: %w", err)
	}
	return nil
}

func (dc *DockerCompose) Logs(ctx context.Context, extraArgs ...string) error {
	args := []string{
		"-f", dc.filepath,
		"-p", dc.projectName,
		"logs",
	}
	args = append(args, extraArgs...)

	cmd := exec.CommandContext(
		ctx,
		"docker-compose",
		args...,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to show logs from docker-compose: %w", err)
	}
	return nil
}

func (dc *DockerCompose) Wrapper(ctx context.Context, extraArgs ...string) error {
	args := []string{
		"-f", dc.filepath,
		"-p", dc.projectName,
	}
	args = append(args, extraArgs...)

	cmd := exec.CommandContext(
		ctx,
		"docker-compose",
		args...,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run docker-compose: %w", err)
	}
	return nil
}

func (dc *DockerCompose) ApplyMetadata(ctx context.Context) error {
	cmd := exec.CommandContext( //nolint:gosec
		ctx,
		"docker-compose",
		"-f", dc.filepath,
		"-p", dc.projectName,
		"exec",
		"console",
		"hasura-cli",
		"metadata", "apply",
		"--endpoint", "http://graphql:8080",
		"--skip-update-check",
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

func (dc *DockerCompose) ApplyMigrations(ctx context.Context) error {
	cmd := exec.CommandContext( //nolint:gosec
		ctx,
		"docker-compose",
		"-f", dc.filepath,
		"-p", dc.projectName,
		"exec",
		"console",
		"hasura-cli",
		"migrate",
		"apply",
		"--endpoint", "http://graphql:8080",
		"--all-databases",
		"--skip-update-check",
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