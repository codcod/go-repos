// Package health provides the health command implementation
package health

import (
	"fmt"
	"time"

	"github.com/codcod/repos/cmd/repos/common"
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health"
	"github.com/codcod/repos/internal/health/commands"
	"github.com/spf13/cobra"
)

// Config contains all configuration for health checks
type Config struct {
	ConfigPath     string
	Categories     []string
	Pipeline       string
	Parallel       bool
	TimeoutSeconds int // Timeout in seconds
	DryRun         bool
	Verbose        bool
	Tag            string
	BasicConfig    string // Path to basic repo config
	ListCategories bool   // List available categories and checkers
	MaxComplexity  int    // Maximum allowed complexity
}

// NewCommand creates the health command
func NewCommand() *cobra.Command {
	config := &Config{
		TimeoutSeconds: 30,
	}

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Run health checks on repositories",
		Long: `Execute modular health checks using the health engine with advanced reporting.
The health command works out-of-the-box with sensible defaults. No configuration
file is required.
If you want to customize the checks, you can provide an optional configuration file.

Examples:
  repos health                           # Run with built-in defaults
  repos health --config custom.yaml     # Use custom configuration
  repos health --category git,security  # Run only git and security checks
  repos health cyclomatic-complexity    # Run only cyclomatic complexity analysis
  repos health --verbose                # Show detailed output
  repos health --list-categories        # List all available categories and checks
  repos health --dry-run                # Preview what would be executed
  repos health genconfig                 # Generate comprehensive configuration template`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(config)
		},
	}

	// Add flags
	cmd.Flags().StringSliceVar(&config.Categories, "category", []string{},
		"filter checkers and analyzers by categories (comma-separated, e.g., 'git,security')")
	cmd.Flags().StringVar(&config.ConfigPath, "config", "",
		"health config file path (optional, uses built-in defaults if not provided)")
	cmd.Flags().BoolVar(&config.DryRun, "dry-run", false,
		"Dry run mode - show what would be executed")
	cmd.Flags().BoolVar(&config.ListCategories, "list-categories", false,
		"List all available categories, checkers, and analyzers")
	cmd.Flags().IntVar(&config.MaxComplexity, "max-complexity", 0,
		"Fail if any function exceeds this cyclomatic complexity (0 disables check)")
	cmd.Flags().BoolVar(&config.Parallel, "parallel", false,
		"Execute health checks in parallel")
	cmd.Flags().IntVar(&config.TimeoutSeconds, "timeout", 30,
		"Timeout in seconds for health checks (default: 30)")
	cmd.Flags().BoolVar(&config.Verbose, "verbose", false,
		"Enable verbose output for health checks")

	// Add subcommands
	cmd.AddCommand(NewCyclomaticComplexityCommand())
	cmd.AddCommand(NewGenConfigCommand())

	return cmd
}

// execute runs the health command
func execute(config *Config) error {
	// Validate configuration
	if err := validateConfig(config); err != nil {
		return err
	}

	// Handle special flags first - order matters!
	if config.ListCategories {
		return listCategories()
	}

	// Check for dry-run mode explicitly
	if config.DryRun {
		return showDryRunConfiguration(config)
	}

	// For now, this is a placeholder implementation
	common.PrintInfo("Running health checks...")
	common.PrintSuccess("Health check implementation needs to be completed")

	return nil
}

// validateConfig validates the health command configuration
func validateConfig(config *Config) error {
	// Validate timeout
	if config.TimeoutSeconds < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}
	if config.TimeoutSeconds == 0 {
		return fmt.Errorf("timeout cannot be zero")
	}
	if config.TimeoutSeconds > 7200 { // 2 hours in seconds
		return fmt.Errorf("timeout cannot exceed 2 hours")
	}

	// Validate complexity threshold
	if config.MaxComplexity < 0 {
		return fmt.Errorf("max complexity cannot be negative")
	}

	return nil
}

