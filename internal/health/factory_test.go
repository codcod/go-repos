package health

import (
	"testing"

	"github.com/codcod/repos/internal/health/analyzers/common"
)

func TestNewFactorySystem(t *testing.T) {
	logger := &common.MockLogger{}

	// Test that we can create an analyzer registry using the new factory system
	registry := NewAnalyzerRegistry(logger)

	// Test that all expected languages are supported
	expectedLanguages := []string{"go", "python", "java", "javascript"}
	for _, lang := range expectedLanguages {
		analyzer, err := registry.GetAnalyzer(lang)
		if err != nil {
			t.Errorf("Failed to get analyzer for language '%s': %v", lang, err)
			continue
		}
		if analyzer.Language() != lang {
			t.Errorf("Expected analyzer for language '%s', got '%s'", lang, analyzer.Language())
		}
	}

	// Test that supported languages include our expected languages
	supportedLanguages := registry.GetSupportedLanguages()
	if len(supportedLanguages) == 0 {
		t.Error("Expected at least one supported language")
	}

	for _, expectedLang := range expectedLanguages {
		found := false
		for _, supportedLang := range supportedLanguages {
			if supportedLang == expectedLang {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected language '%s' to be supported", expectedLang)
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
