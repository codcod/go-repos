package go_analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/platform/filesystem"
)

// MockLogger for testing
type MockLogger struct {
	InfoCalls  [][]interface{}
	ErrorCalls [][]interface{}
	DebugCalls [][]interface{}
	WarnCalls  [][]interface{}
}

func (m *MockLogger) Info(msg string, fields ...core.Field) {
	args := make([]interface{}, len(fields)+1)
	args[0] = msg
	for i, field := range fields {
		args[i+1] = field
	}
	m.InfoCalls = append(m.InfoCalls, args)
}

func (m *MockLogger) Error(msg string, fields ...core.Field) {
	args := make([]interface{}, len(fields)+1)
	args[0] = msg
	for i, field := range fields {
		args[i+1] = field
	}
	m.ErrorCalls = append(m.ErrorCalls, args)
}

func (m *MockLogger) Debug(msg string, fields ...core.Field) {
	args := make([]interface{}, len(fields)+1)
	args[0] = msg
	for i, field := range fields {
		args[i+1] = field
	}
	m.DebugCalls = append(m.DebugCalls, args)
}

func (m *MockLogger) Warn(msg string, fields ...core.Field) {
	args := make([]interface{}, len(fields)+1)
	args[0] = msg
	for i, field := range fields {
		args[i+1] = field
	}
	m.WarnCalls = append(m.WarnCalls, args)
}

func (m *MockLogger) Fatal(msg string, fields ...core.Field) {
	args := make([]interface{}, len(fields)+1)
	args[0] = msg
	for i, field := range fields {
		args[i+1] = field
	}
	// For testing, we'll just record the call instead of exiting
	m.ErrorCalls = append(m.ErrorCalls, args)
}

func TestNewGoAnalyzer(t *testing.T) {
	logger := &MockLogger{}
	fs := filesystem.NewOSFileSystem()

	analyzer := NewGoAnalyzer(fs, logger)

	if analyzer == nil {
		t.Fatal("NewGoAnalyzer returned nil")
	}

	if analyzer.Name() != "go-analyzer" {
		t.Errorf("Expected name 'go-analyzer', got %s", analyzer.Name())
	}

	if analyzer.Language() != "go" {
		t.Errorf("Expected language 'go', got %s", analyzer.Language())
	}

	extensions := analyzer.SupportedExtensions()
	if len(extensions) != 1 || extensions[0] != ".go" {
		t.Errorf("Expected extensions ['.go'], got %v", extensions)
	}
}

