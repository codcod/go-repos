package main

import (
	"context"
	"fmt"
	"time"

	"github.com/codcod/repos/internal/analyzers/registry"
	"github.com/codcod/repos/internal/checkers/quality/complexity"
	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/platform/filesystem"
)

// DemoLogger implements core.Logger for demonstration
type DemoLogger struct{}

func (l *DemoLogger) Debug(message string, fields ...core.Field) {
	fmt.Printf("[DEBUG] %s\n", message)
}

func (l *DemoLogger) Info(message string, fields ...core.Field) {
	fmt.Printf("[INFO] %s\n", message)
}

func (l *DemoLogger) Warn(message string, fields ...core.Field) {
	fmt.Printf("[WARN] %s\n", message)
}

func (l *DemoLogger) Error(message string, fields ...core.Field) {
	fmt.Printf("[ERROR] %s\n", message)
}

// DemoCache implements core.Cache for demonstration
type DemoCache struct {
	data map[string]interface{}
}

func NewDemoCache() *DemoCache {
	return &DemoCache{
		data: make(map[string]interface{}),
	}
}

func (c *DemoCache) Get(key string) (interface{}, bool) {
	value, exists := c.data[key]
	return value, exists
}

func (c *DemoCache) Set(key string, value interface{}, ttl time.Duration) {
	c.data[key] = value
}

func (c *DemoCache) Delete(key string) {
	delete(c.data, key)
}

func (c *DemoCache) Clear() {
	c.data = make(map[string]interface{})
}

func main() {
	fmt.Println("Demonstrating Phase 2: Language Analyzers")
	fmt.Println("==========================================")

	// Create dependencies
	logger := &DemoLogger{}
	fs := filesystem.NewOSFileSystem()
	cache := NewDemoCache()

	// Create analyzer registry with all standard analyzers
	analyzerRegistry := registry.NewRegistryWithStandardAnalyzers(fs, logger)

	// Load configuration (create minimal config if not available)
	cfg := createMinimalConfig()

	fmt.Printf("✓ Configuration loaded with %d checkers\n", len(cfg.Checkers))
	fmt.Printf("✓ Analyzer registry created with %d analyzers\n", len(analyzerRegistry.GetAllAnalyzers()))

	// List available analyzers
	fmt.Println("\n📋 Available Language Analyzers:")
	for _, analyzer := range analyzerRegistry.GetAllAnalyzers() {
		fmt.Printf("  • %s (%s) - Extensions: %v\n",
			analyzer.Name(),
			analyzer.Language(),
			analyzer.SupportedExtensions())
	}

	// Test repository analysis
	testRepo := core.Repository{
		Name:     "test-repo",
		Path:     ".",
		Language: "go",
	}

	fmt.Println("\n🔍 Testing Repository Analysis...")

	// Get supported analyzers for this repository
	supportedAnalyzers := analyzerRegistry.GetSupportedAnalyzers(testRepo)
	fmt.Printf("✓ Found %d supported analyzers for repository\n", len(supportedAnalyzers))

	// Run analysis with each supported analyzer
	ctx := context.Background()
	totalFunctions := 0
	totalFiles := 0

	for _, analyzer := range supportedAnalyzers {
		fmt.Printf("\n🔄 Running %s analyzer...\n", analyzer.Name())

		result, err := analyzer.Analyze(ctx, testRepo)
		if err != nil {
			fmt.Printf("❌ Analysis failed: %v\n", err)
			continue
		}

		fmt.Printf("✓ Analysis completed:")
		fmt.Printf("  - Files analyzed: %d\n", len(result.Files))
		fmt.Printf("  - Functions found: %d\n", len(result.Functions))
		fmt.Printf("  - Language: %s\n", result.Language)

		// Show metrics
		if metrics := result.Metrics; metrics != nil {
			if avgComplexity, ok := metrics["average_complexity"].(float64); ok {
				fmt.Printf("  - Average complexity: %.2f\n", avgComplexity)
			}
			if maxComplexity, ok := metrics["max_complexity"].(int); ok {
				fmt.Printf("  - Max complexity: %d\n", maxComplexity)
			}
		}

		// Show top 3 most complex functions
		if len(result.Functions) > 0 {
			fmt.Printf("  - Top complex functions:\n")
			// Sort by complexity (simple bubble sort for demo)
			functions := make([]core.FunctionInfo, len(result.Functions))
			copy(functions, result.Functions)

			for i := 0; i < len(functions)-1; i++ {
				for j := 0; j < len(functions)-i-1; j++ {
					if functions[j].Complexity < functions[j+1].Complexity {
						functions[j], functions[j+1] = functions[j+1], functions[j]
					}
				}
			}

			// Show top 3
			limit := 3
			if len(functions) < limit {
				limit = len(functions)
			}
			for i := 0; i < limit; i++ {
				fn := functions[i]
				fmt.Printf("    %d. %s (complexity: %d)\n", i+1, fn.Name, fn.Complexity)
			}
		}

		totalFunctions += len(result.Functions)
		totalFiles += len(result.Files)
	}

	// Create a complexity checker to demonstrate integration
	fmt.Println("\n🔧 Testing Complexity Checker Integration...")

	repoCtx := core.RepositoryContext{
		Repository: testRepo,
		Config:     cfg,
		FileSystem: fs,
		Cache:      cache,
		Logger:     logger,
	}

	// Create complexity config
	complexityConfig := complexity.ComplexityConfig{
		CheckerConfig: core.CheckerConfig{
			Enabled:    true,
			Severity:   "medium",
			Timeout:    30 * time.Second,
			Categories: []string{"quality"},
		},
		Thresholds: map[string]int{
			"go":         10,
			"python":     8,
			"java":       12,
			"javascript": 10,
		},
		ReportLevel:    "function",
		IncludeTests:   false,
		DetailedReport: true,
	}

	complexityChecker := complexity.NewComplexityChecker(analyzerRegistry, complexityConfig)

	result, err := complexityChecker.Check(ctx, repoCtx)
	if err != nil {
		fmt.Printf("❌ Complexity check failed: %v\n", err)
	} else {
		fmt.Printf("✓ Complexity check completed:\n")
		fmt.Printf("  - Status: %s\n", result.Status)
		fmt.Printf("  - Score: %d/%d\n", result.Score, result.MaxScore)
		fmt.Printf("  - Issues: %d\n", len(result.Issues))
		fmt.Printf("  - Duration: %v\n", result.Duration)

		// Show some issues
		if len(result.Issues) > 0 {
			fmt.Printf("  - Sample issues:\n")
			for i, issue := range result.Issues {
				if i >= 3 { // Show max 3 issues
					break
				}
				fmt.Printf("    • %s: %s\n", issue.Severity, issue.Message)
			}
		}
	}

	// Summary
	fmt.Println("\n📊 Analysis Summary:")
	fmt.Printf("• Total files analyzed: %d\n", totalFiles)
	fmt.Printf("• Total functions found: %d\n", totalFunctions)
	fmt.Printf("• Languages supported: %d\n", len(analyzerRegistry.GetAllAnalyzers()))

	fmt.Println("\n✨ Phase 2 (Language Analyzers) demonstration complete!")
	fmt.Println("\nKey Phase 2 Achievements:")
	fmt.Println("• ✅ Language-specific analyzers extracted from monolithic code")
	fmt.Println("• ✅ Comprehensive analysis for Go, Python, Java, JavaScript/TypeScript")
	fmt.Println("• ✅ Function-level complexity analysis")
	fmt.Println("• ✅ Import/dependency analysis")
	fmt.Println("• ✅ Registry-based analyzer discovery")
	fmt.Println("• ✅ Seamless integration with existing checker framework")
}

