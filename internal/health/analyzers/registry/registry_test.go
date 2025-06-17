package registry

import (
	"context"
	"strings"
	"testing"

	"github.com/codcod/repos/internal/core"
)

// MockAnalyzer is a mock implementation of core.Analyzer for testing
type MockAnalyzer struct {
	name          string
	language      string
	extensions    []string
	canAnalyze    bool
	analyzeResult *core.AnalysisResult
	analyzeError  error
}

func NewMockAnalyzer(name, language string, extensions []string) *MockAnalyzer {
	return &MockAnalyzer{
		name:       name,
		language:   language,
		extensions: extensions,
		canAnalyze: true,
	}
}

func (m *MockAnalyzer) Name() string {
	return m.name
}

func (m *MockAnalyzer) Language() string {
	return m.language
}

func (m *MockAnalyzer) SupportedExtensions() []string {
	return m.extensions
}

func (m *MockAnalyzer) CanAnalyze(repo core.Repository) bool {
	return m.canAnalyze
}

func (m *MockAnalyzer) SetCanAnalyze(can bool) {
	m.canAnalyze = can
}

func (m *MockAnalyzer) SetAnalyzeResult(result *core.AnalysisResult, err error) {
	m.analyzeResult = result
	m.analyzeError = err
}

func (m *MockAnalyzer) Analyze(ctx context.Context, repoPath string, config core.AnalyzerConfig) (*core.AnalysisResult, error) {
	return m.analyzeResult, m.analyzeError
}

// Test helper functions
func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func assertTrue(t *testing.T, condition bool) {
	t.Helper()
	if !condition {
		t.Error("Expected true, got false")
	}
}

func assertFalse(t *testing.T, condition bool) {
	t.Helper()
	if condition {
		t.Error("Expected false, got true")
	}
}

func assertNil(t *testing.T, value interface{}) {
	t.Helper()
	if value != nil {
		t.Errorf("Expected nil, got %v", value)
	}
}

func assertNotNil(t *testing.T, value interface{}) {
	t.Helper()
	if value == nil {
		t.Error("Expected non-nil value")
	}
}

func assertLen(t *testing.T, slice interface{}, expectedLen int) {
	t.Helper()
	var actualLen int
	switch s := slice.(type) {
	case []core.Analyzer:
		actualLen = len(s)
	case map[string]core.Analyzer:
		actualLen = len(s)
	case []string:
		actualLen = len(s)
	default:
		t.Errorf("Unsupported type for length check: %T", slice)
		return
	}
	if actualLen != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, actualLen)
	}
}

func assertContains(t *testing.T, slice []core.Analyzer, item core.Analyzer) {
	t.Helper()
	for _, s := range slice {
		if s == item {
			return
		}
	}
	t.Errorf("Expected slice to contain %v", item)
}

func assertNotContains(t *testing.T, slice []core.Analyzer, item core.Analyzer) {
	t.Helper()
	for _, s := range slice {
		if s == item {
			t.Errorf("Expected slice not to contain %v", item)
			return
		}
	}
}

func assertContainsString(t *testing.T, slice []string, item string) {
	t.Helper()
	for _, s := range slice {
		if s == item {
			return
		}
	}
	t.Errorf("Expected slice to contain %s", item)
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func assertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Error("Expected error, got nil")
		return
	}
	if !strings.Contains(err.Error(), substr) {
		t.Errorf("Expected error to contain %q, got %q", substr, err.Error())
	}
}

// Test Registry creation
func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	assertNotNil(t, registry)
	assertNotNil(t, registry.analyzers)
	assertLen(t, registry.analyzers, 0)
}

// Test analyzer registration
func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	analyzer := NewMockAnalyzer("go-analyzer", "go", []string{".go"})

	registry.Register(analyzer)

	assertLen(t, registry.analyzers, 1)
	if _, exists := registry.analyzers["go"]; !exists {
		t.Error("Expected registry to contain 'go' analyzer")
	}
	if registry.analyzers["go"] != analyzer {
		t.Error("Expected registered analyzer to match the original")
	}
}

// Test analyzer unregistration
func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	analyzer := NewMockAnalyzer("go-analyzer", "go", []string{".go"})

	registry.Register(analyzer)
	assertLen(t, registry.analyzers, 1)

	registry.Unregister("go")
	assertLen(t, registry.analyzers, 0)
	if _, exists := registry.analyzers["go"]; exists {
		t.Error("Expected registry not to contain 'go' analyzer after unregistration")
	}
}

