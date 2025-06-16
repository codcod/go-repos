package security

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codcod/repos/internal/checkers/base"
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/platform/commands"
)

// VulnerabilityChecker checks for security vulnerabilities
type VulnerabilityChecker struct {
	*base.BaseChecker
	executor commands.CommandExecutor
}

// NewVulnerabilityChecker creates a new vulnerability checker
func NewVulnerabilityChecker(executor commands.CommandExecutor) *VulnerabilityChecker {
	config := core.CheckerConfig{
		Enabled:    true,
		Severity:   "high",
		Timeout:    120 * time.Second, // Vulnerability checks can take longer
		Categories: []string{"security"},
	}

	return &VulnerabilityChecker{
		BaseChecker: base.NewBaseChecker(
			"vulnerability-scan",
			"Vulnerability Scanner",
			"security",
			config,
		),
		executor: executor,
	}
}

// Check performs the vulnerability check
func (c *VulnerabilityChecker) Check(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	return c.Execute(ctx, repoCtx, func() (core.CheckResult, error) {
		return c.checkVulnerabilities(ctx, repoCtx)
	})
}

// checkVulnerabilities performs the actual vulnerability check
func (c *VulnerabilityChecker) checkVulnerabilities(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	builder := base.NewResultBuilder(c.ID(), c.Name(), c.Category())

	// Detect project type and run appropriate scanner
	projectType := c.detectProjectType(repoCtx.Repository.Path)
	builder.AddMetric("project_type", projectType)

	switch projectType {
	case "go":
		return c.checkGoVulnerabilities(ctx, repoCtx, builder)
	case "node":
		return c.checkNodeVulnerabilities(ctx, repoCtx, builder)
	case "python":
		return c.checkPythonVulnerabilities(ctx, repoCtx, builder)
	case "java":
		return c.checkJavaVulnerabilities(ctx, repoCtx, builder)
	default:
		return c.checkGeneralSecurity(repoCtx, builder)
	}
}

// detectProjectType detects the type of project
func (c *VulnerabilityChecker) detectProjectType(repoPath string) string {
	if _, err := os.Stat(filepath.Join(repoPath, "go.mod")); err == nil {
		return "go"
	}
	if _, err := os.Stat(filepath.Join(repoPath, "package.json")); err == nil {
		return "node"
	}
	if _, err := os.Stat(filepath.Join(repoPath, "requirements.txt")); err == nil {
		return "python"
	}
	if _, err := os.Stat(filepath.Join(repoPath, "pyproject.toml")); err == nil {
		return "python"
	}
	if _, err := os.Stat(filepath.Join(repoPath, "pom.xml")); err == nil {
		return "java"
	}
	if _, err := os.Stat(filepath.Join(repoPath, "build.gradle")); err == nil {
		return "java"
	}
	return "unknown"
}

// checkGoVulnerabilities checks for Go vulnerabilities using govulncheck
func (c *VulnerabilityChecker) checkGoVulnerabilities(ctx context.Context, repoCtx core.RepositoryContext, builder *base.ResultBuilder) (core.CheckResult, error) {
	// Check if govulncheck is available
	result := c.executor.Execute(ctx, "which", "govulncheck")
	if result.Error != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddIssue(base.NewIssueWithSuggestion(
			"scanner_not_available",
			core.SeverityMedium,
			"Go vulnerability scanner not available",
			"Install govulncheck: go install golang.org/x/vuln/cmd/govulncheck@latest",
		))
		return builder.Build(), nil
	}

	// Run govulncheck
	result = c.executor.ExecuteInDir(ctx, repoCtx.Repository.Path, "govulncheck", "./...")
	builder.AddMetric("scan_exit_code", result.ExitCode)

	if result.ExitCode == 0 {
		builder.WithStatus(core.StatusHealthy)
		builder.WithScore(100, 100)
		builder.AddMetric("vulnerabilities_found", 0)
	} else {
		builder.WithStatus(core.StatusCritical)
		builder.WithScore(20, 100)

		// Parse output for vulnerabilities
		vulnerabilities := c.parseGoVulnOutput(result.Stdout)
		builder.AddMetric("vulnerabilities_found", len(vulnerabilities))

		for i, vuln := range vulnerabilities {
			if i >= 5 { // Limit to first 5 to avoid too much output
				builder.AddMetric("additional_vulnerabilities", len(vulnerabilities)-5)
				break
			}
			builder.AddIssue(base.NewIssueWithSuggestion(
				"vulnerability_found",
				core.SeverityHigh,
				fmt.Sprintf("Vulnerability: %s", vuln),
				"Update vulnerable dependencies to patched versions",
			))
		}
	}

	return builder.Build(), nil
}

// parseGoVulnOutput parses govulncheck output to extract vulnerabilities
func (c *VulnerabilityChecker) parseGoVulnOutput(output string) []string {
	var vulnerabilities []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Vulnerability") || strings.Contains(line, "GO-") {
			vulnerabilities = append(vulnerabilities, line)
		}
	}

	return vulnerabilities
}

