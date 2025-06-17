package dependencies

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

// OutdatedChecker checks for outdated dependencies
type OutdatedChecker struct {
	*base.BaseChecker
	executor commands.CommandExecutor
}

// NewOutdatedChecker creates a new outdated dependencies checker
func NewOutdatedChecker(executor commands.CommandExecutor) *OutdatedChecker {
	config := core.CheckerConfig{
		Enabled:    true,
		Severity:   "medium",
		Timeout:    60 * time.Second,
		Categories: []string{"dependencies"},
	}

	return &OutdatedChecker{
		BaseChecker: base.NewBaseChecker(
			"dependencies-outdated",
			"Outdated Dependencies",
			"dependencies",
			config,
		),
		executor: executor,
	}
}

// Check performs the outdated dependencies check
func (c *OutdatedChecker) Check(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	return c.Execute(ctx, repoCtx, func() (core.CheckResult, error) {
		return c.checkOutdatedDependencies(ctx, repoCtx)
	})
}

// checkOutdatedDependencies performs the actual outdated dependencies check
func (c *OutdatedChecker) checkOutdatedDependencies(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	builder := base.NewResultBuilder(c.ID(), c.Name(), c.Category())

	// Find dependency files
	depFiles := c.findDependencyFiles(repoCtx.Repository.Path)
	builder.AddMetric("dependency_files_found", len(depFiles))

	if len(depFiles) == 0 {
		builder.WithStatus(core.StatusHealthy)
		builder.WithScore(100, 100)
		builder.AddMetric("status", "no_dependencies")
		return builder.Build(), nil
	}

	// Add found files to metrics
	for i, file := range depFiles {
		builder.AddMetric(fmt.Sprintf("dependency_file_%d", i), file)
	}

	// Check dependencies by project type
	return c.checkDependenciesByType(ctx, repoCtx, builder, depFiles)
}

// findDependencyFiles finds dependency files in the repository
func (c *OutdatedChecker) findDependencyFiles(repoPath string) []string {
	depFiles := []string{
		"go.mod", "package.json", "requirements.txt", "pyproject.toml",
		"Gemfile", "Cargo.toml", "pom.xml", "build.gradle", "build.gradle.kts",
		"composer.json", "Package.swift",
	}
	var foundFiles []string

	for _, file := range depFiles {
		if _, err := os.Stat(filepath.Join(repoPath, file)); err == nil {
			foundFiles = append(foundFiles, file)
		}
	}
	return foundFiles
}

// checkDependenciesByType checks dependencies based on project type
func (c *OutdatedChecker) checkDependenciesByType(ctx context.Context, repoCtx core.RepositoryContext, builder *base.ResultBuilder, foundFiles []string) (core.CheckResult, error) {
	repoPath := repoCtx.Repository.Path

	// Check Go dependencies
	if c.contains(foundFiles, "go.mod") {
		return c.checkGoMod(ctx, repoPath, builder)
	}

	// Check Node.js dependencies
	if c.contains(foundFiles, "package.json") {
		return c.checkPackageJSON(ctx, repoPath, builder)
	}

	// Check Python dependencies
	if c.contains(foundFiles, "requirements.txt") || c.contains(foundFiles, "pyproject.toml") {
		return c.checkPythonDependencies(ctx, repoPath, builder)
	}

	// Check Java dependencies
	if c.contains(foundFiles, "pom.xml") {
		return c.checkMavenPom(ctx, repoPath, builder)
	}

	if c.contains(foundFiles, "build.gradle") || c.contains(foundFiles, "build.gradle.kts") {
		return c.checkGradleBuild(ctx, repoPath, builder)
	}

	// Generic handling for unsupported types
	builder.WithStatus(core.StatusWarning)
	builder.WithScore(60, 100)
	builder.AddIssue(base.NewIssueWithSuggestion(
		"unsupported_dependency_type",
		core.SeverityMedium,
		fmt.Sprintf("Dependency checking not implemented for: %s", strings.Join(foundFiles, ", ")),
		"Consider implementing dependency checking for this project type",
	))

	return builder.Build(), nil
}

// checkGoMod checks Go module dependencies
func (c *OutdatedChecker) checkGoMod(ctx context.Context, repoPath string, builder *base.ResultBuilder) (core.CheckResult, error) {
	builder.AddMetric("project_type", "go")

	// Check if go mod tidy would make changes
	result := c.executor.ExecuteInDir(ctx, repoPath, "go", "list", "-u", "-m", "all")
	if result.Error != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddWarning(core.Warning{
			Type:    "go_command_error",
			Message: fmt.Sprintf("Unable to check Go dependencies: %v", result.Error),
		})
		return builder.Build(), nil
	}

	// Parse output for outdated dependencies
	outdatedDeps := c.parseGoListOutput(result.Stdout)
	builder.AddMetric("outdated_dependencies", len(outdatedDeps))

	if len(outdatedDeps) == 0 {
		builder.WithStatus(core.StatusHealthy)
		builder.WithScore(100, 100)
		builder.AddMetric("status", "up_to_date")
	} else {
		builder.WithStatus(core.StatusWarning)
		builder.WithScore(70, 100)
		builder.AddMetric("status", "outdated_found")

		builder.AddIssue(base.NewIssueWithSuggestion(
			"outdated_go_dependencies",
			core.SeverityMedium,
			fmt.Sprintf("Found %d outdated Go dependencies", len(outdatedDeps)),
			"Run 'go get -u ./...' to update dependencies, then 'go mod tidy'",
		))

		// Add details about outdated dependencies
		for i, dep := range outdatedDeps {
			if i >= 5 { // Limit to first 5
				builder.AddMetric("additional_outdated", len(outdatedDeps)-5)
				break
			}
			builder.AddMetric(fmt.Sprintf("outdated_%d", i), dep)
		}
	}

	return builder.Build(), nil
}

