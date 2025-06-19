package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_VAR_EXISTS",
			defaultValue: "default",
			envValue:     "env_value",
			setEnv:       true,
			expected:     "env_value",
		},
		{
			name:         "environment variable does not exist",
			key:          "TEST_VAR_NOT_EXISTS",
			defaultValue: "default_value",
			envValue:     "",
			setEnv:       false,
			expected:     "default_value",
		},
		{
			name:         "environment variable is empty string",
			key:          "TEST_VAR_EMPTY",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			expected:     "default",
		},
		{
			name:         "both env and default are empty",
			key:          "TEST_VAR_BOTH_EMPTY",
			defaultValue: "",
			envValue:     "",
			setEnv:       true,
			expected:     "",
		},
		{
			name:         "environment variable with spaces",
			key:          "TEST_VAR_SPACES",
			defaultValue: "default",
			envValue:     "  value with spaces  ",
			setEnv:       true,
			expected:     "  value with spaces  ",
		},
		{
			name:         "environment variable with special characters",
			key:          "TEST_VAR_SPECIAL",
			defaultValue: "default",
			envValue:     "value!@#$%^&*()",
			setEnv:       true,
			expected:     "value!@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variable before and after test
			originalValue := os.Getenv(tt.key)
			defer func() {
				if originalValue != "" {
					_ = os.Setenv(tt.key, originalValue)
				} else {
					_ = os.Unsetenv(tt.key)
				}
			}()

			// Set up test environment
			if tt.setEnv {
				_ = os.Setenv(tt.key, tt.envValue)
			} else {
				_ = os.Unsetenv(tt.key)
			}

			// Test the function
			result := getEnvOrDefault(tt.key, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("getEnvOrDefault(%q, %q) = %q, want %q",
					tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestVersionVariables(t *testing.T) {
	// Test that version variables are initialized
	if version == "" {
		t.Error("version should not be empty")
	}
	if commit == "" {
		t.Error("commit should not be empty")
	}
	if date == "" {
		t.Error("date should not be empty")
	}

	// Test default values when no environment variables are set
	oldVersion := os.Getenv("VERSION")
	oldCommit := os.Getenv("COMMIT")
	oldDate := os.Getenv("BUILD_DATE")

	_ = os.Unsetenv("VERSION")
	_ = os.Unsetenv("COMMIT")
	_ = os.Unsetenv("BUILD_DATE")

	// Reinitialize to test defaults
	testVersion := getEnvOrDefault("VERSION", "dev")
	testCommit := getEnvOrDefault("COMMIT", "unknown")
	testDate := getEnvOrDefault("BUILD_DATE", "unknown")

	if testVersion != "dev" {
		t.Errorf("Expected version to default to 'dev', got %q", testVersion)
	}
	if testCommit != "unknown" {
		t.Errorf("Expected commit to default to 'unknown', got %q", testCommit)
	}
	if testDate != "unknown" {
		t.Errorf("Expected date to default to 'unknown', got %q", testDate)
	}

	// Restore original environment variables
	if oldVersion != "" {
		_ = os.Setenv("VERSION", oldVersion)
	}
	if oldCommit != "" {
		_ = os.Setenv("COMMIT", oldCommit)
	}
	if oldDate != "" {
		_ = os.Setenv("BUILD_DATE", oldDate)
	}
}

func TestVersionVariablesWithEnv(t *testing.T) {
	// Test with environment variables set
	testCases := []struct {
		envVar   string
		envValue string
		getter   func() string
	}{
		{
			envVar:   "VERSION",
			envValue: "1.2.3",
			getter:   func() string { return getEnvOrDefault("VERSION", "dev") },
		},
		{
			envVar:   "COMMIT",
			envValue: "abc123",
			getter:   func() string { return getEnvOrDefault("COMMIT", "unknown") },
		},
		{
			envVar:   "BUILD_DATE",
			envValue: "2024-12-19",
			getter:   func() string { return getEnvOrDefault("BUILD_DATE", "unknown") },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.envVar, func(t *testing.T) {
			// Save original value
			original := os.Getenv(tc.envVar)
			defer func() {
				if original != "" {
					_ = os.Setenv(tc.envVar, original)
				} else {
					_ = os.Unsetenv(tc.envVar)
				}
			}()

			// Set test value
			_ = os.Setenv(tc.envVar, tc.envValue)

			// Test that environment variable is used
			result := tc.getter()
			if result != tc.envValue {
				t.Errorf("Expected %s to be %q when env var is set, got %q",
					tc.envVar, tc.envValue, result)
			}
		})
	}
}

func TestDefaultLogDir(t *testing.T) {
	expectedDefault := "logs"
	if defaultLogs != expectedDefault {
		t.Errorf("Expected defaultLogs to be %q, got %q", expectedDefault, defaultLogs)
	}
}

func TestGlobalVariablesInitialization(t *testing.T) {
	// Test that global variables are properly initialized to reasonable defaults
	if configFile == "" {
		t.Error("configFile should have a default value")
	}

	// Check that boolean flags have proper zero values
	if parallel != false {
		t.Error("parallel should default to false")
	}
	if prDraft != false {
		t.Error("prDraft should default to false")
	}
	if createOnly != false {
		t.Error("createOnly should default to false")
	}
	if overwrite != false {
		t.Error("overwrite should default to false")
	}

	// Check that string flags are initialized (empty is OK)
	_ = tag        // tag can be empty
	_ = logDir     // logDir can be empty
	_ = prTitle    // prTitle can be empty
	_ = prBody     // prBody can be empty
	_ = prBranch   // prBranch can be empty
	_ = baseBranch // baseBranch can be empty
	_ = commitMsg  // commitMsg can be empty
	_ = prToken    // prToken can be empty
	_ = outputFile // outputFile can be empty
}

func TestEnvironmentVariableEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envValue string
		expected string
	}{
		{
			name:     "very long environment variable",
			key:      "TEST_LONG_VAR",
			envValue: "very_long_" + strings.Repeat("x", 1000) + "_value",
			expected: "very_long_" + strings.Repeat("x", 1000) + "_value",
		},
		{
			name:     "environment variable with newlines",
			key:      "TEST_NEWLINE_VAR",
			envValue: "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "environment variable with unicode",
			key:      "TEST_UNICODE_VAR",
			envValue: "测试🚀🎉",
			expected: "测试🚀🎉",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := os.Getenv(tt.key)
			defer func() {
				if original != "" {
					_ = os.Setenv(tt.key, original)
				} else {
					_ = os.Unsetenv(tt.key)
				}
			}()

			// Set test value
			_ = os.Setenv(tt.key, tt.envValue)

			// Test
			result := getEnvOrDefault(tt.key, "default")
			if result != tt.expected {
				t.Errorf("getEnvOrDefault with %s failed: got %q, want %q",
					tt.name, result, tt.expected)
			}
		})
	}
}

