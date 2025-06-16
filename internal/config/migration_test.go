package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codcod/repos/internal/core"
)

// mockLogger for testing
type mockLogger struct {
	messages []string
}

func (l *mockLogger) Debug(msg string, fields ...core.Field) {
	l.messages = append(l.messages, msg)
}

func (l *mockLogger) Info(msg string, fields ...core.Field) {
	l.messages = append(l.messages, msg)
}

func (l *mockLogger) Warn(msg string, fields ...core.Field) {
	l.messages = append(l.messages, msg)
}

func (l *mockLogger) Error(msg string, fields ...core.Field) {
	l.messages = append(l.messages, msg)
}

func TestMigrationManager_LoadConfig(t *testing.T) {
	// Create temporary directory for test configs
	tempDir, err := os.MkdirTemp("", "migration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create legacy config
	legacyConfig := `
repositories:
  - name: test-repo
    url: https://github.com/test/repo
    branch: main

checkers:
  enabled: ["git", "dependencies"]
  timeout: 30s

complexity:
  threshold: 10
  languages:
    python: 8
    java: 12
`

	legacyPath := filepath.Join(tempDir, "legacy-config.yaml")
	err = os.WriteFile(legacyPath, []byte(legacyConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write legacy config: %v", err)
	}

	// Test migration manager
	logger := &mockLogger{}
	manager := NewMigrationManager(logger)

	// Load config with migration
	config, err := manager.LoadConfig(legacyPath)
	if err != nil {
		t.Fatalf("Failed to load config with migration: %v", err)
	}

	// Verify it's an advanced config
	advConfig, ok := config.(*AdvancedConfig)
	if !ok {
		t.Fatalf("Expected AdvancedConfig, got %T", config)
	}

	// Verify basic properties
	if advConfig.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", advConfig.Version)
	}

	// Check if migrated file was created
	advancedPath := filepath.Join(tempDir, "legacy-config-advanced.yaml")
	if _, err := os.Stat(advancedPath); os.IsNotExist(err) {
		t.Errorf("Expected migrated config file to be created at %s", advancedPath)
	}
}

func TestConfigMigrator_DetectConfigFormat(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "format_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger := &mockLogger{}
	migrator := NewConfigMigrator(logger)

	// Test legacy format detection
	legacyConfig := `
repositories:
  - name: test-repo
    url: https://github.com/test/repo
`
	legacyPath := filepath.Join(tempDir, "legacy.yaml")
	err = os.WriteFile(legacyPath, []byte(legacyConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write legacy config: %v", err)
	}

	format, err := migrator.DetectConfigFormat(legacyPath)
	if err != nil {
		t.Fatalf("Failed to detect format: %v", err)
	}
	if format != "legacy" {
		t.Errorf("Expected 'legacy', got '%s'", format)
	}

	// Test advanced format detection
	advancedConfig := `
version: "1.0"
engine:
  max_concurrency: 4
profiles:
  default:
    name: "Default Profile"
`
	advancedPath := filepath.Join(tempDir, "advanced.yaml")
	err = os.WriteFile(advancedPath, []byte(advancedConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write advanced config: %v", err)
	}

	format, err = migrator.DetectConfigFormat(advancedPath)
	if err != nil {
		t.Fatalf("Failed to detect format: %v", err)
	}
	if format != "advanced" {
		t.Errorf("Expected 'advanced', got '%s'", format)
	}
}

func TestGetExtensionsForLanguage(t *testing.T) {
	tests := []struct {
		language string
		expected []string
	}{
		{"go", []string{".go"}},
		{"python", []string{".py"}},
		{"javascript", []string{".js", ".jsx", ".ts", ".tsx"}},
		{"unknown", []string{}},
	}

	for _, test := range tests {
		result := getExtensionsForLanguage(test.language)
		if len(result) != len(test.expected) {
			t.Errorf("For language %s, expected %v, got %v", test.language, test.expected, result)
			continue
		}
		for i, ext := range test.expected {
			if result[i] != ext {
				t.Errorf("For language %s, expected extension %s at index %d, got %s", test.language, ext, i, result[i])
			}
		}
	}
}
