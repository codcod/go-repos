# Architecture and Implementation Guide

This document provides an overview of the repos tool architecture, its key components, and implementation details after the comprehensive improvements made in 2025.

## Table of Contents
- [Overview](#overview)
- [Architecture](#architecture)
- [Key Components](#key-components)
- [Configuration Management](#configuration-management)
- [Error Handling](#error-handling)
- [Observability](#observability)
- [Testing](#testing)
- [Usage Examples](#usage-examples)

## Overview

The repos tool has been redesigned with a modular, maintainable architecture that follows industry best practices. The improvements focus on:

- **Structured Configuration Management**: Validation, profiles, and extensible configuration
- **Comprehensive Error Handling**: Contextual errors with severity and recovery
- **Advanced Observability**: Structured logging, metrics collection, and performance monitoring
- **Modular Design**: Clear separation of concerns and dependency injection
- **Robust Testing**: Comprehensive test coverage with builders and utilities

## Architecture

```
repos/
├── cmd/repos/                    # CLI entry points
│   ├── main.go                   # Application main
│   └── commands/                 # Command implementations
│       ├── health.go             # Health check command
│       └── health_test.go        # Command tests
├── internal/                     # Internal packages
│   ├── config/                   # Configuration management
│   │   ├── validator.go          # Config validation system
│   │   └── validator_test.go     # Validation tests
│   ├── errors/                   # Error handling
│   │   ├── errors.go             # Contextual error system
│   │   └── errors_test.go        # Error handling tests
│   ├── observability/            # Logging and metrics
│   │   ├── logger.go             # Structured logging
│   │   ├── logger_test.go        # Logging tests
│   │   ├── metrics.go            # Metrics collection
│   │   └── metrics_test.go       # Metrics tests
│   ├── health/                   # Health checking system
│   │   ├── core/                 # Core utilities
│   │   │   ├── result_builder.go # Result building utility
│   │   │   └── result_builder_test.go
│   │   ├── checkers/             # Health checkers
│   │   ├── orchestration/        # Execution orchestration
│   │   └── reporting/            # Result formatting
│   ├── testutil/                 # Test utilities
│   │   └── testutil.go           # Test builders and helpers
│   └── core/                     # Core types and interfaces
└── docs/                         # Documentation
    ├── architecture.md           # This file
    ├── 09_maintainability_improvements.md
    ├── 12_code_structure_analysis.md
    └── ...
```

## Key Components

### 1. Configuration Management (`internal/config`)

**ConfigValidator**: Extensible validation system with rule-based validation:

```go
// Create validator with default rules
validator := config.NewConfigValidator()

// Add custom validation rule
validator.AddRule(&CustomValidationRule{})

// Validate configuration
if err := validator.ValidateAdvanced(advConfig); err != nil {
    // Handle validation errors
}
```

**Features:**
- Rule-based validation system
- Support for basic and advanced configurations
- Extensible with custom validation rules
- Comprehensive error reporting

### 2. Error Handling (`internal/errors`)

**ContextualError**: Rich error system with context and severity:

```go
// Create contextual error
err := errors.NewContextualError("operation_failed", baseErr).
    WithSeverity(errors.SeverityHigh).
    WithContext("repository", "my-repo").
    WithContext("file", "/path/to/file")

// Check error context
if contextErr, ok := err.(*errors.ContextualError); ok {
    fmt.Printf("Severity: %s\n", contextErr.Severity())
    fmt.Printf("Context: %v\n", contextErr.Context())
}
```

**Features:**
- Severity levels (Low, Medium, High, Critical)
- Rich context information
- Error wrapping and unwrapping
- User-friendly error messages

### 3. Observability (`internal/observability`)

**StructuredLogger**: Contextual logging with structured fields:

```go
// Create logger with context
logger := observability.NewStructuredLogger(observability.LevelInfo).
    WithPrefix("health-check").
    WithField("repository", "my-repo")

// Log with additional fields
logger.Info("analysis started", 
    core.String("language", "go"),
    core.Int("files", 42))

// Track operations
opLogger, done := logger.StartOperation("code-analysis")
defer done()
```

**MetricsCollector**: Performance monitoring and metrics:

```go
// Create metrics collector
metrics := observability.NewMetricsCollector()

// Track counters and gauges
metrics.IncrementCounter("repositories_processed")
metrics.SetGauge("active_connections", 10)

// Record histograms
metrics.RecordHistogram("response_time_ms", 150.0)

// Measure operations
err := metrics.MeasureOperation("health_check", func() error {
    return performHealthCheck()
})

// Get summary
summary := metrics.GetSummary()
metrics.PrintSummary()
```

### 4. Health System (`internal/health`)

**ResultBuilder**: Fluent interface for building health results:

```go
// Create result builder
builder := core.NewResultBuilder(repository)

// Add check results
builder.AddCheckResult(securityResult).
    AddCheckResult(qualityResult).
    AddWarning("minor-issue", "Non-critical warning", location)

// Build final result
result := builder.Build()
```

**Features:**
- Fluent API for result construction
- Automatic aggregation and scoring
- Support for warnings and issues
- Metadata management

### 5. Testing (`internal/testutil`)

**Test Builders**: DRY test data construction:

```go
// Build test repository
repo := testutil.NewRepositoryBuilder().
    WithName("test-repo").
    WithURL("git@github.com:owner/test-repo.git").
    WithLanguage("go").
    Build()

// Build test check result
checkResult := testutil.NewCheckResultBuilder().
    WithName("security-check").
    WithStatus(core.StatusHealthy).
    WithScore(85).
    AddIssue("medium", "Security issue found").
    Build()

// Create test environment
env := testutil.NewTestEnvironment(t).
    WithRepository(repo).
    WithTempDir()
defer env.Cleanup()
```

## Configuration Management

The configuration system supports both basic repository definitions and advanced health check configurations:

### Basic Configuration (YAML)
```yaml
repositories:
  - name: my-service
    url: git@github.com:org/my-service.git
    tags: [backend, go]
    language: go
```

### Advanced Configuration (YAML)
```yaml
profiles:
  strict:
    checkers:
      security:
        enabled: true
        severity_threshold: low
      code_quality:
        enabled: true
        max_complexity: 10

checkers:
  security:
    enabled: true
    config:
      scan_dependencies: true
      check_secrets: true
```

### Validation

All configurations are validated using the `ConfigValidator`:

```go
validator := config.NewConfigValidator()

// Validate basic config
if err := validator.ValidateBasic(basicConfig); err != nil {
    log.Fatal("Basic config validation failed:", err)
}

// Validate advanced config
if err := validator.ValidateAdvanced(advConfig); err != nil {
    log.Fatal("Advanced config validation failed:", err)
}
```

## Error Handling

The error handling system provides rich context and recovery options:

### Error Types

1. **ContextualError**: General errors with context
2. **FileError**: File-specific errors
3. **ValidationError**: Configuration validation errors

### Usage Patterns

```go
// File operation error
if err := readConfigFile(path); err != nil {
    return errors.NewFileError("read_config", path, err)
}

// Validation error with context
if repo.Name == "" {
    return errors.NewValidationError("repository name is required").
        WithContext("field", "name").
        WithSeverity(errors.SeverityHigh)
}

// Operation error with recovery
if err := performOperation(); err != nil {
    return errors.NewContextualError("operation_failed", err).
        WithContext("operation", "health_check").
        WithRecoveryHint("Try running with --verbose for more details")
}
```

## Observability

### Structured Logging

Logging follows structured patterns with consistent field naming:

```go
// Create contextual logger
logger := observability.NewStructuredLogger(observability.LevelInfo).
    WithPrefix("health-check")

// Log with context
logger.Info("starting repository analysis",
    core.String("repository", repo.Name),
    core.String("language", repo.Language),
    core.Int("files_found", fileCount))

// Track operations with automatic timing
opLogger, done := logger.StartOperation("dependency_scan")
defer done()

// Operation completion automatically logs duration
```

### Metrics Collection

Comprehensive metrics for performance monitoring:

```go
// Create metrics collector
metrics := observability.NewMetricsCollector()

// Track various metric types
metrics.IncrementCounter("checks_executed")
metrics.SetGauge("repositories_active", float64(activeCount))
metrics.RecordHistogram("check_duration_ms", duration.Milliseconds())

// Measure operations automatically
err := metrics.MeasureOperation("security_scan", func() error {
    return scanner.Scan(repository)
})

// Get comprehensive summary
summary := metrics.GetSummary()
fmt.Printf("Processed %d repositories in %v\n", 
    summary.RepositoriesCount, summary.TotalDuration)
```

## Testing

### Test Structure

Tests use builders for maintainable, readable test data:

```go
func TestHealthChecker(t *testing.T) {
    // Arrange
    repo := testutil.NewRepositoryBuilder().
        WithName("test-repo").
        WithLanguage("go").
        Build()
    
    env := testutil.NewTestEnvironment(t).
        WithRepository(repo).
        WithTempDir()
    defer env.Cleanup()
    
    checker := NewSecurityChecker(env.Config)
    
    // Act
    result := checker.Check(repo)
    
    // Assert
    assert.Equal(t, core.StatusHealthy, result.Status)
    assert.GreaterOrEqual(t, result.Score, 80)
}
```

### Test Utilities

1. **RepositoryBuilder**: Fluent repository creation
2. **CheckResultBuilder**: Health check result construction
3. **TestEnvironment**: Isolated test environment management
4. **MockLogger**: Structured logging for tests

## Usage Examples

### Basic Health Check

```go
// Create health command configuration
config := &commands.HealthConfig{
    ConfigPath:  "config/health.yaml",
    BasicConfig: "config.yaml",
    Timeout:     5 * time.Minute,
    Verbose:     true,
}

// Create and execute command
cmd := commands.NewHealthCommand(config)
if err := cmd.Execute(context.Background()); err != nil {
    log.Fatal("Health check failed:", err)
}
```

### Custom Checker Implementation

```go
type CustomChecker struct {
    logger  *observability.CheckerLogger
    metrics *observability.MetricsCollector
}

func (c *CustomChecker) Check(repo core.Repository) core.CheckResult {
    opLogger, done := c.logger.StartOperation("custom_check")
    defer done()
    
    builder := core.NewResultBuilder(repo)
    
    // Perform custom checks
    if err := c.performCustomAnalysis(repo); err != nil {
        c.metrics.IncrementCounter("custom_check_errors")
        builder.AddIssue(core.SeverityHigh, "Custom check failed", nil)
        return builder.WithStatus(core.StatusCritical).Build()
    }
    
    c.metrics.IncrementCounter("custom_check_success")
    opLogger.Info("custom check completed successfully")
    return builder.WithStatus(core.StatusHealthy).WithScore(100).Build()
}
```

### Configuration Validation

```go
// Create validator with custom rules
validator := config.NewConfigValidator()
validator.AddRule(&CustomValidationRule{
    Name: "custom-rule",
    ValidateFunc: func(cfg interface{}) error {
        // Custom validation logic
        return nil
    },
})

// Validate configuration
if err := validator.ValidateAdvanced(advConfig); err != nil {
    if validationErr, ok := err.(*errors.ValidationError); ok {
        fmt.Printf("Validation failed: %s\n", validationErr.Error())
        fmt.Printf("Context: %v\n", validationErr.Context())
    }
}
```

This architecture provides a solid foundation for maintainable, observable, and testable code while supporting the tool's core functionality of multi-repository management and health checking.
