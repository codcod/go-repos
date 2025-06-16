package core

import (
	"context"
	"time"
)

// Checker represents a health checker interface
type Checker interface {
	ID() string
	Name() string
	Category() string
	Check(ctx context.Context, repoCtx RepositoryContext) (CheckResult, error)
	SupportsRepository(repo Repository) bool
}

// Analyzer represents a language-specific analyzer interface
type Analyzer interface {
	Name() string
	Language() string
	SupportedExtensions() []string
	CanAnalyze(repo Repository) bool
	Analyze(ctx context.Context, repo Repository) (*AnalysisResult, error)
}

// LegacyAnalyzer represents the legacy analyzer interface for backward compatibility
type LegacyAnalyzer interface {
	Language() string
	FileExtensions() []string
	SupportsComplexity() bool
	SupportsFunctionLevel() bool
	AnalyzeComplexity(ctx context.Context, repoPath string) (ComplexityResult, error)
	AnalyzeFunctions(ctx context.Context, repoPath string) ([]FunctionComplexity, error)
	DetectPatterns(ctx context.Context, content string, patterns []Pattern) ([]PatternMatch, error)
}

// Reporter represents a result reporter interface
type Reporter interface {
	ID() string
	Name() string
	SupportedFormats() []string
	Report(ctx context.Context, results []CheckResult, config ReporterConfig) error
}

// ComplexityResult represents complexity analysis results
type ComplexityResult struct {
	TotalFiles        int                    `json:"total_files"`
	TotalFunctions    int                    `json:"total_functions"`
	AverageComplexity float64                `json:"average_complexity"`
	MaxComplexity     int                    `json:"max_complexity"`
	Functions         []FunctionComplexity   `json:"functions"`
	FileMetrics       map[string]FileMetrics `json:"file_metrics"`
}

// FunctionComplexity represents complexity information for a function
type FunctionComplexity struct {
	Name       string `json:"name"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Complexity int    `json:"complexity"`
	Length     int    `json:"length"`
}

// FileMetrics represents metrics for a file
type FileMetrics struct {
	Path              string  `json:"path"`
	Language          string  `json:"language"`
	Lines             int     `json:"lines"`
	Functions         int     `json:"functions"`
	AverageComplexity float64 `json:"average_complexity"`
	MaxComplexity     int     `json:"max_complexity"`
}

// Pattern represents a pattern to detect
type Pattern struct {
	Name       string   `json:"name"`
	Patterns   []string `json:"patterns"`
	Severity   Severity `json:"severity"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion"`
}

// PatternMatch represents a pattern match result
type PatternMatch struct {
	Pattern   Pattern  `json:"pattern"`
	Location  Location `json:"location"`
	MatchText string   `json:"match_text"`
	Context   string   `json:"context"`
}

// CheckReport represents a complete check report
type CheckReport struct {
	Results   []RepositoryResult `json:"results"`
	Summary   Summary            `json:"summary"`
	Duration  time.Duration      `json:"duration"`
	Timestamp time.Time          `json:"timestamp"`
}

// RepositoryResult represents results for a single repository
type RepositoryResult struct {
	Repository Repository    `json:"repository"`
	Results    []CheckResult `json:"results"`
	Status     HealthStatus  `json:"status"`
	Score      int           `json:"score"`
	MaxScore   int           `json:"max_score"`
	Error      error         `json:"error,omitempty"`
}

// Summary represents a summary of check results
type Summary struct {
	TotalRepositories int                        `json:"total_repositories"`
	HealthyCount      int                        `json:"healthy_count"`
	WarningCount      int                        `json:"warning_count"`
	CriticalCount     int                        `json:"critical_count"`
	CategorySummary   map[string]CategorySummary `json:"category_summary"`
	OverallScore      float64                    `json:"overall_score"`
}

// CategorySummary represents summary for a category
type CategorySummary struct {
	ChecksRun     int     `json:"checks_run"`
	IssuesFound   int     `json:"issues_found"`
	WarningsFound int     `json:"warnings_found"`
	AverageScore  float64 `json:"average_score"`
}
