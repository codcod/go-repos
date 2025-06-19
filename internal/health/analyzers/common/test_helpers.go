// Package common provides shared test helpers for health analyzers
package common

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/codcod/repos/internal/core"
)

// MockFileWalker provides a mock file system for testing
type MockFileWalker struct {
	Files map[string][]byte
	Dirs  map[string][]string
}

// NewMockFileWalker creates a new mock file walker
func NewMockFileWalker() *MockFileWalker {
	return &MockFileWalker{
		Files: make(map[string][]byte),
		Dirs:  make(map[string][]string),
	}
}

// AddFile adds a file to the mock file system
func (m *MockFileWalker) AddFile(path string, content []byte) {
	m.Files[path] = content

	// Add to directory listing
	dir := filepath.Dir(path)
	if _, exists := m.Dirs[dir]; !exists {
		m.Dirs[dir] = []string{}
	}
	m.Dirs[dir] = append(m.Dirs[dir], path)
}

// FindFiles finds all files with given extensions in a directory, excluding patterns
func (m *MockFileWalker) FindFiles(rootPath string, extensions []string, excludePatterns []string) ([]string, error) {
	var files []string

	for filePath := range m.Files {
		// Check if file is under root path using proper path checking
		cleanFilePath := filepath.Clean(filePath)
		cleanRootPath := filepath.Clean(rootPath)

		// Use filepath.Rel to check if the file is under the root path
		relPath, err := filepath.Rel(cleanRootPath, cleanFilePath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			continue
		}

		// Check if file should be excluded
		if ShouldExcludeFile(filePath, excludePatterns) {
			continue
		}

		// Check if file has supported extension
		ext := filepath.Ext(filePath)
		for _, supportedExt := range extensions {
			if ext == supportedExt {
				files = append(files, filePath)
				break
			}
		}
	}

	return files, nil
}

// ReadFile reads the content of a file from the mock file system
func (m *MockFileWalker) ReadFile(filePath string) ([]byte, error) {
	content, exists := m.Files[filePath]
	if !exists {
		return nil, &AnalyzerError{
			Type:    ErrorTypeFileSystem,
			Message: "file not found",
			Cause:   nil,
		}
	}
	return content, nil
}

// MockLogger provides a simple logger for testing
type MockLogger struct {
	Messages []LogMessage
}

// LogMessage represents a logged message
type LogMessage struct {
	Level  string
	Text   string
	Fields []core.Field
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		Messages: []LogMessage{},
	}
}

// Debug logs a debug message
func (l *MockLogger) Debug(msg string, fields ...core.Field) {
	l.Messages = append(l.Messages, LogMessage{Level: "DEBUG", Text: msg, Fields: fields})
}

// Info logs an info message
func (l *MockLogger) Info(msg string, fields ...core.Field) {
	l.Messages = append(l.Messages, LogMessage{Level: "INFO", Text: msg, Fields: fields})
}

// Warn logs a warning message
func (l *MockLogger) Warn(msg string, fields ...core.Field) {
	l.Messages = append(l.Messages, LogMessage{Level: "WARN", Text: msg, Fields: fields})
}

// Error logs an error message
func (l *MockLogger) Error(msg string, fields ...core.Field) {
	l.Messages = append(l.Messages, LogMessage{Level: "ERROR", Text: msg, Fields: fields})
}

// Fatal logs a fatal message
func (l *MockLogger) Fatal(msg string, fields ...core.Field) {
	l.Messages = append(l.Messages, LogMessage{Level: "FATAL", Text: msg, Fields: fields})
}

// GetLastMessage returns the last logged message or nil if no messages
func (l *MockLogger) GetLastMessage() *LogMessage {
	if len(l.Messages) == 0 {
		return nil
	}
	return &l.Messages[len(l.Messages)-1]
}

// AssertComplexityResult asserts that a complexity result has expected values
func AssertComplexityResult(t *testing.T, result core.ComplexityResult, expectedFiles, expectedFunctions int) {
	t.Helper()

	if result.TotalFiles != expectedFiles {
		t.Errorf("Expected %d files, got %d", expectedFiles, result.TotalFiles)
	}

	if result.TotalFunctions != expectedFunctions {
		t.Errorf("Expected %d functions, got %d", expectedFunctions, result.TotalFunctions)
	}

	if expectedFunctions > 0 && result.AverageComplexity == 0 {
		t.Error("Expected non-zero average complexity")
	}
}

// CreateTestRepository creates a test repository structure for testing
func CreateTestRepository(t *testing.T, files map[string]string) *MockFileWalker {
	t.Helper()

	walker := NewMockFileWalker()
	for path, content := range files {
		walker.AddFile(path, []byte(content))
	}
	return walker
}
