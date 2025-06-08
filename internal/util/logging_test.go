package util

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/codcod/repos/internal/config"
	"github.com/fatih/color"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Error("NewLogger() should not return nil")
	}

	if logger.repoColor == nil {
		t.Error("NewLogger() should initialize repoColor function")
	}
}

func TestLoggerSuccess(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture stdout using color package's output
	var buf bytes.Buffer
	color.Output = &buf

	logger := NewLogger()
	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
	}

	logger.Success(repo, "Test success message")

	// Restore color output
	color.Output = os.Stdout

	output := buf.String()

	if !strings.Contains(output, "test-repo") {
		t.Errorf("Success output should contain repo name, got: %s", output)
	}
	if !strings.Contains(output, "Test success message") {
		t.Errorf("Success output should contain message, got: %s", output)
	}
}

func TestLoggerError(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture stdout using color package's output (color.Red also writes to color.Output)
	var buf bytes.Buffer
	color.Output = &buf

	logger := NewLogger()
	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
	}

	logger.Error(repo, "Test error message")

	// Restore color output
	color.Output = os.Stdout

	output := buf.String()

	if !strings.Contains(output, "test-repo") {
		t.Errorf("Error output should contain repo name, got: %s", output)
	}
	if !strings.Contains(output, "Test error message") {
		t.Errorf("Error output should contain message, got: %s", output)
	}
}

func TestLoggerInfo(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture stdout using color package's output
	var buf bytes.Buffer
	color.Output = &buf

	logger := NewLogger()
	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
	}

	logger.Info(repo, "Test info message")

	// Restore color output
	color.Output = os.Stdout

	output := buf.String()

	if !strings.Contains(output, "test-repo") {
		t.Errorf("Info output should contain repo name, got: %s", output)
	}
	if !strings.Contains(output, "Test info message") {
		t.Errorf("Info output should contain message, got: %s", output)
	}
}

func TestLoggerWarn(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture stdout using color package's output
	var buf bytes.Buffer
	color.Output = &buf

	logger := NewLogger()
	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
	}

	logger.Warn(repo, "Test warning message")

	// Restore color output
	color.Output = os.Stdout

	output := buf.String()

	if !strings.Contains(output, "test-repo") {
		t.Errorf("Warn output should contain repo name, got: %s", output)
	}
	if !strings.Contains(output, "Test warning message") {
		t.Errorf("Warn output should contain message, got: %s", output)
	}
}

func TestLoggerWithFormatting(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture stdout using color package's output
	var buf bytes.Buffer
	color.Output = &buf

	logger := NewLogger()
	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
	}

	logger.Success(repo, "Test message with %s and %d", "string", 42)

	// Restore color output
	color.Output = os.Stdout

	output := buf.String()

	if !strings.Contains(output, "test-repo") {
		t.Errorf("Output should contain repo name, got: %s", output)
	}
	if !strings.Contains(output, "string") {
		t.Errorf("Output should contain formatted string, got: %s", output)
	}
	if !strings.Contains(output, "42") {
		t.Errorf("Output should contain formatted number, got: %s", output)
	}
}

func TestLoggerWithEmptyRepoName(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture stdout using color package's output
	var buf bytes.Buffer
	color.Output = &buf

	logger := NewLogger()
	repo := config.Repository{
		Name: "",
		URL:  "git@github.com:owner/test-repo.git",
	}

	logger.Info(repo, "Test message")

	// Restore color output
	color.Output = os.Stdout

	output := buf.String()

	// Should still work with empty repo name
	if !strings.Contains(output, "Test message") {
		t.Errorf("Output should contain message even with empty repo name, got: %s", output)
	}
}

