// Package health provides health command subcommands
package health

import (
	"github.com/codcod/repos/cmd/repos/common"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// ComplexityConfig contains configuration for complexity analysis
type ComplexityConfig struct {
	Tag           string
	MaxComplexity int
}

// NewComplexityCommand creates the complexity subcommand
func NewComplexityCommand() *cobra.Command {
	complexityConfig := &ComplexityConfig{
		MaxComplexity: 10,
	}

	cmd := &cobra.Command{
		Use:   "complexity",
		Short: "Run cyclomatic complexity analysis",
		Long:  `Analyze the cyclomatic complexity of code in repositories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Flags().GetString("config")

			cfg, err := common.LoadConfig(configFile)
			if err != nil {
				return err
			}

			repositories := cfg.FilterRepositoriesByTag(complexityConfig.Tag)
			if len(repositories) == 0 {
				color.Yellow("No repositories found with tag: %s", complexityConfig.Tag)
				return nil
			}

			color.Green("Running cyclomatic complexity analysis on %d repositories...", len(repositories))

			// For now, this is a placeholder - the actual complexity analysis
			// would need to be implemented based on the internal health package structure
			color.Cyan("Complexity analysis feature needs to be implemented")
			color.Cyan("Maximum complexity threshold: %d", complexityConfig.MaxComplexity)

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&complexityConfig.Tag, "tag", "t", "", "Filter repositories by tag")
	cmd.Flags().IntVar(&complexityConfig.MaxComplexity, "max-complexity", 10, "Maximum allowed cyclomatic complexity")

	return cmd
}
