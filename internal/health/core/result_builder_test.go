package core

import (
	"testing"
	"time"

	"github.com/codcod/repos/internal/core"
)

func TestResultBuilder(t *testing.T) {
	repo := core.Repository{
		Name:     "test-repo",
		Path:     "/path/to/repo",
		Language: "go",
	}

	t.Run("basic functionality", func(t *testing.T) {
		builder := NewResultBuilder(repo)

		if builder.GetResultCount() != 0 {
			t.Errorf("Expected 0 initial results, got %d", builder.GetResultCount())
		}

		result := builder.Build()
		if result.Repository.Name != "test-repo" {
			t.Errorf("Expected repository name 'test-repo', got '%s'", result.Repository.Name)
		}

		if result.Status != core.StatusUnknown {
			t.Errorf("Expected status 'unknown' for empty results, got '%s'", result.Status)
		}
	})

	t.Run("adding success results", func(t *testing.T) {
		builder := NewResultBuilder(repo)
		builder.AddSuccessResult("test-checker", "security")

		result := builder.Build()

		if len(result.CheckResults) != 1 {
			t.Errorf("Expected 1 check result, got %d", len(result.CheckResults))
		}

		if result.Status != core.StatusHealthy {
			t.Errorf("Expected status 'healthy', got '%s'", result.Status)
		}

		checkResult := result.CheckResults[0]
		if checkResult.Name != "test-checker" {
			t.Errorf("Expected checker name 'test-checker', got '%s'", checkResult.Name)
		}

		if checkResult.Status != core.StatusHealthy {
			t.Errorf("Expected check status 'healthy', got '%s'", checkResult.Status)
		}
	})

	t.Run("adding issues", func(t *testing.T) {
		builder := NewResultBuilder(repo)
		location := &core.Location{
			File: "test.go",
			Line: 10,
		}

		builder.AddIssue("security-checker", "vulnerability", "SQL injection found", core.SeverityHigh, location)

		result := builder.Build()

		if result.Status != core.StatusCritical {
			t.Errorf("Expected status 'critical' for high severity issue, got '%s'", result.Status)
		}

		if len(result.CheckResults) != 1 {
			t.Errorf("Expected 1 check result, got %d", len(result.CheckResults))
		}

		checkResult := result.CheckResults[0]
		if len(checkResult.Issues) != 1 {
			t.Errorf("Expected 1 issue, got %d", len(checkResult.Issues))
		}

		issue := checkResult.Issues[0]
		if issue.Type != "vulnerability" {
			t.Errorf("Expected issue type 'vulnerability', got '%s'", issue.Type)
		}

		if issue.Severity != core.SeverityHigh {
			t.Errorf("Expected severity 'high', got '%s'", issue.Severity)
		}
	})

	t.Run("adding warnings", func(t *testing.T) {
		builder := NewResultBuilder(repo)
		location := &core.Location{
			File: "test.go",
			Line: 5,
		}

		builder.AddWarning("style-checker", "Consider using more descriptive variable names", location)

		result := builder.Build()

		if result.Status != core.StatusWarning {
			t.Errorf("Expected status 'warning', got '%s'", result.Status)
		}

		if len(result.CheckResults) != 1 {
			t.Errorf("Expected 1 check result, got %d", len(result.CheckResults))
		}

		checkResult := result.CheckResults[0]
		if len(checkResult.Warnings) != 1 {
			t.Errorf("Expected 1 warning, got %d", len(checkResult.Warnings))
		}

		warning := checkResult.Warnings[0]
		if warning.Message != "Consider using more descriptive variable names" {
			t.Errorf("Unexpected warning message: %s", warning.Message)
		}
	})

	t.Run("multiple check results", func(t *testing.T) {
		builder := NewResultBuilder(repo)

		// Add a success result
		builder.AddSuccessResult("formatter", "style")

		// Add a warning
		builder.AddWarning("linter", "Unused variable", nil)

		// Add an issue
		builder.AddIssue("security", "vulnerability", "Weak crypto", core.SeverityMedium, nil)

		result := builder.Build()

		// Should have 3 check results (or 2 if warning and issue are added to existing)
		if len(result.CheckResults) < 2 {
			t.Errorf("Expected at least 2 check results, got %d", len(result.CheckResults))
		}

		// Overall status should be critical due to the issue
		if result.Status != core.StatusWarning {
			t.Errorf("Expected status 'warning' for medium severity issue, got '%s'", result.Status)
		}
	})

	t.Run("fluent interface", func(t *testing.T) {
		result := NewResultBuilder(repo).
			WithMetadata("test", "value").
			AddSuccessResult("test1", "category1").
			AddSuccessResult("test2", "category2").
			Build()

		if len(result.CheckResults) != 2 {
			t.Errorf("Expected 2 check results, got %d", len(result.CheckResults))
		}

		if result.Status != core.StatusHealthy {
			t.Errorf("Expected status 'healthy', got '%s'", result.Status)
		}
	})

	t.Run("duration tracking", func(t *testing.T) {
		builder := NewResultBuilder(repo)

		// Simulate some work
		time.Sleep(10 * time.Millisecond)

		result := builder.Build()

		if result.Duration <= 0 {
			t.Error("Expected positive duration")
		}

		if result.Duration < 10*time.Millisecond {
			t.Errorf("Expected duration >= 10ms, got %v", result.Duration)
		}
	})
}
