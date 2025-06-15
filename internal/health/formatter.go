package health

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

// FormatHealthReport formats the health report according to the specified format
func FormatHealthReport(report HealthReport, options HealthOptions) (string, error) {
	switch strings.ToLower(options.Format) {
	case "json":
		return formatJSON(report)
	case "html":
		return formatHTML(report)
	case "table", "":
		return formatTable(report), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", options.Format)
	}
}

// formatJSON formats the report as JSON
func formatJSON(report HealthReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// formatTable formats the report as a colored table
func formatTable(report HealthReport) string {
	var output strings.Builder

	// Header
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("Repository Health Report\n"))
	output.WriteString(strings.Repeat("=", 50) + "\n\n")

	// Summary
	summary := report.Summary
	output.WriteString(color.New(color.FgWhite, color.Bold).Sprint("Summary:\n"))
	output.WriteString(fmt.Sprintf("  Total repositories: %d\n", summary.Total))

	if summary.Healthy > 0 {
		output.WriteString(fmt.Sprintf("  %s %d healthy\n",
			color.New(color.FgGreen).Sprint("✅"), summary.Healthy))
	}
	if summary.Warning > 0 {
		output.WriteString(fmt.Sprintf("  %s %d with warnings\n",
			color.New(color.FgYellow).Sprint("⚠️"), summary.Warning))
	}
	if summary.Critical > 0 {
		output.WriteString(fmt.Sprintf("  %s %d critical\n",
			color.New(color.FgRed).Sprint("❌"), summary.Critical))
	}
	if summary.Unknown > 0 {
		output.WriteString(fmt.Sprintf("  %s %d unknown\n",
			color.New(color.FgMagenta).Sprint("❓"), summary.Unknown))
	}

	output.WriteString("\n")

	// Repository details
	for _, repoHealth := range report.Repositories {
		output.WriteString(formatRepositoryHealth(repoHealth))
		output.WriteString("\n")
	}

	return output.String()
}

// formatRepositoryHealth formats a single repository's health status
func formatRepositoryHealth(repoHealth RepositoryHealth) string {
	var output strings.Builder

	// Repository header
	statusIcon := getStatusIcon(repoHealth.Status)
	statusColor := getStatusColor(repoHealth.Status)

	repoName := color.New(color.FgCyan, color.Bold).SprintFunc()(repoHealth.Repository.Name)
	status := statusColor.SprintFunc()(string(repoHealth.Status))

	output.WriteString(fmt.Sprintf("%s %s (%s) - Score: %d/100\n",
		statusIcon, repoName, status, repoHealth.Score))

	if repoHealth.Summary != "" {
		output.WriteString(fmt.Sprintf("   %s\n", repoHealth.Summary))
	}

	// Show failed/warning checks
	for _, check := range repoHealth.Checks {
		if check.Status != HealthStatusHealthy {
			checkIcon := getStatusIcon(check.Status)
			checkColor := getStatusColor(check.Status)
			output.WriteString(fmt.Sprintf("   %s %s: %s\n",
				checkIcon, check.Name, checkColor.SprintFunc()(check.Message)))

			if check.Details != "" {
				output.WriteString(fmt.Sprintf("      %s\n",
					color.New(color.FgHiBlack).SprintFunc()(check.Details)))
			}
		}
	}

	return output.String()
}

