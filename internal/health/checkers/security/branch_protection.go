package security

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/checkers/base"
	"github.com/codcod/repos/internal/platform/commands"
)

// BranchProtectionChecker checks if the main branch has protection enabled
type BranchProtectionChecker struct {
	*base.BaseChecker
	executor commands.CommandExecutor
}

// NewBranchProtectionChecker creates a new branch protection checker
func NewBranchProtectionChecker(executor commands.CommandExecutor) *BranchProtectionChecker {
	config := core.CheckerConfig{
		Enabled:    true,
		Severity:   "high",
		Timeout:    30 * time.Second,
		Categories: []string{"security"},
	}

	return &BranchProtectionChecker{
		BaseChecker: base.NewBaseChecker(
			"branch-protection",
			"Branch Protection",
			"security",
			config,
		),
		executor: executor,
	}
}

// Check performs the branch protection check
func (c *BranchProtectionChecker) Check(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	return c.Execute(ctx, repoCtx, func() (core.CheckResult, error) {
		return c.checkBranchProtection(ctx, repoCtx)
	})
}

// checkBranchProtection performs the actual branch protection check
func (c *BranchProtectionChecker) checkBranchProtection(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	builder := base.NewResultBuilder(c.ID(), c.Name(), c.Category())

	// Check if it's a git repository
	if !c.isGitRepository(repoCtx.Repository.Path) {
		builder.WithStatus(core.StatusCritical)
		builder.AddIssue(base.NewIssueWithSuggestion(
			"not_git_repo",
			core.SeverityCritical,
			"Not a git repository",
			"Initialize git repository with 'git init' or check if this is the correct path",
		))
		return builder.Build(), nil
	}

	// Get the default branch
	defaultBranch, err := c.getDefaultBranch(ctx, repoCtx.Repository.Path)
	if err != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddWarning(core.Warning{
			Type:    "branch_detection_error",
			Message: fmt.Sprintf("Unable to determine default branch: %v", err),
		})
		defaultBranch = "main" // fallback
	}

	builder.AddMetric("default_branch", defaultBranch)

	// Check for local protection configuration
	hasLocalConfig := c.checkLocalProtectionConfig(repoCtx.Repository.Path)
	builder.AddMetric("has_local_config", hasLocalConfig)

	// Check for GitHub CLI and protection
	hasGitHubProtection, ghError := c.checkGitHubProtection(ctx, repoCtx.Repository.Path, defaultBranch)
	builder.AddMetric("has_github_protection", hasGitHubProtection)

	// Check for common protection patterns
	protectionIndicators := c.checkCommonProtectionPatterns(repoCtx.Repository.Path)
	builder.AddMetric("protection_indicators", len(protectionIndicators))

	// Check merge patterns
	hasMergePatterns := c.checkMergePatterns(ctx, repoCtx.Repository.Path, defaultBranch)
	builder.AddMetric("has_merge_patterns", hasMergePatterns)

	// Evaluate overall protection status
	c.evaluateProtectionStatus(builder, defaultBranch, hasLocalConfig, hasGitHubProtection, protectionIndicators, hasMergePatterns, ghError)

	return builder.Build(), nil
}

