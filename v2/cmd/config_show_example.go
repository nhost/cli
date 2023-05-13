package cmd

import (
	"github.com/spf13/cobra"
)

func configShowFullExampleCmd() *cobra.Command {
	return &cobra.Command{ //nolint:exhaustruct
		Use:  "show-full-example",
		Long: `Show full example configuration`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Println("Not implemented yet")
			// exampleConf, err := config.FullExampleConfig()
			// if err != nil {
			// 	return fmt.Errorf("failed to get full example config: %w", err)
			// }

			// cmd.Println(string(exampleConf))
			return nil
		},
	}
}
