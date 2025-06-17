package compliance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/checkers/base"
)

// LicenseChecker checks for license files and compliance
type LicenseChecker struct {
	*base.BaseChecker
}

// NewLicenseChecker creates a new license checker
func NewLicenseChecker() *LicenseChecker {
	config := core.CheckerConfig{
		Enabled:    true,
		Severity:   "medium",
		Timeout:    30 * time.Second,
		Categories: []string{"compliance"},
	}

	return &LicenseChecker{
		BaseChecker: base.NewBaseChecker(
			"license-check",
			"License Compliance",
			"compliance",
			config,
		),
	}
}

// Check performs the license check
func (c *LicenseChecker) Check(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	return c.Execute(ctx, repoCtx, func() (core.CheckResult, error) {
		return c.checkLicense(repoCtx)
	})
}

// checkLicense performs the actual license check
func (c *LicenseChecker) checkLicense(repoCtx core.RepositoryContext) (core.CheckResult, error) {
	builder := base.NewResultBuilder(c.ID(), c.Name(), c.Category())

	// Look for license files
	licenseFiles := c.findLicenseFiles(repoCtx.Repository.Path)
	builder.AddMetric("license_files_found", len(licenseFiles))

	if len(licenseFiles) == 0 {
		builder.WithStatus(core.StatusCritical)
		builder.WithScore(0, 100)
		builder.AddIssue(base.NewIssueWithSuggestion(
			"no_license",
			core.SeverityHigh,
			"No license file found",
			"Add a LICENSE file to clarify how others can use your project",
		))
		return builder.Build(), nil
	}

	// Analyze the main license file
	mainLicense := c.selectMainLicense(licenseFiles)
	builder.AddMetric("main_license_file", mainLicense)

	// Analyze license content
	licenseType, confidence, issues, warnings := c.analyzeLicenseContent(repoCtx.Repository.Path, mainLicense)
	builder.AddMetric("license_type", licenseType)
	builder.AddMetric("detection_confidence", confidence)

	// Add issues and warnings
	for _, issue := range issues {
		builder.AddIssue(issue)
	}
	for _, warning := range warnings {
		builder.AddWarning(warning)
	}

	// Calculate score based on license presence and clarity
	score := c.calculateLicenseScore(licenseType, confidence, len(issues))
	builder.WithScore(score, 100)

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

// findLicenseFiles finds license files in the repository
func (c *LicenseChecker) findLicenseFiles(repoPath string) []string {
	licensePatterns := []string{
		"LICENSE", "LICENSE.txt", "LICENSE.md", "LICENSE.rst",
		"license", "license.txt", "license.md", "license.rst",
		"License", "License.txt", "License.md", "License.rst", "LICENSE", "LICENSE.txt", "LICENSE.md", "LICENSE.rst",
		"COPYING", "COPYING.txt", "COPYRIGHT", "COPYRIGHT.txt",
	}

	var foundFiles []string
	for _, pattern := range licensePatterns {
		fullPath := filepath.Join(repoPath, pattern)
		if _, err := os.Stat(fullPath); err == nil {
			foundFiles = append(foundFiles, pattern)
		}
	}

	return foundFiles
}

// selectMainLicense selects the main license file from found files
func (c *LicenseChecker) selectMainLicense(licenseFiles []string) string {
	// Prefer LICENSE, then LICENSE.txt, then others
	priorities := []string{"LICENSE", "LICENSE.txt", "LICENSE.md", "license", "license.txt"}

	for _, priority := range priorities {
		for _, file := range licenseFiles {
			if strings.EqualFold(file, priority) {
				return file
			}
		}
	}

	// Return the first found file if no preferred format
	if len(licenseFiles) > 0 {
		return licenseFiles[0]
	}

	return ""
}

// analyzeLicenseContent analyzes the content of a license file
func (c *LicenseChecker) analyzeLicenseContent(repoPath, licenseFile string) (string, string, []core.Issue, []core.Warning) {
	var issues []core.Issue
	var warnings []core.Warning

	licensePath := filepath.Join(repoPath, licenseFile)
	content, err := os.ReadFile(licensePath) //nolint:gosec // License file path is from repository analysis
	if err != nil {
		issues = append(issues, base.NewIssueWithSuggestion(
			"license_read_error",
			core.SeverityMedium,
			fmt.Sprintf("Unable to read license file: %v", err),
			"Check file permissions and ensure the license file is accessible",
		))
		return "unknown", "low", issues, warnings
	}

	contentStr := strings.ToLower(string(content))

	// Detect license type
	licenseType, confidence := c.detectLicenseType(contentStr)

	// Check for common issues
	if len(content) < 50 {
		issues = append(issues, base.NewIssueWithSuggestion(
			"license_too_short",
			core.SeverityMedium,
			"License file is very short and may not be valid",
			"Ensure the license file contains the complete license text",
		))
	}

	// Check for placeholders that need to be filled
	placeholders := c.findPlaceholders(string(content))
	if len(placeholders) > 0 {
		issues = append(issues, base.NewIssueWithSuggestion(
			"license_placeholders",
			core.SeverityMedium,
			fmt.Sprintf("License contains unfilled placeholders: %s", strings.Join(placeholders, ", ")),
			"Fill in the placeholder values (e.g., [year], [fullname]) in your license",
		))
	}

	// Check for copyright notice
	if !c.hasCopyrightNotice(string(content)) {
		warnings = append(warnings, core.Warning{
			Type:    "no_copyright_notice",
			Message: "License lacks a copyright notice",
		})
	}

	return licenseType, confidence, issues, warnings
}

// detectLicenseType attempts to detect the type of license
func (c *LicenseChecker) detectLicenseType(content string) (string, string) {
	licenses := map[string][]string{
		"MIT": {
			"mit license",
			"permission is hereby granted, free of charge",
			"the above copyright notice and this permission notice",
		},
		"Apache-2.0": {
			"apache license",
			"version 2.0",
			"licensed under the apache license",
		},
		"GPL-3.0": {
			"gnu general public license",
			"version 3",
			"copyleft",
		},
		"GPL-2.0": {
			"gnu general public license",
			"version 2",
			"copyleft",
		},
		"BSD-3-Clause": {
			"bsd license",
			"redistribution and use in source and binary forms",
			"neither the name of",
		},
		"BSD-2-Clause": {
			"bsd license",
			"redistribution and use in source and binary forms",
			"contributors may be used to endorse",
		},
		"ISC": {
			"isc license",
			"permission to use, copy, modify",
		},
		"Unlicense": {
			"unlicense",
			"this is free and unencumbered software",
		},
	}

	bestMatch := "unknown"
	bestScore := 0

	for licenseType, keywords := range licenses {
		score := 0
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				score++
			}
		}

		if score > bestScore {
			bestScore = score
			bestMatch = licenseType
		}
	}

	// Determine confidence based on matches
	confidence := "low"
	if bestScore >= len(licenses[bestMatch]) {
		confidence = "high"
	} else if bestScore >= len(licenses[bestMatch])/2 {
		confidence = "medium"
	}

	return bestMatch, confidence
}

