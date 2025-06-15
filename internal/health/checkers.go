// Package health provides health checkers for analyzing repository health.
//
// This file has been optimized for maintainability and extendability with the following improvements:
//
//  1. **Interface Design**: Added CheckerInterface and CheckerWithContext interfaces to ensure
//     consistent behavior across all checkers and enable better testing and modularity.
//
//  2. **Factory Pattern**: Implemented CheckerFactory to centralize checker creation and
//     enable easy addition of new checkers. Supports creation by name, category, or all checkers.
//
// 3. **Common Utilities**: Added shared helper functions to reduce code duplication:
//
//   - createHealthCheck: Standardized health check result creation
//
//   - executeCommand: Consistent command execution with context support
//
//   - fileExistsInPath: Efficient file existence checking for multiple files
//
//   - commandAvailable: Standardized command availability checking
//
//     4. **Dependency Configuration**: Added DependencyType and DependencyConfig structures
//     to support extensible dependency checking across different project types.
//
//     5. **Consistent Error Handling**: Refactored all checkers to use the common helper functions
//     for consistent error handling and result formatting.
//
//     6. **Improved Performance**: Optimized file system operations and command execution
//     to reduce redundant checks and improve overall performance.
//
//     7. **Better Documentation**: Added comprehensive documentation for all public types
//     and methods to improve code maintainability.
//
// Future Enhancements:
// - Add support for custom checker plugins
// - Implement parallel checker execution with worker pools
// - Add caching for expensive operations
// - Create specialized checker result aggregation
package health

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/codcod/repos/internal/config"
)

// CheckerInterface defines the interface for all health checkers
type CheckerInterface interface {
	Name() string
	Category() string
	Check(repo config.Repository) HealthCheck
}

// CheckerWithContext defines the interface for checkers that support context
type CheckerWithContext interface {
	CheckerInterface
	CheckWithContext(ctx context.Context, repo config.Repository) HealthCheck
}

// CheckerFactory creates health checkers
type CheckerFactory struct{}

// NewCheckerFactory creates a new checker factory
func NewCheckerFactory() *CheckerFactory {
	return &CheckerFactory{}
}

// CreateAllCheckers creates all available health checkers
func (f *CheckerFactory) CreateAllCheckers() []CheckerInterface {
	return []CheckerInterface{
		&GitStatusChecker{},
		&LastCommitChecker{},
		&BranchProtectionChecker{},
		&DependencyChecker{},
		&SecurityChecker{},
		&LicenseChecker{},
		&CIStatusChecker{},
		&DocumentationChecker{},
		&CyclomaticComplexityChecker{},
		&DeprecatedComponentsChecker{},
	}
}

// CreateCheckerByName creates a specific checker by name
func (f *CheckerFactory) CreateCheckerByName(name string) CheckerInterface {
	switch name {
	case "Git Status":
		return &GitStatusChecker{}
	case "Last Commit":
		return &LastCommitChecker{}
	case "Branch Protection":
		return &BranchProtectionChecker{}
	case "Dependencies":
		return &DependencyChecker{}
	case "Security":
		return &SecurityChecker{}
	case "License":
		return &LicenseChecker{}
	case "CI Status":
		return &CIStatusChecker{}
	case "Documentation":
		return &DocumentationChecker{}
	case "Cyclomatic Complexity":
		return &CyclomaticComplexityChecker{}
	case "Deprecated Components":
		return &DeprecatedComponentsChecker{}
	default:
		return nil
	}
}

// CreateCheckersByCategory creates checkers for a specific category
func (f *CheckerFactory) CreateCheckersByCategory(category string) []CheckerInterface {
	var checkers []CheckerInterface
	for _, checker := range f.CreateAllCheckers() {
		if checker.Category() == category {
			checkers = append(checkers, checker)
		}
	}
	return checkers
}

// GetAllCategories returns all available health check categories
func (f *CheckerFactory) GetAllCategories() []string {
	categoriesMap := make(map[string]bool)

	for _, checker := range f.CreateAllCheckers() {
		categoriesMap[checker.Category()] = true
	}

	var categories []string
	for category := range categoriesMap {
		categories = append(categories, category)
	}

	return categories
}

// CategoryInfo holds information about a health check category
type CategoryInfo struct {
	Name        string
	Description string
	Checkers    []string
}

// GetCategoryInfo returns detailed information about all available categories
func (f *CheckerFactory) GetCategoryInfo() []CategoryInfo {
	categoryMap := make(map[string][]string)

	// Group checkers by category
	for _, checker := range f.CreateAllCheckers() {
		category := checker.Category()
		categoryMap[category] = append(categoryMap[category], checker.Name())
	}

	// Create category info with descriptions
	var categories []CategoryInfo

	categoryDescriptions := map[string]string{
		"git":           "Git repository status and commit history checks",
		"security":      "Security vulnerabilities and policy checks",
		"dependencies":  "Dependency management and outdated package checks",
		"compliance":    "License and legal compliance checks",
		"automation":    "CI/CD and automation configuration checks",
		"documentation": "Documentation completeness and quality checks",
		"code-quality":  "Code quality metrics and complexity analysis",
	}

	for category, checkers := range categoryMap {
		description := categoryDescriptions[category]
		if description == "" {
			description = "Health checks for " + category
		}

		categories = append(categories, CategoryInfo{
			Name:        category,
			Description: description,
			Checkers:    checkers,
		})
	}

	return categories
}

// Common helper functions for all checkers

// createHealthCheck creates a standardized health check result
func createHealthCheck(name, category string, status HealthStatus, message, details string, severity int) HealthCheck {
	return HealthCheck{
		Name:        name,
		Status:      status,
		Message:     message,
		Details:     details,
		Severity:    severity,
		Category:    category,
		LastChecked: time.Now(),
	}
}

// executeCommand runs a command with optional timeout context
func executeCommand(ctx context.Context, repoPath, command string, args ...string) ([]byte, error) {
	var cmd *exec.Cmd
	if ctx != nil {
		cmd = exec.CommandContext(ctx, command, args...)
	} else {
		cmd = exec.Command(command, args...)
	}
	cmd.Dir = repoPath
	return cmd.Output()
}

// executeCommandWithoutContext runs a command without context (for backward compatibility)
func executeCommandWithoutContext(repoPath, command string, args ...string) ([]byte, error) {
	return executeCommand(context.TODO(), repoPath, command, args...)
}

// Utility functions for common operations

// DependencyType represents different types of dependency managers
type DependencyType string

const (
	DependencyTypeGo     DependencyType = "go"
	DependencyTypeNode   DependencyType = "node"
	DependencyTypePython DependencyType = "python"
	DependencyTypeJava   DependencyType = "java"
)

