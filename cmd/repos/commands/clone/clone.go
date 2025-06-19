// Package clone provides the clone command implementation
package clone

import (
	"github.com/codcod/repos/cmd/repos/common"
	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/git"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Config contains all configuration for clone command
type Config struct {
	Tag      string
	Parallel bool
}

// NewCommand creates the clone command
func NewCommand() *cobra.Command {
	cloneConfig := &Config{}

	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone repositories specified in config",
		Long:  `Clone all repositories listed in the config file. Filter by tag if specified.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Flags().GetString("config")

			cfg, err := common.LoadConfig(configFile)
			if err != nil {
				return err
			}

			repositories := cfg.FilterRepositoriesByTag(cloneConfig.Tag)
			if len(repositories) == 0 {
				color.Yellow("No repositories found with tag: %s", cloneConfig.Tag)
				return nil
			}

			color.Green("Cloning %d repositories...", len(repositories))

			err = common.ProcessRepos(repositories, cloneConfig.Parallel, func(r config.Repository) error {
				err := git.CloneRepository(r)
				// Only show "Successfully cloned" if no error AND repository didn't already exist
				if err != nil {
					return err
				}
				// git.CloneRepository returns nil when repo exists (skipping clone) without showing success message
				// We don't need to output additional success message here
				return nil
			})

			if err != nil {
				return err
			}

			color.Green("Done cloning repositories")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&cloneConfig.Tag, "tag", "t", "", "Filter repositories by tag")
	cmd.Flags().BoolVarP(&cloneConfig.Parallel, "parallel", "p", false, "Clone repositories in parallel")

	return cmd
}
