// Package core provides result building utilities for health checks
package core

import (
	"time"

	"github.com/codcod/repos/internal/core"
)

// ResultBuilder provides a fluent interface for building repository check results
type ResultBuilder struct {
	repository core.Repository
	startTime  time.Time
	results    []core.CheckResult
	metadata   map[string]interface{}
}

// NewResultBuilder creates a new result builder for the specified repository
func NewResultBuilder(repository core.Repository) *ResultBuilder {
	return &ResultBuilder{
		repository: repository,
		startTime:  time.Now(),
		results:    make([]core.CheckResult, 0),
		metadata:   make(map[string]interface{}),
	}
}

// AddCheckResult adds a check result to the builder
func (rb *ResultBuilder) AddCheckResult(result core.CheckResult) *ResultBuilder {
	rb.results = append(rb.results, result)
	return rb
}

// AddCheckResults adds multiple check results to the builder
func (rb *ResultBuilder) AddCheckResults(results []core.CheckResult) *ResultBuilder {
	rb.results = append(rb.results, results...)
	return rb
}

// AddWarning adds a warning to a check result
func (rb *ResultBuilder) AddWarning(checkerName, message string, location *core.Location) *ResultBuilder {
	warning := core.Warning{
		Type:     "warning",
		Message:  message,
		Location: location,
	}

	// Find existing result or create new one
	for i, result := range rb.results {
		if result.Name == checkerName {
			rb.results[i].Warnings = append(rb.results[i].Warnings, warning)
			return rb
		}
	}

	// Create new result with warning
	result := core.CheckResult{
		Name:      checkerName,
		Category:  "general",
		Status:    core.StatusWarning,
		Warnings:  []core.Warning{warning},
		Timestamp: time.Now(),
	}

	return rb.AddCheckResult(result)
}

// AddIssue adds an issue to a check result
func (rb *ResultBuilder) AddIssue(checkerName, issueType, message string, severity core.Severity, location *core.Location) *ResultBuilder {
	issue := core.Issue{
		Type:     issueType,
		Message:  message,
		Severity: severity,
		Location: location,
	}

	// Find existing result or create new one
	for i, result := range rb.results {
		if result.Name == checkerName {
			rb.results[i].Issues = append(rb.results[i].Issues, issue)
			// Update status based on severity
			if severity == core.SeverityHigh || severity == core.SeverityCritical {
				rb.results[i].Status = core.StatusCritical
			}
			return rb
		}
	}

	// Determine status based on severity
	status := core.StatusWarning
	if severity == core.SeverityHigh || severity == core.SeverityCritical {
		status = core.StatusCritical
	}

	// Create new result with issue
	result := core.CheckResult{
		Name:      checkerName,
		Category:  "general",
		Status:    status,
		Issues:    []core.Issue{issue},
		Timestamp: time.Now(),
	}

	return rb.AddCheckResult(result)
}

// AddSuccessResult adds a successful check result
func (rb *ResultBuilder) AddSuccessResult(checkerName, category string) *ResultBuilder {
	result := core.CheckResult{
		Name:      checkerName,
		Category:  category,
		Status:    core.StatusHealthy,
		Score:     100,
		MaxScore:  100,
		Issues:    []core.Issue{},
		Warnings:  []core.Warning{},
		Duration:  time.Since(rb.startTime),
		Timestamp: time.Now(),
	}

	return rb.AddCheckResult(result)
}

// WithMetadata adds metadata to the result
func (rb *ResultBuilder) WithMetadata(key string, value interface{}) *ResultBuilder {
	rb.metadata[key] = value
	return rb
}

// Build creates the final RepositoryResult
func (rb *ResultBuilder) Build() core.RepositoryResult {
	endTime := time.Now()
	return core.RepositoryResult{
		Repository:   rb.repository,
		CheckResults: rb.results,
		Status:       rb.calculateOverallStatus(),
		Score:        rb.calculateScore(),
		MaxScore:     rb.calculateMaxScore(),
		StartTime:    rb.startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(rb.startTime),
	}
}

// calculateOverallStatus determines the overall status based on check results
func (rb *ResultBuilder) calculateOverallStatus() core.HealthStatus {
	if len(rb.results) == 0 {
		return core.StatusUnknown
	}

	hasCritical := false
	hasWarnings := false

	for _, result := range rb.results {
		switch result.Status {
		case core.StatusCritical:
			hasCritical = true
		case core.StatusWarning:
			hasWarnings = true
		}
	}

	if hasCritical {
		return core.StatusCritical
	}
	if hasWarnings {
		return core.StatusWarning
	}

	return core.StatusHealthy
}

// calculateScore calculates the overall score based on check results
func (rb *ResultBuilder) calculateScore() int {
	if len(rb.results) == 0 {
		return 0
	}

	totalScore := 0
	for _, result := range rb.results {
		totalScore += result.Score
	}

	return totalScore / len(rb.results)
}

// calculateMaxScore calculates the maximum possible score
func (rb *ResultBuilder) calculateMaxScore() int {
	if len(rb.results) == 0 {
		return 0
	}

	totalMaxScore := 0
	for _, result := range rb.results {
		totalMaxScore += result.MaxScore
	}

	return totalMaxScore / len(rb.results)
}

// GetResultCount returns the number of check results
func (rb *ResultBuilder) GetResultCount() int {
	return len(rb.results)
}

// GetWarningCount returns the number of warnings across all results
func (rb *ResultBuilder) GetWarningCount() int {
	count := 0
	for _, result := range rb.results {
		count += len(result.Warnings)
	}
	return count
}

// HasFailures returns true if any check results indicate failure
func (rb *ResultBuilder) HasFailures() bool {
	for _, result := range rb.results {
		if result.Status == core.StatusCritical {
			return true
		}
	}
	return false
}