func TestGoAnalyzer_CanAnalyze(t *testing.T) {
	logger := &MockLogger{}
	fs := filesystem.NewOSFileSystem()
	analyzer := NewGoAnalyzer(fs, logger)

	// Create a temporary directory with Go files
	tempDir, err := os.MkdirTemp("", "go-analyzer-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a Go file
	goFile := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(goFile, []byte(`package main

func main() {
	println("hello")
}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	repo := core.Repository{
		Name: "test-repo",
		Path: tempDir,
	}

	if !analyzer.CanAnalyze(repo) {
		t.Error("Expected CanAnalyze to return true for repository with Go files")
	}

	// Test with directory without Go files
	emptyDir, err := os.MkdirTemp("", "empty-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(emptyDir)

	emptyRepo := core.Repository{
		Name: "empty-repo",
		Path: emptyDir,
	}

	if analyzer.CanAnalyze(emptyRepo) {
		t.Error("Expected CanAnalyze to return false for repository without Go files")
	}
}

func TestGoAnalyzer_Analyze(t *testing.T) {
	logger := &MockLogger{}
	fs := filesystem.NewOSFileSystem()
	analyzer := NewGoAnalyzer(fs, logger)

	// Create a temporary directory with Go files
	tempDir, err := os.MkdirTemp("", "go-analyzer-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go file with functions
	goFile := filepath.Join(tempDir, "example.go")
	goContent := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}

func add(a, b int) int {
	if a > 0 {
		return a + b
	}
	return b
}

func complexFunc(x int) int {
	switch x {
	case 1:
		return 1
	case 2:
		return 2
	default:
		for i := 0; i < x; i++ {
			if i%2 == 0 {
				x++
			}
		}
		return x
	}
}
`
	err = os.WriteFile(goFile, []byte(goContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	config := core.AnalyzerConfig{}

	result, err := analyzer.Analyze(ctx, tempDir, config)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze returned nil result")
	}

	if result.Language != "go" {
		t.Errorf("Expected language 'go', got %s", result.Language)
	}

	if len(result.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(result.Files))
	}

	// Check file analysis
	fileAnalysis, exists := result.Files[goFile]
	if !exists {
		t.Fatal("File analysis not found")
	}

	if len(fileAnalysis.Functions) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(fileAnalysis.Functions))
	}

	// Check function names
	expectedFunctions := map[string]bool{
		"main":        false,
		"add":         false,
		"complexFunc": false,
	}

	for _, fn := range fileAnalysis.Functions {
		if _, exists := expectedFunctions[fn.Name]; exists {
			expectedFunctions[fn.Name] = true
		}
	}

	for name, found := range expectedFunctions {
		if !found {
			t.Errorf("Function %s not found in analysis", name)
		}
	}

	// Check metrics
	if result.Metrics["total_files"] != 1 {
		t.Errorf("Expected total_files to be 1, got %v", result.Metrics["total_files"])
	}

	if result.Metrics["total_functions"] != 3 {
		t.Errorf("Expected total_functions to be 3, got %v", result.Metrics["total_functions"])
	}

	// Check that complexity was calculated
	if result.Metrics["total_complexity"] == nil {
		t.Error("Expected total_complexity to be set")
	}

	if result.Metrics["average_complexity"] == nil {
		t.Error("Expected average_complexity to be set")
	}
}

func TestGoAnalyzer_AnalyzeWithContext(t *testing.T) {
	logger := &MockLogger{}
	fs := filesystem.NewOSFileSystem()
	analyzer := NewGoAnalyzer(fs, logger)

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "go-analyzer-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a Go file
	goFile := filepath.Join(tempDir, "simple.go")
	err = os.WriteFile(goFile, []byte(`package main
func simple() {}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test with canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = analyzer.Analyze(ctx, tempDir, core.AnalyzerConfig{})
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestGoAnalyzer_ComplexityCalculation(t *testing.T) {
	logger := &MockLogger{}
	fs := filesystem.NewOSFileSystem()
	analyzer := NewGoAnalyzer(fs, logger)

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "go-analyzer-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a Go file with various complexity patterns
	goFile := filepath.Join(tempDir, "complexity.go")
	goContent := `package main

func simple() {
	// Complexity: 1 (base)
}

func withIf(x int) {
	// Complexity: 2 (base + if)
	if x > 0 {
		return
	}
}

func withLoop(n int) {
	// Complexity: 2 (base + for)
	for i := 0; i < n; i++ {
		println(i)
	}
}

func withSwitch(x int) {
	// Complexity: 4 (base + switch + 2 cases)
	switch x {
	case 1:
		println("one")
	case 2:
		println("two")
	default:
		println("other")
	}
}

func complex(x int) {
	// Higher complexity with multiple conditions
	if x > 0 && x < 10 {
		for i := 0; i < x; i++ {
			if i%2 == 0 {
				println("even")
			}
		}
	}
}
`
	err = os.WriteFile(goFile, []byte(goContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	result, err := analyzer.Analyze(ctx, tempDir, core.AnalyzerConfig{})
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Check that different functions have different complexities
	complexities := make(map[string]int)
	for _, fn := range result.Functions {
		complexities[fn.Name] = fn.Complexity
	}

	// Simple function should have complexity 1
	if complexities["simple"] != 1 {
		t.Errorf("Expected simple function complexity 1, got %d", complexities["simple"])
	}

	// Functions with conditions should have higher complexity
	if complexities["withIf"] <= 1 {
		t.Errorf("Expected withIf function complexity > 1, got %d", complexities["withIf"])
	}

	if complexities["withLoop"] <= 1 {
		t.Errorf("Expected withLoop function complexity > 1, got %d", complexities["withLoop"])
	}

	if complexities["withSwitch"] <= 1 {
		t.Errorf("Expected withSwitch function complexity > 1, got %d", complexities["withSwitch"])
	}

	// Complex function should have highest complexity
	if complexities["complex"] <= complexities["simple"] {
		t.Errorf("Expected complex function to have higher complexity than simple, got %d vs %d",
			complexities["complex"], complexities["simple"])
	}
}

func TestGoAnalyzer_ExcludedFiles(t *testing.T) {
	logger := &MockLogger{}
	fs := filesystem.NewOSFileSystem()
	analyzer := NewGoAnalyzer(fs, logger)

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "go-analyzer-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create regular Go file
	goFile := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(goFile, []byte(`package main
func main() {}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create test file (should be excluded)
	testFile := filepath.Join(tempDir, "main_test.go")
	err = os.WriteFile(testFile, []byte(`package main
import "testing"
func TestMain(t *testing.T) {}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create vendor directory with Go file (should be excluded)
	vendorDir := filepath.Join(tempDir, "vendor", "github.com", "example")
	err = os.MkdirAll(vendorDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	vendorFile := filepath.Join(vendorDir, "vendor.go")
	err = os.WriteFile(vendorFile, []byte(`package example
func VendorFunc() {}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	result, err := analyzer.Analyze(ctx, tempDir, core.AnalyzerConfig{})
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Should only find the main.go file, not test or vendor files
	if len(result.Files) != 1 {
		t.Errorf("Expected 1 file (excluding test and vendor), got %d", len(result.Files))
	}

	if _, exists := result.Files[goFile]; !exists {
		t.Error("Expected main.go to be analyzed")
	}

	if _, exists := result.Files[testFile]; exists {
		t.Error("Expected test file to be excluded")
	}

	if _, exists := result.Files[vendorFile]; exists {
		t.Error("Expected vendor file to be excluded")
	}
}