// formatHTML formats the report as HTML
func formatHTML(report HealthReport) (string, error) {
	var output strings.Builder

	output.WriteString(`<!DOCTYPE html>
<html>
<head>
    <title>Repository Health Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px; }
        .summary { background: #ecf0f1; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .repo { margin: 20px 0; padding: 15px; border-left: 4px solid #bdc3c7; }
        .healthy { border-left-color: #27ae60; }
        .warning { border-left-color: #f39c12; }
        .critical { border-left-color: #e74c3c; }
        .unknown { border-left-color: #9b59b6; }
        .check { margin: 5px 0; padding: 5px 10px; }
        .check-healthy { background: #d5f4e6; }
        .check-warning { background: #fef9e7; }
        .check-critical { background: #fadbd8; }
        .score { font-weight: bold; float: right; }
    </style>
</head>
<body>
    <h1 class="header">Repository Health Report</h1>
`)

	// Summary
	summary := report.Summary
	output.WriteString(`    <div class="summary">
        <h2>Summary</h2>
        <ul>`)
	output.WriteString(fmt.Sprintf("<li>Total repositories: %d</li>", summary.Total))
	output.WriteString(fmt.Sprintf("<li>✅ Healthy: %d</li>", summary.Healthy))
	output.WriteString(fmt.Sprintf("<li>⚠️ Warning: %d</li>", summary.Warning))
	output.WriteString(fmt.Sprintf("<li>❌ Critical: %d</li>", summary.Critical))
	output.WriteString(fmt.Sprintf("<li>❓ Unknown: %d</li>", summary.Unknown))
	output.WriteString(`        </ul>
    </div>
`)

	// Repository details
	for _, repoHealth := range report.Repositories {
		statusClass := strings.ToLower(string(repoHealth.Status))
		output.WriteString(fmt.Sprintf(`    <div class="repo %s">
        <h3>%s <span class="score">%d/100</span></h3>
        <p>%s</p>
`, repoHealth.Repository.Name, statusClass, repoHealth.Score, repoHealth.Summary))

		for _, check := range repoHealth.Checks {
			if check.Status != HealthStatusHealthy {
				checkClass := fmt.Sprintf("check--%s", strings.ToLower(string(check.Status)))
				output.WriteString(fmt.Sprintf(`        <div class="check %s">
            <strong>%s:</strong> %s
`, checkClass, check.Name, check.Message))
				if check.Details != "" {
					output.WriteString(fmt.Sprintf("            <br><small>%s</small>", check.Details))
				}
				output.WriteString("        </div>\n")
			}
		}

		output.WriteString("    </div>\n")
	}

	output.WriteString(`    <footer>
        <p><small>Generated at: ` + report.GeneratedAt.Format("2006-01-02 15:04:05") + `</small></p>
    </footer>
</body>
</html>`)

	return output.String(), nil
}

// PrintHealthReport prints the health report to stdout or file
func PrintHealthReport(report HealthReport, options HealthOptions) error {
	formatted, err := FormatHealthReport(report, options)
	if err != nil {
		return fmt.Errorf("failed to format report: %w", err)
	}

	if options.OutputFile != "" {
		err := os.WriteFile(options.OutputFile, []byte(formatted), 0600)
		if err != nil {
			return fmt.Errorf("failed to write to file %s: %w", options.OutputFile, err)
		}
		fmt.Printf("Health report saved to: %s\n", options.OutputFile)
	} else {
		fmt.Print(formatted)
	}

	return nil
}

// PrintSummaryTable prints a compact summary table
func PrintSummaryTable(report HealthReport) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		color.New(color.FgWhite, color.Bold).Sprint("Repository"),
		color.New(color.FgWhite, color.Bold).Sprint("Status"),
		color.New(color.FgWhite, color.Bold).Sprint("Score"),
		color.New(color.FgWhite, color.Bold).Sprint("Issues"))

	_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		strings.Repeat("-", 20),
		strings.Repeat("-", 10),
		strings.Repeat("-", 5),
		strings.Repeat("-", 30))

	for _, repoHealth := range report.Repositories {
		statusIcon := getStatusIcon(repoHealth.Status)
		statusColor := getStatusColor(repoHealth.Status)

		repoName := repoHealth.Repository.Name
		if len(repoName) > 20 {
			repoName = repoName[:17] + "..."
		}

		status := statusColor.SprintFunc()(statusIcon + " " + string(repoHealth.Status))
		score := fmt.Sprintf("%d", repoHealth.Score)

		// Count issues
		critical := countByStatus(repoHealth.Checks, HealthStatusCritical)
		warning := countByStatus(repoHealth.Checks, HealthStatusWarning)

		var issues string
		if critical > 0 {
			issues = fmt.Sprintf("%d critical", critical)
			if warning > 0 {
				issues += fmt.Sprintf(", %d warnings", warning)
			}
		} else if warning > 0 {
			issues = fmt.Sprintf("%d warnings", warning)
		} else {
			issues = "none"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", repoName, status, score, issues)
	}

	_ = w.Flush()
}

// Helper functions for formatting
func getStatusIcon(status HealthStatus) string {
	switch status {
	case HealthStatusHealthy:
		return "✅"
	case HealthStatusWarning:
		return "⚠️"
	case HealthStatusCritical:
		return "❌"
	default:
		return "❓"
	}
}

func getStatusColor(status HealthStatus) *color.Color {
	switch status {
	case HealthStatusHealthy:
		return color.New(color.FgGreen)
	case HealthStatusWarning:
		return color.New(color.FgYellow)
	case HealthStatusCritical:
		return color.New(color.FgRed)
	default:
		return color.New(color.FgMagenta)
	}
}
