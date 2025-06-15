package health

import (
	"fmt"
	"os"
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

	text := string(content)
	complexity := 1 // Base complexity

	keywords := []string{
		"if (", "else if", "while (", "for (", "switch (", "case ", "default:",
		"catch (", "&&", "||", "?", ":",
	}

	for _, keyword := range keywords {
		complexity += strings.Count(text, keyword)
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

	text := string(content)
	complexity := 1 // Base complexity

	keywords := []string{
		"if (", "else if", "while (", "for (", "switch (", "case ", "default:",
		"catch (", "&&", "||", "?", ":", "=> {",
	}

	for _, keyword := range keywords {
		complexity += strings.Count(text, keyword)
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
