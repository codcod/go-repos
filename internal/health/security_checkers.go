package health

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codcod/repos/internal/config"
)

// BranchProtectionChecker checks if the main branch has protection enabled
type BranchProtectionChecker struct{}

func (c *BranchProtectionChecker) Name() string     { return "Branch Protection" }
func (c *BranchProtectionChecker) Category() string { return "security" }

func (c *BranchProtectionChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)

	if !IsGitRepository(repoPath) {
		return c.createHealthCheck(HealthStatusCritical, "Not a git repository", "", 3)
	}

	defaultBranch, err := c.getDefaultBranch(repoPath)
	if err != nil {
		return c.createHealthCheck(HealthStatusWarning, "Unable to determine default branch", err.Error(), 2)
	}

	var issues, warnings, info []string

	c.checkLocalProtectionConfig(repoPath, &info)
	c.checkGitHubProtection(repoPath, defaultBranch, &issues, &warnings, &info)
	c.checkCommonProtectionPatterns(repoPath, &info)
	c.checkDefaultBranchName(defaultBranch, &warnings)
	c.checkMergePatterns(repoPath, defaultBranch, &info)

	return c.createHealthCheckFromResults(defaultBranch, issues, warnings, info)
}

func (c *BranchProtectionChecker) createHealthCheck(status HealthStatus, message, details string, severity int) HealthCheck {
	return HealthCheck{
		Name:        c.Name(),
		Status:      status,
		Message:     message,
		Details:     details,
		Severity:    severity,
		Category:    c.Category(),
		LastChecked: time.Now(),
	}
}

func (c *BranchProtectionChecker) getDefaultBranch(repoPath string) (string, error) {
	output, err := executeCommandWithoutContext(repoPath, "git", "symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		// Fallback to checking common branch names
		branches := []string{"main", "master", "develop"}
		for _, branch := range branches {
			_, err := executeCommandWithoutContext(repoPath, "git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
			if err == nil {
				return branch, nil
			}
		}
		return "main", fmt.Errorf("unable to determine default branch")
	}

	// Extract branch name from refs/remotes/origin/HEAD -> origin/main
	branch := strings.TrimSpace(string(output))
	if strings.Contains(branch, "refs/remotes/origin/") {
		branch = strings.TrimPrefix(branch, "refs/remotes/origin/")
	}
	if branch == "" {
		return "main", nil
	}
	return branch, nil
}

// checkLocalProtectionConfig checks local branch protection configuration
func (c *BranchProtectionChecker) checkLocalProtectionConfig(repoPath string, info *[]string) {
	branchProtectionFiles := []string{
		".github/branch-protection.yml",
		".github/branch-protection.yaml",
		".github/workflows/branch-protection.yml",
		".github/workflows/branch-protection.yaml",
	}

	for _, file := range branchProtectionFiles {
		if _, err := os.Stat(filepath.Join(repoPath, file)); err == nil {
			*info = append(*info, fmt.Sprintf("Found branch protection config: %s", file))
			*info = append(*info, "Local branch protection configuration detected")
			break
		}
	}
}

// checkGitHubProtection checks GitHub branch protection via API
type protectionStatus struct {
	hasProtection bool
	details       string
	error         string
}

func (c *BranchProtectionChecker) checkGitHubProtection(repoPath, defaultBranch string, issues, warnings, info *[]string) {
	if !c.commandExists("gh") {
		*warnings = append(*warnings, "GitHub CLI not available - install 'gh' for complete branch protection checks")
		return
	}

	protectionStatus := c.checkGitHubBranchProtection(repoPath, defaultBranch)
	if protectionStatus.hasProtection {
		*info = append(*info, fmt.Sprintf("Branch '%s' has GitHub protection enabled", defaultBranch))
		if protectionStatus.details != "" {
			*info = append(*info, protectionStatus.details)
		}
	} else if protectionStatus.error != "" {
		*warnings = append(*warnings, fmt.Sprintf("GitHub CLI check failed: %s", protectionStatus.error))
	} else {
		*issues = append(*issues, fmt.Sprintf("Branch '%s' has no GitHub protection rules", defaultBranch))
	}
}

func (c *BranchProtectionChecker) checkGitHubBranchProtection(repoPath, defaultBranch string) protectionStatus {
	ctx, cancel := CreateHealthContext()
	defer cancel()

	cmd := exec.CommandContext(ctx, "gh", "api", fmt.Sprintf("repos/:owner/:repo/branches/%s/protection", defaultBranch))
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return protectionStatus{error: "GitHub API check timed out"}
		}
		return protectionStatus{error: fmt.Sprintf("GitHub API error: %v", err)}
	}

	if len(output) > 0 && !strings.Contains(string(output), "Branch not protected") {
		return protectionStatus{hasProtection: true, details: "GitHub branch protection rules detected"}
	}

	return protectionStatus{hasProtection: false}
}

