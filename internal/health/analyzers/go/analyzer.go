package go_analyzer

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/codcod/repos/internal/core"
)

// GoAnalyzer implements language-specific analysis for Go code
type GoAnalyzer struct {
	name       string
	language   string
	extensions []string
	excludes   []string
	filesystem core.FileSystem
	logger     core.Logger
}

// NewGoAnalyzer creates a new Go language analyzer
func NewGoAnalyzer(fs core.FileSystem, logger core.Logger) *GoAnalyzer {
	return &GoAnalyzer{
		name:       "go-analyzer",
		language:   "go",
		extensions: []string{".go"},
		excludes:   []string{"vendor/", "_test.go", ".git/"},
		filesystem: fs,
		logger:     logger,
	}
}

// Name returns the analyzer name
func (g *GoAnalyzer) Name() string {
	return g.name
}

// Language returns the supported language
func (g *GoAnalyzer) Language() string {
	return g.language
}

// SupportedExtensions returns supported file extensions
func (g *GoAnalyzer) SupportedExtensions() []string {
	return g.extensions
}

// CanAnalyze checks if the analyzer can process the given repository
func (g *GoAnalyzer) CanAnalyze(repo core.Repository) bool {
	// Check if repository has Go files
	return g.hasGoFiles(repo.Path)
}

// Analyze performs language-specific analysis on the repository
func (g *GoAnalyzer) Analyze(ctx context.Context, repoPath string, config core.AnalyzerConfig) (*core.AnalysisResult, error) {
	g.logger.Info("Starting Go analysis", core.Field{Key: "path", Value: repoPath})

	result := &core.AnalysisResult{
		Language:  g.language,
		Files:     make(map[string]*core.FileAnalysis),
		Functions: []core.FunctionInfo{},
		Metrics:   make(map[string]interface{}),
	}

	// Find Go files
	files, err := g.findGoFiles(repoPath)
	if err != nil {
		return nil, err
	}

	totalComplexity := 0
	totalFunctions := 0
	maxComplexity := 0

	// Analyze each file
	for _, file := range files {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		fileAnalysis, err := g.analyzeFile(file)
		if err != nil {
			g.logger.Warn("Failed to analyze file",
				core.Field{Key: "file", Value: file},
				core.Field{Key: "error", Value: err.Error()})
			continue
		}

		result.Files[file] = fileAnalysis

		// Collect function information
		for _, fn := range fileAnalysis.Functions {
			result.Functions = append(result.Functions, fn)
			totalFunctions++
			totalComplexity += fn.Complexity
			if fn.Complexity > maxComplexity {
				maxComplexity = fn.Complexity
			}
		}
	}

	// Calculate metrics
	avgComplexity := 0.0
	if totalFunctions > 0 {
		avgComplexity = float64(totalComplexity) / float64(totalFunctions)
	}

	result.Metrics["total_files"] = len(result.Files)
	result.Metrics["total_functions"] = totalFunctions
	result.Metrics["total_complexity"] = totalComplexity
	result.Metrics["max_complexity"] = maxComplexity
	result.Metrics["average_complexity"] = avgComplexity

	g.logger.Info("Go analysis completed",
		core.Field{Key: "files", Value: len(result.Files)},
		core.Field{Key: "functions", Value: totalFunctions})

	return result, nil
}

// hasGoFiles checks if the repository contains Go files
func (g *GoAnalyzer) hasGoFiles(repoPath string) bool {
	files, err := g.findGoFiles(repoPath)
	return err == nil && len(files) > 0
}

// findGoFiles finds all Go source files in the repository
func (g *GoAnalyzer) findGoFiles(repoPath string) ([]string, error) {
	var goFiles []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a Go file
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip excluded patterns
		relPath, _ := filepath.Rel(repoPath, path)
		for _, exclude := range g.excludes {
			if strings.Contains(relPath, exclude) {
				return nil
			}
		}

		goFiles = append(goFiles, path)
		return nil
	})

	return goFiles, err
}

// analyzeFile analyzes a single Go file
func (g *GoAnalyzer) analyzeFile(filePath string) (*core.FileAnalysis, error) {
	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	analysis := &core.FileAnalysis{
		Path:      filePath,
		Language:  g.language,
		Functions: []core.FunctionInfo{},
		Metrics:   make(map[string]interface{}),
	}

	// Analyze functions
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name != nil {
				fnInfo := g.analyzeFunctionDecl(x, fset)
				analysis.Functions = append(analysis.Functions, fnInfo)
			}
		}
		return true
	})

	// Calculate file-level metrics
	analysis.Metrics["function_count"] = len(analysis.Functions)
	if len(analysis.Functions) > 0 {
		totalComplexity := 0
		for _, fn := range analysis.Functions {
			totalComplexity += fn.Complexity
		}
		analysis.Metrics["average_complexity"] = float64(totalComplexity) / float64(len(analysis.Functions))
	}

	return analysis, nil
}

