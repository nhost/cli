package cmd

import (
	"fmt"

	"github.com/nhost/cli/v2/nhostclient/credentials"
	"github.com/nhost/cli/v2/system"
	"github.com/spf13/cobra"
)

const (
	flagDomain = "domain"
)

func GetNhostCredentials() (credentials.Credentials, error) {
	f, err := system.GetStateAuthFile()
	if err != nil {
		return credentials.Credentials{}, fmt.Errorf("failed to get auth file: %w", err)
	}
	defer f.Close()

	creds, err := credentials.FromReader(f)
	if err != nil {
		return credentials.Credentials{}, fmt.Errorf("failed to get credentials: %w", err)
	}
	return creds, nil
}

func Register(rootCmd *cobra.Command) {
	configCmd := configCmd()
	rootCmd.AddCommand(configCmd)

	configPullCmd := configPullCmd()
	configCmd.AddCommand(configPullCmd)

	loginCmd := logincCmd()
	rootCmd.AddCommand(loginCmd)
	loginCmd.PersistentFlags().StringP(flagEmail, "e", "", "Email of your Nhost account")
	loginCmd.PersistentFlags().StringP(flagPassword, "p", "", "Password of your Nhost account")

	logoutCmd := logoutCmd()
	rootCmd.AddCommand(logoutCmd)

	linkCmd := linkCmd()
	rootCmd.AddCommand(linkCmd)

	listCmd := listCmd()
	rootCmd.AddCommand(listCmd)
}
