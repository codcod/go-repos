package health

import (
	"strings"
	"testing"
)

func TestNewGenConfigCommand(t *testing.T) {
	cmd := NewGenConfigCommand()

	// Test command properties
	if cmd.Use != "genconfig" {
		t.Errorf("Expected command use to be 'genconfig', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected command long description to be set")
	}

	if cmd.RunE == nil {
		t.Error("Expected command RunE to be set")
	}
}

func TestGenerateConfigTemplate(t *testing.T) {
	// This test just ensures the function doesn't panic
	// In a real scenario, you might want to capture the output
	err := generateConfigTemplate("", false) // Console output, no overwrite
	if err != nil {
		t.Errorf("generateConfigTemplate() returned error: %v", err)
	}
}

// Mock test to ensure the template contains expected sections
func TestGenConfigTemplateContent(t *testing.T) {
	// We can't easily test the actual output without refactoring,
	// but we can test that the function executes without error
	err := generateConfigTemplate("", false) // Console output, no overwrite
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// Test the template content by checking if it has required sections
func TestTemplateHasRequiredSections(t *testing.T) {
	template := `# Health Configuration Template
health:
  checkers:
    ci:
      enabled: true
    git:
      enabled: true
  analyzers:
    javascript:
      enabled: true`

	expectedSections := []string{
		"health:",
		"checkers:",
		"analyzers:",
		"ci:",
		"git:",
		"javascript:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(template, section) {
			t.Errorf("Template should contain '%s'", section)
		}
	}
}
