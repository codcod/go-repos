// Package common provides shared types and utilities for health analyzers
package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/codcod/repos/internal/core"
)

// AnalyzerError represents standardized error types for analyzers
type AnalyzerError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// ErrorType represents different categories of analyzer errors
type ErrorType string

const (
	ErrorTypeFileSystem  ErrorType = "filesystem"
	ErrorTypeParsing     ErrorType = "parsing"
	ErrorTypeUnsupported ErrorType = "unsupported"
	ErrorTypeInternal    ErrorType = "internal"
)

// Error implements the error interface
func (e *AnalyzerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// NewAnalyzerError creates a new standardized analyzer error
func NewAnalyzerError(errorType ErrorType, message string, cause error) *AnalyzerError {
	return &AnalyzerError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// FileWalker provides file system operations for analyzers
type FileWalker interface {
	// FindFiles finds all files with given extensions in a directory, excluding patterns
	FindFiles(rootPath string, extensions []string, excludePatterns []string) ([]string, error)

	// ReadFile reads the content of a file
	ReadFile(filePath string) ([]byte, error)
}

// MakeRelativePath converts an absolute path to relative path within repository
func MakeRelativePath(repoPath, filePath string) string {
	relativePath, err := filepath.Rel(repoPath, filePath)
	if err != nil {
		// If we can't make it relative, use the original path
		return filePath
	}
	return relativePath
}

// FilterFilesByExtensions filters files by their extensions
func FilterFilesByExtensions(files []string, extensions []string) []string {
	var filtered []string
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[ext] = true
	}

	for _, file := range files {
		if extMap[filepath.Ext(file)] {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// ShouldExcludeFile checks if a file should be excluded based on patterns
func ShouldExcludeFile(filePath string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if strings.Contains(filePath, pattern) {
			return true
		}
	}
	return false
}

// CalculateAverageComplexity calculates average complexity from function complexities
func CalculateAverageComplexity(functions []core.FunctionComplexity) float64 {
	if len(functions) == 0 {
		return 0.0
	}

	total := 0
	for _, fn := range functions {
		total += fn.Complexity
	}

	return float64(total) / float64(len(functions))
}

// FindMaxComplexity finds the maximum complexity from function complexities
func FindMaxComplexity(functions []core.FunctionComplexity) int {
	maxComplexity := 0
	for _, fn := range functions {
		if fn.Complexity > maxComplexity {
			maxComplexity = fn.Complexity
		}
	}
	return maxComplexity
}
