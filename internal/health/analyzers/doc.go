/*
Package analyzers provides language-specific static code analysis capabilities.

This package contains analyzers for various programming languages that examine
source code and extract metrics, patterns, and quality indicators.

# Supported Languages

  - Go: Complexity analysis, function detection, package structure
  - Java: Class analysis, dependency detection, complexity metrics
  - JavaScript/TypeScript: Module analysis, complexity measurement
  - Python: Function analysis, import detection, complexity assessment

# Architecture

Each analyzer implements the core.Analyzer interface and integrates with
the analyzer registry for automated discovery and execution.

The registry provides:
  - Automatic analyzer discovery and registration
  - Language detection and analyzer selection
  - Configuration management for language-specific settings
  - Coordinated execution across multiple languages

# Usage

Basic usage through the analyzer registry:

	registry := analyzers.NewRegistry()
	analyzer := registry.GetAnalyzer("go")
	result, err := analyzer.Analyze(ctx, repoPath, config)

Integration with the health package:

	registry := health.NewAnalyzerRegistry(fs, logger)
	result, err := registry.AnalyzeRepository(ctx, repo)

# Extension

To add support for new languages:

1. Implement the core.Analyzer interface
2. Register the analyzer with the registry
3. Add language detection logic
4. Configure file extension mappings

Example analyzer implementation:

	type CustomAnalyzer struct {
		name string
		lang string
	}

	func (a *CustomAnalyzer) Language() string { return a.lang }
	func (a *CustomAnalyzer) CanAnalyze(repo core.Repository) bool {
		// Language detection logic
	}
	func (a *CustomAnalyzer) Analyze(ctx context.Context, repoPath string, config core.AnalyzerConfig) (*core.AnalysisResult, error) {
		// Analysis implementation
	}
*/
package analyzers
