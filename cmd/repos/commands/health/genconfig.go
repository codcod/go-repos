// Package health provides health command subcommands
package health

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"
)

// GenConfigConfig contains configuration for generating config
type GenConfigConfig struct {
	OutputFile string
	Overwrite  bool
}

// NewGenConfigCommand creates the genconfig subcommand
func NewGenConfigCommand() *cobra.Command {
	genConfigConfig := &GenConfigConfig{
		OutputFile: "health-config.yaml",
	}

	cmd := &cobra.Command{
		Use:   "genconfig",
		Short: "Generate a sample health configuration file",
		Long:  `Create a sample health configuration file with default settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if file exists and we're not overwriting
			if _, err := os.Stat(genConfigConfig.OutputFile); err == nil && !genConfigConfig.Overwrite {
				color.Yellow("Health configuration file already exists: %s", genConfigConfig.OutputFile)
				color.Yellow("Use --overwrite to replace it.")
				return nil
			}

			// Create sample health configuration
			sampleConfig := map[string]interface{}{
				"health": map[string]interface{}{
					"timeout": "30s",
					"checkers": map[string]interface{}{
						"git": map[string]interface{}{
							"enabled": true,
						},
						"build": map[string]interface{}{
							"enabled": true,
						},
						"test": map[string]interface{}{
							"enabled": true,
						},
					},
				},
			}

			// Convert to YAML
			data, err := yaml.Marshal(sampleConfig)
			if err != nil {
				return err
			}

			// Write to file
			err = os.WriteFile(genConfigConfig.OutputFile, data, 0600)
			if err != nil {
				return err
			}

			color.Green("Health configuration file created: %s", genConfigConfig.OutputFile)
			color.Cyan("Edit the file to customize health check settings.")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&genConfigConfig.OutputFile, "output", "o", "health-config.yaml", "Output health configuration file")
	cmd.Flags().BoolVar(&genConfigConfig.Overwrite, "overwrite", false, "Overwrite existing configuration file")

	return cmd
}
