// Package testutil provides common testing utilities across all packages.
package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/core"
	"gopkg.in/yaml.v2"
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

// RepositoryBuilder provides a fluent interface for building test repositories
type RepositoryBuilder struct {
	repo core.Repository
}

// NewRepositoryBuilder creates a new repository builder with sensible defaults
func NewRepositoryBuilder() *RepositoryBuilder {
	return &RepositoryBuilder{
		repo: core.Repository{
			Name:     "test-repo",
			Path:     "/tmp/test",
			Language: "go",
			Tags:     make([]string, 0),
			Metadata: make(map[string]string),
		},
	}
}

// WithName sets the repository name
func (rb *RepositoryBuilder) WithName(name string) *RepositoryBuilder {
	rb.repo.Name = name
	return rb
}

// WithPath sets the repository path
func (rb *RepositoryBuilder) WithPath(path string) *RepositoryBuilder {
	rb.repo.Path = path
	return rb
}

// WithURL sets the repository URL
func (rb *RepositoryBuilder) WithURL(url string) *RepositoryBuilder {
	rb.repo.URL = url
	return rb
}

// WithLanguage sets the repository language
func (rb *RepositoryBuilder) WithLanguage(language string) *RepositoryBuilder {
	rb.repo.Language = language
	return rb
}

// WithFramework sets the repository framework
func (rb *RepositoryBuilder) WithFramework(framework string) *RepositoryBuilder {
	rb.repo.Framework = framework
	return rb
}

// WithTags adds tags to the repository
func (rb *RepositoryBuilder) WithTags(tags ...string) *RepositoryBuilder {
	rb.repo.Tags = append(rb.repo.Tags, tags...)
	return rb
}

// WithBranch sets the repository branch
func (rb *RepositoryBuilder) WithBranch(branch string) *RepositoryBuilder {
	rb.repo.Branch = branch
	return rb
}

// WithMetadata adds metadata to the repository
func (rb *RepositoryBuilder) WithMetadata(key, value string) *RepositoryBuilder {
	rb.repo.Metadata[key] = value
	return rb
}

// Build creates the final Repository
func (rb *RepositoryBuilder) Build() core.Repository {
	return rb.repo
}

// CheckResultBuilder provides a fluent interface for building test check results
type CheckResultBuilder struct {
	result core.CheckResult
}

// NewCheckResultBuilder creates a new check result builder with sensible defaults
func NewCheckResultBuilder() *CheckResultBuilder {
	return &CheckResultBuilder{
		result: core.CheckResult{
			ID:         "test-check",
			Name:       "Test Check",
			Category:   "test",
			Status:     core.StatusHealthy,
			Score:      100,
			MaxScore:   100,
			Issues:     make([]core.Issue, 0),
			Warnings:   make([]core.Warning, 0),
			Metrics:    make(map[string]interface{}),
			Metadata:   make(map[string]string),
			Duration:   time.Millisecond * 100,
			Timestamp:  time.Now(),
			Repository: "test-repo",
		},
	}
}

// WithID sets the check result ID
func (crb *CheckResultBuilder) WithID(id string) *CheckResultBuilder {
	crb.result.ID = id
	return crb
}

// WithName sets the check result name
func (crb *CheckResultBuilder) WithName(name string) *CheckResultBuilder {
	crb.result.Name = name
	return crb
}

// WithCategory sets the check result category
func (crb *CheckResultBuilder) WithCategory(category string) *CheckResultBuilder {
	crb.result.Category = category
	return crb
}

// WithStatus sets the check result status
func (crb *CheckResultBuilder) WithStatus(status core.HealthStatus) *CheckResultBuilder {
	crb.result.Status = status
	return crb
}

// WithScore sets the check result score
func (crb *CheckResultBuilder) WithScore(score, maxScore int) *CheckResultBuilder {
	crb.result.Score = score
	crb.result.MaxScore = maxScore
	return crb
}

// WithIssue adds an issue to the check result
func (crb *CheckResultBuilder) WithIssue(issueType, message string, severity core.Severity) *CheckResultBuilder {
	issue := core.Issue{
		Type:     issueType,
		Message:  message,
		Severity: severity,
	}
	crb.result.Issues = append(crb.result.Issues, issue)
	return crb
}

// WithWarning adds a warning to the check result
func (crb *CheckResultBuilder) WithWarning(warningType, message string) *CheckResultBuilder {
	warning := core.Warning{
		Type:    warningType,
		Message: message,
	}
	crb.result.Warnings = append(crb.result.Warnings, warning)
	return crb
}

// WithMetric adds a metric to the check result
func (crb *CheckResultBuilder) WithMetric(key string, value interface{}) *CheckResultBuilder {
	crb.result.Metrics[key] = value
	return crb
}

// WithRepository sets the repository name for the check result
func (crb *CheckResultBuilder) WithRepository(repo string) *CheckResultBuilder {
	crb.result.Repository = repo
	return crb
}

// Build creates the final CheckResult
func (crb *CheckResultBuilder) Build() core.CheckResult {
	return crb.result
}

// TestEnvironment provides a comprehensive test environment with common test utilities
type TestEnvironment struct {
	TempDir   string
	ConfigDir string
	LogDir    string
	t         testing.TB
}

// NewTestEnvironment creates a new test environment with temporary directories
func NewTestEnvironment(t testing.TB) *TestEnvironment {
	t.Helper()

	tempDir := t.TempDir()

	env := &TestEnvironment{
		TempDir:   tempDir,
		ConfigDir: filepath.Join(tempDir, "config"),
		LogDir:    filepath.Join(tempDir, "logs"),
		t:         t,
	}

	// Create subdirectories
	if err := os.MkdirAll(env.ConfigDir, 0750); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	if err := os.MkdirAll(env.LogDir, 0750); err != nil {
		t.Fatalf("Failed to create log directory: %v", err)
	}

	return env
}

// CreateConfig creates a test configuration file
func (env *TestEnvironment) CreateConfig(filename, content string) string {
	env.t.Helper()

	configPath := filepath.Join(env.ConfigDir, filename)
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		env.t.Fatalf("Failed to create config file %s: %v", filename, err)
	}

	return configPath
}

// CreateRepoConfig creates a basic repository configuration
func (env *TestEnvironment) CreateRepoConfig(repos ...core.Repository) string {
	env.t.Helper()

	config := struct {
		Repositories []core.Repository `yaml:"repositories"`
	}{
		Repositories: repos,
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		env.t.Fatalf("Failed to marshal config: %v", err)
	}

	return env.CreateConfig("config.yaml", string(data))
}

// CreateAdvancedConfig creates an advanced configuration file
func (env *TestEnvironment) CreateAdvancedConfig(content string) string {
	env.t.Helper()
	return env.CreateConfig("advanced.yaml", content)
}

// GetConfigPath returns the path to a config file
func (env *TestEnvironment) GetConfigPath(filename string) string {
	return filepath.Join(env.ConfigDir, filename)
}

// GetLogPath returns the path to a log file
func (env *TestEnvironment) GetLogPath(filename string) string {
	return filepath.Join(env.LogDir, filename)
}
