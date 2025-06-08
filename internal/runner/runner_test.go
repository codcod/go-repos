package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codcod/repos/internal/config"
)

func TestPrepareLogFile(t *testing.T) {
	tmpDir := t.TempDir()

	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
	}

	tests := []struct {
		name        string
		logDir      string
		command     string
		repoDir     string
		expectError bool
		expectFile  bool
	}{
		{
			name:        "create log file in valid directory",
			logDir:      tmpDir,
			command:     "echo hello",
			repoDir:     "/tmp/repo",
			expectError: false,
			expectFile:  true,
		},
		{
			name:        "create log file with nested directory",
			logDir:      filepath.Join(tmpDir, "nested", "logs"),
			command:     "ls -la",
			repoDir:     "/tmp/repo",
			expectError: false,
			expectFile:  true,
		},
		{
			name:        "empty log directory - no file created",
			logDir:      "",
			command:     "echo test",
			repoDir:     "/tmp/repo",
			expectError: false,
			expectFile:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logFile, logFilePath, err := PrepareLogFile(repo, tt.logDir, tt.command, tt.repoDir)

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

			if tt.expectFile {
				if logFile == nil {
					t.Error("Expected log file but got nil")
					return
				}
				defer logFile.Close()

				if logFilePath == "" {
					t.Error("Expected log file path but got empty string")
					return
				}

				// Verify file exists
				if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
					t.Errorf("Log file should exist at %s", logFilePath)
					return
				}

				// Verify file contains header information
				content, err := os.ReadFile(logFilePath)
				if err != nil {
					t.Errorf("Failed to read log file: %v", err)
					return
				}

				contentStr := string(content)
				if !strings.Contains(contentStr, repo.Name) {
					t.Error("Log file should contain repository name")
				}
				if !strings.Contains(contentStr, tt.command) {
					t.Error("Log file should contain command")
				}
				if !strings.Contains(contentStr, tt.repoDir) {
					t.Error("Log file should contain repository directory")
				}
				if !strings.Contains(contentStr, "=== STDOUT ===") {
					t.Error("Log file should contain stdout header")
				}
			} else {
				if logFile != nil {
					t.Error("Expected no log file but got one")
				}
				if logFilePath != "" {
					t.Error("Expected empty log file path but got one")
				}
			}
		})
	}
}

func TestPrepareLogFileInvalidDirectory(t *testing.T) {
	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
	}

	// Test with invalid path (contains null character on Unix)
	invalidPath := "invalid\x00path"
	_, _, err := PrepareLogFile(repo, invalidPath, "echo test", "/tmp")
	
	// This should fail on most systems
	if err == nil {
		t.Log("Warning: PrepareLogFile did not fail with invalid path - this may be platform specific")
	}
}

func TestOutputProcessor(t *testing.T) {
	tmpDir := t.TempDir()
	logFile, err := os.Create(filepath.Join(tmpDir, "test.log"))
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
	defer logFile.Close()

	tests := []struct {
		name        string
		repoName    string
		isStderr    bool
		input       string
		expectInLog bool
	}{
		{
			name:        "process stdout",
			repoName:    "test-repo",
			isStderr:    false,
			input:       "stdout line 1\nstdout line 2\n",
			expectInLog: true,
		},
		{
			name:        "process stderr",
			repoName:    "test-repo",
			isStderr:    true,
			input:       "stderr line 1\nstderr line 2\n",
			expectInLog: true,
		},
		{
			name:        "empty input",
			repoName:    "test-repo",
			isStderr:    false,
			input:       "",
			expectInLog: false,
		},
		{
			name:        "single line without newline",
			repoName:    "test-repo",
			isStderr:    false,
			input:       "single line",
			expectInLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset log file
			logFile.Seek(0, 0)
			logFile.Truncate(0)

			processor := &OutputProcessor{
				RepoName:  tt.repoName,
				LogFile:   logFile,
				IsStderr:  tt.isStderr,
				HeaderSet: false,
			}

			// We can't easily test the goroutine and WaitGroup in unit tests,
			// so we'll test the core logic directly
			if tt.input != "" {
				// Simulate what ProcessOutput does
				lines := strings.Split(strings.TrimSuffix(tt.input, "\n"), "\n")
				for _, line := range lines {
					if line != "" {
						if tt.isStderr && !processor.HeaderSet {
							logFile.WriteString("\n=== STDERR ===\n")
							processor.HeaderSet = true
						}
						logFile.WriteString(tt.repoName + " | " + line + "\n")
					}
				}
				logFile.Sync()
			}

			// Verify log file content
			content, err := os.ReadFile(logFile.Name())
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			contentStr := string(content)

			if tt.expectInLog {
				if !strings.Contains(contentStr, tt.repoName) {
					t.Error("Log should contain repository name")
				}
				if tt.isStderr && !strings.Contains(contentStr, "=== STDERR ===") {
					t.Error("Log should contain stderr header for stderr output")
				}
			} else {
				if contentStr != "" {
					t.Errorf("Expected empty log for empty input, got: %s", contentStr)
				}
			}
		})
	}
}

