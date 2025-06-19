package common

import (
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/codcod/repos/internal/config"
)

func TestCLIError(t *testing.T) {
	err := &CLIError{
		Message:  "test error",
		ExitCode: 42,
	}

	if err.Error() != "test error" {
		t.Errorf("Expected error message 'test error', got %q", err.Error())
	}
}

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
			key:          "TEST_COMMON_VAR_EXISTS",
			defaultValue: "default",
			envValue:     "env_value",
			setEnv:       true,
			expected:     "env_value",
		},
		{
			name:         "environment variable does not exist",
			key:          "TEST_COMMON_VAR_NOT_EXISTS",
			defaultValue: "default_value",
			envValue:     "",
			setEnv:       false,
			expected:     "default_value",
		},
		{
			name:         "environment variable is empty string",
			key:          "TEST_COMMON_VAR_EMPTY",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			expected:     "default",
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
			result := GetEnvOrDefault(tt.key, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("GetEnvOrDefault(%q, %q) = %q, want %q",
					tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Test with valid config file
	tmpDir := t.TempDir()
	configFile := tmpDir + "/test_config.yaml"

	validConfig := `repositories:
  - name: test-repo
    url: https://github.com/test/repo.git
    path: /tmp/test-repo
    tags: [test]
`

	err := os.WriteFile(configFile, []byte(validConfig), 0600)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadConfig returned nil config")
	}

	if len(cfg.Repositories) != 1 {
		t.Errorf("Expected 1 repository, got %d", len(cfg.Repositories))
	}

	if cfg.Repositories[0].Name != "test-repo" {
		t.Errorf("Expected repository name 'test-repo', got %q", cfg.Repositories[0].Name)
	}
}

func TestLoadConfigInvalidFile(t *testing.T) {
	// Test with non-existent config file
	_, err := LoadConfig("/path/that/does/not/exist.yaml")
	if err == nil {
		t.Error("LoadConfig should return error for non-existent file")
	}
}

func TestProcessReposSequential(t *testing.T) {
	repos := []config.Repository{
		{Name: "repo1", URL: "https://github.com/test/repo1.git"},
		{Name: "repo2", URL: "https://github.com/test/repo2.git"},
	}

	processedCount := 0
	processor := func(repo config.Repository) error {
		processedCount++
		return nil
	}

	err := ProcessRepos(repos, false, processor)
	if err != nil {
		t.Errorf("ProcessRepos failed: %v", err)
	}

	if processedCount != 2 {
		t.Errorf("Expected 2 repos to be processed, got %d", processedCount)
	}
}

func TestProcessReposWithError(t *testing.T) {
	repos := []config.Repository{
		{Name: "repo1", URL: "https://github.com/test/repo1.git"},
		{Name: "repo2", URL: "https://github.com/test/repo2.git"},
	}

	processor := func(repo config.Repository) error {
		if repo.Name == "repo2" {
			return &CLIError{Message: "test error", ExitCode: 1}
		}
		return nil
	}

	err := ProcessRepos(repos, false, processor)
	if err == nil {
		t.Error("ProcessRepos should return error when processor fails")
	}

	if !strings.Contains(err.Error(), "repo2") {
		t.Errorf("Error should contain repository name, got: %v", err)
	}
}

func TestProcessReposParallel(t *testing.T) {
	repos := []config.Repository{
		{Name: "repo1", URL: "https://github.com/test/repo1.git"},
		{Name: "repo2", URL: "https://github.com/test/repo2.git"},
	}

	var processedCount int32
	processor := func(repo config.Repository) error {
		atomic.AddInt32(&processedCount, 1)
		return nil
	}

	err := ProcessRepos(repos, true, processor)
	if err != nil {
		t.Errorf("ProcessRepos parallel failed: %v", err)
	}

	if atomic.LoadInt32(&processedCount) != 2 {
		t.Errorf("Expected 2 repos to be processed, got %d", atomic.LoadInt32(&processedCount))
	}
}

func BenchmarkGetEnvOrDefault(b *testing.B) {
	key := "BENCHMARK_COMMON_TEST_VAR"
	defaultValue := "default_value"

	// Test with environment variable set
	_ = os.Setenv(key, "env_value")
	defer func() { _ = os.Unsetenv(key) }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetEnvOrDefault(key, defaultValue)
	}
}

func BenchmarkProcessRepos(b *testing.B) {
	repos := []config.Repository{
		{Name: "repo1", URL: "https://github.com/test/repo1.git"},
		{Name: "repo2", URL: "https://github.com/test/repo2.git"},
		{Name: "repo3", URL: "https://github.com/test/repo3.git"},
	}

	processor := func(repo config.Repository) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ProcessRepos(repos, false, processor)
	}
}
