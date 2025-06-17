/*
Package checkers provides repository health validation capabilities.

This package contains specialized checkers that validate different aspects
of repository health and quality standards.

# Checker Categories

  - base: Fundamental repository structure validation
  - ci: Continuous integration configuration checks
  - compliance: License and legal compliance validation
  - dependencies: Dependency management and security checks
  - docs: Documentation quality and completeness assessment
  - git: Git repository health and hygiene validation
  - security: Security-focused validation and vulnerability detection

# Architecture

Each checker implements the core.Checker interface and can be registered
with the checker registry for coordinated execution.

The registry provides:
  - Automatic checker discovery and registration
  - Category-based organization and filtering
  - Configuration management for checker-specific settings
  - Parallel and sequential execution coordination
  - Result aggregation and reporting

# Usage

Basic usage through the checker registry:

	executor := commands.NewOSCommandExecutor(30 * time.Second)
	registry := checkers.NewRegistry(executor)
	checker := registry.GetChecker("git-status")
	result, err := checker.Check(ctx, repoContext)

Integration with the health package:

	registry := health.NewCheckerRegistry(executor)
	results, err := registry.RunChecks(ctx, repo, categories)

# Configuration

Checkers can be configured through the configuration system:

	checkers:
	  git-status:
	    enabled: true
	    severity: "medium"
	    timeout: "10s"
	    options:
	      check_uncommitted: true
	      check_unpushed: true

# Extension

To add new checkers:

1. Implement the core.Checker interface
2. Register the checker with the registry
3. Define checker category and configuration
4. Add appropriate error handling and logging

Example checker implementation:

	type CustomChecker struct {
		id       string
		name     string
		category string
		config   core.CheckerConfig
	}

	func (c *CustomChecker) ID() string { return c.id }
	func (c *CustomChecker) Name() string { return c.name }
	func (c *CustomChecker) Category() string { return c.category }
	func (c *CustomChecker) Check(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
		// Checker implementation
	}

# Best Practices

When implementing checkers:
  - Use appropriate timeout handling
  - Provide clear, actionable error messages
  - Support configuration options for flexibility
  - Include proper logging for debugging
  - Handle edge cases gracefully
  - Follow naming conventions for consistency
*/
package checkers
