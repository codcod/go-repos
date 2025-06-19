package python_analyzer

import (
	"context"
	"testing"

	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/analyzers/common"
)

func TestPythonAnalyzer_NewAnalyzer(t *testing.T) {
	mockWalker := common.NewMockFileWalker()
	mockLogger := &common.MockLogger{}

	analyzer := NewPythonAnalyzer(mockWalker, mockLogger)

	if analyzer == nil {
		t.Fatal("Expected analyzer to be created, got nil")
	}

	if analyzer.Language() != "python" {
		t.Errorf("Expected language 'python', got '%s'", analyzer.Language())
	}

	extensions := analyzer.FileExtensions()
	if len(extensions) != 1 || extensions[0] != ".py" {
		t.Errorf("Expected extensions ['.py'], got %v", extensions)
	}

	if !analyzer.SupportsComplexity() {
		t.Error("Expected analyzer to support complexity analysis")
	}

	if !analyzer.SupportsFunctionLevel() {
		t.Error("Expected analyzer to support function-level analysis")
	}
}

func TestPythonAnalyzer_CanAnalyze(t *testing.T) {
	mockWalker := common.NewMockFileWalker()
	mockLogger := &common.MockLogger{}

	// Setup mock to return Python files
	mockWalker.AddFile("/test/main.py", []byte("print('hello')"))
	mockWalker.AddFile("/test/utils.py", []byte("def helper(): pass"))

	analyzer := NewPythonAnalyzer(mockWalker, mockLogger)

	repo := core.Repository{Path: "/test"}
	canAnalyze := analyzer.CanAnalyze(repo)

	if !canAnalyze {
		t.Error("Expected analyzer to be able to analyze repository with Python files")
	}
}

func TestPythonAnalyzer_CannotAnalyze(t *testing.T) {
	mockWalker := common.NewMockFileWalker()
	mockLogger := &common.MockLogger{}

	// Setup mock to return no files (empty repository)
	// Don't add any files

	analyzer := NewPythonAnalyzer(mockWalker, mockLogger)

	repo := core.Repository{Path: "/empty"}
	canAnalyze := analyzer.CanAnalyze(repo)

	if canAnalyze {
		t.Error("Expected analyzer to not be able to analyze repository without Python files")
	}
}

func TestPythonAnalyzer_AnalyzeComplexity(t *testing.T) {
	mockWalker := common.NewMockFileWalker()
	mockLogger := &common.MockLogger{}

	// Setup mock files and content
	pythonCode := `def simple_function():
    return "hello"

def complex_function(x):
    if x > 0:
        if x > 10:
            return "big"
        else:
            return "small"
    else:
        return "negative"
`

	mockWalker.AddFile("/test/main.py", []byte(pythonCode))

	analyzer := NewPythonAnalyzer(mockWalker, mockLogger)

	ctx := context.Background()
	result, err := analyzer.AnalyzeComplexity(ctx, "/test")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.TotalFiles != 1 {
		t.Errorf("Expected 1 file, got %d", result.TotalFiles)
	}

	if len(result.Functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(result.Functions))
	}

	// Find the complex function and check its complexity
	var complexFunc *core.FunctionComplexity
	for i := range result.Functions {
		if result.Functions[i].Name == "complex_function" {
			complexFunc = &result.Functions[i]
			break
		}
	}

	if complexFunc == nil {
		t.Fatal("Expected to find 'complex_function'")
	}

	// The complex function should have higher complexity due to nested conditions
	if complexFunc.Complexity <= 1 {
		t.Errorf("Expected complex function to have complexity > 1, got %d", complexFunc.Complexity)
	}
}