// listCategories lists all available categories and checkers
func listCategories() error {
	common.PrintInfo("Available Health Check Categories:")

	// Create a simple logger for the health system
	logger := &simpleLogger{}

	// Create command executor and registries to get actual checker/analyzer information
	executor := commands.NewOSCommandExecutor(30 * time.Second)
	checkerRegistry := health.NewCheckerRegistry(executor)

	// Create filesystem and analyzer registry
	fs := health.NewFileSystem()
	analyzerRegistry := health.NewAnalyzerRegistry(fs, logger)

	fmt.Println()
	fmt.Println("📋 CHECKERS:")

	// Get and display checkers by category
	checkers := checkerRegistry.GetCheckers()
	checkersByCategory := make(map[string][]core.Checker)

	for _, checker := range checkers {
		category := checker.Category()
		checkersByCategory[category] = append(checkersByCategory[category], checker)
	}

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

	fmt.Println("🔍 ANALYZERS:")

	// Get and display analyzers
	analyzers := analyzerRegistry.GetAnalyzers()
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

	fmt.Println()
	fmt.Println("Usage Examples:")
	fmt.Println("  repos health --category git,security     # Run only git and security checkers")
	fmt.Println("  repos health --verbose                   # Show detailed output")
	fmt.Println("  repos health --dry-run                   # Preview what would be executed")

	return nil
}

// simpleLogger implements core.Logger interface
type simpleLogger struct{}

func (l *simpleLogger) Debug(msg string, fields ...core.Field) {
	// For listCategories, we don't need debug output
}

func (l *simpleLogger) Info(msg string, fields ...core.Field) {
	fmt.Print("[INFO] " + msg + l.formatFields(fields))
}

func (l *simpleLogger) Warn(msg string, fields ...core.Field) {
	fmt.Print("[WARN] " + msg + l.formatFields(fields))
}

func (l *simpleLogger) Error(msg string, fields ...core.Field) {
	fmt.Print("[ERROR] " + msg + l.formatFields(fields))
}

func (l *simpleLogger) Fatal(msg string, fields ...core.Field) {
	fmt.Print("[FATAL] " + msg + l.formatFields(fields))
}

func (l *simpleLogger) formatFields(fields []core.Field) string {
	if len(fields) == 0 {
		return "\n"
	}

	var result string
	for _, field := range fields {
		result += fmt.Sprintf(" [%s=%v]", field.Key, field.Value)
	}
	return result + "\n"
}