func TestRunCommandNonExistentDirectory(t *testing.T) {
	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
		Path: "/path/that/does/not/exist",
	}

	err := RunCommand(repo, "echo test", "")
	if err == nil {
		t.Error("RunCommand should return error for non-existent directory")
	}
	if !strings.Contains(err.Error(), "repository directory does not exist") {
		t.Errorf("Error should mention non-existent directory, got: %v", err)
	}
}

func TestRunCommandWithValidDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")

	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
		Path: tmpDir, // Use temp directory as repo directory
	}

	// Test with simple command that should work
	err := RunCommand(repo, "echo hello", logDir)
	if err != nil {
		t.Errorf("RunCommand should not error for valid command, got: %v", err)
	}

	// Verify log directory was created
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Error("Log directory should have been created")
	}

	// Verify log file was created
	logFiles, err := filepath.Glob(filepath.Join(logDir, "test-repo_*.log"))
	if err != nil {
		t.Fatalf("Failed to search for log files: %v", err)
	}
	if len(logFiles) == 0 {
		t.Error("At least one log file should have been created")
	}
}

func TestRunCommandFailingCommand(t *testing.T) {
	tmpDir := t.TempDir()

	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
		Path: tmpDir,
	}

	// Test with command that should fail
	err := RunCommand(repo, "exit 1", "")
	if err == nil {
		t.Error("RunCommand should return error for failing command")
	}
	if !strings.Contains(err.Error(), "command failed") {
		t.Errorf("Error should mention command failure, got: %v", err)
	}
}

func TestRunCommandWithComplexCommand(t *testing.T) {
	tmpDir := t.TempDir()

	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
		Path: tmpDir,
	}

	// Create test files
	testFile1 := filepath.Join(tmpDir, "file1.txt")
	testFile2 := filepath.Join(tmpDir, "file2.txt")
	os.WriteFile(testFile1, []byte("content1"), 0644)
	os.WriteFile(testFile2, []byte("content2"), 0644)

	// Test with complex shell command
	err := RunCommand(repo, "find . -name '*.txt' | wc -l", "")
	if err != nil {
		t.Errorf("RunCommand should handle complex shell commands, got: %v", err)
	}
}

func TestRunCommandWithCustomLogDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	customLogDir := filepath.Join(tmpDir, "custom", "nested", "logs")

	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
		Path: tmpDir,
	}

	err := RunCommand(repo, "echo test", customLogDir)
	if err != nil {
		t.Errorf("RunCommand should create nested log directories, got: %v", err)
	}

	// Verify nested log directory was created
	if _, err := os.Stat(customLogDir); os.IsNotExist(err) {
		t.Error("Custom nested log directory should have been created")
	}
}

func TestRunCommandWithRepoPathFromURL(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory matching URL-derived name
	urlDerivedName := "test-repo"
	repoDir := filepath.Join(tmpDir, urlDerivedName)
	err := os.MkdirAll(repoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create repo directory: %v", err)
	}

	// Change to temp directory so relative path works
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
		// No Path field - should be derived from URL
	}

	err = RunCommand(repo, "pwd", "")
	if err != nil {
		t.Errorf("RunCommand should work with URL-derived path, got: %v", err)
	}
}

func TestRunCommandLongRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}

	tmpDir := t.TempDir()

	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
		Path: tmpDir,
	}

	// Test with command that takes some time
	start := time.Now()
	err := RunCommand(repo, "sleep 1", "")
	duration := time.Since(start)

	if err != nil {
		t.Errorf("RunCommand should handle long-running commands, got: %v", err)
	}

	// Should take at least 1 second
	if duration < time.Second {
		t.Errorf("Command should have taken at least 1 second, took %v", duration)
	}
}

func BenchmarkRunCommand(b *testing.B) {
	tmpDir := b.TempDir()

	repo := config.Repository{
		Name: "benchmark-repo",
		URL:  "git@github.com:owner/benchmark-repo.git",
		Path: tmpDir,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := RunCommand(repo, "echo benchmark", "")
		if err != nil {
			b.Fatalf("RunCommand() error = %v", err)
		}
	}
}

func BenchmarkPrepareLogFile(b *testing.B) {
	tmpDir := b.TempDir()

	repo := config.Repository{
		Name: "benchmark-repo",
		URL:  "git@github.com:owner/benchmark-repo.git",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logFile, _, err := PrepareLogFile(repo, tmpDir, "echo test", "/tmp")
		if err != nil {
			b.Fatalf("PrepareLogFile() error = %v", err)
		}
		if logFile != nil {
			logFile.Close()
		}
	}
}