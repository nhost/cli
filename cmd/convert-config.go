package cmd

import (
	"fmt"
	"github.com/nhost/be/services/mimir/schema"
	"github.com/nhost/cli/config/converter"
	"github.com/nhost/cli/nhost"
	"github.com/nhost/cli/util"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"

	"path"
)

var (
	legacyConfigPath string
	newConfigPath    string
	overwrite        bool
)

// envCmd represents the env command
var convertConfigCmd = &cobra.Command{
	Use:   "convert-config",
	Short: "Convert deprecated config to new format",
	RunE: func(cmd *cobra.Command, args []string) error {
		var legacyConfig nhost.Configuration
		legacyData, err := os.ReadFile(legacyConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read legacy config file: %w", err)
		}

		err = yaml.Unmarshal(legacyData, &legacyConfig)
		if err != nil {
			return fmt.Errorf("failed to unmarshal legacy config file: %w", err)
		}

		if util.PathExists(newConfigPath) && !overwrite {
			return fmt.Errorf("new config file already exists. Use --overwrite to overwrite existing file")
		}

		newConfig, err := converter.Convert(log, &legacyConfig)
		if err != nil {
			return fmt.Errorf("failed to convert config: %w", err)
		}

		tomlConfigData, err := toml.Marshal(newConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal new config: %w", err)
		}

		s, err := schema.New()
		if err != nil {
			return err
		}

		if err = s.ValidateConfig(newConfig); err != nil {
			log.Infof("Generated config:\n\n%s\n\n", string(tomlConfigData))
			return fmt.Errorf("validation failed for new config: %w", err)
		}

		if err = os.WriteFile(newConfigPath, tomlConfigData, 0644); err != nil {
			return fmt.Errorf("failed to write new config file: %w", err)
		}

		log.Info("Successfully converted config file")

		return nil
	},
}

func init() {
	nhost.Init()
	rootCmd.AddCommand(convertConfigCmd)

	convertConfigCmd.Flags().StringVar(&legacyConfigPath, "legacy-config-path", path.Join(nhost.NHOST_DIR, "config.yaml"), "Path to legacy config file")
	convertConfigCmd.Flags().StringVar(&newConfigPath, "new-config-path", nhost.CONFIG_PATH, "Path to new config file")
	convertConfigCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing new config file")
}
