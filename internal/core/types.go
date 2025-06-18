package core

import (
	"time"
)

// Config represents the main configuration interface
type Config interface {
	GetCheckerConfig(checkerID string) (CheckerConfig, bool)
	GetAnalyzerConfig(language string) (AnalyzerConfig, bool)
	GetReporterConfig(reporterID string) (ReporterConfig, bool)
	GetEngineConfig() EngineConfig
}

// Logger represents a structured logger interface
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

// FileSystem represents a file system interface
type FileSystem interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte) error
	Exists(filename string) bool
	IsDir(filename string) bool
	ListFiles(path string, pattern string) ([]string, error)
	Walk(root string, walkFn func(path string, info FileInfo) error) error
}

// FileInfo represents file information
type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Mode    uint32    `json:"mode"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}

// FileAnalysis represents analysis of a single file
type FileAnalysis struct {
	Path       string                 `json:"path"`
	Language   string                 `json:"language"`
	Lines      int                    `json:"lines"`
	Functions  []FunctionInfo         `json:"functions"`
	Classes    []ClassInfo            `json:"classes"`
	Imports    []ImportInfo           `json:"imports"`
	Complexity int                    `json:"complexity"`
	Issues     []Issue                `json:"issues"`
	Metrics    map[string]interface{} `json:"metrics"`
}

// FunctionInfo represents information about a function
type FunctionInfo struct {
	Name       string `json:"name"`
	File       string `json:"file"`
	Language   string `json:"language"`
	Line       int    `json:"line"`
	Complexity int    `json:"complexity"`
}

// ClassInfo represents information about a class
type ClassInfo struct {
	Name     string         `json:"name"`
	File     string         `json:"file"`
	Language string         `json:"language"`
	Line     int            `json:"line"`
	Methods  []FunctionInfo `json:"methods"`
	Fields   []FieldInfo    `json:"fields"`
}

// FieldInfo represents information about a class field
type FieldInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Line int    `json:"line"`
}

// ImportInfo represents information about an import
type ImportInfo struct {
	Name     string `json:"name,omitempty"`
	Path     string `json:"path"`
	Alias    string `json:"alias,omitempty"`
	Line     int    `json:"line"`
	IsStatic bool   `json:"is_static,omitempty"`
	IsLocal  bool   `json:"is_local,omitempty"`
}

// CheckerConfig represents configuration for a checker
type CheckerConfig struct {
	Enabled    bool                   `yaml:"enabled" json:"enabled"`
	Severity   string                 `yaml:"severity" json:"severity"`
	Timeout    time.Duration          `yaml:"timeout" json:"timeout"`
	Categories []string               `yaml:"categories" json:"categories"`
	Options    map[string]interface{} `yaml:"options" json:"options"`
	Exclusions []string               `yaml:"exclusions" json:"exclusions"`
}

// AnalyzerConfig represents configuration for an analyzer
type AnalyzerConfig struct {
	Enabled           bool                   `yaml:"enabled" json:"enabled"`
	FileExtensions    []string               `yaml:"file_extensions" json:"file_extensions"`
	ExcludePatterns   []string               `yaml:"exclude_patterns" json:"exclude_patterns"`
	ComplexityEnabled bool                   `yaml:"complexity_enabled" json:"complexity_enabled"`
	FunctionLevel     bool                   `yaml:"function_level" json:"function_level"`
	Categories        []string               `yaml:"categories" json:"categories"`
	Options           map[string]interface{} `yaml:"options" json:"options"`
}

// ReporterConfig represents configuration for a reporter
type ReporterConfig struct {
	Enabled    bool                   `yaml:"enabled" json:"enabled"`
	Format     string                 `yaml:"format" json:"format"`
	Output     string                 `yaml:"output" json:"output"`
	OutputFile string                 `yaml:"output_file" json:"output_file"`
	Template   string                 `yaml:"template" json:"template"`
	Options    map[string]interface{} `yaml:"options" json:"options"`
}

// EngineConfig represents configuration for the health engine
type EngineConfig struct {
	MaxConcurrency int           `yaml:"max_concurrency" json:"max_concurrency"`
	Timeout        time.Duration `yaml:"timeout" json:"timeout"`
	CacheEnabled   bool          `yaml:"cache_enabled" json:"cache_enabled"`
	CacheTTL       time.Duration `yaml:"cache_ttl" json:"cache_ttl"`
	Parallel       bool          `yaml:"parallel" json:"parallel"`
}

// Repository represents a repository to be analyzed
type Repository struct {
	Name      string            `yaml:"name" json:"name"`
	URL       string            `yaml:"url" json:"url"`
	Branch    string            `yaml:"branch" json:"branch"`
	Path      string            `yaml:"path" json:"path"`
	Tags      []string          `yaml:"tags" json:"tags"`
	Language  string            `yaml:"language" json:"language"`
	Framework string            `yaml:"framework" json:"framework"`
	Metadata  map[string]string `yaml:"metadata" json:"metadata"`
}

// RepositoryContext provides context for repository operations
type RepositoryContext struct {
	Repository Repository        `json:"repository"`
	Config     Config            `json:"config"`
	Metadata   map[string]string `json:"metadata"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Category   string                 `json:"category"`
	Repository string                 `json:"repository"`
	Status     HealthStatus           `json:"status"`
	Score      int                    `json:"score"`
	MaxScore   int                    `json:"max_score"`
	Issues     []Issue                `json:"issues"`
	Warnings   []Warning              `json:"warnings"`
	Metrics    map[string]interface{} `json:"metrics"`
	Metadata   map[string]string      `json:"metadata"`
	Duration   time.Duration          `json:"duration"`
	Timestamp  time.Time              `json:"timestamp"`
	Error      string                 `json:"error,omitempty"`
}

// AnalysisResult represents the result of code analysis
type AnalysisResult struct {
	Language          string                   `json:"language"`
	TotalFiles        int                      `json:"total_files"`
	TotalLines        int                      `json:"total_lines"`
	TotalFunctions    int                      `json:"total_functions"`
	AverageComplexity float64                  `json:"average_complexity"`
	Functions         []FunctionInfo           `json:"functions"`
	Files             map[string]*FileAnalysis `json:"files"`
	Patterns          []PatternMatch           `json:"patterns"`
	Metrics           map[string]interface{}   `json:"metrics"`
}

// HealthStatus represents the health status of a check
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

// Issue represents a health check issue
type Issue struct {
	Type        string                 `json:"type"`
	Severity    Severity               `json:"severity"`
	Message     string                 `json:"message"`
	Description string                 `json:"description,omitempty"`
	Location    *Location              `json:"location,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Suggestion  string                 `json:"suggestion,omitempty"`
}

// Warning represents a health check warning
type Warning struct {
	Type        string    `json:"type"`
	Message     string    `json:"message"`
	Description string    `json:"description,omitempty"`
	Location    *Location `json:"location,omitempty"`
	Context     string    `json:"context,omitempty"`
}

// Location represents a location in code
type Location struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column,omitempty"`
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// Logger helper functions for structured logging
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

func Error(key string, value error) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}
