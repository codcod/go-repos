package errors

import (
	"errors"
	"fmt"
	"testing"
)

// Test ConfigError creation and methods
func TestConfigError_Creation(t *testing.T) {
	underlyingErr := errors.New("file not found")

	configErr := &ConfigError{
		Op:   "load",
		Path: "/path/to/config.yaml",
		Err:  underlyingErr,
	}

	if configErr.Op != "load" {
		t.Errorf("Expected Op to be 'load', got %s", configErr.Op)
	}
	if configErr.Path != "/path/to/config.yaml" {
		t.Errorf("Expected Path to be '/path/to/config.yaml', got %s", configErr.Path)
	}
	if configErr.Err != underlyingErr {
		t.Errorf("Expected Err to be the underlying error, got %v", configErr.Err)
	}
}

func TestConfigError_Error_WithPath(t *testing.T) {
	underlyingErr := errors.New("file not found")
	configErr := &ConfigError{
		Op:   "load",
		Path: "/path/to/config.yaml",
		Err:  underlyingErr,
	}

	expected := "config load /path/to/config.yaml: file not found"
	if configErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, configErr.Error())
	}
}

func TestConfigError_Error_WithoutPath(t *testing.T) {
	underlyingErr := errors.New("invalid format")
	configErr := &ConfigError{
		Op:   "parse",
		Path: "",
		Err:  underlyingErr,
	}

	expected := "config parse: invalid format"
	if configErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, configErr.Error())
	}
}

func TestConfigError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("original error")
	configErr := &ConfigError{
		Op:   "validate",
		Path: "config.yaml",
		Err:  underlyingErr,
	}

	unwrapped := configErr.Unwrap()
	if unwrapped != underlyingErr {
		t.Errorf("Expected unwrapped error to be original error, got %v", unwrapped)
	}
}

func TestNewConfigError(t *testing.T) {
	underlyingErr := errors.New("permission denied")
	configErr := NewConfigError("save", "/etc/config.yaml", underlyingErr)

	if configErr.Op != "save" {
		t.Errorf("Expected Op to be 'save', got %s", configErr.Op)
	}
	if configErr.Path != "/etc/config.yaml" {
		t.Errorf("Expected Path to be '/etc/config.yaml', got %s", configErr.Path)
	}
	if configErr.Err != underlyingErr {
		t.Errorf("Expected Err to be the underlying error, got %v", configErr.Err)
	}

	expected := "config save /etc/config.yaml: permission denied"
	if configErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, configErr.Error())
	}
}

// Test GitError creation and methods
func TestGitError_Creation(t *testing.T) {
	underlyingErr := errors.New("repository not found")

	gitErr := &GitError{
		Op:   "clone",
		Repo: "/path/to/repo",
		Err:  underlyingErr,
	}

	if gitErr.Op != "clone" {
		t.Errorf("Expected Op to be 'clone', got %s", gitErr.Op)
	}
	if gitErr.Repo != "/path/to/repo" {
		t.Errorf("Expected Repo to be '/path/to/repo', got %s", gitErr.Repo)
	}
	if gitErr.Err != underlyingErr {
		t.Errorf("Expected Err to be the underlying error, got %v", gitErr.Err)
	}
}

func TestGitError_Error_WithRepo(t *testing.T) {
	underlyingErr := errors.New("branch not found")
	gitErr := &GitError{
		Op:   "checkout",
		Repo: "/path/to/repo",
		Err:  underlyingErr,
	}

	expected := "git checkout in /path/to/repo: branch not found"
	if gitErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, gitErr.Error())
	}
}

func TestGitError_Error_WithoutRepo(t *testing.T) {
	underlyingErr := errors.New("git not installed")
	gitErr := &GitError{
		Op:   "status",
		Repo: "",
		Err:  underlyingErr,
	}

	expected := "git status: git not installed"
	if gitErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, gitErr.Error())
	}
}

func TestGitError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("network timeout")
	gitErr := &GitError{
		Op:   "push",
		Repo: "/path/to/repo",
		Err:  underlyingErr,
	}

	unwrapped := gitErr.Unwrap()
	if unwrapped != underlyingErr {
		t.Errorf("Expected unwrapped error to be original error, got %v", unwrapped)
	}
}

func TestNewGitError(t *testing.T) {
	underlyingErr := errors.New("authentication failed")
	gitErr := NewGitError("push", "/project/repo", underlyingErr)

	if gitErr.Op != "push" {
		t.Errorf("Expected Op to be 'push', got %s", gitErr.Op)
	}
	if gitErr.Repo != "/project/repo" {
		t.Errorf("Expected Repo to be '/project/repo', got %s", gitErr.Repo)
	}
	if gitErr.Err != underlyingErr {
		t.Errorf("Expected Err to be the underlying error, got %v", gitErr.Err)
	}

	expected := "git push in /project/repo: authentication failed"
	if gitErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, gitErr.Error())
	}
}

