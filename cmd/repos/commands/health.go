// Package commands provides command implementations for the repos CLI
package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/core"
	"github.com/codcod/repos/internal/errors"
	"github.com/codcod/repos/internal/health"
	healthconfig "github.com/codcod/repos/internal/health/config"
	"github.com/codcod/repos/internal/observability"
)

// HealthCommand handles the health check command execution
type HealthCommand struct {
	config    *HealthConfig
	validator *healthconfig.ConfigValidator
	executor  *HealthExecutor
	logger    *observability.StructuredLogger
	metrics   *observability.MetricsCollector
}

// HealthConfig contains all configuration for health checks
type HealthConfig struct {
	ConfigPath     string
	Categories     []string
	Pipeline       string
	Parallel       bool
	Timeout        time.Duration
	DryRun         bool
	Verbose        bool
	Tag            string
	BasicConfig    string // Path to basic repo config
	ListCategories bool   // List available categories and checkers
	GenConfig      bool   // Generate configuration template
}

// NewHealthCommand creates a new health command instance
func NewHealthCommand(healthConfig *HealthConfig) *HealthCommand {
	logger := observability.NewStructuredLogger(observability.LevelInfo)
	if healthConfig.Verbose {
		logger = observability.NewStructuredLogger(observability.LevelDebug)
	}

	return &HealthCommand{
		config:    healthConfig,
		validator: healthconfig.NewConfigValidator(),
		executor:  NewHealthExecutor(),
		logger:    logger.WithPrefix("health-cmd"),
		metrics:   observability.NewMetricsCollector(),
	}
}

// Validate validates the health command configuration
func (hc *HealthCommand) Validate() error {
	hc.logger.Debug("validating health command configuration")

	// Config path is optional - if empty, we'll use default config
	if hc.config.ConfigPath == "" {
		hc.config.ConfigPath = "orchestration.yaml" // Default config file name
		hc.logger.Debug("using default config path", core.String("path", hc.config.ConfigPath))
	}

	if hc.config.Timeout <= 0 {
		hc.config.Timeout = 5 * time.Minute // Set default timeout
	}

	if hc.config.Timeout > 60*time.Minute {
		return errors.NewContextualError("validate_config", fmt.Errorf("timeout too high: %v (max: 60m)", hc.config.Timeout)).
			WithContext("field", "timeout").
			WithContext("value", hc.config.Timeout)
	}

	return nil
}

// Execute runs the health check command
func (hc *HealthCommand) Execute(ctx context.Context) error {
	operationLogger, operationDone := hc.logger.StartOperation("health_check")
	defer operationDone()

	err := hc.metrics.MeasureOperation("health_command", func() error {
		if err := hc.Validate(); err != nil {
			return err
		}

		return hc.executor.Run(ctx, hc.config, operationLogger, hc.metrics)
	})

	// Print metrics summary if verbose
	if hc.config.Verbose {
		hc.metrics.PrintSummary()
	}

	return err
}

// HealthExecutor handles the actual execution of health checks
type HealthExecutor struct {
	logger core.Logger
}

// NewHealthExecutor creates a new health executor
func NewHealthExecutor() *HealthExecutor {
	return &HealthExecutor{
		logger: &simpleLogger{},
	}
}

// SetLogger sets a custom logger for the executor
func (he *HealthExecutor) SetLogger(logger core.Logger) {
	he.logger = logger
}

// Run executes the health checks with the given configuration
func (he *HealthExecutor) Run(ctx context.Context, config *HealthConfig, logger *observability.StructuredLogger, metrics *observability.MetricsCollector) error {
	opLogger := logger.WithField("operation", "health_executor")

	// Setup context with timeout
	ctx, cancel := he.setupContext(ctx, config, opLogger)
	defer cancel()

	// Load and validate configuration
	advConfig, err := he.loadAndValidateConfig(config, opLogger, metrics)
	if err != nil {
		return err
	}

	// Load and prepare repositories
	coreRepos, err := he.loadAndPrepareRepositories(config, opLogger, metrics)
	if err != nil {
		return err
	}

	// Handle case where no repositories are found
	if coreRepos == nil {
		return nil
	}

	// Execute dry run if requested
	if config.DryRun {
		opLogger.Info("dry run mode - would execute health checks",
			core.String("pipeline", config.Pipeline),
			core.Int("repositories", len(coreRepos)))
		return nil
	}

	// Execute and report results
	return he.executeAndReportResults(ctx, coreRepos, advConfig, config, opLogger, metrics)
}

