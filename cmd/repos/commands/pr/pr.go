// Package pr provides the pr command implementation
package pr

import (
	"github.com/codcod/repos/cmd/repos/common"
	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/github"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Config contains all configuration for pr command
type Config struct {
	Tag        string
	Parallel   bool
	Title      string
	Body       string
	Branch     string
	BaseBranch string
	CommitMsg  string
	Draft      bool
	Token      string
	CreateOnly bool
}

// NewCommand creates the pr command
func NewCommand() *cobra.Command {
	prConfig := &Config{}

	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Create pull requests in repositories",
		Long:  `Create pull requests in each repository. Filter by tag if specified.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Flags().GetString("config")

			cfg, err := common.LoadConfig(configFile)
			if err != nil {
				return err
			}

			repositories := cfg.FilterRepositoriesByTag(prConfig.Tag)
			if len(repositories) == 0 {
				color.Yellow("No repositories found with tag: %s", prConfig.Tag)
				return nil
			}

			color.Green("Creating pull requests in %d repositories...", len(repositories))

			err = common.ProcessRepos(repositories, prConfig.Parallel, func(r config.Repository) error {
				return github.CreatePullRequest(r, github.PROptions{
					Title:      prConfig.Title,
					Body:       prConfig.Body,
					BranchName: prConfig.Branch,
					BaseBranch: prConfig.BaseBranch,
					CommitMsg:  prConfig.CommitMsg,
					Draft:      prConfig.Draft,
					Token:      prConfig.Token,
					CreateOnly: prConfig.CreateOnly,
				})
			})

			if err != nil {
				return err
			}

			color.Green("Done creating pull requests")
			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&prConfig.Tag, "tag", "t", "", "Filter repositories by tag")
	cmd.Flags().BoolVarP(&prConfig.Parallel, "parallel", "p", false, "Create pull requests in parallel")
	cmd.Flags().StringVar(&prConfig.Title, "title", "", "Pull request title")
	cmd.Flags().StringVar(&prConfig.Body, "body", "", "Pull request body")
	cmd.Flags().StringVar(&prConfig.Branch, "branch", "", "Source branch name")
	cmd.Flags().StringVar(&prConfig.BaseBranch, "base", "main", "Base branch name")
	cmd.Flags().StringVar(&prConfig.CommitMsg, "commit", "", "Commit message")
	cmd.Flags().BoolVar(&prConfig.Draft, "draft", false, "Create as draft pull request")
	cmd.Flags().StringVar(&prConfig.Token, "token", "", "GitHub token")
	cmd.Flags().BoolVar(&prConfig.CreateOnly, "create-only", false, "Only create PR, don't commit changes")

	return cmd
}
