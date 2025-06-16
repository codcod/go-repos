package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/codcod/repos/internal/checkers/registry"
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/platform/commands"
)

func main() {
	fmt.Println("=== Phase 3 Checkers Demo ===")
	fmt.Println("Testing modular checker implementation")
	fmt.Println()

	// Create command executor
	executor := commands.NewOSCommandExecutor(30 * time.Second)

	// Create checker registry
	checkerRegistry := registry.NewCheckerRegistry(executor)

	// Get current working directory as test repository
	repoPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Create repository context
	repo := core.Repository{
		Name: "test-repo",
		Path: repoPath,
		URL:  "https://github.com/example/test-repo",
	}

	repoCtx := core.RepositoryContext{
		Repository: repo,
	}

	// List all available checkers
	fmt.Println("📋 Available Checkers:")
	checkerInfos := checkerRegistry.ListCheckers()
	for _, info := range checkerInfos {
		status := "disabled"
		if info.Enabled {
			status = "enabled"
		}
		fmt.Printf("  • %s (%s) - %s [%s]\n", info.Name, info.ID, info.Category, status)
	}
	fmt.Println()

	// Get registry stats
	stats := checkerRegistry.GetStats()
	fmt.Printf("📊 Registry Stats:\n")
	fmt.Printf("  • Total checkers: %d\n", stats.TotalCheckers)
	fmt.Printf("  • Enabled checkers: %d\n", stats.EnabledCheckers)
	fmt.Printf("  • Categories: %v\n", stats.Categories)
	fmt.Println()

	// Run all enabled checkers
	fmt.Println("🚀 Running all enabled checkers...")
	ctx := context.Background()

	// Create a simple config that enables all checkers
	config := make(map[string]core.CheckerConfig)
	for _, info := range checkerInfos {
		config[info.ID] = core.CheckerConfig{
			Enabled:  true,
			Severity: info.Severity,
			Timeout:  30 * time.Second,
		}
	}

	results, err := checkerRegistry.RunAllEnabledCheckers(ctx, repoCtx, config)
	if err != nil {
		log.Printf("Some checkers failed: %v", err)
	}

	// Display results
	fmt.Printf("\n✅ Completed %d checks\n\n", len(results))

	for i, result := range results {
		fmt.Printf("--- Check %d: %s ---\n", i+1, result.Name)
		fmt.Printf("Status: %s\n", result.Status)
		fmt.Printf("Score: %d/%d\n", result.Score, result.MaxScore)
		fmt.Printf("Category: %s\n", result.Category)

		if len(result.Issues) > 0 {
			fmt.Printf("Issues (%d):\n", len(result.Issues))
			for _, issue := range result.Issues {
				fmt.Printf("  • [%s] %s", issue.Severity, issue.Message)
				if issue.Suggestion != "" {
					fmt.Printf(" → %s", issue.Suggestion)
				}
				fmt.Println()
			}
		}

		if len(result.Warnings) > 0 {
			fmt.Printf("Warnings (%d):\n", len(result.Warnings))
			for _, warning := range result.Warnings {
				fmt.Printf("  • %s\n", warning.Message)
			}
		}

		if len(result.Metrics) > 0 {
			fmt.Printf("Metrics (%d):\n", len(result.Metrics))
			for key, value := range result.Metrics {
				fmt.Printf("  • %s: %v\n", key, value)
			}
		}

		fmt.Printf("Duration: %v\n", result.Duration)
		fmt.Println()
	}

	// Summary
	var healthyCount, warningCount, criticalCount int
	for _, result := range results {
		switch result.Status {
		case core.StatusHealthy:
			healthyCount++
		case core.StatusWarning:
			warningCount++
		case core.StatusCritical:
			criticalCount++
		}
	}

	fmt.Println("📈 Summary:")
	fmt.Printf("  • Healthy: %d\n", healthyCount)
	fmt.Printf("  • Warning: %d\n", warningCount)
	fmt.Printf("  • Critical: %d\n", criticalCount)
	fmt.Println()

	fmt.Println("✅ Phase 3 Checkers Demo completed successfully!")
}
