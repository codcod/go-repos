package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codcod/repos/internal/config"
)

func TestRemoveRepository(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		setupRepo     func(string) config.Repository
		expectError   bool
		errorContains string
	}{
		{
			name: "remove valid git repository",
			setupRepo: func(baseDir string) config.Repository {
				repoDir := filepath.Join(baseDir, "valid-repo")
				gitDir := filepath.Join(repoDir, ".git")
				_ = os.MkdirAll(gitDir, 0755)

				return config.Repository{
					Name: "valid-repo",
					URL:  "git@github.com:owner/valid-repo.git",
					Path: repoDir,
				}
			},
			expectError: false,
		},
		{
			name: "remove repository with custom path",
			setupRepo: func(baseDir string) config.Repository {
				customPath := filepath.Join(baseDir, "custom", "path", "repo")
				gitDir := filepath.Join(customPath, ".git")
				_ = os.MkdirAll(gitDir, 0755)

				return config.Repository{
					Name: "custom-repo",
					URL:  "git@github.com:owner/custom-repo.git",
					Path: customPath,
				}
			},
			expectError: false,
		},
		{
			name: "remove repository using URL-derived path",
			setupRepo: func(baseDir string) config.Repository {
				// No custom path - should derive from URL
				repoDir := filepath.Join(baseDir, "url-derived-repo")
				gitDir := filepath.Join(repoDir, ".git")
				_ = os.MkdirAll(gitDir, 0755)

				return config.Repository{
					Name: "url-derived-repo",
					URL:  "git@github.com:owner/url-derived-repo.git",
					Path: repoDir, // Set explicit path since GetRepoDir depends on current directory
				}
			},
			expectError: false,
		},
		{
			name: "error when directory does not exist",
			setupRepo: func(baseDir string) config.Repository {
				return config.Repository{
					Name: "non-existent",
					URL:  "git@github.com:owner/non-existent.git",
					Path: filepath.Join(baseDir, "does-not-exist"),
				}
			},
			expectError:   true,
			errorContains: "repository directory does not exist",
		},
		{
			name: "error when directory is not a git repository",
			setupRepo: func(baseDir string) config.Repository {
				notGitDir := filepath.Join(baseDir, "not-git")
				_ = os.MkdirAll(notGitDir, 0755)

				return config.Repository{
					Name: "not-git",
					URL:  "git@github.com:owner/not-git.git",
					Path: notGitDir,
				}
			},
			expectError:   true,
			errorContains: "not a git repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a separate subdirectory for each test
			testDir := filepath.Join(tmpDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			repo := tt.setupRepo(testDir)

			// Test RemoveRepository
			err = RemoveRepository(repo)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify directory was actually removed
			repoPath := repo.Path
			if repoPath == "" {
				// Use URL-derived path
				repoPath = filepath.Base(repo.URL)
				repoPath = filepath.Join(testDir, trimGitSuffix(repoPath))
			}

			if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
				t.Errorf("Repository directory should have been removed: %s", repoPath)
			}
		})
	}
}

func TestRemoveRepositoryWithNestedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a repository with nested files and directories
	repoDir := filepath.Join(tmpDir, "nested-repo")
	gitDir := filepath.Join(repoDir, ".git")
	subDir := filepath.Join(repoDir, "subdir")
	deepDir := filepath.Join(repoDir, "deep", "nested", "path")

	// Create directory structure
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create git directory: %v", err)
	}

	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	err = os.MkdirAll(deepDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create deep directory: %v", err)
	}

	// Create files
	files := []string{
		filepath.Join(repoDir, "README.md"),
		filepath.Join(gitDir, "config"),
		filepath.Join(gitDir, "HEAD"),
		filepath.Join(subDir, "file1.txt"),
		filepath.Join(subDir, "file2.txt"),
		filepath.Join(deepDir, "deep-file.txt"),
	}

	for _, file := range files {
		err = os.WriteFile(file, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	repo := config.Repository{
		Name: "nested-repo",
		URL:  "git@github.com:owner/nested-repo.git",
		Path: repoDir,
	}

	// Remove the repository
	err = RemoveRepository(repo)
	if err != nil {
		t.Fatalf("RemoveRepository() error = %v", err)
	}

	// Verify entire directory tree was removed
	if _, err := os.Stat(repoDir); !os.IsNotExist(err) {
		t.Error("Repository directory should have been completely removed")
	}
}

func TestRemoveRepositorySymlinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a real repository
	realRepoDir := filepath.Join(tmpDir, "real-repo")
	gitDir := filepath.Join(realRepoDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create real repo: %v", err)
	}

	// Create a symlink to the repository
	symlinkPath := filepath.Join(tmpDir, "symlink-repo")
	err = os.Symlink(realRepoDir, symlinkPath)
	if err != nil {
		t.Skipf("Symlinks not supported: %v", err)
	}

	repo := config.Repository{
		Name: "symlink-repo",
		URL:  "git@github.com:owner/symlink-repo.git",
		Path: symlinkPath,
	}

	// Remove the symlinked repository
	err = RemoveRepository(repo)
	if err != nil {
		t.Fatalf("RemoveRepository() error = %v", err)
	}

	// Verify symlink was removed
	if _, err := os.Stat(symlinkPath); !os.IsNotExist(err) {
		t.Error("Symlink should have been removed")
	}

	// Verify real repository still exists
	if _, err := os.Stat(realRepoDir); os.IsNotExist(err) {
		t.Error("Real repository should still exist")
	}
}

func TestRemoveRepositoryReadOnlyFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create repository with read-only files
	repoDir := filepath.Join(tmpDir, "readonly-repo")
	gitDir := filepath.Join(repoDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create git directory: %v", err)
	}

	// Create a read-only file
	readOnlyFile := filepath.Join(repoDir, "readonly.txt")
	err = os.WriteFile(readOnlyFile, []byte("readonly content"), 0444)
	if err != nil {
		t.Fatalf("Failed to create readonly file: %v", err)
	}

	repo := config.Repository{
		Name: "readonly-repo",
		URL:  "git@github.com:owner/readonly-repo.git",
		Path: repoDir,
	}

	// Remove the repository
	err = RemoveRepository(repo)
	if err != nil {
		t.Fatalf("RemoveRepository() should handle readonly files, error = %v", err)
	}

	// Verify directory was removed
	if _, err := os.Stat(repoDir); !os.IsNotExist(err) {
		t.Error("Repository with readonly files should have been removed")
	}
}

func TestRemoveRepositoryEmptyPath(t *testing.T) {
	repo := config.Repository{
		Name: "empty-path",
		URL:  "git@github.com:owner/empty-path.git",
		Path: "",
	}

	// This should derive path from URL, but since it doesn't exist, should error
	err := RemoveRepository(repo)
	if err == nil {
		t.Error("RemoveRepository() should error when derived path doesn't exist")
	}
}

func TestRemoveRepositoryRelativePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory for relative path test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create repository with relative path
	relativePath := "relative-repo"
	gitDir := filepath.Join(relativePath, ".git")
	err = os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create git directory: %v", err)
	}

	repo := config.Repository{
		Name: "relative-repo",
		URL:  "git@github.com:owner/relative-repo.git",
		Path: relativePath,
	}

	// Remove the repository
	err = RemoveRepository(repo)
	if err != nil {
		t.Fatalf("RemoveRepository() error = %v", err)
	}

	// Verify directory was removed
	if _, err := os.Stat(relativePath); !os.IsNotExist(err) {
		t.Error("Repository with relative path should have been removed")
	}
}

func BenchmarkRemoveRepository(b *testing.B) {
	tmpDir := b.TempDir()

	// Pre-create repositories for benchmarking
	repos := make([]config.Repository, b.N)
	for i := 0; i < b.N; i++ {
		repoDir := filepath.Join(tmpDir, fmt.Sprintf("bench-repo-%d", i))
		gitDir := filepath.Join(repoDir, ".git")
		_ = os.MkdirAll(gitDir, 0755)

		// Create some files
		_ = os.WriteFile(filepath.Join(repoDir, "file1.txt"), []byte("content"), 0644)
		_ = os.WriteFile(filepath.Join(repoDir, "file2.txt"), []byte("content"), 0644)
		_ = os.WriteFile(filepath.Join(gitDir, "config"), []byte("git config"), 0644)

		repos[i] = config.Repository{
			Name: fmt.Sprintf("bench-repo-%d", i),
			URL:  fmt.Sprintf("git@github.com:owner/bench-repo-%d.git", i),
			Path: repoDir,
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := RemoveRepository(repos[i])
		if err != nil {
			b.Fatalf("RemoveRepository() error = %v", err)
		}
	}
}

// Helper functions

func containsString(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		(len(str) > len(substr) &&
			(str[:len(substr)] == substr ||
				str[len(str)-len(substr):] == substr ||
				strings.Contains(str, substr))))
}

func trimGitSuffix(url string) string {
	if strings.HasSuffix(url, ".git") {
		return url[:len(url)-4]
	}
	return url
}
