package pr

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	// Test command properties
	if cmd.Use != "pr" {
		t.Errorf("Expected command use to be 'pr', got %q", cmd.Use)
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

func TestPRCommandFlags(t *testing.T) {
	cmd := NewCommand()
	flags := cmd.Flags()

	// Test required flags exist
	expectedFlags := []string{
		"parallel",
		"title",
		"body",
		"branch",
		"base",
		"commit",
		"draft",
		"token",
		"create-only",
	}

	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %s to exist", flagName)
		}
	}

	// Test some default values
	baseBranchFlag := flags.Lookup("base")
	if baseBranchFlag != nil && baseBranchFlag.DefValue != "main" {
		t.Errorf("Expected default base to be 'main', got %q", baseBranchFlag.DefValue)
	}

	draftFlag := flags.Lookup("draft")
	if draftFlag != nil && draftFlag.DefValue != "false" {
		t.Errorf("Expected default draft to be 'false', got %q", draftFlag.DefValue)
	}
}

func TestConfig(t *testing.T) {
	config := &Config{
		Tag:        "test",
		Parallel:   true,
		Title:      "Test PR",
		Body:       "Test body",
		Branch:     "feature/test",
		BaseBranch: "main",
		CommitMsg:  "Test commit",
		Draft:      false,
		Token:      "test-token",
		CreateOnly: true,
	}

	if config.Tag != "test" {
		t.Errorf("Expected tag to be 'test', got %q", config.Tag)
	}

	if !config.Parallel {
		t.Error("Expected parallel to be true")
	}

	if config.Title != "Test PR" {
		t.Errorf("Expected title to be 'Test PR', got %q", config.Title)
	}

	if config.Body != "Test body" {
		t.Errorf("Expected body to be 'Test body', got %q", config.Body)
	}

	if config.Branch != "feature/test" {
		t.Errorf("Expected branch to be 'feature/test', got %q", config.Branch)
	}

	if config.BaseBranch != "main" {
		t.Errorf("Expected base branch to be 'main', got %q", config.BaseBranch)
	}

	if config.CommitMsg != "Test commit" {
		t.Errorf("Expected commit msg to be 'Test commit', got %q", config.CommitMsg)
	}

	if config.Draft {
		t.Error("Expected draft to be false")
	}

	if config.Token != "test-token" {
		t.Errorf("Expected token to be 'test-token', got %q", config.Token)
	}

	if !config.CreateOnly {
		t.Error("Expected create only to be true")
	}
}

func TestPRCommandCreation(t *testing.T) {
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
		t.Error("Expected the added command to be the pr command")
	}
}