// getDefaultBranch attempts to determine the default branch
func (c *BranchProtectionChecker) getDefaultBranch(ctx context.Context, repoPath string) (string, error) {
	// Try to get default branch from remote
	result := c.executor.ExecuteInDir(ctx, repoPath, "git", "symbolic-ref", "refs/remotes/origin/HEAD")
	if result.Error == nil && len(result.Stdout) > 0 {
		branch := strings.TrimSpace(result.Stdout)
		if strings.Contains(branch, "refs/remotes/origin/") {
			branch = strings.TrimPrefix(branch, "refs/remotes/origin/")
		}
		if branch != "" {
			return branch, nil
		}
	}

	// Fallback to checking common branch names
	branches := []string{"main", "master", "develop"}
	for _, branch := range branches {
		result = c.executor.ExecuteInDir(ctx, repoPath, "git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
		if result.Error == nil {
			return branch, nil
		}
	}

	return "main", fmt.Errorf("unable to determine default branch")
}

// checkLocalProtectionConfig checks for local branch protection configuration
func (c *BranchProtectionChecker) checkLocalProtectionConfig(repoPath string) bool {
	branchProtectionFiles := []string{
		".github/branch-protection.yml",
		".github/branch-protection.yaml",
		".github/workflows/branch-protection.yml",
		".github/workflows/branch-protection.yaml",
	}

	for _, file := range branchProtectionFiles {
		if _, err := os.Stat(filepath.Join(repoPath, file)); err == nil {
			return true
		}
	}
	return false
}

// checkGitHubProtection checks GitHub branch protection via CLI
func (c *BranchProtectionChecker) checkGitHubProtection(ctx context.Context, repoPath, defaultBranch string) (bool, error) {
	// Check if GitHub CLI is available
	result := c.executor.Execute(ctx, "which", "gh")
	if result.Error != nil {
		return false, fmt.Errorf("GitHub CLI not available")
	}

	// Try to get branch protection info
	result = c.executor.ExecuteInDir(ctx, repoPath, "gh", "api", fmt.Sprintf("repos/:owner/:repo/branches/%s/protection", defaultBranch))
	if result.Error != nil {
		return false, result.Error
	}

	// If we get output and it's not a "not protected" message, assume protection exists
	output := strings.TrimSpace(result.Stdout)
	if len(output) > 0 && !strings.Contains(output, "Branch not protected") {
		return true, nil
	}

	return false, nil
}

// checkCommonProtectionPatterns checks for files that indicate protection awareness
func (c *BranchProtectionChecker) checkCommonProtectionPatterns(repoPath string) []string {
	var indicators []string
	protectionPatterns := []string{
		".github/CODEOWNERS",
		".github/pull_request_template.md",
		".github/workflows/ci.yml",
		".github/workflows/ci.yaml",
		".github/workflows/test.yml",
		".github/workflows/test.yaml",
	}

	for _, pattern := range protectionPatterns {
		if _, err := os.Stat(filepath.Join(repoPath, pattern)); err == nil {
			indicators = append(indicators, pattern)
		}
	}

	return indicators
}

// checkMergePatterns checks for merge commit patterns in recent history
func (c *BranchProtectionChecker) checkMergePatterns(ctx context.Context, repoPath, defaultBranch string) bool {
	result := c.executor.ExecuteInDir(ctx, repoPath, "git", "log", "--oneline", "--merges", "-10", defaultBranch)
	if result.Error != nil {
		return false
	}

	output := strings.TrimSpace(result.Stdout)
	if len(output) > 0 {
		lines := strings.Split(output, "\n")
		return len(lines) > 0 && lines[0] != ""
	}

	return false
}

// evaluateProtectionStatus evaluates the overall protection status
func (c *BranchProtectionChecker) evaluateProtectionStatus(builder *base.ResultBuilder, defaultBranch string, hasLocalConfig, hasGitHubProtection bool, protectionIndicators []string, hasMergePatterns bool, ghError error) {
	score := 0
	maxScore := 100

	// GitHub protection is the gold standard
	if hasGitHubProtection {
		score += 50
		builder.AddMetric("github_protection_status", "enabled")
	} else if ghError != nil {
		builder.AddWarning(core.Warning{
			Type:    "github_cli_error",
			Message: fmt.Sprintf("Unable to check GitHub protection: %v", ghError),
		})
		builder.AddMetric("github_protection_status", "unknown")
	} else {
		builder.AddMetric("github_protection_status", "disabled")
	}

	// Local configuration files
	if hasLocalConfig {
		score += 20
		builder.AddMetric("local_config_status", "present")
	} else {
		builder.AddMetric("local_config_status", "missing")
	}

	// Protection indicators (CODEOWNERS, workflows, etc.)
	if len(protectionIndicators) > 0 {
		score += 20
		builder.AddMetric("protection_indicators_status", "present")
		for i, indicator := range protectionIndicators {
			builder.AddMetric(fmt.Sprintf("indicator_%d", i), indicator)
		}
	} else {
		builder.AddMetric("protection_indicators_status", "missing")
	}

	// Merge patterns indicate PR workflow
	if hasMergePatterns {
		score += 10
		builder.AddMetric("merge_patterns_status", "present")
	} else {
		builder.AddMetric("merge_patterns_status", "missing")
	}

	builder.WithScore(score, maxScore)

	// Determine overall status and issues
	if score >= 70 {
		builder.WithStatus(core.StatusHealthy)
	} else if score >= 40 {
		builder.WithStatus(core.StatusWarning)
		builder.AddIssue(base.NewIssueWithSuggestion(
			"incomplete_branch_protection",
			core.SeverityMedium,
			fmt.Sprintf("Branch protection for '%s' appears incomplete", defaultBranch),
			"Consider enabling GitHub branch protection rules or adding local protection configuration",
		))
	} else {
		builder.WithStatus(core.StatusCritical)
		builder.AddIssue(base.NewIssueWithSuggestion(
			"no_branch_protection",
			core.SeverityHigh,
			fmt.Sprintf("No branch protection detected for '%s'", defaultBranch),
			"Enable GitHub branch protection rules and add CODEOWNERS file for better security",
		))
	}
}

// isGitRepository checks if the path is a git repository
func (c *BranchProtectionChecker) isGitRepository(path string) bool {
	result := c.executor.ExecuteInDir(context.Background(), path, "git", "rev-parse", "--is-inside-work-tree")
	return result.Error == nil && strings.TrimSpace(result.Stdout) == "true"
}

// SupportsRepository checks if this checker supports the repository
func (c *BranchProtectionChecker) SupportsRepository(repo core.Repository) bool {
	return c.isGitRepository(repo.Path)
}