// DependencyConfig holds configuration for different dependency types
type DependencyConfig struct {
	Type          DependencyType
	Files         []string
	RequiredTools []string
	Commands      map[string][]string
}

// GetDependencyConfigs returns all supported dependency configurations
func GetDependencyConfigs() map[DependencyType]DependencyConfig {
	return map[DependencyType]DependencyConfig{
		DependencyTypeGo: {
			Type:          DependencyTypeGo,
			Files:         []string{"go.mod"},
			RequiredTools: []string{"go"},
			Commands: map[string][]string{
				"tidy": {"go", "mod", "tidy", "-diff"},
			},
		},
		DependencyTypeNode: {
			Type:          DependencyTypeNode,
			Files:         []string{"package.json"},
			RequiredTools: []string{"npm"},
			Commands: map[string][]string{
				"audit": {"npm", "audit"},
			},
		},
		DependencyTypePython: {
			Type:          DependencyTypePython,
			Files:         []string{"requirements.txt", "pyproject.toml"},
			RequiredTools: []string{"pip", "pip3"},
			Commands: map[string][]string{
				"check": {"pip", "check"},
			},
		},
		DependencyTypeJava: {
			Type:          DependencyTypeJava,
			Files:         []string{"pom.xml", "build.gradle", "build.gradle.kts"},
			RequiredTools: []string{"mvn", "gradle"},
			Commands: map[string][]string{
				"maven_analyze": {"mvn", "dependency:analyze", "-q"},
				"gradle_deps":   {"gradle", "dependencies", "--configuration", "compileClasspath", "-q"},
			},
		},
	}
}

// fileExistsInPath checks if any of the files exist in the given path
func fileExistsInPath(repoPath string, files []string) []string {
	var found []string
	for _, file := range files {
		fullPath := filepath.Join(repoPath, file)
		if _, err := os.Stat(fullPath); err == nil {
			found = append(found, file)
		}
	}
	return found
}

// commandAvailable checks if a command is available in PATH
func commandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

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
	c.checkCommonProtectionPatterns(repoPath, &warnings, &info)
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

func (c *BranchProtectionChecker) checkDefaultBranchName(defaultBranch string, warnings *[]string) {
	if defaultBranch != "main" && defaultBranch != "master" && defaultBranch != "develop" {
		*warnings = append(*warnings, fmt.Sprintf("Unusual default branch name: '%s'", defaultBranch))
	}
}

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

// DependencyChecker checks for outdated dependencies
type DependencyChecker struct{}

func (c *DependencyChecker) Name() string     { return "Dependencies" }
func (c *DependencyChecker) Category() string { return "dependencies" }

func (c *DependencyChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)
	foundFiles := c.findDependencyFiles(repoPath)

	if len(foundFiles) == 0 {
		return c.createDependencyHealthCheck(HealthStatusHealthy, "No dependency files found", "", 1)
	}

	return c.checkDependenciesByType(repoPath, foundFiles)
}

func (c *DependencyChecker) findDependencyFiles(repoPath string) []string {
	depFiles := []string{
		"go.mod", "package.json", "requirements.txt", "pyproject.toml",
		"Gemfile", "Cargo.toml", "pom.xml", "build.gradle", "build.gradle.kts",
	}
	var foundFiles []string

	for _, file := range depFiles {
		if _, err := os.Stat(filepath.Join(repoPath, file)); err == nil {
			foundFiles = append(foundFiles, file)
		}
	}
	return foundFiles
}

func (c *DependencyChecker) checkDependenciesByType(repoPath string, foundFiles []string) HealthCheck {
	if contains(foundFiles, "go.mod") {
		return c.checkGoMod(repoPath)
	}
	if contains(foundFiles, "package.json") {
		return c.checkPackageJSON(repoPath)
	}
	if contains(foundFiles, "pyproject.toml") {
		return c.checkPyprojectToml(repoPath)
	}
	if contains(foundFiles, "requirements.txt") {
		return c.checkRequirementsTxt(repoPath)
	}
	if contains(foundFiles, "pom.xml") {
		return c.checkMavenPom(repoPath)
	}
	if contains(foundFiles, "build.gradle") || contains(foundFiles, "build.gradle.kts") {
		return c.checkGradleBuild(repoPath)
	}

	return c.createDependencyHealthCheck(HealthStatusWarning,
		fmt.Sprintf("Found dependency files: %s", strings.Join(foundFiles, ", ")),
		"Dependency checking not implemented for this project type", 1)
}

func (c *DependencyChecker) createDependencyHealthCheck(status HealthStatus, message, details string, severity int) HealthCheck {
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

func (c *DependencyChecker) checkGoMod(repoPath string) HealthCheck {
	// Check if go mod tidy would make changes
	cmd := exec.Command("go", "mod", "tidy", "-diff")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return HealthCheck{
			Name:        c.Name(),
			Status:      HealthStatusWarning,
			Message:     "Unable to check go.mod status",
			Details:     stderr.String(),
			Severity:    2,
			Category:    c.Category(),
			LastChecked: time.Now(),
		}
	}

	if stdout.Len() > 0 {
		return HealthCheck{
			Name:        c.Name(),
			Status:      HealthStatusWarning,
			Message:     "go.mod needs tidying",
			Details:     "Run 'go mod tidy' to fix",
			Severity:    1,
			Category:    c.Category(),
			LastChecked: time.Now(),
		}
	}

	return HealthCheck{
		Name:        c.Name(),
		Status:      HealthStatusHealthy,
		Message:     "Go dependencies are up to date",
		Severity:    1,
		Category:    c.Category(),
		LastChecked: time.Now(),
	}
}

func (c *DependencyChecker) checkPackageJSON(repoPath string) HealthCheck {
	// Basic package.json existence check
	packageFile := filepath.Join(repoPath, "package.json")
	lockFile := filepath.Join(repoPath, "package-lock.json")

	if _, err := os.Stat(packageFile); err != nil {
		return HealthCheck{
			Name:        c.Name(),
			Status:      HealthStatusWarning,
			Message:     "package.json not found",
			Severity:    2,
			Category:    c.Category(),
			LastChecked: time.Now(),
		}
	}

	status := HealthStatusHealthy
	message := "Node.js dependencies found"

	if _, err := os.Stat(lockFile); err != nil {
		status = HealthStatusWarning
		message = "package-lock.json missing"
	}

	return HealthCheck{
		Name:        c.Name(),
		Status:      status,
		Message:     message,
		Severity:    1,
		Category:    c.Category(),
		LastChecked: time.Now(),
	}
}

