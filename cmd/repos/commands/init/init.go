// Package init provides the init command implementation
package init

import (
	"os"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/util"
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

			// Get current directory
			currentDir, err := os.Getwd()
			if err != nil {
				color.Red("Error getting current directory: %v", err)
				return err
			}

			// Find Git repositories
			color.Green("Scanning for Git repositories in %s...", currentDir)
			repos, err := util.FindGitRepositories(currentDir)
			if err != nil {
				color.Red("Error scanning for repositories: %v", err)
				return err
			}

			if len(repos) == 0 {
				color.Yellow("No Git repositories found in %s", currentDir)
				// Create sample configuration instead
				repos = []config.Repository{
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
				}
				color.Yellow("Creating sample configuration with example repositories")
			} else {
				color.Green("Found %d Git repositories", len(repos))
			}

			// Create configuration
			cfg := &config.Config{
				Repositories: repos,
			}

			// Convert to YAML
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return err
			}
			// Write to file
			err = os.WriteFile(initConfig.OutputFile, data, 0600)
			if err != nil {
				return err
			}

			color.Green("Configuration file created: %s", initConfig.OutputFile)
			if len(repos) == 0 {
				color.Cyan("Edit the file to add your repositories.")
			} else {
				color.Green("Successfully created %s with %d repositories", initConfig.OutputFile, len(repos))
			}
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&initConfig.OutputFile, "output", "o", "config.yaml", "Output configuration file")
	cmd.Flags().BoolVar(&initConfig.Overwrite, "overwrite", false, "Overwrite existing configuration file")

	return cmd
}
