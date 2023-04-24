package cmd

import (
	"fmt"

	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage your Nhost configuration",
}

var showFullExampleConfigCmd = &cobra.Command{
	Use:  "show-full-example",
	Long: `Show full example configuration`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exampleConf, err := config.FullExampleConfig()
		if err != nil {
			return fmt.Errorf("failed to get full example config: %v", err)
		}

		fmt.Println(string(exampleConf))
		return nil
	},
}

// func init() {
// 	rootCmd.AddCommand(configCmd)
// 	configCmd.AddCommand(pullConfigCmd)
// 	configCmd.AddCommand(validateConfigCmd)
// 	configCmd.AddCommand(showFullExampleConfigCmd)
// 	validateConfigCmd.Flags().Bool("local", false, "Validate local configuration")
// 	validateConfigCmd.Flags().Bool("remote", false, "Validate remote configuration")
// 	validateConfigCmd.MarkFlagsMutuallyExclusive("local", "remote")
// }

func anonymizeAppSecrets(secrets model.Secrets) model.Secrets {
	defaultSecretsMapping := map[string]string{}
	defaultSecrets := config.DefaultSecrets()
	for _, defaultSecret := range defaultSecrets {
		defaultSecretsMapping[defaultSecret.GetName()] = defaultSecret.GetValue()
	}

	anonymized := model.Secrets{}
	for _, v := range secrets {
		if defaultSecretValue, ok := defaultSecretsMapping[v.GetName()]; ok {
			anonymized = append(anonymized, &model.ConfigEnvironmentVariable{
				Name:  v.GetName(),
				Value: defaultSecretValue,
			})
			continue
		}

		anonymized = append(anonymized, &model.ConfigEnvironmentVariable{
			Name:  v.GetName(),
			Value: "FIXME",
		})
	}
	return anonymized
}
