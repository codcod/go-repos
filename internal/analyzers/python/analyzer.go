package python_analyzer

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codcod/repos/internal/core"
)

// PythonAnalyzer implements language-specific analysis for Python code
type PythonAnalyzer struct {
	name       string
	language   string
	extensions []string
	excludes   []string
	filesystem core.FileSystem
	logger     core.Logger
}

// NewPythonAnalyzer creates a new Python language analyzer
func NewPythonAnalyzer(fs core.FileSystem, logger core.Logger) *PythonAnalyzer {
	return &PythonAnalyzer{
		name:       "python-analyzer",
		language:   "python",
		extensions: []string{".py"},
		excludes:   []string{".venv/", "__pycache__/", ".git/", "venv/", "env/", ".pytest_cache/"},
		filesystem: fs,
		logger:     logger,
	}
}

// Name returns the analyzer name
func (p *PythonAnalyzer) Name() string {
	return p.name
}

// Language returns the supported language
func (p *PythonAnalyzer) Language() string {
	return p.language
}

// SupportedExtensions returns supported file extensions
func (p *PythonAnalyzer) SupportedExtensions() []string {
	return p.extensions
}

// CanAnalyze checks if the analyzer can process the given repository
func (p *PythonAnalyzer) CanAnalyze(repo core.Repository) bool {
	// Check if repository has Python files
	return p.hasPythonFiles(repo.Path)
}

// Analyze performs language-specific analysis on the repository
func (p *PythonAnalyzer) Analyze(ctx context.Context, repo core.Repository) (*core.AnalysisResult, error) {
	p.logger.Info("Starting Python analysis", core.Field{Key: "repo", Value: repo.Name})

	result := &core.AnalysisResult{
		Language:  p.language,
		Files:     make(map[string]*core.FileAnalysis),
		Functions: []core.FunctionInfo{},
		Metrics:   make(map[string]interface{}),
	}

	// Find Python files
	files, err := p.findPythonFiles(repo.Path)
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

		fileAnalysis, err := p.analyzeFile(file)
		if err != nil {
			p.logger.Warn("Failed to analyze file",
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

	p.logger.Info("Python analysis completed",
		core.Field{Key: "files", Value: len(result.Files)},
		core.Field{Key: "functions", Value: totalFunctions})

	return result, nil
}

// hasPythonFiles checks if the repository contains Python files
func (p *PythonAnalyzer) hasPythonFiles(repoPath string) bool {
	files, err := p.findPythonFiles(repoPath)
	return err == nil && len(files) > 0
}

// findPythonFiles finds all Python source files in the repository
func (p *PythonAnalyzer) findPythonFiles(repoPath string) ([]string, error) {
	var pythonFiles []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a Python file
		if !strings.HasSuffix(path, ".py") {
			return nil
		}

		// Skip excluded patterns
		relPath, _ := filepath.Rel(repoPath, path)
		for _, exclude := range p.excludes {
			if strings.Contains(relPath, exclude) {
				return nil
			}
		}

		pythonFiles = append(pythonFiles, path)
		return nil
	})

	return pythonFiles, err
}

