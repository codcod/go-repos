package commands

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestOSCommandExecutor_NewOSCommandExecutor(t *testing.T) {
	executor := NewOSCommandExecutor(10 * time.Second)
	if executor == nil {
		t.Fatal("NewOSCommandExecutor() returned nil")
	}

	if executor.defaultTimeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", executor.defaultTimeout)
	}

	// Test default timeout when zero is provided
	executor2 := NewOSCommandExecutor(0)
	if executor2.defaultTimeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", executor2.defaultTimeout)
	}
}

func TestOSCommandExecutor_Execute(t *testing.T) {
	executor := NewOSCommandExecutor(10 * time.Second)
	ctx := context.Background()

	// Test successful command
	result := executor.Execute(ctx, "echo", "hello world")
	if result.Error != nil {
		t.Fatalf("Execute failed: %v", result.Error)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	expectedOutput := "hello world\n"
	if result.Stdout != expectedOutput {
		t.Errorf("Expected stdout %q, got %q", expectedOutput, result.Stdout)
	}

	if result.Duration <= 0 {
		t.Error("Expected positive duration")
	}
}

func TestOSCommandExecutor_Execute_NonExistentCommand(t *testing.T) {
	executor := NewOSCommandExecutor(10 * time.Second)
	ctx := context.Background()

	// Test non-existent command
	result := executor.Execute(ctx, "non-existent-command-xyz")
	if result.Error == nil {
		t.Error("Expected error for non-existent command")
	}

	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for failed command")
	}
}

func TestOSCommandExecutor_Execute_FailingCommand(t *testing.T) {
	executor := NewOSCommandExecutor(10 * time.Second)
	ctx := context.Background()

	// Test command that fails (exit code 1)
	result := executor.Execute(ctx, "sh", "-c", "exit 1")
	if result.Error == nil {
		t.Error("Expected error for failing command")
	}

	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", result.ExitCode)
	}
}

func TestOSCommandExecutor_ExecuteInDir(t *testing.T) {
	executor := NewOSCommandExecutor(10 * time.Second)
	ctx := context.Background()

	// Test command in specific directory
	result := executor.ExecuteInDir(ctx, "/tmp", "pwd")
	if result.Error != nil {
		t.Fatalf("ExecuteInDir failed: %v", result.Error)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	// Output should contain /tmp (may have /private/tmp on macOS)
	if !contains(result.Stdout, "tmp") {
		t.Errorf("Expected stdout to contain 'tmp', got %q", result.Stdout)
	}
}

func TestOSCommandExecutor_ExecuteWithTimeout(t *testing.T) {
	executor := NewOSCommandExecutor(10 * time.Second)
	ctx := context.Background()

	// Test fast command with timeout
	result := executor.ExecuteWithTimeout(ctx, 5*time.Second, "echo", "hello")
	if result.Error != nil {
		t.Fatalf("ExecuteWithTimeout failed: %v", result.Error)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}

func TestOSCommandExecutor_ExecuteWithTimeout_Timeout(t *testing.T) {
	executor := NewOSCommandExecutor(10 * time.Second)
	ctx := context.Background()

	// Test command that times out
	result := executor.ExecuteWithTimeout(ctx, 100*time.Millisecond, "sleep", "1")
	if result.Error == nil {
		t.Error("Expected timeout error")
	}

	if result.ExitCode != -1 {
		t.Errorf("Expected exit code -1 for timeout, got %d", result.ExitCode)
	}

	// The error message could be "timed out" or "signal: killed" depending on the system
	if result.Error == nil {
		t.Error("Expected error for timeout")
	}
}

func TestMockCommandExecutor_NewMockCommandExecutor(t *testing.T) {
	executor := NewMockCommandExecutor()
	if executor == nil {
		t.Fatal("NewMockCommandExecutor() returned nil")
	}

	if executor.responses == nil {
		t.Fatal("responses map is nil")
	}

	if executor.calls == nil {
		t.Fatal("calls slice is nil")
	}
}

func TestMockCommandExecutor_SetResponse(t *testing.T) {
	executor := NewMockCommandExecutor()

	expectedResult := CommandResult{
		ExitCode: 42,
		Stdout:   "custom output",
		Stderr:   "custom error",
		Duration: 500 * time.Millisecond,
		Error:    fmt.Errorf("custom error"),
	}

	executor.SetResponse("echo hello", expectedResult)

	ctx := context.Background()
	result := executor.Execute(ctx, "echo", "hello")

	if result.ExitCode != expectedResult.ExitCode {
		t.Errorf("Expected exit code %d, got %d", expectedResult.ExitCode, result.ExitCode)
	}

	if result.Stdout != expectedResult.Stdout {
		t.Errorf("Expected stdout %q, got %q", expectedResult.Stdout, result.Stdout)
	}

	if result.Stderr != expectedResult.Stderr {
		t.Errorf("Expected stderr %q, got %q", expectedResult.Stderr, result.Stderr)
	}
}

func TestMockCommandExecutor_Execute(t *testing.T) {
	executor := NewMockCommandExecutor()
	ctx := context.Background()

	// Test default response
	result := executor.Execute(ctx, "test", "command")
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if result.Stdout != "mock output" {
		t.Errorf("Expected default stdout, got %q", result.Stdout)
	}

	if result.Duration <= 0 {
		t.Error("Expected positive duration")
	}

	// Verify call was recorded
	calls := executor.GetCalls()
	if len(calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(calls))
	}

	if calls[0].Command != "test" {
		t.Errorf("Expected command 'test', got %q", calls[0].Command)
	}

	if len(calls[0].Args) != 1 || calls[0].Args[0] != "command" {
		t.Errorf("Expected args ['command'], got %v", calls[0].Args)
	}
}

func TestMockCommandExecutor_ExecuteInDir(t *testing.T) {
	executor := NewMockCommandExecutor()
	ctx := context.Background()

	result := executor.ExecuteInDir(ctx, "/test/dir", "ls", "-la")
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !contains(result.Stdout, "/test/dir") {
		t.Errorf("Expected stdout to contain directory, got %q", result.Stdout)
	}

	// Verify call was recorded
	calls := executor.GetCalls()
	if len(calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(calls))
	}

	if calls[0].Dir != "/test/dir" {
		t.Errorf("Expected dir '/test/dir', got %q", calls[0].Dir)
	}
}

