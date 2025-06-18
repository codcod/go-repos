package orchestration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/codcod/repos/internal/core"
)

// Mock implementations for testing

type mockCheckerRegistry struct {
	checkers map[string]core.Checker
}

func (m *mockCheckerRegistry) Register(checker core.Checker) {
	if m.checkers == nil {
		m.checkers = make(map[string]core.Checker)
	}
	m.checkers[checker.ID()] = checker
}

func (m *mockCheckerRegistry) Unregister(checkerID string) {
	if m.checkers != nil {
		delete(m.checkers, checkerID)
	}
}

func (m *mockCheckerRegistry) GetChecker(checkerID string) (core.Checker, error) {
	if checker, exists := m.checkers[checkerID]; exists {
		return checker, nil
	}
	return nil, fmt.Errorf("checker %s not found", checkerID)
}

func (m *mockCheckerRegistry) GetCheckers() []core.Checker {
	checkers := make([]core.Checker, 0, len(m.checkers))
	for _, checker := range m.checkers {
		checkers = append(checkers, checker)
	}
	return checkers
}

func (m *mockCheckerRegistry) GetCheckersForRepository(repo core.Repository) []core.Checker {
	return m.GetCheckers()
}

func (m *mockCheckerRegistry) GetCheckersByCategory(category string) []core.Checker {
	return m.GetCheckers()
}

func (m *mockCheckerRegistry) GetEnabledCheckers(config map[string]core.CheckerConfig) []core.Checker {
	return m.GetCheckers()
}

func (m *mockCheckerRegistry) RunChecker(ctx context.Context, checkerID string, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	checker, err := m.GetChecker(checkerID)
	if err != nil {
		return core.CheckResult{}, err
	}
	return checker.Check(ctx, repoCtx)
}

