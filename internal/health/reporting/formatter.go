// Package reporting provides result formatting and output functionality for health analysis
package reporting

import (
	"fmt"
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

	if f.verbose {
		f.displayDetailedResults(result.RepositoryResults)
	} else {
		f.displayCompactResults(result.RepositoryResults)
	}

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

// displayCompactResults shows a compact view of results
func (f *Formatter) displayCompactResults(results []core.RepositoryResult) {
	color.Green("=== Repository Results ===")

	for _, result := range results {
		if result.Error != "" {
			color.Red("❌ %s: %s", result.Repository.Name, result.Error)
			continue
		}

		issueCount := f.countIssues(result.CheckResults)
		if issueCount > 0 {
			color.Yellow("⚠️  %s: %d issues found", result.Repository.Name, issueCount)
		} else {
			color.Green("✅ %s: All checks passed", result.Repository.Name)
		}
	}
	fmt.Println()
}

// displayDetailedResults shows detailed results for each repository
func (f *Formatter) displayDetailedResults(results []core.RepositoryResult) {
	color.Green("=== Detailed Results ===")

	for _, result := range results {
		f.displayRepositoryResult(result)
	}
}

// displayRepositoryResult shows detailed results for a single repository
func (f *Formatter) displayRepositoryResult(result core.RepositoryResult) {
	fmt.Printf("\n--- Repository: %s ---\n", result.Repository.Name)
	fmt.Printf("Path: %s\n", result.Repository.Path)
	fmt.Printf("Language: %s\n", result.Repository.Language)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Score: %d/%d\n", result.Score, result.MaxScore)

	if result.Error != "" {
		color.Red("Error: %s\n", result.Error)
		return
	}

	// Display analysis results
	if result.AnalysisResult != nil {
		fmt.Println("\nAnalysis Results:")
		f.displayAnalysisResult("Language Analysis", *result.AnalysisResult)
	}

	// Display checker results
	if len(result.CheckResults) > 0 {
		fmt.Println("\nChecker Results:")
		for _, checkResult := range result.CheckResults {
			f.displayCheckResult(checkResult)
		}
	}
}

// displayAnalysisResult shows the result of a single analyzer
func (f *Formatter) displayAnalysisResult(analyzerName string, result core.AnalysisResult) {
	fmt.Printf("  [%s]\n", analyzerName)
	fmt.Printf("    Language: %s\n", result.Language)
	fmt.Printf("    Files analyzed: %d\n", len(result.Files))
	fmt.Printf("    Functions found: %d\n", len(result.Functions))

	if len(result.Metrics) > 0 {
		fmt.Println("    Metrics:")
		for key, value := range result.Metrics {
			fmt.Printf("      %s: %v\n", key, value)
		}
	}
}

// displayCheckResult shows the result of a single checker
func (f *Formatter) displayCheckResult(result core.CheckResult) {
	statusIcon := "✅"
	statusColor := color.GreenString

	switch result.Status {
	case core.StatusCritical:
		statusIcon = "❌"
		statusColor = color.RedString
	case core.StatusWarning:
		statusIcon = "⚠️"
		statusColor = color.YellowString
	}

	fmt.Printf("  %s %s (%s): Score %d/%d\n", statusIcon, statusColor(result.Name), result.Category, result.Score, result.MaxScore)

	if len(result.Issues) > 0 {
		for _, issue := range result.Issues {
			f.displayIssue(issue)
		}
	}

	if len(result.Warnings) > 0 {
		for _, warning := range result.Warnings {
			f.displayWarning(warning)
		}
	}
}

// displayIssue shows details of a single issue
func (f *Formatter) displayIssue(issue core.Issue) {
	severityColor := color.GreenString
	switch issue.Severity {
	case core.SeverityHigh:
		severityColor = color.RedString
	case core.SeverityMedium:
		severityColor = color.YellowString
	case core.SeverityLow:
		severityColor = color.CyanString
	}

	fmt.Printf("    - [%s] %s", severityColor(string(issue.Severity)), issue.Message)

	if issue.Location != nil {
		fmt.Printf(" (%s", issue.Location.File)
		if issue.Location.Line > 0 {
			fmt.Printf(":%d", issue.Location.Line)
		}
		fmt.Printf(")")
	}
	fmt.Println()

	if f.verbose && issue.Suggestion != "" {
		fmt.Printf("      Suggestion: %s\n", issue.Suggestion)
	}
}

// displayWarning shows details of a single warning
func (f *Formatter) displayWarning(warning core.Warning) {
	fmt.Printf("    ⚠ %s", warning.Message)

	if warning.Location != nil {
		fmt.Printf(" (%s", warning.Location.File)
		if warning.Location.Line > 0 {
			fmt.Printf(":%d", warning.Location.Line)
		}
		fmt.Printf(")")
	}
	fmt.Println()
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

// ExitCode determines the appropriate exit code based on results
func ExitCode(result core.WorkflowResult) int {
	if result.Summary.FailedRepos > 0 {
		return 2 // Critical issues found
	}
	return 0 // Success
}
