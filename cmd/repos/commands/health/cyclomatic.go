// Package health provides health command subcommands
package health

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/codcod/repos/cmd/repos/common"
	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/analyzers"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// ComplexityConfig contains configuration for cyclomatic analysis
type ComplexityConfig struct {
	Tag           string
	MaxComplexity int
}

// NewCyclomaticComplexityCommand creates the cyclomatic subcommand
func NewCyclomaticComplexityCommand() *cobra.Command {
	complexityConfig := &ComplexityConfig{
		MaxComplexity: 10,
	}

	cmd := &cobra.Command{
		Use:   "cyclomatic",
		Short: "Run cyclomatic analysis",
		Long:  `Analyze the cyclomatic complexity of code in repositories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, _ := cmd.Flags().GetString("config")

			cfg, err := common.LoadConfig(configFile)
			if err != nil {
				return err
			}

			repositories := cfg.FilterRepositoriesByTag(complexityConfig.Tag)
			if len(repositories) == 0 {
				color.Yellow("No repositories found with tag: %s", complexityConfig.Tag)
				return nil
			}

			color.Green("Running cyclomatic analysis on %d repositories...", len(repositories))

			// Analyze cyclomatic for each repository
			return analyzeComplexity(repositories, complexityConfig.MaxComplexity)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&complexityConfig.Tag, "tag", "t", "", "Filter repositories by tag")
	cmd.Flags().IntVar(&complexityConfig.MaxComplexity, "max-complexity", 10, "Maximum allowed cyclomatic complexity")

	return cmd
}

// analyzeComplexity performs cyclomatic analysis on repositories
//
//nolint:gocyclo
func analyzeComplexity(repositories []config.Repository, maxComplexity int) error {
	// Create logger
	logger := &complexityLogger{}

	ctx := context.Background()
	hasHighComplexity := false
	successfulRepos := 0

	fmt.Println()
	color.Cyan("=== Cyclomatic Report (Maximum complexity threshold: %d) ===", maxComplexity)
	fmt.Println()

	for _, repo := range repositories {
		repoPath := repo.Path
		if repoPath == "" {
			repoPath = filepath.Join(".", repo.Name)
		}

		// Check if repository path exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			color.Yellow("%s | Repository path not found: %s", repo.Name, repoPath)
			continue
		}

		// Detect language by examining files
		language := detectLanguage(repoPath)
		if language == "" {
			color.Yellow("%s | No supported language detected", repo.Name)
			continue
		}

		// Get analyzer for the detected language
		analyzer, err := analyzers.GetAnalyzer(language, logger)
		if err != nil {
			color.Yellow("%s | No analyzer available for language: %s", repo.Name, language)
			continue
		}

		// Check if analyzer supports cyclomatic analysis
		if analyzer.SupportsComplexity() {
			// Use new cyclomatic analysis
			result, err := analyzer.AnalyzeComplexity(ctx, repoPath)
			if err != nil {
				color.Red("%s | Error analyzing cyclomatic: %v", repo.Name, err)
				continue
			}

			// Display results for this repository
			displayComplexityResult(repo.Name, result, maxComplexity, &hasHighComplexity)
			successfulRepos++
		} else {
			color.Yellow("%s | Cyclomatic analysis not supported for language: %s", repo.Name, language)
		}
	}

	// Check for threshold violations and exit with non-zero status
	if hasHighComplexity && maxComplexity > 0 {
		os.Exit(1)
	}

	if successfulRepos == 0 {
		color.Yellow("⚠️  No repositories could be analyzed")
		return nil
	}

	return nil
}

// detectLanguage detects the primary language of a repository
//
//nolint:gocyclo
func detectLanguage(repoPath string) string {
	languageFiles := map[string][]string{
		"go":         {".go"},
		"python":     {".py"},
		"javascript": {".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		"java":       {".java"},
	}

	languageCounts := make(map[string]int)

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		// Skip hidden directories and common excludes
		if info.IsDir() {
			name := info.Name()
			if name[0] == '.' || name == "node_modules" || name == "vendor" || name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		for lang, extensions := range languageFiles {
			for _, langExt := range extensions {
				if ext == langExt {
					languageCounts[lang]++
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		return ""
	}

	// Find the language with the most files
	maxCount := 0
	detectedLanguage := ""
	for lang, count := range languageCounts {
		if count > maxCount {
			maxCount = count
			detectedLanguage = lang
		}
	}

	return detectedLanguage
}

// displayComplexityResult displays complexity results for a repository
func displayComplexityResult(repoName string, result core.ComplexityResult, threshold int, hasHighComplexity *bool) {
	color.Blue("Repository: %s", repoName)
	fmt.Printf("  Files analyzed: %d\n", result.TotalFiles)
	fmt.Printf("  Total functions: %d\n", result.TotalFunctions)

	if result.TotalFunctions > 0 {
		fmt.Printf("  Average complexity: %.2f\n", result.AverageComplexity)
		fmt.Printf("  Maximum complexity: %d\n", result.MaxComplexity)
	}

	// Show functions that exceed threshold
	if threshold > 0 {
		highComplexityFunctions := make([]core.FunctionComplexity, 0)
		for _, fn := range result.Functions {
			if fn.Complexity > threshold {
				highComplexityFunctions = append(highComplexityFunctions, fn)
			}
		}

		if len(highComplexityFunctions) > 0 {
			*hasHighComplexity = true
			color.Red("  ⚠️  Functions exceeding threshold (%d):", threshold)

			// Sort by complexity (highest first)
			sort.Slice(highComplexityFunctions, func(i, j int) bool {
				return highComplexityFunctions[i].Complexity > highComplexityFunctions[j].Complexity
			})

			for _, fn := range highComplexityFunctions {
				color.Red("    %s:%d - %s (complexity: %d)",
					fn.File, fn.Line, fn.Name, fn.Complexity)
			}
		} else {
			color.Green("  ✅ All functions are within the complexity threshold")
		}
	}

	fmt.Println()
}

// complexityLogger implements core.Logger interface for cyclomatic analysis
type complexityLogger struct{}

func (l *complexityLogger) Debug(msg string, fields ...core.Field) {
	// Debug messages are suppressed for cleaner output
}

func (l *complexityLogger) Info(msg string, fields ...core.Field) {
	// Info messages are suppressed for cleaner output
}

func (l *complexityLogger) Warn(msg string, fields ...core.Field) {
	fmt.Print("[WARN] " + msg + l.formatFields(fields))
}

func (l *complexityLogger) Error(msg string, fields ...core.Field) {
	fmt.Print("[ERROR] " + msg + l.formatFields(fields))
}

func (l *complexityLogger) Fatal(msg string, fields ...core.Field) {
	fmt.Print("[FATAL] " + msg + l.formatFields(fields))
}

func (l *complexityLogger) formatFields(fields []core.Field) string {
	if len(fields) == 0 {
		return "\n"
	}

	var result string
	for _, field := range fields {
		result += fmt.Sprintf(" [%s=%v]", field.Key, field.Value)
	}
	return result + "\n"
}
