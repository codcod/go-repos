package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/codcod/repos/internal/checkers/ci"
	"github.com/codcod/repos/internal/checkers/compliance"
	"github.com/codcod/repos/internal/checkers/git"
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/platform/commands"
)

func main() {
	fmt.Println("=== Phase 3 Basic Checkers Test ===")
	fmt.Println("Testing basic checkers without external dependencies")
	fmt.Println()

	// Create command executor
	executor := commands.NewOSCommandExecutor(30 * time.Second)

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

	// Test basic checkers
	checkers := []struct {
		name    string
		checker core.Checker
	}{
		{"Git Status", git.NewGitStatusChecker(executor)},
		{"Last Commit", git.NewLastCommitChecker(executor)},
		{"License", compliance.NewLicenseChecker()},
		{"CI Config", ci.NewCIConfigChecker()},
	}

	fmt.Println("🔍 Testing basic checkers...")

	for _, test := range checkers {
		fmt.Printf("Testing %s...", test.name)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		start := time.Now()
		result, err := test.checker.Check(ctx, repoCtx)
		duration := time.Since(start)
		cancel()

		if err != nil {
			fmt.Printf(" ❌ FAILED (%v) - %v\n", duration, err)
			continue
		}

		fmt.Printf(" ✅ %s (%v) - Score: %d/%d\n", result.Status, duration, result.Score, result.MaxScore)

		// Show a few key metrics
		if len(result.Metrics) > 0 {
			fmt.Printf("    Key metrics: ")
			count := 0
			for key, value := range result.Metrics {
				if count >= 3 { // Limit to 3 metrics
					break
				}
				fmt.Printf("%s=%v ", key, value)
				count++
			}
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Println("✅ Basic checker testing completed!")
}
