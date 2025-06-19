// Package common provides base analyzer implementation
package common

import (
	"os"
	"path/filepath"

	"github.com/codcod/repos/internal/core"
)

// BaseAnalyzerImpl provides common functionality for all analyzers
type BaseAnalyzerImpl struct {
	name       string
	language   string
	extensions []string
	excludes   []string
	walker     FileWalker
	logger     core.Logger
}

// NewBaseAnalyzer creates a new base analyzer
func NewBaseAnalyzer(name, language string, extensions, excludes []string, walker FileWalker, logger core.Logger) *BaseAnalyzerImpl {
	return &BaseAnalyzerImpl{
		name:       name,
		language:   language,
		extensions: extensions,
		excludes:   excludes,
		walker:     walker,
		logger:     logger,
	}
}

// Language returns the programming language name
func (b *BaseAnalyzerImpl) Language() string {
	return b.language
}

// FileExtensions returns supported file extensions
func (b *BaseAnalyzerImpl) FileExtensions() []string {
	return b.extensions
}

// CanAnalyze checks if the analyzer can process the given repository
func (b *BaseAnalyzerImpl) CanAnalyze(repo core.Repository) bool {
	// Basic implementation - can be overridden by specific analyzers
	return b.hasFiles(repo.Path)
}

// hasFiles checks if the repository contains files with supported extensions
func (b *BaseAnalyzerImpl) hasFiles(repoPath string) bool {
	files, err := b.walker.FindFiles(repoPath, b.extensions, b.excludes)
	return err == nil && len(files) > 0
}

// FindFiles finds all files with supported extensions in the repository
func (b *BaseAnalyzerImpl) FindFiles(repoPath string) ([]string, error) {
	return b.walker.FindFiles(repoPath, b.extensions, b.excludes)
}

// ReadFile reads the content of a file
func (b *BaseAnalyzerImpl) ReadFile(filePath string) ([]byte, error) {
	return b.walker.ReadFile(filePath)
}

// Logger returns the logger instance
func (b *BaseAnalyzerImpl) Logger() core.Logger {
	return b.logger
}

// DefaultFileWalker provides a default file system implementation
type DefaultFileWalker struct{}

// NewDefaultFileWalker creates a new default file walker
func NewDefaultFileWalker() *DefaultFileWalker {
	return &DefaultFileWalker{}
}

// FindFiles finds all files with given extensions in a directory, excluding patterns
func (w *DefaultFileWalker) FindFiles(rootPath string, extensions []string, excludePatterns []string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories and common excludes
			name := info.Name()
			if name[0] == '.' || name == "node_modules" || name == "vendor" || name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be excluded
		if ShouldExcludeFile(path, excludePatterns) {
			return nil
		}

		// Check if file has supported extension
		ext := filepath.Ext(path)
		for _, supportedExt := range extensions {
			if ext == supportedExt {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	return files, err
}

// ReadFile reads the content of a file
func (w *DefaultFileWalker) ReadFile(filePath string) ([]byte, error) {
	// Validate file path to prevent potential security issues
	if filePath == "" {
		return nil, &AnalyzerError{
			Type:    "invalid_path",
			Message: "file path cannot be empty",
		}
	}

	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(filePath)
	return os.ReadFile(cleanPath) // #nosec G304 - path is cleaned and validated
}
