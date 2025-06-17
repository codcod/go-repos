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

func TestFormatter_countIssues(t *testing.T) {
	formatter := NewFormatter(false)

	// Test empty check results
	emptyResults := []core.CheckResult{}
	if count := formatter.countIssues(emptyResults); count != 0 {
		t.Errorf("Expected 0 issues for empty results, got %d", count)
	}

	// Test check results with issues
	resultsWithIssues := []core.CheckResult{
		{
			Issues: []core.Issue{
				{Type: "test", Message: "Issue 1"},
				{Type: "test", Message: "Issue 2"},
			},
		},
		{
			Issues: []core.Issue{
				{Type: "test", Message: "Issue 3"},
			},
		},
	}
	if count := formatter.countIssues(resultsWithIssues); count != 3 {
		t.Errorf("Expected 3 issues, got %d", count)
	}

	// Test check results without issues
	resultsWithoutIssues := []core.CheckResult{
		{Issues: []core.Issue{}},
		{Issues: nil},
	}
	if count := formatter.countIssues(resultsWithoutIssues); count != 0 {
		t.Errorf("Expected 0 issues for results without issues, got %d", count)
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
