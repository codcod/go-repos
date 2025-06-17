package docs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codcod/repos/internal/checkers/base"
	"github.com/codcod/repos/internal/core"
)

const (
	// Code block indicators
	CodeBlockMarker    = "```"
	IndentedCodeSpaces = "    " // 4 spaces indicate code block

	// Badge patterns
	BadgeImageSyntax    = "!["
	BadgeKeyword        = "badge"
	ShieldsIOService    = "shields.io"
	TravisCIBadge       = "travis-ci"
	GitHubWorkflowBadge = "github/workflow"

	// License keywords
	LicenseKeyword = "license"
	MITLicense     = "mit"
	ApacheLicense  = "apache"
	GPLLicense     = "gpl"

	// Content thresholds
	MinReadmeLength      = 100
	MinDescriptionLength = 200
	TitleCheckLineLimit  = 10
)

var (
	ErrEmptyContent = errors.New("content cannot be empty")
	ErrNilChecker   = errors.New("checker cannot be nil")
)

// BadgePattern represents a badge detection pattern
type BadgePattern struct {
	Pattern     string
	Description string
}

// LicenseType represents different license types to detect
type LicenseType struct {
	Keywords    []string
	DisplayName string
}

// ReadmeCheckerConfig holds configuration for README analysis
type ReadmeCheckerConfig struct {
	RequireBadges         bool
	RequireCodeExamples   bool
	RequireLicenseInfo    bool
	CustomBadgePatterns   []string
	CustomLicenseKeywords []string
}

// ContentAnalyzer provides reusable content analysis methods
type ContentAnalyzer struct {
	content      string
	lowerContent string
}

// NewContentAnalyzer creates a new content analyzer
func NewContentAnalyzer(content string) *ContentAnalyzer {
	return &ContentAnalyzer{
		content:      content,
		lowerContent: strings.ToLower(content),
	}
}

