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
	ConfigPath       string
	Categories       []string
	Pipeline         string
	Parallel         bool
	TimeoutSeconds   int // Timeout in seconds
	DryRun           bool
	Verbose          bool
	Tag              string
	BasicConfig      string // Path to basic repo config
	ListCategories   bool   // List available categories and checkers
	GenConfig        bool   // Generate configuration template
	ComplexityReport bool   // Run complexity analysis only
	MaxComplexity    int    // Maximum allowed complexity
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
  repos health --complexity-report      # Run only cyclomatic complexity analysis
  repos health --complexity-report --category docs,security # Run complexity and other checks
  repos health --verbose                # Show detailed output
  repos health --list-categories        # List all available categories and checks
  repos health --gen-config             # Generate comprehensive configuration template
  repos health --dry-run                # Preview what would be executed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(config)
		},
	}

	// Add flags
	cmd.Flags().StringSliceVar(&config.Categories, "category", []string{},
		"filter checkers and analyzers by categories (comma-separated, e.g., 'git,security')")
	cmd.Flags().BoolVar(&config.ComplexityReport, "complexity-report", false,
		"Generate a cyclomatic complexity report for the codebase")
	cmd.Flags().StringVar(&config.ConfigPath, "config", "",
		"health config file path (optional, uses built-in defaults if not provided)")
	cmd.Flags().BoolVar(&config.DryRun, "dry-run", false,
		"Dry run mode - show what would be executed")
	cmd.Flags().BoolVar(&config.GenConfig, "gen-config", false,
		"Generate a comprehensive configuration template with all available options")
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
	cmd.AddCommand(NewComplexityCommand())
	cmd.AddCommand(NewDryRunCommand())
	cmd.AddCommand(NewGenConfigCommand())

	return cmd
}

// execute runs the health command
func execute(config *Config) error {
	// Validate configuration
	if err := validateConfig(config); err != nil {
		return err
	}

	// Handle special flags first
	if config.ListCategories {
		return listCategories()
	}

	if config.GenConfig {
		return generateConfig()
	}

	// For now, this is a placeholder implementation
	common.PrintInfo("Running health checks...")
	common.PrintSuccess("Health check implementation needs to be completed")

	return nil
}

// validateConfig validates the health command configuration
func validateConfig(config *Config) error {
	// Validate timeout
	if config.TimeoutSeconds > 7200 { // 2 hours in seconds
		return fmt.Errorf("timeout cannot exceed 2 hours")
	}
	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = 30
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

// generateConfig generates a comprehensive configuration template
func generateConfig() error {
	common.PrintInfo("Generating comprehensive health configuration template...")

	// This would delegate to the internal health package
	// For now, we'll show a placeholder
	common.PrintSuccess("Configuration template would be generated here")

	return nil
}