// Test ValidationError creation and methods
func TestValidationError_Creation(t *testing.T) {
	underlyingErr := errors.New("must be a valid email address")
	validationErr := &ValidationError{
		Field: "email",
		Value: "invalid-email",
		Err:   underlyingErr,
	}

	if validationErr.Field != "email" {
		t.Errorf("Expected Field to be 'email', got %s", validationErr.Field)
	}
	if validationErr.Value != "invalid-email" {
		t.Errorf("Expected Value to be 'invalid-email', got %s", validationErr.Value)
	}
	if validationErr.Err != underlyingErr {
		t.Errorf("Expected Err to be the underlying error, got %v", validationErr.Err)
	}
}

func TestValidationError_Error(t *testing.T) {
	underlyingErr := errors.New("must be a valid port number")
	validationErr := &ValidationError{
		Field: "port",
		Value: "invalid",
		Err:   underlyingErr,
	}

	expected := "validation failed for field port (value: invalid): must be a valid port number"
	if validationErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, validationErr.Error())
	}
}

func TestValidationError_Error_WithoutField(t *testing.T) {
	underlyingErr := errors.New("validation failed")
	validationErr := &ValidationError{
		Field: "",
		Value: "",
		Err:   underlyingErr,
	}

	expected := "validation failed: validation failed"
	if validationErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, validationErr.Error())
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("original validation error")
	validationErr := &ValidationError{
		Field: "username",
		Value: "",
		Err:   underlyingErr,
	}

	unwrapped := validationErr.Unwrap()
	if unwrapped != underlyingErr {
		t.Errorf("Expected unwrapped error to be original error, got %v", unwrapped)
	}
}

func TestNewValidationError(t *testing.T) {
	underlyingErr := errors.New("cannot be empty")
	validationErr := NewValidationError("username", "", underlyingErr)

	if validationErr.Field != "username" {
		t.Errorf("Expected Field to be 'username', got %s", validationErr.Field)
	}
	if validationErr.Value != "" {
		t.Errorf("Expected Value to be empty, got %s", validationErr.Value)
	}
	if validationErr.Err != underlyingErr {
		t.Errorf("Expected Err to be the underlying error, got %v", validationErr.Err)
	}

	expected := "validation failed for field username (value: ): cannot be empty"
	if validationErr.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, validationErr.Error())
	}
}

// Test error type checking functions
func TestIsConfigError(t *testing.T) {
	configErr := NewConfigError("load", "config.yaml", errors.New("file not found"))
	gitErr := NewGitError("clone", "/repo", errors.New("repo not found"))
	regularErr := errors.New("regular error")

	if !IsConfigError(configErr) {
		t.Error("Expected IsConfigError to return true for ConfigError")
	}
	if IsConfigError(gitErr) {
		t.Error("Expected IsConfigError to return false for GitError")
	}
	if IsConfigError(regularErr) {
		t.Error("Expected IsConfigError to return false for regular error")
	}
}

func TestIsGitError(t *testing.T) {
	configErr := NewConfigError("load", "config.yaml", errors.New("file not found"))
	gitErr := NewGitError("clone", "/repo", errors.New("repo not found"))
	regularErr := errors.New("regular error")

	if !IsGitError(gitErr) {
		t.Error("Expected IsGitError to return true for GitError")
	}
	if IsGitError(configErr) {
		t.Error("Expected IsGitError to return false for ConfigError")
	}
	if IsGitError(regularErr) {
		t.Error("Expected IsGitError to return false for regular error")
	}
}

func TestIsValidationError(t *testing.T) {
	validationErr := NewValidationError("field", "value", errors.New("validation failed"))
	configErr := NewConfigError("load", "config.yaml", errors.New("file not found"))
	regularErr := errors.New("regular error")

	if !IsValidationError(validationErr) {
		t.Error("Expected IsValidationError to return true for ValidationError")
	}
	if IsValidationError(configErr) {
		t.Error("Expected IsValidationError to return false for ConfigError")
	}
	if IsValidationError(regularErr) {
		t.Error("Expected IsValidationError to return false for regular error")
	}
}

// Test error unwrapping with errors.Is and errors.As
func TestErrorUnwrapping(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	configErr := NewConfigError("load", "config.yaml", originalErr)

	// Test errors.Is
	if !errors.Is(configErr, originalErr) {
		t.Error("Expected errors.Is to return true for unwrapped error")
	}

	// Test errors.As
	var targetConfigErr *ConfigError
	if !errors.As(configErr, &targetConfigErr) {
		t.Error("Expected errors.As to return true for ConfigError type")
	}
	if targetConfigErr.Op != "load" {
		t.Errorf("Expected Op to be 'load', got %s", targetConfigErr.Op)
	}
}

// Test error chaining
func TestErrorChaining(t *testing.T) {
	baseErr := errors.New("base error")
	configErr := NewConfigError("parse", "config.yaml", baseErr)
	wrappedErr := fmt.Errorf("wrapper: %w", configErr)

	// Test that we can unwrap through multiple layers
	if !errors.Is(wrappedErr, baseErr) {
		t.Error("Expected errors.Is to work through multiple wrapping layers")
	}

	var targetConfigErr *ConfigError
	if !errors.As(wrappedErr, &targetConfigErr) {
		t.Error("Expected errors.As to work through multiple wrapping layers")
	}
}
