// Package testutil provides common testing utilities across all packages.
package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/codcod/repos/internal/config"
)

// CreateTempConfig writes the given YAML to a temp file and returns its path.
func CreateTempConfig(t testing.TB, configYAML string) string {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0600); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	return configPath
}

// CreateGitConfig creates a minimal git config for testing.
func CreateGitConfig(t testing.TB, gitDir, remoteURL string) {
	t.Helper()
	gitConfig := fmt.Sprintf(`[core]
	repositoryformatversion = 0
	filemode = true
[remote "origin"]
	url = %s
	fetch = +refs/heads/*:refs/remotes/origin/*`, remoteURL)

	configPath := filepath.Join(gitDir, "config")
	if err := os.WriteFile(configPath, []byte(gitConfig), 0600); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}
}

// CreateMockGitRepo creates a mock git repository with the given URL.
func CreateMockGitRepo(t testing.TB, baseDir, repoName, remoteURL string) string {
	t.Helper()
	repoDir := filepath.Join(baseDir, repoName)
	gitDir := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(gitDir, 0750); err != nil {
		t.Fatalf("Failed to create git directory: %v", err)
	}
	CreateGitConfig(t, gitDir, remoteURL)
	return repoDir
}

// CreateRealGitRepo creates a real git repository for integration tests.
func CreateRealGitRepo(t testing.TB, repoDir string) {
	t.Helper()

	// Initialize git repo
	if err := exec.Command("git", "init", repoDir).Run(); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Configure git user
	_ = exec.Command("git", "-C", repoDir, "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "-C", repoDir, "config", "user.name", "Test User").Run()

	// Create initial commit
	testFile := filepath.Join(repoDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repository"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	_ = exec.Command("git", "-C", repoDir, "add", "README.md").Run()
	_ = exec.Command("git", "-C", repoDir, "commit", "-m", "initial commit").Run()
}

// Contains checks if a string is in a slice.
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetRepoNames extracts repository names from a slice of repositories.
func GetRepoNames(repos []config.Repository) []string {
	names := make([]string, len(repos))
	for i, repo := range repos {
		names[i] = repo.Name
	}
	return names
}

// StandardTestConfig provides a consistent test configuration for benchmarks and tests.
const StandardTestConfig = `repositories:
  - name: go-app
    url: git@github.com:owner/go-app.git
    tags: [go, backend]
    branch: main
  - name: react-ui
    url: git@github.com:owner/react-ui.git
    tags: [javascript, frontend]
    branch: develop
  - name: python-api
    url: git@github.com:owner/python-api.git
    tags: [python, backend, api]`

// CreateBenchmarkRepos creates repositories for benchmarking with specified count.
func CreateBenchmarkRepos(b testing.TB, count int) []config.Repository {
	b.Helper()
	repos := make([]config.Repository, count)
	for i := 0; i < count; i++ {
		repos[i] = config.Repository{
			Name: fmt.Sprintf("repo-%d", i),
			URL:  fmt.Sprintf("git@github.com:owner/repo-%d.git", i),
			Tags: []string{"test", "benchmark"},
		}
	}
	return repos
}

// SkipIfGitNotAvailable skips the test if git is not available.
func SkipIfGitNotAvailable(t testing.TB) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available, skipping test")
	}
}
