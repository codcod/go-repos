// Package init provides the init command implementation
package init

import (
	"os"

	"github.com/codcod/repos/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"
)

// Config contains all configuration for init command
type Config struct {
	OutputFile string
	Overwrite  bool
}

// NewCommand creates the init command
func NewCommand() *cobra.Command {
	initConfig := &Config{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new configuration file",
		Long:  `Create a new configuration file with sample repository entries.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if file exists and we're not overwriting
			if _, err := os.Stat(initConfig.OutputFile); err == nil && !initConfig.Overwrite {
				color.Yellow("Configuration file already exists: %s", initConfig.OutputFile)
				color.Yellow("Use --overwrite to replace it.")
				return nil
			}

			// Create sample configuration
			sampleConfig := &config.Config{
				Repositories: []config.Repository{
					{
						Name: "example-repo",
						URL:  "https://github.com/user/example-repo.git",
						Tags: []string{"example"},
					},
					{
						Name: "another-repo",
						URL:  "https://github.com/user/another-repo.git",
						Tags: []string{"example", "demo"},
					},
				},
			}

			// Convert to YAML
			data, err := yaml.Marshal(sampleConfig)
			if err != nil {
				return err
			}
			// Write to file
			err = os.WriteFile(initConfig.OutputFile, data, 0600)
			if err != nil {
				return err
			}

			color.Green("Configuration file created: %s", initConfig.OutputFile)
			color.Cyan("Edit the file to add your repositories.")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&initConfig.OutputFile, "output", "o", "config.yaml", "Output configuration file")
	cmd.Flags().BoolVar(&initConfig.Overwrite, "overwrite", false, "Overwrite existing configuration file")

	return cmd
}
