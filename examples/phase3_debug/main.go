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
	fmt.Println("=== Phase 3 Checkers Demo (Debug) ===")
	fmt.Println("Testing individual checkers with timeouts")
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
		fmt.Printf("  • %s (%s) - %s\n", info.Name, info.ID, info.Category)
	}
	fmt.Println()

	// Test each checker individually with timeout
	fmt.Println("🔍 Testing each checker individually...")

	for _, info := range checkerInfos {
		fmt.Printf("Testing %s (%s)...", info.Name, info.ID)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		start := time.Now()
		result, err := checkerRegistry.RunChecker(ctx, info.ID, repoCtx)
		duration := time.Since(start)
		cancel()

		if err != nil {
			fmt.Printf(" ❌ FAILED (%v) - %v\n", duration, err)
			continue
		}

		fmt.Printf(" ✅ %s (%v) - Score: %d/%d\n", result.Status, duration, result.Score, result.MaxScore)

		// Show issues if any
		if len(result.Issues) > 0 {
			fmt.Printf("    Issues: %d\n", len(result.Issues))
			for i, issue := range result.Issues {
				if i >= 2 { // Limit to first 2 issues
					fmt.Printf("    ... and %d more\n", len(result.Issues)-2)
					break
				}
				fmt.Printf("    - %s\n", issue.Message)
			}
		}

		// Show warnings if any
		if len(result.Warnings) > 0 {
			fmt.Printf("    Warnings: %d\n", len(result.Warnings))
			for i, warning := range result.Warnings {
				if i >= 2 { // Limit to first 2 warnings
					fmt.Printf("    ... and %d more\n", len(result.Warnings)-2)
					break
				}
				fmt.Printf("    - %s\n", warning.Message)
			}
		}
	}

	fmt.Println()
	fmt.Println("✅ Individual checker testing completed!")
}