// analyzeFile analyzes a single Python file
func (p *PythonAnalyzer) analyzeFile(filePath string) (*core.FileAnalysis, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	analysis := &core.FileAnalysis{
		Path:      filePath,
		Language:  p.language,
		Functions: []core.FunctionInfo{},
		Imports:   []core.ImportInfo{},
		Metrics:   make(map[string]interface{}),
	}

	// Analyze functions and imports
	functions, imports := p.parseFile(string(content), filePath)
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

// parseFile parses a Python file to extract functions and imports
func (p *PythonAnalyzer) parseFile(content, filePath string) ([]core.FunctionInfo, []core.ImportInfo) {
	var functions []core.FunctionInfo
	var imports []core.ImportInfo

	lines := strings.Split(content, "\n")
	inFunction := false
	var currentFunction *core.FunctionInfo
	indentLevel := 0

	// Regex patterns
	functionPattern := regexp.MustCompile(`^\s*def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	importPattern := regexp.MustCompile(`^\s*(?:from\s+([a-zA-Z_][a-zA-Z0-9_.]*)\s+)?import\s+([a-zA-Z_][a-zA-Z0-9_.*,\s]+)`)
	classPattern := regexp.MustCompile(`^\s*class\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*(?:\(.*\))?\s*:`)

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Calculate current line's indentation
		currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))

		// Check for imports
		if matches := importPattern.FindStringSubmatch(line); matches != nil {
			fromModule := matches[1] // Could be empty for direct imports
			importItems := strings.Split(matches[2], ",")

			for _, item := range importItems {
				item = strings.TrimSpace(item)
				if item != "" {
					importInfo := core.ImportInfo{
						Name:    item,
						Path:    fromModule,
						Line:    lineNum,
						IsLocal: !strings.Contains(item, ".") && fromModule == "",
					}

					// Handle aliases
					if strings.Contains(item, " as ") {
						parts := strings.Split(item, " as ")
						if len(parts) == 2 {
							importInfo.Name = strings.TrimSpace(parts[0])
							importInfo.Alias = strings.TrimSpace(parts[1])
						}
					}

					imports = append(imports, importInfo)
				}
			}
		}

		// Check for function definitions
		if matches := functionPattern.FindStringSubmatch(line); matches != nil {
			functionName := matches[1]

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
				Language:   p.language,
			}

			inFunction = true
			indentLevel = currentIndent
		} else if inFunction && currentFunction != nil {
			// We're inside a function
			if currentIndent <= indentLevel && trimmedLine != "" {
				// Function ended
				functions = append(functions, *currentFunction)
				inFunction = false
				currentFunction = nil
			} else {
				// Still inside function, calculate complexity
				currentFunction.Complexity += p.calculateLineComplexity(trimmedLine)
			}
		}

		// Check for class definitions (for future enhancement)
		if matches := classPattern.FindStringSubmatch(line); matches != nil {
			// TODO: Implement class analysis
			_ = matches[1] // className
		}
	}

	// Don't forget the last function if the file ends while in a function
	if inFunction && currentFunction != nil {
		functions = append(functions, *currentFunction)
	}

	return functions, imports
}

// calculateLineComplexity calculates complexity contribution of a single line
func (p *PythonAnalyzer) calculateLineComplexity(line string) int {
	complexity := 0
	line = strings.TrimSpace(line)

	// Skip comments and empty lines
	if line == "" || strings.HasPrefix(line, "#") {
		return 0
	}

	// Decision points that increase McCabe complexity:

	// Conditional branches
	if strings.HasPrefix(line, "if ") || strings.Contains(line, " if ") {
		complexity++
		// Count logical operators in conditional statements
		complexity += strings.Count(line, " and ")
		complexity += strings.Count(line, " or ")
	}
	if strings.HasPrefix(line, "elif ") {
		complexity++
		complexity += strings.Count(line, " and ")
		complexity += strings.Count(line, " or ")
	}

	// Loops
	if strings.HasPrefix(line, "for ") {
		complexity++
	}
	if strings.HasPrefix(line, "while ") {
		complexity++
		complexity += strings.Count(line, " and ")
		complexity += strings.Count(line, " or ")
	}

	// Exception handling - only except clauses
	if strings.HasPrefix(line, "except ") {
		complexity++
	}

	// Context managers
	if strings.HasPrefix(line, "with ") {
		complexity++
	}

	// Lambda functions
	if strings.Contains(line, "lambda ") {
		complexity++
	}

	// Assert statements with conditions
	if strings.HasPrefix(line, "assert ") {
		complexity++
	}

	// List/dict comprehensions with conditions
	if strings.Contains(line, " if ") && (strings.Contains(line, "[") || strings.Contains(line, "{")) {
		complexity++
	}

	return complexity
}
