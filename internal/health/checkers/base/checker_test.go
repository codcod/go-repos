package base

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/codcod/repos/internal/core"
)

func TestNewBaseChecker(t *testing.T) {
	config := core.CheckerConfig{
		Enabled:  true,
		Severity: "high",
		Timeout:  30 * time.Second,
	}

	checker := NewBaseChecker("test-id", "Test Checker", "test-category", config)

	if checker == nil {
		t.Fatal("NewBaseChecker() returned nil")
	}

	if checker.ID() != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", checker.ID())
	}

	if checker.Name() != "Test Checker" {
		t.Errorf("Expected name 'Test Checker', got %s", checker.Name())
	}

	if checker.Category() != "test-category" {
		t.Errorf("Expected category 'test-category', got %s", checker.Category())
	}

	if !checker.Config().Enabled {
		t.Error("Expected config enabled to be true")
	}

	if checker.Config().Severity != "high" {
		t.Errorf("Expected severity 'high', got %s", checker.Config().Severity)
	}
}

func TestBaseChecker_SupportsRepository(t *testing.T) {
	checker := NewBaseChecker("test", "Test", "test", core.CheckerConfig{})

	repo := core.Repository{
		Name:     "test-repo",
		Path:     "/path/to/repo",
		Language: "go",
	}

	// Default implementation should support all repositories
	if !checker.SupportsRepository(repo) {
		t.Error("Expected checker to support repository by default")
	}
}

func TestBaseChecker_Execute_Success(t *testing.T) {
	checker := NewBaseChecker("test-checker", "Test Checker", "test", core.CheckerConfig{})

	repoCtx := core.RepositoryContext{
		Repository: core.Repository{
			Name: "test-repo",
			Path: "/path/to/repo",
		},
	}

	// Mock check function that succeeds
	checkFn := func() (core.CheckResult, error) {
		return core.CheckResult{
			Status:   core.StatusHealthy,
			Score:    100,
			MaxScore: 100,
			Issues:   []core.Issue{},
		}, nil
	}

	ctx := context.Background()
	result, err := checker.Execute(ctx, repoCtx, checkFn)

	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	// Verify common fields are filled
	if result.ID != "test-checker" {
		t.Errorf("Expected ID 'test-checker', got %s", result.ID)
	}

	if result.Name != "Test Checker" {
		t.Errorf("Expected name 'Test Checker', got %s", result.Name)
	}

	if result.Category != "test" {
		t.Errorf("Expected category 'test', got %s", result.Category)
	}

	if result.Repository != "test-repo" {
		t.Errorf("Expected repository 'test-repo', got %s", result.Repository)
	}

	if result.Status != core.StatusHealthy {
		t.Errorf("Expected status healthy, got %s", result.Status)
	}

	if result.Duration <= 0 {
		t.Error("Expected positive duration")
	}

	if result.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestBaseChecker_Execute_Error(t *testing.T) {
	checker := NewBaseChecker("test-checker", "Test Checker", "test", core.CheckerConfig{})

	repoCtx := core.RepositoryContext{
		Repository: core.Repository{
			Name: "test-repo",
			Path: "/path/to/repo",
		},
	}

	// Mock check function that fails
	expectedError := fmt.Errorf("check failed")
	checkFn := func() (core.CheckResult, error) {
		return core.CheckResult{}, expectedError
	}

	ctx := context.Background()
	result, err := checker.Execute(ctx, repoCtx, checkFn)

	// Execute should not return an error - it handles errors internally
	if err != nil {
		t.Fatalf("Execute() should not return error, got: %v", err)
	}

	// Verify error is handled in result
	if result.Status != core.StatusCritical {
		t.Errorf("Expected status critical for failed check, got %s", result.Status)
	}

	if len(result.Issues) == 0 {
		t.Error("Expected issues to be present for failed check")
	}

	issue := result.Issues[0]
	if issue.Type != "execution_error" {
		t.Errorf("Expected issue type 'execution_error', got %s", issue.Type)
	}

	if issue.Severity != core.SeverityCritical {
		t.Errorf("Expected severity critical, got %s", issue.Severity)
	}

	if issue.Message != expectedError.Error() {
		t.Errorf("Expected error message '%s', got %s", expectedError.Error(), issue.Message)
	}
}

func TestNewResultBuilder(t *testing.T) {
	builder := NewResultBuilder("test-id", "Test Name", "test-category")

	if builder == nil {
		t.Fatal("NewResultBuilder() returned nil")
	}

	// The result should be initialized with basic fields
	if builder.result.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", builder.result.ID)
	}

	if builder.result.Name != "Test Name" {
		t.Errorf("Expected name 'Test Name', got %s", builder.result.Name)
	}

	if builder.result.Category != "test-category" {
		t.Errorf("Expected category 'test-category', got %s", builder.result.Category)
	}
}

