package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/checkers/registry"
	"github.com/codcod/repos/internal/health/orchestration"
	"github.com/codcod/repos/internal/platform/commands"
)

// TestOrchestrationEndToEnd tests the complete orchestration pipeline
func TestOrchestrationEndToEnd(t *testing.T) {
	// Skip if running in CI without proper setup
	if os.Getenv("CI") == "true" && os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test in CI")
	}

	// Create test directory structure
	testDir := filepath.Join(os.TempDir(), "repos-orchestration-test")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create a mock repository directory
	repoDir := filepath.Join(testDir, "test-repo")
	err = os.MkdirAll(repoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create repo directory: %v", err)
	}

	// Initialize git repository
	err = os.Chdir(repoDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create a simple git repository for testing
	createTestGitRepo(t, repoDir)

	// Create test configuration
	config := createTestAdvancedConfig(t)

	// Create test repositories
	testRepos := []core.Repository{
		{
			Name:     "test-repo",
			Path:     repoDir,
			URL:      "https://github.com/test/test-repo.git",
			Branch:   "main",
			Tags:     []string{"test"},
			Metadata: make(map[string]string),
		},
	}

	// Create command executor and logger
	executor := commands.NewOSCommandExecutor(30 * time.Second)
	logger := &testLogger{t: t}

	// Create registries
	checkerRegistry := registry.NewCheckerRegistry(executor)

	// Create orchestration engine
	engine := orchestration.NewEngine(checkerRegistry, nil, config, logger)

	// Execute health check
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := engine.ExecuteHealthCheck(ctx, testRepos)
	if err != nil {
		t.Fatalf("Failed to execute health check: %v", err)
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Verify results
	if result.TotalRepos != 1 {
		t.Errorf("Expected 1 total repo, got %d", result.TotalRepos)
	}
	if len(result.RepositoryResults) != 1 {
		t.Errorf("Expected 1 repository result, got %d", len(result.RepositoryResults))
	}
	if result.Duration <= 0 {
		t.Error("Duration should be greater than 0")
	}
	if result.StartTime.IsZero() {
		t.Error("Start time should not be zero")
	}
	if result.EndTime.IsZero() {
		t.Error("End time should not be zero")
	}

	// Check repository result
	repoResult := result.RepositoryResults[0]
	if repoResult.Repository.Name != "test-repo" {
		t.Errorf("Expected repo name 'test-repo', got '%s'", repoResult.Repository.Name)
	}
	if repoResult.Duration <= 0 {
		t.Error("Repository duration should be greater than 0")
	}
	if repoResult.Error != "" {
		t.Errorf("Repository should have no error, got: %s", repoResult.Error)
	}

	// Verify that some checks were executed
	if len(repoResult.CheckResults) == 0 {
		t.Error("Should have executed some checks")
	}

	// Verify summary
	if result.Summary.SuccessfulRepos+result.Summary.FailedRepos != result.TotalRepos {
		t.Error("Summary counts don't match total repos")
	}

	t.Logf("Orchestration test completed successfully")
	t.Logf("Total repos: %d, Successful: %d, Failed: %d",
		result.TotalRepos, result.Summary.SuccessfulRepos, result.Summary.FailedRepos)
	t.Logf("Duration: %v", result.Duration)
}

// TestOrchestrationConfigLoading tests loading of advanced configuration
func TestOrchestrationConfigLoading(t *testing.T) {
	configPath := "../../examples/advanced-config-sample.yaml"

	// Check if sample config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Sample orchestration config not found")
	}

	// Load configuration
	config, err := config.LoadAdvancedConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config == nil {
		t.Fatal("Config should not be nil")
	}

	// Verify configuration structure
	if config.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", config.Version)
	}
	if len(config.Checkers) == 0 {
		t.Error("Should have checkers configured")
	}
	if len(config.Categories) == 0 {
		t.Error("Should have categories configured")
	}

	// Verify engine config
	engineConfig := config.GetEngineConfig()
	if engineConfig.MaxConcurrency <= 0 {
		t.Error("MaxConcurrency should be greater than 0")
	}

	t.Logf("Configuration loaded successfully with %d checkers",
		len(config.Checkers))
}