// checkCommonProtectionPatterns checks for common protection patterns
func (c *BranchProtectionChecker) checkCommonProtectionPatterns(repoPath string, info *[]string) {
	// Check for common patterns that indicate branch protection awareness
	protectionPatterns := []string{
		".github/CODEOWNERS",
		".github/pull_request_template.md",
		".github/workflows/ci.yml",
		".github/workflows/ci.yaml",
	}

	for _, pattern := range protectionPatterns {
		if _, err := os.Stat(filepath.Join(repoPath, pattern)); err == nil {
			*info = append(*info, fmt.Sprintf("Found protection-related file: %s", pattern))
		}
	}
}

func (c *BranchProtectionChecker) checkMergeCommitPatterns(repoPath, defaultBranch string) string {
	output, err := executeCommandWithoutContext(repoPath, "git", "log", "--oneline", "--merges", "-10", defaultBranch)
	if err != nil {
		return ""
	}

	mergeCount := len(strings.Split(strings.TrimSpace(string(output)), "\n"))
	if mergeCount > 0 && len(output) > 0 {
		return fmt.Sprintf("Found %d recent merge commits (indicates pull request workflow)", mergeCount)
	}

	return ""
}

// checkMergePatterns checks for merge patterns in the repository
func (c *BranchProtectionChecker) checkMergePatterns(repoPath, defaultBranch string, info *[]string) {
	if mergeCommitInfo := c.checkMergeCommitPatterns(repoPath, defaultBranch); mergeCommitInfo != "" {
		*info = append(*info, mergeCommitInfo)
	}
}

func (c *BranchProtectionChecker) createHealthCheckFromResults(defaultBranch string, issues, warnings, info []string) HealthCheck {
	status := HealthStatusHealthy
	message := fmt.Sprintf("Branch protection appears configured for '%s'", defaultBranch)
	severity := 1

	if len(issues) > 0 {
		status = HealthStatusCritical
		message = fmt.Sprintf("Branch protection issues: %s", strings.Join(issues, ", "))
		severity = 3
	} else if len(warnings) > 0 {
		status = HealthStatusWarning
		message = fmt.Sprintf("Branch protection warnings: %s", strings.Join(warnings, ", "))
		severity = 2
	} else if len(info) == 0 {
		status = HealthStatusWarning
		message = "Unable to verify branch protection status"
		severity = 2
	}

	details := ""
	if len(issues) > 0 || len(warnings) > 0 || len(info) > 0 {
		allItems := append(append(issues, warnings...), info...)
		details = strings.Join(allItems, "\n")
	}

	return c.createHealthCheck(status, message, details, severity)
}

// commandExists checks if a command is available
func (c *BranchProtectionChecker) commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// checkDefaultBranchName checks if the default branch name follows conventions
func (c *BranchProtectionChecker) checkDefaultBranchName(defaultBranch string, warnings *[]string) {
	if defaultBranch != "main" && defaultBranch != "master" && defaultBranch != "develop" {
		*warnings = append(*warnings, fmt.Sprintf("Unusual default branch name: '%s'", defaultBranch))
	}
}

// SecurityChecker checks for security vulnerabilities
type SecurityChecker struct{}

func (c *SecurityChecker) Name() string     { return "Security" }
func (c *SecurityChecker) Category() string { return "security" }

func (c *SecurityChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)

	// Check for Go vulnerabilities if it's a Go project
	if _, err := os.Stat(filepath.Join(repoPath, "go.mod")); err == nil {
		return c.checkGoVulnerabilities(repoPath)
	}

	// Basic security file checks
	securityFiles := []string{"SECURITY.md", ".github/SECURITY.md", "security.md"}
	found := fileExistsInPath(repoPath, securityFiles)

	if len(found) > 0 {
		return createHealthCheck(
			c.Name(),
			c.Category(),
			HealthStatusHealthy,
			"Security policy found",
			fmt.Sprintf("Found: %s", strings.Join(found, ", ")),
			1,
		)
	}

	return createHealthCheck(
		c.Name(),
		c.Category(),
		HealthStatusWarning,
		"No security policy found",
		"Consider adding a SECURITY.md file",
		1,
	)
}

func (c *SecurityChecker) checkGoVulnerabilities(repoPath string) HealthCheck {
	// Check if govulncheck is available
	if !commandAvailable("govulncheck") {
		return createHealthCheck(
			c.Name(),
			c.Category(),
			HealthStatusWarning,
			"Security scanner not available",
			"Install govulncheck: go install golang.org/x/vuln/cmd/govulncheck@latest",
			2,
		)
	}

	// Run govulncheck
	output, err := executeCommandWithoutContext(repoPath, "govulncheck", "./...")
	if err != nil {
		return createHealthCheck(
			c.Name(),
			c.Category(),
			HealthStatusCritical,
			"Security vulnerabilities found",
			string(output),
			3,
		)
	}

	return createHealthCheck(
		c.Name(),
		c.Category(),
		HealthStatusHealthy,
		"No security vulnerabilities found",
		"",
		1,
	)
}