func TestResultBuilder_WithStatus(t *testing.T) {
	builder := NewResultBuilder("test", "Test", "test")

	result := builder.WithStatus(core.StatusWarning).Build()

	if result.Status != core.StatusWarning {
		t.Errorf("Expected status warning, got %s", result.Status)
	}
}

func TestResultBuilder_WithScore(t *testing.T) {
	builder := NewResultBuilder("test", "Test", "test")

	result := builder.WithScore(85, 100).Build()

	if result.Score != 85 {
		t.Errorf("Expected score 85, got %d", result.Score)
	}

	if result.MaxScore != 100 {
		t.Errorf("Expected max score 100, got %d", result.MaxScore)
	}
}

func TestResultBuilder_AddIssue(t *testing.T) {
	builder := NewResultBuilder("test", "Test", "test")

	issue := core.Issue{
		Type:     "test-issue",
		Severity: core.SeverityMedium,
		Message:  "Test issue message",
	}

	result := builder.AddIssue(issue).Build()

	if len(result.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(result.Issues))
	}

	if result.Issues[0].Message != "Test issue message" {
		t.Errorf("Expected issue message 'Test issue message', got %s", result.Issues[0].Message)
	}

	// Should set status to warning for medium severity
	if result.Status != core.StatusWarning {
		t.Errorf("Expected status warning for medium severity issue, got %s", result.Status)
	}
}

func TestResultBuilder_AddWarning(t *testing.T) {
	builder := NewResultBuilder("test", "Test", "test")

	warning := core.Warning{
		Message: "Test warning message",
	}

	result := builder.AddWarning(warning).Build()

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}

	if result.Warnings[0].Message != "Test warning message" {
		t.Errorf("Expected warning message 'Test warning message', got %s", result.Warnings[0].Message)
	}

	// Should set status to warning
	if result.Status != core.StatusWarning {
		t.Errorf("Expected status warning, got %s", result.Status)
	}
}

func TestResultBuilder_AddMetric(t *testing.T) {
	builder := NewResultBuilder("test", "Test", "test")

	result := builder.AddMetric("coverage", 85.5).AddMetric("complexity", 12).Build()

	if len(result.Metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(result.Metrics))
	}

	if result.Metrics["coverage"] != 85.5 {
		t.Errorf("Expected coverage metric 85.5, got %v", result.Metrics["coverage"])
	}

	if result.Metrics["complexity"] != 12 {
		t.Errorf("Expected complexity metric 12, got %v", result.Metrics["complexity"])
	}
}

func TestResultBuilder_ChainedCalls(t *testing.T) {
	builder := NewResultBuilder("chain-test", "Chain Test", "test")

	issue := core.Issue{
		Type:    "test-issue",
		Message: "Chained issue",
	}

	warning := core.Warning{
		Message: "Chained warning",
	}

	result := builder.
		WithStatus(core.StatusWarning).
		WithScore(75, 100).
		AddIssue(issue).
		AddWarning(warning).
		AddMetric("test_metric", 42).
		Build()

	// Verify all chained operations worked
	if result.Status != core.StatusWarning {
		t.Error("Chained status not set correctly")
	}

	if result.Score != 75 || result.MaxScore != 100 {
		t.Error("Chained score not set correctly")
	}

	if len(result.Issues) != 1 || result.Issues[0].Message != "Chained issue" {
		t.Error("Chained issue not set correctly")
	}

	if len(result.Warnings) != 1 || result.Warnings[0].Message != "Chained warning" {
		t.Error("Chained warning not set correctly")
	}

	if result.Metrics["test_metric"] != 42 {
		t.Error("Chained metric not set correctly")
	}
}
