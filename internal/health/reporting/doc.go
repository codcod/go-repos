/*
Package reporting provides result formatting and output functionality for health analysis.

This package handles the presentation and formatting of health analysis results,
offering both compact and detailed views with configurable verbosity levels.

# Features

  - Colorized console output with status indicators
  - Compact summary view for CI/CD environments
  - Detailed verbose output for debugging and investigation
  - Issue severity classification and highlighting
  - Timing and performance information
  - Exit code determination for automated workflows

# Usage

Create a formatter and display results:

	// Create formatter with verbose output
	formatter := reporting.NewFormatter(true)

	// Display the results
	formatter.DisplayResults(workflowResult)

	// Get appropriate exit code for the shell
	exitCode := reporting.ExitCode(workflowResult)
	os.Exit(exitCode)

# Output Formats

## Compact Mode (verbose=false)

Shows a summary with repository status indicators:
  - ✅ Repository: All checks passed
  - ⚠️  Repository: X issues found
  - ❌ Repository: Error occurred

## Verbose Mode (verbose=true)

Includes detailed information:
  - Complete summary statistics
  - Individual repository analysis results
  - Checker results with issue details
  - Analysis metrics and findings
  - Execution timing information

# Status Indicators

  - ✅ Success/Healthy status
  - ⚠️  Warning status or issues found
  - ❌ Critical errors or failures

# Issue Severity Colors

  - 🔴 High severity (red)
  - 🟡 Medium severity (yellow)
  - 🔵 Low severity (cyan)
  - 🟢 Info level (green)

# Exit Codes

The package provides standardized exit codes:
  - 0: Success, no critical issues
  - 2: Critical issues found, requires attention

These codes are suitable for use in CI/CD pipelines and automated workflows.
*/
package reporting
