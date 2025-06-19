package health

import (
	"testing"

	"github.com/codcod/repos/internal/health/analyzers/common"
)

func TestNewFactorySystem(t *testing.T) {
	logger := &common.MockLogger{}

	// Test that we can create an analyzer map using the new factory system
	analyzerMap := NewAnalyzerMap(logger)

	if len(analyzerMap) == 0 {
		t.Error("Expected analyzer map to have at least one analyzer")
	}

	// Test that all expected languages are supported
	expectedLanguages := []string{"go", "python", "java", "javascript"}
	for _, lang := range expectedLanguages {
		if analyzer, exists := analyzerMap[lang]; !exists {
			t.Errorf("Expected language '%s' to be in analyzer map", lang)
		} else {
			if analyzer.Language() != lang {
				t.Errorf("Expected analyzer for language '%s', got '%s'", lang, analyzer.Language())
			}
		}
	}
}

func TestNewAnalyzerRegistry_BackwardCompatibility(t *testing.T) {
	logger := &common.MockLogger{}

	// Test that the legacy registry interface still works
	registry := NewAnalyzerRegistry(logger)

	if registry == nil {
		t.Fatal("Expected registry, got nil")
	}

	// Test that we can get supported languages
	languages := registry.GetSupportedLanguages()
	if len(languages) == 0 {
		t.Error("Expected at least one supported language")
	}

	// Test that we can get an analyzer
	analyzer, err := registry.GetAnalyzer("go")
	if err != nil {
		t.Fatalf("Failed to get Go analyzer: %v", err)
	}

	if analyzer.Language() != "go" {
		t.Errorf("Expected language 'go', got '%s'", analyzer.Language())
	}
}
