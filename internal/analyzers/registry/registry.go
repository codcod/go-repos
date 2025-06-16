package registry

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	golang "github.com/codcod/repos/internal/analyzers/go"
	java_analyzer "github.com/codcod/repos/internal/analyzers/java"
	javascript_analyzer "github.com/codcod/repos/internal/analyzers/javascript"
	python_analyzer "github.com/codcod/repos/internal/analyzers/python"
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

// GetByLanguage gets an analyzer by language
func (r *Registry) GetByLanguage(language string) (core.Analyzer, bool) {
	analyzer, exists := r.analyzers[language]
	return analyzer, exists
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

// detectLanguages detects programming languages in a repository
func (r *Registry) detectLanguages(repoPath string) []string {
	languageMap := make(map[string]bool)

	// Walk through files and detect languages based on extensions
	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		if info.IsDir() {
			// Skip common directories that shouldn't be analyzed
			name := filepath.Base(path)
			skipDirs := []string{
				".git", ".svn", ".hg",
				"node_modules", "vendor", "target", "build", "dist",
				".venv", "venv", "env", "__pycache__",
				".gradle", ".next", ".nuxt",
			}

			for _, skipDir := range skipDirs {
				if name == skipDir {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if analyzer, exists := r.GetByFileExtension(ext); exists {
			languageMap[analyzer.Language()] = true
		}

		return nil
	})

	// Convert map to slice
	var languages []string
	for lang := range languageMap {
		languages = append(languages, lang)
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

// NewRegistryWithStandardAnalyzers creates a registry with all standard language analyzers
func NewRegistryWithStandardAnalyzers(fs core.FileSystem, logger core.Logger) *Registry {
	registry := NewRegistry()

	// Register all standard language analyzers
	registry.Register(golang.NewGoAnalyzer(fs, logger))
	registry.Register(python_analyzer.NewPythonAnalyzer(fs, logger))
	registry.Register(java_analyzer.NewJavaAnalyzer(fs, logger))
	registry.Register(javascript_analyzer.NewJavaScriptAnalyzer(fs, logger))

	return registry
}

// RegisterStandardAnalyzers registers all standard analyzers to an existing registry
func (r *Registry) RegisterStandardAnalyzers(fs core.FileSystem, logger core.Logger) {
	r.Register(golang.NewGoAnalyzer(fs, logger))
	r.Register(python_analyzer.NewPythonAnalyzer(fs, logger))
	r.Register(java_analyzer.NewJavaAnalyzer(fs, logger))
	r.Register(javascript_analyzer.NewJavaScriptAnalyzer(fs, logger))
}
