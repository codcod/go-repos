package health

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ComplexityAnalyzer analyzes cyclomatic complexity for different languages
type ComplexityAnalyzer struct {
	pathValidator *PathValidator
	fsHelper      *FileSystemHelper
}

// NewComplexityAnalyzer creates a new complexity analyzer
func NewComplexityAnalyzer() *ComplexityAnalyzer {
	return &ComplexityAnalyzer{
		pathValidator: NewPathValidator(),
		fsHelper:      NewFileSystemHelper(),
	}
}

// AnalyzeComplexity analyzes cyclomatic complexity for all supported languages
func (a *ComplexityAnalyzer) AnalyzeComplexity(repoPath string) []ComplexityResult {
	var results []ComplexityResult

	// Analyze different language files
	results = append(results, a.analyzeGoFiles(repoPath)...)
	results = append(results, a.analyzeJavaFiles(repoPath)...)
	results = append(results, a.analyzeJavaScriptFiles(repoPath)...)
	results = append(results, a.analyzePythonFiles(repoPath)...)
	results = append(results, a.analyzeCFiles(repoPath)...)

	return results
}

// CalculateMetrics calculates complexity metrics from results
func (a *ComplexityAnalyzer) CalculateMetrics(results []ComplexityResult) ComplexityMetrics {
	metrics := ComplexityMetrics{
		totalFiles: len(results),
	}

	for _, result := range results {
		metrics.totalComplexity += result.Complexity
		if result.Complexity > metrics.maxComplexity {
			metrics.maxComplexity = result.Complexity
		}
		if result.Complexity > 10 {
			metrics.highComplexityFiles++
		}
		if result.Complexity > 20 {
			metrics.veryHighComplexityFiles++
		}
	}

	if metrics.totalFiles > 0 {
		metrics.avgComplexity = metrics.totalComplexity / metrics.totalFiles
	}

	return metrics
}

// AnalyzeRepository analyzes the complexity of an entire repository
func (a *ComplexityAnalyzer) AnalyzeRepository(repoPath string) ComplexityMetrics {
	results := a.AnalyzeComplexity(repoPath)
	return a.CalculateMetrics(results)
}

// analyzeGoFiles analyzes Go files for cyclomatic complexity
func (a *ComplexityAnalyzer) analyzeGoFiles(repoPath string) []ComplexityResult {
	var results []ComplexityResult

	files, err := a.fsHelper.FindFiles(repoPath, GoFilePattern)
	if err != nil {
		return results
	}

	for _, file := range files {
		if strings.Contains(file, "vendor/") || strings.Contains(file, "_test.go") {
			continue
		}

		complexity := a.calculateGoComplexity(file)
		if complexity > 0 {
			results = append(results, ComplexityResult{
				File:       file,
				Language:   "Go",
				Complexity: complexity,
				Functions:  []FunctionComplexity{}, // TODO: Implement function-level analysis
			})
		}
	}

	return results
}

// analyzeJavaFiles analyzes Java files for cyclomatic complexity
func (a *ComplexityAnalyzer) analyzeJavaFiles(repoPath string) []ComplexityResult {
	var results []ComplexityResult

	files, err := a.fsHelper.FindFiles(repoPath, JavaFilePattern)
	if err != nil {
		return results
	}

	for _, file := range files {
		complexity := a.calculateJavaComplexity(file)
		if complexity > 0 {
			results = append(results, ComplexityResult{
				File:       file,
				Language:   "Java",
				Complexity: complexity,
				Functions:  []FunctionComplexity{}, // TODO: Implement function-level analysis
			})
		}
	}

	return results
}

// analyzeJavaScriptFiles analyzes JavaScript/TypeScript files for cyclomatic complexity
func (a *ComplexityAnalyzer) analyzeJavaScriptFiles(repoPath string) []ComplexityResult {
	var results []ComplexityResult

	files, err := a.fsHelper.FindFiles(repoPath, JavaScriptFilePattern)
	if err != nil {
		return results
	}

	for _, file := range files {
		complexity := a.calculateJSComplexity(file)
		language := "JavaScript"
		if strings.HasSuffix(file, ".ts") || strings.HasSuffix(file, ".tsx") {
			language = "TypeScript"
		}

		if complexity > 0 {
			results = append(results, ComplexityResult{
				File:       file,
				Language:   language,
				Complexity: complexity,
				Functions:  []FunctionComplexity{}, // TODO: Implement function-level analysis
			})
		}
	}

	return results
}

// analyzePythonFiles analyzes Python files for cyclomatic complexity
func (a *ComplexityAnalyzer) analyzePythonFiles(repoPath string) []ComplexityResult {
	var results []ComplexityResult

	files, err := a.fsHelper.FindFiles(repoPath, PythonFilePattern)
	if err != nil {
		return results
	}

	for _, file := range files {
		complexity := a.calculatePythonComplexity(file)
		if complexity > 0 {
			results = append(results, ComplexityResult{
				File:       file,
				Language:   "Python",
				Complexity: complexity,
				Functions:  []FunctionComplexity{}, // TODO: Implement function-level analysis
			})
		}
	}

	return results
}

