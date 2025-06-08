package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		expectError bool
		wantRepos   int
	}{
		{
			name: "valid config with single repository",
			configYAML: `repositories:
  - name: test-repo
    url: git@github.com:owner/test-repo.git
    tags: [go, backend]
    branch: main
    path: /tmp/test-repo`,
			expectError: false,
			wantRepos:   1,
		},
		{
			name: "valid config with multiple repositories",
			configYAML: `repositories:
  - name: repo1
    url: git@github.com:owner/repo1.git
    tags: [go, backend]
  - name: repo2
    url: https://github.com/owner/repo2.git
    tags: [javascript, frontend]
    branch: develop`,
			expectError: false,
			wantRepos:   2,
		},
		{
			name: "empty repositories",
			configYAML: `repositories: []`,
			expectError: false,
			wantRepos:   0,
		},
		{
			name: "invalid YAML",
			configYAML: `repositories:
  - name: test-repo
    url: git@github.com:owner/test-repo.git
    tags: [go, backend
    branch: main`,
			expectError: true,
			wantRepos:   0,
		},
		{
			name: "missing repositories key",
			configYAML: `some_other_key:
  - name: test-repo`,
			expectError: false,
			wantRepos:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			
			err := os.WriteFile(configPath, []byte(tt.configYAML), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			// Test LoadConfig
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
				repo := config.Repositories[0]
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
		})
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("non-existent-file.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestFilterRepositoriesByTag(t *testing.T) {
	config := &Config{
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

			if len(filtered) != len(tt.expected) {
				t.Errorf("Expected %d repositories, got %d", len(tt.expected), len(filtered))
				return
			}

			// Check that we got the expected repositories
			for i, expectedName := range tt.expected {
				if filtered[i].Name != expectedName {
					t.Errorf("Expected repository %s at index %d, got %s", expectedName, i, filtered[i].Name)
				}
			}
		})
	}
}

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

	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	err := os.WriteFile(configPath, []byte(configYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(config.Repositories) != 2 {
		t.Fatalf("Expected 2 repositories, got %d", len(config.Repositories))
	}

	// Test full config repository
	fullRepo := config.Repositories[0]
	if fullRepo.Name != "full-config-repo" {
		t.Errorf("Expected name 'full-config-repo', got '%s'", fullRepo.Name)
	}
	if fullRepo.URL != "git@github.com:owner/repo.git" {
		t.Errorf("Expected URL 'git@github.com:owner/repo.git', got '%s'", fullRepo.URL)
	}
	if fullRepo.Branch != "develop" {
		t.Errorf("Expected branch 'develop', got '%s'", fullRepo.Branch)
	}
	if fullRepo.Path != "/custom/path" {
		t.Errorf("Expected path '/custom/path', got '%s'", fullRepo.Path)
	}
	if len(fullRepo.Tags) != 2 || fullRepo.Tags[0] != "go" || fullRepo.Tags[1] != "backend" {
		t.Errorf("Expected tags [go, backend], got %v", fullRepo.Tags)
	}

	// Test minimal config repository
	minimalRepo := config.Repositories[1]
	if minimalRepo.Name != "minimal-config-repo" {
		t.Errorf("Expected name 'minimal-config-repo', got '%s'", minimalRepo.Name)
	}
	if minimalRepo.URL != "https://github.com/owner/minimal.git" {
		t.Errorf("Expected URL 'https://github.com/owner/minimal.git', got '%s'", minimalRepo.URL)
	}
	if minimalRepo.Branch != "" {
		t.Errorf("Expected empty branch, got '%s'", minimalRepo.Branch)
	}
	if minimalRepo.Path != "" {
		t.Errorf("Expected empty path, got '%s'", minimalRepo.Path)
	}
	if len(minimalRepo.Tags) != 1 || minimalRepo.Tags[0] != "frontend" {
		t.Errorf("Expected tags [frontend], got %v", minimalRepo.Tags)
	}
}

func TestFilterRepositoriesByTagEdgeCases(t *testing.T) {
	config := &Config{
		Repositories: []Repository{
			{
				Name: "no-tags-repo",
				URL:  "git@github.com:owner/no-tags.git",
				Tags: []string{},
			},
			{
				Name: "nil-tags-repo",
				URL:  "git@github.com:owner/nil-tags.git",
				Tags: nil,
			},
		},
	}

	// Test filtering when repositories have no tags
	filtered := config.FilterRepositoriesByTag("any-tag")
	if len(filtered) != 0 {
		t.Errorf("Expected 0 repositories when filtering repos with no tags, got %d", len(filtered))
	}

	// Test filtering with empty tag (should return all)
	filtered = config.FilterRepositoriesByTag("")
	if len(filtered) != 2 {
		t.Errorf("Expected 2 repositories when filtering with empty tag, got %d", len(filtered))
	}
}

func BenchmarkFilterRepositoriesByTag(b *testing.B) {
	// Create a large config for benchmarking
	config := &Config{
		Repositories: make([]Repository, 1000),
	}

	// Fill with test data
	for i := 0; i < 1000; i++ {
		config.Repositories[i] = Repository{
			Name: fmt.Sprintf("repo-%d", i),
			URL:  fmt.Sprintf("git@github.com:owner/repo-%d.git", i),
			Tags: []string{"tag1", "tag2", "tag3"},
		}
	}

	// Add some repositories with the target tag
	for i := 0; i < 100; i++ {
		config.Repositories[i].Tags = append(config.Repositories[i].Tags, "backend")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		filtered := config.FilterRepositoriesByTag("backend")
		if len(filtered) != 100 {
			b.Errorf("Expected 100 repositories, got %d", len(filtered))
		}
	}
}