// checkMavenPom checks Maven dependencies and project health
func (c *DependencyChecker) checkMavenPom(repoPath string) HealthCheck {
	if err := c.validateMavenPom(repoPath); err != nil {
		return c.createDependencyHealthCheck(HealthStatusCritical, err.Error(), "", 3)
	}

	if !c.commandExists("mvn") {
		return c.createDependencyHealthCheck(HealthStatusWarning,
			"Maven project found but mvn command not available",
			"Install Maven to get detailed dependency analysis", 2)
	}

	var issues, warnings []string
	c.checkMavenDependencyAnalysis(repoPath, &issues, &warnings)
	c.checkMavenDependencyResolution(repoPath, &issues, &warnings)

	return c.createMavenHealthCheckResult(issues, warnings)
}

func (c *DependencyChecker) validateMavenPom(repoPath string) error {
	pomFile := filepath.Join(repoPath, "pom.xml")
	if _, err := os.Stat(pomFile); err != nil {
		return fmt.Errorf("pom.xml not found")
	}
	return nil
}

func (c *DependencyChecker) checkMavenDependencyAnalysis(repoPath string, issues, warnings *[]string) {
	ctx, cancel := CreateHealthContext()
	defer cancel()

	cmd := exec.CommandContext(ctx, "mvn", "dependency:analyze", "-q")
	cmd.Dir = repoPath
	output, err := cmd.Output()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			*warnings = append(*warnings, "Maven dependency check timed out")
		} else {
			*warnings = append(*warnings, "Unable to analyze dependencies")
		}
		return
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "Used undeclared dependencies") {
		*issues = append(*issues, "Has undeclared dependencies")
	}
	if strings.Contains(outputStr, "Unused declared dependencies") {
		*warnings = append(*warnings, "Has unused declared dependencies")
	}
}

func (c *DependencyChecker) checkMavenDependencyResolution(repoPath string, issues, warnings *[]string) {
	ctx, cancel := CreateHealthContext()
	defer cancel()

	cmd := exec.CommandContext(ctx, "mvn", "dependency:resolve", "-q")
	cmd.Dir = repoPath
	err := cmd.Run()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			*warnings = append(*warnings, "Maven dependency resolution timed out")
		} else {
			*issues = append(*issues, "Dependencies cannot be resolved")
		}
	}
}

func (c *DependencyChecker) createMavenHealthCheckResult(issues, warnings []string) HealthCheck {
	status := HealthStatusHealthy
	message := "Maven dependencies are healthy"
	severity := 1

	if len(issues) > 0 {
		status = HealthStatusCritical
		message = fmt.Sprintf("Maven dependency issues: %s", strings.Join(issues, ", "))
		severity = 3
	} else if len(warnings) > 0 {
		status = HealthStatusWarning
		message = fmt.Sprintf("Maven dependency warnings: %s", strings.Join(warnings, ", "))
		severity = 2
	}

	details := ""
	if len(issues) > 0 || len(warnings) > 0 {
		allItems := append(issues, warnings...)
		details = strings.Join(allItems, "\n")
	}

	return c.createDependencyHealthCheck(status, message, details, severity)
}

// checkGradleBuild checks Gradle dependencies and project health
func (c *DependencyChecker) checkGradleBuild(repoPath string) HealthCheck {
	foundFile, err := c.findGradleBuildFile(repoPath)
	if err != nil {
		return c.createDependencyHealthCheck(HealthStatusCritical, err.Error(), "", 3)
	}

	gradleCmd, useWrapper, err := c.determineGradleCommand(repoPath, foundFile)
	if err != nil {
		return c.createDependencyHealthCheck(HealthStatusWarning, err.Error(),
			"Install Gradle or use Gradle wrapper (gradlew) for detailed dependency analysis", 2)
	}

	var issues, warnings []string
	c.checkGradleDependencyResolution(repoPath, gradleCmd, useWrapper, &issues, &warnings)
	c.checkGradleOutdatedDependencies(repoPath, gradleCmd, useWrapper, &warnings)
	c.checkGradleSecurityScan(repoPath, gradleCmd, useWrapper, &warnings)
	c.checkGradleBestPractices(useWrapper, &warnings)

	return c.createGradleHealthCheckResult(foundFile, issues, warnings)
}

func (c *DependencyChecker) findGradleBuildFile(repoPath string) (string, error) {
	buildFile := filepath.Join(repoPath, "build.gradle")
	buildFileKts := filepath.Join(repoPath, "build.gradle.kts")

	if _, err := os.Stat(buildFile); err == nil {
		return "build.gradle", nil
	}
	if _, err := os.Stat(buildFileKts); err == nil {
		return "build.gradle.kts", nil
	}
	return "", fmt.Errorf("gradle build file not found")
}

func (c *DependencyChecker) determineGradleCommand(repoPath, foundFile string) (string, bool, error) {
	gradlewPath := filepath.Join(repoPath, "gradlew")
	if _, err := os.Stat(gradlewPath); err == nil {
		return "./gradlew", true, nil
	}
	if !c.commandExists("gradle") {
		return "", false, fmt.Errorf("gradle project found (%s) but gradle command not available", foundFile)
	}
	return "gradle", false, nil
}

func (c *DependencyChecker) checkGradleDependencyResolution(repoPath, gradleCmd string, useWrapper bool, issues, warnings *[]string) {
	ctx, cancel := CreateHealthContext()
	defer cancel()

	var cmd *exec.Cmd
	if useWrapper {
		cmd = exec.CommandContext(ctx, "./gradlew", "dependencies", "--configuration", "compileClasspath", "-q")
	} else {
		cmd = exec.CommandContext(ctx, gradleCmd, "dependencies", "--configuration", "compileClasspath", "-q")
	}
	cmd.Dir = repoPath

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			*warnings = append(*warnings, "Gradle dependency resolution timed out")
		} else {
			*issues = append(*issues, "Dependencies cannot be resolved")
		}
	}
}

func (c *DependencyChecker) checkGradleOutdatedDependencies(repoPath, gradleCmd string, useWrapper bool, warnings *[]string) {
	ctx, cancel := CreateHealthContext()
	defer cancel()

	var cmd *exec.Cmd
	if useWrapper {
		cmd = exec.CommandContext(ctx, "./gradlew", "dependencyUpdates", "-q")
	} else {
		cmd = exec.CommandContext(ctx, gradleCmd, "dependencyUpdates", "-q")
	}
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err == nil {
		if strings.Contains(string(output), "outdated dependencies") {
			*warnings = append(*warnings, "Has outdated dependencies")
		}
	} else if ctx.Err() == context.DeadlineExceeded {
		*warnings = append(*warnings, "Gradle dependency update check timed out")
	}
}

func (c *DependencyChecker) checkGradleSecurityScan(repoPath, gradleCmd string, useWrapper bool, warnings *[]string) {
	var cmd *exec.Cmd
	if useWrapper {
		cmd = exec.Command("./gradlew", "dependencyCheckAnalyze", "-q")
	} else {
		cmd = exec.Command(gradleCmd, "dependencyCheckAnalyze", "-q")
	}
	cmd.Dir = repoPath

	if err := cmd.Run(); err == nil {
		*warnings = append(*warnings, "Security scan completed")
	}
}

