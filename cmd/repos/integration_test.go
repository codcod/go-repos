package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Integration tests for the repos CLI tool
// These tests require the binary to be built and test end-to-end functionality

func TestMain(m *testing.M) {
	// Build the binary before running integration tests
	if err := buildBinary(); err != nil {
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	cleanupBinary()

	os.Exit(code)
}

func buildBinary() error {
	cmd := exec.Command("go", "build", "-o", "repos-test", ".")
	return cmd.Run()
}

func cleanupBinary() {
	_ = os.Remove("repos-test")
}

func TestCLIVersion(t *testing.T) {
	cmd := exec.Command("./repos-test", "version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "repos") {
		t.Errorf("version output should contain 'repos', got: %s", outputStr)
	}
}

func TestCLIHelp(t *testing.T) {
	cmd := exec.Command("./repos-test", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	outputStr := string(output)
	expectedSections := []string{
		"Usage:",
		"Available Commands:",
		"Flags:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(outputStr, section) {
			t.Errorf("help output should contain '%s', got: %s", section, outputStr)
		}
	}
}

func TestCLIInvalidCommand(t *testing.T) {
	cmd := exec.Command("./repos-test", "invalid-command")
	_, err := cmd.Output()
	if err == nil {
		t.Error("invalid command should return an error")
	}
}

func TestCLICloneWithoutConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "clone")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("clone without config should fail")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Error") {
		t.Errorf("error output should contain 'Error', got: %s", outputStr)
	}
}

func TestCLIInitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Create a mock git repository
	repoDir := filepath.Join(tmpDir, "test-repo")
	gitDir := filepath.Join(repoDir, ".git")
	_ = os.MkdirAll(gitDir, 0755)

	// Create a mock git config
	configContent := `[remote "origin"]
	url = git@github.com:owner/test-repo.git`
	_ = os.WriteFile(filepath.Join(gitDir, "config"), []byte(configContent), 0644)

	// Run init command
	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "init", "--overwrite")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("init command failed: %v, output: %s", err, string(output))
	}

	// Check that config.yaml was created
	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		t.Error("config.yaml should have been created")
	}

	// Check content of config.yaml
	content, err := os.ReadFile("config.yaml")
	if err != nil {
		t.Fatalf("failed to read config.yaml: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "test-repo") {
		t.Errorf("config.yaml should contain 'test-repo', got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "repositories:") {
		t.Errorf("config.yaml should contain 'repositories:', got: %s", contentStr)
	}
}

func TestCLIRunCommandWithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Create a test repository directory
	repoDir := filepath.Join(tmpDir, "test-repo")
	_ = os.MkdirAll(repoDir, 0755)

	// Create a test config.yaml
	configContent := `repositories:
  - name: test-repo
    url: git@github.com:owner/test-repo.git
    tags: [test]
    path: ` + repoDir

	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config.yaml: %v", err)
	}

	// Run a simple command
	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "run", "echo", "hello world")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("run command failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "test-repo") {
		t.Errorf("output should contain repository name, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "hello world") {
		t.Errorf("output should contain command output, got: %s", outputStr)
	}
}

func TestCLIRunCommandWithTag(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Create test repository directories
	repo1Dir := filepath.Join(tmpDir, "repo1")
	repo2Dir := filepath.Join(tmpDir, "repo2")
	_ = os.MkdirAll(repo1Dir, 0755)
	_ = os.MkdirAll(repo2Dir, 0755)

	// Create a test config.yaml with multiple repos and tags
	configContent := `repositories:
  - name: repo1
    url: git@github.com:owner/repo1.git
    tags: [backend, go]
    path: ` + repo1Dir + `
  - name: repo2
    url: git@github.com:owner/repo2.git
    tags: [frontend, react]
    path: ` + repo2Dir

	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config.yaml: %v", err)
	}

	// Run command with specific tag
	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "run", "-t", "backend", "echo", "backend-only")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("run command with tag failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "repo1") {
		t.Errorf("output should contain repo1 (tagged with backend), got: %s", outputStr)
	}
	if strings.Contains(outputStr, "repo2") {
		t.Errorf("output should not contain repo2 (not tagged with backend), got: %s", outputStr)
	}
}

func TestCLIWithCustomConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Create a test repository directory
	repoDir := filepath.Join(tmpDir, "custom-repo")
	_ = os.MkdirAll(repoDir, 0755)

	// Create a custom config file
	customConfigContent := `repositories:
  - name: custom-repo
    url: git@github.com:owner/custom-repo.git
    tags: [custom]
    path: ` + repoDir

	err := os.WriteFile("custom-config.yaml", []byte(customConfigContent), 0644)
	if err != nil {
		t.Fatalf("failed to create custom config: %v", err)
	}

	// Run command with custom config
	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "-c", "custom-config.yaml", "run", "pwd")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("run command with custom config failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "custom-repo") {
		t.Errorf("output should contain custom repository name, got: %s", outputStr)
	}
}

func TestCLILogging(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Create a test repository directory
	repoDir := filepath.Join(tmpDir, "log-test-repo")
	_ = os.MkdirAll(repoDir, 0755)

	// Create a test config.yaml
	configContent := `repositories:
  - name: log-test-repo
    url: git@github.com:owner/log-test-repo.git
    tags: [test]
    path: ` + repoDir

	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config.yaml: %v", err)
	}

	// Run command with custom log directory
	logDir := filepath.Join(tmpDir, "custom-logs")
	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "run", "-l", logDir, "echo", "test logging")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("run command with logging failed: %v, output: %s", err, string(output))
	}

	// Check that log directory was created
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("custom log directory should have been created")
	}

	// Check that log files were created
	logFiles, err := filepath.Glob(filepath.Join(logDir, "log-test-repo_*.log"))
	if err != nil {
		t.Fatalf("failed to search for log files: %v", err)
	}
	if len(logFiles) == 0 {
		t.Error("at least one log file should have been created")
	}

	// Check log file content
	if len(logFiles) > 0 {
		logContent, err := os.ReadFile(logFiles[0])
		if err != nil {
			t.Fatalf("failed to read log file: %v", err)
		}

		logStr := string(logContent)
		if !strings.Contains(logStr, "test logging") {
			t.Errorf("log file should contain command output, got: %s", logStr)
		}
		if !strings.Contains(logStr, "Repository: log-test-repo") {
			t.Errorf("log file should contain repository info, got: %s", logStr)
		}
	}
}

func TestCLIEnvironmentVariables(t *testing.T) {
	// Test version command with custom environment variables
	cmd := exec.Command("./repos-test", "version")
	cmd.Env = append(os.Environ(),
		"VERSION=test-version",
		"COMMIT=test-commit",
		"BUILD_DATE=test-date")

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("version command with env vars failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "test-version") {
		t.Errorf("version output should contain custom version, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "test-commit") {
		t.Errorf("version output should contain custom commit, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "test-date") {
		t.Errorf("version output should contain custom date, got: %s", outputStr)
	}
}

func TestCLIParallelExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping parallel execution test in short mode")
	}

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Create multiple test repository directories
	numRepos := 3
	var configLines []string

	for i := 0; i < numRepos; i++ {
		repoDir := filepath.Join(tmpDir, fmt.Sprintf("parallel-repo-%d", i))
		_ = os.MkdirAll(repoDir, 0755)

		configLines = append(configLines, fmt.Sprintf(`  - name: parallel-repo-%d
    url: git@github.com:owner/parallel-repo-%d.git
    tags: [parallel]
    path: %s`, i, i, repoDir))
	}

	// Create config with multiple repositories
	configContent := "repositories:\n" + strings.Join(configLines, "\n")
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config.yaml: %v", err)
	}

	// Run command in parallel and measure time
	start := time.Now()
	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "run", "-p", "sleep", "1")
	output, err := cmd.Output()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("parallel run command failed: %v, output: %s", err, string(output))
	}

	// Parallel execution should complete faster than sequential (less than numRepos seconds)
	maxExpectedDuration := time.Duration(numRepos-1) * time.Second
	if duration > maxExpectedDuration {
		t.Errorf("parallel execution took too long: %v (expected less than %v)", duration, maxExpectedDuration)
	}

	// Verify all repositories were processed
	outputStr := string(output)
	for i := 0; i < numRepos; i++ {
		repoName := fmt.Sprintf("parallel-repo-%d", i)
		if !strings.Contains(outputStr, repoName) {
			t.Errorf("output should contain %s, got: %s", repoName, outputStr)
		}
	}
}

func TestCLIErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Create a test repository directory
	repoDir := filepath.Join(tmpDir, "error-test-repo")
	_ = os.MkdirAll(repoDir, 0755)

	// Create a test config.yaml
	configContent := `repositories:
  - name: error-test-repo
    url: git@github.com:owner/error-test-repo.git
    tags: [test]
    path: ` + repoDir

	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config.yaml: %v", err)
	}

	// Run a command that should fail
	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "run", "exit", "1")
	output, err := cmd.CombinedOutput()

	// Command should fail but repos should handle it gracefully
	if err == nil {
		t.Error("command with exit 1 should fail")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "error-test-repo") {
		t.Errorf("error output should still contain repository name, got: %s", outputStr)
	}
}

func TestCLIHealthCommand(t *testing.T) {
	cmd := exec.Command("./repos-test", "health", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("health help command failed: %v", err)
	}

	outputStr := string(output)
	expectedSections := []string{
		"Execute modular health checks",
		"Available Commands:",
		"complexity",
		"genconfig",
		"--list-categories",
		"--gen-config",
		"--complexity-report",
	}

	for _, section := range expectedSections {
		if !strings.Contains(outputStr, section) {
			t.Errorf("health help output should contain '%s', got: %s", section, outputStr)
		}
	}
}

func TestCLIHealthListCategories(t *testing.T) {
	cmd := exec.Command("./repos-test", "health", "--list-categories")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("health list-categories command failed: %v", err)
	}

	outputStr := string(output)
	expectedSections := []string{
		"Available Health Check Categories:",
		"CHECKERS:",
		"ANALYZERS:",
		"Total Checkers:",
		"Total Analyzers:",
		"Usage Examples:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(outputStr, section) {
			t.Errorf("health list-categories output should contain '%s', got: %s", section, outputStr)
		}
	}
}

func TestCLIHealthGenConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "health", "genconfig", "--overwrite")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("health genconfig command failed: %v, output: %s", err, string(output))
	}

	// Check that health-config.yaml was created
	if _, err := os.Stat("health-config.yaml"); os.IsNotExist(err) {
		t.Error("health-config.yaml should have been created")
	}

	// Check content of health-config.yaml
	content, err := os.ReadFile("health-config.yaml")
	if err != nil {
		t.Fatalf("failed to read health-config.yaml: %v", err)
	}

	contentStr := string(content)
	expectedSections := []string{
		"health:",
		"checkers:",
		"timeout:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(contentStr, section) {
			t.Errorf("health-config.yaml should contain '%s', got: %s", section, contentStr)
		}
	}
}

func TestCLIHealthDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Create a test repository directory
	repoDir := filepath.Join(tmpDir, "health-test-repo")
	_ = os.MkdirAll(repoDir, 0755)

	// Create a basic config.yaml for testing
	configContent := `repositories:
  - name: health-test-repo
    url: git@github.com:owner/health-test-repo.git
    tags: [test]
    path: ` + repoDir

	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config.yaml: %v", err)
	}

	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "health", "--dry-run")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("health --dry-run command failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "DRY RUN MODE") || !strings.Contains(outputStr, "Configuration Preview") {
		t.Errorf("dryrun output should contain dry run indicators, got: %s", outputStr)
	}
}

func TestCLIHealthComplexity(t *testing.T) {
	cmd := exec.Command("./repos-test", "health", "complexity", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("health complexity help command failed: %v", err)
	}

	outputStr := string(output)
	expectedSections := []string{
		"cyclomatic complexity",
		"--max-complexity",
		"--tag",
	}

	for _, section := range expectedSections {
		if !strings.Contains(outputStr, section) {
			t.Errorf("health complexity help should contain '%s', got: %s", section, outputStr)
		}
	}
}

func TestCLIHealthInvalidFlag(t *testing.T) {
	cmd := exec.Command("./repos-test", "health", "--invalid-flag")
	_, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("health command with invalid flag should fail")
	}
}

func TestCLIAllCommandsHelp(t *testing.T) {
	commands := []string{"clone", "health", "init", "pr", "rm", "run"}

	for _, cmdName := range commands {
		t.Run(cmdName+"_help", func(t *testing.T) {
			cmd := exec.Command("./repos-test", cmdName, "--help")
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("%s help command failed: %v", cmdName, err)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "Usage:") {
				t.Errorf("%s help should contain 'Usage:', got: %s", cmdName, outputStr)
			}
		})
	}
}

func TestCLIGlobalFlags(t *testing.T) {
	cmd := exec.Command("./repos-test", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("global help command failed: %v", err)
	}

	outputStr := string(output)
	expectedGlobalFlags := []string{
		"--config",
		"--tag",
		"--parallel",
	}

	for _, flag := range expectedGlobalFlags {
		if !strings.Contains(outputStr, flag) {
			t.Errorf("global help should contain flag '%s', got: %s", flag, outputStr)
		}
	}
}

