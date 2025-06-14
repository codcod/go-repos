package util

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/codcod/repos/internal/config"
)

func TestFindGitRepositories(t *testing.T) {
	// Create temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test directory structure:
	// tmpDir/
	//   ├── repo1/           (git repo)
	//   │   └── .git/
	//   ├── repo2/           (git repo)
	//   │   ├── .git/
	//   │   │   └── config
	//   │   └── subdir/
	//   ├── nested/          (directory - will be ignored)
	//   │   └── deep/
	//   │       └── repo3/   (git repo - will be ignored)
	//   │           └── .git/
	//   ├── not-git/         (regular directory)
	//   └── empty-git/       (has .git but no config)
	//       └── .git/

	// Setup repo1
	repo1Dir := filepath.Join(tmpDir, "repo1")
	repo1GitDir := filepath.Join(repo1Dir, ".git")
	err := os.MkdirAll(repo1GitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create repo1 .git directory: %v", err)
	}
	createGitConfig(t, repo1GitDir, "git@github.com:owner/repo1.git")

	// Setup repo2
	repo2Dir := filepath.Join(tmpDir, "repo2")
	repo2GitDir := filepath.Join(repo2Dir, ".git")
	repo2SubDir := filepath.Join(repo2Dir, "subdir")
	err = os.MkdirAll(repo2GitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create repo2 .git directory: %v", err)
	}
	err = os.MkdirAll(repo2SubDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create repo2 subdir: %v", err)
	}
	createGitConfig(t, repo2GitDir, "https://github.com/owner/repo2.git")

	// Setup nested repo3 (will be ignored)
	repo3Dir := filepath.Join(tmpDir, "nested", "deep", "repo3")
	repo3GitDir := filepath.Join(repo3Dir, ".git")
	err = os.MkdirAll(repo3GitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create repo3 .git directory: %v", err)
	}
	createGitConfig(t, repo3GitDir, "git@github.com:owner/repo3.git")

	// Setup not-git directory
	notGitDir := filepath.Join(tmpDir, "not-git")
	err = os.MkdirAll(notGitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create not-git directory: %v", err)
	}

	// Setup empty-git directory (no config)
	emptyGitDir := filepath.Join(tmpDir, "empty-git", ".git")
	err = os.MkdirAll(emptyGitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create empty-git directory: %v", err)
	}

	repos, err := FindGitRepositories(tmpDir)
	if err != nil {
		t.Fatalf("FindGitRepositories() error = %v", err)
	}

	expectedRepos := []string{"repo1", "repo2"}
	if len(repos) != len(expectedRepos) {
		t.Errorf("FindGitRepositories() found %d repositories, expected %d", len(repos), len(expectedRepos))
		t.Errorf("Found: %v", getRepoNames(repos))
		t.Errorf("Expected: %v", expectedRepos)
		return
	}

	// Check that we found the expected repositories
	foundNames := getRepoNames(repos)
	for _, expectedName := range expectedRepos {
		if !contains(foundNames, expectedName) {
			t.Errorf("Expected to find repository '%s' but didn't", expectedName)
		}
	}

	// Verify repository details
	for _, repo := range repos {
		if repo.Name == "" {
			t.Error("Repository name should not be empty")
		}
		if repo.URL == "" {
			t.Error("Repository URL should not be empty")
		}
		if len(repo.Tags) == 0 {
			t.Error("Repository should have auto-discovered tag")
		}
		if repo.Tags[0] != "auto-discovered" {
			t.Errorf("Expected tag 'auto-discovered', got '%s'", repo.Tags[0])
		}
		if repo.Path == "" {
			t.Error("Repository path should not be empty")
		}

		// Verify path exists and is a git repository
		if !IsGitRepository(repo.Path) {
			t.Errorf("Repository path '%s' is not a git repository", repo.Path)
		}
	}
}

func TestFindGitRepositoriesEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	repos, err := FindGitRepositories(tmpDir)
	if err != nil {
		t.Fatalf("FindGitRepositories() error = %v", err)
	}

	if len(repos) != 0 {
		t.Errorf("Expected 0 repositories in empty directory, got %d", len(repos))
	}
}

