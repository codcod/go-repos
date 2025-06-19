package java_analyzer

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/health/analyzers/common"
)

// JavaAnalyzer implements language-specific analysis for Java code
type JavaAnalyzer struct {
	*common.BaseAnalyzerImpl
}

// NewJavaAnalyzer creates a new Java language analyzer
func NewJavaAnalyzer(walker common.FileWalker, logger core.Logger) *JavaAnalyzer {
	baseAnalyzer := common.NewBaseAnalyzer(
		"java-analyzer",
		"java",
		[]string{".java"},
		[]string{"target/", "build/", ".git/", "bin/", "out/"},
		walker,
		logger,
	)

	return &JavaAnalyzer{
		BaseAnalyzerImpl: baseAnalyzer,
	}
}

// NewJavaAnalyzerWithFS creates a Java analyzer with file system dependency
func NewJavaAnalyzerWithFS(fs core.FileSystem, logger core.Logger) *JavaAnalyzer {
	walker := common.NewDefaultFileWalker()
	return NewJavaAnalyzer(walker, logger)
}

// SupportsComplexity returns whether complexity analysis is supported
func (j *JavaAnalyzer) SupportsComplexity() bool {
	return true
}

// SupportsFunctionLevel returns whether function-level analysis is supported
func (j *JavaAnalyzer) SupportsFunctionLevel() bool {
	return true
}

// CanAnalyze checks if the analyzer can process the given repository
func (j *JavaAnalyzer) CanAnalyze(repo core.Repository) bool {
	// Check if repository has Java files
	return j.hasJavaFiles(repo.Path)
}

// Analyze performs language-specific analysis on the repository
func (j *JavaAnalyzer) Analyze(ctx context.Context, repoPath string, config core.AnalyzerConfig) (*core.AnalysisResult, error) {
	j.Logger().Info("Starting Java analysis", core.Field{Key: "repo", Value: repoPath})

	result := &core.AnalysisResult{
		Language:  j.Language(),
		Files:     make(map[string]*core.FileAnalysis),
		Functions: []core.FunctionInfo{},
		Metrics:   make(map[string]interface{}),
	}

	// Find Java files
	files, err := j.findJavaFiles(repoPath)
	if err != nil {
		return nil, err
	}

	totalComplexity := 0
	totalFunctions := 0
	totalClasses := 0
	maxComplexity := 0

	// Analyze each file
	for _, file := range files {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		fileAnalysis, err := j.analyzeFile(file)
		if err != nil {
			j.Logger().Warn("Failed to analyze file",
				core.Field{Key: "file", Value: file},
				core.Field{Key: "error", Value: err.Error()})
			continue
		}

		result.Files[file] = fileAnalysis

		// Collect function and class information
		for _, fn := range fileAnalysis.Functions {
			result.Functions = append(result.Functions, fn)
			totalFunctions++
			totalComplexity += fn.Complexity
			if fn.Complexity > maxComplexity {
				maxComplexity = fn.Complexity
			}
		}

		totalClasses += len(fileAnalysis.Classes)
	}

	// Calculate metrics
	avgComplexity := 0.0
	if totalFunctions > 0 {
		avgComplexity = float64(totalComplexity) / float64(totalFunctions)
	}

	result.Metrics["total_files"] = len(result.Files)
	result.Metrics["total_classes"] = totalClasses
	result.Metrics["total_functions"] = totalFunctions
	result.Metrics["total_complexity"] = totalComplexity
	result.Metrics["max_complexity"] = maxComplexity
	result.Metrics["average_complexity"] = avgComplexity

	j.Logger().Info("Java analysis completed",
		core.Field{Key: "files", Value: len(result.Files)},
		core.Field{Key: "classes", Value: totalClasses},
		core.Field{Key: "functions", Value: totalFunctions})

	return result, nil
}

// hasJavaFiles checks if the repository contains Java files
func (j *JavaAnalyzer) hasJavaFiles(repoPath string) bool {
	files, err := j.findJavaFiles(repoPath)
	return err == nil && len(files) > 0
}

// findJavaFiles finds all Java source files in the repository
func (j *JavaAnalyzer) findJavaFiles(repoPath string) ([]string, error) {
	return j.FindFiles(repoPath)
}

