package reporting

import (
	"testing"
	"time"

	"github.com/codcod/repos/internal/core"
)

func TestNewFormatter(t *testing.T) {
	// Test creating formatter with verbose=false
	formatter := NewFormatter(false)
	if formatter == nil {
		t.Fatal("NewFormatter() returned nil")
	}
	if formatter.verbose {
		t.Error("Expected verbose to be false")
	}

	// Test creating formatter with verbose=true
	verboseFormatter := NewFormatter(true)
	if !verboseFormatter.verbose {
		t.Error("Expected verbose to be true")
	}
}

func TestExitCode(t *testing.T) {
	// Test successful result
	successResult := core.WorkflowResult{
		Summary: core.WorkflowSummary{
			SuccessfulRepos: 5,
			FailedRepos:     0,
		},
	}
	if ExitCode(successResult) != 0 {
		t.Errorf("Expected exit code 0 for successful result, got %d", ExitCode(successResult))
	}

	// Test failed result
	failedResult := core.WorkflowResult{
		Summary: core.WorkflowSummary{
			SuccessfulRepos: 3,
			FailedRepos:     2,
		},
	}
	if ExitCode(failedResult) != 2 {
		t.Errorf("Expected exit code 2 for failed result, got %d", ExitCode(failedResult))
	}
}

func TestFormatter_DisplayResults_Compact(t *testing.T) {
	formatter := NewFormatter(false) // Non-verbose mode

	// Create test results
	workflowResult := core.WorkflowResult{
		Summary: core.WorkflowSummary{
			SuccessfulRepos: 2,
			FailedRepos:     1,
			AverageScore:    85,
			TotalIssues:     3,
		},
		RepositoryResults: []core.RepositoryResult{
			{
				Repository: core.Repository{Name: "repo1", Path: "/path/to/repo1"},
				Status:     core.StatusHealthy,
				CheckResults: []core.CheckResult{
					{Name: "Test Checker", Status: core.StatusHealthy, Issues: []core.Issue{}},
				},
			},
			{
				Repository: core.Repository{Name: "repo2", Path: "/path/to/repo2"},
				Status:     core.StatusCritical,
				Error:      "Failed to run checks",
			},
		},
	}

	// Test that it doesn't panic and processes the results
	formatter.DisplayResults(workflowResult)

	// If we get here without panicking, the basic functionality works
	t.Log("DisplayResults completed successfully for compact mode")
}

func TestFormatter_DisplayResults_Verbose(t *testing.T) {
	formatter := NewFormatter(true) // Verbose mode

	// Create test results with analysis
	workflowResult := core.WorkflowResult{
		StartTime: time.Now().Add(-5 * time.Second),
		EndTime:   time.Now(),
		Summary: core.WorkflowSummary{
			SuccessfulRepos: 1,
			FailedRepos:     0,
		},
		RepositoryResults: []core.RepositoryResult{
			{
				Repository: core.Repository{
					Name:     "detailed-repo",
					Path:     "/path/to/detailed-repo",
					Language: "go",
				},
				Status:   core.StatusHealthy,
				Score:    95,
				MaxScore: 100,
				AnalysisResult: &core.AnalysisResult{
					Language: "go",
					Files: map[string]*core.FileAnalysis{
						"main.go": {Path: "main.go", Language: "go"},
					},
					Functions: []core.FunctionInfo{
						{Name: "main", File: "main.go", Line: 10},
					},
					Metrics: map[string]interface{}{
						"complexity": 5.5,
					},
				},
				CheckResults: []core.CheckResult{
					{
						Name:     "Go Linter",
						Category: "quality",
						Status:   core.StatusHealthy,
						Score:    90,
						MaxScore: 100,
						Issues:   []core.Issue{},
						Warnings: []core.Warning{},
					},
				},
			},
		},
	}

	// Test that verbose mode works without panicking
	formatter.DisplayResults(workflowResult)

	t.Log("DisplayResults completed successfully for verbose mode")
}

