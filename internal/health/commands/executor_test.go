package commands

import (
	"context"
	"testing"
	"time"
)

func TestNewOSCommandExecutor(t *testing.T) {
	executor := NewOSCommandExecutor(30 * time.Second)
	if executor == nil {
		t.Fatal("Expected executor to be created")
	}

	if executor.defaultTimeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", executor.defaultTimeout)
	}
}

func TestNewOSCommandExecutor_ZeroTimeout(t *testing.T) {
	executor := NewOSCommandExecutor(0)
	if executor == nil {
		t.Fatal("Expected executor to be created")
	}

	if executor.defaultTimeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", executor.defaultTimeout)
	}
}

func TestOSCommandExecutor_Execute(t *testing.T) {
	executor := NewOSCommandExecutor(5 * time.Second)

	ctx := context.Background()
	result := executor.Execute(ctx, "echo", "hello")

	if result.Error != nil {
		t.Fatalf("Expected no error, got %v", result.Error)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if result.Stdout != "hello" {
		t.Errorf("Expected stdout 'hello', got '%s'", result.Stdout)
	}
}

func TestOSCommandExecutor_ExecuteWithTimeout(t *testing.T) {
	executor := NewOSCommandExecutor(5 * time.Second)

	ctx := context.Background()
	result := executor.ExecuteWithTimeout(ctx, 1*time.Second, "echo", "test")

	if result.Error != nil {
		t.Fatalf("Expected no error, got %v", result.Error)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}
