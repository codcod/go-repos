package rm

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	// Test command properties
	if cmd.Use != "rm" {
		t.Errorf("Expected command use to be 'rm', got %q", cmd.Use)
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

func TestRmCommandFlags(t *testing.T) {
	cmd := NewCommand()
	flags := cmd.Flags()

	// Test required flags exist
	expectedFlags := []string{
		"parallel",
	}

	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %s to exist", flagName)
		}
	}

	// Test flag defaults
	parallelFlag := flags.Lookup("parallel")
	if parallelFlag != nil && parallelFlag.DefValue != "false" {
		t.Errorf("Expected default parallel to be 'false', got %q", parallelFlag.DefValue)
	}
}

func TestConfig(t *testing.T) {
	config := &Config{
		Tag:      "test",
		Parallel: true,
	}

	if config.Tag != "test" {
		t.Errorf("Expected tag to be 'test', got %q", config.Tag)
	}

	if !config.Parallel {
		t.Error("Expected parallel to be true")
	}
}

func TestRmCommandCreation(t *testing.T) {
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
		t.Error("Expected the added command to be the rm command")
	}
}