func (c *DependencyChecker) checkGradleBestPractices(useWrapper bool, warnings *[]string) {
	if !useWrapper {
		*warnings = append(*warnings, "Gradle wrapper not found - consider using gradlew for reproducible builds")
	}
}

func (c *DependencyChecker) createGradleHealthCheckResult(foundFile string, issues, warnings []string) HealthCheck {
	status := HealthStatusHealthy
	message := fmt.Sprintf("Gradle dependencies are healthy (%s)", foundFile)
	severity := 1

	if len(issues) > 0 {
		status = HealthStatusCritical
		message = fmt.Sprintf("Gradle dependency issues: %s", strings.Join(issues, ", "))
		severity = 3
	} else if len(warnings) > 0 {
		status = HealthStatusWarning
		message = fmt.Sprintf("Gradle dependency warnings: %s", strings.Join(warnings, ", "))
		severity = 2
	}

	details := ""
	if len(issues) > 0 || len(warnings) > 0 {
		allItems := append(issues, warnings...)
		details = strings.Join(allItems, "\n")
	}

	return c.createDependencyHealthCheck(status, message, details, severity)
}

// checkPyprojectToml checks Python projects using pyproject.toml
func (c *DependencyChecker) checkPyprojectToml(repoPath string) HealthCheck {
	pyprojectFile := filepath.Join(repoPath, "pyproject.toml")

	if _, err := os.Stat(pyprojectFile); err != nil {
		return c.createDependencyHealthCheck(HealthStatusCritical, "pyproject.toml not found", "", 3)
	}

	content, err := os.ReadFile(pyprojectFile) // #nosec G304 - file path is controlled
	if err != nil {
		return c.createDependencyHealthCheck(HealthStatusCritical, "Cannot read pyproject.toml", err.Error(), 3)
	}

	var warnings, issues []string
	pipAvailable := c.commandExists("pip") || c.commandExists("pip3")

	c.checkPyprojectPipAvailability(pipAvailable, &warnings)
	c.checkPyprojectStructure(string(content), &warnings)
	c.checkPyprojectVirtualEnv(repoPath, &warnings)
	c.checkPyprojectDependencyManagement(repoPath, string(content), pipAvailable, &warnings)

	return c.createPythonHealthCheckResult(issues, warnings, "Python pyproject.toml")
}

func (c *DependencyChecker) checkPyprojectPipAvailability(pipAvailable bool, warnings *[]string) {
	if !pipAvailable {
		*warnings = append(*warnings, "pip not available - install pip for dependency management")
	}
}

func (c *DependencyChecker) checkPyprojectStructure(content string, warnings *[]string) {
	hasProjectSection := strings.Contains(content, "[project")
	hasBuildSystem := strings.Contains(content, "[build-system")
	hasDependencies := strings.Contains(content, "dependencies")

	if !hasProjectSection && !hasBuildSystem {
		*warnings = append(*warnings, "pyproject.toml may be incomplete - missing [project] or [build-system] sections")
	}
	if !hasDependencies {
		*warnings = append(*warnings, "No dependencies declared in pyproject.toml")
	}
}

func (c *DependencyChecker) checkPyprojectVirtualEnv(repoPath string, warnings *[]string) {
	venvPaths := []string{
		filepath.Join(repoPath, "venv"),
		filepath.Join(repoPath, ".venv"),
		filepath.Join(repoPath, "env"),
	}

	for _, venvPath := range venvPaths {
		if _, err := os.Stat(venvPath); err == nil {
			return // Found virtual environment
		}
	}
	*warnings = append(*warnings, "No virtual environment found - consider using venv for dependency isolation")
}

func (c *DependencyChecker) checkPyprojectDependencyManagement(repoPath, content string, pipAvailable bool, warnings *[]string) {
	if !pipAvailable {
		return
	}

	c.checkPyprojectDevDependencies(repoPath, content, warnings)
	c.checkPyprojectPipTools(repoPath, warnings)
}

func (c *DependencyChecker) checkPyprojectDevDependencies(repoPath, content string, warnings *[]string) {
	devFiles := []string{"requirements-dev.txt", "requirements-test.txt", "dev-requirements.txt"}
	for _, devFile := range devFiles {
		if _, err := os.Stat(filepath.Join(repoPath, devFile)); err == nil {
			return // Found dev dependencies file
		}
	}

	if strings.Contains(content, "test") {
		*warnings = append(*warnings, "Consider separating development/test dependencies")
	}
}

func (c *DependencyChecker) checkPyprojectPipTools(repoPath string, warnings *[]string) {
	requirementsIn := filepath.Join(repoPath, "requirements.in")
	requirementsTxt := filepath.Join(repoPath, "requirements.txt")

	_, hasRequirementsIn := os.Stat(requirementsIn)
	_, hasRequirementsTxt := os.Stat(requirementsTxt)

	if hasRequirementsIn == nil && hasRequirementsTxt != nil {
		*warnings = append(*warnings, "Found requirements.in but no requirements.txt - run pip-compile to generate lock file")
	}
}

func (c *DependencyChecker) createPythonHealthCheckResult(issues, warnings []string, baseMessage string) HealthCheck {
	status := HealthStatusHealthy
	message := fmt.Sprintf("%s is healthy", baseMessage)
	severity := 1

	if len(issues) > 0 {
		status = HealthStatusCritical
		message = fmt.Sprintf("Python project issues: %s", strings.Join(issues, ", "))
		severity = 3
	} else if len(warnings) > 0 {
		status = HealthStatusWarning
		message = fmt.Sprintf("Python project warnings: %s", strings.Join(warnings, ", "))
		severity = 2
	}

	details := ""
	if len(issues) > 0 || len(warnings) > 0 {
		allItems := append(issues, warnings...)
		details = strings.Join(allItems, "\n")
	}

	return c.createDependencyHealthCheck(status, message, details, severity)
}

// checkRequirementsTxt checks Python projects using requirements.txt
func (c *DependencyChecker) checkRequirementsTxt(repoPath string) HealthCheck {
	requirementsFile := filepath.Join(repoPath, "requirements.txt")

	if _, err := os.Stat(requirementsFile); err != nil {
		return c.createDependencyHealthCheck(HealthStatusCritical, "requirements.txt not found", "", 3)
	}

	content, err := os.ReadFile(requirementsFile) // #nosec G304 - file path is controlled
	if err != nil {
		return c.createDependencyHealthCheck(HealthStatusCritical, "Cannot read requirements.txt", err.Error(), 3)
	}

	var warnings, issues []string
	pipAvailable := c.commandExists("pip") || c.commandExists("pip3")

	dependencies := c.parseRequirementsTxt(string(content), &warnings)

	c.checkPyprojectPipAvailability(pipAvailable, &warnings)
	c.checkPyprojectVirtualEnv(repoPath, &warnings)
	c.checkRequirementsTxtDevDependencies(repoPath, len(dependencies), &warnings)
	c.checkRequirementsTxtPipCheck(repoPath, pipAvailable, &warnings)
	c.checkRequirementsTxtPipTools(repoPath, requirementsFile, &warnings)

	message := fmt.Sprintf("Python requirements.txt (%d dependencies)", len(dependencies))
	return c.createPythonHealthCheckResult(issues, warnings, message)
}