// analyzeCFiles analyzes C/C++ files for cyclomatic complexity
func (a *ComplexityAnalyzer) analyzeCFiles(repoPath string) []ComplexityResult {
	var results []ComplexityResult

	files, err := a.fsHelper.FindFiles(repoPath, CFilePattern)
	if err != nil {
		return results
	}

	for _, file := range files {
		complexity := a.calculateCComplexity(file)
		language := "C"
		if strings.HasSuffix(file, ".cpp") || strings.HasSuffix(file, ".hpp") {
			language = "C++"
		}

		if complexity > 0 {
			results = append(results, ComplexityResult{
				File:       file,
				Language:   language,
				Complexity: complexity,
				Functions:  []FunctionComplexity{}, // TODO: Implement function-level analysis
			})
		}
	}

	return results
}

// calculateGoComplexity calculates cyclomatic complexity for Go files
func (a *ComplexityAnalyzer) calculateGoComplexity(filePath string) int {
	if !a.pathValidator.IsValidFilePath(filePath) {
		return 0
	}

	content, err := os.ReadFile(filePath) // #nosec G304 - filePath is validated
	if err != nil {
		return 0
	}

	text := string(content)
	complexity := 1 // Base complexity

	keywords := []string{
		"if ", "else if", "for ", "switch ", "case ", "default:",
		"&&", "||", "select ", "go ", "defer ",
	}

	for _, keyword := range keywords {
		complexity += strings.Count(text, keyword)
	}

	return complexity
}

// calculateJavaComplexity calculates cyclomatic complexity for Java files
func (a *ComplexityAnalyzer) calculateJavaComplexity(filePath string) int {
	if !a.pathValidator.IsValidFilePath(filePath) {
		return 0
	}

	content, err := os.ReadFile(filePath) // #nosec G304 - filePath is validated
	if err != nil {
		return 0
	}

	lines := strings.Split(string(content), "\n")
	complexity := 1 // Base complexity

	for _, line := range lines {
		complexity += a.countJavaComplexityIndicators(line)
	}

	return complexity
}

// calculateJSComplexity calculates cyclomatic complexity for JavaScript/TypeScript files
func (a *ComplexityAnalyzer) calculateJSComplexity(filePath string) int {
	if !a.pathValidator.IsValidFilePath(filePath) {
		return 0
	}

	content, err := os.ReadFile(filePath) // #nosec G304 - filePath is validated
	if err != nil {
		return 0
	}

	lines := strings.Split(string(content), "\n")
	complexity := 1 // Base complexity

	for _, line := range lines {
		complexity += a.countJavaScriptComplexityIndicators(line)
	}

	return complexity
}

// calculatePythonComplexity calculates cyclomatic complexity for Python files
func (a *ComplexityAnalyzer) calculatePythonComplexity(filePath string) int {
	if !a.pathValidator.IsValidFilePath(filePath) {
		return 0
	}

	content, err := os.ReadFile(filePath) // #nosec G304 - filePath is validated
	if err != nil {
		return 0
	}

	text := string(content)
	complexity := 1 // Base complexity

	keywords := []string{
		"if ", "elif ", "while ", "for ", "except ", "with ",
		"and ", "or ", "lambda ", "assert ",
	}

	for _, keyword := range keywords {
		complexity += strings.Count(text, keyword)
	}

	return complexity
}

// calculateCComplexity calculates cyclomatic complexity for C/C++ files
func (a *ComplexityAnalyzer) calculateCComplexity(filePath string) int {
	if !a.pathValidator.IsValidFilePath(filePath) {
		return 0
	}

	content, err := os.ReadFile(filePath) // #nosec G304 - filePath is validated
	if err != nil {
		return 0
	}

	text := string(content)
	complexity := 1 // Base complexity

	keywords := []string{
		"if (", "else if", "while (", "for (", "switch (", "case ", "default:",
		"&&", "||", "?", ":",
	}

	for _, keyword := range keywords {
		complexity += strings.Count(text, keyword)
	}

	return complexity
}

// ComplexityResult represents the complexity analysis result for a single file
type ComplexityResult struct {
	File       string
	Language   string
	Complexity int
	Functions  []FunctionComplexity
}

// FunctionComplexity represents complexity metrics for a single function
type FunctionComplexity struct {
	Name       string
	Complexity int
	StartLine  int
	EndLine    int
	File       string
}

// ComplexityMetrics represents aggregated complexity metrics
type ComplexityMetrics struct {
	totalFiles              int
	totalComplexity         int
	maxComplexity           int
	avgComplexity           int
	highComplexityFiles     int
	veryHighComplexityFiles int
}

// FormatSummary formats a summary of complexity metrics
func (m ComplexityMetrics) FormatSummary() string {
	if m.totalFiles == 0 {
		return "No files analyzed"
	}

	summary := fmt.Sprintf("Analyzed %d files, average complexity: %.1f, max complexity: %d",
		m.totalFiles, float64(m.avgComplexity), m.maxComplexity)

	if m.highComplexityFiles > 0 {
		summary += fmt.Sprintf(", %d files with high complexity (>10)", m.highComplexityFiles)
	}

	if m.veryHighComplexityFiles > 0 {
		summary += fmt.Sprintf(", %d files with very high complexity (>20)", m.veryHighComplexityFiles)
	}

	return summary
}

// GetAverageComplexity returns the average complexity across all files
func (m ComplexityMetrics) GetAverageComplexity() float64 {
	return float64(m.avgComplexity)
}

