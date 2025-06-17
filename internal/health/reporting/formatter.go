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
	verbose bool
}

// NewFormatter creates a new result formatter
func NewFormatter(verbose bool) *Formatter {
	return &Formatter{
		verbose: verbose,
	}
}

// DisplayResults formats and displays the health analysis results
func (f *Formatter) DisplayResults(result core.WorkflowResult) {
	f.displaySummary(result.Summary)

	// Display each repository individually
	f.displayRepositoryReports(result.RepositoryResults)

	f.displayTiming(result)
}

// displaySummary shows the overall summary of the health check
func (f *Formatter) displaySummary(summary core.WorkflowSummary) {
	color.Green("\n=== Health Check Summary ===")
	fmt.Printf("Total Repositories: %d\n", summary.SuccessfulRepos+summary.FailedRepos)
	fmt.Printf("Successful Checks: %d\n", summary.SuccessfulRepos)

	if summary.FailedRepos > 0 {
		color.Red("Failed Checks: %d", summary.FailedRepos)
	} else {
		color.Green("Failed Checks: %d", summary.FailedRepos)
	}

	fmt.Printf("Average Score: %d\n", summary.AverageScore)
	fmt.Printf("Total Issues Found: %d\n", summary.TotalIssues)

	// Display status counts
	if len(summary.StatusCounts) > 0 {
		fmt.Println("Status Distribution:")
		for status, count := range summary.StatusCounts {
			fmt.Printf("  %s: %d\n", status, count)
		}
	}
	fmt.Println()
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
	// Repository header
	color.Cyan("┌─────────────────────────────────────────────────────────────────────────────────┐")
	color.Cyan("│ Repository: %-67s │", result.Repository.Name)
	color.Cyan("├─────────────────────────────────────────────────────────────────────────────────┤")

	// Basic information
	fmt.Printf("│ Path: %-73s │\n", result.Repository.Path)
	fmt.Printf("│ Language: %-67s │\n", result.Repository.Language)

	// Status and score
	statusDisplay := f.getStatusDisplay(result.Status)
	fmt.Printf("│ Status: %-69s │\n", statusDisplay)
	fmt.Printf("│ Health Score: %d/%d %-58s │\n", result.Score, result.MaxScore, f.getScoreBar(result.Score, result.MaxScore))

	if result.Error != "" {
		color.Cyan("├─────────────────────────────────────────────────────────────────────────────────┤")
		color.Red("│ Error: %-72s │", result.Error)
		color.Cyan("└─────────────────────────────────────────────────────────────────────────────────┘")
		return
	}

	// Checks summary
	checksRun := len(result.CheckResults)
	passedChecks := f.countPassedChecks(result.CheckResults)
	failedChecks := f.countFailedChecks(result.CheckResults)
	totalIssues := f.countIssues(result.CheckResults)

	color.Cyan("├─────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Printf("│ Checks: %d total, %d passed, %d failed, %d issues found %-20s │\n",
		checksRun, passedChecks, failedChecks, totalIssues, "")

	// Cyclomatic Complexity Analysis
	f.displayCyclomaticComplexitySection(result)

	// Health Checks Results
	if len(result.CheckResults) > 0 {
		color.Cyan("├─────────────────────────────────────────────────────────────────────────────────┤")
		color.Cyan("│ Health Check Results                                                            │")
		color.Cyan("├─────────────────────────────────────────────────────────────────────────────────┤")

		for _, checkResult := range result.CheckResults {
			f.displayCheckResultInBox(checkResult)
		}
	}

	color.Cyan("└─────────────────────────────────────────────────────────────────────────────────┘")
}

