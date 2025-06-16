package orchestration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codcod/repos/internal/core"
)

// Engine orchestrates the execution of health checks across repositories
type Engine struct {
	checkerRegistry  core.CheckerRegistry
	analyzerRegistry core.AnalyzerRegistry
	config           core.Config
	logger           core.Logger
	maxConcurrency   int
	timeout          time.Duration
}

// NewEngine creates a new orchestration engine
func NewEngine(
	checkerRegistry core.CheckerRegistry,
	analyzerRegistry core.AnalyzerRegistry,
	config core.Config,
	logger core.Logger,
) *Engine {
	engineConfig := config.GetEngineConfig()

	return &Engine{
		checkerRegistry:  checkerRegistry,
		analyzerRegistry: analyzerRegistry,
		config:           config,
		logger:           logger,
		maxConcurrency:   engineConfig.MaxConcurrency,
		timeout:          engineConfig.Timeout,
	}
}

// ExecuteHealthCheck runs a complete health check workflow for repositories
func (e *Engine) ExecuteHealthCheck(ctx context.Context, repos []core.Repository) (*core.WorkflowResult, error) {
	e.logger.Info("Starting health check workflow",
		core.Int("repository_count", len(repos)),
		core.Int("max_concurrency", e.maxConcurrency))

	startTime := time.Now()

	// Create workflow context with timeout
	workflowCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Execute checks for all repositories
	repoResults, err := e.executeRepositoryChecks(workflowCtx, repos)
	if err != nil {
		return nil, fmt.Errorf("failed to execute repository checks: %w", err)
	}

	// Aggregate results
	workflowResult := &core.WorkflowResult{
		StartTime:         startTime,
		EndTime:           time.Now(),
		Duration:          time.Since(startTime),
		TotalRepos:        len(repos),
		RepositoryResults: repoResults,
		Summary:           e.generateSummary(repoResults),
	}

	e.logger.Info("Health check workflow completed",
		core.Duration("duration", workflowResult.Duration),
		core.Int("total_repos", workflowResult.TotalRepos),
		core.Int("successful_repos", workflowResult.Summary.SuccessfulRepos))

	return workflowResult, nil
}

// executeRepositoryChecks runs checks for all repositories with concurrency control
//
//nolint:unparam // error return kept for future extensibility
func (e *Engine) executeRepositoryChecks(ctx context.Context, repos []core.Repository) ([]core.RepositoryResult, error) {
	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, e.maxConcurrency)

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]core.RepositoryResult, len(repos))

	// Process repositories concurrently
	for i, repo := range repos {
		wg.Add(1)

		go func(index int, repository core.Repository) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := e.executeRepositoryCheck(ctx, repository)

			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(i, repo)
	}

	wg.Wait()

	return results, nil // No errors in current implementation
}

// executeRepositoryCheck runs all checks for a single repository
func (e *Engine) executeRepositoryCheck(ctx context.Context, repo core.Repository) core.RepositoryResult {
	e.logger.Debug("Starting repository check", core.String("repository", repo.Name))

	startTime := time.Now()
	result := core.RepositoryResult{
		Repository: repo,
		StartTime:  startTime,
		Status:     core.StatusHealthy,
	}

	// Create repository context
	repoCtx := core.RepositoryContext{
		Repository: repo,
		Config:     e.config,
		// FileSystem and Cache would be injected from platforms
	}

	// Run analysis if language is detected
	if repo.Language != "" {
		analysisResult, err := e.runAnalysis(ctx, repoCtx)
		if err != nil {
			e.logger.Warn("Analysis failed",
				core.String("repository", repo.Name),
				core.Error("error", err))
		} else {
			result.AnalysisResult = analysisResult
		}
	}

	// Get enabled checkers for this repository
	checkerConfigs := e.getCheckerConfigs()
	checkResults, err := e.runCheckers(ctx, repoCtx, checkerConfigs)
	if err != nil {
		e.logger.Error("Checker execution failed",
			core.String("repository", repo.Name),
			core.Error("error", err))
		result.Status = core.StatusCritical
		result.Error = err.Error()
	} else {
		result.CheckResults = checkResults
		result.Status = e.calculateOverallStatus(checkResults)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)
	result.Score = e.calculateScore(checkResults)

	e.logger.Debug("Repository check completed",
		core.String("repository", repo.Name),
		core.String("status", string(result.Status)),
		core.Int("score", result.Score),
		core.Duration("duration", result.Duration))

	return result
}

// runAnalysis executes language-specific analysis
func (e *Engine) runAnalysis(ctx context.Context, repoCtx core.RepositoryContext) (*core.AnalysisResult, error) {
	analyzer, err := e.analyzerRegistry.GetAnalyzer(repoCtx.Repository.Language)
	if err != nil {
		return nil, fmt.Errorf("analyzer not found for language %s: %w", repoCtx.Repository.Language, err)
	}

	return analyzer.Analyze(ctx, repoCtx.Repository.Path, core.AnalyzerConfig{
		Enabled:           true,
		ComplexityEnabled: true,
		FunctionLevel:     true,
	})
}

