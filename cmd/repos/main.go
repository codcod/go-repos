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
	"github.com/codcod/repos/internal/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	configFile  string
	tag         string
	parallel    bool
	logDir      string
	defaultLogs = "logs"

	// Version information
	version = "0.2.1"
	commit  = "f51155d"
	date    = "2025-06-08"

	// PR command flags
	prTitle    string
	prBody     string
	prBranch   string
	baseBranch string
	commitMsg  string
	prDraft    bool
	prToken    string
	createOnly bool

	// Init command flags
	maxDepth   int
	outputFile string
	overwrite  bool
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
			err := git.CloneRepository(r)
			// Only show "Successfully cloned" if no error AND repository didn't already exist
			if err != nil {
				return err
			}
			// git.CloneRepository returns nil when repo exists (skipping clone) without showing success message
			// We don't need to output additional success message here
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
			return runner.RunCommand(r, command, absLogDir)
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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a config.yaml file from discovered Git repositories",
	Long:  `Scan the current directory for Git repositories and generate a config.yaml file based on discovered repositories.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current directory
		currentDir, err := os.Getwd()
		if err != nil {
			color.Red("Error getting current directory: %v", err)
			os.Exit(1)
		}

		// Check if output file already exists
		if _, err := os.Stat(outputFile); err == nil && !overwrite {
			color.Red("File %s already exists. Use --overwrite to replace it.", outputFile)
			os.Exit(1)
		}

		// Find Git repositories
		color.Green("Scanning for Git repositories in %s (max depth: %d)...", currentDir, maxDepth)
		repos, err := util.FindGitRepositories(currentDir, maxDepth)
		if err != nil {
			color.Red("Error scanning for repositories: %v", err)
			os.Exit(1)
		}

		if len(repos) == 0 {
			color.Yellow("No Git repositories found in %s", currentDir)
			os.Exit(0)
		}

		color.Green("Found %d Git repositories", len(repos))

		// Create config structure
		cfg := config.Config{
			Repositories: repos,
		}

		// Convert to YAML
		yamlData, err := yaml.Marshal(cfg)
		if err != nil {
			color.Red("Error creating YAML: %v", err)
			os.Exit(1)
		}

		// Write to file
		err = os.WriteFile(outputFile, yamlData, 0644)
		if err != nil {
			color.Red("Error writing to file %s: %v", outputFile, err)
			os.Exit(1)
		}

		color.Green("Successfully created %s with %d repositories", outputFile, len(repos))

		// Print preview of the generated file
		fmt.Println("\nConfig file preview:")
		color.Cyan("---")
		fmt.Println(string(yamlData))
	},
}

// Process repositories with clean error handling
func processRepos(repositories []config.Repository, parallel bool, processor func(config.Repository) error) {
	logger := util.NewLogger()

	if parallel {
		var wg sync.WaitGroup
		wg.Add(len(repositories))

		for _, repo := range repositories {
			go func(r config.Repository) {
				defer wg.Done()
				if err := processor(r); err != nil {
					logger.Error(r, "%v", err)
				}
			}(repo)
		}

		wg.Wait()
	} else {
		for _, repo := range repositories {
			if err := processor(repo); err != nil {
				logger.Error(repo, "%v", err)
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

	// Init command flags
	initCmd.Flags().IntVar(&maxDepth, "depth", 3, "Maximum directory depth to scan for repositories")
	initCmd.Flags().StringVarP(&outputFile, "output", "o", "config.yaml", "Output file name")
	initCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing file if it exists")

	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(prCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(initCmd) // Add the init command

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("repos %s (%s) built on %s\n", version, commit, date)
		},
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