func (m *mockCheckerRegistry) RunCheckers(ctx context.Context, checkerIDs []string, repoCtx core.RepositoryContext) ([]core.CheckResult, error) {
	results := make([]core.CheckResult, 0, len(checkerIDs))
	for _, id := range checkerIDs {
		result, err := m.RunChecker(ctx, id, repoCtx)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (m *mockCheckerRegistry) RunAllEnabledCheckers(ctx context.Context, repoCtx core.RepositoryContext, config map[string]core.CheckerConfig) ([]core.CheckResult, error) {
	results := make([]core.CheckResult, 0, len(m.checkers))
	for _, checker := range m.checkers {
		result, err := checker.Check(ctx, repoCtx)
		if err != nil {
			return nil, fmt.Errorf("checker %s failed: %w", checker.ID(), err)
		}
		results = append(results, result)
	}
	return results, nil
}

type mockAnalyzerRegistry struct {
	analyzers map[string]core.Analyzer
}

func (m *mockAnalyzerRegistry) Register(analyzer core.Analyzer) {
	if m.analyzers == nil {
		m.analyzers = make(map[string]core.Analyzer)
	}
	m.analyzers[analyzer.Language()] = analyzer
}

func (m *mockAnalyzerRegistry) Unregister(language string) {
	if m.analyzers != nil {
		delete(m.analyzers, language)
	}
}

func (m *mockAnalyzerRegistry) GetAnalyzer(language string) (core.Analyzer, error) {
	if analyzer, exists := m.analyzers[language]; exists {
		return analyzer, nil
	}
	return nil, fmt.Errorf("analyzer for %s not found", language)
}

func (m *mockAnalyzerRegistry) GetAnalyzers() []core.Analyzer {
	analyzers := make([]core.Analyzer, 0, len(m.analyzers))
	for _, analyzer := range m.analyzers {
		analyzers = append(analyzers, analyzer)
	}
	return analyzers
}

func (m *mockAnalyzerRegistry) GetSupportedLanguages() []string {
	languages := make([]string, 0, len(m.analyzers))
	for lang := range m.analyzers {
		languages = append(languages, lang)
	}
	return languages
}

type mockChecker struct {
	id       string
	name     string
	category string
	config   core.CheckerConfig
	result   core.CheckResult
	err      error
}

func (m *mockChecker) ID() string {
	return m.id
}

func (m *mockChecker) Name() string {
	return m.name
}

func (m *mockChecker) Category() string {
	return m.category
}

func (m *mockChecker) Config() core.CheckerConfig {
	return m.config
}

func (m *mockChecker) Check(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	if m.err != nil {
		return core.CheckResult{}, m.err
	}
	return m.result, nil
}

func (m *mockChecker) SupportsRepository(repo core.Repository) bool {
	return true
}

type mockConfig struct {
	engineConfig core.EngineConfig
}

func (m *mockConfig) GetCheckerConfig(checkerID string) (core.CheckerConfig, bool) {
	return core.CheckerConfig{Enabled: true}, true
}

func (m *mockConfig) GetAnalyzerConfig(language string) (core.AnalyzerConfig, bool) {
	return core.AnalyzerConfig{Enabled: true}, true
}

func (m *mockConfig) GetReporterConfig(reporterID string) (core.ReporterConfig, bool) {
	return core.ReporterConfig{Enabled: true}, true
}

func (m *mockConfig) GetEngineConfig() core.EngineConfig {
	if m.engineConfig.MaxConcurrency == 0 {
		m.engineConfig.MaxConcurrency = 5
	}
	if m.engineConfig.Timeout == 0 {
		m.engineConfig.Timeout = 30 * time.Second
	}
	return m.engineConfig
}

type mockLogger struct {
	logs []string
}

func (m *mockLogger) Info(msg string, fields ...core.Field) {
	m.logs = append(m.logs, fmt.Sprintf("INFO: %s", msg))
}

func (m *mockLogger) Debug(msg string, fields ...core.Field) {
	m.logs = append(m.logs, fmt.Sprintf("DEBUG: %s", msg))
}

func (m *mockLogger) Warn(msg string, fields ...core.Field) {
	m.logs = append(m.logs, fmt.Sprintf("WARN: %s", msg))
}

func (m *mockLogger) Error(msg string, fields ...core.Field) {
	m.logs = append(m.logs, fmt.Sprintf("ERROR: %s", msg))
}

func (m *mockLogger) Fatal(msg string, fields ...core.Field) {
	m.logs = append(m.logs, fmt.Sprintf("FATAL: %s", msg))
}

func TestNewEngine(t *testing.T) {
	checkerRegistry := &mockCheckerRegistry{}
	analyzerRegistry := &mockAnalyzerRegistry{}
	config := &mockConfig{}
	logger := &mockLogger{}

	engine := NewEngine(checkerRegistry, analyzerRegistry, config, logger)

	if engine == nil {
		t.Fatal("NewEngine() returned nil")
	}

	if engine.maxConcurrency != 5 {
		t.Errorf("Expected max concurrency 5, got %d", engine.maxConcurrency)
	}

	if engine.timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", engine.timeout)
	}
}

func TestEngine_ExecuteHealthCheck_EmptyRepos(t *testing.T) {
	checkerRegistry := &mockCheckerRegistry{}
	analyzerRegistry := &mockAnalyzerRegistry{}
	config := &mockConfig{}
	logger := &mockLogger{}

	engine := NewEngine(checkerRegistry, analyzerRegistry, config, logger)

	ctx := context.Background()
	result, err := engine.ExecuteHealthCheck(ctx, []core.Repository{})

	if err != nil {
		t.Fatalf("ExecuteHealthCheck failed with empty repos: %v", err)
	}

	if result == nil {
		t.Fatal("ExecuteHealthCheck returned nil result")
	}

	if result.TotalRepos != 0 {
		t.Errorf("Expected 0 total repos, got %d", result.TotalRepos)
	}

	if len(result.RepositoryResults) != 0 {
		t.Errorf("Expected 0 repository results, got %d", len(result.RepositoryResults))
	}

	if result.Duration <= 0 {
		t.Error("Expected positive duration")
	}
}

func TestEngine_ExecuteHealthCheck_SingleRepo(t *testing.T) {
	checkerRegistry := &mockCheckerRegistry{}
	analyzerRegistry := &mockAnalyzerRegistry{}
	config := &mockConfig{}
	logger := &mockLogger{}

	// Register a mock checker
	mockChecker := &mockChecker{
		id:       "test-checker",
		name:     "Test Checker",
		category: "test",
		result: core.CheckResult{
			ID:       "test-checker",
			Name:     "Test Checker",
			Category: "test",
			Status:   core.StatusHealthy,
			Score:    100,
			MaxScore: 100,
			Issues:   []core.Issue{},
			Warnings: []core.Warning{},
			Metrics:  map[string]interface{}{"test": "passed"},
		},
	}
	checkerRegistry.Register(mockChecker)

	engine := NewEngine(checkerRegistry, analyzerRegistry, config, logger)

	repos := []core.Repository{
		{
			Name: "test-repo",
			Path: "/path/to/repo",
		},
	}

	ctx := context.Background()
	result, err := engine.ExecuteHealthCheck(ctx, repos)

	if err != nil {
		t.Fatalf("ExecuteHealthCheck failed: %v", err)
	}

	if result == nil {
		t.Fatal("ExecuteHealthCheck returned nil result")
	}

	if result.TotalRepos != 1 {
		t.Errorf("Expected 1 total repo, got %d", result.TotalRepos)
	}

	if len(result.RepositoryResults) != 1 {
		t.Errorf("Expected 1 repository result, got %d", len(result.RepositoryResults))
	}

	if result.Summary.SuccessfulRepos != 1 {
		t.Errorf("Expected 1 successful repo, got %d", result.Summary.SuccessfulRepos)
	}
}
