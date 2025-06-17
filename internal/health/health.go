package health

import (
	"time"

	"github.com/codcod/repos/internal/core"
	analyzer_registry "github.com/codcod/repos/internal/health/analyzers/registry"
	checker_registry "github.com/codcod/repos/internal/health/checkers/registry"
	"github.com/codcod/repos/internal/health/orchestration"
	"github.com/codcod/repos/internal/health/reporting"
	"github.com/codcod/repos/internal/platform/commands"
	"github.com/codcod/repos/internal/platform/filesystem"
)

// Re-export key types for cleaner imports
type (
	AnalyzerRegistry = analyzer_registry.Registry
	CheckerRegistry  = checker_registry.CheckerRegistry
	Engine           = orchestration.Engine
	Formatter        = reporting.Formatter
)

// NewAnalyzerRegistry creates a new analyzer registry with all standard analyzers
func NewAnalyzerRegistry(fs core.FileSystem, logger core.Logger) *AnalyzerRegistry {
	return analyzer_registry.NewRegistryWithStandardAnalyzers(fs, logger)
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

	analyzerRegistry := NewAnalyzerRegistry(fs, logger)
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
