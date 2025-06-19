package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/checkers/base"
	"github.com/codcod/repos/internal/health/commands"
)

// LastCommitChecker checks when the last commit was made
type LastCommitChecker struct {
	*base.BaseChecker
	executor commands.CommandExecutor
}

// NewLastCommitChecker creates a new last commit checker
func NewLastCommitChecker(executor commands.CommandExecutor) *LastCommitChecker {
	config := core.CheckerConfig{
		Enabled:    true,
		Severity:   "low",
		Timeout:    30 * time.Second,
		Categories: []string{"git"},
	}

	return &LastCommitChecker{
		BaseChecker: base.NewBaseChecker(
			"git-last-commit",
			"Last Commit",
			"git",
			config,
		),
		executor: executor,
	}
}

// Check performs the last commit check
func (c *LastCommitChecker) Check(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	return c.Execute(ctx, repoCtx, func() (core.CheckResult, error) {
		return c.checkLastCommit(ctx, repoCtx)
	})
}

// checkLastCommit performs the actual last commit check
func (c *LastCommitChecker) checkLastCommit(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	builder := base.NewResultBuilder(c.ID(), c.Name(), c.Category())

	// Check if it's a git repository
	if !c.isGitRepository(repoCtx.Repository.Path) {
		builder.WithStatus(core.StatusWarning)
		builder.AddWarning(core.Warning{
			Type:    "not_git_repo",
			Message: "Not a git repository",
		})
		return builder.Build(), nil
	}

	// Get last commit date
	result := c.executor.ExecuteInDir(ctx, repoCtx.Repository.Path, "git", "log", "-1", "--format=%ct")
	if result.Error != nil || len(result.Stdout) == 0 {
		builder.WithStatus(core.StatusWarning)
		builder.AddWarning(core.Warning{
			Type:    "git_command_error",
			Message: fmt.Sprintf("Unable to get last commit date: %v", result.Error),
		})
		return builder.Build(), nil
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(result.Stdout), 10, 64)
	if err != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddWarning(core.Warning{
			Type:    "parse_error",
			Message: fmt.Sprintf("Unable to parse commit timestamp: %v", err),
		})
		return builder.Build(), nil
	}

	lastCommit := time.Unix(timestamp, 0)
	daysSince := int(time.Since(lastCommit).Hours() / 24)

	builder.AddMetric("last_commit_timestamp", timestamp)
	builder.AddMetric("last_commit_date", lastCommit.Format("2006-01-02 15:04:05"))
	builder.AddMetric("days_since_last_commit", daysSince)

	// Evaluate freshness
	if daysSince <= 7 {
		builder.WithStatus(core.StatusHealthy)
		builder.WithScore(100, 100)
		builder.AddMetric("freshness", "excellent")
	} else if daysSince <= 30 {
		builder.WithStatus(core.StatusHealthy)
		builder.WithScore(80, 100)
		builder.AddMetric("freshness", "good")
	} else if daysSince <= 90 {
		builder.WithStatus(core.StatusWarning)
		builder.WithScore(60, 100)
		builder.AddMetric("freshness", "moderate")
		builder.AddWarning(core.Warning{
			Type:    "stale_repository",
			Message: fmt.Sprintf("Repository has not been updated for %d days", daysSince),
		})
	} else {
		builder.WithStatus(core.StatusCritical)
		builder.WithScore(30, 100)
		builder.AddMetric("freshness", "stale")
		builder.AddIssue(base.NewIssueWithSuggestion(
			"very_stale_repository",
			core.SeverityHigh,
			fmt.Sprintf("Repository has not been updated for %d days", daysSince),
			"Consider updating the repository or archiving if no longer maintained",
		))
	}

	return builder.Build(), nil
}

// isGitRepository checks if the path is a git repository
func (c *LastCommitChecker) isGitRepository(path string) bool {
	// Use git command to check if it's a repository
	result := c.executor.ExecuteInDir(context.Background(), path, "git", "rev-parse", "--is-inside-work-tree")
	return result.Error == nil && strings.TrimSpace(result.Stdout) == "true"
}

// SupportsRepository checks if this checker supports the repository
func (c *LastCommitChecker) SupportsRepository(repo core.Repository) bool {
	return c.isGitRepository(repo.Path)
}
