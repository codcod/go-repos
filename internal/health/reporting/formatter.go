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
	f.displayChecksSummary(result.RepositoryResults)

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
		checksRun := len(result.CheckResults)

		if result.Status == core.StatusCritical {
			color.Red("❌ %s: %d/%d checks failed (%d issues, score: %d)",
				result.Repository.Name, f.countFailedChecks(result.CheckResults), checksRun, issueCount, result.Score)
		} else if issueCount > 0 {
			color.Yellow("⚠️  %s: %d/%d checks passed (%d issues, score: %d)",
				result.Repository.Name, f.countPassedChecks(result.CheckResults), checksRun, issueCount, result.Score)
		} else {
			color.Green("✅ %s: %d/%d checks passed (score: %d)",
				result.Repository.Name, checksRun, checksRun, result.Score)
		}

		// Show categories of checks that were run
		if checksRun > 0 {
			categories := f.getCheckCategories(result.CheckResults)
			fmt.Printf("   └── Checks: %s\n", categories)
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

	// Show checks summary
	checksRun := len(result.CheckResults)
	passedChecks := f.countPassedChecks(result.CheckResults)
	failedChecks := f.countFailedChecks(result.CheckResults)
	totalIssues := f.countIssues(result.CheckResults)

	fmt.Printf("Checks: %d total (%d passed, %d failed, %d issues)\n",
		checksRun, passedChecks, failedChecks, totalIssues)

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
//
//nolint:gocyclo
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

	// Show check name, category, and score
	fmt.Printf("  %s %s (%s): Score %d/%d", statusIcon, statusColor(result.Name), result.Category, result.Score, result.MaxScore)

	// Show duration if available
	if result.Duration > 0 {
		fmt.Printf(" [%v]", result.Duration.Round(time.Millisecond))
	}

	// Show check ID for reference
	if result.ID != "" {
		fmt.Printf(" (ID: %s)", result.ID)
	}
	fmt.Println()

	// Show metrics if available
	if len(result.Metrics) > 0 {
		fmt.Printf("    Metrics: ")
		first := true
		for key, value := range result.Metrics {
			if !first {
				fmt.Printf(", ")
			}
			fmt.Printf("%s=%v", key, value)
			first = false
		}
		fmt.Println()
	}

	if len(result.Issues) > 0 {
		fmt.Printf("    Issues (%d):\n", len(result.Issues))
		for _, issue := range result.Issues {
			f.displayIssue(issue)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("    Warnings (%d):\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			f.displayWarning(warning)
		}
	}
}

// displayIssue shows details of a single issue
func (f *Formatter) displayIssue(issue core.Issue) {
	severityColor := color.GreenString
	severityIcon := "•"

	switch issue.Severity {
	case core.SeverityHigh:
		severityColor = color.RedString
		severityIcon = "🔴"
	case core.SeverityMedium:
		severityColor = color.YellowString
		severityIcon = "🟡"
	case core.SeverityLow:
		severityColor = color.CyanString
		severityIcon = "🔵"
	}

	fmt.Printf("      %s [%s] %s", severityIcon, severityColor(string(issue.Severity)), issue.Message)

	if issue.Location != nil {
		fmt.Printf(" (%s", issue.Location.File)
		if issue.Location.Line > 0 {
			fmt.Printf(":%d", issue.Location.Line)
		}
		fmt.Printf(")")
	}

	if issue.Type != "" {
		fmt.Printf(" [type: %s]", issue.Type)
	}
	fmt.Println()

	if issue.Suggestion != "" {
		fmt.Printf("        💡 Suggestion: %s\n", issue.Suggestion)
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

// getCheckCategories returns a comma-separated list of check categories
func (f *Formatter) getCheckCategories(checkResults []core.CheckResult) string {
	categoryMap := make(map[string]bool)
	for _, result := range checkResults {
		categoryMap[result.Category] = true
	}

	categories := make([]string, 0, len(categoryMap))
	for category := range categoryMap {
		categories = append(categories, category)
	}

	if len(categories) == 0 {
		return "none"
	}

	result := ""
	for i, category := range categories {
		if i > 0 {
			result += ", "
		}
		result += category
	}
	return result
}

// ExitCode determines the appropriate exit code based on results
func ExitCode(result core.WorkflowResult) int {
	if result.Summary.FailedRepos > 0 {
		return 2 // Critical issues found
	}
	return 0 // Success
}

// displayChecksSummary shows summary of all checks executed across repositories
//
//nolint:gocyclo
func (f *Formatter) displayChecksSummary(results []core.RepositoryResult) {
	color.Green("=== Checks Executed ===")

	// Collect unique checks across all repositories
	checkMap := make(map[string]CheckSummary)

	for _, result := range results {
		for _, checkResult := range result.CheckResults {
			key := fmt.Sprintf("%s (%s)", checkResult.Name, checkResult.Category)
			summary, exists := checkMap[key]
			if !exists {
				summary = CheckSummary{
					Name:       checkResult.Name,
					Category:   checkResult.Category,
					ID:         checkResult.ID,
					RunCount:   0,
					PassCount:  0,
					IssueCount: 0,
				}
			}

			summary.RunCount++
			if checkResult.Status != core.StatusCritical {
				summary.PassCount++
			}
			summary.IssueCount += len(checkResult.Issues)

			checkMap[key] = summary
		}
	}

	if len(checkMap) == 0 {
		fmt.Println("No checks were executed")
		fmt.Println()
		return
	}

	// Display checks by category
	categories := make(map[string][]CheckSummary)
	for _, summary := range checkMap {
		categories[summary.Category] = append(categories[summary.Category], summary)
	}

	for category, checks := range categories {
		fmt.Printf("📋 %s:\n", category)
		for _, check := range checks {
			statusIcon := "✅"
			if check.IssueCount > 0 {
				statusIcon = "⚠️"
			}
			if check.PassCount == 0 && check.RunCount > 0 {
				statusIcon = "❌"
			}

			fmt.Printf("   %s %s (ran on %d repos", statusIcon, check.Name, check.RunCount)
			if check.IssueCount > 0 {
				fmt.Printf(", %d issues found", check.IssueCount)
			}
			fmt.Printf(")\n")
		}
	}
	fmt.Println()
}

// CheckSummary represents aggregated information about a check across repositories
type CheckSummary struct {
	Name       string
	Category   string
	ID         string
	RunCount   int
	PassCount  int
	IssueCount int
}
