package health

import (
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

func TestGenConfigCommandFlags(t *testing.T) {
	cmd := NewGenConfigCommand()
	flags := cmd.Flags()

	// Test required flags exist
	expectedFlags := []string{
		"output",
		"overwrite",
	}

	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %s to exist", flagName)
		}
	}

	// Test default values
	outputFlag := flags.Lookup("output")
	if outputFlag != nil && outputFlag.DefValue != "health-config.yaml" {
		t.Errorf("Expected default output to be 'health-config.yaml', got %q", outputFlag.DefValue)
	}

	overwriteFlag := flags.Lookup("overwrite")
	if overwriteFlag != nil && overwriteFlag.DefValue != "false" {
		t.Errorf("Expected default overwrite to be 'false', got %q", overwriteFlag.DefValue)
	}
}

func TestGenConfigConfig(t *testing.T) {
	config := &GenConfigConfig{
		OutputFile: "test-health.yaml",
		Overwrite:  true,
	}

	if config.OutputFile != "test-health.yaml" {
		t.Errorf("Expected output file to be 'test-health.yaml', got %q", config.OutputFile)
	}

	if !config.Overwrite {
		t.Error("Expected overwrite to be true")
	}
}
