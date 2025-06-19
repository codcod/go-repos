// Package run provides the run command implementation
package run

import (
	"github.com/codcod/repos/cmd/repos/common"
	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/runner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Config contains all configuration for run command
type Config struct {
	Tag      string
	Parallel bool
	LogDir   string
}

// NewCommand creates the run command
func NewCommand() *cobra.Command {
	runConfig := &Config{
		LogDir: "logs",
	}

	cmd := &cobra.Command{
		Use:   "run [command]",
		Short: "Run a command in each repository",
		Long:  `Execute an arbitrary command in each repository. Filter by tag if specified.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := args[0]
			if len(args) > 1 {
				command = args[0] + " " + args[1]
				for _, arg := range args[2:] {
					command += " " + arg
				}
			}

			configFile, _ := cmd.Flags().GetString("config")

			cfg, err := common.LoadConfig(configFile)
			if err != nil {
				return err
			}

			repositories := cfg.FilterRepositoriesByTag(runConfig.Tag)
			if len(repositories) == 0 {
				color.Yellow("No repositories found with tag: %s", runConfig.Tag)
				return nil
			}

			color.Green("Running command '%s' in %d repositories...", command, len(repositories))

			err = common.ProcessRepos(repositories, runConfig.Parallel, func(r config.Repository) error {
				return runner.RunCommand(r, command, runConfig.LogDir)
			})

			if err != nil {
				return err
			}

			color.Green("Done running command in repositories")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&runConfig.Tag, "tag", "t", "", "Filter repositories by tag")
	cmd.Flags().BoolVarP(&runConfig.Parallel, "parallel", "p", false, "Run command in repositories in parallel")
	cmd.Flags().StringVarP(&runConfig.LogDir, "log-dir", "l", "logs", "Directory to store command output logs")

	return cmd
}
