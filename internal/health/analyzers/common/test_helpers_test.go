package common

import (
	"testing"

	"github.com/codcod/repos/internal/core"
)

func TestMockFileWalker(t *testing.T) {
	walker := NewMockFileWalker()

	// Add test files
	walker.AddFile("/test/main.go", []byte(`package main

func main() {
	if true {
		println("hello")
	}
}`))

	walker.AddFile("/test/utils.py", []byte(`def hello():
	if True:
		print("hello")`))

	// Test finding Go files
	goFiles, err := walker.FindFiles("/test", []string{".go"}, []string{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(goFiles) != 1 {
		t.Errorf("Expected 1 Go file, got %d", len(goFiles))
	}

	// Test reading file
	content, err := walker.ReadFile("/test/main.go")
	if err != nil {
		t.Fatalf("Expected no error reading file, got %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected file content, got empty")
	}
}

func TestMockLogger(t *testing.T) {
	logger := NewMockLogger()

	// Test logging
	logger.Info("test message", core.Field{Key: "test", Value: "value"})
	logger.Error("error message")

	if len(logger.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(logger.Messages))
	}

	lastMsg := logger.GetLastMessage()
	if lastMsg == nil {
		t.Fatal("Expected last message, got nil")
	}

	if lastMsg.Level != "ERROR" {
		t.Errorf("Expected ERROR level, got %s", lastMsg.Level)
	}
}

func TestComplexityHelpers(t *testing.T) {
	functions := []core.FunctionComplexity{
		{Name: "func1", Complexity: 2},
		{Name: "func2", Complexity: 4},
		{Name: "func3", Complexity: 6},
	}

	avg := CalculateAverageComplexity(functions)
	if avg != 4.0 {
		t.Errorf("Expected average complexity 4.0, got %f", avg)
	}

	maxComplexity := FindMaxComplexity(functions)
	if maxComplexity != 6 {
		t.Errorf("Expected max complexity 6, got %d", maxComplexity)
	}
}
