// cmd/repos/main.go
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
	yaml "gopkg.in/yaml.v3"
)

var (
	configFile  string
	tag         string
	parallel    bool
	logDir      string
	defaultLogs = "logs"

	// Version information - will be set via build flags or environment variables
	version = getEnvOrDefault("VERSION", "dev")
	commit  = getEnvOrDefault("COMMIT", "unknown")
	date    = getEnvOrDefault("BUILD_DATE", "unknown")

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
	outputFile string
	overwrite  bool
)

// getEnvOrDefault returns the value of the environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var rootCmd = &cobra.Command{
	Use:   "repos",
	Short: "A tool to manage multiple GitHub repositories",
	Long:  `Clone multiple GitHub repositories and run arbitrary commands inside them.`,
}

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone repositories specified in config",
	Long:  `Clone all repositories listed in the config file. Filter by tag if specified.`,
	Run: func(_ *cobra.Command, _ []string) {
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

		err = processRepos(repositories, parallel, func(r config.Repository) error {
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
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		color.Green("Done cloning repositories")
	},
}

var runCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "Run a command in each repository",
	Long:  `Execute an arbitrary command in each repository. Filter by tag if specified.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
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

		err = processRepos(repositories, parallel, func(r config.Repository) error {
			return runner.RunCommand(r, command, absLogDir)
		})

		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		color.Green("Done running commands in all repositories")
	},
}

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create pull requests for repositories with changes",
	Long:  `Check for changes in repositories and create pull requests to GitHub.`,
	Run: func(_ *cobra.Command, _ []string) {
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

		err = processRepos(repositories, parallel, func(r config.Repository) error {
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

		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		color.Green("Created %d pull requests", successCount)
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove cloned repositories",
	Long:  `Remove repositories that were previously cloned. Filter by tag if specified.`,
	Run: func(_ *cobra.Command, _ []string) {
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

		err = processRepos(repositories, parallel, func(r config.Repository) error {
			if err := git.RemoveRepository(r); err != nil {
				return err
			}
			color.Green("%s | Successfully removed", color.New(color.FgCyan, color.Bold).SprintFunc()(r.Name))
			return nil
		})

		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		color.Green("Done removing repositories")
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a config.yaml file from discovered Git repositories",
	Long:  `Scan the current directory for Git repositories and generate a config.yaml file based on discovered repositories.`,
	Run: func(_ *cobra.Command, _ []string) {
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
		color.Green("Scanning for Git repositories in %s...", currentDir)
		repos, err := util.FindGitRepositories(currentDir)
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
		err = os.WriteFile(outputFile, yamlData, 0600)
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
func processRepos(repositories []config.Repository, parallel bool, processor func(config.Repository) error) error {
	logger := util.NewLogger()
	var hasErrors bool

	if parallel {
		var wg sync.WaitGroup
		var mu sync.Mutex
		wg.Add(len(repositories))

		for _, repo := range repositories {
			go func(r config.Repository) {
				defer wg.Done()
				if err := processor(r); err != nil {
					logger.Error(r, "%v", err)
					mu.Lock()
					hasErrors = true
					mu.Unlock()
				}
			}(repo)
		}

		wg.Wait()
	} else {
		for _, repo := range repositories {
			if err := processor(repo); err != nil {
				logger.Error(repo, "%v", err)
				hasErrors = true
			}
		}
	}

	if hasErrors {
		return fmt.Errorf("one or more commands failed")
	}
	return nil
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
		Run: func(_ *cobra.Command, _ []string) {
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
