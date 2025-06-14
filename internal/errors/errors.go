// Package errors provides custom error types for better error handling.
package errors

import "fmt"

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
