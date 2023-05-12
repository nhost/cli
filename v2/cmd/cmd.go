package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

const (
	flagDomain          = "domain"
	flagRemote          = "remote"
	flagHTTPPort        = "http-port"
	flagDisableTLS      = "disable-tls"
	flagProjectName     = "project-name"
	flagPostgresPort    = "postgres-port"
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

func getFolders(cmd *cobra.Command) (string, string, string, string, error) {
	dotNhostFolder, err := cmd.Flags().GetString(flagDotNhostFolder)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to parse .nhost folder: %w", err)
	}

	dataFolder, err := cmd.Flags().GetString(flagDataFolder)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to parse data folder: %w", err)
	}

	nhostFolder, err := cmd.Flags().GetString(flagNhostFolder)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to parse nhost folder: %w", err)
	}

	functionsFolder, err := cmd.Flags().GetString(flagFunctionsFolder)
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to parse functions folder: %w", err)
	}

	return dotNhostFolder, dataFolder, nhostFolder, functionsFolder, nil
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

	rootCmd.Flags().StringP(flagDotNhostFolder, "", dotNhostFolder, "Path to .nhost folder")
	rootCmd.Flags().StringP(flagDataFolder, "", dataFolder, "Data folder to persist data. Defaults to ")
	rootCmd.Flags().StringP(flagNhostFolder, "", nhostFolder, "Path to nhost folder")
	rootCmd.Flags().StringP(flagFunctionsFolder, "", functionsFolder, "Path to functions folder")

	{
		configCmd := configCmd()
		rootCmd.AddCommand(configCmd)

		configPullCmd := configPullCmd()
		configCmd.AddCommand(configPullCmd)

		configShowFullExampleCmd := configShowFullExampleCmd()
		configCmd.AddCommand(configShowFullExampleCmd)

		configValidateCmd := configValidateCmd()
		configCmd.AddCommand(configValidateCmd)
		configValidateCmd.Flags().Bool(
			flagRemote, false, "Validate remote configuration. Defaults to validation of local config.",
		)
	}

	{
		devCmd := devCmd()
		rootCmd.AddCommand(devCmd)
		devCmd.Flags().UintP(flagHTTPPort, "", defaultHTTPSPort, "HTTP port for the local development server")
		devCmd.Flags().BoolP(flagDisableTLS, "", false, "Disable TLS for the local development server")
		devCmd.Flags().UintP(flagPostgresPort, "", defaultPostgresPort, "Postgres port for the local development server")
		devCmd.Flags().StringP(flagProjectName, "", defaultProjectName, "Project name for the local development server")
	}

	{
		downCmd := downCmd()
		rootCmd.AddCommand(downCmd)
		downCmd.Flags().StringP(flagProjectName, "", defaultProjectName, "Project name for the local development server")
	}

	{
		initCmd := initCmd()
		rootCmd.AddCommand(initCmd)
		initCmd.Flags().Bool(
			flagRemote, false, "Validate remote configuration. Defaults to validation of local config.",
		)
	}

	{
		loginCmd := logincCmd()
		rootCmd.AddCommand(loginCmd)
		loginCmd.PersistentFlags().StringP(flagEmail, "e", "", "Email of your Nhost account")
		loginCmd.PersistentFlags().StringP(flagPassword, "p", "", "Password of your Nhost account")
	}

	{
		logoutCmd := logoutCmd()
		rootCmd.AddCommand(logoutCmd)
	}

	{
		linkCmd := linkCmd()
		rootCmd.AddCommand(linkCmd)
	}

	{
		listCmd := listCmd()
		rootCmd.AddCommand(listCmd)
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