// ContainsAny checks if content contains any of the provided patterns (case-insensitive)
func (ca *ContentAnalyzer) ContainsAny(patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(ca.lowerContent, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// ContainsAll checks if content contains all of the provided patterns (case-insensitive)
func (ca *ContentAnalyzer) ContainsAll(patterns []string) bool {
	for _, pattern := range patterns {
		if !strings.Contains(ca.lowerContent, strings.ToLower(pattern)) {
			return false
		}
	}
	return true
}

// DefaultConfig returns the default configuration
func DefaultConfig() *ReadmeCheckerConfig {
	return &ReadmeCheckerConfig{
		RequireBadges:       true,
		RequireCodeExamples: true,
		RequireLicenseInfo:  true,
	}
}

// getBadgePatterns returns all badge patterns to check
func getBadgePatterns() []BadgePattern {
	return []BadgePattern{
		{BadgeImageSyntax, "Markdown image syntax"},
		{BadgeKeyword, "Generic badge keyword"},
		{ShieldsIOService, "Shields.io badge service"},
		{TravisCIBadge, "Travis CI badge"},
		{GitHubWorkflowBadge, "GitHub Actions badge"},
	}
}

// getSupportedLicenses returns all supported license types
func getSupportedLicenses() []LicenseType {
	return []LicenseType{
		{[]string{"mit"}, "MIT License"},
		{[]string{"apache", "apache-2.0"}, "Apache License"},
		{[]string{"gpl", "gpl-3.0", "gpl-2.0"}, "GPL License"},
		{[]string{"bsd"}, "BSD License"},
		{[]string{"creative commons", "cc"}, "Creative Commons"},
	}
}

// ReadmeChecker analyzes repository README files for completeness and quality.
// It checks for essential documentation elements including:
// - Installation instructions
// - Usage examples and code blocks
// - Project badges for build status, coverage, etc.
// - License information

// ReadmeChecker checks for README files and their quality
type ReadmeChecker struct {
	*base.BaseChecker
	config *ReadmeCheckerConfig
}

// NewReadmeChecker creates a new README checker with default configuration
func NewReadmeChecker() *ReadmeChecker {
	return NewReadmeCheckerWithConfig(DefaultConfig())
}

// NewReadmeCheckerWithConfig creates a new README checker with the provided configuration
func NewReadmeCheckerWithConfig(config *ReadmeCheckerConfig) *ReadmeChecker {
	if config == nil {
		config = DefaultConfig()
	}

	checkerConfig := core.CheckerConfig{
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
			checkerConfig,
		),
		config: config,
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
	//nolint:gosec // This is intentional file reading for code analysis
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
	if len(contentStr) < MinReadmeLength {
		issues = append(issues, base.NewIssueWithSuggestion(
			"readme_too_short",
			core.SeverityMedium,
			fmt.Sprintf("README file is very short (less than %d characters)", MinReadmeLength),
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
	checkLimit := minInt(TitleCheckLineLimit, len(lines))
	for i, line := range lines[:checkLimit] {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
			return true
		}
		// Also check for underlined titles
		if len(line) > 0 && i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			if strings.Repeat("=", len(line)) == nextLine || strings.Repeat("-", len(line)) == nextLine {
				return true
			}
		}
	}
	return false
}

// hasDescription checks if README has a description section.
// It looks for common description indicators and checks if content is substantial
// enough to likely contain a meaningful description.
func (c *ReadmeChecker) hasDescription(content string) bool {
	if content == "" {
		return false
	}

	analyzer := NewContentAnalyzer(content)

	// Look for description indicators
	indicators := []string{"description", "about", "overview", "what", "this project"}
	if analyzer.ContainsAny(indicators) {
		return true
	}

	// Also check if content is substantial enough to likely contain a description
	return len(strings.TrimSpace(content)) > MinDescriptionLength
}

// hasInstallationInstructions checks if README contains installation or setup instructions
func (c *ReadmeChecker) hasInstallationInstructions(content string) bool {
	if content == "" {
		return false
	}

	analyzer := NewContentAnalyzer(content)
	indicators := []string{"install", "setup", "getting started", "build", "compile"}
	return analyzer.ContainsAny(indicators)
}

// hasUsageExamples checks if README contains usage examples or instructions
func (c *ReadmeChecker) hasUsageExamples(content string) bool {
	if content == "" {
		return false
	}

	analyzer := NewContentAnalyzer(content)
	indicators := []string{"usage", "example", "how to", "tutorial", "guide"}

	// Check for usage section keywords
	if analyzer.ContainsAny(indicators) {
		return true
	}

	// Check for code examples which often contain usage instructions
	return c.hasCodeExamples(content)
}

// hasCodeExamples analyzes README content for code examples and usage instructions.
// It detects:
// - Markdown code blocks (```)
// - Indented code blocks (4+ spaces)
// - Common usage section keywords
//
// Returns true if any code examples or usage instructions are found.
func (c *ReadmeChecker) hasCodeExamples(content string) bool {
	if content == "" {
		return false
	}

	if strings.Contains(content, CodeBlockMarker) {
		return true
	}

	if strings.Contains(content, IndentedCodeSpaces) {
		return true
	}

	// Check for usage sections that typically contain examples
	analyzer := NewContentAnalyzer(content)
	usageKeywords := []string{"usage", "example", "getting started", "quickstart", "how to"}

	return analyzer.ContainsAny(usageKeywords)
}

// hasBadges checks if README contains project badges
func (c *ReadmeChecker) hasBadges(content string) bool {
	if content == "" {
		return false
	}

	analyzer := NewContentAnalyzer(content)

	// Check standard badge patterns
	for _, badgePattern := range getBadgePatterns() {
		if strings.Contains(analyzer.lowerContent, badgePattern.Pattern) {
			return true
		}
	}

	// Check custom badge patterns from config
	if c.config != nil && len(c.config.CustomBadgePatterns) > 0 {
		return analyzer.ContainsAny(c.config.CustomBadgePatterns)
	}

	return false
}

// hasLicenseInfo checks if README mentions license information
func (c *ReadmeChecker) hasLicenseInfo(content string) bool {
	if content == "" {
		return false
	}

	analyzer := NewContentAnalyzer(content)

	// Check for generic license keyword first
	if strings.Contains(analyzer.lowerContent, LicenseKeyword) {
		return true
	}

	// Check for specific license types
	for _, license := range getSupportedLicenses() {
		if analyzer.ContainsAny(license.Keywords) {
			return true
		}
	}

	// Check custom license keywords from config
	if c.config != nil && len(c.config.CustomLicenseKeywords) > 0 {
		return analyzer.ContainsAny(c.config.CustomLicenseKeywords)
	}

	return false
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SupportsRepository checks if this checker supports the repository with validation
func (c *ReadmeChecker) SupportsRepository(repo core.Repository) bool {
	if repo.Name == "" {
		return false // Invalid repository
	}
	// This checker supports all valid repositories
	return true
}