// createMinimalConfig creates a minimal configuration for the demo
func createMinimalConfig() *config.ModularConfig {
	return &config.ModularConfig{
		Checkers: map[string]core.CheckerConfig{
			"cyclomatic-complexity": {
				Enabled:  true,
				Severity: "medium",
				Timeout:  30 * time.Second,
				Options: map[string]interface{}{
					"thresholds": map[string]interface{}{
						"go":         10,
						"python":     8,
						"java":       12,
						"javascript": 10,
					},
				},
				Categories: []string{"quality"},
			},
		},
		Analyzers: map[string]core.AnalyzerConfig{
			"go": {
				Enabled:           true,
				FileExtensions:    []string{".go"},
				ExcludePatterns:   []string{"vendor/", "_test.go"},
				ComplexityEnabled: true,
				FunctionLevel:     true,
			},
			"python": {
				Enabled:           true,
				FileExtensions:    []string{".py"},
				ExcludePatterns:   []string{".venv/", "__pycache__/"},
				ComplexityEnabled: true,
				FunctionLevel:     true,
			},
			"java": {
				Enabled:           true,
				FileExtensions:    []string{".java"},
				ExcludePatterns:   []string{"target/", "build/"},
				ComplexityEnabled: true,
				FunctionLevel:     true,
			},
			"javascript": {
				Enabled:           true,
				FileExtensions:    []string{".js", ".jsx", ".ts", ".tsx"},
				ExcludePatterns:   []string{"node_modules/", "dist/"},
				ComplexityEnabled: true,
				FunctionLevel:     true,
			},
		},
		Reporters: map[string]core.ReporterConfig{
			"console": {
				Enabled:  true,
				Template: "table",
				Options: map[string]interface{}{
					"show_summary": true,
				},
			},
		},
		Engine: core.EngineConfig{
			MaxConcurrency: 4,
			Timeout:        5 * time.Minute,
			CacheEnabled:   true,
			CacheTTL:       5 * time.Minute,
		},
	}
}
