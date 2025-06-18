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
	healthGenConfig      bool
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
	healthCmd.Flags().BoolVar(&healthGenConfig, "gen-config", false, "Generate a comprehensive configuration template with all available options")

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
  repos health --gen-config             # Generate comprehensive configuration template
  repos health --dry-run                # Preview what would be executed`,
	Run: func(_ *cobra.Command, _ []string) {
		// Handle list-categories option first
		if healthListCategories {
			listHealthCategories()
			return
		}

		// Handle gen-config option
		if healthGenConfig {
			generateHealthConfig()
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
			advConfig = advConfig.FilterByCategories(healthCategories)
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
			showDryRunDetails(coreRepos, advConfig, analyzerReg, healthCategories)
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

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
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

// generateHealthConfig generates a comprehensive configuration template with all available options
//
//nolint:gocyclo
func generateHealthConfig() {
	// Create simple logger for initialization
	logger := &simpleLogger{}

	// Create command executor and registries to get available checkers and analyzers
	executor := health.NewCommandExecutor(30 * time.Second)
	checkerRegistry := health.NewCheckerRegistry(executor)

	// Create filesystem and analyzer registry
	fs := health.NewFileSystem()
	analyzerReg := health.NewAnalyzerRegistry(fs, logger)

	fmt.Println("# Comprehensive Health Configuration Template")
	fmt.Println("# This file demonstrates all available configuration options")
	fmt.Println("# Customize as needed for your project requirements")
	fmt.Println()
	fmt.Println("version: \"1.0\"")
	fmt.Println()

	// Engine configuration
	fmt.Println("# Engine configuration for parallel execution and performance")
	fmt.Println("engine:")
	fmt.Println("  max_concurrency: 4        # Maximum parallel checkers (default: 4)")
	fmt.Println("  timeout: 5m                # Global timeout for all checks")
	fmt.Println("  cache_enabled: true        # Enable result caching")
	fmt.Println("  cache_ttl: 1h             # Cache time-to-live")
	fmt.Println()

	// Checkers configuration
	fmt.Println("# Checker configurations - all available checkers with their options")
	fmt.Println("checkers:")

	checkers := checkerRegistry.GetCheckers()
	checkersByCategory := make(map[string][]core.Checker)

	// Group checkers by category
	for _, checker := range checkers {
		category := checker.Category()
		checkersByCategory[category] = append(checkersByCategory[category], checker)
	}

	// Generate configuration for each category
	for category, categoryCheckers := range checkersByCategory {
		fmt.Printf("  # %s category checkers\n", capitalizeFirst(category))

		for _, checker := range categoryCheckers {
			config := checker.Config()
			fmt.Printf("  %s:\n", checker.ID())
			fmt.Printf("    enabled: %t             # Enable/disable this checker\n", config.Enabled)
			fmt.Printf("    severity: %s           # Severity level: low, medium, high, critical\n", config.Severity)
			fmt.Printf("    timeout: %s            # Timeout for this specific checker\n", config.Timeout)
			fmt.Printf("    categories: [\"%s\"]      # Category classification\n", category)

			// Add common options based on checker type
			fmt.Println("    options:")
			switch checker.ID() {
			case "git-status":
				fmt.Println("      check_uncommitted: true    # Check for uncommitted changes")
				fmt.Println("      check_untracked: true      # Check for untracked files")
				fmt.Println("      check_unpushed: true       # Check for unpushed commits")

			case "git-last-commit":
				fmt.Println("      max_days_since_commit: 30  # Alert if last commit is older than N days")
				fmt.Println("      check_commit_messages: true # Validate commit message format")

			case "dependencies-outdated":
				fmt.Println("      package_managers: [\"npm\", \"pip\", \"go\", \"maven\"] # Supported package managers")
				fmt.Println("      severity_threshold: \"minor\" # Minimum severity to report: patch, minor, major")
				fmt.Println("      max_age_days: 180          # Consider packages outdated after N days")

			case "vulnerability-scan":
				fmt.Println("      scan_dependencies: true    # Scan dependencies for vulnerabilities")
				fmt.Println("      scan_code: false          # Scan source code (requires additional tools)")
				fmt.Println("      severity_threshold: \"medium\" # Minimum severity to report")

			case "branch-protection":
				fmt.Println("      require_reviews: true      # Require pull request reviews")
				fmt.Println("      require_status_checks: true # Require status checks to pass")
				fmt.Println("      enforce_admins: false      # Enforce restrictions for admins")

			case "license-check":
				fmt.Println("      allowed_licenses:          # List of allowed licenses")
				fmt.Println("        - \"MIT\"")
				fmt.Println("        - \"Apache-2.0\"")
				fmt.Println("        - \"BSD-3-Clause\"")
				fmt.Println("      check_compatibility: true  # Check license compatibility")

			case "ci-config":
				fmt.Println("      platforms: [\"github\", \"gitlab\", \"jenkins\"] # CI platforms to check")
				fmt.Println("      require_tests: true        # Require test execution in CI")
				fmt.Println("      require_linting: false     # Require linting in CI")

			default:
				fmt.Println("      # Checker-specific options would be documented here")
			}

			fmt.Printf("    exclusions:              # Files/patterns to exclude\n")
			fmt.Printf("      - \"test/\"\n")
			fmt.Printf("      - \"*.tmp\"\n")
			fmt.Println()
		}
	}

	// Analyzers configuration
	fmt.Println("# Language analyzer configurations")
	fmt.Println("analyzers:")

	analyzers := analyzerReg.GetAnalyzers()
	for _, analyzer := range analyzers {
		language := analyzer.Language()
		fmt.Printf("  %s:\n", language)
		fmt.Printf("    enabled: true              # Enable analyzer for %s\n", language)
		fmt.Printf("    file_extensions: %v # Supported file extensions\n", analyzer.SupportedExtensions())

		// Add language-specific exclude patterns
		switch language {
		case "go":
			fmt.Println("    exclude_patterns: [\"vendor\", \"_test.go\", \"*.pb.go\"]")
		case "python":
			fmt.Println("    exclude_patterns: [\"__pycache__\", \"*.pyc\", \".venv\", \"venv\"]")
		case "javascript":
			fmt.Println("    exclude_patterns: [\"node_modules\", \"dist\", \"build\", \"*.min.js\"]")
		case "java":
			fmt.Println("    exclude_patterns: [\"target\", \"*.class\", \"*.jar\"]")
		default:
			fmt.Println("    exclude_patterns: [\"build\", \"dist\", \"target\"]")
		}

		fmt.Println("    complexity_enabled: true   # Enable complexity analysis")
		fmt.Println("    function_level: true       # Analyze at function level")
		fmt.Println("    categories: [\"quality\", \"analysis\"]")
		fmt.Println()
	}

	// Reporters configuration
	fmt.Println("# Reporter configurations for output formatting")
	fmt.Println("reporters:")
	fmt.Println("  console:")
	fmt.Println("    enabled: true              # Console output")
	fmt.Println("    template: table            # Output format: table, list, summary")
	fmt.Println("    options:")
	fmt.Println("      show_summary: true       # Show summary statistics")
	fmt.Println("      show_details: true       # Show detailed results")
	fmt.Println("      color_output: true       # Use colored output")
	fmt.Println()
	fmt.Println("  json:")
	fmt.Println("    enabled: false             # JSON file output")
	fmt.Println("    output_file: \"health-report.json\"")
	fmt.Println("    template: detailed         # JSON format: simple, detailed, structured")
	fmt.Println("    options:")
	fmt.Println("      pretty_print: true       # Format JSON for readability")
	fmt.Println()
	fmt.Println("  html:")
	fmt.Println("    enabled: false             # HTML report output")
	fmt.Println("    output_file: \"health-report.html\"")
	fmt.Println("    template: dashboard        # HTML format: simple, dashboard, detailed")
	fmt.Println("    options:")
	fmt.Println("      include_charts: true     # Include visual charts")
	fmt.Println("      theme: \"light\"           # Theme: light, dark")
	fmt.Println()

	// Categories configuration
	fmt.Println("# Category configurations for organizing checks")
	fmt.Println("categories:")

	categories := make(map[string]bool)
	for category := range checkersByCategory {
		categories[category] = true
	}

	for category := range categories {
		fmt.Printf("  %s:\n", category)
		fmt.Printf("    name: \"%s Checks\"\n", capitalizeFirst(category))

		var description string
		switch category {
		case "git":
			description = "Git repository status and history checks"
		case "security":
			description = "Security vulnerability and configuration checks"
		case "dependencies":
			description = "Dependency management and freshness checks"
		case "compliance":
			description = "License and regulatory compliance checks"
		case "ci":
			description = "Continuous integration and deployment checks"
		default:
			description = fmt.Sprintf("%s related checks", capitalizeFirst(category))
		}

		fmt.Printf("    description: \"%s\"\n", description)
		fmt.Printf("    enabled: true              # Enable all checkers in this category\n")

		var severity string
		switch category {
		case "security":
			severity = "critical"
		case "dependencies":
			severity = "high"
		default:
			severity = "medium"
		}

		fmt.Printf("    severity: %s              # Default severity for category\n", severity)
		fmt.Println()
	}

	// Override configurations
	fmt.Println("# Override configurations for specific conditions")
	fmt.Println("# overrides:")
	fmt.Println("#   - name: \"legacy-repositories\"")
	fmt.Println("#     description: \"Special configuration for legacy repositories\"")
	fmt.Println("#     conditions:")
	fmt.Println("#       - type: \"tag\"")
	fmt.Println("#         field: \"tags\"")
	fmt.Println("#         operator: \"contains\"")
	fmt.Println("#         value: \"legacy\"")
	fmt.Println("#     checkers:")
	fmt.Println("#       security-vulnerabilities:")
	fmt.Println("#         enabled: false          # Disable for legacy repos")
	fmt.Println("#     engine:")
	fmt.Println("#       max_concurrency: 1       # Run sequentially for legacy repos")
	fmt.Println()

	fmt.Println("# Usage Instructions:")
	fmt.Println("# 1. Save this output to a file (e.g., health-config.yaml)")
	fmt.Println("# 2. Customize the options according to your project needs")
	fmt.Println("# 3. Use with: repos health --config health-config.yaml")
	fmt.Println("# 4. Test with: repos health --config health-config.yaml --dry-run")
}

// showDryRunDetails displays comprehensive dry-run information based on actual configuration
//
//nolint:gocyclo
func showDryRunDetails(repos []core.Repository, advConfig *healthconfig.AdvancedConfig, analyzerReg *health.AnalyzerRegistry, categories []string) {
	fmt.Println()
	color.Yellow("=== DRY RUN MODE - HEALTH CHECK EXECUTION PLAN ===")
	fmt.Println()

	// Repository information
	color.Cyan("📁 REPOSITORIES TO ANALYZE:")
	for i, repo := range repos {
		fmt.Printf("  %d. %s", i+1, repo.Name)
		if repo.Language != "" {
			fmt.Printf(" (Language: %s)", repo.Language)
		}
		if len(repo.Tags) > 0 {
			fmt.Printf(" [Tags: %v]", repo.Tags)
		}
		fmt.Printf("\n     Path: %s\n", repo.Path)
	}
	fmt.Printf("  Total repositories: %d\n", len(repos))
	fmt.Println()

	// Get checkers from the actual configuration (after filtering)
	allAnalyzers := analyzerReg.GetAnalyzers()

	// Show category filtering if applied
	if len(categories) > 0 {
		color.Blue("🔍 CATEGORY FILTERING APPLIED: %v", categories)
		fmt.Println()
	}

	// Show checkers that would be executed (from actual configuration)
	color.Cyan("🔧 CHECKERS TO EXECUTE:")
	if len(advConfig.Checkers) == 0 {
		color.Red("  No checkers configured")
	} else {
		// Group checkers by category from configuration
		checkersByCategory := make(map[string][]string)
		enabledCount := 0
		disabledCount := 0

		for checkerID, checkerConfig := range advConfig.Checkers {
			// Get category from checker categories
			category := "uncategorized"
			if len(checkerConfig.Categories) > 0 {
				category = checkerConfig.Categories[0] // Use first category
			}
			checkersByCategory[category] = append(checkersByCategory[category], checkerID)

			if checkerConfig.Enabled {
				enabledCount++
			} else {
				disabledCount++
			}
		}

		for category, checkerIDs := range checkersByCategory {
			fmt.Printf("  Category: %s\n", capitalizeFirst(category))
			for _, checkerID := range checkerIDs {
				checkerConfig := advConfig.Checkers[checkerID]
				status := "enabled"
				if !checkerConfig.Enabled {
					status = "disabled"
					color.Red("    ❌ %s - %s [%s] - DISABLED",
						checkerID, status, checkerConfig.Severity)
				} else {
					color.Green("    ✓ %s - %s [%s]",
						checkerID, status, checkerConfig.Severity)
				}
				if checkerConfig.Timeout > 0 {
					fmt.Printf(" (timeout: %s)", checkerConfig.Timeout)
				}
				fmt.Println()
			}
			fmt.Println()
		}
	}

	// Show analyzers that would be executed
	color.Cyan("🔬 ANALYZERS TO EXECUTE:")
	if len(allAnalyzers) == 0 {
		color.Red("  No analyzers available")
	} else {
		// Group analyzers by language and show which repositories they'd analyze
		analyzersByLang := make(map[string][]core.Analyzer)
		for _, analyzer := range allAnalyzers {
			lang := analyzer.Language()
			analyzersByLang[lang] = append(analyzersByLang[lang], analyzer)
		}

		for language, langAnalyzers := range analyzersByLang {
			fmt.Printf("  Language: %s\n", capitalizeFirst(language))

			// Count repositories that match this language
			repoCount := 0
			var matchingRepos []string
			for _, repo := range repos {
				if repo.Language == language {
					repoCount++
					matchingRepos = append(matchingRepos, repo.Name)
				}
			}

			for _, analyzer := range langAnalyzers {
				color.Green("    ✓ %s", analyzer.Name())
				fmt.Printf(" (Extensions: %v)", analyzer.SupportedExtensions())
				if repoCount > 0 {
					fmt.Printf(" - Would analyze %d repositories: %v", repoCount, matchingRepos)
				} else {
					color.Yellow(" - No matching repositories")
				}
				fmt.Println()
			}
			fmt.Println()
		}
	}

	// Configuration summary
	color.Cyan("⚙️  CONFIGURATION SUMMARY:")
	if advConfig != nil {
		fmt.Printf("  Engine max concurrency: %d\n", advConfig.Engine.MaxConcurrency)
		if advConfig.Engine.Timeout > 0 {
			fmt.Printf("  Engine timeout: %s\n", advConfig.Engine.Timeout)
		}
		fmt.Printf("  Cache enabled: %t\n", advConfig.Engine.CacheEnabled)
		if advConfig.Engine.CacheTTL > 0 {
			fmt.Printf("  Cache TTL: %s\n", advConfig.Engine.CacheTTL)
		}
	}
	fmt.Println()

	// Execution summary
	color.Cyan("📊 EXECUTION SUMMARY:")
	enabledCheckers := 0
	totalCheckers := 0
	for _, checkerConfig := range advConfig.Checkers {
		totalCheckers++
		if checkerConfig.Enabled {
			enabledCheckers++
		}
	}

	fmt.Printf("  Total repositories: %d\n", len(repos))
	fmt.Printf("  Total checkers: %d (enabled: %d, disabled: %d)\n",
		totalCheckers, enabledCheckers, totalCheckers-enabledCheckers)
	fmt.Printf("  Total analyzers: %d\n", len(allAnalyzers))
	if len(categories) > 0 {
		fmt.Printf("  Category filter: %v\n", categories)
	}
	fmt.Println()

	color.Yellow("=== This was a DRY RUN - no actual checks were performed ===")
	color.Blue("To execute the checks, run the same command without --dry-run")
	fmt.Println()
}
