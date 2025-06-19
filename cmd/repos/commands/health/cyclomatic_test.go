package health

import (
	"testing"
)

func TestNewCyclomaticComplexityCommand(t *testing.T) {
	cmd := NewCyclomaticComplexityCommand()

	// Test command properties
	if cmd.Use != "cyclomatic" {
		t.Errorf("Expected command use to be 'cyclomatic', got %q", cmd.Use)
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

func TestComplexityFlags(t *testing.T) {
	cmd := NewCyclomaticComplexityCommand()
	flags := cmd.Flags()

	// Test required flags exist
	expectedFlags := []string{
		"tag",
		"max-complexity",
	}

	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %s to exist", flagName)
		}
	}

	// Test default values
	maxComplexityFlag := flags.Lookup("max-complexity")
	if maxComplexityFlag != nil && maxComplexityFlag.DefValue != "10" {
		t.Errorf("Expected default max-complexity to be '10', got %q", maxComplexityFlag.DefValue)
	}
}

func TestComplexityConfig(t *testing.T) {
	config := &ComplexityConfig{
		Tag:           "test",
		MaxComplexity: 15,
	}

	if config.Tag != "test" {
		t.Errorf("Expected tag to be 'test', got %q", config.Tag)
	}

	if config.MaxComplexity != 15 {
		t.Errorf("Expected max complexity to be 15, got %d", config.MaxComplexity)
	}
}
