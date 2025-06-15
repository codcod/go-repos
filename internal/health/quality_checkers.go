package health

import (
	"github.com/codcod/repos/internal/config"
)

// CyclomaticComplexityChecker analyzes code complexity using cyclomatic complexity metrics
type CyclomaticComplexityChecker struct {
	analyzer *ComplexityAnalyzer
}

func (c *CyclomaticComplexityChecker) Name() string {
	return "Cyclomatic Complexity"
}

func (c *CyclomaticComplexityChecker) Category() string {
	return "code-quality"
}

func (c *CyclomaticComplexityChecker) Check(repo config.Repository) HealthCheck {
	if c.analyzer == nil {
		c.analyzer = NewComplexityAnalyzer()
	}

	repoPath := GetRepoPath(repo)
	metrics := c.analyzer.AnalyzeRepository(repoPath)

	return createHealthCheck(
		c.Name(),
		"quality",
		c.determineHealthStatus(metrics),
		c.formatMessage(metrics),
		c.formatDetails(metrics),
		c.calculateSeverity(metrics),
	)
}

func (c *CyclomaticComplexityChecker) determineHealthStatus(metrics ComplexityMetrics) HealthStatus {
	avgComplexity := metrics.GetAverageComplexity()
	if avgComplexity <= 5.0 {
		return HealthStatusHealthy
	} else if avgComplexity <= 10.0 {
		return HealthStatusWarning
	}
	return HealthStatusCritical
}

func (c *CyclomaticComplexityChecker) formatMessage(metrics ComplexityMetrics) string {
	avgComplexity := metrics.GetAverageComplexity()
	if avgComplexity <= 5.0 {
		return "Code complexity is within acceptable limits"
	} else if avgComplexity <= 10.0 {
		return "Code complexity is moderate - consider refactoring"
	}
	return "Code complexity is high - refactoring recommended"
}

func (c *CyclomaticComplexityChecker) formatDetails(metrics ComplexityMetrics) string {
	return metrics.FormatSummary()
}

func (c *CyclomaticComplexityChecker) calculateSeverity(metrics ComplexityMetrics) int {
	avgComplexity := metrics.GetAverageComplexity()
	if avgComplexity <= 5.0 {
		return 0
	} else if avgComplexity <= 10.0 {
		return 1
	}
	return 2
}
