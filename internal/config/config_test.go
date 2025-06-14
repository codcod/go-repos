package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// writeTempConfig writes the given YAML to a temp file and returns its path.
func writeTempConfig(t *testing.T, configYAML string) string {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	return configPath
}

// Test data constants for consistency
const (
	validSingleRepoConfig = `repositories:
  - name: test-repo
    url: git@github.com:owner/test-repo.git
    tags: [go, backend]
    branch: main
    path: /tmp/test-repo`

	validMultiRepoConfig = `repositories:
  - name: repo1
    url: git@github.com:owner/repo1.git
    tags: [go, backend]
  - name: repo2
    url: https://github.com/owner/repo2.git
    tags: [javascript, frontend]
    branch: develop`

	invalidYAMLConfig = `repositories:
  - name: test-repo
    url: git@github.com:owner/test-repo.git
    tags: [go, backend
    branch: main`

	emptyReposConfig = `repositories: []`

	missingReposKeyConfig = `some_other_key:
  - name: test-repo`
)

// TestLoadConfig verifies loading various config YAML scenarios.
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		expectError bool
		wantRepos   int
	}{
		{
			name:        "valid config with single repository",
			configYAML:  validSingleRepoConfig,
			expectError: false,
			wantRepos:   1,
		},
		{
			name:        "valid config with multiple repositories",
			configYAML:  validMultiRepoConfig,
			expectError: false,
			wantRepos:   2,
		},
		{
			name:        "empty repositories",
			configYAML:  emptyReposConfig,
			expectError: false,
			wantRepos:   0,
		},
		{
			name:        "invalid YAML",
			configYAML:  invalidYAMLConfig,
			expectError: true,
			wantRepos:   0,
		},
		{
			name:        "missing repositories key",
			configYAML:  missingReposKeyConfig,
			expectError: false,
			wantRepos:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := writeTempConfig(t, tt.configYAML)
			config, err := LoadConfig(configPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(config.Repositories) != tt.wantRepos {
				t.Errorf("Expected %d repositories, got %d", tt.wantRepos, len(config.Repositories))
			}

			// Additional validation for non-empty configs
			if tt.wantRepos > 0 {
				validateFirstRepository(t, config.Repositories[0])
			}
		})
	}
}

// TestLoadConfigFileNotFound checks error on missing config file.
func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("non-existent-file.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestFilterRepositoriesByTag tests filtering repositories by tag.
func TestFilterRepositoriesByTag(t *testing.T) {
	config := createTestConfigWithRepos()

	tests := []struct {
		name     string
		tag      string
		expected []string // repository names
	}{
		{
			name:     "filter by backend tag",
			tag:      "backend",
			expected: []string{"go-app", "python-api"},
		},
		{
			name:     "filter by frontend tag",
			tag:      "frontend",
			expected: []string{"react-ui"},
		},
		{
			name:     "filter by go tag",
			tag:      "go",
			expected: []string{"go-app"},
		},
		{
			name:     "filter by non-existent tag",
			tag:      "java",
			expected: []string{},
		},
		{
			name:     "empty tag returns all",
			tag:      "",
			expected: []string{"go-app", "react-ui", "python-api", "docs"},
		},
		{
			name:     "case sensitive filtering",
			tag:      "Backend",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := config.FilterRepositoriesByTag(tt.tag)
			validateFilteredResults(t, filtered, tt.expected)
		})
	}
}

// createTestConfigWithRepos creates a config with test repositories.
func createTestConfigWithRepos() *Config {
	return &Config{
		Repositories: []Repository{
			{
				Name: "go-app",
				URL:  "git@github.com:owner/go-app.git",
				Tags: []string{"go", "backend", "microservice"},
			},
			{
				Name: "react-ui",
				URL:  "git@github.com:owner/react-ui.git",
				Tags: []string{"javascript", "frontend", "react"},
			},
			{
				Name: "python-api",
				URL:  "git@github.com:owner/python-api.git",
				Tags: []string{"python", "backend", "api"},
			},
			{
				Name: "docs",
				URL:  "git@github.com:owner/docs.git",
				Tags: []string{"documentation"},
			},
		},
	}
}

// validateFilteredResults validates filtered repository results.
func validateFilteredResults(t *testing.T, filtered []Repository, expected []string) {
	t.Helper()
	if len(filtered) != len(expected) {
		t.Errorf("Expected %d repositories, got %d", len(expected), len(filtered))
		return
	}
	for i, expectedName := range expected {
		if filtered[i].Name != expectedName {
			t.Errorf("Expected repository %s at index %d, got %s", expectedName, i, filtered[i].Name)
		}
	}
}

