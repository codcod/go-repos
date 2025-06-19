package health

import (
	"context"
	"time"

	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/analyzers"
	"github.com/codcod/repos/internal/health/analyzers/common"
	analyzer_registry "github.com/codcod/repos/internal/health/analyzers/registry"
	checker_registry "github.com/codcod/repos/internal/health/checkers/registry"
	"github.com/codcod/repos/internal/health/commands"
	"github.com/codcod/repos/internal/health/filesystem"
	"github.com/codcod/repos/internal/health/orchestration"
	"github.com/codcod/repos/internal/health/reporting"
)

// Re-export key types for cleaner imports
type (
	AnalyzerRegistry = analyzer_registry.Registry
	CheckerRegistry  = checker_registry.CheckerRegistry
	Engine           = orchestration.Engine
	Formatter        = reporting.Formatter
)

// NewAnalyzerMap creates a map of analyzers using the new factory system
// This is the recommended way to get analyzers in the new architecture
func NewAnalyzerMap(logger core.Logger) map[string]common.FullAnalyzer {
	registry := make(map[string]common.FullAnalyzer)

	for _, lang := range analyzers.GetSupportedLanguages() {
		if analyzer, err := analyzers.GetAnalyzer(lang, logger); err == nil {
			registry[lang] = analyzer
		}
	}

	return registry
}

// NewAnalyzerRegistry creates a new analyzer registry using the new factory system
// This method creates a legacy registry for backward compatibility
func NewAnalyzerRegistry(logger core.Logger) *AnalyzerRegistry {
	registry := analyzer_registry.NewRegistry()

	// Register analyzers from the new factory system with adapters
	for _, lang := range analyzers.GetSupportedLanguages() {
		if analyzer, err := analyzers.GetAnalyzer(lang, logger); err == nil {
			adapter := &NewToLegacyAnalyzerAdapter{analyzer: analyzer}
			registry.Register(adapter)
		}
	}

	return registry
}

// NewToLegacyAnalyzerAdapter adapts the new analyzer interface to the legacy core.Analyzer interface
type NewToLegacyAnalyzerAdapter struct {
	analyzer common.FullAnalyzer
}

// Name returns the analyzer name (legacy interface)
func (a *NewToLegacyAnalyzerAdapter) Name() string {
	return a.analyzer.Language() + "-analyzer"
}

// Language returns the programming language
func (a *NewToLegacyAnalyzerAdapter) Language() string {
	return a.analyzer.Language()
}

// SupportedExtensions returns supported file extensions (legacy interface)
func (a *NewToLegacyAnalyzerAdapter) SupportedExtensions() []string {
	return a.analyzer.FileExtensions()
}

// CanAnalyze checks if the analyzer can process the given repository
func (a *NewToLegacyAnalyzerAdapter) CanAnalyze(repo core.Repository) bool {
	return a.analyzer.CanAnalyze(repo)
}

// Analyze performs full analysis and returns results
func (a *NewToLegacyAnalyzerAdapter) Analyze(ctx context.Context, repoPath string, config core.AnalyzerConfig) (*core.AnalysisResult, error) {
	// Use the new analyzer's Analyze method if available, otherwise synthesize from complexity analysis
	if fullAnalyzer, ok := a.analyzer.(interface {
		Analyze(context.Context, string, core.AnalyzerConfig) (*core.AnalysisResult, error)
	}); ok {
		return fullAnalyzer.Analyze(ctx, repoPath, config)
	}

	// Fallback: create a basic analysis result from complexity analysis
	result := &core.AnalysisResult{
		Language:  a.analyzer.Language(),
		Files:     make(map[string]*core.FileAnalysis),
		Functions: []core.FunctionInfo{},
		Metrics:   make(map[string]interface{}),
	}

	if complexityResult, err := a.analyzer.AnalyzeComplexity(ctx, repoPath); err == nil {
		result.Metrics["total_files"] = complexityResult.TotalFiles
		result.Metrics["total_functions"] = complexityResult.TotalFunctions
		result.Metrics["average_complexity"] = complexityResult.AverageComplexity
		result.Metrics["max_complexity"] = complexityResult.MaxComplexity

		// Convert complexity functions to FunctionInfo
		for _, fn := range complexityResult.Functions {
			result.Functions = append(result.Functions, core.FunctionInfo{
				Name:       fn.Name,
				File:       fn.File,
				Line:       fn.Line,
				Complexity: fn.Complexity,
				Language:   a.analyzer.Language(),
			})
		}
	}

	return result, nil
}

// NewAnalyzerRegistryLegacy creates a legacy analyzer registry for backward compatibility
// This function is deprecated - use NewAnalyzerRegistry(logger) instead
func NewAnalyzerRegistryLegacy(fs core.FileSystem, logger core.Logger) *AnalyzerRegistry {
	// Fallback to new registry since legacy implementation was removed
	return NewAnalyzerRegistry(logger)
}

// NewCheckerRegistry creates a new checker registry with command executor
func NewCheckerRegistry(executor commands.CommandExecutor) *CheckerRegistry {
	return checker_registry.NewCheckerRegistry(executor)
}

// NewOrchestrationEngine creates a new orchestration engine
func NewOrchestrationEngine(
	checkerRegistry core.CheckerRegistry,
	analyzerRegistry core.AnalyzerRegistry,
	config core.Config,
	logger core.Logger,
) *Engine {
	return orchestration.NewEngine(checkerRegistry, analyzerRegistry, config, logger)
}

// NewFileSystem creates a new OS filesystem implementation
func NewFileSystem() core.FileSystem {
	return filesystem.NewOSFileSystem()
}

// NewCommandExecutor creates a new OS command executor with timeout
func NewCommandExecutor(timeout time.Duration) commands.CommandExecutor {
	return commands.NewOSCommandExecutor(timeout)
}

// NewFormatter creates a new result formatter
func NewFormatter(verbose bool) *Formatter {
	return reporting.NewFormatter(verbose)
}

// GetExitCode determines the appropriate exit code based on results
func GetExitCode(result core.WorkflowResult) int {
	return reporting.ExitCode(result)
}

// HealthPackage provides a unified interface for all health analysis functionality
type HealthPackage struct {
	AnalyzerRegistry *AnalyzerRegistry
	CheckerRegistry  *CheckerRegistry
	Engine           *Engine
	FileSystem       core.FileSystem
	Logger           core.Logger
}

// NewHealthPackage creates a complete health analysis package with all components
func NewHealthPackage(config core.Config, logger core.Logger, timeout time.Duration) *HealthPackage {
	fs := NewFileSystem()
	executor := NewCommandExecutor(timeout)

	analyzerRegistry := NewAnalyzerRegistry(logger)
	checkerRegistry := NewCheckerRegistry(executor)
	engine := NewOrchestrationEngine(checkerRegistry, analyzerRegistry, config, logger)

	return &HealthPackage{
		AnalyzerRegistry: analyzerRegistry,
		CheckerRegistry:  checkerRegistry,
		Engine:           engine,
		FileSystem:       fs,
		Logger:           logger,
	}
}
