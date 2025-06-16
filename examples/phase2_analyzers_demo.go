package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/codcod/repos/internal/analyzers/registry"
	"github.com/codcod/repos/internal/checkers/quality/complexity"
	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/platform/filesystem"
)

// SimpleLogger implements core.Logger for demonstration
type SimpleLogger struct{}

func (l *SimpleLogger) Debug(message string, fields ...core.Field) {
	fmt.Printf("[DEBUG] %s\n", message)
}

func (l *SimpleLogger) Info(message string, fields ...core.Field) {
	fmt.Printf("[INFO] %s\n", message)
}

func (l *SimpleLogger) Warn(message string, fields ...core.Field) {
	fmt.Printf("[WARN] %s\n", message)
}

func (l *SimpleLogger) Error(message string, fields ...core.Field) {
	fmt.Printf("[ERROR] %s\n", message)
}

// SimpleCache implements core.Cache for demonstration
type SimpleCache struct {
	data map[string]interface{}
}

func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		data: make(map[string]interface{}),
	}
}

func (c *SimpleCache) Get(key string) (interface{}, bool) {
	value, exists := c.data[key]
	return value, exists
}

func (c *SimpleCache) Set(key string, value interface{}, ttl time.Duration) {
	c.data[key] = value
}

func (c *SimpleCache) Delete(key string) {
	delete(c.data, key)
}

func (c *SimpleCache) Clear() {
	c.data = make(map[string]interface{})
}

func main() {
	fmt.Println("Demonstrating Phase 2: Language Analyzers")
	fmt.Println("==========================================")

	// Create dependencies
	logger := &SimpleLogger{}
	filesystem := filesystem.NewOSFileSystem()
	cache := NewSimpleCache()

	// Create analyzer registry with all standard analyzers
	analyzerRegistry := registry.NewRegistryWithStandardAnalyzers(filesystem, logger)

	// Load configuration
	configLoader := config.NewModularConfigLoader(filesystem, logger)
	cfg, err := configLoader.LoadConfig("config/modular-health.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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
			functions := result.Functions
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
		FileSystem: filesystem,
		Cache:      cache,
		Logger:     logger,
	}

	complexityChecker := complexity.NewComplexityChecker(analyzerRegistry, filesystem, logger)

	result, err := complexityChecker.Check(ctx, repoCtx)
	if err != nil {
		fmt.Printf("❌ Complexity check failed: %v\n", err)
	} else {
		fmt.Printf("✓ Complexity check completed:\n")
		fmt.Printf("  - Status: %s\n", result.Status)
		fmt.Printf("  - Score: %d/%d\n", result.Score, result.MaxScore)
		fmt.Printf("  - Issues: %d\n", len(result.Issues))
		fmt.Printf("  - Duration: %v\n", result.Duration)
	}

	// Summary
	fmt.Println("\n📊 Analysis Summary:")
	fmt.Printf("• Total files analyzed: %d\n", totalFiles)
	fmt.Printf("• Total functions found: %d\n", totalFunctions)
	fmt.Printf("• Languages supported: %d\n", len(analyzerRegistry.GetAllAnalyzers()))

	fmt.Println("\n✨ Phase 2 (Language Analyzers) demonstration complete!")
	fmt.Println("\nKey Phase 2 Achievements:")
	fmt.Println("• Language-specific analyzers extracted from monolithic code")
	fmt.Println("• Comprehensive analysis for Go, Python, Java, JavaScript/TypeScript")
	fmt.Println("• Function-level complexity analysis")
	fmt.Println("• Import/dependency analysis")
	fmt.Println("• Registry-based analyzer discovery")
	fmt.Println("• Seamless integration with existing checker framework")
}
