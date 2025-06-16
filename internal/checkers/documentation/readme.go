package documentation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codcod/repos/internal/checkers/base"
	"github.com/codcod/repos/internal/core"
)

// ReadmeChecker checks for README files and their quality
type ReadmeChecker struct {
	*base.BaseChecker
}

// NewReadmeChecker creates a new README checker
func NewReadmeChecker() *ReadmeChecker {
	config := core.CheckerConfig{
		Enabled:    true,
		Severity:   "medium",
		Timeout:    30 * time.Second,
		Categories: []string{"documentation"},
	}

	return &ReadmeChecker{
		BaseChecker: base.NewBaseChecker(
			"readme-check",
			"README Documentation",
			"documentation",
			config,
		),
	}
}

// Check performs the README check
func (c *ReadmeChecker) Check(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	return c.Execute(ctx, repoCtx, func() (core.CheckResult, error) {
		return c.checkReadme(repoCtx)
	})
}

// checkReadme performs the actual README check
func (c *ReadmeChecker) checkReadme(repoCtx core.RepositoryContext) (core.CheckResult, error) {
	builder := base.NewResultBuilder(c.ID(), c.Name(), c.Category())

	// Look for README files
	readmeFiles := c.findReadmeFiles(repoCtx.Repository.Path)
	builder.AddMetric("readme_files_found", len(readmeFiles))

	if len(readmeFiles) == 0 {
		builder.WithStatus(core.StatusCritical)
		builder.WithScore(0, 100)
		builder.AddIssue(base.NewIssueWithSuggestion(
			"no_readme",
			core.SeverityHigh,
			"No README file found",
			"Create a README.md file with project description, installation, and usage instructions",
		))
		return builder.Build(), nil
	}

	// Analyze the main README file
	mainReadme := c.selectMainReadme(readmeFiles)
	builder.AddMetric("main_readme", mainReadme)

	// Analyze README quality
	score, issues, warnings := c.analyzeReadmeQuality(repoCtx.Repository.Path, mainReadme)
	builder.WithScore(score, 100)

	// Add issues and warnings
	for _, issue := range issues {
		builder.AddIssue(issue)
	}
	for _, warning := range warnings {
		builder.AddWarning(warning)
	}

	// Set overall status
	if score >= 80 {
		builder.WithStatus(core.StatusHealthy)
	} else if score >= 50 {
		builder.WithStatus(core.StatusWarning)
	} else {
		builder.WithStatus(core.StatusCritical)
	}

	return builder.Build(), nil
}

// findReadmeFiles finds README files in the repository
func (c *ReadmeChecker) findReadmeFiles(repoPath string) []string {
	readmePatterns := []string{
		"README.md", "README.txt", "README.rst", "README",
		"readme.md", "readme.txt", "readme.rst", "readme",
		"Readme.md", "Readme.txt", "Readme.rst", "Readme",
	}

	var foundFiles []string
	for _, pattern := range readmePatterns {
		fullPath := filepath.Join(repoPath, pattern)
		if _, err := os.Stat(fullPath); err == nil {
			foundFiles = append(foundFiles, pattern)
		}
	}

	return foundFiles
}

// selectMainReadme selects the main README file from found files
func (c *ReadmeChecker) selectMainReadme(readmeFiles []string) string {
	// Prefer README.md, then README.txt, then others
	priorities := []string{"README.md", "README.txt", "README.rst", "README"}

	for _, priority := range priorities {
		for _, file := range readmeFiles {
			if strings.EqualFold(file, priority) {
				return file
			}
		}
	}

	// Return the first found file if no preferred format
	if len(readmeFiles) > 0 {
		return readmeFiles[0]
	}

	return ""
}

