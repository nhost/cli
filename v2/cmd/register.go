package cmd

import "github.com/spf13/cobra"

func Register(rootCmd *cobra.Command) {
	loginCmd := logincCmd()
	rootCmd.AddCommand(loginCmd)
	loginCmd.PersistentFlags().StringP(flagEmail, "e", "", "Email of your Nhost account")
	loginCmd.PersistentFlags().StringP(flagPassword, "p", "", "Password of your Nhost account")

	logoutCmd := logoutCmd()
	rootCmd.AddCommand(logoutCmd)

	listCmd := listCmd()
	rootCmd.AddCommand(listCmd)
}
