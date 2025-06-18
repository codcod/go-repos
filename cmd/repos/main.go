// cmd/repos/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/git"
	"github.com/codcod/repos/internal/github"
	"github.com/codcod/repos/internal/health"
	healthconfig "github.com/codcod/repos/internal/health/config"
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

	// Version information - will be set via build flags, with environment variable fallback
	version = "dev"
	commit  = "unknown"
	date    = "unknown"

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

	// Health command flags
	healthConfig         string
	healthCategories     []string
	healthParallel       bool
	healthTimeout        int
	healthDryRun         bool
	healthVerbose        bool
	healthListCategories bool
)

// getEnvOrDefault returns the environment variable value or default if empty
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// init function to handle environment variable fallback for version info
func init() {
	// Use environment variables as fallback when build-time flags weren't set
	if version == "dev" {
		version = getEnvOrDefault("VERSION", version)
	}
	if commit == "unknown" {
		commit = getEnvOrDefault("COMMIT", commit)
	}
	if date == "unknown" {
		date = getEnvOrDefault("BUILD_DATE", date)
	}
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

	// Health command flags
	healthCmd.Flags().StringVar(&healthConfig, "config", "", "health config file path (optional, uses built-in defaults if not provided)")
	healthCmd.Flags().StringSliceVar(&healthCategories, "category", []string{}, "filter checkers and analyzers by categories (comma-separated, e.g., 'git,security')")
	healthCmd.Flags().BoolVar(&healthParallel, "parallel", false, "Execute health checks in parallel")
	healthCmd.Flags().IntVar(&healthTimeout, "timeout", 30, "Timeout in seconds for health checks (default: 30)")
	healthCmd.Flags().BoolVar(&healthDryRun, "dry-run", false, "Dry run mode - show what would be executed")
	healthCmd.Flags().BoolVar(&healthVerbose, "verbose", false, "Enable verbose output for health checks")
	healthCmd.Flags().BoolVar(&healthListCategories, "list-categories", false, "List all available categories, checkers, and analyzers")

	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(prCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(initCmd)   // Add the init command
	rootCmd.AddCommand(healthCmd) // Add the health command

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

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Run comprehensive health checks with advanced analysis",
	Long: `Execute modular health checks using the health engine with advanced reporting.

The health command works out-of-the-box with sensible defaults. No configuration file is required.
If you want to customize the checks, you can provide an optional configuration file.

Examples:
  repos health                           # Run with built-in defaults
  repos health --config custom.yaml     # Use custom configuration
  repos health --category git,security  # Run only git and security checks
  repos health --verbose                # Show detailed output
  repos health --list-categories        # List all available categories and checks
  repos health --dry-run                # Preview what would be executed`,
	Run: func(_ *cobra.Command, _ []string) {
		// Handle list-categories option first
		if healthListCategories {
			listHealthCategories()
			return
		}

		// Create simple logger
		logger := &simpleLogger{}

		// If no config file is specified, use default name or empty for built-in defaults
		configPath := healthConfig
		if configPath == "" {
			configPath = "orchestration.yaml" // Try default file, will use built-in defaults if not found
		}

		// Load advanced configuration or use defaults if file doesn't exist
		advConfig, err := healthconfig.LoadAdvancedConfigOrDefault(configPath)
		if err != nil {
			color.Red("Error loading health config: %v", err)
			os.Exit(1)
		}

		// Load basic config for repositories
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

		// Convert repositories to core.Repository format
		coreRepos := make([]core.Repository, len(repositories))
		for i, repo := range repositories {
			// Use the actual repository path if it exists, otherwise use the specified path
			repoPath := repo.Path
			if repoPath == "" {
				repoPath = filepath.Join("cloned_repos", repo.Name)
			}

			// Detect language from repository tags or directory structure
			language := detectRepositoryLanguage(repo, repoPath)

			coreRepos[i] = core.Repository{
				Name:     repo.Name,
				Path:     repoPath,
				URL:      repo.URL,
				Branch:   repo.Branch,
				Tags:     repo.Tags,
				Language: language,
				Metadata: make(map[string]string),
			}
		}

		color.Green("Running comprehensive health checks on %d repositories...", len(repositories))

		// Apply category filtering if specified
		if len(healthCategories) > 0 {
			color.Blue("Filtering by categories: %v", healthCategories)
			advConfig.FilterByCategories(healthCategories)
		}

		// Create command executor and registries
		executor := health.NewCommandExecutor(time.Duration(healthTimeout) * time.Second)
		checkerRegistry := health.NewCheckerRegistry(executor)

		// Create filesystem and analyzer registry
		fs := health.NewFileSystem()
		analyzerReg := health.NewAnalyzerRegistry(fs, logger)

		// Create orchestration engine
		engine := health.NewOrchestrationEngine(checkerRegistry, analyzerReg, advConfig, logger)

		// Execute health checks
		if healthDryRun {
			color.Yellow("Dry run mode - would execute health checks on %d repositories", len(coreRepos))
			return
		}

		ctx := context.Background()
		if healthTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(healthTimeout)*time.Second)
			defer cancel()
		}

		result, err := engine.ExecuteHealthCheck(ctx, coreRepos)
		if err != nil {
			color.Red("Error executing code analysis: %v", err)
			os.Exit(1)
		}

		// Display results using the formatter
		formatter := health.NewFormatter(healthVerbose)
		formatter.DisplayResults(*result)

		// Exit with appropriate code based on results
		os.Exit(health.GetExitCode(*result))
	},
}

