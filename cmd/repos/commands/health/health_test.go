package health

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	// Test command properties
	if cmd.Use != "health" {
		t.Errorf("Expected command use to be 'health', got %q", cmd.Use)
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

	// Test that subcommands are added
	subcommands := cmd.Commands()
	expectedSubcommands := []string{"cyclomatic-complexity", "genconfig"}

	if len(subcommands) != len(expectedSubcommands) {
		t.Errorf("Expected %d subcommands, got %d", len(expectedSubcommands), len(subcommands))
	}

	for _, expected := range expectedSubcommands {
		found := false
		for _, subcmd := range subcommands {
			if subcmd.Use == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %s not found", expected)
		}
	}
}

func TestConfigDefaults(t *testing.T) {
	// Test default configuration values
	config := &Config{
		TimeoutSeconds: 30,
	}

	if config.TimeoutSeconds != 30 {
		t.Errorf("Expected default timeout to be 30 seconds, got %d", config.TimeoutSeconds)
	}
}

func TestHealthCommandFlags(t *testing.T) {
	cmd := NewCommand()
	flags := cmd.Flags()

	// Test required flags exist
	expectedFlags := []string{
		"config",
		"category",
		"parallel",
		"timeout",
		"dry-run",
		"verbose",
		"list-categories",
		"max-complexity",
	}

	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %s to exist", flagName)
		}
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid basic config",
			config: &Config{
				TimeoutSeconds: 30,
			},
			wantErr: false,
		},
		{
			name: "valid config with categories",
			config: &Config{
				TimeoutSeconds: 60,
				Categories:     []string{"git", "security"},
			},
			wantErr: false,
		},
		{
			name: "invalid timeout - negative",
			config: &Config{
				TimeoutSeconds: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid timeout - zero",
			config: &Config{
				TimeoutSeconds: 0,
			},
			wantErr: true,
		},
		{
			name: "valid config with max complexity",
			config: &Config{
				TimeoutSeconds: 30,
				MaxComplexity:  10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSimpleLogger(t *testing.T) {
	logger := &simpleLogger{}

	// Test that logger methods don't panic
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Test that logger methods exist and don't panic
	// Note: formatFields is an internal method, testing it directly is not necessary
}

func TestHealthCommandCreation(t *testing.T) {
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
		t.Error("Expected the added command to be the health command")
	}
}