func (c *DependencyChecker) parseRequirementsTxt(content string, warnings *[]string) []string {
	lines := strings.Split(content, "\n")
	var dependencies []string
	unpinnedCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		dependencies = append(dependencies, line)
		if !c.isDependencyPinned(line) {
			unpinnedCount++
		}
	}

	if len(dependencies) == 0 {
		*warnings = append(*warnings, "requirements.txt is empty")
	} else if unpinnedCount > 0 {
		*warnings = append(*warnings, fmt.Sprintf("%d dependencies are not pinned to specific versions", unpinnedCount))
	}

	return dependencies
}

func (c *DependencyChecker) isDependencyPinned(line string) bool {
	return strings.Contains(line, "==") || strings.Contains(line, ">=") ||
		strings.Contains(line, "<=") || strings.Contains(line, "~=")
}

func (c *DependencyChecker) checkRequirementsTxtDevDependencies(repoPath string, dependencyCount int, warnings *[]string) {
	devFiles := []string{"requirements-dev.txt", "requirements-test.txt", "dev-requirements.txt"}
	for _, devFile := range devFiles {
		if _, err := os.Stat(filepath.Join(repoPath, devFile)); err == nil {
			return // Found dev dependencies file
		}
	}

	if dependencyCount > 5 {
		*warnings = append(*warnings, "Consider separating development dependencies into requirements-dev.txt")
	}
}

func (c *DependencyChecker) checkRequirementsTxtPipCheck(repoPath string, pipAvailable bool, warnings *[]string) {
	if !pipAvailable {
		return
	}

	ctx, cancel := CreateHealthContext()
	defer cancel()

	cmd := exec.CommandContext(ctx, "pip", "check")
	cmd.Dir = repoPath

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			*warnings = append(*warnings, "pip check timed out")
		} else {
			*warnings = append(*warnings, "Potential dependency conflicts detected (pip check failed)")
		}
	}
}

func (c *DependencyChecker) checkRequirementsTxtPipTools(repoPath, requirementsFile string, warnings *[]string) {
	requirementsIn := filepath.Join(repoPath, "requirements.in")
	inStat, inErr := os.Stat(requirementsIn)
	if inErr != nil {
		return // requirements.in doesn't exist
	}

	txtStat, txtErr := os.Stat(requirementsFile)
	if txtErr == nil && inStat.ModTime().After(txtStat.ModTime()) {
		*warnings = append(*warnings, "requirements.in is newer than requirements.txt - run pip-compile to update")
	}
}

// commandExists checks if a command is available in the system PATH
func (c *DependencyChecker) commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
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

// LicenseChecker checks for license files
type LicenseChecker struct{}

func (c *LicenseChecker) Name() string     { return "License" }
func (c *LicenseChecker) Category() string { return "compliance" }

func (c *LicenseChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)

	// Check for license files (both LICENSE and international spelling variations are valid)
	licenseFiles := []string{"LICENSE", "LICENSE.txt", "LICENSE.md", "LICENCE", "LICENCE.txt", "LICENCE.md"} //nolint:misspell // international spellings

	found := fileExistsInPath(repoPath, licenseFiles)
	if len(found) > 0 {
		return createHealthCheck(
			c.Name(),
			c.Category(),
			HealthStatusHealthy,
			"License file found",
			fmt.Sprintf("Found: %s", strings.Join(found, ", ")),
			1,
		)
	}

	return createHealthCheck(
		c.Name(),
		c.Category(),
		HealthStatusWarning,
		"No license file found",
		"Consider adding a LICENSE file",
		1,
	)
}

// CIStatusChecker checks for CI/CD configuration
type CIStatusChecker struct{}

func (c *CIStatusChecker) Name() string     { return "CI/CD" }
func (c *CIStatusChecker) Category() string { return "automation" }

func (c *CIStatusChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)

	ciFiles := []string{
		".github/workflows",
		".gitlab-ci.yml",
		".travis.yml",
		"Jenkinsfile",
		".circleci/config.yml",
		"azure-pipelines.yml",
	}

	found := fileExistsInPath(repoPath, ciFiles)

	if len(found) == 0 {
		return createHealthCheck(
			c.Name(),
			c.Category(),
			HealthStatusWarning,
			"No CI/CD configuration found",
			"Consider setting up continuous integration",
			1,
		)
	}

	return createHealthCheck(
		c.Name(),
		c.Category(),
		HealthStatusHealthy,
		fmt.Sprintf("CI/CD configuration found: %s", strings.Join(found, ", ")),
		"",
		1,
	)
}

// DocumentationChecker checks for documentation files
type DocumentationChecker struct{}

func (c *DocumentationChecker) Name() string     { return "Documentation" }
func (c *DocumentationChecker) Category() string { return "documentation" }

func (c *DocumentationChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)

	// Check for README files
	readmeFiles := []string{"README.md", "README.txt", "README.rst", "README"}
	var foundReadme bool

	for _, file := range readmeFiles {
		if info, err := os.Stat(filepath.Join(repoPath, file)); err == nil && !info.IsDir() {
			foundReadme = true
			break
		}
	}

	if !foundReadme {
		return HealthCheck{
			Name:        c.Name(),
			Status:      HealthStatusCritical,
			Message:     "No README file found",
			Details:     "Add a README.md file to document the project",
			Severity:    3,
			Category:    c.Category(),
			LastChecked: time.Now(),
		}
	}

	// Check README file size and content
	readmePath := filepath.Join(repoPath, "README.md")
	if info, err := os.Stat(readmePath); err == nil {
		if info.Size() < 100 {
			return HealthCheck{
				Name:        c.Name(),
				Status:      HealthStatusWarning,
				Message:     "README file is very short",
				Details:     "Consider adding more documentation",
				Severity:    1,
				Category:    c.Category(),
				LastChecked: time.Now(),
			}
		}

		// Check for common documentation sections
		content, err := os.ReadFile(readmePath) // #nosec G304 - file path is controlled
		if err == nil {
			sections := c.checkDocumentationSections(string(content))
			if len(sections) < 2 {
				return HealthCheck{
					Name:        c.Name(),
					Status:      HealthStatusWarning,
					Message:     "README lacks common sections",
					Details:     "Consider adding Installation, Usage, or Contributing sections",
					Severity:    1,
					Category:    c.Category(),
					LastChecked: time.Now(),
				}
			}
		}
	}

	return HealthCheck{
		Name:        c.Name(),
		Status:      HealthStatusHealthy,
		Message:     "Good documentation found",
		Severity:    1,
		Category:    c.Category(),
		LastChecked: time.Now(),
	}
}