// setupContext configures the context with timeout if specified
func (he *HealthExecutor) setupContext(ctx context.Context, config *HealthConfig, logger *observability.StructuredLogger) (context.Context, context.CancelFunc) {
	if config.Timeout <= 0 {
		return ctx, func() {}
	}

	logger.Debug("timeout configured", core.String("timeout", config.Timeout.String()))
	return context.WithTimeout(ctx, config.Timeout)
}

// loadAndValidateConfig loads the advanced configuration
func (he *HealthExecutor) loadAndValidateConfig(config *HealthConfig, logger *observability.StructuredLogger, metrics *observability.MetricsCollector) (*healthconfig.AdvancedConfig, error) {
	// Load advanced configuration
	logger.Info("loading advanced configuration", core.String("config_path", config.ConfigPath))
	advConfig, err := he.loadAdvancedConfig(config.ConfigPath)
	if err != nil {
		metrics.IncrementCounter("config_load_errors")
		return nil, errors.NewFileError("load_config", config.ConfigPath, err)
	}
	metrics.IncrementCounter("config_load_success")

	return advConfig, nil
}

// loadAndPrepareRepositories loads repositories and converts them to core format
func (he *HealthExecutor) loadAndPrepareRepositories(config *HealthConfig, logger *observability.StructuredLogger, metrics *observability.MetricsCollector) ([]core.Repository, error) {
	// Load repositories from basic config
	logger.Info("loading repositories",
		core.String("basic_config", config.BasicConfig),
		core.String("tag", config.Tag))
	repositories, err := he.loadRepositories(config.BasicConfig, config.Tag)
	if err != nil {
		metrics.IncrementCounter("repo_load_errors")
		return nil, err
	}
	metrics.IncrementCounter("repo_load_success")
	metrics.SetGauge("repositories_found", float64(len(repositories)))

	if len(repositories) == 0 {
		logger.Warn("no repositories found", core.String("tag", config.Tag))
		return nil, nil
	}

	// Convert to core repositories
	coreRepos := he.convertRepositories(repositories)
	logger.Info("converted repositories", core.Int("count", len(coreRepos)))
	return coreRepos, nil
}

// executeAndReportResults executes health checks and reports the results
func (he *HealthExecutor) executeAndReportResults(ctx context.Context, coreRepos []core.Repository, advConfig *healthconfig.AdvancedConfig, config *HealthConfig, logger *observability.StructuredLogger, metrics *observability.MetricsCollector) error {
	// Execute health checks
	logger.Info("executing comprehensive health checks",
		core.Int("repositories", len(coreRepos)),
		core.String("pipeline", config.Pipeline))

	result, err := he.executeHealthChecks(ctx, coreRepos, advConfig, config, logger, metrics)
	if err != nil {
		metrics.IncrementCounter("health_check_execution_errors")
		return errors.NewContextualError("execute_health_checks", err).
			WithContext("repositories", len(coreRepos))
	}
	metrics.IncrementCounter("health_check_execution_success")

	// Record overall results in metrics
	for _, repoResult := range result.RepositoryResults {
		metrics.RecordRepositoryProcessed()
		for _, checkResult := range repoResult.CheckResults {
			metrics.RecordCheckResult(checkResult)
		}
	}

	// Display results
	logger.Info("displaying results", core.Bool("verbose", config.Verbose))
	formatter := health.NewFormatter(config.Verbose)
	formatter.DisplayResults(*result)

	// Return appropriate exit code
	exitCode := health.GetExitCode(*result)
	if exitCode != 0 {
		logger.Warn("health checks failed", core.Int("exit_code", exitCode))
		metrics.IncrementCounter("health_check_failures")
		return errors.NewContextualError("health_checks_failed",
			fmt.Errorf("health checks completed with exit code %d", exitCode)).
			WithContext("exit_code", exitCode)
	}

	logger.Info("health checks completed successfully")
	metrics.IncrementCounter("health_check_success")
	return nil
}

