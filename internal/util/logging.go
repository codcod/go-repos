package util

import (
	"github.com/codcod/repos/internal/config"
	"github.com/fatih/color"
)

// Logger defines structured logging functions for repositories
type Logger struct {
	repoColor func(a ...interface{}) string
}

// NewLogger creates a new repository logger
func NewLogger() *Logger {
	return &Logger{
		repoColor: color.New(color.FgCyan, color.Bold).SprintFunc(),
	}
}

// Success logs a success message with standardized formatting
func (l *Logger) Success(repo config.Repository, format string, args ...interface{}) {
	color.Green("%s | "+format, append([]interface{}{l.repoColor(repo.Name)}, args...)...)
}

// Error logs an error message with standardized formatting
func (l *Logger) Error(repo config.Repository, format string, args ...interface{}) {
	color.Red("%s | "+format, append([]interface{}{l.repoColor(repo.Name)}, args...)...)
}

// Info logs an informational message with standardized formatting
func (l *Logger) Info(repo config.Repository, format string, args ...interface{}) {
	color.Cyan("%s | "+format, append([]interface{}{l.repoColor(repo.Name)}, args...)...)
}

// Warn logs a warning message with standardized formatting
func (l *Logger) Warn(repo config.Repository, format string, args ...interface{}) {
	color.Yellow("%s | "+format, append([]interface{}{l.repoColor(repo.Name)}, args...)...)
}