// checkNodeVulnerabilities checks for Node.js vulnerabilities using npm audit
func (c *VulnerabilityChecker) checkNodeVulnerabilities(ctx context.Context, repoCtx core.RepositoryContext, builder *base.ResultBuilder) (core.CheckResult, error) {
	// Check if npm is available
	result := c.executor.Execute(ctx, "which", "npm")
	if result.Error != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddIssue(base.NewIssueWithSuggestion(
			"scanner_not_available",
			core.SeverityMedium,
			"Node.js vulnerability scanner not available",
			"Install Node.js and npm to enable vulnerability scanning",
		))
		return builder.Build(), nil
	}

	// Run npm audit
	result = c.executor.ExecuteInDir(ctx, repoCtx.Repository.Path, "npm", "audit", "--audit-level=moderate")
	builder.AddMetric("audit_exit_code", result.ExitCode)

	if result.ExitCode == 0 {
		builder.WithStatus(core.StatusHealthy)
		builder.WithScore(100, 100)
		builder.AddMetric("vulnerabilities_found", 0)
	} else {
		builder.WithStatus(core.StatusCritical)
		builder.WithScore(30, 100)

		builder.AddIssue(base.NewIssueWithSuggestion(
			"npm_vulnerabilities",
			core.SeverityHigh,
			"npm audit found vulnerabilities",
			"Run 'npm audit fix' to automatically fix vulnerabilities, or update packages manually",
		))

		// Store audit output for details
		if len(result.Stdout) > 0 {
			builder.AddMetric("audit_summary", c.summarizeNpmAudit(result.Stdout))
		}
	}

	return builder.Build(), nil
}

// summarizeNpmAudit summarizes npm audit output
func (c *VulnerabilityChecker) summarizeNpmAudit(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "found") && strings.Contains(line, "vulnerabilit") {
			return strings.TrimSpace(line)
		}
	}
	return "npm audit completed with issues"
}

// checkPythonVulnerabilities checks for Python vulnerabilities using safety
func (c *VulnerabilityChecker) checkPythonVulnerabilities(ctx context.Context, repoCtx core.RepositoryContext, builder *base.ResultBuilder) (core.CheckResult, error) {
	// Check if safety is available
	result := c.executor.Execute(ctx, "which", "safety")
	if result.Error != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddIssue(base.NewIssueWithSuggestion(
			"scanner_not_available",
			core.SeverityMedium,
			"Python vulnerability scanner not available",
			"Install safety: pip install safety",
		))
		return builder.Build(), nil
	}

	// Run safety check
	result = c.executor.ExecuteInDir(ctx, repoCtx.Repository.Path, "safety", "check")
	builder.AddMetric("safety_exit_code", result.ExitCode)

	if result.ExitCode == 0 {
		builder.WithStatus(core.StatusHealthy)
		builder.WithScore(100, 100)
		builder.AddMetric("vulnerabilities_found", 0)
	} else {
		builder.WithStatus(core.StatusCritical)
		builder.WithScore(30, 100)

		builder.AddIssue(base.NewIssueWithSuggestion(
			"python_vulnerabilities",
			core.SeverityHigh,
			"Safety check found vulnerabilities",
			"Update vulnerable packages to secure versions",
		))
	}

	return builder.Build(), nil
}

// checkJavaVulnerabilities checks for Java vulnerabilities
func (c *VulnerabilityChecker) checkJavaVulnerabilities(_ context.Context, _ core.RepositoryContext, builder *base.ResultBuilder) (core.CheckResult, error) {
	// For Java, we'll check for OWASP dependency check or similar tools
	builder.WithStatus(core.StatusWarning)
	builder.AddIssue(base.NewIssueWithSuggestion(
		"java_scanner_not_implemented",
		core.SeverityMedium,
		"Java vulnerability scanning not implemented",
		"Consider using OWASP Dependency Check or Snyk for Java vulnerability scanning",
	))
	return builder.Build(), nil
}

// checkGeneralSecurity performs general security checks
func (c *VulnerabilityChecker) checkGeneralSecurity(repoCtx core.RepositoryContext, builder *base.ResultBuilder) (core.CheckResult, error) {
	// Check for security policy file
	securityFiles := []string{"SECURITY.md", ".github/SECURITY.md", "security.md"}
	var foundFiles []string

	for _, file := range securityFiles {
		if _, err := os.Stat(filepath.Join(repoCtx.Repository.Path, file)); err == nil {
			foundFiles = append(foundFiles, file)
		}
	}

	builder.AddMetric("security_policy_files", len(foundFiles))

	if len(foundFiles) > 0 {
		builder.WithStatus(core.StatusHealthy)
		builder.WithScore(70, 100) // Not perfect but shows security awareness
		builder.AddMetric("security_policy_status", "present")
		for i, file := range foundFiles {
			builder.AddMetric(fmt.Sprintf("security_file_%d", i), file)
		}
	} else {
		builder.WithStatus(core.StatusWarning)
		builder.WithScore(40, 100)
		builder.AddMetric("security_policy_status", "missing")
		builder.AddIssue(base.NewIssueWithSuggestion(
			"no_security_policy",
			core.SeverityMedium,
			"No security policy found",
			"Add a SECURITY.md file with vulnerability reporting guidelines",
		))
	}

	return builder.Build(), nil
}

// SupportsRepository checks if this checker supports the repository
func (c *VulnerabilityChecker) SupportsRepository(repo core.Repository) bool {
	// This checker supports all repositories
	return true
}
