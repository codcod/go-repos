package health

import (
	"path/filepath"
	"strings"
)

// LanguageAnalyzer interface for pluggable language analysis
type LanguageAnalyzer interface {
	Name() string
	FilePatterns() []string
	ExcludePatterns() []string
	SupportsComplexity() bool
	SupportsFunctionLevel() bool
	AnalyzeComplexity(filePath string) (ComplexityResult, error)
	AnalyzeFunctions(filePath string) ([]FunctionComplexity, error)
}

// BaseAnalyzer provides common functionality for language analyzers
type BaseAnalyzer struct {
	name            string
	filePatterns    []string
	excludePatterns []string
	logger          Logger
}

// NewBaseAnalyzer creates a new base analyzer
func NewBaseAnalyzer(name string, patterns, excludes []string, logger Logger) *BaseAnalyzer {
	if logger == nil {
		logger = &NoOpLogger{}
	}

	return &BaseAnalyzer{
		name:            name,
		filePatterns:    patterns,
		excludePatterns: excludes,
		logger:          logger,
	}
}

func (b *BaseAnalyzer) Name() string {
	return b.name
}

func (b *BaseAnalyzer) FilePatterns() []string {
	return b.filePatterns
}

func (b *BaseAnalyzer) ExcludePatterns() []string {
	return b.excludePatterns
}

// ShouldAnalyzeFile checks if a file should be analyzed based on patterns
func (b *BaseAnalyzer) ShouldAnalyzeFile(filePath string) bool {
	// Check if file matches any include patterns
	matches := false
	for _, pattern := range b.filePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			matches = true
			break
		}
	}

	if !matches {
		return false
	}

	// Check if file matches any exclude patterns
	for _, pattern := range b.excludePatterns {
		if strings.Contains(filePath, pattern) {
			return false
		}
	}

	return true
}

// AnalyzerRegistry manages language analyzers
type AnalyzerRegistry struct {
	analyzers map[string]LanguageAnalyzer
	logger    Logger
}

// NewAnalyzerRegistry creates a new analyzer registry
func NewAnalyzerRegistry(logger Logger) *AnalyzerRegistry {
	if logger == nil {
		logger = &NoOpLogger{}
	}

	return &AnalyzerRegistry{
		analyzers: make(map[string]LanguageAnalyzer),
		logger:    logger,
	}
}

// Register registers a language analyzer
func (r *AnalyzerRegistry) Register(analyzer LanguageAnalyzer) {
	r.analyzers[analyzer.Name()] = analyzer
	r.logger.Info("registered language analyzer", String("language", analyzer.Name()))
}

// GetByName returns an analyzer by name
func (r *AnalyzerRegistry) GetByName(name string) (LanguageAnalyzer, bool) {
	analyzer, exists := r.analyzers[name]
	return analyzer, exists
}

// GetByExtension returns an analyzer that can handle the given file extension
func (r *AnalyzerRegistry) GetByExtension(ext string) LanguageAnalyzer {
	for _, analyzer := range r.analyzers {
		for _, pattern := range analyzer.FilePatterns() {
			if matched, _ := filepath.Match(pattern, "*"+ext); matched {
				return analyzer
			}
		}
	}
	return nil
}

// GetByFilePath returns an analyzer that can handle the given file path
func (r *AnalyzerRegistry) GetByFilePath(filePath string) LanguageAnalyzer {
	for _, analyzer := range r.analyzers {
		if r.shouldAnalyzeWithAnalyzer(analyzer, filePath) {
			return analyzer
		}
	}
	return nil
}

// ListAnalyzers returns all registered analyzers
func (r *AnalyzerRegistry) ListAnalyzers() map[string]LanguageAnalyzer {
	return r.analyzers
}

// shouldAnalyzeWithAnalyzer checks if a file should be analyzed with a specific analyzer
func (r *AnalyzerRegistry) shouldAnalyzeWithAnalyzer(analyzer LanguageAnalyzer, filePath string) bool {
	// Check if file matches any include patterns
	matches := false
	for _, pattern := range analyzer.FilePatterns() {
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			matches = true
			break
		}
	}

	if !matches {
		return false
	}

	// Check if file matches any exclude patterns
	for _, pattern := range analyzer.ExcludePatterns() {
		if strings.Contains(filePath, pattern) {
			return false
		}
	}

	return true
}