func TestMockCommandExecutor_ExecuteWithTimeout(t *testing.T) {
	executor := NewMockCommandExecutor()
	ctx := context.Background()

	result := executor.ExecuteWithTimeout(ctx, 5*time.Second, "test")
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	// Verify call was recorded
	calls := executor.GetCalls()
	if len(calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(calls))
	}
}

func TestMockCommandExecutor_GetCalls(t *testing.T) {
	executor := NewMockCommandExecutor()
	ctx := context.Background()

	// Execute multiple commands
	executor.Execute(ctx, "cmd1", "arg1")
	executor.ExecuteInDir(ctx, "/dir", "cmd2", "arg2", "arg3")
	executor.ExecuteWithTimeout(ctx, time.Second, "cmd3")

	calls := executor.GetCalls()
	if len(calls) != 3 {
		t.Errorf("Expected 3 calls, got %d", len(calls))
	}

	// Verify first call
	if calls[0].Command != "cmd1" || len(calls[0].Args) != 1 || calls[0].Args[0] != "arg1" {
		t.Errorf("First call incorrect: %+v", calls[0])
	}

	// Verify second call
	if calls[1].Command != "cmd2" || calls[1].Dir != "/dir" || len(calls[1].Args) != 2 {
		t.Errorf("Second call incorrect: %+v", calls[1])
	}

	// Verify third call
	if calls[2].Command != "cmd3" || len(calls[2].Args) != 0 {
		t.Errorf("Third call incorrect: %+v", calls[2])
	}
}

func TestCommandExists(t *testing.T) {
	executor := NewMockCommandExecutor()

	// Set up mock responses
	executor.SetResponse("which existing-command", CommandResult{ExitCode: 0})
	executor.SetResponse("which non-existing-command", CommandResult{ExitCode: 1})

	// Test existing command
	if !CommandExists(executor, "existing-command") {
		t.Error("Expected existing command to be found")
	}

	// Test non-existing command
	if CommandExists(executor, "non-existing-command") {
		t.Error("Expected non-existing command to not be found")
	}
}

func TestCommandExistsWindows(t *testing.T) {
	executor := NewMockCommandExecutor()

	// Set up mock responses
	executor.SetResponse("where existing-command", CommandResult{ExitCode: 0})
	executor.SetResponse("where non-existing-command", CommandResult{ExitCode: 1})

	// Test existing command
	if !CommandExistsWindows(executor, "existing-command") {
		t.Error("Expected existing command to be found on Windows")
	}

	// Test non-existing command
	if CommandExistsWindows(executor, "non-existing-command") {
		t.Error("Expected non-existing command to not be found on Windows")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
