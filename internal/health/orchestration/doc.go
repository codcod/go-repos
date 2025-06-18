/*
Package orchestration provides workflow orchestration capabilities for coordinating
repository analysis and health checking operations.

This package implements a sophisticated execution engine that coordinates
analyzers and checkers in configurable workflows.

# Features

  - Configurable execution workflows
  - Parallel and sequential operation support
  - Advanced error handling and retry logic
  - Progress tracking and detailed logging
  - Result aggregation and scoring
  - Timeout and cancellation support
  - Resource management and cleanup

# Architecture

The orchestration engine coordinates multiple components:

  - Engine: Main orchestration controller
  - Types: Orchestration-specific type definitions
  - Context: Execution context and state management

# Usage

Basic orchestration through the engine:

	engine := orchestration.NewEngine(checkerRegistry, analyzerRegistry, config, logger)
	result, err := engine.ExecuteHealthCheck(ctx, repositories)

Integration with the health package:

	engine := health.NewOrchestrationEngine(checkerRegistry, analyzerRegistry, config, logger)
	result, err := engine.ExecuteHealthCheck(ctx, repos)

# Configuration

Health checks can be configured through YAML:

	profiles:
	  production:
	    categories: ["security", "quality", "compliance"]
	    checkers:
	      security-scan:
	        enabled: true
	        timeout: "30s"
	      quality-check:
	        enabled: true
	        parallel: true
	        timeout: "60s"

# Advanced Features

The orchestration engine supports:

  - Dynamic workflow construction
  - Conditional execution
  - Resource pooling and management
  - Distributed execution capabilities
  - Custom checker types and handlers
  - Profile composition and inheritance

# Error Handling

The engine provides comprehensive error handling:

  - Checker-level error isolation
  - Retry mechanisms with backoff
  - Graceful degradation options
  - Detailed error reporting and logging
  - Recovery and cleanup procedures

# Performance Optimization

Built-in optimizations include:

  - Intelligent parallelization
  - Resource reuse and pooling
  - Caching of intermediate results
  - Lazy loading of expensive operations
  - Memory management and cleanup

# Monitoring and Observability

The engine provides extensive monitoring:

  - Execution metrics and timing
  - Progress reporting and status updates
  - Detailed logging at multiple levels
  - Integration with external monitoring systems
  - Performance profiling capabilities
*/
package orchestration