// AnalyzeRepositoryDetailed analyzes the complexity of an entire repository with function-level details
func (a *ComplexityAnalyzer) AnalyzeRepositoryDetailed(repoPath string, maxComplexity int) ComplexityDetailedReport {
	results := a.AnalyzeComplexityDetailed(repoPath)

	// If detailed analysis returns no results, fall back to file-level analysis
	if len(results) == 0 {
		fileResults := a.AnalyzeComplexity(repoPath)
		// Convert file-level results to detailed format for consistency
		for _, fileResult := range fileResults {
			if fileResult.Complexity > maxComplexity {
				// Create a synthetic function representing the whole file
				syntheticFunction := FunctionComplexity{
					Name:       "file-level",
					File:       fileResult.File,
					Complexity: fileResult.Complexity,
					StartLine:  1,
					EndLine:    -1, // Indicates whole file
				}

				results = append(results, ComplexityResult{
					File:       fileResult.File,
					Language:   fileResult.Language,
					Complexity: fileResult.Complexity,
					Functions:  []FunctionComplexity{syntheticFunction},
				})
			} else {
				// Include file in results even if complexity is below threshold for metrics
				results = append(results, ComplexityResult{
					File:       fileResult.File,
					Language:   fileResult.Language,
					Complexity: fileResult.Complexity,
					Functions:  []FunctionComplexity{},
				})
			}
		}
	}

	metrics := a.CalculateMetrics(results)

	// Filter functions that exceed the complexity threshold
	var highComplexityFunctions []FunctionComplexity
	for _, result := range results {
		for _, function := range result.Functions {
			if function.Complexity > maxComplexity {
				highComplexityFunctions = append(highComplexityFunctions, function)
			}
		}
	}

	return ComplexityDetailedReport{
		Metrics:                 metrics,
		Results:                 results,
		HighComplexityFunctions: highComplexityFunctions,
		MaxComplexity:           maxComplexity,
	}
}

// AnalyzeComplexityDetailed analyzes cyclomatic complexity with function-level details
func (a *ComplexityAnalyzer) AnalyzeComplexityDetailed(repoPath string) []ComplexityResult {
	var results []ComplexityResult

	// Analyze different language files with detailed function analysis
	results = append(results, a.analyzeGoFilesDetailed(repoPath)...)
	results = append(results, a.analyzeJavaFilesDetailed(repoPath)...)
	results = append(results, a.analyzeJavaScriptFilesDetailed(repoPath)...)
	results = append(results, a.analyzePythonFilesDetailed(repoPath)...)
	results = append(results, a.analyzeCFilesDetailed(repoPath)...)

	return results
}

// analyzeGoFilesDetailed analyzes Go files for cyclomatic complexity with function-level details
func (a *ComplexityAnalyzer) analyzeGoFilesDetailed(repoPath string) []ComplexityResult {
	var results []ComplexityResult

	files, err := a.fsHelper.FindFiles(repoPath, GoFilePattern)
	if err != nil {
		return results
	}

	for _, file := range files {
		if strings.Contains(file, "vendor/") || strings.Contains(file, "_test.go") {
			continue
		}

		functions := a.analyzeGoFunctions(file)
		if len(functions) > 0 {
			totalComplexity := 0
			for _, fn := range functions {
				totalComplexity += fn.Complexity
			}

			results = append(results, ComplexityResult{
				File:       file,
				Language:   "Go",
				Complexity: totalComplexity,
				Functions:  functions,
			})
		}
	}

	return results
}

// analyzeJavaFilesDetailed analyzes Java files for cyclomatic complexity with function-level details
func (a *ComplexityAnalyzer) analyzeJavaFilesDetailed(repoPath string) []ComplexityResult {
	var results []ComplexityResult

	files, err := a.fsHelper.FindFiles(repoPath, JavaFilePattern)
	if err != nil {
		return results
	}

	for _, file := range files {
		functions := a.analyzeJavaFunctions(file)
		if len(functions) > 0 {
			totalComplexity := 0
			for _, fn := range functions {
				totalComplexity += fn.Complexity
			}

			results = append(results, ComplexityResult{
				File:       file,
				Language:   "Java",
				Complexity: totalComplexity,
				Functions:  functions,
			})
		}
	}

	return results
}

// analyzeJavaScriptFilesDetailed analyzes JavaScript files for cyclomatic complexity with function-level details
func (a *ComplexityAnalyzer) analyzeJavaScriptFilesDetailed(repoPath string) []ComplexityResult {
	var results []ComplexityResult

	files, err := a.fsHelper.FindFiles(repoPath, JavaScriptFilePattern)
	if err != nil {
		return results
	}

	for _, file := range files {
		if strings.Contains(file, "node_modules/") {
			continue
		}

		functions := a.analyzeJavaScriptFunctions(file)
		if len(functions) > 0 {
			totalComplexity := 0
			for _, fn := range functions {
				totalComplexity += fn.Complexity
			}

			results = append(results, ComplexityResult{
				File:       file,
				Language:   "JavaScript",
				Complexity: totalComplexity,
				Functions:  functions,
			})
		}
	}

	return results
}

// analyzePythonFilesDetailed analyzes Python files for cyclomatic complexity with function-level details
func (a *ComplexityAnalyzer) analyzePythonFilesDetailed(repoPath string) []ComplexityResult {
	var results []ComplexityResult

	files, err := a.fsHelper.FindFiles(repoPath, PythonFilePattern)
	if err != nil {
		return results
	}

	for _, file := range files {
		functions := a.analyzePythonFunctions(file)
		if len(functions) > 0 {
			totalComplexity := 0
			for _, fn := range functions {
				totalComplexity += fn.Complexity
			}

			results = append(results, ComplexityResult{
				File:       file,
				Language:   "Python",
				Complexity: totalComplexity,
				Functions:  functions,
			})
		}
	}

	return results
}

