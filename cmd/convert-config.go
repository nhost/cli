/*
MIT License

# Copyright (c) Nhost

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package cmd

import (
	"fmt"
	"github.com/nhost/cli/config/converter"
	"github.com/nhost/cli/nhost"
	"github.com/nhost/cli/util"
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

		tomlConfigData, err := newConfig.Marshal()
		if err != nil {
			return fmt.Errorf("failed to marshal new config: %w", err)
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