// analyzeFunctionDecl analyzes a function declaration
func (g *GoAnalyzer) analyzeFunctionDecl(fn *ast.FuncDecl, fset *token.FileSet) core.FunctionInfo {
	pos := fset.Position(fn.Pos())

	info := core.FunctionInfo{
		Name:       fn.Name.Name,
		File:       pos.Filename,
		Line:       pos.Line,
		Complexity: 1, // Base complexity
		Language:   g.language,
	}

	// Calculate cyclomatic complexity
	if fn.Body != nil {
		info.Complexity = g.calculateComplexity(fn.Body)
	}

	return info
}

// calculateComplexity calculates cyclomatic complexity for a function body
//
//nolint:gocyclo // Complex parsing logic requires high cyclomatic complexity
func (g *GoAnalyzer) calculateComplexity(body *ast.BlockStmt) int {
	complexity := 1 // Base complexity

	ast.Inspect(body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.IfStmt:
			complexity++
		case *ast.ForStmt:
			complexity++
		case *ast.RangeStmt:
			complexity++
		case *ast.SwitchStmt:
			complexity++
		case *ast.TypeSwitchStmt:
			complexity++
		case *ast.SelectStmt:
			complexity++
		case *ast.CaseClause:
			// Don't count default case
			if n.List != nil {
				complexity++
			}
		case *ast.BinaryExpr:
			// Count logical operators in conditions
			if n.Op == token.LAND || n.Op == token.LOR {
				complexity++
			}
		}
		return true
	})

	return complexity
}

// Legacy analyzer interface methods for backward compatibility

// FileExtensions returns supported file extensions (LegacyAnalyzer interface)
func (g *GoAnalyzer) FileExtensions() []string {
	return g.extensions
}

// SupportsComplexity returns whether complexity analysis is supported (LegacyAnalyzer interface)
func (g *GoAnalyzer) SupportsComplexity() bool {
	return true
}

// SupportsFunctionLevel returns whether function-level analysis is supported (LegacyAnalyzer interface)
func (g *GoAnalyzer) SupportsFunctionLevel() bool {
	return true
}

// AnalyzeComplexity performs complexity analysis and returns results (LegacyAnalyzer interface)
func (g *GoAnalyzer) AnalyzeComplexity(ctx context.Context, repoPath string) (core.ComplexityResult, error) {
	result := core.ComplexityResult{
		Functions: []core.FunctionComplexity{},
	}

	// Find all Go files
	goFiles, err := g.findGoFiles(repoPath)
	if err != nil {
		return result, err
	}

	result.TotalFiles = len(goFiles)

	var totalComplexity int
	var maxComplexity int

	// Analyze each file
	for _, filePath := range goFiles {
		fileAnalysis, err := g.analyzeFile(filePath)
		if err != nil {
			g.logger.Warn("Failed to analyze file",
				core.Field{Key: "file", Value: filePath},
				core.Field{Key: "error", Value: err})
			continue
		}

		// Convert function info to complexity format
		for _, fn := range fileAnalysis.Functions {
			complexity := core.FunctionComplexity{
				Name:       fn.Name,
				File:       fn.File,
				Line:       fn.Line,
				Complexity: fn.Complexity,
			}

			result.Functions = append(result.Functions, complexity)
			totalComplexity += fn.Complexity

			if fn.Complexity > maxComplexity {
				maxComplexity = fn.Complexity
			}
		}

		result.TotalFunctions += len(fileAnalysis.Functions)
	}

	// Calculate average complexity
	if result.TotalFunctions > 0 {
		result.AverageComplexity = float64(totalComplexity) / float64(result.TotalFunctions)
	}
	result.MaxComplexity = maxComplexity

	return result, nil
}

// AnalyzeFunctions performs function-level analysis (LegacyAnalyzer interface)
func (g *GoAnalyzer) AnalyzeFunctions(ctx context.Context, repoPath string) ([]core.FunctionComplexity, error) {
	complexityResult, err := g.AnalyzeComplexity(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	return complexityResult.Functions, nil
}

// DetectPatterns detects patterns in code content (LegacyAnalyzer interface)
func (g *GoAnalyzer) DetectPatterns(ctx context.Context, content string, patterns []core.Pattern) ([]core.PatternMatch, error) {
	// Basic pattern detection implementation
	var matches []core.PatternMatch

	for _, pattern := range patterns {
		// Check each pattern string in the pattern
		for _, patternStr := range pattern.Patterns {
			// Simple string matching for now - could be enhanced with regex
			if strings.Contains(content, patternStr) {
				match := core.PatternMatch{
					Pattern: pattern,
					Location: core.Location{
						File:   "", // Would need file path context
						Line:   1,  // Would need line-by-line analysis for accurate line numbers
						Column: 1,
					},
					MatchText: patternStr,
					Context:   content, // Could be trimmed to surrounding context
				}
				matches = append(matches, match)
			}
		}
	}

	return matches, nil
}