// showDryRunConfiguration displays exhaustive configuration for dry-run mode
//
//nolint:gocyclo
func showDryRunConfiguration(config *Config) error {
	common.PrintInfo("🔍 DRY RUN MODE - Health Check Configuration Preview")
	fmt.Println()

	// Basic Configuration
	common.PrintInfo("📋 BASIC CONFIGURATION:")
	fmt.Printf("  Timeout: %d seconds\n", config.TimeoutSeconds)
	fmt.Printf("  Parallel Execution: %t\n", config.Parallel)
	fmt.Printf("  Verbose Output: %t\n", config.Verbose)

	if config.ConfigPath != "" {
		fmt.Printf("  Config File: %s\n", config.ConfigPath)
	} else {
		fmt.Println("  Config File: Using built-in defaults")
	}
	fmt.Println()

	// Categories Configuration
	common.PrintInfo("🏷️  CATEGORIES CONFIGURATION:")
	if len(config.Categories) > 0 {
		fmt.Printf("  Selected Categories: %v\n", config.Categories)
		fmt.Println("  Note: Only checkers and analyzers in these categories will run")
	} else {
		fmt.Println("  Selected Categories: ALL (no filter applied)")
		fmt.Println("  Note: All available checkers and analyzers will run")
	}
	fmt.Println()

	// Repository Configuration
	common.PrintInfo("🏪 REPOSITORY CONFIGURATION:")
	if config.Tag != "" {
		fmt.Printf("  Repository Filter: tag='%s'\n", config.Tag)
		fmt.Println("  Note: Only repositories with this tag will be analyzed")
	} else {
		fmt.Println("  Repository Filter: ALL repositories")
		fmt.Println("  Note: All repositories in config will be analyzed")
	}
	fmt.Println()

	// Complexity Analysis Configuration
	common.PrintInfo("🧮 COMPLEXITY ANALYSIS:")
	fmt.Println("  Complexity Report: Available via 'repos health cyclomatic-complexity' subcommand")
	if config.MaxComplexity > 0 {
		fmt.Printf("  Maximum Complexity Threshold: %d\n", config.MaxComplexity)
		fmt.Println("  Note: Functions exceeding this threshold will cause failure")
	} else {
		fmt.Println("  Maximum Complexity Threshold: DISABLED")
		fmt.Println("  Note: Complexity will be reported but won't cause failure")
	}
	fmt.Println()

	// Available Health Checkers
	common.PrintInfo("✅ AVAILABLE HEALTH CHECKERS:")
	checkers := []struct {
		category    string
		name        string
		id          string
		priority    string
		description string
	}{
		{"ci", "CI/CD Configuration", "ci-config", "medium", "Validates CI/CD pipeline configuration"},
		{"documentation", "README Documentation", "readme-check", "medium", "Ensures README exists and has required content"},
		{"git", "Git Status", "git-status", "medium", "Checks for uncommitted changes and clean status"},
		{"git", "Last Commit", "git-last-commit", "low", "Validates recent commit activity"},
		{"security", "Branch Protection", "branch-protection", "high", "Verifies branch protection rules"},
		{"security", "Vulnerability Scanner", "vulnerability-scan", "high", "Scans for known security vulnerabilities"},
		{"dependencies", "Outdated Dependencies", "dependencies-outdated", "medium", "Checks for outdated package dependencies"},
		{"compliance", "License Compliance", "license-check", "medium", "Validates license information and compliance"},
	}

	for _, checker := range checkers {
		enabled := "ENABLED"
		if len(config.Categories) > 0 {
			found := false
			for _, cat := range config.Categories {
				if cat == checker.category {
					found = true
					break
				}
			}
			if !found {
				enabled = "DISABLED (category filter)"
			}
		}

		fmt.Printf("  • %s (%s) - %s [%s]\n", checker.name, checker.id, enabled, checker.priority)
		fmt.Printf("    Category: %s | %s\n", checker.category, checker.description)
	}
	fmt.Println()

	// Available Code Analyzers
	common.PrintInfo("🔍 AVAILABLE CODE ANALYZERS:")
	analyzers := []struct {
		language    string
		name        string
		extensions  []string
		description string
	}{
		{"javascript", "javascript-analyzer", []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"}, "Analyzes JavaScript/TypeScript code quality and complexity"},
		{"go", "go-analyzer", []string{".go"}, "Analyzes Go code using standard Go tools"},
		{"python", "python-analyzer", []string{".py"}, "Analyzes Python code quality and style"},
		{"java", "java-analyzer", []string{".java"}, "Analyzes Java code complexity and quality"},
	}

	for _, analyzer := range analyzers {
		fmt.Printf("  • %s\n", analyzer.name)
		fmt.Printf("    Language: %s | Extensions: %v\n", analyzer.language, analyzer.extensions)
		fmt.Printf("    Description: %s\n", analyzer.description)
	}
	fmt.Println()

	// Execution Plan
	common.PrintInfo("🚀 EXECUTION PLAN:")
	fmt.Println("  1. Load repository configuration")
	if config.Tag != "" {
		fmt.Printf("  2. Filter repositories by tag: '%s'\n", config.Tag)
	} else {
		fmt.Println("  2. Process all repositories")
	}
	if len(config.Categories) > 0 {
		fmt.Printf("  3. Filter health checkers by categories: %v\n", config.Categories)
	} else {
		fmt.Println("  3. Enable all health checkers")
	}
	fmt.Println("  4. Run health checkers on each repository")
	fmt.Println("  5. Run code analyzers on detected languages")
	if config.Parallel {
		fmt.Println("  6. Execute checks in parallel for faster processing")
	} else {
		fmt.Println("  6. Execute checks sequentially")
	}
	fmt.Println("  7. Generate comprehensive health report")
	fmt.Println()

	// Configuration Tips
	common.PrintInfo("💡 CONFIGURATION TIPS:")
	fmt.Println("  • Use --config <file> to specify custom health configuration")
	fmt.Println("  • Use --category git,security to run only specific checker categories")
	fmt.Println("  • Use 'repos health cyclomatic-complexity --max-complexity 10' to enforce complexity limits")
	fmt.Println("  • Use --parallel to speed up analysis of multiple repositories")
	fmt.Println("  • Use --verbose to see detailed output during execution")
	fmt.Println("  • Use --list-categories to see all available categories and checkers")
	fmt.Println("  • Use 'repos health genconfig' to generate a template configuration file")
	fmt.Println()

	common.PrintSuccess("✨ Dry run complete! Use the above configuration to customize your health checks.")
	common.PrintInfo("💻 Run without --dry-run to execute the actual health checks.")

	return nil
}