func TestLoggerWithSpecialCharacters(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Capture stdout using color package's output
	var buf bytes.Buffer
	color.Output = &buf

	logger := NewLogger()
	repo := config.Repository{
		Name: "repo-with-special-chars_123",
		URL:  "git@github.com:owner/repo-with-special-chars_123.git",
	}

	logger.Info(repo, "Message with special chars: !@#$%%%%^&*()")

	// Restore color output
	color.Output = os.Stdout

	output := buf.String()

	if !strings.Contains(output, "repo-with-special-chars_123") {
		t.Errorf("Output should contain repo name with special chars, got: %s", output)
	}
	if !strings.Contains(output, "!@#$%%^&*()") {
		t.Errorf("Output should contain special characters, got: %s", output)
	}
}

func TestLoggerRepoColorFunction(t *testing.T) {
	logger := NewLogger()
	
	// Test that repoColor function works
	testString := "test"
	coloredString := logger.repoColor(testString)
	
	// When colors are enabled, the string should be different (contain ANSI codes)
	// When colors are disabled, it should be the same
	if color.NoColor {
		if coloredString != testString {
			t.Error("repoColor should not modify string when colors are disabled")
		}
	} else {
		// Should contain the original string somewhere
		if !strings.Contains(coloredString, testString) {
			t.Error("repoColor should preserve the original string content")
		}
	}
}

func TestLoggerMultipleRepos(t *testing.T) {
	// Disable color for consistent testing
	color.NoColor = true
	defer func() { color.NoColor = false }()

	// Test logging with multiple different repositories
	repos := []config.Repository{
		{Name: "repo1", URL: "git@github.com:owner/repo1.git"},
		{Name: "repo2", URL: "git@github.com:owner/repo2.git"},
		{Name: "repo3", URL: "git@github.com:owner/repo3.git"},
	}
	
	// Capture stdout using color package's output
	var buf bytes.Buffer
	color.Output = &buf

	logger := NewLogger()
	
	for i, repo := range repos {
		logger.Info(repo, "Message %d", i+1)
	}

	// Restore color output
	color.Output = os.Stdout

	output := buf.String()

	// Check that all repo names appear in output
	for _, repo := range repos {
		if !strings.Contains(output, repo.Name) {
			t.Errorf("Output should contain repo name %s, got: %s", repo.Name, output)
		}
	}
	
	// Check that all messages appear
	for i := 1; i <= 3; i++ {
		expectedMsg := fmt.Sprintf("Message %d", i)
		if !strings.Contains(output, expectedMsg) {
			t.Errorf("Output should contain %s, got: %s", expectedMsg, output)
		}
	}
}

func TestLoggerColorDisabled(t *testing.T) {
	// Disable color for this test
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = oldNoColor }()

	// Capture stdout using color package's output
	var buf bytes.Buffer
	color.Output = &buf

	logger := NewLogger()
	repo := config.Repository{
		Name: "test-repo",
		URL:  "git@github.com:owner/test-repo.git",
	}

	logger.Success(repo, "Test message")

	// Restore color output
	color.Output = os.Stdout

	output := buf.String()

	// Should still contain the repo name and message
	if !strings.Contains(output, "test-repo") {
		t.Errorf("Output should contain repo name even with colors disabled, got: %s", output)
	}
	if !strings.Contains(output, "Test message") {
		t.Errorf("Output should contain message even with colors disabled, got: %s", output)
	}
}

func BenchmarkLoggerSuccess(b *testing.B) {
	// Redirect output to discard for benchmarking
	var buf bytes.Buffer
	color.Output = &buf
	defer func() { color.Output = os.Stdout }()

	logger := NewLogger()
	repo := config.Repository{
		Name: "benchmark-repo",
		URL:  "git@github.com:owner/benchmark-repo.git",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Success(repo, "Benchmark message %d", i)
	}
}

func BenchmarkLoggerError(b *testing.B) {
	// Redirect output to discard for benchmarking
	var buf bytes.Buffer
	color.Output = &buf
	defer func() { color.Output = os.Stdout }()

	logger := NewLogger()
	repo := config.Repository{
		Name: "benchmark-repo",
		URL:  "git@github.com:owner/benchmark-repo.git",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Error(repo, "Benchmark error %d", i)
	}
}