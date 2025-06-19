// Package rm provides the rm command implementation
package rm

import (
	"github.com/codcod/repos/cmd/repos/common"
	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/git"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Config contains all configuration for rm command
type Config struct {
	Tag      string
	Parallel bool
}

// NewCommand creates the rm command
func NewCommand() *cobra.Command {
	rmConfig := &Config{}

	cmd := &cobra.Command{
		Use:   "rm",
		Short: "Remove repositories",
		Long:  `Remove repositories from the local filesystem. Filter by tag if specified.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Flags().GetString("config")

			cfg, err := common.LoadConfig(configFile)
			if err != nil {
				return err
			}

			repositories := cfg.FilterRepositoriesByTag(rmConfig.Tag)
			if len(repositories) == 0 {
				color.Yellow("No repositories found with tag: %s", rmConfig.Tag)
				return nil
			}

			color.Green("Removing %d repositories...", len(repositories))

			err = common.ProcessRepos(repositories, rmConfig.Parallel, func(r config.Repository) error {
				return git.RemoveRepository(r)
			})

			if err != nil {
				return err
			}

			color.Green("Done removing repositories")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&rmConfig.Tag, "tag", "t", "", "Filter repositories by tag")
	cmd.Flags().BoolVarP(&rmConfig.Parallel, "parallel", "p", false, "Remove repositories in parallel")

	return cmd
}
