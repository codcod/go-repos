package health

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Severity levels for health checks
type Severity int

const (
	SeverityInfo Severity = iota + 1
	SeverityWarning
	SeverityCritical
)

// String returns the string representation of severity
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ErrorCode represents different types of checker errors
type ErrorCode string

const (
	ErrorCodeToolNotFound ErrorCode = "TOOL_NOT_FOUND"
	ErrorCodeFileNotFound ErrorCode = "FILE_NOT_FOUND"
	ErrorCodeTimeout      ErrorCode = "TIMEOUT"
	ErrorCodePermission   ErrorCode = "PERMISSION_DENIED"
	ErrorCodeInvalidInput ErrorCode = "INVALID_INPUT"
	ErrorCodeNetworkError ErrorCode = "NETWORK_ERROR"
)

// CheckerError represents an error that occurred during health checking
type CheckerError struct {
	Checker   string
	Operation string
	Err       error
	Code      ErrorCode
}

func (c *CheckerError) Error() string {
	return fmt.Sprintf("[%s] %s: %s (%s)", c.Checker, c.Operation, c.Err.Error(), c.Code)
}

// NewCheckerError creates a new checker error
func NewCheckerError(checker, operation string, err error, code ErrorCode) *CheckerError {
	return &CheckerError{
		Checker:   checker,
		Operation: operation,
		Err:       err,
		Code:      code,
	}
}

// FilePattern represents file patterns for different languages
type FilePattern struct {
	Extensions []string
	Paths      []string
	Exclude    []string
}

// Common file patterns
var (
	GoFilePattern = FilePattern{
		Extensions: []string{".go"},
		Exclude:    []string{"vendor/", ".git/", "*_test.go"},
	}

	JavaFilePattern = FilePattern{
		Extensions: []string{".java"},
		Exclude:    []string{"target/", ".git/", "build/"},
	}

	JavaScriptFilePattern = FilePattern{
		Extensions: []string{".js", ".ts", ".jsx", ".tsx"},
		Exclude:    []string{"node_modules/", ".git/", "dist/", "build/"},
	}

	PythonFilePattern = FilePattern{
		Extensions: []string{".py"},
		Exclude:    []string{"__pycache__/", ".git/", "venv/", ".venv/"},
	}

	CFilePattern = FilePattern{
		Extensions: []string{".c", ".cpp", ".h", ".hpp"},
		Exclude:    []string{".git/"},
	}
)

// CommandExecutor handles command execution with consistent timeout and context management
type CommandExecutor struct {
	timeout time.Duration
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	Output   []byte
	ExitCode int
	Duration time.Duration
	Error    error
}

// NewCommandExecutor creates a new command executor with default timeout
func NewCommandExecutor(timeout time.Duration) *CommandExecutor {
	return &CommandExecutor{timeout: timeout}
}

// Execute runs a command with context and returns the result
func (e *CommandExecutor) Execute(ctx context.Context, workDir, command string, args ...string) CommandResult {
	start := time.Now()

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = workDir

	output, err := cmd.Output()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return CommandResult{
		Output:   output,
		ExitCode: exitCode,
		Duration: duration,
		Error:    err,
	}
}

// IsCommandAvailable checks if a command is available in PATH
func (e *CommandExecutor) IsCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// FileSystemHelper provides file system operations
type FileSystemHelper struct{}

// NewFileSystemHelper creates a new file system helper
func NewFileSystemHelper() *FileSystemHelper {
	return &FileSystemHelper{}
}

// Exists checks if a file or directory exists
func (f *FileSystemHelper) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFile reads a file and returns its content
func (f *FileSystemHelper) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path) // #nosec G304 - path validation should be done by caller
}

