package orchestration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codcod/repos/internal/core"
)

// DefaultPipelineExecutor implements PipelineExecutor
type DefaultPipelineExecutor struct {
	stepExecutors map[StepType]StepExecutor
	engine        *Engine
	logger        core.Logger
	mu            sync.RWMutex
}

// NewDefaultPipelineExecutor creates a new pipeline executor
func NewDefaultPipelineExecutor(engine *Engine, logger core.Logger) *DefaultPipelineExecutor {
	executor := &DefaultPipelineExecutor{
		stepExecutors: make(map[StepType]StepExecutor),
		engine:        engine,
		logger:        logger,
	}

	// Register default step executors
	executor.RegisterStepExecutor(StepTypeAnalysis, &AnalysisStepExecutor{engine: engine, logger: logger})
	executor.RegisterStepExecutor(StepTypeCheckers, &CheckersStepExecutor{engine: engine, logger: logger})
	executor.RegisterStepExecutor(StepTypeReporting, &ReportingStepExecutor{logger: logger})
	executor.RegisterStepExecutor(StepTypeValidation, &ValidationStepExecutor{logger: logger})

	return executor
}

// RegisterStepExecutor registers a step executor for a specific step type
func (p *DefaultPipelineExecutor) RegisterStepExecutor(stepType StepType, executor StepExecutor) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stepExecutors[stepType] = executor
}