// Test unregistering non-existent analyzer
func TestRegistry_Unregister_NonExistent(t *testing.T) {
	registry := NewRegistry()

	// Should not panic
	registry.Unregister("nonexistent")
	assertLen(t, registry.analyzers, 0)
}

// Test getting analyzer by language
func TestRegistry_GetByLanguage(t *testing.T) {
	registry := NewRegistry()
	analyzer := NewMockAnalyzer("go-analyzer", "go", []string{".go"})
	registry.Register(analyzer)

	// Test existing analyzer
	found, exists := registry.GetByLanguage("go")
	assertTrue(t, exists)
	assertEqual(t, analyzer, found)

	// Test non-existing analyzer
	notFound, exists := registry.GetByLanguage("rust")
	assertFalse(t, exists)
	assertNil(t, notFound)
}

// Test GetAnalyzer method (AnalyzerRegistry interface)
func TestRegistry_GetAnalyzer(t *testing.T) {
	registry := NewRegistry()
	analyzer := NewMockAnalyzer("go-analyzer", "go", []string{".go"})
	registry.Register(analyzer)

	// Test existing analyzer
	found, err := registry.GetAnalyzer("go")
	assertNoError(t, err)
	assertEqual(t, analyzer, found)

	// Test non-existing analyzer
	notFound, err := registry.GetAnalyzer("rust")
	assertError(t, err)
	assertNil(t, notFound)
	assertErrorContains(t, err, "analyzer not found for language: rust")
}

// Test GetAnalyzers method (AnalyzerRegistry interface)
func TestRegistry_GetAnalyzers(t *testing.T) {
	registry := NewRegistry()

	// Empty registry
	analyzers := registry.GetAnalyzers()
	assertLen(t, analyzers, 0)

	// Add analyzers
	goAnalyzer := NewMockAnalyzer("go-analyzer", "go", []string{".go"})
	pythonAnalyzer := NewMockAnalyzer("python-analyzer", "python", []string{".py"})
	registry.Register(goAnalyzer)
	registry.Register(pythonAnalyzer)

	analyzers = registry.GetAnalyzers()
	assertLen(t, analyzers, 2)
	assertContains(t, analyzers, goAnalyzer)
	assertContains(t, analyzers, pythonAnalyzer)
}

// Test getting analyzer by file extension
func TestRegistry_GetByFileExtension(t *testing.T) {
	registry := NewRegistry()
	goAnalyzer := NewMockAnalyzer("go-analyzer", "go", []string{".go"})
	jsAnalyzer := NewMockAnalyzer("js-analyzer", "javascript", []string{".js", ".ts"})

	registry.Register(goAnalyzer)
	registry.Register(jsAnalyzer)

	// Test existing extensions
	found, exists := registry.GetByFileExtension(".go")
	assertTrue(t, exists)
	assertEqual(t, goAnalyzer, found)

	found, exists = registry.GetByFileExtension(".js")
	assertTrue(t, exists)
	assertEqual(t, jsAnalyzer, found)

	found, exists = registry.GetByFileExtension(".ts")
	assertTrue(t, exists)
	assertEqual(t, jsAnalyzer, found)

	// Test non-existing extension
	notFound, exists := registry.GetByFileExtension(".rs")
	assertFalse(t, exists)
	assertNil(t, notFound)
}

// Test getting supported analyzers for a repository
func TestRegistry_GetSupportedAnalyzers(t *testing.T) {
	registry := NewRegistry()

	goAnalyzer := NewMockAnalyzer("go-analyzer", "go", []string{".go"})
	pythonAnalyzer := NewMockAnalyzer("python-analyzer", "python", []string{".py"})
	jsAnalyzer := NewMockAnalyzer("js-analyzer", "javascript", []string{".js"})

	// Set up which analyzers can analyze the repo
	goAnalyzer.SetCanAnalyze(true)
	pythonAnalyzer.SetCanAnalyze(false)
	jsAnalyzer.SetCanAnalyze(true)

	registry.Register(goAnalyzer)
	registry.Register(pythonAnalyzer)
	registry.Register(jsAnalyzer)

	repo := core.Repository{
		Name:     "test-repo",
		Language: "go",
	}

	supported := registry.GetSupportedAnalyzers(repo)
	assertLen(t, supported, 2)
	assertContains(t, supported, goAnalyzer)
	assertContains(t, supported, jsAnalyzer)
	assertNotContains(t, supported, pythonAnalyzer)
}