func TestCLIConfigFlag(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Create a custom config file with different name
	customConfigContent := `repositories:
  - name: custom-config-repo
    url: git@github.com:owner/custom-config-repo.git
    tags: [custom]
    path: ` + filepath.Join(tmpDir, "custom-config-repo")

	_ = os.MkdirAll(filepath.Join(tmpDir, "custom-config-repo"), 0755)
	err := os.WriteFile("my-custom.yaml", []byte(customConfigContent), 0644)
	if err != nil {
		t.Fatalf("failed to create custom config: %v", err)
	}

	// Test using --config flag
	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "--config", "my-custom.yaml", "run", "echo", "test")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("command with --config flag failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "custom-config-repo") {
		t.Errorf("output should contain custom repository name, got: %s", outputStr)
	}
}

func TestCLIVersionCommand(t *testing.T) {
	// Test version command output format
	cmd := exec.Command("./repos-test", "version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	outputStr := string(output)
	expectedFields := []string{
		"repos",
		"built",
	}

	for _, field := range expectedFields {
		if !strings.Contains(outputStr, field) {
			t.Errorf("version output should contain '%s', got: %s", field, outputStr)
		}
	}
}

func TestCLICommandCompletion(t *testing.T) {
	// Test that all expected commands are available
	cmd := exec.Command("./repos-test", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	outputStr := string(output)
	expectedCommands := []string{
		"clone",
		"health",
		"init",
		"pr",
		"rm",
		"run",
		"version",
	}

	for _, command := range expectedCommands {
		if !strings.Contains(outputStr, command) {
			t.Errorf("help output should list command '%s', got: %s", command, outputStr)
		}
	}
}

func TestCLIHealthComplexityCommand(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Create a simple Go repository for complexity analysis
	repoDir := filepath.Join(tmpDir, "complexity-test-repo")
	_ = os.MkdirAll(repoDir, 0755)

	// Create a simple Go file
	goContent := `package main

func simpleFunction() {
	println("Hello, World!")
}

func complexFunction(x int) int {
	if x > 10 {
		if x > 20 {
			return x * 2
		}
		return x + 5
	}
	return x
}
`
	_ = os.WriteFile(filepath.Join(repoDir, "main.go"), []byte(goContent), 0644)

	// Create config
	configContent := `repositories:
  - name: complexity-test-repo
    url: git@github.com:owner/complexity-test-repo.git
    tags: [test]
    path: ` + repoDir

	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config.yaml: %v", err)
	}

	// Run complexity analysis
	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "health", "complexity")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("health complexity command failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Running cyclomatic complexity analysis") {
		t.Errorf("complexity output should contain complexity analysis message, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Maximum complexity threshold") {
		t.Errorf("complexity output should contain complexity threshold message, got: %s", outputStr)
	}
}

func TestCLIHealthDryRunCommand(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Create minimal config
	configContent := `repositories:
  - name: dryrun-test-repo
    url: git@github.com:owner/dryrun-test-repo.git
    tags: [test]
    path: ` + filepath.Join(tmpDir, "dryrun-test-repo")

	_ = os.MkdirAll(filepath.Join(tmpDir, "dryrun-test-repo"), 0755)
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config.yaml: %v", err)
	}

	// Run dry run
	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "health", "--dry-run")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("health --dry-run command failed: %v, output: %s", err, string(output))
	}

	outputStr := string(output)
	// Dry run should show what would be executed
	if !strings.Contains(outputStr, "DRY RUN MODE") || !strings.Contains(outputStr, "Configuration Preview") {
		t.Errorf("dry run output should contain repository name, got: %s", outputStr)
	}
}

func TestCLIHealthGenConfigCommand(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Run genconfig command
	cmd := exec.Command(filepath.Join(originalDir, "repos-test"), "health", "genconfig")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("health genconfig command failed: %v, output: %s", err, string(output))
	}

	// Check if config file was generated
	if _, err := os.Stat("health-config.yaml"); os.IsNotExist(err) {
		t.Error("health-config.yaml should have been generated")
	}

	// Check content
	content, err := os.ReadFile("health-config.yaml")
	if err != nil {
		t.Fatalf("failed to read generated config: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "checkers") {
		t.Errorf("generated config should contain 'checkers', got: %s", contentStr)
	}
}
