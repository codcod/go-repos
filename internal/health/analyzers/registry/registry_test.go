package registry

import (
	"context"
	"testing"

	"github.com/codcod/repos/internal/core"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("Expected registry, got nil")
	}

	// Initially should be empty
	languages := registry.GetSupportedLanguages()
	if len(languages) != 0 {
		t.Errorf("Expected 0 languages initially, got %d", len(languages))
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	// Create a mock analyzer that implements the interface
	mockAnalyzer := &TestAnalyzer{
		name:       "test-analyzer",
		language:   "test",
		extensions: []string{".test"},
	}

	registry.Register(mockAnalyzer)

	// Check that it was registered
	languages := registry.GetSupportedLanguages()
	if len(languages) != 1 {
		t.Errorf("Expected 1 language, got %d", len(languages))
	}

	if languages[0] != "test" {
		t.Errorf("Expected language 'test', got '%s'", languages[0])
	}

	// Check that we can retrieve it
	analyzer, err := registry.GetAnalyzer("test")
	if err != nil {
		t.Fatalf("Failed to get analyzer: %v", err)
	}

	if analyzer.Language() != "test" {
		t.Errorf("Expected language 'test', got '%s'", analyzer.Language())
	}
}

func TestRegistry_GetAnalyzer(t *testing.T) {
	registry := NewRegistry()

	// Test getting non-existent analyzer
	_, err := registry.GetAnalyzer("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent analyzer")
	}

	// Register an analyzer and test retrieval
	mockAnalyzer := &TestAnalyzer{
		name:       "test-analyzer",
		language:   "test",
		extensions: []string{".test"},
	}

	registry.Register(mockAnalyzer)

	analyzer, err := registry.GetAnalyzer("test")
	if err != nil {
		t.Fatalf("Failed to get analyzer: %v", err)
	}

	if analyzer.Language() != "test" {
		t.Errorf("Expected language 'test', got '%s'", analyzer.Language())
	}
}

// TestAnalyzer is a simple test implementation that satisfies the core.Analyzer interface
type TestAnalyzer struct {
	name       string
	language   string
	extensions []string
}

func (t *TestAnalyzer) Name() string {
	return t.name
}

func (t *TestAnalyzer) Language() string {
	return t.language
}

func (t *TestAnalyzer) SupportedExtensions() []string {
	return t.extensions
}

func (t *TestAnalyzer) CanAnalyze(repo core.Repository) bool {
	return repo.Language == t.language
}

func (t *TestAnalyzer) Analyze(ctx context.Context, repoPath string, config core.AnalyzerConfig) (*core.AnalysisResult, error) {
	return &core.AnalysisResult{
		Language: t.language,
		Files:    make(map[string]*core.FileAnalysis),
		Metrics:  make(map[string]interface{}),
	}, nil
}