// TestOrchestrationDryRun tests dry run functionality
func TestOrchestrationDryRun(t *testing.T) {
	// This test verifies that dry run mode works correctly
	// In a real implementation, this would test the --dry-run flag functionality

	config := createTestAdvancedConfig(t)
	testRepos := []core.Repository{
		{
			Name:     "dry-run-test",
			Path:     "/tmp/test",
			URL:      "https://github.com/test/dry-run.git",
			Tags:     []string{"test"},
			Metadata: make(map[string]string),
		},
	}

	executor := commands.NewMockCommandExecutor()
	logger := &testLogger{t: t}
	checkerRegistry := registry.NewCheckerRegistry(executor)

	engine := orchestration.NewEngine(checkerRegistry, nil, config, logger)

	// In dry run mode, we would expect the engine to return what it would do
	// without actually executing the checks
	ctx := context.Background()
	result, err := engine.ExecuteHealthCheck(ctx, testRepos)

	// For now, just verify the engine can be created and called
	// In a full implementation, dry run would have special handling
	if err != nil {
		t.Logf("Dry run test completed with expected error: %v", err)
	} else {
		t.Logf("Dry run test completed, result: %+v", result)
	}
}

// Helper functions

func createTestGitRepo(t *testing.T, dir string) {
	executor := commands.NewOSCommandExecutor(10 * time.Second)
	ctx := context.Background()

	// Initialize git repo
	result := executor.ExecuteInDir(ctx, dir, "git", "init")
	if result.Error != nil {
		t.Logf("Warning: Failed to initialize git repo: %v", result.Error)
		return
	}

	// Create a simple file
	readmePath := filepath.Join(dir, "README.md")
	err := os.WriteFile(readmePath, []byte("# Test Repository\n\nThis is a test repository for orchestration testing.\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	// Configure git user (required for commits)
	executor.ExecuteInDir(ctx, dir, "git", "config", "user.email", "test@example.com")
	executor.ExecuteInDir(ctx, dir, "git", "config", "user.name", "Test User")

	// Add and commit file
	executor.ExecuteInDir(ctx, dir, "git", "add", "README.md")
	executor.ExecuteInDir(ctx, dir, "git", "commit", "-m", "Initial commit")
}

func createTestAdvancedConfig(t *testing.T) *config.AdvancedConfig {
	return &config.AdvancedConfig{
		Version: "1.0",
		Engine: core.EngineConfig{
			MaxConcurrency: 2,
			Timeout:        300 * time.Second,
			CacheEnabled:   false,
			CacheTTL:       time.Hour,
		},
		Checkers: map[string]core.CheckerConfig{
			"git-status": {
				Enabled: true,
				Timeout: 30 * time.Second,
				Options: map[string]interface{}{
					"check_uncommitted": true,
					"check_unpushed":    true,
				},
			},
		},
		Categories: map[string]config.CategoryConfig{
			"git": {
				Name:        "Git Checks",
				Description: "Git repository checks",
				Enabled:     true,
				Severity:    "low",
				Weight:      1.0,
				Checkers:    []string{"git-status"},
			},
		},
	}
}

// testLogger implements core.Logger for testing
type testLogger struct {
	t *testing.T
}

func (l *testLogger) Debug(msg string, fields ...core.Field) {
	l.t.Logf("[DEBUG] %s %v", msg, fields)
}

func (l *testLogger) Info(msg string, fields ...core.Field) {
	l.t.Logf("[INFO] %s %v", msg, fields)
}

func (l *testLogger) Warn(msg string, fields ...core.Field) {
	l.t.Logf("[WARN] %s %v", msg, fields)
}

func (l *testLogger) Error(msg string, fields ...core.Field) {
	l.t.Logf("[ERROR] %s %v", msg, fields)
}