// analyzeCFilesDetailed analyzes C files for cyclomatic complexity with function-level details
func (a *ComplexityAnalyzer) analyzeCFilesDetailed(repoPath string) []ComplexityResult {
	var results []ComplexityResult

	files, err := a.fsHelper.FindFiles(repoPath, CFilePattern)
	if err != nil {
		return results
	}

	for _, file := range files {
		functions := a.analyzeCFunctions(file)
		if len(functions) > 0 {
			totalComplexity := 0
			for _, fn := range functions {
				totalComplexity += fn.Complexity
			}

			results = append(results, ComplexityResult{
				File:       file,
				Language:   "C",
				Complexity: totalComplexity,
				Functions:  functions,
			})
		}
	}

	return results
}

// ComplexityDetailedReport represents a detailed complexity analysis report
type ComplexityDetailedReport struct {
	Metrics                 ComplexityMetrics
	Results                 []ComplexityResult
	HighComplexityFunctions []FunctionComplexity
	MaxComplexity           int
}

// FormatDetailedReport formats the detailed complexity report for display
func (r ComplexityDetailedReport) FormatDetailedReport() string {
	if len(r.HighComplexityFunctions) == 0 {
		return fmt.Sprintf("No functions exceed the complexity threshold of %d\n%s",
			r.MaxComplexity, r.Metrics.FormatSummary())
	}

	report := fmt.Sprintf("Functions exceeding complexity threshold of %d:\n\n", r.MaxComplexity)

	for _, function := range r.HighComplexityFunctions {
		report += fmt.Sprintf("  %s() - Complexity: %d (Lines: %d-%d)\n",
			function.Name, function.Complexity, function.StartLine, function.EndLine)
		if function.File != "" {
			report += fmt.Sprintf("    File: %s\n", function.File)
		}
		report += "\n"
	}

	report += fmt.Sprintf("Summary: %s\n", r.Metrics.FormatSummary())
	return report
}

// analyzeGoFunctions analyzes individual Go functions for complexity
//
//nolint:gocyclo // Go function parsing - inherently complex due to language syntax
func (a *ComplexityAnalyzer) analyzeGoFunctions(filePath string) []FunctionComplexity {
	var functions []FunctionComplexity

	content, err := a.safeReadFile(filePath)
	if err != nil {
		return functions
	}

	lines := strings.Split(string(content), "\n")
	inFunction := false
	var currentFunction FunctionComplexity

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Detect function start (simplified pattern)
		if strings.HasPrefix(trimmedLine, "func ") && strings.Contains(trimmedLine, "(") {
			if inFunction {
				// Save previous function
				currentFunction.EndLine = lineNum - 1
				functions = append(functions, currentFunction)
			}

			// Extract function name
			funcName := a.extractGoFunctionName(trimmedLine)
			currentFunction = FunctionComplexity{
				Name:       funcName,
				StartLine:  lineNum,
				File:       filePath,
				Complexity: 1, // Base complexity
			}
			inFunction = true
		}

		// Count complexity indicators
		if inFunction {
			currentFunction.Complexity += a.countGoComplexityIndicators(trimmedLine)
		}

		// Detect function end (simplified - when we hit a closing brace at the beginning of line)
		if inFunction && trimmedLine == "}" && a.isGoFunctionEnd(lines, i) {
			currentFunction.EndLine = lineNum
			functions = append(functions, currentFunction)
			inFunction = false
		}
	}

	// Handle case where file ends while in a function
	if inFunction {
		currentFunction.EndLine = len(lines)
		functions = append(functions, currentFunction)
	}

	return functions
}

// extractGoFunctionName extracts the function name from a Go function declaration
func (a *ComplexityAnalyzer) extractGoFunctionName(line string) string {
	// Remove "func " prefix
	line = strings.TrimPrefix(line, "func ")

	// Handle method receivers like "func (r *Receiver) MethodName("
	if strings.HasPrefix(line, "(") {
		// Find the end of the receiver
		parenEnd := strings.Index(line, ")")
		if parenEnd != -1 {
			line = line[parenEnd+1:]
			line = strings.TrimSpace(line)
		}
	}

	// Extract function name (everything before the opening parenthesis)
	parenIndex := strings.Index(line, "(")
	if parenIndex != -1 {
		return strings.TrimSpace(line[:parenIndex])
	}

	return "unknown"
}

// countGoComplexityIndicators counts cyclomatic complexity indicators in a Go line
func (a *ComplexityAnalyzer) countGoComplexityIndicators(line string) int {
	complexity := 0

	// Control flow keywords that increase complexity
	keywords := []string{"if ", "else if", "for ", "switch ", "case ", "&&", "||", "select "}

	for _, keyword := range keywords {
		if strings.Contains(line, keyword) {
			complexity++
		}
	}

	return complexity
}

// isGoFunctionEnd checks if a closing brace represents the end of a function
func (a *ComplexityAnalyzer) isGoFunctionEnd(lines []string, index int) bool {
	// This is a simplified check - in practice, you'd need proper Go AST parsing
	// For now, we assume any standalone closing brace could be a function end
	return true
}

