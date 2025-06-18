package config

import (
	"testing"
)

func TestNewDefaultAdvancedConfig(t *testing.T) {
	config := NewDefaultAdvancedConfig()

	if config == nil {
		t.Fatal("NewDefaultAdvancedConfig should not return nil")
	}

	if config.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", config.Version)
	}

	if config.Engine.MaxConcurrency != 4 {
		t.Errorf("Expected MaxConcurrency 4, got %d", config.Engine.MaxConcurrency)
	}

	// Check that categories are initialized
	if len(config.Categories) == 0 {
		t.Error("Categories should be initialized with default values")
	}

	expectedCategories := []string{"security", "quality", "compliance", "ci", "docs"}
	for _, cat := range expectedCategories {
		if _, exists := config.Categories[cat]; !exists {
			t.Errorf("Expected category '%s' to exist", cat)
		}
	}
}

func TestLoadAdvancedConfigOrDefault(t *testing.T) {
	// Test with non-existent file (should return default config)
	config, err := LoadAdvancedConfigOrDefault("non-existent-file.yaml")
	if err != nil {
		t.Fatalf("LoadAdvancedConfigOrDefault should not error for non-existent file: %v", err)
	}

	if config == nil {
		t.Fatal("Config should not be nil")
	}

	// Should have default values
	if config.Version != "1.0" {
		t.Errorf("Expected default version '1.0', got '%s'", config.Version)
	}

	if len(config.Categories) == 0 {
		t.Error("Default config should have categories")
	}
}
