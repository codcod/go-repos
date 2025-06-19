package javascript_analyzer

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codcod/repos/internal/core"
)

// JavaScriptAnalyzer implements language-specific analysis for JavaScript/TypeScript code
type JavaScriptAnalyzer struct {
	name       string
	language   string
	extensions []string
	excludes   []string
	filesystem core.FileSystem
	logger     core.Logger
}

// NewJavaScriptAnalyzer creates a new JavaScript/TypeScript language analyzer
func NewJavaScriptAnalyzer(fs core.FileSystem, logger core.Logger) *JavaScriptAnalyzer {
	return &JavaScriptAnalyzer{
		name:       "javascript-analyzer",
		language:   "javascript",
		extensions: []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
		excludes:   []string{"node_modules/", "dist/", "build/", ".git/", "coverage/", ".next/"},
		filesystem: fs,
		logger:     logger,
	}
}

// Name returns the analyzer name
func (js *JavaScriptAnalyzer) Name() string {
	return js.name
}

// Language returns the supported language
func (js *JavaScriptAnalyzer) Language() string {
	return js.language
}

// SupportedExtensions returns supported file extensions
func (js *JavaScriptAnalyzer) SupportedExtensions() []string {
	return js.extensions
}

// CanAnalyze checks if the analyzer can process the given repository
func (js *JavaScriptAnalyzer) CanAnalyze(repo core.Repository) bool {
	// Check if repository has JavaScript/TypeScript files
	return js.hasJavaScriptFiles(repo.Path)
}

// Analyze performs language-specific analysis on the repository
func (js *JavaScriptAnalyzer) Analyze(ctx context.Context, repoPath string, config core.AnalyzerConfig) (*core.AnalysisResult, error) {
	js.logger.Info("Starting JavaScript/TypeScript analysis", core.Field{Key: "repo", Value: repoPath})

	result := &core.AnalysisResult{
		Language:  js.language,
		Files:     make(map[string]*core.FileAnalysis),
		Functions: []core.FunctionInfo{},
		Metrics:   make(map[string]interface{}),
	}

	// Find JavaScript/TypeScript files
	files, err := js.findJavaScriptFiles(repoPath)
	if err != nil {
		return nil, err
	}

	totalComplexity := 0
	totalFunctions := 0
	maxComplexity := 0
	jsFiles := 0
	tsFiles := 0

	// Analyze each file
	for _, file := range files {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		fileAnalysis, err := js.analyzeFile(file)
		if err != nil {
			js.logger.Warn("Failed to analyze file",
				core.Field{Key: "file", Value: file},
				core.Field{Key: "error", Value: err.Error()})
			continue
		}

		result.Files[file] = fileAnalysis

		// Count file types
		if strings.HasSuffix(file, ".ts") || strings.HasSuffix(file, ".tsx") {
			tsFiles++
		} else {
			jsFiles++
		}

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
	result.Metrics["js_files"] = jsFiles
	result.Metrics["ts_files"] = tsFiles
	result.Metrics["total_functions"] = totalFunctions
	result.Metrics["total_complexity"] = totalComplexity
	result.Metrics["max_complexity"] = maxComplexity
	result.Metrics["average_complexity"] = avgComplexity

	js.logger.Info("JavaScript/TypeScript analysis completed",
		core.Field{Key: "files", Value: len(result.Files)},
		core.Field{Key: "js_files", Value: jsFiles},
		core.Field{Key: "ts_files", Value: tsFiles},
		core.Field{Key: "functions", Value: totalFunctions})

	return result, nil
}

// hasJavaScriptFiles checks if the repository contains JavaScript/TypeScript files
func (js *JavaScriptAnalyzer) hasJavaScriptFiles(repoPath string) bool {
	files, err := js.findJavaScriptFiles(repoPath)
	return err == nil && len(files) > 0
}

// findJavaScriptFiles finds all JavaScript/TypeScript source files in the repository
func (js *JavaScriptAnalyzer) findJavaScriptFiles(repoPath string) ([]string, error) {
	var jsFiles []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a JavaScript/TypeScript file
		isJSFile := false
		for _, ext := range js.extensions {
			if strings.HasSuffix(path, ext) {
				isJSFile = true
				break
			}
		}
		if !isJSFile {
			return nil
		}

		// Skip excluded patterns
		relPath, _ := filepath.Rel(repoPath, path)
		for _, exclude := range js.excludes {
			if strings.Contains(relPath, exclude) {
				return nil
			}
		}

		jsFiles = append(jsFiles, path)
		return nil
	})

	return jsFiles, err
}

