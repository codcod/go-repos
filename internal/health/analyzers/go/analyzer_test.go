package go_analyzer

import (
	"context"
	"testing"

	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/analyzers/common"
)

func TestNewGoAnalyzer(t *testing.T) {
	logger := &common.MockLogger{}
	mockWalker := common.NewMockFileWalker()

	analyzer := NewGoAnalyzer(mockWalker, logger)

	if analyzer == nil {
		t.Fatal("NewGoAnalyzer returned nil")
	}

	if analyzer.Language() != "go" {
		t.Errorf("Expected language 'go', got %s", analyzer.Language())
	}

	extensions := analyzer.FileExtensions()
	if len(extensions) != 1 || extensions[0] != ".go" {
		t.Errorf("Expected extensions ['.go'], got %v", extensions)
	}

	if !analyzer.SupportsComplexity() {
		t.Error("Expected analyzer to support complexity analysis")
	}

	if !analyzer.SupportsFunctionLevel() {
		t.Error("Expected analyzer to support function-level analysis")
	}
}

func TestGoAnalyzer_CanAnalyze(t *testing.T) {
	logger := &common.MockLogger{}
	mockWalker := common.NewMockFileWalker()
	analyzer := NewGoAnalyzer(mockWalker, logger)

	// Add Go files to the mock walker
	mockWalker.AddFile("/test/main.go", []byte("package main"))
	mockWalker.AddFile("/test/utils.go", []byte("package utils"))

	// Test with Go files
	repoWithGo := core.Repository{
		Name: "test-repo",
		Path: "/test",
		Tags: []string{"go"},
	}

	if !analyzer.CanAnalyze(repoWithGo) {
		t.Error("Expected CanAnalyze to return true for repository with Go files")
	}

	// Test without Go files
	emptyRepo := core.Repository{
		Name: "empty-repo",
		Path: "/empty",
		Tags: []string{},
	}

	if analyzer.CanAnalyze(emptyRepo) {
		t.Error("Expected CanAnalyze to return false for repository without Go files")
	}
}

func TestGoAnalyzer_Basic(t *testing.T) {
	logger := &common.MockLogger{}
	mockWalker := common.NewMockFileWalker()
	analyzer := NewGoAnalyzer(mockWalker, logger)

	// Setup mock data
	goCode := `package main

import "fmt"

func hello() {
	fmt.Println("Hello, World!")
}

func main() {
	hello()
}
`
	mockWalker.AddFile("/test/main.go", []byte(goCode))

	ctx := context.Background()
	config := core.AnalyzerConfig{}

	result, err := analyzer.Analyze(ctx, "/test", config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze returned nil result")
	}

	if result.Language != "go" {
		t.Errorf("Expected language 'go', got %s", result.Language)
	}

	// Check that we have at least some basic metrics
	if result.Metrics == nil {
		t.Error("Expected metrics to be set")
	}
}