// TestRepositoryFields checks parsing of all repository fields.
func TestRepositoryFields(t *testing.T) {
	configYAML := `repositories:
  - name: full-config-repo
    url: git@github.com:owner/repo.git
    tags: [go, backend]
    branch: develop
    path: /custom/path
  - name: minimal-config-repo
    url: https://github.com/owner/minimal.git
    tags: [frontend]`

	configPath := writeTempConfig(t, configYAML)
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(config.Repositories) != 2 {
		t.Fatalf("Expected 2 repositories, got %d", len(config.Repositories))
	}

	validateFullConfigRepo(t, config.Repositories[0])
	validateMinimalConfigRepo(t, config.Repositories[1])
}

// validateFullConfigRepo validates a repository with all fields set.
func validateFullConfigRepo(t *testing.T, repo Repository) {
	t.Helper()
	expected := Repository{
		Name:   "full-config-repo",
		URL:    "git@github.com:owner/repo.git",
		Branch: "develop",
		Path:   "/custom/path",
		Tags:   []string{"go", "backend"},
	}

	if repo.Name != expected.Name {
		t.Errorf("Expected name '%s', got '%s'", expected.Name, repo.Name)
	}
	if repo.URL != expected.URL {
		t.Errorf("Expected URL '%s', got '%s'", expected.URL, repo.URL)
	}
	if repo.Branch != expected.Branch {
		t.Errorf("Expected branch '%s', got '%s'", expected.Branch, repo.Branch)
	}
	if repo.Path != expected.Path {
		t.Errorf("Expected path '%s', got '%s'", expected.Path, repo.Path)
	}
	if len(repo.Tags) != len(expected.Tags) {
		t.Errorf("Expected tags %v, got %v", expected.Tags, repo.Tags)
	}
}

// validateMinimalConfigRepo validates a repository with minimal fields.
func validateMinimalConfigRepo(t *testing.T, repo Repository) {
	t.Helper()
	if repo.Name != "minimal-config-repo" {
		t.Errorf("Expected name 'minimal-config-repo', got '%s'", repo.Name)
	}
	if repo.URL != "https://github.com/owner/minimal.git" {
		t.Errorf("Expected URL 'https://github.com/owner/minimal.git', got '%s'", repo.URL)
	}
	if repo.Branch != "" {
		t.Errorf("Expected empty branch, got '%s'", repo.Branch)
	}
	if repo.Path != "" {
		t.Errorf("Expected empty path, got '%s'", repo.Path)
	}
	if len(repo.Tags) != 1 || repo.Tags[0] != "frontend" {
		t.Errorf("Expected tags [frontend], got %v", repo.Tags)
	}
}

// TestFilterRepositoriesByTagEdgeCases checks filtering with empty/nil tags.
func TestFilterRepositoriesByTagEdgeCases(t *testing.T) {
	config := &Config{
		Repositories: []Repository{
			{Name: "no-tags-repo", URL: "git@github.com:owner/no-tags.git", Tags: []string{}},
			{Name: "nil-tags-repo", URL: "git@github.com:owner/nil-tags.git", Tags: nil},
		},
	}

	filtered := config.FilterRepositoriesByTag("any-tag")
	if len(filtered) != 0 {
		t.Errorf("Expected 0 repositories when filtering repos with no tags, got %d", len(filtered))
	}

	filtered = config.FilterRepositoriesByTag("")
	if len(filtered) != 2 {
		t.Errorf("Expected 2 repositories when filtering with empty tag, got %d", len(filtered))
	}
}

// BenchmarkFilterRepositoriesByTag benchmarks tag filtering on a large config.
func BenchmarkFilterRepositoriesByTag(b *testing.B) {
	config := makeLargeConfigForBenchmark()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filtered := config.FilterRepositoriesByTag("backend")
		if len(filtered) != 100 {
			b.Errorf("Expected 100 repositories, got %d", len(filtered))
		}
	}
}

// makeLargeConfigForBenchmark creates a large config for benchmarking.
func makeLargeConfigForBenchmark() *Config {
	config := &Config{
		Repositories: make([]Repository, 1000),
	}
	for i := 0; i < 1000; i++ {
		config.Repositories[i] = Repository{
			Name: fmt.Sprintf("repo-%d", i),
			URL:  fmt.Sprintf("git@github.com:owner/repo-%d.git", i),
			Tags: []string{"tag1", "tag2", "tag3"},
		}
	}
	for i := 0; i < 100; i++ {
		config.Repositories[i].Tags = append(config.Repositories[i].Tags, "backend")
	}
	return config
}

// validateFirstRepository validates that a repository has required fields.
func validateFirstRepository(t *testing.T, repo Repository) {
	t.Helper()
	if repo.Name == "" {
		t.Error("Repository name should not be empty")
	}
	if repo.URL == "" {
		t.Error("Repository URL should not be empty")
	}
	if len(repo.Tags) == 0 {
		t.Error("Repository should have at least one tag")
	}
}