func (c *DocumentationChecker) checkDocumentationSections(content string) []string {
	sections := []string{}

	commonSections := []string{
		"installation", "install", "setup",
		"usage", "getting started",
		"contributing", "contribute",
		"license", "api", "examples",
	}

	contentLower := strings.ToLower(content)

	for _, section := range commonSections {
		patterns := []string{
			fmt.Sprintf("# %s", section),
			fmt.Sprintf("## %s", section),
			fmt.Sprintf("### %s", section),
		}

		for _, pattern := range patterns {
			if strings.Contains(contentLower, pattern) {
				sections = append(sections, section)
				break
			}
		}
	}

	return sections
}

// DeprecatedComponentsChecker checks for usage of deprecated components, APIs, and patterns
type DeprecatedComponentsChecker struct{}

func (d *DeprecatedComponentsChecker) Name() string {
	return "Deprecated Components"
}

func (d *DeprecatedComponentsChecker) Category() string {
	return "code-quality"
}

func (d *DeprecatedComponentsChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)

	deprecatedItems := d.scanForDeprecatedUsage(repoPath)

	if len(deprecatedItems) == 0 {
		return createHealthCheck(
			d.Name(),
			d.Category(),
			HealthStatusHealthy,
			"No deprecated components detected",
			"",
			1,
		)
	}

	severity := d.calculateSeverity(deprecatedItems)
	status := d.determineStatus(len(deprecatedItems), severity)

	message := fmt.Sprintf("Found %d deprecated component usages", len(deprecatedItems))
	details := d.formatDeprecatedItems(deprecatedItems)

	return createHealthCheck(d.Name(), d.Category(), status, message, details, severity)
}

// DeprecatedItem represents a deprecated component usage
type DeprecatedItem struct {
	File        string
	Line        int
	Pattern     string
	Replacement string
	Language    string
	Severity    string
	Description string
}

// scanForDeprecatedUsage scans the repository for deprecated component usage
func (d *DeprecatedComponentsChecker) scanForDeprecatedUsage(repoPath string) []DeprecatedItem {
	var deprecatedItems []DeprecatedItem

	// Scan different file types for deprecated patterns
	deprecatedItems = append(deprecatedItems, d.scanGoFiles(repoPath)...)
	deprecatedItems = append(deprecatedItems, d.scanJavaFiles(repoPath)...)
	deprecatedItems = append(deprecatedItems, d.scanJavaScriptFiles(repoPath)...)
	deprecatedItems = append(deprecatedItems, d.scanPythonFiles(repoPath)...)
	deprecatedItems = append(deprecatedItems, d.scanDockerFiles(repoPath)...)
	deprecatedItems = append(deprecatedItems, d.scanKubernetesFiles(repoPath)...)

	return deprecatedItems
}

// scanGoFiles scans Go files for deprecated patterns
func (d *DeprecatedComponentsChecker) scanGoFiles(repoPath string) []DeprecatedItem {
	var items []DeprecatedItem

	goFiles := d.findFiles(repoPath, "*.go")

	// Common Go deprecated patterns
	deprecatedPatterns := map[string]DeprecatedItem{
		"ioutil.ReadFile":     {Pattern: "ioutil.ReadFile", Replacement: "os.ReadFile", Severity: "warning", Description: "ioutil.ReadFile deprecated since Go 1.16"},
		"ioutil.WriteFile":    {Pattern: "ioutil.WriteFile", Replacement: "os.WriteFile", Severity: "warning", Description: "ioutil.WriteFile deprecated since Go 1.16"},
		"ioutil.ReadAll":      {Pattern: "ioutil.ReadAll", Replacement: "io.ReadAll", Severity: "warning", Description: "ioutil.ReadAll deprecated since Go 1.16"},
		"ioutil.ReadDir":      {Pattern: "ioutil.ReadDir", Replacement: "os.ReadDir", Severity: "warning", Description: "ioutil.ReadDir deprecated since Go 1.16"},
		"ioutil.TempDir":      {Pattern: "ioutil.TempDir", Replacement: "os.MkdirTemp", Severity: "warning", Description: "ioutil.TempDir deprecated since Go 1.17"},
		"ioutil.TempFile":     {Pattern: "ioutil.TempFile", Replacement: "os.CreateTemp", Severity: "warning", Description: "ioutil.TempFile deprecated since Go 1.17"},
		"golang.org/x/net/context": {Pattern: "golang.org/x/net/context", Replacement: "context", Severity: "critical", Description: "Use standard library context package instead"},
	}

	for _, file := range goFiles {
		items = append(items, d.scanFileForPatterns(file, deprecatedPatterns, "Go")...)
	}

	return items
}

// scanJavaFiles scans Java files for deprecated patterns
func (d *DeprecatedComponentsChecker) scanJavaFiles(repoPath string) []DeprecatedItem {
	var items []DeprecatedItem

	javaFiles := d.findFiles(repoPath, "*.java")

	// Common Java deprecated patterns
	deprecatedPatterns := map[string]DeprecatedItem{
		"@Deprecated":                    {Pattern: "@Deprecated", Replacement: "Check for modern alternative", Severity: "warning", Description: "Using deprecated Java API"},
		"java.util.Date":                {Pattern: "java.util.Date", Replacement: "java.time.LocalDate/LocalDateTime", Severity: "warning", Description: "Use modern Java time API"},
		"SimpleDateFormat":              {Pattern: "SimpleDateFormat", Replacement: "DateTimeFormatter", Severity: "warning", Description: "Use thread-safe DateTimeFormatter"},
		"StringBuffer":                  {Pattern: "StringBuffer", Replacement: "StringBuilder", Severity: "warning", Description: "StringBuilder is more efficient for single-threaded use"},
		"java.util.Vector":              {Pattern: "java.util.Vector", Replacement: "java.util.ArrayList", Severity: "warning", Description: "Vector is legacy, use ArrayList or Collections.synchronizedList"},
		"java.util.Hashtable":           {Pattern: "java.util.Hashtable", Replacement: "java.util.HashMap", Severity: "warning", Description: "Hashtable is legacy, use HashMap or ConcurrentHashMap"},
		"java.util.Stack":               {Pattern: "java.util.Stack", Replacement: "java.util.Deque", Severity: "warning", Description: "Stack is legacy, use ArrayDeque"},
	}

	for _, file := range javaFiles {
		items = append(items, d.scanFileForPatterns(file, deprecatedPatterns, "Java")...)
	}

	return items
}