// simpleLogger provides a basic logger implementation
type simpleLogger struct{}

func (l *simpleLogger) Debug(msg string, fields ...core.Field) {
	if healthVerbose {
		fmt.Print("[DEBUG] " + msg + l.formatFieldsAsString(fields))
	}
}

func (l *simpleLogger) Info(msg string, fields ...core.Field) {
	fmt.Print("[INFO] " + msg + l.formatFieldsAsString(fields))
}

func (l *simpleLogger) Warn(msg string, fields ...core.Field) {
	color.Yellow("[WARN] " + msg + l.formatFieldsAsString(fields))
}

func (l *simpleLogger) Error(msg string, fields ...core.Field) {
	color.Red("[ERROR] " + msg + l.formatFieldsAsString(fields))
}

func (l *simpleLogger) Fatal(msg string, fields ...core.Field) {
	color.Red("[FATAL] " + msg + l.formatFieldsAsString(fields))
	os.Exit(1)
}

func (l *simpleLogger) formatFieldsAsString(fields []core.Field) string {
	if len(fields) == 0 {
		return "\n"
	}

	var result string
	for _, field := range fields {
		result += fmt.Sprintf(" [%s=%v]", field.Key, field.Value)
	}
	return result + "\n"
}

// detectRepositoryLanguage attempts to detect the primary language of a repository
//
//nolint:gocyclo
func detectRepositoryLanguage(repo config.Repository, repoPath string) string {
	// First, check tags for language hints
	for _, tag := range repo.Tags {
		switch tag {
		case "go", "golang":
			return "go"
		case "python", "py":
			return "python"
		case "javascript", "js", "node", "nodejs":
			return "javascript"
		case "java":
			return "java"
		case "rust":
			return "rust"
		case "cpp", "c++":
			return "cpp"
		case "c":
			return "c"
		}
	}

	// If no language tag found, try to detect from directory structure
	if _, err := os.Stat(repoPath); err == nil {
		// Check for language-specific files
		if hasFile(repoPath, "go.mod", "main.go", "*.go") {
			return "go"
		}
		if hasFile(repoPath, "requirements.txt", "setup.py", "pyproject.toml", "*.py") {
			return "python"
		}
		if hasFile(repoPath, "package.json", "*.js", "*.ts") {
			return "javascript"
		}
		if hasFile(repoPath, "pom.xml", "build.gradle", "*.java") {
			return "java"
		}
		if hasFile(repoPath, "Cargo.toml", "*.rs") {
			return "rust"
		}
	}

	return "" // Unknown language
}

// hasFile checks if any of the specified files exist in the repository path
func hasFile(repoPath string, patterns ...string) bool {
	for _, pattern := range patterns {
		if strings.Contains(pattern, "*") {
			// Use glob pattern
			matches, err := filepath.Glob(filepath.Join(repoPath, pattern))
			if err == nil && len(matches) > 0 {
				return true
			}
		} else {
			// Check for exact file
			if _, err := os.Stat(filepath.Join(repoPath, pattern)); err == nil {
				return true
			}
		}
	}
	return false
}

// listHealthCategories lists all available categories, checkers, and analyzers
func listHealthCategories() {
	logger := &simpleLogger{}

	// Create registries to discover available checkers and analyzers
	executor := health.NewCommandExecutor(30 * time.Second)
	checkerRegistry := health.NewCheckerRegistry(executor)

	fs := health.NewFileSystem()
	analyzerRegistry := health.NewAnalyzerRegistry(fs, logger)

	fmt.Println("=== Available Health Check Categories ===")
	fmt.Println()

	// List checkers by category
	checkers := checkerRegistry.GetCheckers()
	checkersByCategory := make(map[string][]core.Checker)

	for _, checker := range checkers {
		category := checker.Category()
		checkersByCategory[category] = append(checkersByCategory[category], checker)
	}

	fmt.Println("📋 CHECKERS:")
	for category, categoryCheckers := range checkersByCategory {
		fmt.Printf("  Category: %s\n", category)
		for _, checker := range categoryCheckers {
			config := checker.Config()
			status := "disabled"
			if config.Enabled {
				status = "enabled"
			}
			fmt.Printf("    • %s (%s) - %s [%s]\n",
				checker.Name(),
				checker.ID(),
				status,
				config.Severity)
		}
		fmt.Println()
	}

	// List analyzers by language
	analyzers := analyzerRegistry.GetAnalyzers()
	fmt.Println("🔍 ANALYZERS:")

	for _, analyzer := range analyzers {
		fmt.Printf("  Language: %s\n", analyzer.Language())
		fmt.Printf("    • Name: %s\n", analyzer.Name())
		fmt.Printf("    • Extensions: %v\n", analyzer.SupportedExtensions())
		fmt.Println()
	}

	// Summary
	fmt.Println("=== Summary ===")
	fmt.Printf("Total Checkers: %d\n", len(checkers))
	fmt.Printf("Total Categories: %d\n", len(checkersByCategory))
	fmt.Printf("Total Analyzers: %d\n", len(analyzers))

	fmt.Println("\nUsage Examples:")
	fmt.Println("  repos health --category git,security     # Run only git and security checkers")
	fmt.Println("  repos health --verbose                   # Show detailed output")
	fmt.Println("  repos health --dry-run                   # Preview what would be executed")
}