// analyzeJavaFunctions analyzes individual Java functions for complexity
//
//nolint:gocyclo // Java function parsing - inherently complex due to language syntax
func (a *ComplexityAnalyzer) analyzeJavaFunctions(filePath string) []FunctionComplexity {
	var functions []FunctionComplexity

	content, err := a.safeReadFile(filePath)
	if err != nil {
		return functions
	}

	lines := strings.Split(string(content), "\n")
	inFunction := false
	var currentFunction FunctionComplexity
	braceLevel := 0
	functionStartBraceLevel := 0

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "/*") || strings.HasPrefix(trimmedLine, "*") {
			continue
		}

		// Update brace level
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")
		braceLevel += openBraces - closeBraces

		// Detect Java method start (simplified - looking for method signatures)
		if a.isJavaMethodSignature(trimmedLine) && !inFunction {
			// Extract method name
			methodName := a.extractJavaMethodName(trimmedLine)
			currentFunction = FunctionComplexity{
				Name:       methodName,
				StartLine:  lineNum,
				File:       filePath,
				Complexity: 1, // Base complexity
			}
			functionStartBraceLevel = braceLevel - openBraces // Before the opening brace
			inFunction = true
			continue
		}

		// Count complexity indicators only if we're inside a function
		if inFunction {
			currentFunction.Complexity += a.countJavaComplexityIndicators(trimmedLine)
		}

		// Detect function end (when brace level returns to function start level)
		if inFunction && braceLevel == functionStartBraceLevel {
			currentFunction.EndLine = lineNum
			functions = append(functions, currentFunction)
			inFunction = false
		}
	}

	// Handle case where file ends while in a function
	if inFunction {
		currentFunction.EndLine = len(lines)
		functions = append(functions, currentFunction)
	}

	return functions
}

// isJavaMethodSignature checks if a line contains a Java method signature
//
//nolint:gocyclo // Java method signature detection - multiple access modifiers and patterns
func (a *ComplexityAnalyzer) isJavaMethodSignature(line string) bool {
	line = strings.TrimSpace(line)

	// Skip constructors, interfaces, and class declarations
	if strings.Contains(line, "class ") || strings.Contains(line, "interface ") || strings.Contains(line, "enum ") {
		return false
	}

	// Check for method patterns: visibility modifier + return type + method name + parameters
	// This is a simplified check - a proper implementation would use AST parsing
	hasVisibility := strings.Contains(line, "public ") || strings.Contains(line, "private ") || strings.Contains(line, "protected ") || strings.Contains(line, "static ")
	hasParentheses := strings.Contains(line, "(") && strings.Contains(line, ")")
	hasMethodBodyStart := strings.Contains(line, "{") || strings.HasSuffix(line, ")")

	return hasVisibility && hasParentheses && hasMethodBodyStart && !strings.Contains(line, "=")
}

// extractJavaMethodName extracts the method name from a Java method signature
func (a *ComplexityAnalyzer) extractJavaMethodName(line string) string {
	line = strings.TrimSpace(line)

	// Find the method name by looking for the pattern: name(
	parenIndex := strings.Index(line, "(")
	if parenIndex == -1 {
		return "unknown"
	}

	// Work backwards from the parenthesis to find the method name
	beforeParen := line[:parenIndex]
	words := strings.Fields(beforeParen)

	if len(words) > 0 {
		return words[len(words)-1]
	}

	return "unknown"
}

// analyzeJavaScriptFunctions analyzes individual JavaScript functions for complexity
//
//nolint:gocyclo // JavaScript function parsing - inherently complex due to language syntax
func (a *ComplexityAnalyzer) analyzeJavaScriptFunctions(filePath string) []FunctionComplexity {
	var functions []FunctionComplexity

	content, err := a.safeReadFile(filePath)
	if err != nil {
		return functions
	}

	lines := strings.Split(string(content), "\n")
	inFunction := false
	var currentFunction FunctionComplexity
	braceLevel := 0
	functionStartBraceLevel := 0

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "/*") || strings.HasPrefix(trimmedLine, "*") {
			continue
		}

		// Update brace level
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")
		braceLevel += openBraces - closeBraces

		// Detect JavaScript function start
		if a.isJavaScriptFunctionSignature(trimmedLine) && !inFunction {
			// Extract function name
			funcName := a.extractJavaScriptFunctionName(trimmedLine)
			currentFunction = FunctionComplexity{
				Name:       funcName,
				StartLine:  lineNum,
				File:       filePath,
				Complexity: 1, // Base complexity
			}
			functionStartBraceLevel = braceLevel - openBraces // Before the opening brace
			inFunction = true
			continue
		}

		// Count complexity indicators only if we're inside a function
		if inFunction {
			currentFunction.Complexity += a.countJavaScriptComplexityIndicators(trimmedLine)
		}

		// Detect function end (when brace level returns to function start level)
		if inFunction && braceLevel == functionStartBraceLevel {
			currentFunction.EndLine = lineNum
			functions = append(functions, currentFunction)
			inFunction = false
		}
	}

	// Handle case where file ends while in a function
	if inFunction {
		currentFunction.EndLine = len(lines)
		functions = append(functions, currentFunction)
	}

	return functions
}

// isJavaScriptFunctionSignature checks if a line contains a JavaScript function signature
//
//nolint:gocyclo // JavaScript function signature detection - multiple function declaration patterns
func (a *ComplexityAnalyzer) isJavaScriptFunctionSignature(line string) bool {
	line = strings.TrimSpace(line)

	// Check for various JavaScript function patterns
	patterns := []string{
		"function ",              // function declaration
		"const ", "let ", "var ", // function expressions (will need additional checks)
		"async function ", // async function
		"() => {",         // arrow function
		"=> {",            // arrow function
	}

	for _, pattern := range patterns {
		if strings.Contains(line, pattern) {
			// Additional checks for function expressions
			if pattern == "const " || pattern == "let " || pattern == "var " {
				// Check if it's actually a function assignment
				if strings.Contains(line, "function") || strings.Contains(line, "=>") {
					return true
				}
			} else {
				return true
			}
		}
	}

	// Check for method definitions in objects/classes
	if strings.Contains(line, "(") && strings.Contains(line, ")") && strings.Contains(line, "{") {
		// Could be a method definition
		return true
	}

	return false
}

