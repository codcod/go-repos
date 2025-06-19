package health

import (
	"testing"
)

func TestNewDryRunCommand(t *testing.T) {
	cmd := NewDryRunCommand()

	// Test command properties
	if cmd.Use != "dryrun" {
		t.Errorf("Expected command use to be 'dryrun', got %q", cmd.Use)
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
