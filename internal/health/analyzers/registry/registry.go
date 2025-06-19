package registry

import (
	"context"
	"fmt"

	"github.com/codcod/repos/internal/core"
)

// Registry manages language analyzers
type Registry struct {
	analyzers map[string]core.Analyzer
}

// NewRegistry creates a new analyzer registry
func NewRegistry() *Registry {
	return &Registry{
		analyzers: make(map[string]core.Analyzer),
	}
}

// Register registers an analyzer
func (r *Registry) Register(analyzer core.Analyzer) {
	r.analyzers[analyzer.Language()] = analyzer
}

// Unregister removes an analyzer by language
func (r *Registry) Unregister(language string) {
	delete(r.analyzers, language)
}

// GetByLanguage gets an analyzer by language
func (r *Registry) GetByLanguage(language string) (core.Analyzer, bool) {
	analyzer, exists := r.analyzers[language]
	return analyzer, exists
}

// GetAnalyzer gets an analyzer by language (core.AnalyzerRegistry interface)
func (r *Registry) GetAnalyzer(language string) (core.Analyzer, error) {
	analyzer, exists := r.analyzers[language]
	if !exists {
		return nil, fmt.Errorf("analyzer not found for language: %s", language)
	}
	return analyzer, nil
}

// GetAnalyzers returns all registered analyzers (core.AnalyzerRegistry interface)
func (r *Registry) GetAnalyzers() []core.Analyzer {
	var analyzers []core.Analyzer
	for _, analyzer := range r.analyzers {
		analyzers = append(analyzers, analyzer)
	}
	return analyzers
}

// GetByFileExtension gets an analyzer by file extension
func (r *Registry) GetByFileExtension(ext string) (core.Analyzer, bool) {
	for _, analyzer := range r.analyzers {
		for _, supportedExt := range analyzer.SupportedExtensions() {
			if ext == supportedExt {
				return analyzer, true
			}
		}
	}
	return nil, false
}

// GetSupportedAnalyzers returns analyzers that support the given repository
func (r *Registry) GetSupportedAnalyzers(repo core.Repository) []core.Analyzer {
	var supported []core.Analyzer

	for _, analyzer := range r.analyzers {
		if analyzer.CanAnalyze(repo) {
			supported = append(supported, analyzer)
		}
	}

	return supported
}

// GetAllAnalyzers returns all registered analyzers
func (r *Registry) GetAllAnalyzers() []core.Analyzer {
	var analyzers []core.Analyzer
	for _, analyzer := range r.analyzers {
		analyzers = append(analyzers, analyzer)
	}
	return analyzers
}

// GetSupportedLanguages returns all supported languages
func (r *Registry) GetSupportedLanguages() []string {
	var languages []string
	for language := range r.analyzers {
		languages = append(languages, language)
	}
	return languages
}

// BaseAnalyzer provides common functionality for analyzers
type BaseAnalyzer struct {
	language      string
	extensions    []string
	complexity    bool
	functionLevel bool
}

// NewBaseAnalyzer creates a new base analyzer
func NewBaseAnalyzer(language string, extensions []string, complexity, functionLevel bool) *BaseAnalyzer {
	return &BaseAnalyzer{
		language:      language,
		extensions:    extensions,
		complexity:    complexity,
		functionLevel: functionLevel,
	}
}

// Language returns the language name
func (a *BaseAnalyzer) Language() string {
	return a.language
}

// FileExtensions returns supported file extensions
func (a *BaseAnalyzer) FileExtensions() []string {
	return a.extensions
}

// SupportsComplexity returns whether complexity analysis is supported
func (a *BaseAnalyzer) SupportsComplexity() bool {
	return a.complexity
}

// SupportsFunctionLevel returns whether function-level analysis is supported
func (a *BaseAnalyzer) SupportsFunctionLevel() bool {
	return a.functionLevel
}

// AnalyzeComplexity provides a default implementation (should be overridden)
func (a *BaseAnalyzer) AnalyzeComplexity(ctx context.Context, repoPath string) (core.ComplexityResult, error) {
	return core.ComplexityResult{}, nil
}

// AnalyzeFunctions provides a default implementation (should be overridden)
func (a *BaseAnalyzer) AnalyzeFunctions(ctx context.Context, repoPath string) ([]core.FunctionComplexity, error) {
	return nil, nil
}

// DetectPatterns provides a default implementation (should be overridden)
func (a *BaseAnalyzer) DetectPatterns(ctx context.Context, content string, patterns []core.Pattern) ([]core.PatternMatch, error) {
	return nil, nil
}

// NOTE: NewRegistryWithStandardAnalyzers has been deprecated
// Use the new factory system in internal/health/analyzers instead
// Example: analyzers.GetAnalyzer("go", logger)
