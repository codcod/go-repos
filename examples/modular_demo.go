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

// Example Go analyzer (simplified)
type GoAnalyzer struct {
	*registry.BaseAnalyzer
}

func NewGoAnalyzer() *GoAnalyzer {
	return &GoAnalyzer{
		BaseAnalyzer: registry.NewBaseAnalyzer("go", []string{".go"}, true, true),
	}
}

func (a *GoAnalyzer) AnalyzeComplexity(ctx context.Context, repoPath string) (core.ComplexityResult, error) {
	// Simplified complexity analysis
	return core.ComplexityResult{
		TotalFiles:        5,
		TotalFunctions:    20,
		AverageComplexity: 6.5,
		MaxComplexity:     12,
		Functions: []core.FunctionComplexity{
			{
				Name:       "ExampleFunction",
				File:       "example.go",
				Line:       10,
				Complexity: 12,
				Length:     50,
			},
		},
		FileMetrics: map[string]core.FileMetrics{
			"example.go": {
				Path:              "example.go",
				Language:          "go",
				Lines:             100,
				Functions:         4,
				AverageComplexity: 8.0,
				MaxComplexity:     12,
			},
		},
	}, nil
}

func main() {
	fmt.Println("Demonstrating Modular Health Check Architecture")
	fmt.Println("==============================================")

	// 1. Create configuration
	cfg := config.DefaultModularConfig()
	fmt.Printf("✓ Configuration loaded with %d checkers\n", len(cfg.Checkers))

	// 2. Create platform services
	fs := filesystem.NewOSFileSystem()
	cache := NewSimpleCache()
	logger := &SimpleLogger{}

	// 3. Create analyzer registry and register analyzers
	analyzerRegistry := registry.NewRegistry()
	goAnalyzer := NewGoAnalyzer()
	analyzerRegistry.Register(goAnalyzer)
	fmt.Printf("✓ Analyzer registry created with %d analyzers\n", len(analyzerRegistry.GetAllAnalyzers()))

	// 4. Create checkers
	complexityConfig := complexity.ComplexityConfig{
		CheckerConfig: core.CheckerConfig{
			Enabled:  true,
			Severity: "medium",
			Timeout:  30 * time.Second,
		},
		Thresholds: map[string]int{
			"go": 10,
		},
		ReportLevel:    "detailed",
		IncludeTests:   false,
		DetailedReport: true,
	}

	complexityChecker := complexity.NewComplexityChecker(analyzerRegistry, complexityConfig)
	fmt.Printf("✓ Complexity checker created: %s\n", complexityChecker.Name())

	// 5. Create repository context
	repo := core.Repository{
		Name: "example-repo",
		Path: "/Users/nicos/Projects/private/repos", // Use current repo for demo
		URL:  "https://github.com/example/repo.git",
	}

	repoCtx := core.RepositoryContext{
		Repository: repo,
		Config:     cfg,
		FileSystem: fs,
		Cache:      cache,
		Logger:     logger,
	}

	// 6. Run health check
	ctx := context.Background()
	fmt.Println("\n🔍 Running health checks...")

	result, err := complexityChecker.Check(ctx, repoCtx)
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}

	// 7. Display results
	fmt.Println("\n📊 Results:")
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Score: %d/%d\n", result.Score, result.MaxScore)
	fmt.Printf("Issues: %d\n", len(result.Issues))
	fmt.Printf("Warnings: %d\n", len(result.Warnings))
	fmt.Printf("Duration: %v\n", result.Duration)

	// Display metrics
	if len(result.Metrics) > 0 {
		fmt.Println("\nMetrics:")
		for key, value := range result.Metrics {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	// Display issues
	if len(result.Issues) > 0 {
		fmt.Println("\nIssues:")
		for i, issue := range result.Issues {
			fmt.Printf("  %d. [%s] %s\n", i+1, issue.Severity, issue.Message)
			if issue.Location != nil {
				fmt.Printf("     Location: %s:%d\n", issue.Location.File, issue.Location.Line)
			}
			if issue.Suggestion != "" {
				fmt.Printf("     Suggestion: %s\n", issue.Suggestion)
			}
		}
	}

	fmt.Println("\n✨ Modular architecture demonstration complete!")
	fmt.Println("\nKey Benefits Demonstrated:")
	fmt.Println("• Clear separation of concerns")
	fmt.Println("• Dependency injection")
	fmt.Println("• Pluggable analyzers")
	fmt.Println("• Configuration-driven setup")
	fmt.Println("• Structured results")
	fmt.Println("• Platform abstraction")
}