// runCheckers executes all enabled checkers for a repository
//
//nolint:unparam // error return kept for future extensibility
func (e *Engine) runCheckers(ctx context.Context, repoCtx core.RepositoryContext, checkerConfigs map[string]core.CheckerConfig) ([]core.CheckResult, error) {
	// This would use the registry's RunAllEnabledCheckers method
	// For now, we'll implement a simple version

	enabledCheckers := e.getEnabledCheckers(repoCtx.Repository, checkerConfigs)
	results := make([]core.CheckResult, 0, len(enabledCheckers))

	for _, checker := range enabledCheckers {
		result, err := checker.Check(ctx, repoCtx)
		if err != nil {
			e.logger.Warn("Checker failed",
				core.String("checker", checker.ID()),
				core.String("repository", repoCtx.Repository.Name),
				core.Error("error", err))

			// Create error result
			result = core.CheckResult{
				ID:         checker.ID(),
				Name:       checker.Name(),
				Category:   checker.Category(),
				Status:     core.StatusCritical,
				Repository: repoCtx.Repository.Name,
				Timestamp:  time.Now(),
				Issues: []core.Issue{{
					Type:     "execution_error",
					Severity: core.SeverityCritical,
					Message:  fmt.Sprintf("Checker execution failed: %v", err),
				}},
			}
		}

		results = append(results, result)
	}

	return results, nil // No errors in current implementation
}

// getEnabledCheckers returns checkers that are enabled and support the repository
func (e *Engine) getEnabledCheckers(repo core.Repository, checkerConfigs map[string]core.CheckerConfig) []core.Checker {
	allCheckers := e.checkerRegistry.GetCheckers()
	var enabledCheckers []core.Checker

	for _, checker := range allCheckers {
		if !checker.SupportsRepository(repo) {
			continue
		}

		config, exists := checkerConfigs[checker.ID()]
		if !exists {
			// Use default config if not specified
			config = checker.Config()
		}

		if config.Enabled {
			enabledCheckers = append(enabledCheckers, checker)
		}
	}

	return enabledCheckers
}

// getCheckerConfigs retrieves checker configurations
func (e *Engine) getCheckerConfigs() map[string]core.CheckerConfig {
	// Get all registered checkers and enable them with default config
	allCheckers := e.checkerRegistry.GetCheckers()
	configs := make(map[string]core.CheckerConfig)

	for _, checker := range allCheckers {
		// Get the checker's default config and enable it
		defaultConfig := checker.Config()
		defaultConfig.Enabled = true
		configs[checker.ID()] = defaultConfig
	}

	return configs
}

// calculateOverallStatus determines the overall status based on check results
func (e *Engine) calculateOverallStatus(results []core.CheckResult) core.HealthStatus {
	if len(results) == 0 {
		return core.StatusUnknown
	}

	hasCritical := false
	hasWarning := false

	for _, result := range results {
		switch result.Status {
		case core.StatusCritical:
			hasCritical = true
		case core.StatusWarning:
			hasWarning = true
		}
	}

	if hasCritical {
		return core.StatusCritical
	}
	if hasWarning {
		return core.StatusWarning
	}

	return core.StatusHealthy
}

// calculateScore calculates an overall score based on check results
func (e *Engine) calculateScore(results []core.CheckResult) int {
	if len(results) == 0 {
		return 0
	}

	totalScore := 0
	totalMaxScore := 0

	for _, result := range results {
		totalScore += result.Score
		totalMaxScore += result.MaxScore
	}

	if totalMaxScore == 0 {
		return 0
	}

	return (totalScore * 100) / totalMaxScore
}

// generateSummary creates a summary of workflow results
func (e *Engine) generateSummary(results []core.RepositoryResult) core.WorkflowSummary {
	summary := core.WorkflowSummary{
		StatusCounts:   make(map[core.HealthStatus]int),
		SeverityCounts: make(map[core.Severity]int),
	}

	totalScore := 0
	totalRepos := len(results)

	for _, result := range results {
		summary.StatusCounts[result.Status]++
		totalScore += result.Score

		if result.Status == core.StatusHealthy || result.Status == core.StatusWarning {
			summary.SuccessfulRepos++
		} else {
			summary.FailedRepos++
		}

		// Aggregate issues
		for _, checkResult := range result.CheckResults {
			summary.TotalIssues += len(checkResult.Issues)
			for _, issue := range checkResult.Issues {
				summary.SeverityCounts[issue.Severity]++
			}
		}
	}

	if totalRepos > 0 {
		summary.AverageScore = totalScore / totalRepos
	}

	return summary
}
