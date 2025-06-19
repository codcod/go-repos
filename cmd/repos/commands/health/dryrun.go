// Package health provides health command subcommands
package health

import (
	"github.com/codcod/repos/cmd/repos/common"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// DryRunConfig contains configuration for dry run
type DryRunConfig struct {
	Tag string
}

// NewDryRunCommand creates the dryrun subcommand
func NewDryRunCommand() *cobra.Command {
	dryRunConfig := &DryRunConfig{}

	cmd := &cobra.Command{
		Use:   "dryrun",
		Short: "Perform a dry run of health checks",
		Long:  `Show what health checks would be performed without actually running them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Flags().GetString("config")

			cfg, err := common.LoadConfig(configFile)
			if err != nil {
				return err
			}

			repositories := cfg.FilterRepositoriesByTag(dryRunConfig.Tag)
			if len(repositories) == 0 {
				color.Yellow("No repositories found with tag: %s", dryRunConfig.Tag)
				return nil
			}

			color.Green("Dry run: would check health of %d repositories", len(repositories))

			for _, repo := range repositories {
				color.Cyan("  - %s (%s)", repo.Name, repo.URL)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&dryRunConfig.Tag, "tag", "t", "", "Filter repositories by tag")

	return cmd
}
