/*
Package health provides a unified interface for repository health analysis and orchestration.

This package consolidates all health analysis functionality including:
  - Language-specific code analyzers
  - Repository health checkers
  - Orchestration engine for coordinated analysis
  - Result formatting and reporting

# Overview

The health package is organized into several sub-packages:

  - analyzers/: Language-specific code analysis (complexity, structure, etc.)
  - checkers/: Repository health checks (dependencies, security, compliance, etc.)
  - orchestration/: Execution engine for coordinated analysis workflows
  - reporting/: Result formatting and output functionality

# Quick Start

To perform a basic health check on repositories:

	// Create filesystem and logger
	fs := health.NewFileSystem()
	logger := &simpleLogger{} // Implement core.Logger interface

	// Create registries
	analyzerRegistry := health.NewAnalyzerRegistry(fs, logger)
	checkerRegistry := health.NewCheckerRegistry(
		health.NewCommandExecutor(30 * time.Second),
	)

	// Create orchestration engine
	engine := health.NewOrchestrationEngine(
		checkerRegistry,
		analyzerRegistry,
		config, // Implement core.Config interface
		logger,
	)

	// Execute health check
	result, err := engine.ExecuteHealthCheck(ctx, repositories)
	if err != nil {
		log.Fatal(err)
	}

	// Display results
	formatter := health.NewFormatter(true) // verbose output
	formatter.DisplayResults(*result)

	// Exit with appropriate code
	os.Exit(health.GetExitCode(*result))

# Analyzers

Language analyzers perform static analysis on source code to extract:
  - Function and class information
  - Complexity metrics
  - Import/dependency analysis
  - Code structure analysis

Supported languages:
  - Go
  - Java
  - JavaScript/TypeScript
  - Python

# Checkers

Health checkers validate various aspects of repository health:
  - Base checker: Common repository structure and standards
  - CI/Config checker: Continuous integration configuration
  - Compliance checker: License and legal compliance
  - Dependencies checker: Outdated or vulnerable dependencies
  - Documentation checker: README and documentation quality
  - Git checker: Repository status and commit history
  - Security checker: Branch protection and vulnerability scanning

# Orchestration

The orchestration engine coordinates the execution of analyzers and checkers:
  - Configurable execution workflows
  - Parallel execution support
  - Error handling and retry logic
  - Progress reporting and logging
  - Result aggregation and scoring

# Configuration

The health analysis is driven by configuration that supports:
  - Checker-specific settings (enabled/disabled, severity, timeouts)
  - Analyzer configuration (file extensions, complexity thresholds)
  - Engine settings (concurrency, caching, timeouts)
  - Configuration-based setup for different environments

# Results and Reporting

Analysis results include:
  - Repository-level scores and status
  - Detailed issue and warning information
  - Performance metrics and timing
  - Configurable output formats (console, JSON, HTML)

The reporting system provides:
  - Colorized console output
  - Compact and detailed result views
  - Issue severity classification
  - Exit code determination for CI/CD integration

# Architecture

The health package follows a modular architecture:

 1. Core interfaces define contracts for all components
 2. Platform-specific implementations provide OS integration
 3. Registry pattern manages available analyzers and checkers
 4. Factory functions simplify component creation and configuration
 5. Unified API reduces complexity for consumers

# Extension Points

The architecture supports extension through:
  - Custom analyzers for new languages or analysis types
  - Custom checkers for organization-specific requirements
  - Custom reporters for different output formats
  - Plugin system for dynamic component loading

# Best Practices

When using the health package:
  - Use factory functions rather than direct instantiation
  - Configure appropriate timeouts for long-running operations
  - Implement proper logging for debugging and monitoring
  - Handle context cancellation for graceful shutdown
  - Use configurations for environment-specific settings
  - Implement caching for improved performance
*/
package health