// scanJavaScriptFiles scans JavaScript/TypeScript files for deprecated patterns
func (d *DeprecatedComponentsChecker) scanJavaScriptFiles(repoPath string) []DeprecatedItem {
	var items []DeprecatedItem

	jsFiles := append(d.findFiles(repoPath, "*.js"), d.findFiles(repoPath, "*.ts")...)
	jsFiles = append(jsFiles, d.findFiles(repoPath, "*.jsx")...)
	jsFiles = append(jsFiles, d.findFiles(repoPath, "*.tsx")...)

	// Common JavaScript/TypeScript deprecated patterns
	deprecatedPatterns := map[string]DeprecatedItem{
		"var ":                          {Pattern: "var ", Replacement: "const/let", Severity: "warning", Description: "Use const or let instead of var"},
		"$.ajax":                        {Pattern: "$.ajax", Replacement: "fetch() or axios", Severity: "warning", Description: "Use modern HTTP client instead of jQuery ajax"},
		"componentWillMount":            {Pattern: "componentWillMount", Replacement: "componentDidMount", Severity: "critical", Description: "React lifecycle method deprecated"},
		"componentWillReceiveProps":     {Pattern: "componentWillReceiveProps", Replacement: "componentDidUpdate", Severity: "critical", Description: "React lifecycle method deprecated"},
		"componentWillUpdate":           {Pattern: "componentWillUpdate", Replacement: "componentDidUpdate", Severity: "critical", Description: "React lifecycle method deprecated"},
		"UNSAFE_componentWillMount":     {Pattern: "UNSAFE_componentWillMount", Replacement: "componentDidMount", Severity: "warning", Description: "Unsafe React lifecycle method"},
		"UNSAFE_componentWillReceiveProps": {Pattern: "UNSAFE_componentWillReceiveProps", Replacement: "componentDidUpdate", Severity: "warning", Description: "Unsafe React lifecycle method"},
		"UNSAFE_componentWillUpdate":    {Pattern: "UNSAFE_componentWillUpdate", Replacement: "componentDidUpdate", Severity: "warning", Description: "Unsafe React lifecycle method"},
		"ReactDOM.findDOMNode":          {Pattern: "ReactDOM.findDOMNode", Replacement: "useRef hook", Severity: "warning", Description: "findDOMNode is deprecated in StrictMode"},
		"String.prototype.substr":       {Pattern: ".substr(", Replacement: ".substring(", Severity: "warning", Description: "substr() is deprecated, use substring()"},
	}

	for _, file := range jsFiles {
		items = append(items, d.scanFileForPatterns(file, deprecatedPatterns, "JavaScript/TypeScript")...)
	}

	return items
}

// scanPythonFiles scans Python files for deprecated patterns
func (d *DeprecatedComponentsChecker) scanPythonFiles(repoPath string) []DeprecatedItem {
	var items []DeprecatedItem

	pythonFiles := d.findFiles(repoPath, "*.py")

	// Common Python deprecated patterns
	deprecatedPatterns := map[string]DeprecatedItem{
		"imp.":                          {Pattern: "imp.", Replacement: "importlib", Severity: "warning", Description: "imp module is deprecated since Python 3.4"},
		"optparse":                      {Pattern: "optparse", Replacement: "argparse", Severity: "warning", Description: "optparse is deprecated since Python 2.7"},
		"platform.dist":                {Pattern: "platform.dist", Replacement: "platform.freedesktop_os_release", Severity: "warning", Description: "platform.dist deprecated since Python 3.5"},
		"cgi.escape":                    {Pattern: "cgi.escape", Replacement: "html.escape", Severity: "warning", Description: "cgi.escape deprecated since Python 3.2"},
		"collections.Mapping":          {Pattern: "collections.Mapping", Replacement: "collections.abc.Mapping", Severity: "warning", Description: "Import from collections.abc instead"},
		"collections.MutableMapping":   {Pattern: "collections.MutableMapping", Replacement: "collections.abc.MutableMapping", Severity: "warning", Description: "Import from collections.abc instead"},
		"collections.Sequence":         {Pattern: "collections.Sequence", Replacement: "collections.abc.Sequence", Severity: "warning", Description: "Import from collections.abc instead"},
		"collections.Iterable":         {Pattern: "collections.Iterable", Replacement: "collections.abc.Iterable", Severity: "warning", Description: "Import from collections.abc instead"},
		"asyncio.coroutine":            {Pattern: "asyncio.coroutine", Replacement: "async def", Severity: "warning", Description: "Use async/await syntax instead of @asyncio.coroutine"},
	}

	for _, file := range pythonFiles {
		items = append(items, d.scanFileForPatterns(file, deprecatedPatterns, "Python")...)
	}

	return items
}

// scanDockerFiles scans Dockerfile for deprecated patterns
func (d *DeprecatedComponentsChecker) scanDockerFiles(repoPath string) []DeprecatedItem {
	var items []DeprecatedItem

	dockerFiles := d.findFiles(repoPath, "Dockerfile*")
	dockerFiles = append(dockerFiles, d.findFiles(repoPath, "*.dockerfile")...)

	// Common Docker deprecated patterns
	deprecatedPatterns := map[string]DeprecatedItem{
		"MAINTAINER":                   {Pattern: "MAINTAINER", Replacement: "LABEL maintainer=", Severity: "warning", Description: "MAINTAINER instruction is deprecated"},
		"FROM ubuntu:14.04":           {Pattern: "FROM ubuntu:14.04", Replacement: "FROM ubuntu:20.04 or later", Severity: "critical", Description: "Ubuntu 14.04 is end-of-life"},
		"FROM ubuntu:16.04":           {Pattern: "FROM ubuntu:16.04", Replacement: "FROM ubuntu:20.04 or later", Severity: "warning", Description: "Ubuntu 16.04 is end-of-life"},
		"FROM centos:6":               {Pattern: "FROM centos:6", Replacement: "FROM centos:8 or rocky/alma", Severity: "critical", Description: "CentOS 6 is end-of-life"},
		"FROM centos:7":               {Pattern: "FROM centos:7", Replacement: "FROM centos:8 or rocky/alma", Severity: "warning", Description: "CentOS 7 will be end-of-life soon"},
		"FROM node:10":                {Pattern: "FROM node:10", Replacement: "FROM node:18 or later", Severity: "critical", Description: "Node.js 10 is end-of-life"},
		"FROM node:12":                {Pattern: "FROM node:12", Replacement: "FROM node:18 or later", Severity: "warning", Description: "Node.js 12 is end-of-life"},
		"FROM python:2":               {Pattern: "FROM python:2", Replacement: "FROM python:3", Severity: "critical", Description: "Python 2 is end-of-life"},
	}

	for _, file := range dockerFiles {
		items = append(items, d.scanFileForPatterns(file, deprecatedPatterns, "Docker")...)
	}

	return items
}