// executeHealthChecks executes the actual health checks
func (he *HealthExecutor) executeHealthChecks(ctx context.Context, repos []core.Repository,
	advConfig *healthconfig.AdvancedConfig, config *HealthConfig, logger *observability.StructuredLogger, metrics *observability.MetricsCollector) (*core.WorkflowResult, error) {
	opLogger := logger.WithField("operation", "execute_health_checks")

	// Create command executor and registries
	opLogger.Debug("creating command executor and registries")
	executor := health.NewCommandExecutor(config.Timeout)
	checkerRegistry := health.NewCheckerRegistry(executor)

	// Create filesystem and analyzer registry
	opLogger.Debug("creating filesystem and analyzer registry")
	fs := health.NewFileSystem()
	analyzerReg := health.NewAnalyzerRegistry(fs, he.logger)

	// Create orchestration engine
	opLogger.Debug("creating orchestration engine")
	engine := health.NewOrchestrationEngine(checkerRegistry, analyzerReg, advConfig, he.logger)

	// Execute health checks with timing
	opLogger.Info("starting health check execution", core.Int("repositories", len(repos)))
	timer := metrics.StartTimer("health_check_execution")
	defer timer.Stop()

	result, err := engine.ExecuteHealthCheck(ctx, repos)
	if err != nil {
		opLogger.Error("health check execution failed", core.String("error", err.Error()))
		return nil, err
	}

	opLogger.Info("health check execution completed",
		core.Int("repositories", len(result.RepositoryResults)),
		core.String("duration", timer.Duration().String()))

	return result, nil
}

// loadAdvancedConfig loads the advanced configuration file or returns default config
func (he *HealthExecutor) loadAdvancedConfig(configPath string) (*healthconfig.AdvancedConfig, error) {
	advConfig, err := healthconfig.LoadAdvancedConfigOrDefault(configPath)
	if err != nil {
		return nil, err
	}

	// Validate the configuration
	validator := healthconfig.NewConfigValidator()
	if err := validator.Validate(advConfig); err != nil {
		return nil, errors.NewContextualError("validate_advanced_config", err).
			WithContext("config_path", configPath)
	}

	return advConfig, nil
}

// loadRepositories loads repositories from the basic configuration
func (he *HealthExecutor) loadRepositories(configPath, tag string) ([]config.Repository, error) {
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, errors.NewFileError("load_repositories", configPath, err)
	}

	// Filter repositories by tag if specified
	if tag == "" {
		return cfg.Repositories, nil
	}

	var filteredRepos []config.Repository
	for _, repo := range cfg.Repositories {
		for _, repoTag := range repo.Tags {
			if repoTag == tag {
				filteredRepos = append(filteredRepos, repo)
				break
			}
		}
	}

	return filteredRepos, nil
}

// convertRepositories converts config repositories to core repositories
func (he *HealthExecutor) convertRepositories(repos []config.Repository) []core.Repository {
	coreRepos := make([]core.Repository, len(repos))
	for i, repo := range repos {
		coreRepos[i] = core.Repository{
			Name:   repo.Name,
			Path:   repo.Path,
			URL:    repo.URL,
			Branch: repo.Branch,
			Tags:   repo.Tags,
			// Language and Metadata can be detected/added later if needed
		}
	}
	return coreRepos
}

// simpleLogger provides a basic logger implementation for health executor
type simpleLogger struct{}

func (l *simpleLogger) Debug(msg string, fields ...core.Field) {
	// Simple debug implementation - could be enhanced
	fmt.Printf("[DEBUG] %s\n", msg)
}

func (l *simpleLogger) Info(msg string, fields ...core.Field) {
	fmt.Printf("[INFO] %s\n", msg)
}

func (l *simpleLogger) Warn(msg string, fields ...core.Field) {
	fmt.Printf("[WARN] %s\n", msg)
}

func (l *simpleLogger) Error(msg string, fields ...core.Field) {
	fmt.Printf("[ERROR] %s\n", msg)
}

func (l *simpleLogger) Fatal(msg string, fields ...core.Field) {
	fmt.Printf("[FATAL] %s\n", msg)
	os.Exit(1)
}
