// Package health provides repository health checking functionality
// for the repos tool, including dependency analysis, security scanning,
// and overall repository maintenance status.
package health

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/codcod/repos/internal/config"
)

// Global timeout context for health checks
var (
	defaultTimeout = 30 * time.Second
	healthTimeout  = defaultTimeout
)

// SetHealthTimeout sets the global timeout for health checks
func SetHealthTimeout(seconds int) {
	if seconds > 0 {
		healthTimeout = time.Duration(seconds) * time.Second
	} else {
		healthTimeout = defaultTimeout
	}
}

// GetHealthTimeout returns the current health check timeout
func GetHealthTimeout() time.Duration {
	return healthTimeout
}

// CreateHealthContext creates a context with the configured timeout
func CreateHealthContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), healthTimeout)
}

// HealthStatus represents the overall health status of a repository
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "healthy"
	HealthStatusWarning  HealthStatus = "warning"
	HealthStatusCritical HealthStatus = "critical"
	HealthStatusUnknown  HealthStatus = "unknown"
)

// HealthCheck represents a single health check result
type HealthCheck struct {
	Name        string       `json:"name"`
	Status      HealthStatus `json:"status"`
	Message     string       `json:"message"`
	Details     string       `json:"details,omitempty"`
	Severity    int          `json:"severity"` // 1=info, 2=warning, 3=critical
	Category    string       `json:"category"`
	LastChecked time.Time    `json:"last_checked"`
}

// RepositoryHealth represents the complete health status of a repository
type RepositoryHealth struct {
	Repository  config.Repository `json:"repository"`
	Status      HealthStatus      `json:"status"`
	Score       int               `json:"score"` // 0-100
	Checks      []HealthCheck     `json:"checks"`
	Summary     string            `json:"summary"`
	LastChecked time.Time         `json:"last_checked"`
}

// HealthReport contains the complete health report for all repositories
type HealthReport struct {
	Summary      HealthSummary      `json:"summary"`
	Repositories []RepositoryHealth `json:"repositories"`
	GeneratedAt  time.Time          `json:"generated_at"`
}

// HealthSummary provides an overview of the health status across all repositories
type HealthSummary struct {
	Total    int `json:"total"`
	Healthy  int `json:"healthy"`
	Warning  int `json:"warning"`
	Critical int `json:"critical"`
	Unknown  int `json:"unknown"`
}

// HealthChecker interface defines the contract for health checkers
type HealthChecker interface {
	Name() string
	Category() string
	Check(repo config.Repository) HealthCheck
}

// HealthOptions configures how health checks are performed
type HealthOptions struct {
	IncludeCategories []string `json:"include_categories,omitempty"`
	ExcludeCategories []string `json:"exclude_categories,omitempty"`
	Threshold         int      `json:"threshold,omitempty"` // Minimum score threshold
	Format            string   `json:"format,omitempty"`    // Output format: table, json, html
	OutputFile        string   `json:"output_file,omitempty"`
	Parallel          bool     `json:"parallel,omitempty"`
	Timeout           int      `json:"timeout,omitempty"` // Timeout in seconds for individual health checks
}

// CheckRepositoryHealth performs all health checks on a single repository
func CheckRepositoryHealth(repo config.Repository, options HealthOptions) RepositoryHealth {
	// Set the global timeout for this health check run
	if options.Timeout > 0 {
		SetHealthTimeout(options.Timeout)
	}

	checkers := GetHealthCheckers(options)
	checks := make([]HealthCheck, 0, len(checkers))

	for _, checker := range checkers {
		check := checker.Check(repo)
		checks = append(checks, check)
	}

	// Calculate overall status and score
	status, score := calculateOverallHealth(checks)
	summary := generateSummary(checks, status)

	return RepositoryHealth{
		Repository:  repo,
		Status:      status,
		Score:       score,
		Checks:      checks,
		Summary:     summary,
		LastChecked: time.Now(),
	}
}

// CheckAllRepositories performs health checks on all repositories
func CheckAllRepositories(repositories []config.Repository, options HealthOptions) HealthReport {
	repoHealths := make([]RepositoryHealth, 0, len(repositories))

	for _, repo := range repositories {
		health := CheckRepositoryHealth(repo, options)
		repoHealths = append(repoHealths, health)
	}

	// Sort by status (critical first, then warnings, then healthy)
	sort.Slice(repoHealths, func(i, j int) bool {
		return getStatusPriority(repoHealths[i].Status) > getStatusPriority(repoHealths[j].Status)
	})

	summary := generateHealthSummary(repoHealths)

	return HealthReport{
		Summary:      summary,
		Repositories: repoHealths,
		GeneratedAt:  time.Now(),
	}
}

