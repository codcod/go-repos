package base

import (
	"context"
	"time"

	"github.com/codcod/repos/internal/core"
)

// BaseChecker provides common functionality for all checkers
type BaseChecker struct {
	id       string
	name     string
	category string
	config   core.CheckerConfig
}

// NewBaseChecker creates a new base checker
func NewBaseChecker(id, name, category string, config core.CheckerConfig) *BaseChecker {
	return &BaseChecker{
		id:       id,
		name:     name,
		category: category,
		config:   config,
	}
}

// ID returns the checker ID
func (c *BaseChecker) ID() string {
	return c.id
}

// Name returns the checker name
func (c *BaseChecker) Name() string {
	return c.name
}

// Category returns the checker category
func (c *BaseChecker) Category() string {
	return c.category
}

// Config returns the checker configuration
func (c *BaseChecker) Config() core.CheckerConfig {
	return c.config
}

// Execute executes a check function with common error handling and timing
func (c *BaseChecker) Execute(ctx context.Context, repoCtx core.RepositoryContext, checkFn func() (core.CheckResult, error)) (core.CheckResult, error) {
	start := time.Now()

	// Note: Timeout is not currently supported as checkFn doesn't accept context
	// TODO: Update checkFn signature to accept context for timeout support

	result, err := checkFn()
	if err != nil {
		return core.CheckResult{
			ID:         c.id,
			Name:       c.name,
			Category:   c.category,
			Status:     core.StatusCritical,
			Duration:   time.Since(start),
			Timestamp:  time.Now(),
			Repository: repoCtx.Repository.Name,
			Issues: []core.Issue{
				{
					Type:     "execution_error",
					Severity: core.SeverityCritical,
					Message:  err.Error(),
				},
			},
		}, nil // Return nil error as we've handled it in the result
	}

	// Fill in common fields
	result.ID = c.id
	result.Name = c.name
	result.Category = c.category
	result.Duration = time.Since(start)
	result.Timestamp = time.Now()
	result.Repository = repoCtx.Repository.Name

	return result, nil
}

// SupportsRepository checks if this checker supports the given repository
func (c *BaseChecker) SupportsRepository(repo core.Repository) bool {
	// Default implementation - can be overridden
	return true
}

// ResultBuilder helps build check results
type ResultBuilder struct {
	result core.CheckResult
}

// NewResultBuilder creates a new result builder
func NewResultBuilder(id, name, category string) *ResultBuilder {
	return &ResultBuilder{
		result: core.CheckResult{
			ID:       id,
			Name:     name,
			Category: category,
			Status:   core.StatusHealthy,
			Score:    100,
			MaxScore: 100,
			Issues:   make([]core.Issue, 0),
			Warnings: make([]core.Warning, 0),
			Metrics:  make(map[string]interface{}),
			Metadata: make(map[string]string),
		},
	}
}

// WithStatus sets the status
func (b *ResultBuilder) WithStatus(status core.HealthStatus) *ResultBuilder {
	b.result.Status = status
	return b
}

// WithScore sets the score
func (b *ResultBuilder) WithScore(score, maxScore int) *ResultBuilder {
	b.result.Score = score
	b.result.MaxScore = maxScore
	return b
}

// AddIssue adds an issue
func (b *ResultBuilder) AddIssue(issue core.Issue) *ResultBuilder {
	b.result.Issues = append(b.result.Issues, issue)

	// Adjust status based on severity
	switch issue.Severity {
	case core.SeverityCritical, core.SeverityHigh:
		b.result.Status = core.StatusCritical
	case core.SeverityMedium:
		if b.result.Status == core.StatusHealthy {
			b.result.Status = core.StatusWarning
		}
	}

	return b
}

// AddWarning adds a warning
func (b *ResultBuilder) AddWarning(warning core.Warning) *ResultBuilder {
	b.result.Warnings = append(b.result.Warnings, warning)

	// Set to warning if currently healthy
	if b.result.Status == core.StatusHealthy {
		b.result.Status = core.StatusWarning
	}

	return b
}

// AddMetric adds a metric
func (b *ResultBuilder) AddMetric(key string, value interface{}) *ResultBuilder {
	b.result.Metrics[key] = value
	return b
}

// AddMetadata adds metadata
func (b *ResultBuilder) AddMetadata(key, value string) *ResultBuilder {
	b.result.Metadata[key] = value
	return b
}

// Build returns the final result
func (b *ResultBuilder) Build() core.CheckResult {
	return b.result
}

// NewIssue creates a new issue
func NewIssue(issueType string, severity core.Severity, message string) core.Issue {
	return core.Issue{
		Type:     issueType,
		Severity: severity,
		Message:  message,
		Context:  make(map[string]interface{}),
	}
}

// NewIssueWithLocation creates a new issue with location
func NewIssueWithLocation(issueType string, severity core.Severity, message, file string, line, column int) core.Issue {
	return core.Issue{
		Type:     issueType,
		Severity: severity,
		Message:  message,
		Location: &core.Location{
			File:   file,
			Line:   line,
			Column: column,
		},
		Context: make(map[string]interface{}),
	}
}

// NewIssueWithSuggestion creates a new issue with suggestion
func NewIssueWithSuggestion(issueType string, severity core.Severity, message, suggestion string) core.Issue {
	return core.Issue{
		Type:       issueType,
		Severity:   severity,
		Message:    message,
		Suggestion: suggestion,
		Context:    make(map[string]interface{}),
	}
}