// displayCyclomaticComplexitySection shows detailed cyclomatic complexity analysis
//
//nolint:gocyclo
func (f *Formatter) displayCyclomaticComplexitySection(result core.RepositoryResult) {
	if result.AnalysisResult == nil {
		color.Cyan("├─────────────────────────────────────────────────────────────────────────────────┤")
		fmt.Printf("│ Cyclomatic Complexity: No analysis data available %-27s │\n", "")
		return
	}

	analysis := *result.AnalysisResult

	color.Cyan("├─────────────────────────────────────────────────────────────────────────────────┤")
	color.Cyan("│ Cyclomatic Complexity Analysis                                                  │")
	color.Cyan("├─────────────────────────────────────────────────────────────────────────────────┤")

	// Overall complexity metrics
	if totalComplexity, ok := analysis.Metrics["total_complexity"].(int); ok {
		fmt.Printf("│ Total Complexity: %-59d │\n", totalComplexity)
	}

	if maxComplexity, ok := analysis.Metrics["max_complexity"].(int); ok {
		complexityLevel := f.getComplexityLevel(maxComplexity)
		fmt.Printf("│ Maximum Complexity: %-54s │\n",
			fmt.Sprintf("%d (%s)", maxComplexity, complexityLevel))
	}

	if avgComplexity, ok := analysis.Metrics["average_complexity"].(float64); ok {
		complexityLevel := f.getComplexityLevel(int(avgComplexity))
		fmt.Printf("│ Average Complexity: %-54s │\n",
			fmt.Sprintf("%.2f (%s)", avgComplexity, complexityLevel))
	}

	if totalFunctions, ok := analysis.Metrics["total_functions"].(int); ok {
		fmt.Printf("│ Total Functions: %-60d │\n", totalFunctions)
	}

	if totalFiles, ok := analysis.Metrics["total_files"].(int); ok {
		fmt.Printf("│ Files Analyzed: %-61d │\n", totalFiles)
	}

	// Function-level complexity breakdown
	if len(analysis.Functions) > 0 && f.verbose {
		color.Cyan("├─────────────────────────────────────────────────────────────────────────────────┤")
		color.Cyan("│ Function Complexity Breakdown                                                   │")
		color.Cyan("├─────────────────────────────────────────────────────────────────────────────────┤")

		// Sort functions by complexity (highest first)
		functions := make([]core.FunctionInfo, len(analysis.Functions))
		copy(functions, analysis.Functions)
		f.sortFunctionsByComplexity(functions)

		// Show top complex functions (limit to 10 for readability)
		limit := 10
		if len(functions) < limit {
			limit = len(functions)
		}

		for i := 0; i < limit; i++ {
			fn := functions[i]
			complexityLevel := f.getComplexityLevel(fn.Complexity)
			fileName := f.getShortFileName(fn.File)

			fmt.Printf("│ %s:%d - %s() %-50s │\n",
				fileName, fn.Line, fn.Name,
				fmt.Sprintf("Complexity: %d (%s)", fn.Complexity, complexityLevel))
		}

		if len(functions) > limit {
			fmt.Printf("│ ... and %d more functions %-50s │\n", len(functions)-limit, "")
		}
	}
}

// Helper functions for the new display format

// getStatusDisplay returns a colored status display string
func (f *Formatter) getStatusDisplay(status core.HealthStatus) string {
	switch status {
	case core.StatusHealthy:
		return color.GreenString("✅ Healthy")
	case core.StatusWarning:
		return color.YellowString("⚠️  Warning")
	case core.StatusCritical:
		return color.RedString("❌ Critical")
	default:
		return string(status)
	}
}

// getScoreBar creates a visual score bar
func (f *Formatter) getScoreBar(score, maxScore int) string {
	if maxScore <= 0 {
		return ""
	}

	percentage := float64(score) / float64(maxScore)
	barLength := 20
	filledLength := int(percentage * float64(barLength))

	bar := "["
	for i := 0; i < barLength; i++ {
		if i < filledLength {
			if percentage >= 0.8 {
				bar += color.GreenString("█")
			} else if percentage >= 0.6 {
				bar += color.YellowString("█")
			} else {
				bar += color.RedString("█")
			}
		} else {
			bar += "░"
		}
	}
	bar += fmt.Sprintf("] %.1f%%", percentage*100)
	return bar
}