// GetHealthCheckers returns all available health checkers based on options
func GetHealthCheckers(options HealthOptions) []HealthChecker {
	// Start with safe, non-git checkers by default
	allCheckers := []HealthChecker{
		&DeprecatedComponentsChecker{},
		&CyclomaticComplexityChecker{},
		&DependencyChecker{},
		&DocumentationChecker{},
		&LicenseChecker{},
		&CIStatusChecker{},
		// Add git checkers only if explicitly requested or no filters specified
	}

	// Add git-based checkers if safe to do so
	if shouldIncludeGitCheckers(options) {
		allCheckers = append(allCheckers,
			&GitStatusChecker{},
			&LastCommitChecker{},
		)
	}

	// Add security checkers (may be slow)
	if shouldIncludeSecurityCheckers(options) {
		allCheckers = append(allCheckers,
			&SecurityChecker{},
			&BranchProtectionChecker{},
		)
	}

	if len(options.IncludeCategories) == 0 && len(options.ExcludeCategories) == 0 {
		return allCheckers
	}

	var filtered []HealthChecker
	for _, checker := range allCheckers {
		category := checker.Category()

		// Check exclusions first
		if contains(options.ExcludeCategories, category) {
			continue
		}

		// If include list is specified, only include those categories
		if len(options.IncludeCategories) > 0 && !contains(options.IncludeCategories, category) {
			continue
		}

		filtered = append(filtered, checker)
	}

	return filtered
}

// shouldIncludeGitCheckers determines if git-based checkers should be included
func shouldIncludeGitCheckers(options HealthOptions) bool {
	// Include git checkers if explicitly requested
	if contains(options.IncludeCategories, "git") {
		return true
	}

	// Exclude if git category is excluded
	if contains(options.ExcludeCategories, "git") {
		return false
	}

	// Include by default if no specific filters
	return len(options.IncludeCategories) == 0
}

// shouldIncludeSecurityCheckers determines if security checkers should be included
func shouldIncludeSecurityCheckers(options HealthOptions) bool {
	// Include security checkers if explicitly requested
	if contains(options.IncludeCategories, "security") {
		return true
	}

	// Exclude if security category is excluded
	if contains(options.ExcludeCategories, "security") {
		return false
	}

	// Include by default if no specific filters
	return len(options.IncludeCategories) == 0
}

// calculateOverallHealth determines the overall health status and score from individual checks
func calculateOverallHealth(checks []HealthCheck) (HealthStatus, int) {
	if len(checks) == 0 {
		return HealthStatusUnknown, 0
	}

	var totalScore, maxScore int
	hasCritical, hasWarning := false, false

	for _, check := range checks {
		maxScore += 100

		switch check.Status {
		case HealthStatusHealthy:
			totalScore += 100
		case HealthStatusWarning:
			totalScore += 60
			hasWarning = true
		case HealthStatusCritical:
			totalScore += 20
			hasCritical = true
		case HealthStatusUnknown:
			totalScore += 50
		}
	}

	score := (totalScore * 100) / maxScore

	// Determine overall status
	var status HealthStatus
	if hasCritical {
		status = HealthStatusCritical
	} else if hasWarning {
		status = HealthStatusWarning
	} else {
		status = HealthStatusHealthy
	}

	return status, score
}

// generateSummary creates a human-readable summary of the health checks
func generateSummary(checks []HealthCheck, _ HealthStatus) string {
	critical := countByStatus(checks, HealthStatusCritical)
	warning := countByStatus(checks, HealthStatusWarning)
	healthy := countByStatus(checks, HealthStatusHealthy)

	if critical > 0 {
		return fmt.Sprintf("%d critical issues, %d warnings", critical, warning)
	} else if warning > 0 {
		return fmt.Sprintf("%d warnings, %d checks passed", warning, healthy)
	}
	return "All checks passed"
}

// generateHealthSummary creates a summary of health across all repositories
func generateHealthSummary(repoHealths []RepositoryHealth) HealthSummary {
	summary := HealthSummary{Total: len(repoHealths)}

	for _, health := range repoHealths {
		switch health.Status {
		case HealthStatusHealthy:
			summary.Healthy++
		case HealthStatusWarning:
			summary.Warning++
		case HealthStatusCritical:
			summary.Critical++
		default:
			summary.Unknown++
		}
	}

	return summary
}

// Helper functions
func getStatusPriority(status HealthStatus) int {
	switch status {
	case HealthStatusCritical:
		return 3
	case HealthStatusWarning:
		return 2
	case HealthStatusHealthy:
		return 1
	default:
		return 0
	}
}

func countByStatus(checks []HealthCheck, status HealthStatus) int {
	count := 0
	for _, check := range checks {
		if check.Status == status {
			count++
		}
	}
	return count
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// GetRepoPath gets the repository path, handling both Path field and URL-derived paths
func GetRepoPath(repo config.Repository) string {
	if repo.Path != "" {
		return repo.Path
	}
	// Fallback to URL-derived path logic if needed
	return filepath.Join(".", repo.Name)
}

// IsGitRepository checks if the directory is a git repository
func IsGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if info, err := os.Stat(gitDir); err == nil {
		return info.IsDir()
	}
	return false
}
