// Package errors provides custom error types for better error handling.
package errors

import (
	"fmt"
	"strings"
)

// ConfigError represents configuration-related errors.
type ConfigError struct {
	Op   string // Operation that failed
	Path string // File path that caused the error
	Err  error  // Underlying error
}

func (e *ConfigError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("config %s %s: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("config %s: %v", e.Op, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new ConfigError.
func NewConfigError(op, path string, err error) *ConfigError {
	return &ConfigError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

// GitError represents git operation errors.
type GitError struct {
	Op   string // Git operation that failed
	Repo string // Repository path
	Err  error  // Underlying error
}

func (e *GitError) Error() string {
	if e.Repo != "" {
		return fmt.Sprintf("git %s in %s: %v", e.Op, e.Repo, e.Err)
	}
	return fmt.Sprintf("git %s: %v", e.Op, e.Err)
}

func (e *GitError) Unwrap() error {
	return e.Err
}

// NewGitError creates a new GitError.
func NewGitError(op, repo string, err error) *GitError {
	return &GitError{
		Op:   op,
		Repo: repo,
		Err:  err,
	}
}

// ValidationError represents validation errors.
type ValidationError struct {
	Field string // Field that failed validation
	Value string // Value that failed validation
	Err   error  // Underlying error
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation failed for field %s (value: %s): %v", e.Field, e.Value, e.Err)
	}
	return fmt.Sprintf("validation failed: %v", e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, value string, err error) *ValidationError {
	return &ValidationError{
		Field: field,
		Value: value,
		Err:   err,
	}
}

// ContextualError provides rich error context for better debugging and user experience
type ContextualError struct {
	Op       string                 // Operation that failed (e.g., "read_file", "analyze_go")
	Path     string                 // File/repository path where error occurred
	Cause    error                  // Underlying error that caused this failure
	Context  map[string]interface{} // Additional context information
	Severity ErrorSeverity          // Severity level of the error
}

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	// SeverityLow indicates a minor error that doesn't prevent operation
	SeverityLow ErrorSeverity = iota
	// SeverityMedium indicates an error that may affect some functionality
	SeverityMedium
	// SeverityHigh indicates a critical error that prevents operation
	SeverityHigh
)

// String returns the string representation of ErrorSeverity
func (s ErrorSeverity) String() string {
	switch s {
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	default:
		return "UNKNOWN"
	}
}

// Error implements the error interface with rich context
func (e *ContextualError) Error() string {
	var builder strings.Builder

	// Add severity prefix
	if e.Severity != SeverityLow {
		builder.WriteString(fmt.Sprintf("[%s] ", e.Severity.String()))
	}

	// Add operation description
	builder.WriteString(fmt.Sprintf("operation '%s' failed", e.Op))

	// Add path if available
	if e.Path != "" {
		builder.WriteString(fmt.Sprintf(" at '%s'", e.Path))
	}

	// Add underlying cause
	if e.Cause != nil {
		builder.WriteString(fmt.Sprintf(": %v", e.Cause))
	}

	// Add context information
	if len(e.Context) > 0 {
		var contextPairs []string
		for key, value := range e.Context {
			contextPairs = append(contextPairs, fmt.Sprintf("%s=%v", key, value))
		}
		builder.WriteString(fmt.Sprintf(" [%s]", strings.Join(contextPairs, ", ")))
	}

	return builder.String()
}

// Unwrap returns the underlying error for error unwrapping
func (e *ContextualError) Unwrap() error {
	return e.Cause
}

// WithContext adds additional context to the error
func (e *ContextualError) WithContext(key string, value interface{}) *ContextualError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewContextualError creates a new contextual error
func NewContextualError(op string, cause error) *ContextualError {
	return &ContextualError{
		Op:       op,
		Cause:    cause,
		Context:  make(map[string]interface{}),
		Severity: SeverityMedium,
	}
}

// NewFileError creates a contextual error for file operations
func NewFileError(op, path string, cause error) *ContextualError {
	return &ContextualError{
		Op:       op,
		Path:     path,
		Cause:    cause,
		Context:  make(map[string]interface{}),
		Severity: SeverityLow, // Default to low severity for file errors
	}
}

// NewAnalysisError creates a contextual error for analysis operations
func NewAnalysisError(analyzer, path string, cause error) *ContextualError {
	err := &ContextualError{
		Op:       "analyze",
		Path:     path,
		Cause:    cause,
		Context:  map[string]interface{}{"analyzer": analyzer},
		Severity: SeverityMedium,
	}
	return err
}

// IsContextualError checks if an error is a ContextualError
func IsContextualError(err error) bool {
	_, ok := err.(*ContextualError)
	return ok
}

// GetContextualError extracts a ContextualError from an error chain
func GetContextualError(err error) *ContextualError {
	if contextErr, ok := err.(*ContextualError); ok {
		return contextErr
	}
	return nil
}

// IsConfigError checks if an error is a configuration error.
func IsConfigError(err error) bool {
	_, ok := err.(*ConfigError)
	return ok
}

// IsGitError checks if an error is a git operation error.
func IsGitError(err error) bool {
	_, ok := err.(*GitError)
	return ok
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}