// analyzeFile analyzes a single JavaScript/TypeScript file
func (js *JavaScriptAnalyzer) analyzeFile(filePath string) (*core.FileAnalysis, error) {
	content, err := os.ReadFile(filePath) //nolint:gosec // File path is from repository analysis
	if err != nil {
		return nil, err
	}

	// Determine specific language variant
	language := "javascript"
	if strings.HasSuffix(filePath, ".ts") || strings.HasSuffix(filePath, ".tsx") {
		language = "typescript"
	}

	analysis := &core.FileAnalysis{
		Path:      filePath,
		Language:  language,
		Functions: []core.FunctionInfo{},
		Imports:   []core.ImportInfo{},
		Metrics:   make(map[string]interface{}),
	}

	// Parse the file
	functions, imports := js.parseFile(string(content), filePath, language)
	analysis.Functions = functions
	analysis.Imports = imports

	// Calculate file-level metrics
	analysis.Metrics["function_count"] = len(analysis.Functions)
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

// parseFile parses a JavaScript/TypeScript file to extract functions and imports
//
//nolint:gocyclo // Complex parsing logic for JavaScript/TypeScript requires high cyclomatic complexity
func (js *JavaScriptAnalyzer) parseFile(content, filePath, language string) ([]core.FunctionInfo, []core.ImportInfo) {
	var functions []core.FunctionInfo
	var imports []core.ImportInfo

	lines := strings.Split(content, "\n")
	inFunction := false
	var currentFunction *core.FunctionInfo
	braceLevel := 0

	// Regex patterns for different function syntaxes
	functionPattern := regexp.MustCompile(`^\s*(?:export\s+)?(?:async\s+)?function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(`)
	arrowFunctionPattern := regexp.MustCompile(`^\s*(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(?:async\s+)?(?:\([^)]*\)\s*|\w+\s*)=>\s*\{?`)
	methodPattern := regexp.MustCompile(`^\s*(?:async\s+)?([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*\{`)

	// Import patterns
	importPattern := regexp.MustCompile(`^\s*import\s+(?:\{([^}]+)\}|([a-zA-Z_$][a-zA-Z0-9_$]*)|(\*\s+as\s+[a-zA-Z_$][a-zA-Z0-9_$]*))\s+from\s+['"]([^'"]+)['"]`)
	requirePattern := regexp.MustCompile(`^\s*(?:const|let|var)\s+(?:\{([^}]+)\}|([a-zA-Z_$][a-zA-Z0-9_$]*))\s*=\s*require\s*\(\s*['"]([^'"]+)['"]\s*\)`)

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "/*") || strings.HasPrefix(trimmedLine, "*") {
			continue
		}

		// Count braces to track nesting (basic approach)
		braceLevel += strings.Count(line, "{") - strings.Count(line, "}")

		// Check for imports (ES6 modules)
		if matches := importPattern.FindStringSubmatch(line); matches != nil {
			modulePath := matches[4]

			// Handle different import types
			if matches[1] != "" {
				// Named imports: import { a, b } from 'module'
				imports = append(imports, js.parseNamedImports(matches[1], modulePath, lineNum)...)
			} else if matches[2] != "" {
				// Default import: import name from 'module'
				importInfo := core.ImportInfo{
					Name:    matches[2],
					Path:    modulePath,
					Line:    lineNum,
					IsLocal: js.isLocalImport(modulePath),
				}
				imports = append(imports, importInfo)
			} else if matches[3] != "" {
				// Namespace import: import * as name from 'module'
				namespaceImport := strings.TrimSpace(strings.Replace(matches[3], "* as ", "", 1))
				importInfo := core.ImportInfo{
					Name:    namespaceImport,
					Path:    modulePath,
					Line:    lineNum,
					IsLocal: js.isLocalImport(modulePath),
				}
				imports = append(imports, importInfo)
			}
		}

		// Check for CommonJS requires
		if matches := requirePattern.FindStringSubmatch(line); matches != nil {
			modulePath := matches[3]

			if matches[1] != "" {
				// Destructured require: const { a, b } = require('module')
				imports = append(imports, js.parseNamedImports(matches[1], modulePath, lineNum)...)
			} else if matches[2] != "" {
				// Simple require: const name = require('module')
				importInfo := core.ImportInfo{
					Name:    matches[2],
					Path:    modulePath,
					Line:    lineNum,
					IsLocal: js.isLocalImport(modulePath),
				}
				imports = append(imports, importInfo)
			}
		}

		// Check for function definitions
		var functionName string
		var isFunction bool

		// Regular function declaration
		if matches := functionPattern.FindStringSubmatch(line); matches != nil {
			functionName = matches[1]
			isFunction = true
		}

		// Arrow function
		if !isFunction {
			if matches := arrowFunctionPattern.FindStringSubmatch(line); matches != nil {
				functionName = matches[1]
				isFunction = true
			}
		}

		// Method definition (in classes or objects)
		if !isFunction {
			if matches := methodPattern.FindStringSubmatch(line); matches != nil {
				functionName = matches[1]
				isFunction = true
			}
		}

		if isFunction {
			// If we were already in a function, finalize the previous one
			if inFunction && currentFunction != nil {
				functions = append(functions, *currentFunction)
			}

			// Start new function
			currentFunction = &core.FunctionInfo{
				Name:       functionName,
				File:       filePath,
				Line:       lineNum,
				Complexity: 1, // Base complexity
				Language:   language,
			}

			inFunction = true
		} else if inFunction && currentFunction != nil {
			// We're inside a function, calculate complexity
			currentFunction.Complexity += js.calculateLineComplexity(trimmedLine)
		}

		// Check if function ended (simplified heuristic)
		if inFunction && braceLevel == 0 && currentFunction != nil {
			functions = append(functions, *currentFunction)
			inFunction = false
			currentFunction = nil
		}
	}

	// Don't forget the last function if the file ends while in a function
	if inFunction && currentFunction != nil {
		functions = append(functions, *currentFunction)
	}

	return functions, imports
}