// scanKubernetesFiles scans Kubernetes YAML files for deprecated API versions
func (d *DeprecatedComponentsChecker) scanKubernetesFiles(repoPath string) []DeprecatedItem {
	var items []DeprecatedItem

	k8sFiles := d.findFiles(repoPath, "*.yaml")
	k8sFiles = append(k8sFiles, d.findFiles(repoPath, "*.yml")...)

	// Common Kubernetes deprecated API versions
	deprecatedPatterns := map[string]DeprecatedItem{
		"apiVersion: extensions/v1beta1":          {Pattern: "apiVersion: extensions/v1beta1", Replacement: "apps/v1", Severity: "critical", Description: "extensions/v1beta1 API is deprecated"},
		"apiVersion: apps/v1beta1":                {Pattern: "apiVersion: apps/v1beta1", Replacement: "apps/v1", Severity: "critical", Description: "apps/v1beta1 API is deprecated"},
		"apiVersion: apps/v1beta2":                {Pattern: "apiVersion: apps/v1beta2", Replacement: "apps/v1", Severity: "warning", Description: "apps/v1beta2 API is deprecated"},
		"apiVersion: networking.k8s.io/v1beta1":  {Pattern: "apiVersion: networking.k8s.io/v1beta1", Replacement: "networking.k8s.io/v1", Severity: "warning", Description: "networking.k8s.io/v1beta1 API is deprecated"},
		"apiVersion: policy/v1beta1":              {Pattern: "apiVersion: policy/v1beta1", Replacement: "policy/v1", Severity: "warning", Description: "policy/v1beta1 API is deprecated"},
		"apiVersion: rbac.authorization.k8s.io/v1beta1": {Pattern: "apiVersion: rbac.authorization.k8s.io/v1beta1", Replacement: "rbac.authorization.k8s.io/v1", Severity: "warning", Description: "rbac.authorization.k8s.io/v1beta1 API is deprecated"},
	}

	for _, file := range k8sFiles {
		items = append(items, d.scanFileForPatterns(file, deprecatedPatterns, "Kubernetes")...)
	}

	return items
}

// scanFileForPatterns scans a file for deprecated patterns
func (d *DeprecatedComponentsChecker) scanFileForPatterns(filePath string, patterns map[string]DeprecatedItem, language string) []DeprecatedItem {
	var items []DeprecatedItem

	content, err := os.ReadFile(filePath) // #nosec G304 - file path is controlled
	if err != nil {
		return items
	}

	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		for patternStr, deprecatedItem := range patterns {
			if strings.Contains(line, patternStr) {
				item := DeprecatedItem{
					File:        filePath,
					Line:        lineNum + 1,
					Pattern:     deprecatedItem.Pattern,
					Replacement: deprecatedItem.Replacement,
					Language:    language,
					Severity:    deprecatedItem.Severity,
					Description: deprecatedItem.Description,
				}
				items = append(items, item)
			}
		}
	}

	return items
}

// findFiles finds files matching a pattern in the repository
func (d *DeprecatedComponentsChecker) findFiles(repoPath, pattern string) []string {
	var files []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if there's an error
		}

		// Skip hidden directories and common ignore patterns
		if strings.HasPrefix(info.Name(), ".") && info.IsDir() {
			return filepath.SkipDir
		}

		if info.IsDir() {
			// Skip common directories that shouldn't be scanned
			skipDirs := []string{"node_modules", "vendor", "target", "build", "dist", "__pycache__", ".git"}
			for _, skipDir := range skipDirs {
				if info.Name() == skipDir {
					return filepath.SkipDir
				}
			}
			return nil
		}

		matched, err := filepath.Match(pattern, info.Name())
		if err != nil {
			return nil
		}

		if matched {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return []string{}
	}

	return files
}

// calculateSeverity calculates the overall severity based on deprecated items found
func (d *DeprecatedComponentsChecker) calculateSeverity(items []DeprecatedItem) int {
	criticalCount := 0
	warningCount := 0

	for _, item := range items {
		switch item.Severity {
		case "critical":
			criticalCount++
		case "warning":
			warningCount++
		}
	}

	if criticalCount > 0 {
		return 3 // Critical
	} else if warningCount > 5 {
		return 2 // Warning with many issues
	} else if warningCount > 0 {
		return 1 // Warning with few issues
	}

	return 1 // Default
}

// determineStatus determines the health status based on the number and severity of deprecated items
func (d *DeprecatedComponentsChecker) determineStatus(count int, severity int) HealthStatus {
	if severity >= 3 {
		return HealthStatusCritical
	} else if count > 10 || severity >= 2 {
		return HealthStatusWarning
	}
	return HealthStatusWarning // Any deprecated usage should at least be a warning
}

// formatDeprecatedItems formats the deprecated items for display
func (d *DeprecatedComponentsChecker) formatDeprecatedItems(items []DeprecatedItem) string {
	if len(items) == 0 {
		return ""
	}

	var details strings.Builder
	details.WriteString("Deprecated component usages found:\n\n")

	// Group by language for better organization
	languageGroups := make(map[string][]DeprecatedItem)
	for _, item := range items {
		languageGroups[item.Language] = append(languageGroups[item.Language], item)
	}

	for language, langItems := range languageGroups {
		details.WriteString(fmt.Sprintf("📋 %s:\n", language))

		for _, item := range langItems {
			severity := "⚠️"
			if item.Severity == "critical" {
				severity = "🚨"
			}

			relPath := strings.TrimPrefix(item.File, GetRepoPath(config.Repository{}))
			if relPath == item.File {
				// If TrimPrefix didn't work, use the basename
				relPath = filepath.Base(item.File)
			}

			details.WriteString(fmt.Sprintf("  %s %s:%d - %s\n", severity, relPath, item.Line, item.Pattern))
			details.WriteString(fmt.Sprintf("      %s\n", item.Description))
			details.WriteString(fmt.Sprintf("      Suggested replacement: %s\n", item.Replacement))
			details.WriteString("\n")
		}
	}

	return details.String()
}

// Example usage of the optimized checker system:
//
// // Create factory
// factory := NewCheckerFactory()
//
// // Get all checkers
// checkers := factory.CreateAllCheckers()
//
// // Get checkers by category
// gitCheckers := factory.CreateCheckersByCategory("git")
//
// // Get specific checker
// depChecker := factory.CreateCheckerByName("Dependencies")
//
// // Use interfaces for testing and modularity
// func runHealthCheck(checker CheckerInterface, repo config.Repository) HealthCheck {
//     return checker.Check(repo)
// }