// Test getting all analyzers
func TestRegistry_GetAllAnalyzers(t *testing.T) {
	registry := NewRegistry()

	// Empty registry
	analyzers := registry.GetAllAnalyzers()
	assertLen(t, analyzers, 0)

	// Add analyzers
	goAnalyzer := NewMockAnalyzer("go-analyzer", "go", []string{".go"})
	pythonAnalyzer := NewMockAnalyzer("python-analyzer", "python", []string{".py"})
	registry.Register(goAnalyzer)
	registry.Register(pythonAnalyzer)

	analyzers = registry.GetAllAnalyzers()
	assertLen(t, analyzers, 2)
	assertContains(t, analyzers, goAnalyzer)
	assertContains(t, analyzers, pythonAnalyzer)
}

// Test getting supported languages
func TestRegistry_GetSupportedLanguages(t *testing.T) {
	registry := NewRegistry()

	// Empty registry
	languages := registry.GetSupportedLanguages()
	assertLen(t, languages, 0)

	// Add analyzers
	registry.Register(NewMockAnalyzer("go-analyzer", "go", []string{".go"}))
	registry.Register(NewMockAnalyzer("python-analyzer", "python", []string{".py"}))
	registry.Register(NewMockAnalyzer("js-analyzer", "javascript", []string{".js"}))

	languages = registry.GetSupportedLanguages()
	assertLen(t, languages, 3)
	assertContainsString(t, languages, "go")
	assertContainsString(t, languages, "python")
	assertContainsString(t, languages, "javascript")
}

// Test BaseAnalyzer
func TestBaseAnalyzer(t *testing.T) {
	extensions := []string{".go", ".mod"}
	analyzer := NewBaseAnalyzer("go", extensions, true, true)

	assertEqual(t, "go", analyzer.Language())
	if len(analyzer.FileExtensions()) != len(extensions) {
		t.Errorf("Expected %d extensions, got %d", len(extensions), len(analyzer.FileExtensions()))
	}
	assertTrue(t, analyzer.SupportsComplexity())
	assertTrue(t, analyzer.SupportsFunctionLevel())
}

// Test BaseAnalyzer default implementations
func TestBaseAnalyzer_DefaultImplementations(t *testing.T) {
	analyzer := NewBaseAnalyzer("go", []string{".go"}, true, true)
	ctx := context.Background()

	// Test default AnalyzeComplexity
	result, err := analyzer.AnalyzeComplexity(ctx, "/test/path")
	assertNoError(t, err)
	// Check individual fields instead of direct comparison
	if result.TotalFiles != 0 {
		t.Errorf("Expected TotalFiles to be 0, got %d", result.TotalFiles)
	}
	if result.TotalFunctions != 0 {
		t.Errorf("Expected TotalFunctions to be 0, got %d", result.TotalFunctions)
	}
	if result.AverageComplexity != 0 {
		t.Errorf("Expected AverageComplexity to be 0, got %f", result.AverageComplexity)
	}
	if result.MaxComplexity != 0 {
		t.Errorf("Expected MaxComplexity to be 0, got %d", result.MaxComplexity)
	}
	// Note: slices and maps in structs are initialized as nil, which is correct

	// Test default AnalyzeFunctions
	functions, err := analyzer.AnalyzeFunctions(ctx, "/test/path")
	assertNoError(t, err)
	if functions != nil {
		t.Errorf("Expected functions to be nil, got %v", functions)
	}

	// Test default DetectPatterns
	patterns, err := analyzer.DetectPatterns(ctx, "content", []core.Pattern{})
	assertNoError(t, err)
	if patterns != nil {
		t.Errorf("Expected patterns to be nil, got %v", patterns)
	}
}

