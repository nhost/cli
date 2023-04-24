package cmd

import (
	"github.com/nhost/be/services/mimir/model"
	"github.com/nhost/cli/config"
)

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
