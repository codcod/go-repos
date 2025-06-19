package init

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	// Test command properties
	if cmd.Use != "init" {
		t.Errorf("Expected command use to be 'init', got %q", cmd.Use)
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

func TestInitCommandFlags(t *testing.T) {
	cmd := NewCommand()
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
	if outputFlag != nil && outputFlag.DefValue != "config.yaml" {
		t.Errorf("Expected default output to be 'config.yaml', got %q", outputFlag.DefValue)
	}

	overwriteFlag := flags.Lookup("overwrite")
	if overwriteFlag != nil && overwriteFlag.DefValue != "false" {
		t.Errorf("Expected default overwrite to be 'false', got %q", overwriteFlag.DefValue)
	}
}

func TestConfig(t *testing.T) {
	config := &Config{
		OutputFile: "test.yaml",
		Overwrite:  true,
	}

	if config.OutputFile != "test.yaml" {
		t.Errorf("Expected output file to be 'test.yaml', got %q", config.OutputFile)
	}

	if !config.Overwrite {
		t.Error("Expected overwrite to be true")
	}
}

func TestInitCommandCreation(t *testing.T) {
	// Test that the command can be created and has required properties
	cmd := NewCommand()

	// Test that the command can be added to a parent command
	rootCmd := &cobra.Command{Use: "test"}
	rootCmd.AddCommand(cmd)

	// Verify the command was added
	if len(rootCmd.Commands()) != 1 {
		t.Error("Expected command to be added to root command")
	}

	if rootCmd.Commands()[0] != cmd {
		t.Error("Expected the added command to be the init command")
	}
}
