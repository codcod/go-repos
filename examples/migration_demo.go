package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/codcod/repos/internal/analyzers/registry"
	checkersRegistry "github.com/codcod/repos/internal/checkers/registry"
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/platform/commands"
)

func main() {
	fmt.Println("🚀 ===== MODULAR ARCHITECTURE MIGRATION DEMO =====")
	fmt.Println("   Showcasing Phases 1-3 Implementation")
	fmt.Println("   Foundation ✅ | Analyzers ✅ | Checkers ✅")
	fmt.Println()

	// Get repository context
	repoPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	repo := core.Repository{
		Name: "repos-health-checker",
		Path: repoPath,
		URL:  "https://github.com/codcod/repos",
	}

	repoCtx := core.RepositoryContext{
		Repository: repo,
	}

	// === PHASE 1: FOUNDATION ===
	fmt.Println("📐 Phase 1: Foundation")
	fmt.Println("   ✅ Core interfaces and types")
	fmt.Println("   ✅ Base checker framework")
	fmt.Println("   ✅ Platform abstractions")
	fmt.Println()

	// Show platform abstraction
	executor := commands.NewOSCommandExecutor(30 * time.Second)
	fmt.Printf("   🔧 Command Executor: %T\n", executor)
	fmt.Println()

	// === PHASE 2: ANALYZERS ===
	fmt.Println("📊 Phase 2: Language Analyzers")

	analyzerRegistry := registry.NewRegistry()
	supportedAnalyzers := analyzerRegistry.GetSupportedAnalyzers(repo)

	fmt.Printf("   📈 Analyzer Registry: %d analyzers registered\n", len(analyzerRegistry.GetAnalyzers()))
	fmt.Printf("   🎯 Supported for this repo: %d analyzers\n", len(supportedAnalyzers))

	for _, analyzer := range supportedAnalyzers {
		fmt.Printf("     • %s (%s) - %v\n",
			analyzer.Name(),
			analyzer.Language(),
			analyzer.SupportedExtensions())
	}
	fmt.Println()

	// Run a quick complexity analysis
	if len(supportedAnalyzers) > 0 {
		fmt.Println("   🔍 Sample Analysis (Go files):")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		for _, analyzer := range supportedAnalyzers {
			if analyzer.Language() == "go" {
				result, err := analyzer.Analyze(ctx, repo)
				if err != nil {
					fmt.Printf("     ❌ Analysis failed: %v\n", err)
				} else {
					fmt.Printf("     ✅ %s: %d files, %.1f avg complexity\n",
						analyzer.Language(),
						result.FileCount,
						result.AverageComplexity)
				}
				break
			}
		}
	}
	fmt.Println()

	// === PHASE 3: CHECKERS ===
	fmt.Println("🔍 Phase 3: Modular Checkers")

	checkerRegistry := checkersRegistry.NewCheckerRegistry(executor)
	checkerInfos := checkerRegistry.ListCheckers()
	stats := checkerRegistry.GetStats()

	fmt.Printf("   🏗️  Checker Registry: %d checkers registered\n", stats.TotalCheckers)
	fmt.Printf("   ⚡ Enabled checkers: %d\n", stats.EnabledCheckers)
	fmt.Printf("   📂 Categories: %v\n", formatCategories(stats.Categories))
	fmt.Println()

	// Show checker details by category
	categories := []string{"git", "security", "dependencies", "compliance", "ci"}
	for _, category := range categories {
		categoryCheckers := checkerRegistry.GetCheckersByCategory(category)
		if len(categoryCheckers) > 0 {
			fmt.Printf("   📋 %s checkers (%d):\n", strings.Title(category), len(categoryCheckers))
			for _, checker := range categoryCheckers {
				fmt.Printf("     • %s (%s)\n", checker.Name(), checker.ID())
			}
		}
	}
	fmt.Println()

	// Run a sample of checkers
	fmt.Println("   🧪 Sample Checker Execution:")
	sampleCheckers := []string{"git-status", "license-check", "ci-config"}

	for _, checkerID := range sampleCheckers {
		fmt.Printf("     Testing %s... ", checkerID)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		start := time.Now()

		result, err := checkerRegistry.RunChecker(ctx, checkerID, repoCtx)
		duration := time.Since(start)

		cancel()

		if err != nil {
			fmt.Printf("❌ Failed (%v)\n", duration)
			continue
		}

		statusIcon := getStatusIcon(result.Status)
		fmt.Printf("%s %s (%v) Score: %d/100\n",
			statusIcon, result.Status, duration, result.Score)
	}
	fmt.Println()

	// === INTEGRATION DEMO ===
	fmt.Println("🔗 Integration Demonstration")
	fmt.Println("   🎯 Running comprehensive health check...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create config that enables safe checkers (no external dependencies)
	config := make(map[string]core.CheckerConfig)
	safeCheckers := []string{"git-status", "git-last-commit", "license-check", "ci-config"}

	for _, checkerID := range safeCheckers {
		config[checkerID] = core.CheckerConfig{
			Enabled:  true,
			Severity: "medium",
			Timeout:  10 * time.Second,
		}
	}

	results, err := checkerRegistry.RunAllEnabledCheckers(ctx, repoCtx, config)
	if err != nil {
		log.Printf("   ⚠️  Some checks failed: %v", err)
	}

	// Calculate overall health score
	totalScore := 0
	maxPossible := 0
	statusCounts := make(map[core.Status]int)

	for _, result := range results {
		totalScore += result.Score
		maxPossible += result.MaxScore
		statusCounts[result.Status]++
	}

	overallScore := 0
	if maxPossible > 0 {
		overallScore = (totalScore * 100) / maxPossible
	}

	fmt.Printf("   📊 Overall Health Score: %d/100\n", overallScore)
	fmt.Printf("   📈 Status Summary: ")
	for status, count := range statusCounts {
		fmt.Printf("%s:%d ", status, count)
	}
	fmt.Println()
	fmt.Printf("   ⏱️  Total execution time: %d checks completed\n", len(results))
	fmt.Println()

	// === SUMMARY ===
	fmt.Println("🎉 MIGRATION SUCCESS SUMMARY")
	fmt.Println("   ✅ Phase 1: Foundation architecture established")
	fmt.Printf("   ✅ Phase 2: %d language analyzers implemented\n", len(analyzerRegistry.GetAnalyzers()))
	fmt.Printf("   ✅ Phase 3: %d modular checkers implemented\n", stats.TotalCheckers)
	fmt.Printf("   📊 Architecture health: %d%% (%d/%d)\n", overallScore, totalScore, maxPossible)
	fmt.Println()
	fmt.Println("🚧 Next Steps:")
	fmt.Println("   • Phase 4: Enhanced orchestration engine")
	fmt.Println("   • Phase 5: Complete migration & optimization")
	fmt.Println()
	fmt.Println("✨ Modular architecture migration is on track!")
}

func formatCategories(categories map[string]int) string {
	var parts []string
	for category, count := range categories {
		parts = append(parts, fmt.Sprintf("%s:%d", category, count))
	}
	return strings.Join(parts, " ")
}

func getStatusIcon(status core.Status) string {
	switch status {
	case core.StatusHealthy:
		return "✅"
	case core.StatusWarning:
		return "⚠️"
	case core.StatusCritical:
		return "❌"
	default:
		return "❓"
	}
}