func BenchmarkGetEnvOrDefault(b *testing.B) {
	key := "BENCHMARK_TEST_VAR"
	defaultValue := "default_value"

	// Test with environment variable set
	_ = os.Setenv(key, "env_value")
	defer func() { _ = os.Unsetenv(key) }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getEnvOrDefault(key, defaultValue)
	}
}

func BenchmarkGetEnvOrDefaultNotSet(b *testing.B) {
	key := "BENCHMARK_TEST_VAR_NOT_SET"
	defaultValue := "default_value"

	// Ensure environment variable is not set
	_ = os.Unsetenv(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getEnvOrDefault(key, defaultValue)
	}
}

func TestHealthCommandWithListCategories(t *testing.T) {
	// Test that the command can be executed with --list-categories flag
	// This integration test covers the functionality that was previously tested
	// by the direct function call (which has been moved to the health command module)
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "repos_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("repos_test")

	// Run the command with --list-categories
	cmd := exec.Command("./repos_test", "health", "--list-categories")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	outputStr := string(output)

	// Verify expected content
	expectedPhrases := []string{
		"Available Health Check Categories",
		"CHECKERS:",
		"ANALYZERS:",
		"Summary",
		"Total Checkers:",
		"Total Categories:",
		"Total Analyzers:",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(outputStr, phrase) {
			t.Errorf("Expected phrase '%s' not found in output", phrase)
		}
	}
}
