package orchestration

import (
	"context"
	"time"

	"github.com/codcod/repos/internal/core"
)

// Pipeline represents a configurable execution pipeline
type Pipeline struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Steps       []PipelineStep    `json:"steps" yaml:"steps"`
	Config      PipelineConfig    `json:"config" yaml:"config"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// PipelineStep represents a single step in the pipeline
type PipelineStep struct {
	Name         string                 `json:"name" yaml:"name"`
	Type         StepType               `json:"type" yaml:"type"`
	Config       map[string]interface{} `json:"config" yaml:"config"`
	Dependencies []string               `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Enabled      bool                   `json:"enabled" yaml:"enabled"`
	Timeout      time.Duration          `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// StepType represents the type of pipeline step
type StepType string

const (
	StepTypeAnalysis   StepType = "analysis"
	StepTypeCheckers   StepType = "checkers"
	StepTypeReporting  StepType = "reporting"
	StepTypeValidation StepType = "validation"
	StepTypeCustom     StepType = "custom"
)

// PipelineConfig represents configuration for pipeline execution
type PipelineConfig struct {
	MaxConcurrency  int                    `json:"max_concurrency" yaml:"max_concurrency"`
	Timeout         time.Duration          `json:"timeout" yaml:"timeout"`
	FailFast        bool                   `json:"fail_fast" yaml:"fail_fast"`
	RetryCount      int                    `json:"retry_count" yaml:"retry_count"`
	RetryDelay      time.Duration          `json:"retry_delay" yaml:"retry_delay"`
	ContinueOnError bool                   `json:"continue_on_error" yaml:"continue_on_error"`
	OutputFormats   []string               `json:"output_formats" yaml:"output_formats"`
	ReportingConfig map[string]interface{} `json:"reporting_config" yaml:"reporting_config"`
}

// ExecutionContext contains context for pipeline execution
type ExecutionContext struct {
	Pipeline     Pipeline
	Repositories []core.Repository
	Config       core.Config
	Logger       core.Logger
	Cache        core.Cache
	StartTime    time.Time
}

// StepResult represents the result of a pipeline step
type StepResult struct {
	StepName  string                 `json:"step_name"`
	Status    ExecutionStatus        `json:"status"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Duration  time.Duration          `json:"duration"`
	Output    map[string]interface{} `json:"output,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

// ExecutionStatus represents the status of execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusSkipped   ExecutionStatus = "skipped"
	StatusCanceled  ExecutionStatus = "canceled"
)

// PipelineResult represents the result of pipeline execution
type PipelineResult struct {
	Pipeline       Pipeline               `json:"pipeline"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        time.Time              `json:"end_time"`
	Duration       time.Duration          `json:"duration"`
	Status         ExecutionStatus        `json:"status"`
	StepResults    []StepResult           `json:"step_results"`
	WorkflowResult *core.WorkflowResult   `json:"workflow_result,omitempty"`
	Error          string                 `json:"error,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionPlan represents a plan for executing checks
type ExecutionPlan struct {
	Repositories  []core.Repository   `json:"repositories"`
	CheckerGroups []CheckerGroup      `json:"checker_groups"`
	Dependencies  map[string][]string `json:"dependencies"`
	Config        ExecutionPlanConfig `json:"config"`
}

// CheckerGroup represents a group of related checkers
type CheckerGroup struct {
	Name       string   `json:"name"`
	Category   string   `json:"category"`
	CheckerIDs []string `json:"checker_ids"`
	Parallel   bool     `json:"parallel"`
	Required   bool     `json:"required"`
}

// ExecutionPlanConfig represents configuration for execution plan
type ExecutionPlanConfig struct {
	MaxParallelRepos    int           `json:"max_parallel_repos"`
	MaxParallelCheckers int           `json:"max_parallel_checkers"`
	Timeout             time.Duration `json:"timeout"`
	CacheEnabled        bool          `json:"cache_enabled"`
	CacheTTL            time.Duration `json:"cache_ttl"`
}

// ProgressReporter reports progress during execution
type ProgressReporter interface {
	ReportProgress(ctx context.Context, progress Progress)
	ReportStepStart(ctx context.Context, stepName string)
	ReportStepComplete(ctx context.Context, stepName string, result StepResult)
	ReportError(ctx context.Context, err error)
}

// Progress represents execution progress
type Progress struct {
	TotalRepos        int           `json:"total_repos"`
	CompletedRepos    int           `json:"completed_repos"`
	TotalSteps        int           `json:"total_steps"`
	CompletedSteps    int           `json:"completed_steps"`
	PercentComplete   float64       `json:"percent_complete"`
	EstimatedTimeLeft time.Duration `json:"estimated_time_left"`
	CurrentStep       string        `json:"current_step"`
	Status            string        `json:"status"`
}

// StepExecutor defines the interface for executing pipeline steps
type StepExecutor interface {
	Execute(ctx context.Context, step PipelineStep, execCtx ExecutionContext) (StepResult, error)
	SupportsStepType(stepType StepType) bool
}

// PipelineExecutor orchestrates pipeline execution
type PipelineExecutor interface {
	ExecutePipeline(ctx context.Context, pipeline Pipeline, repos []core.Repository) (*PipelineResult, error)
	ValidatePipeline(pipeline Pipeline) error
	RegisterStepExecutor(stepType StepType, executor StepExecutor)
}
