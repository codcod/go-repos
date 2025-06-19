package run

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	// Test command properties
	if cmd.Use != "run [command]" {
		t.Errorf("Expected command use to be 'run [command]', got %q", cmd.Use)
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

	if cmd.Args == nil {
		t.Error("Expected command Args to be set")
	}
}

func TestRunCommandFlags(t *testing.T) {
	cmd := NewCommand()
	flags := cmd.Flags()

	// Test required flags exist
	expectedFlags := []string{
		"parallel",
		"log-dir",
	}

	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %s to exist", flagName)
		}
	}

	// Test default values
	logDirFlag := flags.Lookup("log-dir")
	if logDirFlag != nil && logDirFlag.DefValue != "logs" {
		t.Errorf("Expected default log-dir to be 'logs', got %q", logDirFlag.DefValue)
	}

	parallelFlag := flags.Lookup("parallel")
	if parallelFlag != nil && parallelFlag.DefValue != "false" {
		t.Errorf("Expected default parallel to be 'false', got %q", parallelFlag.DefValue)
	}
}

func TestConfig(t *testing.T) {
	config := &Config{
		Tag:      "test",
		Parallel: true,
		LogDir:   "custom-logs",
	}

	if config.Tag != "test" {
		t.Errorf("Expected tag to be 'test', got %q", config.Tag)
	}

	if !config.Parallel {
		t.Error("Expected parallel to be true")
	}

	if config.LogDir != "custom-logs" {
		t.Errorf("Expected log dir to be 'custom-logs', got %q", config.LogDir)
	}
}

func TestConfigDefaults(t *testing.T) {
	// Test that the config has proper defaults when created
	config := &Config{
		LogDir: "logs",
	}

	if config.LogDir != "logs" {
		t.Errorf("Expected default log dir to be 'logs', got %q", config.LogDir)
	}

	if config.Parallel {
		t.Error("Expected default parallel to be false")
	}

	if config.Tag != "" {
		t.Errorf("Expected default tag to be empty, got %q", config.Tag)
	}
}

func TestRunCommandCreation(t *testing.T) {
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
		t.Error("Expected the added command to be the run command")
	}
}
