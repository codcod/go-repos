package complexity

import (
	"context"
	"fmt"

	"github.com/codcod/repos/internal/analyzers/registry"
	"github.com/codcod/repos/internal/checkers/base"
	"github.com/codcod/repos/internal/core"
)

// ComplexityConfig represents configuration for complexity checking
type ComplexityConfig struct {
	core.CheckerConfig
	Thresholds     map[string]int `json:"thresholds" yaml:"thresholds"`
	ReportLevel    string         `json:"report_level" yaml:"report_level"`
	IncludeTests   bool           `json:"include_tests" yaml:"include_tests"`
	DetailedReport bool           `json:"detailed_report" yaml:"detailed_report"`
}

// ComplexityChecker checks cyclomatic complexity
type ComplexityChecker struct {
	*base.BaseChecker
	registry   *registry.Registry
	thresholds map[string]int
	config     ComplexityConfig
}

// NewComplexityChecker creates a new complexity checker
func NewComplexityChecker(analyzerRegistry *registry.Registry, config ComplexityConfig) *ComplexityChecker {
	return &ComplexityChecker{
		BaseChecker: base.NewBaseChecker(
			"cyclomatic-complexity",
			"Cyclomatic Complexity",
			"quality",
			config.CheckerConfig,
		),
		registry:   analyzerRegistry,
		thresholds: config.Thresholds,
		config:     config,
	}
}

// Check performs complexity analysis
func (c *ComplexityChecker) Check(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	return c.Execute(ctx, repoCtx, func() (core.CheckResult, error) {
		return c.analyzeComplexity(ctx, repoCtx)
	})
}

// analyzeComplexity performs the actual complexity analysis
func (c *ComplexityChecker) analyzeComplexity(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	builder := base.NewResultBuilder(c.ID(), c.Name(), c.Category())

	// Get supported analyzers for this repository
	analyzers := c.registry.GetSupportedAnalyzers(repoCtx.Repository)

	if len(analyzers) == 0 {
		builder.WithStatus(core.StatusWarning)
		builder.AddWarning(core.Warning{
			Type:    "no_analyzers",
			Message: "No supported analyzers found for this repository",
		})
		return builder.Build(), nil
	}

	totalScore := 0
	maxScore := 0
	overallComplexity := 0.0
	totalFunctions := 0

	for _, analyzer := range analyzers {
		if !analyzer.SupportsComplexity() {
			continue
		}

		repoCtx.Logger.Debug("analyzing complexity",
			core.String("language", analyzer.Language()),
			core.String("repository", repoCtx.Repository.Name))

		result, err := analyzer.AnalyzeComplexity(ctx, repoCtx.Repository.Path)
		if err != nil {
			builder.AddWarning(core.Warning{
				Type:    "analysis_error",
				Message: fmt.Sprintf("Failed to analyze %s files: %v", analyzer.Language(), err),
			})
			continue
		}

		// Calculate score based on complexity
		threshold := c.getThreshold(analyzer.Language())
		score := c.calculateScore(result, threshold)

		totalScore += score
		maxScore += 100
		overallComplexity += result.AverageComplexity
		totalFunctions += result.TotalFunctions

		// Add metrics
		builder.AddMetric(fmt.Sprintf("%s_average_complexity", analyzer.Language()), result.AverageComplexity)
		builder.AddMetric(fmt.Sprintf("%s_max_complexity", analyzer.Language()), result.MaxComplexity)
		builder.AddMetric(fmt.Sprintf("%s_total_functions", analyzer.Language()), result.TotalFunctions)
		builder.AddMetric(fmt.Sprintf("%s_total_files", analyzer.Language()), result.TotalFiles)

		// Check for high complexity functions
		c.checkFunctionComplexity(builder, analyzer.Language(), result.Functions, threshold)
	}

	if maxScore > 0 {
		avgScore := totalScore / len(analyzers)
		builder.WithScore(avgScore, 100)

		// Set overall status based on score
		if avgScore >= 80 {
			builder.WithStatus(core.StatusHealthy)
		} else if avgScore >= 60 {
			builder.WithStatus(core.StatusWarning)
		} else {
			builder.WithStatus(core.StatusCritical)
		}
	}

	// Add overall metrics
	if len(analyzers) > 0 {
		builder.AddMetric("overall_average_complexity", overallComplexity/float64(len(analyzers)))
		builder.AddMetric("total_functions_analyzed", totalFunctions)
		builder.AddMetric("languages_analyzed", len(analyzers))
	}

	builder.AddMetadata("analysis_type", "cyclomatic_complexity")
	builder.AddMetadata("detailed_report", fmt.Sprintf("%t", c.config.DetailedReport))

	return builder.Build(), nil
}

// getThreshold gets the complexity threshold for a language
func (c *ComplexityChecker) getThreshold(language string) int {
	if threshold, exists := c.thresholds[language]; exists {
		return threshold
	}
	// Default threshold
	return 10
}

// calculateScore calculates a score based on complexity results
func (c *ComplexityChecker) calculateScore(result core.ComplexityResult, threshold int) int {
	if result.TotalFunctions == 0 {
		return 100
	}

	highComplexityCount := 0
	for _, fn := range result.Functions {
		if fn.Complexity > threshold {
			highComplexityCount++
		}
	}

	// Score based on percentage of functions under threshold
	percentage := float64(result.TotalFunctions-highComplexityCount) / float64(result.TotalFunctions)
	return int(percentage * 100)
}

// checkFunctionComplexity checks individual function complexity
func (c *ComplexityChecker) checkFunctionComplexity(builder *base.ResultBuilder, language string, functions []core.FunctionComplexity, threshold int) {
	for _, fn := range functions {
		if fn.Complexity > threshold {
			severity := c.getSeverity(fn.Complexity, threshold)

			issue := base.NewIssueWithLocation(
				"high_complexity",
				severity,
				fmt.Sprintf("Function '%s' has high cyclomatic complexity: %d (threshold: %d)", fn.Name, fn.Complexity, threshold),
				fn.File,
				fn.Line,
				0,
			)

			issue = base.NewIssueWithSuggestion(
				issue.Type,
				issue.Severity,
				issue.Message,
				fmt.Sprintf("Consider refactoring '%s' to reduce complexity by extracting methods or simplifying logic", fn.Name),
			)

			builder.AddIssue(issue)
		}
	}
}

// getSeverity determines severity based on complexity
func (c *ComplexityChecker) getSeverity(complexity, threshold int) core.Severity {
	if complexity > threshold*2 {
		return core.SeverityCritical
	} else if complexity > threshold+threshold/2 { // threshold * 1.5
		return core.SeverityHigh
	} else {
		return core.SeverityMedium
	}
}

// SupportsRepository checks if this checker supports the repository
func (c *ComplexityChecker) SupportsRepository(repo core.Repository) bool {
	analyzers := c.registry.GetSupportedAnalyzers(repo)
	for _, analyzer := range analyzers {
		if analyzer.SupportsComplexity() {
			return true
		}
	}
	return false
}