// parseGoListOutput parses go list -u -m all output for outdated dependencies
func (c *OutdatedChecker) parseGoListOutput(output string) []string {
	var outdated []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			// This indicates an available update
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				outdated = append(outdated, fmt.Sprintf("%s (current: %s)", parts[0], parts[1]))
			}
		}
	}

	return outdated
}

// checkPackageJSON checks Node.js package dependencies
func (c *OutdatedChecker) checkPackageJSON(ctx context.Context, repoPath string, builder *base.ResultBuilder) (core.CheckResult, error) {
	builder.AddMetric("project_type", "node")

	// Check if npm is available
	result := c.executor.Execute(ctx, "which", "npm")
	if result.Error != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddIssue(base.NewIssueWithSuggestion(
			"npm_not_available",
			core.SeverityMedium,
			"npm not available for dependency checking",
			"Install Node.js and npm to enable dependency checking",
		))
		return builder.Build(), nil
	}

	// Run npm outdated
	result = c.executor.ExecuteInDir(ctx, repoPath, "npm", "outdated", "--json")
	builder.AddMetric("npm_outdated_exit_code", result.ExitCode)

	if result.ExitCode == 0 {
		builder.WithStatus(core.StatusHealthy)
		builder.WithScore(100, 100)
		builder.AddMetric("status", "up_to_date")
		builder.AddMetric("outdated_packages", 0)
	} else {
		// npm outdated returns non-zero when outdated packages are found
		builder.WithStatus(core.StatusWarning)
		builder.WithScore(70, 100)
		builder.AddMetric("status", "outdated_found")

		// Count outdated packages from JSON output
		outdatedCount := c.countNpmOutdated(result.Stdout)
		builder.AddMetric("outdated_packages", outdatedCount)

		builder.AddIssue(base.NewIssueWithSuggestion(
			"outdated_npm_packages",
			core.SeverityMedium,
			fmt.Sprintf("Found %d outdated npm packages", outdatedCount),
			"Run 'npm update' to update packages or 'npm outdated' to see details",
		))
	}

	return builder.Build(), nil
}

// countNpmOutdated counts outdated packages from npm outdated JSON output
func (c *OutdatedChecker) countNpmOutdated(jsonOutput string) int {
	// Simple count based on JSON structure - this could be improved with proper JSON parsing
	if strings.TrimSpace(jsonOutput) == "{}" || strings.TrimSpace(jsonOutput) == "" {
		return 0
	}

	// Count the number of package entries in the JSON
	count := strings.Count(jsonOutput, "\"current\":")
	return count
}

// checkPythonDependencies checks Python dependencies
func (c *OutdatedChecker) checkPythonDependencies(ctx context.Context, repoPath string, builder *base.ResultBuilder) (core.CheckResult, error) {
	builder.AddMetric("project_type", "python")

	// Check if pip is available
	result := c.executor.Execute(ctx, "which", "pip")
	if result.Error != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddIssue(base.NewIssueWithSuggestion(
			"pip_not_available",
			core.SeverityMedium,
			"pip not available for dependency checking",
			"Install Python and pip to enable dependency checking",
		))
		return builder.Build(), nil
	}

	// Run pip list --outdated
	result = c.executor.ExecuteInDir(ctx, repoPath, "pip", "list", "--outdated")
	if result.Error != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddWarning(core.Warning{
			Type:    "pip_command_error",
			Message: fmt.Sprintf("Unable to check Python dependencies: %v", result.Error),
		})
		return builder.Build(), nil
	}

	outdatedPackages := c.parsePipOutdated(result.Stdout)
	builder.AddMetric("outdated_packages", len(outdatedPackages))

	if len(outdatedPackages) == 0 {
		builder.WithStatus(core.StatusHealthy)
		builder.WithScore(100, 100)
		builder.AddMetric("status", "up_to_date")
	} else {
		builder.WithStatus(core.StatusWarning)
		builder.WithScore(70, 100)
		builder.AddMetric("status", "outdated_found")

		builder.AddIssue(base.NewIssueWithSuggestion(
			"outdated_python_packages",
			core.SeverityMedium,
			fmt.Sprintf("Found %d outdated Python packages", len(outdatedPackages)),
			"Update packages using pip install --upgrade or update requirements.txt",
		))
	}

	return builder.Build(), nil
}

