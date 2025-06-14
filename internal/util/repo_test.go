package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codcod/repos/internal/config"
	"github.com/fatih/color"
)

func TestGetRepoDir(t *testing.T) {
	tests := []struct {
		name     string
		repo     config.Repository
		expected string
	}{
		{
			name: "with custom path",
			repo: config.Repository{
				Name: "test-repo",
				URL:  "git@github.com:owner/test-repo.git",
				Path: "/custom/path/test-repo",
			},
			expected: "/custom/path/test-repo",
		},
		{
			name: "without custom path - SSH URL",
			repo: config.Repository{
				Name: "test-repo",
				URL:  "git@github.com:owner/test-repo.git",
			},
			expected: "test-repo",
		},
		{
			name: "without custom path - HTTPS URL",
			repo: config.Repository{
				Name: "my-project",
				URL:  "https://github.com/owner/my-project.git",
			},
			expected: "my-project",
		},
		{
			name: "without custom path - URL without .git suffix",
			repo: config.Repository{
				Name: "simple-repo",
				URL:  "git@github.com:owner/simple-repo",
			},
			expected: "simple-repo",
		},
		{
			name: "complex repository name",
			repo: config.Repository{
				Name: "complex-name",
				URL:  "git@github.com:org/my-complex-repo-name.git",
			},
			expected: "my-complex-repo-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRepoDir(tt.repo)
			if result != tt.expected {
				t.Errorf("GetRepoDir() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsGitRepository(t *testing.T) {
	// Create temporary directories for testing
	tmpDir := t.TempDir()

	// Create a mock git repository
	gitRepoDir := filepath.Join(tmpDir, "git-repo")
	err := os.MkdirAll(gitRepoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	gitDir := filepath.Join(gitRepoDir, ".git")
	err = os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create a non-git directory
	nonGitDir := filepath.Join(tmpDir, "non-git")
	err = os.MkdirAll(nonGitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create non-git directory: %v", err)
	}

	// Create a directory with .git file instead of directory
	gitFileDir := filepath.Join(tmpDir, "git-file")
	err = os.MkdirAll(gitFileDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create git-file directory: %v", err)
	}
	gitFile := filepath.Join(gitFileDir, ".git")
	err = os.WriteFile(gitFile, []byte("gitdir: /some/path"), 0644)
	if err != nil {
		t.Fatalf("Failed to create .git file: %v", err)
	}

	tests := []struct {
		name     string
		dir      string
		expected bool
	}{
		{
			name:     "valid git repository",
			dir:      gitRepoDir,
			expected: true,
		},
		{
			name:     "non-git directory",
			dir:      nonGitDir,
			expected: false,
		},
		{
			name:     "directory with .git file",
			dir:      gitFileDir,
			expected: false,
		},
		{
			name:     "non-existent directory",
			dir:      filepath.Join(tmpDir, "does-not-exist"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitRepository(tt.dir)
			if result != tt.expected {
				t.Errorf("IsGitRepository(%s) = %v, want %v", tt.dir, result, tt.expected)
			}
		})
	}
}

func TestExtractOwnerAndRepo(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantOwner   string
		wantRepo    string
		expectError bool
	}{
		{
			name:        "SSH URL with .git",
			url:         "git@github.com:owner/repo.git",
			wantOwner:   "owner",
			wantRepo:    "repo",
			expectError: false,
		},
		{
			name:        "SSH URL without .git",
			url:         "git@github.com:owner/repo",
			wantOwner:   "owner",
			wantRepo:    "repo",
			expectError: false,
		},
		{
			name:        "HTTPS URL with .git",
			url:         "https://github.com/owner/repo.git",
			wantOwner:   "owner",
			wantRepo:    "repo",
			expectError: false,
		},
		{
			name:        "HTTPS URL without .git",
			url:         "https://github.com/owner/repo",
			wantOwner:   "owner",
			wantRepo:    "repo",
			expectError: false,
		},
		{
			name:        "complex owner and repo names",
			url:         "git@github.com:my-org/my-complex-repo-name.git",
			wantOwner:   "my-org",
			wantRepo:    "my-complex-repo-name",
			expectError: false,
		},
		{
			name:        "invalid SSH URL - missing colon",
			url:         "git@github.com/owner/repo.git",
			wantOwner:   "",
			wantRepo:    "",
			expectError: true,
		},
		{
			name:        "invalid SSH URL - too many parts",
			url:         "git@github.com:owner/repo/extra.git",
			wantOwner:   "",
			wantRepo:    "",
			expectError: true,
		},
		{
			name:        "invalid HTTPS URL - missing github.com",
			url:         "https://gitlab.com/owner/repo.git",
			wantOwner:   "",
			wantRepo:    "",
			expectError: true,
		},
		{
			name:        "invalid HTTPS URL - too many parts",
			url:         "https://github.com/owner/repo/extra.git",
			wantOwner:   "",
			wantRepo:    "",
			expectError: true,
		},
		{
			name:        "unsupported protocol",
			url:         "http://github.com/owner/repo.git",
			wantOwner:   "",
			wantRepo:    "",
			expectError: true,
		},
		{
			name:        "empty URL",
			url:         "",
			wantOwner:   "",
			wantRepo:    "",
			expectError: true,
		},
		{
			name:        "malformed URL",
			url:         "not-a-url",
			wantOwner:   "",
			wantRepo:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOwner, gotRepo, err := ExtractOwnerAndRepo(tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("ExtractOwnerAndRepo() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractOwnerAndRepo() unexpected error: %v", err)
				return
			}

			if gotOwner != tt.wantOwner {
				t.Errorf("ExtractOwnerAndRepo() owner = %v, want %v", gotOwner, tt.wantOwner)
			}

			if gotRepo != tt.wantRepo {
				t.Errorf("ExtractOwnerAndRepo() repo = %v, want %v", gotRepo, tt.wantRepo)
			}
		})
	}
}

func TestColoredRepoName(t *testing.T) {
	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
	}

	// Test with different colors
	redColor := color.New(color.FgRed)
	blueColor := color.New(color.FgBlue)

	redResult := ColoredRepoName(repo, redColor)
	blueResult := ColoredRepoName(repo, blueColor)

	// Both should contain the repo name
	if !strings.Contains(redResult, repo.Name) {
		t.Errorf("ColoredRepoName() with red color should contain repo name, got: %s", redResult)
	}

	if !strings.Contains(blueResult, repo.Name) {
		t.Errorf("ColoredRepoName() with blue color should contain repo name, got: %s", blueResult)
	}

	// When colors are disabled, results should be the same
	// When colors are enabled, results should be different (different color codes)
	if !color.NoColor && redResult == blueResult {
		t.Error("ColoredRepoName() should produce different results for different colors when colors are enabled")
	}
}

func TestEnsureDirectoryExists(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "create single directory",
			path:        filepath.Join(tmpDir, "single"),
			expectError: false,
		},
		{
			name:        "create nested directories",
			path:        filepath.Join(tmpDir, "nested", "deep", "path"),
			expectError: false,
		},
		{
			name:        "directory already exists",
			path:        tmpDir, // This already exists
			expectError: false,
		},
		{
			name:        "create in existing directory",
			path:        filepath.Join(tmpDir, "existing", "new"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureDirectoryExists(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("EnsureDirectoryExists() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("EnsureDirectoryExists() unexpected error: %v", err)
				return
			}

			// Verify directory was created
			if _, err := os.Stat(tt.path); os.IsNotExist(err) {
				t.Errorf("EnsureDirectoryExists() directory was not created: %s", tt.path)
			}
		})
	}
}

func TestEnsureDirectoryExistsInvalidPath(t *testing.T) {
	// Test with invalid path (on Unix systems, null character is invalid)
	invalidPath := "invalid\x00path"
	err := EnsureDirectoryExists(invalidPath)

	// This should fail on most systems
	if err == nil {
		t.Log("Warning: EnsureDirectoryExists did not fail with invalid path - this may be platform specific")
	}
}

func BenchmarkGetRepoDir(b *testing.B) {
	repo := config.Repository{
		Name: "benchmark-repo",
		URL:  "git@github.com:owner/benchmark-repo.git",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetRepoDir(repo)
	}
}

func BenchmarkExtractOwnerAndRepo(b *testing.B) {
	url := "git@github.com:owner/repo.git"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ExtractOwnerAndRepo(url)
	}
}

func BenchmarkIsGitRepository(b *testing.B) {
	// Create a temporary git repository for benchmarking
	tmpDir := b.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		b.Fatalf("Failed to create test git directory: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsGitRepository(tmpDir)
	}
}