// ExecutePipeline executes a complete pipeline
func (p *DefaultPipelineExecutor) ExecutePipeline(ctx context.Context, pipeline Pipeline, repos []core.Repository) (*PipelineResult, error) {
	p.logger.Info("Starting pipeline execution",
		core.String("pipeline", pipeline.Name),
		core.Int("repositories", len(repos)),
		core.Int("steps", len(pipeline.Steps)))

	startTime := time.Now()

	// Validate pipeline first
	if err := p.ValidatePipeline(pipeline); err != nil {
		return nil, fmt.Errorf("pipeline validation failed: %w", err)
	}

	// Create execution context
	execCtx := ExecutionContext{
		Pipeline:     pipeline,
		Repositories: repos,
		Config:       p.engine.config,
		Logger:       p.logger,
		Cache:        nil, // Would be injected
		StartTime:    startTime,
	}

	// Create pipeline context with timeout
	pipelineCtx, cancel := context.WithTimeout(ctx, pipeline.Config.Timeout)
	defer cancel()

	result := &PipelineResult{
		Pipeline:    pipeline,
		StartTime:   startTime,
		Status:      StatusRunning,
		StepResults: make([]StepResult, 0, len(pipeline.Steps)),
	}

	// Execute steps in order
	var workflowResult *core.WorkflowResult
	for _, step := range pipeline.Steps {
		if !step.Enabled {
			p.logger.Debug("Skipping disabled step", core.String("step", step.Name))
			result.StepResults = append(result.StepResults, StepResult{
				StepName:  step.Name,
				Status:    StatusSkipped,
				StartTime: time.Now(),
				EndTime:   time.Now(),
			})
			continue
		}

		stepResult, err := p.executeStep(pipelineCtx, step, execCtx)
		result.StepResults = append(result.StepResults, stepResult)

		if err != nil {
			p.logger.Error("Step execution failed",
				core.String("step", step.Name),
				core.Error("error", err))

			if pipeline.Config.FailFast {
				result.Status = StatusFailed
				result.Error = fmt.Sprintf("Step '%s' failed: %v", step.Name, err)
				break
			}
		}

		// If this is the main checkers step, capture the workflow result
		if step.Type == StepTypeCheckers {
			if output, exists := stepResult.Output["workflow_result"]; exists {
				if wr, ok := output.(*core.WorkflowResult); ok {
					workflowResult = wr
				}
			}
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)
	result.WorkflowResult = workflowResult

	if result.Status == StatusRunning {
		result.Status = StatusCompleted
	}

	p.logger.Info("Pipeline execution completed",
		core.String("pipeline", pipeline.Name),
		core.String("status", string(result.Status)),
		core.Duration("duration", result.Duration))

	return result, nil
}

// executeStep executes a single pipeline step
func (p *DefaultPipelineExecutor) executeStep(ctx context.Context, step PipelineStep, execCtx ExecutionContext) (StepResult, error) {
	p.logger.Debug("Executing step", core.String("step", step.Name), core.String("type", string(step.Type)))

	p.mu.RLock()
	executor, exists := p.stepExecutors[step.Type]
	p.mu.RUnlock()

	if !exists {
		return StepResult{
			StepName:  step.Name,
			Status:    StatusFailed,
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Error:     fmt.Sprintf("No executor found for step type '%s'", step.Type),
		}, fmt.Errorf("no executor found for step type '%s'", step.Type)
	}

	// Create step context with timeout
	stepCtx := ctx
	if step.Timeout > 0 {
		var cancel context.CancelFunc
		stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
	}

	return executor.Execute(stepCtx, step, execCtx)
}

// ValidatePipeline validates a pipeline configuration
func (p *DefaultPipelineExecutor) ValidatePipeline(pipeline Pipeline) error {
	if pipeline.Name == "" {
		return fmt.Errorf("pipeline name is required")
	}

	if len(pipeline.Steps) == 0 {
		return fmt.Errorf("pipeline must have at least one step")
	}

	// Validate step dependencies
	stepNames := make(map[string]bool)
	for _, step := range pipeline.Steps {
		stepNames[step.Name] = true
	}

	for _, step := range pipeline.Steps {
		for _, dep := range step.Dependencies {
			if !stepNames[dep] {
				return fmt.Errorf("step '%s' depends on non-existent step '%s'", step.Name, dep)
			}
		}

		// Validate step executor exists
		p.mu.RLock()
		_, exists := p.stepExecutors[step.Type]
		p.mu.RUnlock()

		if !exists {
			return fmt.Errorf("no executor registered for step type '%s' in step '%s'", step.Type, step.Name)
		}
	}

	return nil
}

// AnalysisStepExecutor executes analysis steps
type AnalysisStepExecutor struct {
	engine *Engine
	logger core.Logger
}

// Execute implements StepExecutor for analysis steps
func (e *AnalysisStepExecutor) Execute(ctx context.Context, step PipelineStep, execCtx ExecutionContext) (StepResult, error) {
	startTime := time.Now()
	result := StepResult{
		StepName:  step.Name,
		Status:    StatusRunning,
		StartTime: startTime,
		Output:    make(map[string]interface{}),
	}

	e.logger.Info("Executing analysis step", core.String("step", step.Name))

	// Run analysis for each repository
	analysisResults := make(map[string]*core.AnalysisResult)
	for _, repo := range execCtx.Repositories {
		if repo.Language == "" {
			continue // Skip repositories without detected language
		}

		repoCtx := core.RepositoryContext{
			Repository: repo,
			Config:     execCtx.Config,
			Logger:     e.logger,
		}

		analysisResult, err := e.engine.runAnalysis(ctx, repoCtx)
		if err != nil {
			e.logger.Warn("Analysis failed for repository",
				core.String("repository", repo.Name),
				core.Error("error", err))
			continue
		}

		analysisResults[repo.Name] = analysisResult
	}

	result.Output["analysis_results"] = analysisResults
	result.Output["repositories_analyzed"] = len(analysisResults)
	result.Status = StatusCompleted
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	e.logger.Info("Analysis step completed",
		core.String("step", step.Name),
		core.Int("repositories_analyzed", len(analysisResults)),
		core.Duration("duration", result.Duration))

	return result, nil
}

// SupportsStepType implements StepExecutor
func (e *AnalysisStepExecutor) SupportsStepType(stepType StepType) bool {
	return stepType == StepTypeAnalysis
}

// CheckersStepExecutor executes checker steps
type CheckersStepExecutor struct {
	engine *Engine
	logger core.Logger
}

// Execute implements StepExecutor for checker steps
func (e *CheckersStepExecutor) Execute(ctx context.Context, step PipelineStep, execCtx ExecutionContext) (StepResult, error) {
	startTime := time.Now()
	result := StepResult{
		StepName:  step.Name,
		Status:    StatusRunning,
		StartTime: startTime,
		Output:    make(map[string]interface{}),
	}

	e.logger.Info("Executing checkers step", core.String("step", step.Name))

	// Execute health check workflow
	workflowResult, err := e.engine.ExecuteHealthCheck(ctx, execCtx.Repositories)
	if err != nil {
		result.Status = StatusFailed
		result.Error = err.Error()
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		return result, err
	}

	result.Output["workflow_result"] = workflowResult
	result.Output["total_repositories"] = workflowResult.TotalRepos
	result.Output["successful_repositories"] = workflowResult.Summary.SuccessfulRepos
	result.Output["failed_repositories"] = workflowResult.Summary.FailedRepos
	result.Status = StatusCompleted
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	e.logger.Info("Checkers step completed",
		core.String("step", step.Name),
		core.Int("total_repositories", workflowResult.TotalRepos),
		core.Int("successful_repositories", workflowResult.Summary.SuccessfulRepos),
		core.Duration("duration", result.Duration))

	return result, nil
}

// SupportsStepType implements StepExecutor
func (e *CheckersStepExecutor) SupportsStepType(stepType StepType) bool {
	return stepType == StepTypeCheckers
}

// ReportingStepExecutor executes reporting steps
type ReportingStepExecutor struct {
	logger core.Logger
}

// Execute implements StepExecutor for reporting steps
func (e *ReportingStepExecutor) Execute(ctx context.Context, step PipelineStep, execCtx ExecutionContext) (StepResult, error) {
	startTime := time.Now()
	result := StepResult{
		StepName:  step.Name,
		Status:    StatusRunning,
		StartTime: startTime,
		Output:    make(map[string]interface{}),
	}

	e.logger.Info("Executing reporting step", core.String("step", step.Name))

	// For now, just log that reporting would happen
	// In a real implementation, this would generate reports based on previous step results
	result.Output["reports_generated"] = []string{"console", "json"}
	result.Status = StatusCompleted
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	e.logger.Info("Reporting step completed",
		core.String("step", step.Name),
		core.Duration("duration", result.Duration))

	return result, nil
}

// SupportsStepType implements StepExecutor
func (e *ReportingStepExecutor) SupportsStepType(stepType StepType) bool {
	return stepType == StepTypeReporting
}

// ValidationStepExecutor executes validation steps
type ValidationStepExecutor struct {
	logger core.Logger
}

// Execute implements StepExecutor for validation steps
func (e *ValidationStepExecutor) Execute(ctx context.Context, step PipelineStep, execCtx ExecutionContext) (StepResult, error) {
	startTime := time.Now()
	result := StepResult{
		StepName:  step.Name,
		Status:    StatusRunning,
		StartTime: startTime,
		Output:    make(map[string]interface{}),
	}

	e.logger.Info("Executing validation step", core.String("step", step.Name))

	// Validate repositories are accessible
	validRepos := 0
	for _, repo := range execCtx.Repositories {
		// Basic validation - check if repository path exists
		// In a real implementation, this would do more thorough validation
		if repo.Path != "" {
			validRepos++
		}
	}

	result.Output["total_repositories"] = len(execCtx.Repositories)
	result.Output["valid_repositories"] = validRepos
	result.Status = StatusCompleted
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	e.logger.Info("Validation step completed",
		core.String("step", step.Name),
		core.Int("valid_repositories", validRepos),
		core.Duration("duration", result.Duration))

	return result, nil
}

// SupportsStepType implements StepExecutor
func (e *ValidationStepExecutor) SupportsStepType(stepType StepType) bool {
	return stepType == StepTypeValidation
}
