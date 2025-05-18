package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/git"
	"github.com/codcod/repos/internal/github"
	"github.com/codcod/repos/internal/runner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	configFile  string
	tag         string
	parallel    bool
	logDir      string
	defaultLogs = "logs"

	// PR command flags
	prTitle    string
	prBody     string
	prBranch   string
	baseBranch string
	commitMsg  string
	prDraft    bool
	prToken    string
	createOnly bool
)

var rootCmd = &cobra.Command{
	Use:   "repos",
	Short: "A tool to manage multiple GitHub repositories",
	Long:  `Clone multiple GitHub repositories and run arbitrary commands inside them.`,
}

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone repositories specified in config",
	Long:  `Clone all repositories listed in the config file. Filter by tag if specified.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		repositories := cfg.FilterRepositoriesByTag(tag)
		if len(repositories) == 0 {
			color.Yellow("No repositories found with tag: %s", tag)
			return
		}

		color.Green("Cloning %d repositories...", len(repositories))

		processRepos(repositories, parallel, func(r config.Repository) error {
			if err := git.CloneRepository(r); err != nil {
				return err
			}
			color.Green("%s | Successfully cloned", color.New(color.FgCyan, color.Bold).SprintFunc()(r.Name))
			return nil
		})

		color.Green("Done cloning repositories")
	},
}

var runCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "Run a command in each repository",
	Long:  `Execute an arbitrary command in each repository. Filter by tag if specified.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		command := args[0]
		if len(args) > 1 {
			command = args[0] + " " + args[1]
			for _, arg := range args[2:] {
				command += " " + arg
			}
		}

		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		repositories := cfg.FilterRepositoriesByTag(tag)
		if len(repositories) == 0 {
			color.Yellow("No repositories found with tag: %s", tag)
			return
		}

		color.Green("Running '%s' in %d repositories...", command, len(repositories))

		// Create log directory if specified
		if logDir == "" {
			logDir = defaultLogs
		}

		// Absolute path for logs
		absLogDir, err := filepath.Abs(logDir)
		if err != nil {
			color.Red("Error resolving log directory path: %v", err)
			os.Exit(1)
		}

		processRepos(repositories, parallel, func(r config.Repository) error {
			if err := runner.RunCommand(r, command, absLogDir); err != nil {
				return err
			}
			color.Green("%s | Command completed successfully", color.New(color.FgCyan, color.Bold).SprintFunc()(r.Name))
			return nil
		})

		color.Green("Done running commands in all repositories")
	},
}

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create pull requests for repositories with changes",
	Long:  `Check for changes in repositories and create pull requests to GitHub.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		repositories := cfg.FilterRepositoriesByTag(tag)
		if len(repositories) == 0 {
			color.Yellow("No repositories found with tag: %s", tag)
			return
		}

		color.Green("Checking %d repositories for changes...", len(repositories))

		// Use environment variable if token not provided via flag
		if prToken == "" {
			prToken = os.Getenv("GITHUB_TOKEN")
			if prToken == "" && !createOnly {
				color.Red("GitHub token not provided. Use --token flag or set GITHUB_TOKEN environment variable.")
				os.Exit(1)
			}
		}

		// Configure PR options
		prOptions := github.PROptions{
			Title:      prTitle,
			Body:       prBody,
			BranchName: prBranch,
			BaseBranch: baseBranch,
			CommitMsg:  commitMsg,
			Draft:      prDraft,
			Token:      prToken,
			CreateOnly: createOnly,
		}

		successCount := 0

		processRepos(repositories, parallel, func(r config.Repository) error {
			if err := github.CreatePullRequest(r, prOptions); err != nil {
				if strings.Contains(err.Error(), "no changes detected") {
					color.Yellow("%s | No changes detected", color.New(color.FgCyan, color.Bold).SprintFunc()(r.Name))
				} else {
					return err
				}
			} else {
				color.Green("%s | Pull request created successfully", color.New(color.FgCyan, color.Bold).SprintFunc()(r.Name))
				successCount++
			}
			return nil
		})

		color.Green("Created %d pull requests", successCount)
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove cloned repositories",
	Long:  `Remove repositories that were previously cloned. Filter by tag if specified.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		repositories := cfg.FilterRepositoriesByTag(tag)
		if len(repositories) == 0 {
			color.Yellow("No repositories found with tag: %s", tag)
			return
		}

		color.Green("Removing %d repositories...", len(repositories))

		processRepos(repositories, parallel, func(r config.Repository) error {
			if err := git.RemoveRepository(r); err != nil {
				return err
			}
			color.Green("%s | Successfully removed", color.New(color.FgCyan, color.Bold).SprintFunc()(r.Name))
			return nil
		})

		color.Green("Done removing repositories")
	},
}

// Process repositories with clean error handling
func processRepos(repositories []config.Repository, parallel bool, processor func(config.Repository) error) {
	repoColor := color.New(color.FgCyan, color.Bold).SprintFunc()

	if parallel {
		var wg sync.WaitGroup
		wg.Add(len(repositories))

		for _, repo := range repositories {
			go func(r config.Repository) {
				defer wg.Done()
				if err := processor(r); err != nil {
					color.Red("%s | Error: %v", repoColor(r.Name), err)
				}
			}(repo)
		}

		wg.Wait()
	} else {
		for _, repo := range repositories {
			if err := processor(repo); err != nil {
				color.Red("%s | Error: %v", repoColor(repo.Name), err)
			}
		}
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "config file path")
	rootCmd.PersistentFlags().StringVarP(&tag, "tag", "t", "", "filter repositories by tag")
	rootCmd.PersistentFlags().BoolVarP(&parallel, "parallel", "p", false, "execute operations in parallel")

	runCmd.Flags().StringVarP(&logDir, "logs", "l", defaultLogs, "directory to store log files")

	// PR command flags
	prCmd.Flags().StringVar(&prTitle, "title", "Automated changes", "Title for the pull request")
	prCmd.Flags().StringVar(&prBody, "body", "This PR was created automatically", "Body text for the pull request")
	prCmd.Flags().StringVar(&prBranch, "branch", "", "Branch name to create (default: automated-changes-{PID})")
	prCmd.Flags().StringVar(&baseBranch, "base", "", "Base branch for the PR (default: main or master)")
	prCmd.Flags().StringVar(&commitMsg, "message", "", "Commit message (defaults to PR title)")
	prCmd.Flags().BoolVar(&prDraft, "draft", false, "Create PR as draft")
	prCmd.Flags().StringVar(&prToken, "token", "", "GitHub token (can also use GITHUB_TOKEN env var)")
	prCmd.Flags().BoolVar(&createOnly, "create-only", false, "Only create PR, don't commit changes")

	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(prCmd)
	rootCmd.AddCommand(rmCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