// analyzeFile analyzes a single Java file
func (j *JavaAnalyzer) analyzeFile(filePath string) (*core.FileAnalysis, error) {
	content, err := j.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	analysis := &core.FileAnalysis{
		Path:      filePath,
		Language:  j.Language(),
		Functions: []core.FunctionInfo{},
		Classes:   []core.ClassInfo{},
		Imports:   []core.ImportInfo{},
		Metrics:   make(map[string]interface{}),
	}

	// Parse the file
	functions, classes, imports := j.parseFile(string(content), filePath)
	analysis.Functions = functions
	analysis.Classes = classes
	analysis.Imports = imports

	// Calculate file-level metrics
	analysis.Metrics["function_count"] = len(analysis.Functions)
	analysis.Metrics["class_count"] = len(analysis.Classes)
	analysis.Metrics["import_count"] = len(analysis.Imports)

	if len(analysis.Functions) > 0 {
		totalComplexity := 0
		for _, fn := range analysis.Functions {
			totalComplexity += fn.Complexity
		}
		analysis.Metrics["average_complexity"] = float64(totalComplexity) / float64(len(analysis.Functions))
	}

	return analysis, nil
}

// parseFile parses a Java file to extract classes, methods, and imports
//
//nolint:gocyclo // Complex parsing logic for Java language requires high cyclomatic complexity
func (j *JavaAnalyzer) parseFile(content, filePath string) ([]core.FunctionInfo, []core.ClassInfo, []core.ImportInfo) {
	var functions []core.FunctionInfo
	var classes []core.ClassInfo
	var imports []core.ImportInfo

	lines := strings.Split(content, "\n")
	inClass := false
	inMethod := false
	var currentClass *core.ClassInfo
	var currentMethod *core.FunctionInfo
	braceLevel := 0

	// Regex patterns
	classPattern := regexp.MustCompile(`^\s*(?:public|private|protected)?\s*(?:abstract|final)?\s*class\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*(?:extends\s+([a-zA-Z_][a-zA-Z0-9_]*))?\s*(?:implements\s+[a-zA-Z0-9_,\s<>]+)?\s*\{?`)
	methodPattern := regexp.MustCompile(`^\s*(?:public|private|protected)?\s*(?:static)?\s*(?:final)?\s*(?:abstract)?\s*[a-zA-Z_<>[\]]+\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\([^)]*\)\s*(?:throws\s+[a-zA-Z0-9_,\s]+)?\s*\{?`)
	importPattern := regexp.MustCompile(`^\s*import\s+(?:static\s+)?([a-zA-Z_][a-zA-Z0-9_.*]+)\s*;`)

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "/*") || strings.HasPrefix(trimmedLine, "*") {
			continue
		}

		// Count braces to track nesting
		braceLevel += strings.Count(line, "{") - strings.Count(line, "}")

		// Check for imports
		if matches := importPattern.FindStringSubmatch(line); matches != nil {
			importPath := matches[1]
			parts := strings.Split(importPath, ".")
			name := parts[len(parts)-1]

			importInfo := core.ImportInfo{
				Name:    name,
				Path:    importPath,
				Line:    lineNum,
				IsLocal: !strings.Contains(importPath, "."),
			}

			imports = append(imports, importInfo)
		}

		// Check for class definitions
		if matches := classPattern.FindStringSubmatch(line); matches != nil {
			className := matches[1]

			// If we were already in a class, finalize the previous one
			if inClass && currentClass != nil {
				classes = append(classes, *currentClass)
			}

			// Start new class
			currentClass = &core.ClassInfo{
				Name:     className,
				File:     filePath,
				Line:     lineNum,
				Language: j.Language(),
				Methods:  []core.FunctionInfo{},
				Fields:   []core.FieldInfo{},
			}

			inClass = true
		}

		// Check for method definitions
		if matches := methodPattern.FindStringSubmatch(line); matches != nil {
			methodName := matches[1]

			// Skip constructors and getters/setters for complexity (they're usually simple)
			isConstructor := currentClass != nil && methodName == currentClass.Name
			isGetterSetter := strings.HasPrefix(methodName, "get") || strings.HasPrefix(methodName, "set") || strings.HasPrefix(methodName, "is")

			// If we were already in a method, finalize the previous one
			if inMethod && currentMethod != nil {
				functions = append(functions, *currentMethod)
				if currentClass != nil {
					currentClass.Methods = append(currentClass.Methods, *currentMethod)
				}
			}

			// Start new method
			currentMethod = &core.FunctionInfo{
				Name: methodName, File: filePath,
				Line:       lineNum,
				Complexity: 1, // Base complexity
				Language:   j.Language(),
			}

			// Constructors and simple getters/setters get lower base complexity
			if isConstructor || isGetterSetter {
				currentMethod.Complexity = 1
			}

			inMethod = true
		} else if inMethod && currentMethod != nil {
			// We're inside a method, calculate complexity
			currentMethod.Complexity += j.calculateLineComplexity(trimmedLine)
		}

		// Check if method or class ended
		if braceLevel == 0 && (inMethod || inClass) {
			if inMethod && currentMethod != nil {
				functions = append(functions, *currentMethod)
				if currentClass != nil {
					currentClass.Methods = append(currentClass.Methods, *currentMethod)
				}
				inMethod = false
				currentMethod = nil
			}

			if inClass && currentClass != nil && !inMethod {
				classes = append(classes, *currentClass)
				inClass = false
				currentClass = nil
			}
		}
	}

	// Don't forget the last method/class if the file ends while in them
	if inMethod && currentMethod != nil {
		functions = append(functions, *currentMethod)
		if currentClass != nil {
			currentClass.Methods = append(currentClass.Methods, *currentMethod)
		}
	}
	if inClass && currentClass != nil {
		classes = append(classes, *currentClass)
	}

	return functions, classes, imports
}