// FindFiles finds files matching a pattern, respecting exclude patterns
func (f *FileSystemHelper) FindFiles(repoPath string, pattern FilePattern) ([]string, error) {
	var files []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		// Skip excluded paths
		relPath, _ := filepath.Rel(repoPath, path)
		for _, exclude := range pattern.Exclude {
			if matched, _ := filepath.Match(exclude, relPath); matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check if file matches extensions
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(info.Name())
		for _, validExt := range pattern.Extensions {
			if ext == validExt {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	return files, err
}

// FindFilesSimple finds files matching a simple glob pattern
func (f *FileSystemHelper) FindFilesSimple(repoPath, pattern string) []string {
	var files []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		if matched, _ := filepath.Match(pattern, info.Name()); matched {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return []string{}
	}

	return files
}

// HealthCheckBuilder provides a fluent interface for building health check results
type HealthCheckBuilder struct {
	name     string
	category string
	status   HealthStatus
	message  string
	details  strings.Builder
	severity int
	issues   []string
	warnings []string
	info     []string
}

// NewHealthCheckBuilder creates a new health check builder
func NewHealthCheckBuilder(name, category string) *HealthCheckBuilder {
	return &HealthCheckBuilder{
		name:     name,
		category: category,
		status:   HealthStatusHealthy,
		severity: 1,
	}
}

// AddIssue adds a critical issue to the health check
func (b *HealthCheckBuilder) AddIssue(issue string) *HealthCheckBuilder {
	b.issues = append(b.issues, issue)
	b.status = HealthStatusCritical
	b.severity = 3
	return b
}

// AddWarning adds a warning to the health check
func (b *HealthCheckBuilder) AddWarning(warning string) *HealthCheckBuilder {
	b.warnings = append(b.warnings, warning)
	if b.status == HealthStatusHealthy {
		b.status = HealthStatusWarning
		b.severity = 2
	}
	return b
}

// AddInfo adds informational content to the health check
func (b *HealthCheckBuilder) AddInfo(info string) *HealthCheckBuilder {
	b.info = append(b.info, info)
	return b
}

// SetStatus explicitly sets the status
func (b *HealthCheckBuilder) SetStatus(status HealthStatus, severity int) *HealthCheckBuilder {
	b.status = status
	b.severity = severity
	return b
}

// SetMessage sets a custom message
func (b *HealthCheckBuilder) SetMessage(message string) *HealthCheckBuilder {
	b.message = message
	return b
}

// Build creates the final HealthCheck
func (b *HealthCheckBuilder) Build() HealthCheck {
	if b.message == "" {
		b.buildMessage()
	}
	b.buildDetails()

	return HealthCheck{
		Name:        b.name,
		Status:      b.status,
		Message:     b.message,
		Details:     b.details.String(),
		Severity:    b.severity,
		Category:    b.category,
		LastChecked: time.Now(),
	}
}

// buildMessage creates a default message based on issues and warnings
func (b *HealthCheckBuilder) buildMessage() {
	if len(b.issues) > 0 {
		b.message = fmt.Sprintf("Found %d critical issues", len(b.issues))
	} else if len(b.warnings) > 0 {
		b.message = fmt.Sprintf("Found %d warnings", len(b.warnings))
	} else {
		b.message = "All checks passed"
	}
}

// buildDetails creates the details section
func (b *HealthCheckBuilder) buildDetails() {
	if len(b.issues) > 0 {
		b.details.WriteString("Issues:\n")
		for _, issue := range b.issues {
			fmt.Fprintf(&b.details, "  ❌ %s\n", issue)
		}
		b.details.WriteString("\n")
	}

	if len(b.warnings) > 0 {
		b.details.WriteString("Warnings:\n")
		for _, warning := range b.warnings {
			fmt.Fprintf(&b.details, "  ⚠️ %s\n", warning)
		}
		b.details.WriteString("\n")
	}

	if len(b.info) > 0 {
		b.details.WriteString("Information:\n")
		for _, info := range b.info {
			fmt.Fprintf(&b.details, "  ℹ️ %s\n", info)
		}
	}
}

// PathValidator provides path validation utilities
type PathValidator struct{}

// NewPathValidator creates a new path validator
func NewPathValidator() *PathValidator {
	return &PathValidator{}
}

// IsValidFilePath validates file paths to prevent directory traversal attacks
func (v *PathValidator) IsValidFilePath(filePath string) bool {
	// Check for directory traversal patterns
	if strings.Contains(filePath, "..") {
		return false
	}

	// Check for absolute paths outside of expected directories
	if filepath.IsAbs(filePath) {
		allowedPrefixes := []string{
			"/Users/",
			"/home/",
			"/tmp/",
			"/var/",
			"/opt/",
		}

		hasValidPrefix := false
		for _, prefix := range allowedPrefixes {
			if strings.HasPrefix(filePath, prefix) {
				hasValidPrefix = true
				break
			}
		}

		if !hasValidPrefix {
			return false
		}
	}

	// Check file extension is reasonable
	ext := filepath.Ext(filePath)
	validExts := []string{".go", ".java", ".js", ".ts", ".py", ".c", ".cpp", ".h", ".hpp", ".yaml", ".yml", ".json", ".md", ".txt"}
	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}

	return false
}

// GetRelativePath gets a relative path for display purposes
func (v *PathValidator) GetRelativePath(filePath, basePath string) string {
	relPath, err := filepath.Rel(basePath, filePath)
	if err != nil || relPath == filePath {
		return filepath.Base(filePath)
	}
	return relPath
}