// getComplexityLevel returns a descriptive level for complexity values
func (f *Formatter) getComplexityLevel(complexity int) string {
	switch {
	case complexity <= 5:
		return color.GreenString("Low")
	case complexity <= 10:
		return color.YellowString("Moderate")
	case complexity <= 20:
		return color.RedString("High")
	default:
		return color.RedString("Very High")
	}
}

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

// getShortFileName returns a shortened file name for display
func (f *Formatter) getShortFileName(filePath string) string {
	if len(filePath) <= 20 {
		return filePath
	}

	// Extract just the filename from the path
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		if len(filename) <= 20 {
			return filename
		}
		// Truncate if still too long
		return filename[:17] + "..."
	}

	return filePath[:17] + "..."
}

// displayCheckResultInBox shows a check result within the repository box format
//
//nolint:gocyclo
func (f *Formatter) displayCheckResultInBox(result core.CheckResult) {
	statusIcon := f.getCheckStatusIcon(result.Status)

	checkLine := fmt.Sprintf("%s %s (%s): Score %d/%d",
		statusIcon, result.Name, result.Category, result.Score, result.MaxScore)

	fmt.Printf("│ %-79s │\n", checkLine)

	// Show metrics on next line if available and verbose
	if len(result.Metrics) > 0 && f.verbose {
		metricsStr := "  Metrics: "
		first := true
		for key, value := range result.Metrics {
			if !first {
				metricsStr += ", "
			}
			metricsStr += fmt.Sprintf("%s=%v", key, value)
			first = false
		}

		// Truncate if too long
		if len(metricsStr) > 75 {
			metricsStr = metricsStr[:72] + "..."
		}

		fmt.Printf("│ %-79s │\n", metricsStr)
	}

	// Show issues summary
	if len(result.Issues) > 0 {
		issuesLine := fmt.Sprintf("  Issues: %d found", len(result.Issues))
		fmt.Printf("│ %-79s │\n", issuesLine)

		// Show first few issues if verbose
		if f.verbose {
			limit := 3
			if len(result.Issues) < limit {
				limit = len(result.Issues)
			}

			for i := 0; i < limit; i++ {
				issue := result.Issues[i]
				issueLine := fmt.Sprintf("    • %s: %s", issue.Severity, issue.Message)
				if len(issueLine) > 75 {
					issueLine = issueLine[:72] + "..."
				}
				fmt.Printf("│ %-79s │\n", issueLine)
			}

			if len(result.Issues) > limit {
				fmt.Printf("│ %-79s │\n", fmt.Sprintf("    ... and %d more issues", len(result.Issues)-limit))
			}
		}
	}
}

// getCheckStatusIcon returns the appropriate icon for a check status
func (f *Formatter) getCheckStatusIcon(status core.HealthStatus) string {
	switch status {
	case core.StatusCritical:
		return "❌"
	case core.StatusWarning:
		return "⚠️"
	default:
		return "✅"
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

// countIssues counts the total number of issues in check results
func (f *Formatter) countIssues(checkResults []core.CheckResult) int {
	count := 0
	for _, result := range checkResults {
		count += len(result.Issues)
	}
	return count
}

// countPassedChecks counts checks that passed (not critical status)
func (f *Formatter) countPassedChecks(checkResults []core.CheckResult) int {
	count := 0
	for _, result := range checkResults {
		if result.Status != core.StatusCritical {
			count++
		}
	}
	return count
}

// countFailedChecks counts checks that failed (critical status)
func (f *Formatter) countFailedChecks(checkResults []core.CheckResult) int {
	count := 0
	for _, result := range checkResults {
		if result.Status == core.StatusCritical {
			count++
		}
	}
	return count
}

// ExitCode determines the appropriate exit code based on results
func ExitCode(result core.WorkflowResult) int {
	if result.Summary.FailedRepos > 0 {
		return 2 // Critical issues found
	}
	return 0 // Success
}
