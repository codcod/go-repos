package health

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codcod/repos/internal/config"
)

// GitStatusChecker checks the git status of the repository
type GitStatusChecker struct{}

func (c *GitStatusChecker) Name() string     { return "Git Status" }
func (c *GitStatusChecker) Category() string { return "git" }

func (c *GitStatusChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)

	if !IsGitRepository(repoPath) {
		return createHealthCheck(
			c.Name(),
			c.Category(),
			HealthStatusCritical,
			"Not a git repository",
			"",
			3,
		)
	}

	// Check for uncommitted changes using git status --porcelain
	output, err := executeCommandWithoutContext(repoPath, "git", "status", "--porcelain")
	if err != nil {
		return createHealthCheck(
			c.Name(),
			c.Category(),
			HealthStatusWarning,
			fmt.Sprintf("Unable to check git status: %v", err),
			"",
			2,
		)
	}

	if len(output) > 0 {
		return createHealthCheck(
			c.Name(),
			c.Category(),
			HealthStatusWarning,
			"Repository has uncommitted changes",
			"Run 'git status' to see changes",
			1,
		)
	}

	return createHealthCheck(
		c.Name(),
		c.Category(),
		HealthStatusHealthy,
		"Git status is clean",
		"",
		1,
	)
}

// LastCommitChecker checks when the last commit was made
type LastCommitChecker struct{}

func (c *LastCommitChecker) Name() string     { return "Last Commit" }
func (c *LastCommitChecker) Category() string { return "git" }

func (c *LastCommitChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)

	if !IsGitRepository(repoPath) {
		return createHealthCheck(
			c.Name(),
			c.Category(),
			HealthStatusUnknown,
			"Not a git repository",
			"",
			1,
		)
	}

	// Get last commit date with timeout
	output, err := executeCommandWithoutContext(repoPath, "git", "log", "-1", "--format=%ct")
	if err != nil || len(output) == 0 {
		return createHealthCheck(
			c.Name(),
			c.Category(),
			HealthStatusWarning,
			"Unable to get last commit date",
			fmt.Sprintf("Error: %v", err),
			2,
		)
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return createHealthCheck(
			c.Name(),
			c.Category(),
			HealthStatusWarning,
			"Invalid commit timestamp",
			"",
			2,
		)
	}

	lastCommit := time.Unix(timestamp, 0)
	daysSince := int(time.Since(lastCommit).Hours() / 24)

	var status HealthStatus
	var message string
	var severity int

	if daysSince > 365 {
		status = HealthStatusCritical
		message = fmt.Sprintf("Last commit was %d days ago", daysSince)
		severity = 3
	} else if daysSince > 90 {
		status = HealthStatusWarning
		message = fmt.Sprintf("Last commit was %d days ago", daysSince)
		severity = 2
	} else {
		status = HealthStatusHealthy
		message = fmt.Sprintf("Last commit was %d days ago", daysSince)
		severity = 1
	}

	return createHealthCheck(
		c.Name(),
		c.Category(),
		status,
		message,
		fmt.Sprintf("Last commit: %s", lastCommit.Format("2006-01-02 15:04:05")),
		severity,
	)
}