// parsePipOutdated parses pip list --outdated output
func (c *OutdatedChecker) parsePipOutdated(output string) []string {
	var outdated []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Package") || strings.HasPrefix(line, "---") {
			continue
		}
		if strings.Contains(line, " ") {
			// Valid package line
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				outdated = append(outdated, fmt.Sprintf("%s (current: %s, latest: %s)", parts[0], parts[1], parts[2]))
			}
		}
	}

	return outdated
}

// checkMavenPom checks Maven dependencies
func (c *OutdatedChecker) checkMavenPom(ctx context.Context, repoPath string, builder *base.ResultBuilder) (core.CheckResult, error) {
	builder.AddMetric("project_type", "maven")

	// Check if mvn is available
	result := c.executor.Execute(ctx, "which", "mvn")
	if result.Error != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddIssue(base.NewIssueWithSuggestion(
			"maven_not_available",
			core.SeverityMedium,
			"Maven not available for dependency checking",
			"Install Maven to enable dependency checking",
		))
		return builder.Build(), nil
	}

	// Run mvn versions:display-dependency-updates
	result = c.executor.ExecuteInDir(ctx, repoPath, "mvn", "versions:display-dependency-updates", "-q")
	if result.Error != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddWarning(core.Warning{
			Type:    "maven_command_error",
			Message: fmt.Sprintf("Unable to check Maven dependencies: %v", result.Error),
		})
		return builder.Build(), nil
	}

	// Simple check for updates in output
	hasUpdates := strings.Contains(result.Stdout, "The following dependencies in Dependencies have newer versions:")
	builder.AddMetric("has_outdated_dependencies", hasUpdates)

	if !hasUpdates {
		builder.WithStatus(core.StatusHealthy)
		builder.WithScore(100, 100)
		builder.AddMetric("status", "up_to_date")
	} else {
		builder.WithStatus(core.StatusWarning)
		builder.WithScore(70, 100)
		builder.AddMetric("status", "outdated_found")

		builder.AddIssue(base.NewIssueWithSuggestion(
			"outdated_maven_dependencies",
			core.SeverityMedium,
			"Maven dependencies have newer versions available",
			"Run 'mvn versions:display-dependency-updates' to see details and consider updating",
		))
	}

	return builder.Build(), nil
}

// checkGradleBuild checks Gradle dependencies
func (c *OutdatedChecker) checkGradleBuild(ctx context.Context, repoPath string, builder *base.ResultBuilder) (core.CheckResult, error) {
	builder.AddMetric("project_type", "gradle")

	// Check if gradle is available
	result := c.executor.Execute(ctx, "which", "gradle")
	if result.Error != nil {
		// Try gradlew
		if _, err := os.Stat(filepath.Join(repoPath, "gradlew")); err == nil {
			return c.checkGradleWrapper(ctx, repoPath, builder)
		} else {
			builder.WithStatus(core.StatusWarning)
			builder.AddIssue(base.NewIssueWithSuggestion(
				"gradle_not_available",
				core.SeverityMedium,
				"Gradle not available for dependency checking",
				"Install Gradle or use Gradle wrapper to enable dependency checking",
			))
			return builder.Build(), nil
		}
	}

	// Run gradle dependencies --configuration compileClasspath
	result = c.executor.ExecuteInDir(ctx, repoPath, "gradle", "dependencies", "--configuration", "compileClasspath")
	if result.Error != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddWarning(core.Warning{
			Type:    "gradle_command_error",
			Message: fmt.Sprintf("Unable to check Gradle dependencies: %v", result.Error),
		})
		return builder.Build(), nil
	}

	// This is a simplified check - a more sophisticated implementation would parse the dependency tree
	builder.WithStatus(core.StatusHealthy)
	builder.WithScore(80, 100)
	builder.AddMetric("status", "checked")
	builder.AddMetric("gradle_dependencies_checked", true)

	return builder.Build(), nil
}

// checkGradleWrapper checks dependencies using Gradle wrapper
func (c *OutdatedChecker) checkGradleWrapper(ctx context.Context, repoPath string, builder *base.ResultBuilder) (core.CheckResult, error) {
	result := c.executor.ExecuteInDir(ctx, repoPath, "./gradlew", "dependencies", "--configuration", "compileClasspath")
	if result.Error != nil {
		builder.WithStatus(core.StatusWarning)
		builder.AddWarning(core.Warning{
			Type:    "gradlew_command_error",
			Message: fmt.Sprintf("Unable to check Gradle dependencies with wrapper: %v", result.Error),
		})
		return builder.Build(), nil
	}

	builder.WithStatus(core.StatusHealthy)
	builder.WithScore(80, 100)
	builder.AddMetric("status", "checked_with_wrapper")
	builder.AddMetric("gradle_dependencies_checked", true)

	return builder.Build(), nil
}

// contains checks if a slice contains a string
func (c *OutdatedChecker) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SupportsRepository checks if this checker supports the repository
func (c *OutdatedChecker) SupportsRepository(repo core.Repository) bool {
	// Check if there are any dependency files
	depFiles := c.findDependencyFiles(repo.Path)
	return len(depFiles) > 0
}
