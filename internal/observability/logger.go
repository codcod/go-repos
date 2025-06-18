// Package observability provides structured logging and metrics for the repos tool.
package observability

import (
	"fmt"
	"os"
	"time"

	"github.com/codcod/repos/internal/core"
	"github.com/fatih/color"
)

// Level represents log levels
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// StructuredLogger provides structured logging with context and fields
type StructuredLogger struct {
	level  Level
	fields map[string]interface{}
	prefix string
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(level Level) *StructuredLogger {
	return &StructuredLogger{
		level:  level,
		fields: make(map[string]interface{}),
	}
}

// WithPrefix creates a new logger with a prefix
func (l *StructuredLogger) WithPrefix(prefix string) *StructuredLogger {
	return &StructuredLogger{
		level:  l.level,
		fields: copyFields(l.fields),
		prefix: prefix,
	}
}

// WithField adds a field to the logger context
func (l *StructuredLogger) WithField(key string, value interface{}) *StructuredLogger {
	fields := copyFields(l.fields)
	fields[key] = value
	return &StructuredLogger{
		level:  l.level,
		fields: fields,
		prefix: l.prefix,
	}
}

// WithFields adds multiple fields to the logger context
func (l *StructuredLogger) WithFields(fields map[string]interface{}) *StructuredLogger {
	newFields := copyFields(l.fields)
	for k, v := range fields {
		newFields[k] = v
	}
	return &StructuredLogger{
		level:  l.level,
		fields: newFields,
		prefix: l.prefix,
	}
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(msg string, fields ...core.Field) {
	if l.level <= LevelDebug {
		l.log(LevelDebug, msg, fields...)
	}
}

// Info logs an info message
func (l *StructuredLogger) Info(msg string, fields ...core.Field) {
	if l.level <= LevelInfo {
		l.log(LevelInfo, msg, fields...)
	}
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(msg string, fields ...core.Field) {
	if l.level <= LevelWarn {
		l.log(LevelWarn, msg, fields...)
	}
}

// Error logs an error message
func (l *StructuredLogger) Error(msg string, fields ...core.Field) {
	if l.level <= LevelError {
		l.log(LevelError, msg, fields...)
	}
}

// log performs the actual logging
func (l *StructuredLogger) log(level Level, msg string, fields ...core.Field) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Build the log message
	var logMsg string
	if l.prefix != "" {
		logMsg = fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level, l.prefix, msg)
	} else {
		logMsg = fmt.Sprintf("[%s] [%s] %s", timestamp, level, msg)
	}

	// Add context fields
	if len(l.fields) > 0 {
		logMsg += " " + formatContextFields(l.fields)
	}

	// Add log fields
	if len(fields) > 0 {
		logMsg += " " + formatLogFields(fields)
	}

	// Color the output based on level
	switch level {
	case LevelDebug:
		color.HiBlack(logMsg)
	case LevelInfo:
		color.Cyan(logMsg)
	case LevelWarn:
		color.Yellow(logMsg)
	case LevelError:
		color.Red(logMsg)
	default:
		_, _ = fmt.Fprintln(os.Stdout, logMsg)
	}
}

// Operation represents a timed operation
type Operation struct {
	name      string
	logger    *StructuredLogger
	startTime time.Time
}

// StartOperation begins tracking a timed operation
func (l *StructuredLogger) StartOperation(name string) (*StructuredLogger, func()) {
	op := &Operation{
		name:      name,
		logger:    l.WithField("operation", name),
		startTime: time.Now(),
	}

	op.logger.Debug("operation started")

	return op.logger, func() {
		duration := time.Since(op.startTime)
		op.logger.WithField("duration", duration).Info("operation completed")
	}
}

// formatContextFields formats context fields for logging
func formatContextFields(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	result := "context={"
	first := true
	for k, v := range fields {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%s=%v", k, v)
		first = false
	}
	result += "}"
	return result
}

// formatLogFields formats log fields for output
func formatLogFields(fields []core.Field) string {
	if len(fields) == 0 {
		return ""
	}

	result := "fields={"
	for i, field := range fields {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s=%v", field.Key, field.Value)
	}
	result += "}"
	return result
}

// copyFields creates a copy of a fields map
func copyFields(fields map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range fields {
		result[k] = v
	}
	return result
}

// CheckerLogger provides logging specifically for health checkers
type CheckerLogger struct {
	*StructuredLogger
	checkerName string
}

// NewCheckerLogger creates a logger for a specific checker
func NewCheckerLogger(checkerName string, baseLogger *StructuredLogger) *CheckerLogger {
	return &CheckerLogger{
		StructuredLogger: baseLogger.WithField("checker", checkerName),
		checkerName:      checkerName,
	}
}

// LogResult logs a health check result
func (l *CheckerLogger) LogResult(result core.RepositoryResult) {
	// Count issues and warnings from check results
	var totalIssues, totalWarnings int
	for _, checkResult := range result.CheckResults {
		totalIssues += len(checkResult.Issues)
		totalWarnings += len(checkResult.Warnings)
	}

	fields := map[string]interface{}{
		"status":    string(result.Status),
		"score":     result.Score,
		"max_score": result.MaxScore,
		"issues":    totalIssues,
		"warnings":  totalWarnings,
		"duration":  result.Duration,
		"end_time":  result.EndTime,
	}

	switch result.Status {
	case core.StatusHealthy:
		l.WithFields(fields).Info("check completed successfully")
	case core.StatusWarning:
		l.WithFields(fields).Warn("check completed with warnings")
	case core.StatusCritical:
		l.WithFields(fields).Error("check completed with critical issues")
	default:
		l.WithFields(fields).Info("check completed")
	}
}
