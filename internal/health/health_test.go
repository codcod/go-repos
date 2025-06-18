package health

import (
	"testing"
	"time"

	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/reporting"
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

// MockFileSystem for testing
type MockFileSystem struct {
	files map[string][]byte
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
	}
}

func (fs *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if data, exists := fs.files[path]; exists {
		return data, nil
	}
	return nil, &FileNotFoundError{Path: path}
}

func (fs *MockFileSystem) WriteFile(path string, data []byte) error {
	fs.files[path] = data
	return nil
}

func (fs *MockFileSystem) Exists(path string) bool {
	_, exists := fs.files[path]
	return exists
}

func (fs *MockFileSystem) IsDir(path string) bool {
	return false
}

func (fs *MockFileSystem) ListFiles(path string, pattern string) ([]string, error) {
	var files []string
	for filepath := range fs.files {
		files = append(files, filepath)
	}
	return files, nil
}

func (fs *MockFileSystem) Walk(path string, walkFn func(path string, info core.FileInfo) error) error {
	for filepath := range fs.files {
		info := core.FileInfo{
			Name:    filepath,
			Size:    int64(len(fs.files[filepath])),
			Mode:    0644,
			ModTime: time.Now(),
			IsDir:   false,
		}
		if err := walkFn(filepath, info); err != nil {
			return err
		}
	}
	return nil
}

// FileNotFoundError for testing
type FileNotFoundError struct {
	Path string
}

func (e *FileNotFoundError) Error() string {
	return "file not found: " + e.Path
}

// Test NewAnalyzerRegistry
func TestNewAnalyzerRegistry(t *testing.T) {
	fs := NewMockFileSystem()
	logger := &MockLogger{}

	registry := NewAnalyzerRegistry(fs, logger)

	if registry == nil {
		t.Error("Expected NewAnalyzerRegistry to return non-nil registry")
	}

	// Test that standard analyzers are registered
	languages := registry.GetSupportedLanguages()
	if len(languages) == 0 {
		t.Error("Expected some languages to be supported")
	}

	// Check for common languages
	expectedLanguages := []string{"go", "python", "java", "javascript"}
	for _, expectedLang := range expectedLanguages {
		found := false
		for _, lang := range languages {
			if lang == expectedLang {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s to be supported, but it wasn't found", expectedLang)
		}
	}
}

// Test NewFileSystem
func TestNewFileSystem(t *testing.T) {
	fs := NewFileSystem()

	if fs == nil {
		t.Error("Expected NewFileSystem to return non-nil filesystem")
	}

	// Test that it implements the FileSystem interface
	var _ = fs
}

// Test NewCommandExecutor
func TestNewCommandExecutor(t *testing.T) {
	timeout := 30 * time.Second
	executor := NewCommandExecutor(timeout)

	if executor == nil {
		t.Error("Expected NewCommandExecutor to return non-nil executor")
	}
}

// Test factory function integration
func TestFactoryFunctions_Integration(t *testing.T) {
	// Create dependencies
	fs := NewFileSystem()
	logger := &MockLogger{}
	timeout := 10 * time.Second

	// Test that all factory functions work together
	analyzerRegistry := NewAnalyzerRegistry(fs, logger)
	if analyzerRegistry == nil {
		t.Error("Failed to create analyzer registry")
	}

	commandExecutor := NewCommandExecutor(timeout)
	if commandExecutor == nil {
		t.Error("Failed to create command executor")
	}

	checkerRegistry := NewCheckerRegistry(commandExecutor)
	if checkerRegistry == nil {
		t.Error("Failed to create checker registry")
	}

	// Verify that the registries have expected content
	if len(analyzerRegistry.GetSupportedLanguages()) == 0 {
		t.Error("Analyzer registry should have some supported languages")
	}

	if len(checkerRegistry.GetCheckers()) == 0 {
		t.Error("Checker registry should have some available checkers")
	}
}

// Test type aliases
func TestTypeAliases(t *testing.T) {
	// Test that type aliases are correctly defined by creating instances
	fs := NewMockFileSystem()
	logger := &MockLogger{}
	timeout := 5 * time.Second

	var analyzerReg = NewAnalyzerRegistry(fs, logger)
	if analyzerReg == nil {
		t.Error("AnalyzerRegistry type alias not working")
	}

	var checkerReg = NewCheckerRegistry(NewCommandExecutor(timeout))
	if checkerReg == nil {
		t.Error("CheckerRegistry type alias not working")
	}

	var formatter = reporting.NewFormatter(true)
	if formatter == nil {
		t.Error("Formatter type alias not working")
	}
}

// Test error handling in factory functions
func TestFactoryFunctions_ErrorHandling(t *testing.T) {
	// Test with nil dependencies (should not panic)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Factory function panicked: %v", r)
		}
	}()

	// Test NewAnalyzerRegistry with nil logger (if it handles gracefully)
	fs := NewMockFileSystem()
	registry := NewAnalyzerRegistry(fs, nil)
	if registry == nil {
		t.Error("Expected NewAnalyzerRegistry to handle nil logger gracefully")
	}

	// Test NewCommandExecutor with zero timeout
	executor := NewCommandExecutor(0)
	if executor == nil {
		t.Error("Expected NewCommandExecutor to handle zero timeout")
	}
}

// Test package-level constants and configuration
func TestPackageConfiguration(t *testing.T) {
	// Test that we can create all the main components
	fs := NewFileSystem()
	logger := &MockLogger{}
	executor := NewCommandExecutor(30 * time.Second)

	analyzerReg := NewAnalyzerRegistry(fs, logger)
	checkerReg := NewCheckerRegistry(executor)

	// Verify they're properly initialized
	if analyzerReg == nil || checkerReg == nil {
		t.Error("Failed to create health components")
	}

	// Test that they have reasonable defaults
	languages := analyzerReg.GetSupportedLanguages()
	checkers := checkerReg.GetCheckers()

	if len(languages) < 2 {
		t.Errorf("Expected at least 2 supported languages, got %d", len(languages))
	}

	if len(checkers) < 1 {
		t.Errorf("Expected at least 1 available checker, got %d", len(checkers))
	}
}

// Benchmark factory functions
func BenchmarkNewAnalyzerRegistry(b *testing.B) {
	fs := NewMockFileSystem()
	logger := &MockLogger{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry := NewAnalyzerRegistry(fs, logger)
		if registry == nil {
			b.Fatal("Failed to create analyzer registry")
		}
	}
}

func BenchmarkNewFileSystem(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs := NewFileSystem()
		if fs == nil {
			b.Fatal("Failed to create filesystem")
		}
	}
}

func BenchmarkNewCommandExecutor(b *testing.B) {
	timeout := 30 * time.Second

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor := NewCommandExecutor(timeout)
		if executor == nil {
			b.Fatal("Failed to create command executor")
		}
	}
}