// parseNamedImports parses named imports like { a, b as c, d }
func (js *JavaScriptAnalyzer) parseNamedImports(namedImports, modulePath string, lineNum int) []core.ImportInfo {
	var imports []core.ImportInfo

	// Split by comma and parse each import
	parts := strings.Split(namedImports, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		importInfo := core.ImportInfo{
			Path:    modulePath,
			Line:    lineNum,
			IsLocal: js.isLocalImport(modulePath),
		}

		// Handle aliases (import { name as alias })
		if strings.Contains(part, " as ") {
			aliasParts := strings.Split(part, " as ")
			if len(aliasParts) == 2 {
				importInfo.Name = strings.TrimSpace(aliasParts[0])
				importInfo.Alias = strings.TrimSpace(aliasParts[1])
			}
		} else {
			importInfo.Name = part
		}

		imports = append(imports, importInfo)
	}

	return imports
}

// isLocalImport determines if an import is local (relative path) or external (npm package)
func (js *JavaScriptAnalyzer) isLocalImport(path string) bool {
	return strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") || strings.HasPrefix(path, "/")
}

// calculateLineComplexity calculates complexity contribution of a single line
//
//nolint:gocyclo // Complex line-by-line analysis requires high cyclomatic complexity
func (js *JavaScriptAnalyzer) calculateLineComplexity(line string) int {
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
	if strings.Contains(line, "do {") {
		complexity++
	}

	// For-in and for-of loops
	if strings.Contains(line, "for (") && (strings.Contains(line, " in ") || strings.Contains(line, " of ")) {
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

	// Array methods with callbacks (simplified detection)
	arrayMethods := []string{".map(", ".filter(", ".reduce(", ".forEach(", ".find(", ".some(", ".every("}
	for _, method := range arrayMethods {
		if strings.Contains(line, method) {
			complexity++
		}
	}

	// Promise chains
	if strings.Contains(line, ".then(") || strings.Contains(line, ".catch(") {
		complexity++
	}

	return complexity
}

// Legacy analyzer interface methods for backward compatibility

// FileExtensions returns supported file extensions (LegacyAnalyzer interface)
func (js *JavaScriptAnalyzer) FileExtensions() []string {
	return js.extensions
}

// SupportsComplexity returns whether complexity analysis is supported (LegacyAnalyzer interface)
func (js *JavaScriptAnalyzer) SupportsComplexity() bool {
	return true
}

// SupportsFunctionLevel returns whether function-level analysis is supported (LegacyAnalyzer interface)
func (js *JavaScriptAnalyzer) SupportsFunctionLevel() bool {
	return true
}

// AnalyzeComplexity performs complexity analysis and returns results (LegacyAnalyzer interface)
func (js *JavaScriptAnalyzer) AnalyzeComplexity(ctx context.Context, repoPath string) (core.ComplexityResult, error) {
	result := core.ComplexityResult{
		Functions: []core.FunctionComplexity{},
	}

	// Find all JavaScript files
	jsFiles, err := js.findJavaScriptFiles(repoPath)
	if err != nil {
		return result, err
	}

	result.TotalFiles = len(jsFiles)

	var totalComplexity int
	var maxComplexity int

	// Analyze each file
	for _, filePath := range jsFiles {
		fileAnalysis, err := js.analyzeFile(filePath)
		if err != nil {
			js.logger.Warn("Failed to analyze file",
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
func (js *JavaScriptAnalyzer) AnalyzeFunctions(ctx context.Context, repoPath string) ([]core.FunctionComplexity, error) {
	complexityResult, err := js.AnalyzeComplexity(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	return complexityResult.Functions, nil
}

// DetectPatterns detects patterns in code content (LegacyAnalyzer interface)
func (js *JavaScriptAnalyzer) DetectPatterns(ctx context.Context, content string, patterns []core.Pattern) ([]core.PatternMatch, error) {
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
