package core

import (
	"time"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	StatusHealthy  HealthStatus = "healthy"
	StatusWarning  HealthStatus = "warning"
	StatusCritical HealthStatus = "critical"
	StatusUnknown  HealthStatus = "unknown"
)

// Severity represents the severity level of an issue
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Repository represents a code repository
type Repository struct {
	Name      string            `json:"name" yaml:"name"`
	Path      string            `json:"path" yaml:"path"`
	URL       string            `json:"url,omitempty" yaml:"url,omitempty"`
	Branch    string            `json:"branch,omitempty" yaml:"branch,omitempty"`
	Tags      []string          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Language  string            `json:"language,omitempty" yaml:"language,omitempty"`
	Framework string            `json:"framework,omitempty" yaml:"framework,omitempty"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Category   string                 `json:"category"`
	Status     HealthStatus           `json:"status"`
	Score      int                    `json:"score"`
	MaxScore   int                    `json:"max_score"`
	Issues     []Issue                `json:"issues"`
	Warnings   []Warning              `json:"warnings"`
	Metrics    map[string]interface{} `json:"metrics"`
	Metadata   map[string]string      `json:"metadata"`
	Duration   time.Duration          `json:"duration"`
	Timestamp  time.Time              `json:"timestamp"`
	Repository string                 `json:"repository"`
}

// Issue represents a specific issue found during checking
type Issue struct {
	Type       string                 `json:"type"`
	Severity   Severity               `json:"severity"`
	Message    string                 `json:"message"`
	Location   *Location              `json:"location,omitempty"`
	Suggestion string                 `json:"suggestion,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// Warning represents a warning found during checking
type Warning struct {
	Type     string    `json:"type"`
	Message  string    `json:"message"`
	Location *Location `json:"location,omitempty"`
}

// Location represents a location in a file
type Location struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Function string `json:"function,omitempty"`
}

// RepositoryContext contains context for repository checking
type RepositoryContext struct {
	Repository Repository
	Config     Config
	FileSystem FileSystem
	Cache      Cache
	Logger     Logger
}

// Config represents configuration interface
type Config interface {
	GetCheckerConfig(checkerID string) (CheckerConfig, bool)
	GetAnalyzerConfig(language string) (AnalyzerConfig, bool)
	GetReporterConfig(reporterID string) (ReporterConfig, bool)
	GetEngineConfig() EngineConfig
}

// CheckerConfig represents configuration for a checker
type CheckerConfig struct {
	Enabled    bool                   `json:"enabled" yaml:"enabled"`
	Severity   string                 `json:"severity" yaml:"severity"`
	Timeout    time.Duration          `json:"timeout" yaml:"timeout"`
	Options    map[string]interface{} `json:"options" yaml:"options"`
	Categories []string               `json:"categories" yaml:"categories"`
	Exclusions []string               `json:"exclusions" yaml:"exclusions"`
}

// AnalyzerConfig represents configuration for an analyzer
type AnalyzerConfig struct {
	Enabled           bool     `json:"enabled" yaml:"enabled"`
	FileExtensions    []string `json:"file_extensions" yaml:"file_extensions"`
	ExcludePatterns   []string `json:"exclude_patterns" yaml:"exclude_patterns"`
	ComplexityEnabled bool     `json:"complexity_enabled" yaml:"complexity_enabled"`
	FunctionLevel     bool     `json:"function_level" yaml:"function_level"`
}

// ReporterConfig represents configuration for a reporter
type ReporterConfig struct {
	Enabled    bool                   `json:"enabled" yaml:"enabled"`
	OutputFile string                 `json:"output_file" yaml:"output_file"`
	Template   string                 `json:"template" yaml:"template"`
	Options    map[string]interface{} `json:"options" yaml:"options"`
}

// EngineConfig represents configuration for the execution engine
type EngineConfig struct {
	MaxConcurrency int           `json:"max_concurrency" yaml:"max_concurrency"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
	CacheEnabled   bool          `json:"cache_enabled" yaml:"cache_enabled"`
	CacheTTL       time.Duration `json:"cache_ttl" yaml:"cache_ttl"`
}

// FileSystem represents file system operations interface
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	Exists(path string) bool
	IsDir(path string) bool
	ListFiles(path string, pattern string) ([]string, error)
	Walk(path string, walkFn func(path string, info FileInfo) error) error
}

// FileInfo represents file information
type FileInfo struct {
	Name    string
	Size    int64
	Mode    uint32
	ModTime time.Time
	IsDir   bool
}

// Cache represents caching interface
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
}

// Logger represents logging interface
type Logger interface {
	Debug(message string, fields ...Field)
	Info(message string, fields ...Field)
	Warn(message string, fields ...Field)
	Error(message string, fields ...Field)
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// Helper functions for creating log fields
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

func Error(key string, err error) Field {
	return Field{Key: key, Value: err}
}

// Analysis types for language analyzers

// AnalysisResult represents the result of language-specific analysis
type AnalysisResult struct {
	Language  string                   `json:"language"`
	Files     map[string]*FileAnalysis `json:"files"`
	Functions []FunctionInfo           `json:"functions"`
	Metrics   map[string]interface{}   `json:"metrics"`
	Timestamp time.Time                `json:"timestamp"`
	Duration  time.Duration            `json:"duration"`
}

// FileAnalysis represents analysis results for a single file
type FileAnalysis struct {
	Path      string                 `json:"path"`
	Language  string                 `json:"language"`
	Functions []FunctionInfo         `json:"functions"`
	Classes   []ClassInfo            `json:"classes,omitempty"`
	Imports   []ImportInfo           `json:"imports,omitempty"`
	Metrics   map[string]interface{} `json:"metrics"`
	Issues    []Issue                `json:"issues,omitempty"`
}

// FunctionInfo represents information about a function
type FunctionInfo struct {
	Name       string                 `json:"name"`
	File       string                 `json:"file"`
	Line       int                    `json:"line"`
	EndLine    int                    `json:"end_line,omitempty"`
	Complexity int                    `json:"complexity"`
	Language   string                 `json:"language"`
	Parameters []ParameterInfo        `json:"parameters,omitempty"`
	ReturnType string                 `json:"return_type,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ClassInfo represents information about a class
type ClassInfo struct {
	Name      string         `json:"name"`
	File      string         `json:"file"`
	Line      int            `json:"line"`
	EndLine   int            `json:"end_line,omitempty"`
	Language  string         `json:"language"`
	Methods   []FunctionInfo `json:"methods"`
	Fields    []FieldInfo    `json:"fields,omitempty"`
	SuperType string         `json:"super_type,omitempty"`
}

// FieldInfo represents information about a class field
type FieldInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Line       int    `json:"line"`
	Visibility string `json:"visibility,omitempty"`
}

// ParameterInfo represents information about a function parameter
type ParameterInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ImportInfo represents information about imports/includes
type ImportInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Alias   string `json:"alias,omitempty"`
	Line    int    `json:"line"`
	IsLocal bool   `json:"is_local"`
}