// extractJavaScriptFunctionName extracts the function name from a JavaScript function signature
func (a *ComplexityAnalyzer) extractJavaScriptFunctionName(line string) string {
	line = strings.TrimSpace(line)

	// Handle different function patterns
	if strings.HasPrefix(line, "function ") {
		// function declaration: function functionName(
		line = strings.TrimPrefix(line, "function ")
		parenIndex := strings.Index(line, "(")
		if parenIndex != -1 {
			return strings.TrimSpace(line[:parenIndex])
		}
	}

	if strings.HasPrefix(line, "async function ") {
		// async function declaration
		line = strings.TrimPrefix(line, "async function ")
		parenIndex := strings.Index(line, "(")
		if parenIndex != -1 {
			return strings.TrimSpace(line[:parenIndex])
		}
	}

	// Handle const/let/var function expressions: const functionName =
	for _, prefix := range []string{"const ", "let ", "var "} {
		if strings.HasPrefix(line, prefix) {
			line = strings.TrimPrefix(line, prefix)
			equalIndex := strings.Index(line, "=")
			if equalIndex != -1 {
				return strings.TrimSpace(line[:equalIndex])
			}
		}
	}

	// Handle method definitions (simplified)
	parenIndex := strings.Index(line, "(")
	if parenIndex != -1 {
		// Work backwards to find the method name
		beforeParen := line[:parenIndex]
		words := strings.Fields(beforeParen)
		if len(words) > 0 {
			return words[len(words)-1]
		}
	}

	return "anonymous"
}

// countPythonComplexityIndicators counts complexity-increasing constructs in Python following McCabe algorithm
// McCabe complexity = number of decision points + 1
// Decision points: if, elif, for, while, except (but not try), and logical operators in conditionals
func (a *ComplexityAnalyzer) countPythonComplexityIndicators(line string) int {
	complexity := 0
	line = strings.TrimSpace(line)

	// Decision points that increase McCabe complexity by 1 each:

	// Conditional branches
	if strings.HasPrefix(line, "if ") {
		complexity++
		// Count logical operators in conditional statements
		andCount := strings.Count(line, " and ")
		orCount := strings.Count(line, " or ")
		complexity += andCount + orCount
	}
	if strings.HasPrefix(line, "elif ") {
		complexity++
		// Count logical operators in conditional statements
		andCount := strings.Count(line, " and ")
		orCount := strings.Count(line, " or ")
		complexity += andCount + orCount
	}

	// Loops
	if strings.HasPrefix(line, "for ") {
		complexity++
	}
	if strings.HasPrefix(line, "while ") {
		complexity++
		// Count logical operators in while conditions
		andCount := strings.Count(line, " and ")
		orCount := strings.Count(line, " or ")
		complexity += andCount + orCount
	}

	// Exception handling - only except clauses, not try
	if strings.HasPrefix(line, "except ") || strings.HasPrefix(line, "except:") {
		complexity++
	}

	return complexity
}

// countJavaComplexityIndicators counts complexity-increasing constructs in Java following McCabe algorithm
//
//nolint:gocyclo // Language-specific complexity indicator counting - multiple conditions needed
func (a *ComplexityAnalyzer) countJavaComplexityIndicators(line string) int {
	complexity := 0
	line = strings.TrimSpace(line)

	// Skip comments and empty lines
	if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
		return 0
	}

	// Decision points that increase McCabe complexity by 1 each:

	// Conditional branches
	if strings.Contains(line, "if (") {
		complexity++
		// Count logical operators in conditional statements
		andCount := strings.Count(line, "&&")
		orCount := strings.Count(line, "||")
		complexity += andCount + orCount
	}
	if strings.Contains(line, "else if") {
		complexity++
		// Count logical operators in conditional statements
		andCount := strings.Count(line, "&&")
		orCount := strings.Count(line, "||")
		complexity += andCount + orCount
	}

	// Loops
	if strings.Contains(line, "for (") {
		complexity++
	}
	if strings.Contains(line, "while (") {
		complexity++
		// Count logical operators in while conditions
		andCount := strings.Count(line, "&&")
		orCount := strings.Count(line, "||")
		complexity += andCount + orCount
	}

	// Switch statements - only case labels, not the switch itself
	if strings.Contains(line, "case ") && strings.Contains(line, ":") {
		complexity++
	}

	// Exception handling - only catch clauses, not try
	if strings.Contains(line, "catch (") {
		complexity++
	}

	// Ternary operators
	if strings.Contains(line, "?") && strings.Contains(line, ":") {
		complexity++
	}

	return complexity
}

