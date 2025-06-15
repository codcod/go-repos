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
	"context"
	"os"
	"os/exec"
	"path/filepath"
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
		&SecurityChecker{},
		&DependencyChecker{},
		&LicenseChecker{},
		&CIStatusChecker{},
		&DocumentationChecker{},
		&DeprecatedComponentsChecker{},
		&CyclomaticComplexityChecker{},
	}
}

// CreateCheckerByName creates a specific checker by name
func (f *CheckerFactory) CreateCheckerByName(name string) CheckerInterface {
	return f.createCheckerByName(name)
}

// createCheckerByName helper function to reduce complexity
func (f *CheckerFactory) createCheckerByName(name string) CheckerInterface {
	gitCheckers := map[string]CheckerInterface{
		"Git Status":  &GitStatusChecker{},
		"Last Commit": &LastCommitChecker{},
	}

	securityCheckers := map[string]CheckerInterface{
		"Branch Protection": &BranchProtectionChecker{},
		"Security":          &SecurityChecker{},
	}

	qualityCheckers := map[string]CheckerInterface{
		"Dependencies":          &DependencyChecker{},
		"Deprecated Components": &DeprecatedComponentsChecker{},
		"Cyclomatic Complexity": &CyclomaticComplexityChecker{},
	}

	otherCheckers := map[string]CheckerInterface{
		"License":       &LicenseChecker{},
		"CI Status":     &CIStatusChecker{},
		"Documentation": &DocumentationChecker{},
	}

	// Search in all checker maps
	if checker := gitCheckers[name]; checker != nil {
		return checker
	}
	if checker := securityCheckers[name]; checker != nil {
		return checker
	}
	if checker := qualityCheckers[name]; checker != nil {
		return checker
	}
	if checker := otherCheckers[name]; checker != nil {
		return checker
	}

	return nil
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

// LicenseChecker checks for license files
// DocumentationChecker checks for documentation files
