package health

import (
	"fmt"
	"time"

	"github.com/codcod/repos/internal/config"
)

// CyclomaticComplexityChecker analyzes code complexity using cyclomatic complexity metrics
type CyclomaticComplexityChecker struct {
	analyzer *ComplexityAnalyzer
	config   *ComplexityConfig
	logger   Logger
	registry *AnalyzerRegistry
}

// NewCyclomaticComplexityChecker creates a new cyclomatic complexity checker
func NewCyclomaticComplexityChecker(registry *AnalyzerRegistry, config ComplexityConfig, logger Logger) *CyclomaticComplexityChecker {
	if logger == nil {
		logger = &NoOpLogger{}
	}

	return &CyclomaticComplexityChecker{
		analyzer: NewComplexityAnalyzer(),
		config:   &config,
		logger:   NewCheckerLogger("cyclomatic_complexity", logger),
		registry: registry,
	}
}

func (c *CyclomaticComplexityChecker) Name() string {
	return "Cyclomatic Complexity"
}

func (c *CyclomaticComplexityChecker) Category() string {
	return "code-quality"
}

func (c *CyclomaticComplexityChecker) Check(repo config.Repository) HealthCheck {
	start := time.Now()

	opLogger, done := c.logger.(*CheckerLogger).StartOperation("analyze_repository")
	defer done()

	repoPath := GetRepoPath(repo)
	opLogger.Info("starting complexity analysis", String("repo_path", repoPath))

	// Use the new result builder for structured results
	builder := NewResultBuilder(c.Name(), c.Category())

	metrics := c.analyzeRepository(repoPath, opLogger)

	c.evaluateMetrics(metrics, builder, opLogger)

	result := builder.WithDuration(time.Since(start)).Build()
	opLogger.Info("analysis completed",
		String("status", string(result.Status)),
		Int("score", result.Score),
		Int("issues_count", len(result.Issues)))

	return c.convertToHealthCheck(result)
}

func (c *CyclomaticComplexityChecker) analyzeRepository(repoPath string, logger Logger) ComplexityMetrics {
	if c.analyzer == nil {
		c.analyzer = NewComplexityAnalyzer()
	}

	logger.Debug("analyzing repository complexity")

	metrics := c.analyzer.AnalyzeRepository(repoPath)

	logger.Info("complexity analysis completed",
		Int("total_files", metrics.totalFiles),
		String("average_complexity", fmt.Sprintf("%.2f", metrics.GetAverageComplexity())))

	return metrics
}

func (c *CyclomaticComplexityChecker) evaluateMetrics(metrics ComplexityMetrics, builder *ResultBuilder, logger Logger) {
	avgComplexity := metrics.GetAverageComplexity()
	threshold := float64(c.config.DefaultThreshold)

	// Add metrics
	builder.AddMetric("average_complexity", avgComplexity).
		AddMetric("max_complexity", metrics.maxComplexity).
		AddMetric("total_files", metrics.totalFiles).
		AddMetric("threshold", threshold)

	// Add metadata
	builder.AddMetadata("analyzer_version", "1.0").
		AddMetadata("threshold_source", "config")

	// Evaluate overall health
	if avgComplexity <= threshold/2 {
		builder.WithStatus(HealthStatusHealthy)
		logger.Debug("complexity within excellent range",
			String("average", fmt.Sprintf("%.2f", avgComplexity)),
			String("threshold", fmt.Sprintf("%.0f", threshold/2)))
	} else if avgComplexity <= threshold {
		builder.WithStatus(HealthStatusWarning).
			AddWarning(NewWarning("Code complexity is moderate - consider refactoring high-complexity functions").
				WithSuggestion("Focus on functions with complexity > " + fmt.Sprintf("%.0f", threshold)))
		logger.Info("complexity in warning range",
			String("average", fmt.Sprintf("%.2f", avgComplexity)),
			String("threshold", fmt.Sprintf("%.0f", threshold)))
	} else {
		builder.WithStatus(HealthStatusCritical).
			AddIssue(NewIssue("high_complexity", SeverityCritical,
				fmt.Sprintf("Average complexity (%.2f) exceeds threshold (%.0f)", avgComplexity, threshold)).
				WithSuggestion("Refactor complex functions to improve maintainability"))
		logger.Warn("complexity exceeds threshold",
			String("average", fmt.Sprintf("%.2f", avgComplexity)),
			String("threshold", fmt.Sprintf("%.0f", threshold)))
	}

	// Add issues for specific high-complexity files/functions if available
	if c.config.DetailedReport {
		c.addDetailedIssues(metrics, builder, logger)
	}
}

func (c *CyclomaticComplexityChecker) addDetailedIssues(metrics ComplexityMetrics, builder *ResultBuilder, logger Logger) {
	// This would be implemented to add specific file/function issues
	// For now, we'll add a placeholder implementation
	logger.Debug("adding detailed complexity issues")

	// Add warnings for files with high complexity if we had that data
	if metrics.highComplexityFiles > 0 {
		builder.AddWarning(NewWarning(
			fmt.Sprintf("%d files have high complexity", metrics.highComplexityFiles)).
			WithSuggestion("Consider breaking down complex functions in these files"))
	}
}

// convertToHealthCheck converts the new CheckResult to the legacy HealthCheck format
func (c *CyclomaticComplexityChecker) convertToHealthCheck(result CheckResult) HealthCheck {
	message := c.formatMessageFromResult(result)
	details := c.formatDetailsFromResult(result)
	severity := c.calculateSeverityFromResult(result)

	return createHealthCheck(
		result.Name,
		result.Category,
		result.Status,
		message,
		details,
		severity,
	)
}

func (c *CyclomaticComplexityChecker) formatMessageFromResult(result CheckResult) string {
	if len(result.Issues) > 0 {
		return result.Issues[0].Message
	}
	if len(result.Warnings) > 0 {
		return result.Warnings[0].Message
	}
	return "Code complexity is within acceptable limits"
}

func (c *CyclomaticComplexityChecker) formatDetailsFromResult(result CheckResult) string {
	details := fmt.Sprintf("Score: %d/100\n", result.Score)

	if avgComplexity, ok := result.Metrics["average_complexity"]; ok {
		details += fmt.Sprintf("Average Complexity: %.2f\n", avgComplexity)
	}
	if maxComplexity, ok := result.Metrics["max_complexity"]; ok {
		details += fmt.Sprintf("Maximum Complexity: %v\n", maxComplexity)
	}
	if totalFiles, ok := result.Metrics["total_files"]; ok {
		details += fmt.Sprintf("Files Analyzed: %v\n", totalFiles)
	}

	return details
}

func (c *CyclomaticComplexityChecker) calculateSeverityFromResult(result CheckResult) int {
	switch result.Status {
	case HealthStatusHealthy:
		return 0
	case HealthStatusWarning:
		return 1
	case HealthStatusCritical:
		return 2
	default:
		return 1
	}
}