func TestFindGitRepositoriesNonExistentDirectory(t *testing.T) {
	nonExistentDir := "/path/that/does/not/exist"

	_, err := FindGitRepositories(nonExistentDir)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestGetRemoteURL(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		gitConfig   string
		expectedURL string
		expectError bool
	}{
		{
			name: "valid SSH URL",
			gitConfig: `[core]
	repositoryformatversion = 0
	filemode = true
[remote "origin"]
	url = git@github.com:owner/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
	merge = refs/heads/main`,
			expectedURL: "git@github.com:owner/repo.git",
			expectError: false,
		},
		{
			name: "valid HTTPS URL",
			gitConfig: `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = https://github.com/owner/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*`,
			expectedURL: "https://github.com/owner/repo.git",
			expectError: false,
		},
		{
			name: "multiple remotes",
			gitConfig: `[core]
	repositoryformatversion = 0
[remote "origin"]
	url = git@github.com:owner/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[remote "upstream"]
	url = git@github.com:upstream/repo.git
	fetch = +refs/heads/*:refs/remotes/upstream/*`,
			expectedURL: "git@github.com:owner/repo.git",
			expectError: false,
		},
		{
			name: "no origin remote",
			gitConfig: `[core]
	repositoryformatversion = 0
[remote "upstream"]
	url = git@github.com:upstream/repo.git
	fetch = +refs/heads/*:refs/remotes/upstream/*`,
			expectedURL: "",
			expectError: false,
		},
		{
			name: "no remotes",
			gitConfig: `[core]
	repositoryformatversion = 0
	filemode = true`,
			expectedURL: "",
			expectError: false,
		},
		{
			name: "malformed config",
			gitConfig: `[core
	repositoryformatversion = 0
[remote "origin"
	url = git@github.com:owner/repo.git`,
			expectedURL: "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary repository directory
			repoDir := filepath.Join(tmpDir, tt.name)
			gitDir := filepath.Join(repoDir, ".git")
			err := os.MkdirAll(gitDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create git directory: %v", err)
			}

			// Write git config
			configPath := filepath.Join(gitDir, "config")
			err = os.WriteFile(configPath, []byte(tt.gitConfig), 0644)
			if err != nil {
				t.Fatalf("Failed to write git config: %v", err)
			}

			// Test GetRemoteURL
			url, err := GetRemoteURL(repoDir)

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

			if url != tt.expectedURL {
				t.Errorf("GetRemoteURL() = %v, want %v", url, tt.expectedURL)
			}
		})
	}
}

func TestGetRemoteURLNoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create repository directory without config file
	repoDir := filepath.Join(tmpDir, "no-config")
	gitDir := filepath.Join(repoDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create git directory: %v", err)
	}

	_, err = GetRemoteURL(repoDir)
	if err == nil {
		t.Error("Expected error for missing config file")
	}
}

func BenchmarkFindGitRepositories(b *testing.B) {
	// Create temporary directory with multiple repositories
	tmpDir := b.TempDir()

	// Create 10 repositories for benchmarking
	for i := 0; i < 10; i++ {
		repoDir := filepath.Join(tmpDir, fmt.Sprintf("repo%d", i))
		gitDir := filepath.Join(repoDir, ".git")
		err := os.MkdirAll(gitDir, 0755)
		if err != nil {
			b.Fatalf("Failed to create repo%d: %v", i, err)
		}
		createGitConfig(b, gitDir, fmt.Sprintf("git@github.com:owner/repo%d.git", i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := FindGitRepositories(tmpDir)
		if err != nil {
			b.Fatalf("FindGitRepositories() error = %v", err)
		}
	}
}

func BenchmarkGetRemoteURL(b *testing.B) {
	tmpDir := b.TempDir()
	repoDir := filepath.Join(tmpDir, "bench-repo")
	gitDir := filepath.Join(repoDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		b.Fatalf("Failed to create git directory: %v", err)
	}

	gitConfig := `[core]
	repositoryformatversion = 0
	filemode = true
[remote "origin"]
	url = git@github.com:owner/repo.git
	fetch = +refs/heads/*:refs/remotes/origin/*`

	configPath := filepath.Join(gitDir, "config")
	err = os.WriteFile(configPath, []byte(gitConfig), 0644)
	if err != nil {
		b.Fatalf("Failed to write git config: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := GetRemoteURL(repoDir)
		if err != nil {
			b.Fatalf("GetRemoteURL() error = %v", err)
		}
	}
}

// Helper functions

func createGitConfig(t testing.TB, gitDir, remoteURL string) {
	gitConfig := fmt.Sprintf(`[core]
	repositoryformatversion = 0
	filemode = true
[remote "origin"]
	url = %s
	fetch = +refs/heads/*:refs/remotes/origin/*`, remoteURL)

	configPath := filepath.Join(gitDir, "config")
	err := os.WriteFile(configPath, []byte(gitConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}
}

func getRepoNames(repos []config.Repository) []string {
	names := make([]string, len(repos))
	for i, repo := range repos {
		names[i] = repo.Name
	}
	return names
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
