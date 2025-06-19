package health

import (
	"context"
	"time"

	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/analyzers"
	"github.com/codcod/repos/internal/health/analyzers/common"
	checker_registry "github.com/codcod/repos/internal/health/checkers/registry"
	"github.com/codcod/repos/internal/health/commands"
	"github.com/codcod/repos/internal/health/filesystem"
	"github.com/codcod/repos/internal/health/orchestration"
	"github.com/codcod/repos/internal/health/reporting"
)

// Re-export key types for cleaner imports
type (
	CheckerRegistry = checker_registry.CheckerRegistry
	Engine          = orchestration.Engine
	Formatter       = reporting.Formatter
)

// NewAnalyzerRegistry creates a new analyzer registry using the new factory system
func NewAnalyzerRegistry(logger core.Logger) core.AnalyzerRegistry {
	return &ModernAnalyzerRegistry{logger: logger}
}

// ModernAnalyzerRegistry implements core.AnalyzerRegistry using the new factory system
type ModernAnalyzerRegistry struct {
	logger core.Logger
}

// Register is not supported in the new factory system (analyzers are auto-registered)
func (r *ModernAnalyzerRegistry) Register(analyzer core.Analyzer) {
	// No-op: analyzers are automatically registered via the factory system
}

// Unregister is not supported in the new factory system
func (r *ModernAnalyzerRegistry) Unregister(language string) {
	// No-op: analyzers are automatically managed via the factory system
}

// GetAnalyzer gets an analyzer for the specified language
func (r *ModernAnalyzerRegistry) GetAnalyzer(language string) (core.Analyzer, error) {
	analyzer, err := analyzers.GetAnalyzer(language, r.logger)
	if err != nil {
		return nil, err
	}
	return &ModernAnalyzerAdapter{analyzer: analyzer}, nil
}

// GetAnalyzers returns all available analyzers
func (r *ModernAnalyzerRegistry) GetAnalyzers() []core.Analyzer {
	var result []core.Analyzer
	for _, lang := range analyzers.GetSupportedLanguages() {
		if analyzer, err := analyzers.GetAnalyzer(lang, r.logger); err == nil {
			result = append(result, &ModernAnalyzerAdapter{analyzer: analyzer})
		}
	}
	return result
}

// GetSupportedLanguages returns all supported languages
func (r *ModernAnalyzerRegistry) GetSupportedLanguages() []string {
	return analyzers.GetSupportedLanguages()
}

// ModernAnalyzerAdapter adapts the new analyzer interface to the core.Analyzer interface
type ModernAnalyzerAdapter struct {
	analyzer common.FullAnalyzer
}

// Name returns the analyzer name
func (a *ModernAnalyzerAdapter) Name() string {
	return a.analyzer.Language() + "-analyzer"
}

// Language returns the programming language
func (a *ModernAnalyzerAdapter) Language() string {
	return a.analyzer.Language()
}

// SupportedExtensions returns supported file extensions
func (a *ModernAnalyzerAdapter) SupportedExtensions() []string {
	return a.analyzer.FileExtensions()
}

// CanAnalyze checks if the analyzer can process the given repository
func (a *ModernAnalyzerAdapter) CanAnalyze(repo core.Repository) bool {
	return a.analyzer.CanAnalyze(repo)
}

// Analyze performs full analysis and returns results
func (a *ModernAnalyzerAdapter) Analyze(ctx context.Context, repoPath string, config core.AnalyzerConfig) (*core.AnalysisResult, error) {
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
	AnalyzerRegistry core.AnalyzerRegistry
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
