// Package common provides shared interfaces and types for health analyzers
package common

import (
	"context"

	"github.com/codcod/repos/internal/core"
)

// BaseAnalyzer defines the basic functionality all analyzers must implement
type BaseAnalyzer interface {
	// Language returns the programming language name (e.g., "go", "python", "java")
	Language() string

	// FileExtensions returns supported file extensions (e.g., []string{".go"})
	FileExtensions() []string

	// CanAnalyze checks if the analyzer can process the given repository
	CanAnalyze(repo core.Repository) bool
}

// ComplexityAnalyzer defines functionality for cyclomatic complexity analysis
type ComplexityAnalyzer interface {
	BaseAnalyzer

	// SupportsComplexity returns whether complexity analysis is supported
	SupportsComplexity() bool

	// AnalyzeComplexity performs complexity analysis and returns results
	AnalyzeComplexity(ctx context.Context, repoPath string) (core.ComplexityResult, error)
}

// FunctionAnalyzer defines functionality for function-level analysis
type FunctionAnalyzer interface {
	BaseAnalyzer

	// SupportsFunctionLevel returns whether function-level analysis is supported
	SupportsFunctionLevel() bool

	// AnalyzeFunctions performs function-level analysis
	AnalyzeFunctions(ctx context.Context, repoPath string) ([]core.FunctionComplexity, error)
}

// PatternAnalyzer defines functionality for pattern detection
type PatternAnalyzer interface {
	BaseAnalyzer

	// DetectPatterns detects patterns in code content
	DetectPatterns(ctx context.Context, content string, patterns []core.Pattern) ([]core.PatternMatch, error)
}

// FullAnalyzer combines all analyzer capabilities
type FullAnalyzer interface {
	ComplexityAnalyzer
	FunctionAnalyzer
	PatternAnalyzer
}
