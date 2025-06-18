package java_analyzer

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codcod/repos/internal/core"
)

// JavaAnalyzer implements language-specific analysis for Java code
type JavaAnalyzer struct {
	name       string
	language   string
	extensions []string
	excludes   []string
	filesystem core.FileSystem
	logger     core.Logger
}

// NewJavaAnalyzer creates a new Java language analyzer
func NewJavaAnalyzer(fs core.FileSystem, logger core.Logger) *JavaAnalyzer {
	return &JavaAnalyzer{
		name:       "java-analyzer",
		language:   "java",
		extensions: []string{".java"},
		excludes:   []string{"target/", "build/", ".git/", "bin/", "out/"},
		filesystem: fs,
		logger:     logger,
	}
}

// Name returns the analyzer name
func (j *JavaAnalyzer) Name() string {
	return j.name
}

// Language returns the supported language
func (j *JavaAnalyzer) Language() string {
	return j.language
}

// SupportedExtensions returns supported file extensions
func (j *JavaAnalyzer) SupportedExtensions() []string {
	return j.extensions
}

// CanAnalyze checks if the analyzer can process the given repository
func (j *JavaAnalyzer) CanAnalyze(repo core.Repository) bool {
	// Check if repository has Java files
	return j.hasJavaFiles(repo.Path)
}

// Analyze performs language-specific analysis on the repository
func (j *JavaAnalyzer) Analyze(ctx context.Context, repoPath string, config core.AnalyzerConfig) (*core.AnalysisResult, error) {
	j.logger.Info("Starting Java analysis", core.Field{Key: "repo", Value: repoPath})

	result := &core.AnalysisResult{
		Language:  j.language,
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
			j.logger.Warn("Failed to analyze file",
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

	j.logger.Info("Java analysis completed",
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
	var javaFiles []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a Java file
		if !strings.HasSuffix(path, ".java") {
			return nil
		}

		// Skip excluded patterns
		relPath, _ := filepath.Rel(repoPath, path)
		for _, exclude := range j.excludes {
			if strings.Contains(relPath, exclude) {
				return nil
			}
		}

		javaFiles = append(javaFiles, path)
		return nil
	})

	return javaFiles, err
}

// analyzeFile analyzes a single Java file
func (j *JavaAnalyzer) analyzeFile(filePath string) (*core.FileAnalysis, error) {
	content, err := os.ReadFile(filePath) //nolint:gosec // File path is from repository analysis
	if err != nil {
		return nil, err
	}

	analysis := &core.FileAnalysis{
		Path:      filePath,
		Language:  j.language,
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
				Language: j.language,
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
				Name:       methodName,
				File:       filePath,
				Line:       lineNum,
				Complexity: 1, // Base complexity
				Language:   j.language,
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

	// Conditional branches
	if strings.Contains(line, "if (") {
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

	// Loops
	if strings.Contains(line, "for (") {
		complexity++
	}
	if strings.Contains(line, "while (") {
		complexity++
		complexity += strings.Count(line, "&&")
		complexity += strings.Count(line, "||")
	}
	if strings.Contains(line, "do {") || strings.Contains(line, "do{") {
		complexity++
	}

	// Enhanced for loop (for-each)
	if strings.Contains(line, "for (") && strings.Contains(line, " : ") {
		complexity++
	}

	// Switch statements - only case labels
	if strings.Contains(line, "case ") && strings.Contains(line, ":") {
		complexity++
	}

	// Exception handling - only catch clauses
	if strings.Contains(line, "catch (") {
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
