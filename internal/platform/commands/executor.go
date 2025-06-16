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

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, e.defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, command, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	result := CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
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
	start := time.Now()

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	result := CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
		Error:    err,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else if timeoutCtx.Err() == context.DeadlineExceeded {
			result.ExitCode = -1
			result.Error = fmt.Errorf("command timed out after %v", timeout)
		} else {
			result.ExitCode = -1
		}
	}

	return result
}

// MockCommandExecutor implements CommandExecutor for testing
type MockCommandExecutor struct {
	responses map[string]CommandResult
	calls     []MockCall
}

// MockCall represents a recorded command call
type MockCall struct {
	Command string
	Args    []string
	Dir     string
}

// NewMockCommandExecutor creates a new mock command executor
func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		responses: make(map[string]CommandResult),
		calls:     make([]MockCall, 0),
	}
}

// SetResponse sets the response for a specific command
func (m *MockCommandExecutor) SetResponse(command string, result CommandResult) {
	m.responses[command] = result
}

// GetCalls returns all recorded calls
func (m *MockCommandExecutor) GetCalls() []MockCall {
	return m.calls
}

// Execute runs a mock command
func (m *MockCommandExecutor) Execute(ctx context.Context, command string, args ...string) CommandResult {
	fullCommand := command
	if len(args) > 0 {
		fullCommand = command + " " + strings.Join(args, " ")
	}

	m.calls = append(m.calls, MockCall{
		Command: command,
		Args:    args,
		Dir:     "",
	})

	if result, exists := m.responses[fullCommand]; exists {
		return result
	}

	// Default response
	return CommandResult{
		ExitCode: 0,
		Stdout:   "mock output",
		Stderr:   "",
		Duration: 100 * time.Millisecond,
		Error:    nil,
	}
}

// ExecuteInDir runs a mock command in a directory
func (m *MockCommandExecutor) ExecuteInDir(ctx context.Context, dir, command string, args ...string) CommandResult {
	fullCommand := command
	if len(args) > 0 {
		fullCommand = command + " " + strings.Join(args, " ")
	}

	m.calls = append(m.calls, MockCall{
		Command: command,
		Args:    args,
		Dir:     dir,
	})

	if result, exists := m.responses[fullCommand]; exists {
		return result
	}

	// Default response
	return CommandResult{
		ExitCode: 0,
		Stdout:   "mock output from " + dir,
		Stderr:   "",
		Duration: 100 * time.Millisecond,
		Error:    nil,
	}
}

// ExecuteWithTimeout runs a mock command with timeout
func (m *MockCommandExecutor) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, command string, args ...string) CommandResult {
	return m.Execute(ctx, command, args...)
}

// CommandExists checks if a command exists in the system PATH
func CommandExists(executor CommandExecutor, command string) bool {
	result := executor.Execute(context.Background(), "which", command)
	return result.ExitCode == 0
}

// CommandExistsWindows checks if a command exists on Windows
func CommandExistsWindows(executor CommandExecutor, command string) bool {
	result := executor.Execute(context.Background(), "where", command)
	return result.ExitCode == 0
}
