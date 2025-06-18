// Package reporting provides result formatting and output functionality for health analysis
package reporting

import (
	"fmt"
	"strings"
	"time"

	"github.com/codcod/repos/internal/core"
	"github.com/fatih/color"
)

// Formatter handles the formatting and display of health analysis results
type Formatter struct {
	verbose             bool
	ComplexityThreshold int // minimum complexity to show, default 10
}

// NewFormatter creates a new result formatter
func NewFormatter(verbose bool) *Formatter {
	return &Formatter{
		verbose:             verbose,
		ComplexityThreshold: 10, // default threshold
	}
}

// NewComplexityFormatterWithThreshold creates a formatter with a specific complexity threshold
func NewComplexityFormatterWithThreshold(verbose bool, threshold int) *Formatter {
	return &Formatter{
		verbose:             verbose,
		ComplexityThreshold: threshold,
	}
}

// DisplayResults formats and displays the health analysis results
func (f *Formatter) DisplayResults(result core.WorkflowResult) {
	// Display each repository individually (removed summary)
	f.displayRepositoryReports(result.RepositoryResults)

	f.displayTiming(result)
}

// displayRepositoryReports shows individual reports for each repository
func (f *Formatter) displayRepositoryReports(results []core.RepositoryResult) {
	color.Green("=== Repository Health Reports ===")

	for i, result := range results {
		if i > 0 {
			fmt.Println() // Add spacing between repositories
		}
		f.displayIndividualRepositoryReport(result)
	}
}

// displayIndividualRepositoryReport shows a comprehensive report for a single repository
func (f *Formatter) displayIndividualRepositoryReport(result core.RepositoryResult) {
	// Repository header in red (removed separator line)
	color.Red("Repository: %s", result.Repository.Name)

	// Language - handle empty case
	language := result.Repository.Language
	if language == "" {
		language = "Unknown"
	}
	fmt.Printf("Language: %s\n", language)

	// Status with emoji and score
	statusEmoji := f.getStatusEmoji(result.Status)
	statusText := f.getStatusText(result.Status)

	// Calculate proper max score if it's 0
	maxScore := result.MaxScore
	if maxScore == 0 {
		maxScore = 100 // Default to 100 if not set
	}

	fmt.Printf("Status: %s %s (%d/%d)\n", statusEmoji, statusText, result.Score, maxScore)

	// Add blank line before health checks
	fmt.Println()

	// Health checks results
	if len(result.CheckResults) > 0 {
		fmt.Println("Health checks results")
		for _, checkResult := range result.CheckResults {
			f.displayCheckResultSimple(checkResult)
		}
	}

	// Add blank line before complexity section
	fmt.Println()

	// Cyclomatic complexity
	f.displayCyclomaticComplexitySimple(result)
}