// analyzeReadmeQuality analyzes the quality of a README file
func (c *ReadmeChecker) analyzeReadmeQuality(repoPath, readmeFile string) (int, []core.Issue, []core.Warning) {
	var issues []core.Issue
	var warnings []core.Warning
	score := 0

	readmePath := filepath.Join(repoPath, readmeFile)
	content, err := os.ReadFile(readmePath)
	if err != nil {
		issues = append(issues, base.NewIssueWithSuggestion(
			"readme_read_error",
			core.SeverityMedium,
			fmt.Sprintf("Unable to read README file: %v", err),
			"Check file permissions and ensure the README file is accessible",
		))
		return 20, issues, warnings
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Basic presence check
	score += 20

	// Check file size/length
	if len(contentStr) < 100 {
		issues = append(issues, base.NewIssueWithSuggestion(
			"readme_too_short",
			core.SeverityMedium,
			"README file is very short (less than 100 characters)",
			"Add more detailed information about your project",
		))
	} else if len(contentStr) > 50 {
		score += 10
	}

	// Check for title/heading
	hasTitle := c.hasTitle(lines)
	if hasTitle {
		score += 15
	} else {
		warnings = append(warnings, core.Warning{
			Type:    "no_title",
			Message: "README lacks a clear title or main heading",
		})
	}

	// Check for description
	hasDescription := c.hasDescription(contentStr)
	if hasDescription {
		score += 15
	} else {
		issues = append(issues, base.NewIssueWithSuggestion(
			"no_description",
			core.SeverityMedium,
			"README lacks a clear project description",
			"Add a section explaining what your project does",
		))
	}

	// Check for installation instructions
	hasInstallation := c.hasInstallationInstructions(contentStr)
	if hasInstallation {
		score += 15
	} else {
		warnings = append(warnings, core.Warning{
			Type:    "no_installation",
			Message: "README lacks installation instructions",
		})
	}

	// Check for usage examples
	hasUsage := c.hasUsageExamples(contentStr)
	if hasUsage {
		score += 15
	} else {
		warnings = append(warnings, core.Warning{
			Type:    "no_usage",
			Message: "README lacks usage examples",
		})
	}

	// Check for badges
	hasBadges := c.hasBadges(contentStr)
	if hasBadges {
		score += 5
	}

	// Check for license information
	hasLicense := c.hasLicenseInfo(contentStr)
	if hasLicense {
		score += 5
	}

	return score, issues, warnings
}

// hasTitle checks if README has a title or main heading
func (c *ReadmeChecker) hasTitle(lines []string) bool {
	for _, line := range lines[:min(10, len(lines))] { // Check first 10 lines
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
			return true
		}
		// Also check for underlined titles
		if len(line) > 0 && len(lines) > 1 {
			nextLine := strings.TrimSpace(lines[1])
			if strings.Repeat("=", len(line)) == nextLine || strings.Repeat("-", len(line)) == nextLine {
				return true
			}
		}
	}
	return false
}

// hasDescription checks if README has a description
func (c *ReadmeChecker) hasDescription(content string) bool {
	lowerContent := strings.ToLower(content)
	// Look for description indicators
	indicators := []string{"description", "about", "overview", "what", "this project"}
	for _, indicator := range indicators {
		if strings.Contains(lowerContent, indicator) {
			return true
		}
	}
	// Also check if content is substantial enough to likely contain a description
	return len(strings.TrimSpace(content)) > 200
}

// hasInstallationInstructions checks if README has installation instructions
func (c *ReadmeChecker) hasInstallationInstructions(content string) bool {
	lowerContent := strings.ToLower(content)
	indicators := []string{"install", "setup", "getting started", "build", "compile"}
	for _, indicator := range indicators {
		if strings.Contains(lowerContent, indicator) {
			return true
		}
	}
	return false
}

// hasUsageExamples checks if README has usage examples
func (c *ReadmeChecker) hasUsageExamples(content string) bool {
	lowerContent := strings.ToLower(content)
	indicators := []string{"usage", "example", "how to", "tutorial", "guide"}
	for _, indicator := range indicators {
		if strings.Contains(lowerContent, indicator) {
			return true
		}
	}
	// Check for code blocks which often contain usage examples
	return strings.Contains(content, "```") || strings.Contains(content, "    ") // 4 spaces indicate code block
}

// hasBadges checks if README has badges
func (c *ReadmeChecker) hasBadges(content string) bool {
	// Look for common badge patterns
	badgePatterns := []string{
		"![",              // Markdown image syntax
		"badge",           // Common in badge URLs
		"shields.io",      // Popular badge service
		"travis-ci",       // CI badge
		"github/workflow", // GitHub Actions badge
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range badgePatterns {
		if strings.Contains(lowerContent, pattern) {
			return true
		}
	}
	return false
}

// hasLicenseInfo checks if README mentions license information
func (c *ReadmeChecker) hasLicenseInfo(content string) bool {
	lowerContent := strings.ToLower(content)
	return strings.Contains(lowerContent, "license") || strings.Contains(lowerContent, "mit") ||
		strings.Contains(lowerContent, "apache") || strings.Contains(lowerContent, "gpl")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SupportsRepository checks if this checker supports the repository
func (c *ReadmeChecker) SupportsRepository(repo core.Repository) bool {
	// This checker supports all repositories
	return true
}