// countJavaScriptComplexityIndicators counts complexity-increasing constructs in JavaScript/TypeScript following McCabe algorithm
//
//nolint:gocyclo // Language-specific complexity indicator counting - multiple conditions needed
func (a *ComplexityAnalyzer) countJavaScriptComplexityIndicators(line string) int {
	complexity := 0
	line = strings.TrimSpace(line)

	// Skip comments and empty lines
	if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
		return 0
	}

	// Decision points that increase McCabe complexity by 1 each:

	// Conditional branches
	if strings.Contains(line, "if (") {
		complexity++
		// Count logical operators in conditional statements
		andCount := strings.Count(line, "&&")
		orCount := strings.Count(line, "||")
		complexity += andCount + orCount
	}
	if strings.Contains(line, "else if") {
		complexity++
		// Count logical operators in conditional statements
		andCount := strings.Count(line, "&&")
		orCount := strings.Count(line, "||")
		complexity += andCount + orCount
	}

	// Loops
	if strings.Contains(line, "for (") {
		complexity++
	}
	if strings.Contains(line, "while (") {
		complexity++
		// Count logical operators in while conditions
		andCount := strings.Count(line, "&&")
		orCount := strings.Count(line, "||")
		complexity += andCount + orCount
	}

	// Switch statements - only case labels, not the switch itself
	if strings.Contains(line, "case ") && strings.Contains(line, ":") {
		complexity++
	}

	// Exception handling - only catch clauses, not try
	if strings.Contains(line, "catch (") {
		complexity++
	}

	// Ternary operators
	if strings.Contains(line, "?") && strings.Contains(line, ":") {
		complexity++
	}

	// Arrow functions with conditionals
	// Note: Arrow functions themselves don't add complexity
	// but we may want to analyze their content in the future
	_ = strings.Contains(line, "=> {") // Placeholder for future enhancement

	return complexity
}

// analyzePythonFunctions analyzes individual Python functions for complexity
//
//nolint:gocyclo // Complex parsing function - inherently has high complexity
func (a *ComplexityAnalyzer) analyzePythonFunctions(filePath string) []FunctionComplexity {
	var functions []FunctionComplexity

	content, err := a.safeReadFile(filePath)
	if err != nil {
		return functions
	}

	lines := strings.Split(string(content), "\n")
	inFunction := false
	var currentFunction FunctionComplexity
	indentLevel := 0
	inFunctionSignature := false

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Calculate current line's indentation
		currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))

		// Detect function start - handle multi-line function signatures
		if strings.HasPrefix(trimmedLine, "def ") {
			if inFunction && currentIndent <= indentLevel {
				// Save previous function if we're at the same or lower indentation
				functions = append(functions, currentFunction)
			}

			// Extract function name
			funcName := a.extractPythonFunctionName(trimmedLine)
			currentFunction = FunctionComplexity{
				Name:       funcName,
				StartLine:  lineNum,
				File:       filePath,
				Complexity: 1, // Base complexity
			}
			indentLevel = currentIndent
			inFunctionSignature = true

			// Check if function signature ends on this line
			if strings.Contains(trimmedLine, ":") {
				inFunction = true
				inFunctionSignature = false
			}
			continue
		}

		// Handle multi-line function signatures
		if inFunctionSignature {
			if strings.Contains(trimmedLine, ":") {
				inFunction = true
				inFunctionSignature = false
			}
			continue
		}

		// Count complexity indicators only if we're inside a function
		if inFunction && currentIndent > indentLevel {
			currentFunction.Complexity += a.countPythonComplexityIndicators(trimmedLine)
		}

		// Detect function end (when we encounter a line at or below the function's indentation level)
		// But exclude lines that are part of the function signature or docstrings
		if inFunction && currentIndent <= indentLevel && lineNum > currentFunction.StartLine && trimmedLine != "" {
			// Don't end if this is part of the function definition, a docstring, or the signature closing
			if !strings.HasPrefix(trimmedLine, "def ") &&
				!strings.HasPrefix(trimmedLine, "\"\"\"") &&
				!strings.HasPrefix(trimmedLine, ")") &&
				!inFunctionSignature {
				currentFunction.EndLine = lineNum - 1
				functions = append(functions, currentFunction)
				inFunction = false
			}
		}
	}

	// Handle case where file ends while in a function
	if inFunction {
		currentFunction.EndLine = len(lines)
		functions = append(functions, currentFunction)
	}

	return functions
}

// extractPythonFunctionName extracts the function name from a Python function declaration
func (a *ComplexityAnalyzer) extractPythonFunctionName(line string) string {
	// Remove "def " prefix
	line = strings.TrimPrefix(line, "def ")
	line = strings.TrimSpace(line)

	// Extract function name (everything before the opening parenthesis)
	parenIndex := strings.Index(line, "(")
	if parenIndex != -1 {
		return strings.TrimSpace(line[:parenIndex])
	}

	return "unknown"
}

// analyzeCFunctions analyzes individual C functions for complexity (simplified implementation)
func (a *ComplexityAnalyzer) analyzeCFunctions(filePath string) []FunctionComplexity {
	// Simplified implementation - would need proper C AST parsing for production use
	return []FunctionComplexity{}
}

// safeReadFile reads a file with path validation to prevent directory traversal
func (a *ComplexityAnalyzer) safeReadFile(filePath string) ([]byte, error) {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(filePath)

	// Ensure the path doesn't contain directory traversal patterns
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid file path: contains directory traversal")
	}

	// Additional validation: ensure file has a valid extension
	ext := strings.ToLower(filepath.Ext(cleanPath))
	validExts := []string{".go", ".py", ".java", ".js", ".ts", ".jsx", ".tsx"}
	isValid := false
	for _, validExt := range validExts {
		if ext == validExt {
			isValid = true
			break
		}
	}

	if !isValid {
		return nil, fmt.Errorf("invalid file extension: %s", ext)
	}

	return os.ReadFile(cleanPath)
}

// PythonAnalyzer implements language-specific analysis for Python
type PythonAnalyzer struct {
	*BaseAnalyzer
	complexityAnalyzer *ComplexityAnalyzer
}