// getStatusText returns a simple text representation of the status
func (f *Formatter) getStatusText(status core.HealthStatus) string {
	switch status {
	case core.StatusHealthy:
		return "Healthy"
	case core.StatusWarning:
		return "Warning"
	case core.StatusCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// getStatusEmoji returns the emoji for the overall status
func (f *Formatter) getStatusEmoji(status core.HealthStatus) string {
	switch status {
	case core.StatusHealthy:
		return "✅"
	case core.StatusWarning:
		return "⚠️"
	case core.StatusCritical:
		return "❌"
	default:
		return "❓"
	}
}

// displayCheckResultSimple shows a check result in the simple format
func (f *Formatter) displayCheckResultSimple(result core.CheckResult) {
	emoji := f.getCheckStatusEmoji(result.Status)

	// Check if this is a warning about tool not being available
	scoreDisplay := fmt.Sprintf("%d", result.Score)
	if result.Status == core.StatusWarning && f.isToolUnavailableWarning(result) {
		scoreDisplay = "unknown"
	}

	fmt.Printf("%s %s (%s): %s\n", emoji, result.Name, result.Category, scoreDisplay)

	// Show top 3 issues in grey
	if len(result.Issues) > 0 {
		limit := 3
		if len(result.Issues) < limit {
			limit = len(result.Issues)
		}

		for i := 0; i < limit; i++ {
			issue := result.Issues[i]
			// Print issues in grey color
			_, _ = color.New(color.FgHiBlack).Printf("  - %s\n", issue.Message)
		}
	}
}

// isToolUnavailableWarning checks if a warning is about a tool not being available
func (f *Formatter) isToolUnavailableWarning(result core.CheckResult) bool {
	if len(result.Issues) == 0 {
		return false
	}

	// Check for common patterns indicating tool unavailability
	for _, issue := range result.Issues {
		msg := strings.ToLower(issue.Message)
		if strings.Contains(msg, "not available") ||
			strings.Contains(msg, "not implemented") ||
			strings.Contains(msg, "scanner not available") ||
			strings.Contains(msg, "not found") {
			return true
		}
	}
	return false
}

// displayCyclomaticComplexitySimple shows complexity issues in the simple format
func (f *Formatter) displayCyclomaticComplexitySimple(result core.RepositoryResult) {
	if result.AnalysisResult == nil {
		return
	}

	analysis := *result.AnalysisResult

	// Only show if there are functions with complexity issues
	complexFunctions := f.getComplexFunctions(analysis.Functions)

	if len(complexFunctions) == 0 {
		return
	}

	fmt.Println("Cyclomatic complexity")

	// Show top complex functions (limit to reasonable number)
	limit := 10
	if len(complexFunctions) < limit {
		limit = len(complexFunctions)
	}

	for i := 0; i < limit; i++ {
		fn := complexFunctions[i]
		// Make file path relative to repository
		relativePath := f.getRelativePath(fn.File, result.Repository.Path)
		// Format: path:line:column: 'function name' is too complex (score)
		// Note: We don't have column info, so using 0
		fmt.Printf("  - %s:%d:0: '%s' is too complex (%d)\n",
			relativePath, fn.Line, fn.Name, fn.Complexity)
	}
}

// getRelativePath returns a path relative to the repository root
func (f *Formatter) getRelativePath(filePath, repoPath string) string {
	// If the file path starts with the repo path, remove it
	if strings.HasPrefix(filePath, repoPath) {
		relativePath := strings.TrimPrefix(filePath, repoPath)
		// Remove leading slash
		relativePath = strings.TrimPrefix(relativePath, "/")
		return relativePath
	}

	// If we can't make it relative, return the original path
	return filePath
}

// getComplexFunctions returns functions that are considered too complex (>= threshold)
// sorted by complexity descending
func (f *Formatter) getComplexFunctions(functions []core.FunctionInfo) []core.FunctionInfo {
	var complexFunctions []core.FunctionInfo

	// Filter functions with complexity >= threshold
	for _, fn := range functions {
		if fn.Complexity >= f.ComplexityThreshold {
			complexFunctions = append(complexFunctions, fn)
		}
	}

	// Sort by complexity (highest first)
	f.sortFunctionsByComplexity(complexFunctions)

	return complexFunctions
}

// getCheckStatusEmoji returns the appropriate emoji for a check status
func (f *Formatter) getCheckStatusEmoji(status core.HealthStatus) string {
	switch status {
	case core.StatusCritical:
		return "❌"
	case core.StatusWarning:
		return "⚠️"
	default:
		return "✅"
	}
}

// Helper functions for the new display format

// sortFunctionsByComplexity sorts functions by complexity in descending order
func (f *Formatter) sortFunctionsByComplexity(functions []core.FunctionInfo) {
	for i := 0; i < len(functions)-1; i++ {
		for j := i + 1; j < len(functions); j++ {
			if functions[i].Complexity < functions[j].Complexity {
				functions[i], functions[j] = functions[j], functions[i]
			}
		}
	}
}

// displayTiming shows execution timing information
func (f *Formatter) displayTiming(result core.WorkflowResult) {
	if !f.verbose {
		return
	}

	fmt.Println("=== Timing Information ===")
	if !result.StartTime.IsZero() && !result.EndTime.IsZero() {
		duration := result.EndTime.Sub(result.StartTime)
		fmt.Printf("Total execution time: %v\n", duration.Round(time.Millisecond))
	}
}

// ExitCode determines the appropriate exit code based on results
func ExitCode(result core.WorkflowResult) int {
	if result.Summary.FailedRepos > 0 {
		return 2 // Critical issues found
	}
	return 0 // Success
}