func TestFormatter_DisplayResults_WithErrors(t *testing.T) {
	formatter := NewFormatter(false)

	// Create test results with errors
	workflowResult := core.WorkflowResult{
		Summary: core.WorkflowSummary{
			SuccessfulRepos: 0,
			FailedRepos:     2,
		},
		RepositoryResults: []core.RepositoryResult{
			{
				Repository: core.Repository{Name: "error-repo1"},
				Status:     core.StatusCritical,
				Error:      "Permission denied",
			},
			{
				Repository: core.Repository{Name: "error-repo2"},
				Status:     core.StatusCritical,
				Error:      "Repository not found",
			},
		},
	}

	// Test that error handling works without panicking
	formatter.DisplayResults(workflowResult)

	t.Log("DisplayResults completed successfully with error handling")
}

func TestFormatter_DisplayResults_EmptyResults(t *testing.T) {
	formatter := NewFormatter(false)

	// Create empty results
	workflowResult := core.WorkflowResult{
		Summary: core.WorkflowSummary{
			SuccessfulRepos: 0,
			FailedRepos:     0,
			AverageScore:    0,
			TotalIssues:     0,
		},
		RepositoryResults: []core.RepositoryResult{},
	}

	// Test that empty results are handled gracefully
	formatter.DisplayResults(workflowResult)

	t.Log("DisplayResults completed successfully with empty results")
}

func TestDisplayResults_EnhancedDetails(t *testing.T) {
	formatter := NewFormatter(false) // compact mode

	// Create a sample result with detailed check information
	result := core.WorkflowResult{
		StartTime: time.Now().Add(-5 * time.Second),
		EndTime:   time.Now(),
		Duration:  5 * time.Second,
		Summary: core.WorkflowSummary{
			SuccessfulRepos: 1,
			FailedRepos:     0,
			AverageScore:    85,
			TotalIssues:     2,
			StatusCounts: map[core.HealthStatus]int{
				core.StatusHealthy: 1,
			},
		},
		RepositoryResults: []core.RepositoryResult{
			{
				Repository: core.Repository{
					Name:     "test-repo",
					Path:     "/path/to/test-repo",
					Language: "python",
				},
				Status:   core.StatusHealthy,
				Score:    85,
				MaxScore: 100,
				CheckResults: []core.CheckResult{
					{
						ID:       "security-scan",
						Name:     "Security Scanner",
						Category: "security",
						Status:   core.StatusHealthy,
						Score:    90,
						MaxScore: 100,
						Duration: 250 * time.Millisecond,
						Issues: []core.Issue{
							{
								Type:     "potential_vulnerability",
								Severity: core.SeverityMedium,
								Message:  "Potential SQL injection vulnerability detected",
								Location: &core.Location{
									File: "app.py",
									Line: 42,
								},
								Suggestion: "Use parameterized queries to prevent SQL injection",
							},
						},
						Metrics: map[string]interface{}{
							"files_scanned":         15,
							"vulnerabilities_found": 1,
						},
					},
					{
						ID:       "license-check",
						Name:     "License Compliance",
						Category: "compliance",
						Status:   core.StatusWarning,
						Score:    80,
						MaxScore: 100,
						Duration: 100 * time.Millisecond,
						Issues: []core.Issue{
							{
								Type:       "missing_license",
								Severity:   core.SeverityLow,
								Message:    "No license file found in repository",
								Suggestion: "Add a LICENSE file to clarify usage rights",
							},
						},
						Metrics: map[string]interface{}{
							"license_files_found": 0,
						},
					},
				},
			},
		},
	}

	// Test that enhanced formatting doesn't panic and produces output
	t.Log("=== Enhanced Health Check Output ===")
	formatter.DisplayResults(result)
	t.Log("DisplayResults completed successfully with enhanced details")
}

// TestComplexityFunctions tests the complexity-related helper functions
func TestComplexityFunctions(t *testing.T) {
	formatter := NewFormatter(true)

	// Test sortFunctionsByComplexity
	functions := []core.FunctionInfo{
		{Name: "low", Complexity: 2},
		{Name: "high", Complexity: 20},
		{Name: "medium", Complexity: 8},
		{Name: "very_high", Complexity: 30},
	}

	formatter.sortFunctionsByComplexity(functions)

	// Verify sorted order (highest first)
	expectedOrder := []string{"very_high", "high", "medium", "low"}
	for i, fn := range functions {
		if fn.Name != expectedOrder[i] {
			t.Errorf("Expected function at position %d to be '%s', got '%s'",
				i, expectedOrder[i], fn.Name)
		}
	}
}
