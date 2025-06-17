package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codcod/repos/internal/config"
)

func TestCreateTempConfig(t *testing.T) {
	testYAML := `repositories:
  - name: test-repo
    url: git@github.com:test/test.git`

	configPath := CreateTempConfig(t, testYAML)

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created: %s", configPath)
	}

	// Verify file contents
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if string(content) != testYAML {
		t.Errorf("Config content mismatch. Expected:\n%s\nGot:\n%s", testYAML, string(content))
	}

	// Verify it's in a temp directory (check if path starts with os.TempDir())
	tempDir := os.TempDir()
	if !strings.HasPrefix(configPath, tempDir) {
		t.Errorf("Config should be in temp directory, got: %s", configPath)
	}
}

func TestCreateGitConfig(t *testing.T) {
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create git directory: %v", err)
	}

	remoteURL := "git@github.com:test/repo.git"
	CreateGitConfig(t, gitDir, remoteURL)

	// Verify git config file exists
	configPath := filepath.Join(gitDir, "config")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Git config file was not created: %s", configPath)
	}

	// Verify config content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read git config: %v", err)
	}

	configStr := string(content)
	if !strings.Contains(configStr, remoteURL) {
		t.Errorf("Git config should contain remote URL %s, got:\n%s", remoteURL, configStr)
	}

	if !strings.Contains(configStr, "[core]") {
		t.Error("Git config should contain [core] section")
	}

	if !strings.Contains(configStr, "[remote \"origin\"]") {
		t.Error("Git config should contain remote origin section")
	}
}

func TestCreateMockGitRepo(t *testing.T) {
	baseDir := t.TempDir()
	repoName := "test-repo"
	remoteURL := "git@github.com:test/repo.git"

	repoDir := CreateMockGitRepo(t, baseDir, repoName, remoteURL)

	// Verify repo directory was created
	expectedPath := filepath.Join(baseDir, repoName)
	if repoDir != expectedPath {
		t.Errorf("Expected repo dir %s, got %s", expectedPath, repoDir)
	}

	// Verify .git directory exists
	gitDir := filepath.Join(repoDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error(".git directory should exist")
	}

	// Verify git config exists
	configPath := filepath.Join(gitDir, "config")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("git config should exist")
	}

	// Verify config contains remote URL
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read git config: %v", err)
	}

	if !strings.Contains(string(content), remoteURL) {
		t.Errorf("Config should contain remote URL %s", remoteURL)
	}
}

func TestCreateRealGitRepo(t *testing.T) {
	// Skip if git is not available
	SkipIfGitNotAvailable(t)

	repoDir := filepath.Join(t.TempDir(), "real-repo")
	CreateRealGitRepo(t, repoDir)

	// Verify .git directory exists
	gitDir := filepath.Join(repoDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error(".git directory should exist")
	}

	// Verify README.md was created
	readmePath := filepath.Join(repoDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("README.md should exist")
	}

	// Verify README content
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README: %v", err)
	}

	expectedContent := "# Test Repository"
	if string(content) != expectedContent {
		t.Errorf("README content mismatch. Expected: %s, Got: %s", expectedContent, string(content))
	}
}

func TestContains(t *testing.T) {
	testCases := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "item exists",
			slice:    []string{"apple", "banana", "cherry"},
			item:     "banana",
			expected: true,
		},
		{
			name:     "item does not exist",
			slice:    []string{"apple", "banana", "cherry"},
			item:     "orange",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "apple",
			expected: false,
		},
		{
			name:     "empty item",
			slice:    []string{"apple", "", "cherry"},
			item:     "",
			expected: true,
		},
		{
			name:     "case sensitive",
			slice:    []string{"Apple", "banana", "Cherry"},
			item:     "apple",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Contains(tc.slice, tc.item)
			if result != tc.expected {
				t.Errorf("Contains(%v, %s) = %v, expected %v", tc.slice, tc.item, result, tc.expected)
			}
		})
	}
}

func TestGetRepoNames(t *testing.T) {
	testCases := []struct {
		name     string
		repos    []config.Repository
		expected []string
	}{
		{
			name:     "empty repos",
			repos:    []config.Repository{},
			expected: []string{},
		},
		{
			name: "single repo",
			repos: []config.Repository{
				{Name: "repo1", URL: "git@github.com:owner/repo1.git"},
			},
			expected: []string{"repo1"},
		},
		{
			name: "multiple repos",
			repos: []config.Repository{
				{Name: "repo1", URL: "git@github.com:owner/repo1.git"},
				{Name: "repo2", URL: "git@github.com:owner/repo2.git"},
				{Name: "repo3", URL: "git@github.com:owner/repo3.git"},
			},
			expected: []string{"repo1", "repo2", "repo3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetRepoNames(tc.repos)

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d names, got %d", len(tc.expected), len(result))
				return
			}

			for i, expected := range tc.expected {
				if result[i] != expected {
					t.Errorf("Expected name[%d] = %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

func TestCreateBenchmarkRepos(t *testing.T) {
	testCases := []struct {
		name  string
		count int
	}{
		{name: "zero repos", count: 0},
		{name: "single repo", count: 1},
		{name: "multiple repos", count: 5},
		{name: "many repos", count: 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repos := CreateBenchmarkRepos(t, tc.count)

			if len(repos) != tc.count {
				t.Errorf("Expected %d repos, got %d", tc.count, len(repos))
			}

			// Verify repository structure
			for _, repo := range repos {
				if !strings.HasPrefix(repo.Name, "repo-") {
					t.Errorf("Repo name should start with 'repo-', got %s", repo.Name)
				}

				if !strings.Contains(repo.URL, "git@github.com") {
					t.Errorf("URL should contain github.com, got %s", repo.URL)
				}

				if !Contains(repo.Tags, "test") {
					t.Error("Repository should have 'test' tag")
				}

				if !Contains(repo.Tags, "benchmark") {
					t.Error("Repository should have 'benchmark' tag")
				}
			}
		})
	}
}

func TestStandardTestConfig(t *testing.T) {
	if StandardTestConfig == "" {
		t.Error("StandardTestConfig should not be empty")
	}

	// Verify it contains expected repositories
	if !strings.Contains(StandardTestConfig, "go-app") {
		t.Error("StandardTestConfig should contain go-app repository")
	}

	if !strings.Contains(StandardTestConfig, "react-ui") {
		t.Error("StandardTestConfig should contain react-ui repository")
	}

	if !strings.Contains(StandardTestConfig, "python-api") {
		t.Error("StandardTestConfig should contain python-api repository")
	}

	// Verify it contains YAML structure
	if !strings.Contains(StandardTestConfig, "repositories:") {
		t.Error("StandardTestConfig should contain repositories section")
	}

	if !strings.Contains(StandardTestConfig, "name:") {
		t.Error("StandardTestConfig should contain name fields")
	}

	if !strings.Contains(StandardTestConfig, "url:") {
		t.Error("StandardTestConfig should contain url fields")
	}
}

func TestSkipIfGitNotAvailable(t *testing.T) {
	// This test is tricky to test directly since we can't easily mock exec.LookPath
	// We'll just verify it doesn't panic when called and that it calls Helper()

	// Create a subtest that we expect might be skipped
	t.Run("git_availability", func(t *testing.T) {
		// This should not panic
		SkipIfGitNotAvailable(t)

		// If we reach here, git is available
		t.Log("Git is available on this system")
	})
}