// NewPythonAnalyzer creates a new Python analyzer
func NewPythonAnalyzer(config LangConfig, complexityAnalyzer *ComplexityAnalyzer, logger Logger) *PythonAnalyzer {
	base := NewBaseAnalyzer("python", config.Patterns, config.Exclusions, logger)
	return &PythonAnalyzer{
		BaseAnalyzer:       base,
		complexityAnalyzer: complexityAnalyzer,
	}
}

func (p *PythonAnalyzer) SupportsComplexity() bool {
	return true
}

func (p *PythonAnalyzer) SupportsFunctionLevel() bool {
	return true
}

func (p *PythonAnalyzer) AnalyzeComplexity(filePath string) (ComplexityResult, error) {
	if !p.ShouldAnalyzeFile(filePath) {
		return ComplexityResult{}, NewCheckerError("python_analyzer", "file_filter",
			fmt.Errorf("file does not match patterns"), ErrorCodeInvalidInput).WithFile(filePath)
	}

	complexity := p.complexityAnalyzer.calculatePythonComplexity(filePath)
	return ComplexityResult{
		File:       filePath,
		Language:   "Python",
		Complexity: complexity,
	}, nil
}

func (p *PythonAnalyzer) AnalyzeFunctions(filePath string) ([]FunctionComplexity, error) {
	if !p.ShouldAnalyzeFile(filePath) {
		return nil, NewCheckerError("python_analyzer", "file_filter",
			fmt.Errorf("file does not match patterns"), ErrorCodeInvalidInput).WithFile(filePath)
	}

	functions := p.complexityAnalyzer.analyzePythonFunctions(filePath)
	return functions, nil
}

// JavaAnalyzer implements language-specific analysis for Java
type JavaAnalyzer struct {
	*BaseAnalyzer
	complexityAnalyzer *ComplexityAnalyzer
}

// NewJavaAnalyzer creates a new Java analyzer
func NewJavaAnalyzer(config LangConfig, complexityAnalyzer *ComplexityAnalyzer, logger Logger) *JavaAnalyzer {
	base := NewBaseAnalyzer("java", config.Patterns, config.Exclusions, logger)
	return &JavaAnalyzer{
		BaseAnalyzer:       base,
		complexityAnalyzer: complexityAnalyzer,
	}
}

func (j *JavaAnalyzer) SupportsComplexity() bool {
	return true
}

func (j *JavaAnalyzer) SupportsFunctionLevel() bool {
	return true
}

func (j *JavaAnalyzer) AnalyzeComplexity(filePath string) (ComplexityResult, error) {
	if !j.ShouldAnalyzeFile(filePath) {
		return ComplexityResult{}, NewCheckerError("java_analyzer", "file_filter",
			fmt.Errorf("file does not match patterns"), ErrorCodeInvalidInput).WithFile(filePath)
	}

	complexity := j.complexityAnalyzer.calculateJavaComplexity(filePath)
	return ComplexityResult{
		File:       filePath,
		Language:   "Java",
		Complexity: complexity,
	}, nil
}

func (j *JavaAnalyzer) AnalyzeFunctions(filePath string) ([]FunctionComplexity, error) {
	if !j.ShouldAnalyzeFile(filePath) {
		return nil, NewCheckerError("java_analyzer", "file_filter",
			fmt.Errorf("file does not match patterns"), ErrorCodeInvalidInput).WithFile(filePath)
	}

	functions := j.complexityAnalyzer.analyzeJavaFunctions(filePath)
	return functions, nil
}

// JavaScriptAnalyzer implements language-specific analysis for JavaScript/TypeScript
type JavaScriptAnalyzer struct {
	*BaseAnalyzer
	complexityAnalyzer *ComplexityAnalyzer
}

// NewJavaScriptAnalyzer creates a new JavaScript analyzer
func NewJavaScriptAnalyzer(config LangConfig, complexityAnalyzer *ComplexityAnalyzer, logger Logger) *JavaScriptAnalyzer {
	base := NewBaseAnalyzer("javascript", config.Patterns, config.Exclusions, logger)
	return &JavaScriptAnalyzer{
		BaseAnalyzer:       base,
		complexityAnalyzer: complexityAnalyzer,
	}
}

func (js *JavaScriptAnalyzer) SupportsComplexity() bool {
	return true
}

func (js *JavaScriptAnalyzer) SupportsFunctionLevel() bool {
	return true
}

func (js *JavaScriptAnalyzer) AnalyzeComplexity(filePath string) (ComplexityResult, error) {
	if !js.ShouldAnalyzeFile(filePath) {
		return ComplexityResult{}, NewCheckerError("javascript_analyzer", "file_filter",
			fmt.Errorf("file does not match patterns"), ErrorCodeInvalidInput).WithFile(filePath)
	}

	complexity := js.complexityAnalyzer.calculateJSComplexity(filePath)
	return ComplexityResult{
		File:       filePath,
		Language:   "JavaScript",
		Complexity: complexity,
	}, nil
}

func (js *JavaScriptAnalyzer) AnalyzeFunctions(filePath string) ([]FunctionComplexity, error) {
	if !js.ShouldAnalyzeFile(filePath) {
		return nil, NewCheckerError("javascript_analyzer", "file_filter",
			fmt.Errorf("file does not match patterns"), ErrorCodeInvalidInput).WithFile(filePath)
	}

	functions := js.complexityAnalyzer.analyzeJavaScriptFunctions(filePath)
	return functions, nil
}
