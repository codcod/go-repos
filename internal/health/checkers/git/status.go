package git

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/checkers/base"
	"github.com/codcod/repos/internal/platform/commands"
)

// GitStatusChecker checks the git status of the repository
type GitStatusChecker struct {
	*base.BaseChecker
	executor commands.CommandExecutor
}

// NewGitStatusChecker creates a new git status checker
func NewGitStatusChecker(executor commands.CommandExecutor) *GitStatusChecker {
	config := core.CheckerConfig{
		Enabled:    true,
		Severity:   "medium",
		Timeout:    30 * time.Second,
		Categories: []string{"git"},
	}

	return &GitStatusChecker{
		BaseChecker: base.NewBaseChecker(
			"git-status",
			"Git Status",
			"git",
			config,
		),
		executor: executor,
	}
}

// Check performs the git status check
func (c *GitStatusChecker) Check(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	return c.Execute(ctx, repoCtx, func() (core.CheckResult, error) {
		return c.checkGitStatus(ctx, repoCtx)
	})
}

// checkGitStatus performs the actual git status check
func (c *GitStatusChecker) checkGitStatus(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	builder := base.NewResultBuilder(c.ID(), c.Name(), c.Category())

	// Check if it's a git repository
	if !c.isGitRepository(repoCtx.Repository.Path) {
		builder.WithStatus(core.StatusCritical)
		issue := base.NewIssueWithSuggestion(
			"not_git_repo",
			core.SeverityCritical,
			"Not a git repository",
			"Initialize git repository with 'git init' or check if this is the correct path",
		)
		issue.Location = &core.Location{File: repoCtx.Repository.Path}
		builder.AddIssue(issue)
		return builder.Build(), nil
	}

	// Check for uncommitted changes using git status --porcelain
	result := c.executor.ExecuteInDir(ctx, repoCtx.Repository.Path, "git", "status", "--porcelain")

	if result.Error != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddWarning(core.Warning{
			Type:    "git_command_error",
			Message: fmt.Sprintf("Unable to check git status: %v", result.Error),
		})
		return builder.Build(), nil
	}

	// Parse git status output
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	if len(lines) == 1 && lines[0] == "" {
		// Clean status
		builder.WithStatus(core.StatusHealthy)
		builder.WithScore(100, 100)
		builder.AddMetric("uncommitted_files", 0)
		builder.AddMetric("status", "clean")
	} else {
		// Uncommitted changes found
		builder.WithStatus(core.StatusWarning)
		builder.WithScore(70, 100)

		uncommittedFiles := c.parseGitStatus(lines)
		builder.AddMetric("uncommitted_files", len(uncommittedFiles))
		builder.AddMetric("status", "dirty")

		builder.AddIssue(base.NewIssueWithSuggestion(
			"uncommitted_changes",
			core.SeverityMedium,
			fmt.Sprintf("Repository has %d uncommitted changes", len(uncommittedFiles)),
			"Review and commit changes with 'git add' and 'git commit', or stash them with 'git stash'",
		))

		// Add details about uncommitted files
		for i, file := range uncommittedFiles {
			if i >= 5 { // Limit to first 5 files to avoid too much output
				builder.AddMetric(fmt.Sprintf("file_%d", i), fmt.Sprintf("... and %d more", len(uncommittedFiles)-5))
				break
			}
			builder.AddMetric(fmt.Sprintf("file_%d", i), fmt.Sprintf("%s (%s)", file.Name, file.Status))
		}
	}

	return builder.Build(), nil
}

// GitFile represents a file with git status
type GitFile struct {
	Name   string
	Status string
}

// parseGitStatus parses git status --porcelain output
func (c *GitStatusChecker) parseGitStatus(lines []string) []GitFile {
	var files []GitFile

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		statusCode := line[:2]
		fileName := line[3:]

		status := c.getStatusDescription(statusCode)
		files = append(files, GitFile{
			Name:   fileName,
			Status: status,
		})
	}

	return files
}

// getStatusDescription converts git status codes to descriptions
//
//nolint:gocyclo // Switch statement for git status codes requires multiple branches
func (c *GitStatusChecker) getStatusDescription(code string) string {
	switch code {
	case "??":
		return "untracked"
	case " M":
		return "modified"
	case "M ":
		return "modified (staged)"
	case "MM":
		return "modified (staged and unstaged)"
	case " A":
		return "added"
	case "A ":
		return "added (staged)"
	case " D":
		return "deleted"
	case "D ":
		return "deleted (staged)"
	case "R ":
		return "renamed"
	case "C ":
		return "copied"
	default:
		return strings.TrimSpace(code)
	}
}

// isGitRepository checks if the path is a git repository
func (c *GitStatusChecker) isGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if info, err := filepath.Glob(gitDir); err == nil && len(info) > 0 {
		return true
	}

	// Also check if git command recognizes it as a repo
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = path
	err := cmd.Run()
	return err == nil
}

// SupportsRepository checks if this checker supports the repository
func (c *GitStatusChecker) SupportsRepository(repo core.Repository) bool {
	return c.isGitRepository(repo.Path)
}
