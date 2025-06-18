package orchestration

import (
	"context"
	"time"

	"github.com/codcod/repos/internal/core"
)

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