// calculateLineComplexity calculates complexity contribution of a single line
//
//nolint:gocyclo // Complex line-by-line analysis requires high cyclomatic complexity
func (j *JavaAnalyzer) calculateLineComplexity(line string) int {
	complexity := 0
	line = strings.TrimSpace(line)

	// Skip comments and empty lines
	if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
		return 0
	}

	// Decision points that increase McCabe complexity:

	// Conditional branches - more flexible matching
	if strings.Contains(line, "if") && (strings.Contains(line, "(") || strings.Contains(line, " ")) {
		complexity++
		// Count logical operators in conditional statements
		complexity += strings.Count(line, "&&")
		complexity += strings.Count(line, "||")
	}
	if strings.Contains(line, "else if") {
		complexity++
		complexity += strings.Count(line, "&&")
		complexity += strings.Count(line, "||")
	}

	// Loops - more flexible matching
	if strings.Contains(line, "for") && strings.Contains(line, "(") {
		complexity++
	}
	if strings.Contains(line, "while") && strings.Contains(line, "(") {
		complexity++
		complexity += strings.Count(line, "&&")
		complexity += strings.Count(line, "||")
	}
	if strings.Contains(line, "do") && (strings.Contains(line, "{") || strings.Contains(line, " ")) {
		complexity++
	}

	// Enhanced for loop (for-each) - more flexible
	if strings.Contains(line, "for") && strings.Contains(line, " : ") {
		complexity++
	}

	// Switch statements - only case labels
	if strings.Contains(line, "case ") && strings.Contains(line, ":") {
		complexity++
	}

	// Exception handling - more flexible matching
	if strings.Contains(line, "catch") && strings.Contains(line, "(") {
		complexity++
	}

	// Ternary operators
	if strings.Contains(line, "?") && strings.Contains(line, ":") {
		complexity++
	}

	// Lambda expressions with conditions
	if strings.Contains(line, "->") && (strings.Contains(line, "if") || strings.Contains(line, "?")) {
		complexity++
	}

	return complexity
}

// AnalyzeComplexity performs complexity analysis and returns results (ComplexityAnalyzer interface)
func (j *JavaAnalyzer) AnalyzeComplexity(ctx context.Context, repoPath string) (core.ComplexityResult, error) {
	result := core.ComplexityResult{
		Functions: []core.FunctionComplexity{},
	}

	// Find all Java files
	javaFiles, err := j.findJavaFiles(repoPath)
	if err != nil {
		return result, err
	}

	result.TotalFiles = len(javaFiles)

	var totalComplexity int
	var maxComplexity int

	// Analyze each file
	for _, filePath := range javaFiles {
		fileAnalysis, err := j.analyzeFile(filePath)
		if err != nil {
			j.Logger().Warn("Failed to analyze file",
				core.Field{Key: "file", Value: filePath},
				core.Field{Key: "error", Value: err})
			continue
		}

		// Convert function info to complexity format
		for _, fn := range fileAnalysis.Functions {
			// Make file path relative to repository root
			relativePath, err := filepath.Rel(repoPath, fn.File)
			if err != nil {
				// If we can't make it relative, use the original path
				relativePath = fn.File
			}

			complexity := core.FunctionComplexity{
				Name:       fn.Name,
				File:       relativePath,
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
func (j *JavaAnalyzer) AnalyzeFunctions(ctx context.Context, repoPath string) ([]core.FunctionComplexity, error) {
	complexityResult, err := j.AnalyzeComplexity(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	return complexityResult.Functions, nil
}

// DetectPatterns detects patterns in code content (LegacyAnalyzer interface)
func (j *JavaAnalyzer) DetectPatterns(ctx context.Context, content string, patterns []core.Pattern) ([]core.PatternMatch, error) {
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