// Test analyzer replacement (registering same language twice)
func TestRegistry_ReplaceAnalyzer(t *testing.T) {
	registry := NewRegistry()

	// Register first analyzer
	analyzer1 := NewMockAnalyzer("go-analyzer-1", "go", []string{".go"})
	registry.Register(analyzer1)

	found, exists := registry.GetByLanguage("go")
	assertTrue(t, exists)
	assertEqual(t, analyzer1, found)

	// Register second analyzer with same language (should replace)
	analyzer2 := NewMockAnalyzer("go-analyzer-2", "go", []string{".go", ".mod"})
	registry.Register(analyzer2)

	found, exists = registry.GetByLanguage("go")
	assertTrue(t, exists)
	assertEqual(t, analyzer2, found)
	if found == analyzer1 {
		t.Error("Expected analyzer to be replaced")
	}

	// Should still have only one analyzer
	assertLen(t, registry.analyzers, 1)
}

// Test multiple analyzers with same extension
func TestRegistry_MultipleAnalyzersWithSameExtension(t *testing.T) {
	registry := NewRegistry()

	// Both analyzers support .js files
	jsAnalyzer := NewMockAnalyzer("js-analyzer", "javascript", []string{".js"})
	tsAnalyzer := NewMockAnalyzer("ts-analyzer", "typescript", []string{".js", ".ts"})

	registry.Register(jsAnalyzer)
	registry.Register(tsAnalyzer)

	// GetByFileExtension should return the first matching analyzer found
	found, exists := registry.GetByFileExtension(".js")
	assertTrue(t, exists)
	assertNotNil(t, found)
	// Could be either analyzer - depends on iteration order
	if found != jsAnalyzer && found != tsAnalyzer {
		t.Error("Expected to find either jsAnalyzer or tsAnalyzer")
	}
}

// Test edge cases
func TestRegistry_EdgeCases(t *testing.T) {
	registry := NewRegistry()

	// Test with empty extension
	found, exists := registry.GetByFileExtension("")
	assertFalse(t, exists)
	assertNil(t, found)

	// Test with nil repository
	var repo core.Repository
	supported := registry.GetSupportedAnalyzers(repo)
	assertLen(t, supported, 0)

	// Test unregistering from empty registry
	registry.Unregister("nonexistent")
	assertLen(t, registry.analyzers, 0)
}

// Test factory function tests would require actual analyzer implementations
// For now, we'll test the basic behavior with mocks

// Mock implementations for FileSystem and Logger for factory tests
type MockFileSystem struct{}

func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	return nil, nil
}

func (m *MockFileSystem) WriteFile(path string, data []byte) error {
	return nil
}

func (m *MockFileSystem) Exists(path string) bool {
	return false
}

func (m *MockFileSystem) IsDir(path string) bool {
	return false
}

func (m *MockFileSystem) ListFiles(path string, pattern string) ([]string, error) {
	return nil, nil
}

func (m *MockFileSystem) Walk(path string, walkFn func(path string, info core.FileInfo) error) error {
	return nil
}

type MockLogger struct{}

func (m *MockLogger) Info(msg string, args ...interface{})  {}
func (m *MockLogger) Error(msg string, args ...interface{}) {}
func (m *MockLogger) Debug(msg string, args ...interface{}) {}
func (m *MockLogger) Warn(msg string, args ...interface{})  {}

// Test registry interface compliance
func TestRegistry_ImplementsAnalyzerRegistry(t *testing.T) {
	registry := NewRegistry()

	// Verify it implements core.AnalyzerRegistry interface
	var _ core.AnalyzerRegistry = registry

	// Test interface methods work
	languages := registry.GetSupportedLanguages()
	assertLen(t, languages, 0)

	analyzers := registry.GetAnalyzers()
	assertLen(t, analyzers, 0)

	_, err := registry.GetAnalyzer("nonexistent")
	assertError(t, err)
}

// Test registry methods consistency
func TestRegistry_MethodConsistency(t *testing.T) {
	registry := NewRegistry()
	analyzer := NewMockAnalyzer("test-analyzer", "test", []string{".test"})

	registry.Register(analyzer)

	// GetAnalyzers and GetAllAnalyzers should return the same results
	analyzers1 := registry.GetAnalyzers()
	analyzers2 := registry.GetAllAnalyzers()

	assertLen(t, analyzers1, 1)
	assertLen(t, analyzers2, 1)
	assertEqual(t, analyzers1[0], analyzers2[0])

	// GetByLanguage and GetAnalyzer should be consistent
	found1, exists := registry.GetByLanguage("test")
	assertTrue(t, exists)

	found2, err := registry.GetAnalyzer("test")
	assertNoError(t, err)

	assertEqual(t, found1, found2)
}
