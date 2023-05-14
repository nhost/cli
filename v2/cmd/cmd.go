package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/nhost/cli/v2/system"
	"github.com/spf13/cobra"
)

const (
	flagDomain          = "domain"
	flagRemote          = "remote"
	flagHTTPPort        = "http-port"
	flagDisableTLS      = "disable-tls"
	flagProjectName     = "project-name"
	flagPostgresPort    = "postgres-port"
	flagRootFolder      = "root-folder"
	flagDataFolder      = "data-folder"
	flagNhostFolder     = "nhost-folder"
	flagDotNhostFolder  = "dot-nhost-folder"
	flagFunctionsFolder = "functions-folder"
)

const (
	defaultPostgresPort = 5432
	defaultHTTPSPort    = 443
	defaultProjectName  = "nhost"
)

func getGitBranchName() string {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "nogit"
	}

	head, err := repo.Head()
	if err != nil {
		return "nogit"
	}

	return head.Name().Short()
}

func getFolders(cmd *cobra.Command) (*system.PathStructure, error) {
	rootFolder, err := cmd.Flags().GetString(flagRootFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to parse root folder: %w", err)
	}

	dotNhostFolder, err := cmd.Flags().GetString(flagDotNhostFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .nhost folder: %w", err)
	}

	dataFolder, err := cmd.Flags().GetString(flagDataFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data folder: %w", err)
	}

	nhostFolder, err := cmd.Flags().GetString(flagNhostFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nhost folder: %w", err)
	}

	functionsFolder, err := cmd.Flags().GetString(flagFunctionsFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to parse functions folder: %w", err)
	}

	return system.NewPathStructure(
		rootFolder,
		dotNhostFolder,
		dataFolder,
		functionsFolder,
		nhostFolder,
	), nil
}

func Register(rootCmd *cobra.Command) { //nolint:funlen
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dotNhostFolder := filepath.Join(curDir, ".nhost")
	dataFolder := filepath.Join(dotNhostFolder, "data", getGitBranchName())
	nhostFolder := filepath.Join(curDir, "nhost")
	functionsFolder := filepath.Join(curDir, "functions")

	rootCmd.Flags().StringP(flagRootFolder, "", "", "Root folder of project")
	rootCmd.Flags().StringP(flagDotNhostFolder, "", dotNhostFolder, "Path to .nhost folder")
	rootCmd.Flags().
		StringP(flagDataFolder, "", dataFolder, "Data folder to persist data. Defaults to ")
	rootCmd.Flags().StringP(flagNhostFolder, "", nhostFolder, "Path to nhost folder")
	rootCmd.Flags().StringP(flagFunctionsFolder, "", functionsFolder, "Path to functions folder")

	{
		logoutCmd := logoutCmd()
		rootCmd.AddCommand(logoutCmd)
	}

	{
		secretsCmd := secretsCmd()
		rootCmd.AddCommand(secretsCmd)

		secretsListCmd := secretsListCmd()
		secretsCmd.AddCommand(secretsListCmd)
		secretsCreateCmd := secretsCreateCmd()
		secretsCmd.AddCommand(secretsCreateCmd)
		secretsDeleteCmd := secretsDeleteCmd()
		secretsCmd.AddCommand(secretsDeleteCmd)
		secretsUpdateCmd := secretsUpdateCmd()
		secretsCmd.AddCommand(secretsUpdateCmd)
	}
}
