package commands

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// CommandResult represents the result of command execution
type CommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	Error    error
}

// CommandExecutor defines the interface for command execution
type CommandExecutor interface {
	Execute(ctx context.Context, command string, args ...string) CommandResult
	ExecuteInDir(ctx context.Context, dir, command string, args ...string) CommandResult
	ExecuteWithTimeout(ctx context.Context, timeout time.Duration, command string, args ...string) CommandResult
}

// OSCommandExecutor implements CommandExecutor using the OS
type OSCommandExecutor struct {
	defaultTimeout time.Duration
}

// NewOSCommandExecutor creates a new OS command executor
func NewOSCommandExecutor(defaultTimeout time.Duration) *OSCommandExecutor {
	if defaultTimeout <= 0 {
		defaultTimeout = 30 * time.Second
	}

	return &OSCommandExecutor{
		defaultTimeout: defaultTimeout,
	}
}

// Execute runs a command
func (e *OSCommandExecutor) Execute(ctx context.Context, command string, args ...string) CommandResult {
	return e.ExecuteWithTimeout(ctx, e.defaultTimeout, command, args...)
}

// ExecuteInDir runs a command in a specific directory
func (e *OSCommandExecutor) ExecuteInDir(ctx context.Context, dir, command string, args ...string) CommandResult {
	start := time.Now()

	// Create command with context
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	result := CommandResult{
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		Duration: duration,
		Error:    err,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}

// ExecuteWithTimeout runs a command with a specific timeout
func (e *OSCommandExecutor) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, command string, args ...string) CommandResult {
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()

	// Create command with timeout context
	cmd := exec.CommandContext(timeoutCtx, command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	result := CommandResult{
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		Duration: duration,
		Error:    err,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}

		// Check if it was a timeout
		if timeoutCtx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Errorf("command timed out after %v: %w", timeout, err)
		}
	}

	return result
}