// findPlaceholders finds common placeholders in license text
func (c *LicenseChecker) findPlaceholders(content string) []string {
	var placeholders []string

	commonPlaceholders := []string{
		"[year]", "[yyyy]", "<year>", "{year}",
		"[name]", "[fullname]", "<name>", "{name}", "[copyright holder]",
		"[author]", "<author>", "{author}",
		"[project]", "<project>", "{project}",
		"[organization]", "<organization>", "{organization}",
	}

	contentLower := strings.ToLower(content)
	for _, placeholder := range commonPlaceholders {
		if strings.Contains(contentLower, placeholder) {
			placeholders = append(placeholders, placeholder)
		}
	}

	return placeholders
}

// hasCopyrightNotice checks if the license has a copyright notice
func (c *LicenseChecker) hasCopyrightNotice(content string) bool {
	contentLower := strings.ToLower(content)
	return strings.Contains(contentLower, "copyright") || strings.Contains(contentLower, "©")
}

// calculateLicenseScore calculates a score based on license quality
func (c *LicenseChecker) calculateLicenseScore(licenseType, confidence string, issueCount int) int {
	score := 50 // Base score for having a license

	// Bonus for recognized license types
	if licenseType != "unknown" {
		score += 30

		// Additional bonus for high confidence detection
		switch confidence {
		case "high":
			score += 20
		case "medium":
			score += 10
		}
	}

	// Penalty for issues
	score -= issueCount * 10

	// Ensure score is within bounds
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// SupportsRepository checks if this checker supports the repository
func (c *LicenseChecker) SupportsRepository(repo core.Repository) bool {
	// This checker supports all repositories
	return true
}
