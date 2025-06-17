package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/codcod/repos/internal/core"
)

// LegacyConfig represents the old configuration format
type LegacyConfig struct {
	Repositories []LegacyRepository `yaml:"repositories"`
	Checkers     LegacyCheckers     `yaml:"checkers,omitempty"`
	Complexity   LegacyComplexity   `yaml:"complexity,omitempty"`
	Reporting    LegacyReporting    `yaml:"reporting,omitempty"`
}

// LegacyRepository represents old repository format
type LegacyRepository struct {
	Name        string   `yaml:"name"`
	URL         string   `yaml:"url"`
	Branch      string   `yaml:"branch,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	Description string   `yaml:"description,omitempty"`
}

// LegacyCheckers represents old checker configuration
type LegacyCheckers struct {
	Enabled []string      `yaml:"enabled,omitempty"`
	Timeout time.Duration `yaml:"timeout,omitempty"`
	Git     LegacyGit     `yaml:"git,omitempty"`
}

// LegacyGit represents old git checker configuration
type LegacyGit struct {
	CheckUncommitted bool `yaml:"check_uncommitted,omitempty"`
	CheckUnpushed    bool `yaml:"check_unpushed,omitempty"`
}

// LegacyComplexity represents old complexity configuration
type LegacyComplexity struct {
	Threshold int            `yaml:"threshold,omitempty"`
	Languages map[string]int `yaml:"languages,omitempty"`
	Enabled   bool           `yaml:"enabled,omitempty"`
}

// LegacyReporting represents old reporting configuration
type LegacyReporting struct {
	Format     string `yaml:"format,omitempty"`
	OutputFile string `yaml:"output_file,omitempty"`
	Summary    bool   `yaml:"summary,omitempty"`
}

// ConfigMigrator handles migration from legacy to advanced configuration
type ConfigMigrator struct {
	logger core.Logger
}

// NewConfigMigrator creates a new configuration migrator
func NewConfigMigrator(logger core.Logger) *ConfigMigrator {
	return &ConfigMigrator{
		logger: logger,
	}
}

// MigrateConfig converts legacy configuration to advanced configuration
func (m *ConfigMigrator) MigrateConfig(legacyPath, advancedPath string) error {
	m.logger.Info("Starting configuration migration",
		core.String("legacy_path", legacyPath),
		core.String("advanced_path", advancedPath))

	// Check if legacy config exists
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		return fmt.Errorf("legacy configuration file not found: %s", legacyPath)
	}

	// Load legacy configuration
	legacyConfig, err := m.loadLegacyConfig(legacyPath)
	if err != nil {
		return fmt.Errorf("failed to load legacy config: %w", err)
	}

	// Convert to advanced configuration
	advancedConfig := m.convertToAdvanced(legacyConfig)

	// Save advanced configuration
	err = m.saveAdvancedConfig(advancedConfig, advancedPath)
	if err != nil {
		return fmt.Errorf("failed to save advanced config: %w", err)
	}

	m.logger.Info("Configuration migration completed successfully",
		core.String("advanced_path", advancedPath))

	return nil
}

// DetectConfigFormat determines if a config file is legacy or advanced format
//
//nolint:gocyclo // Complex config format detection requires high cyclomatic complexity
func (m *ConfigMigrator) DetectConfigFormat(configPath string) (string, error) {
	data, err := os.ReadFile(configPath) //nolint:gosec // Config path is from user input
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	// Try parsing as advanced config first
	var advancedTest AdvancedConfig
	if err := yaml.Unmarshal(data, &advancedTest); err == nil {
		// Check for advanced config markers - any of these indicate advanced format
		if advancedTest.Version != "" ||
			len(advancedTest.Profiles) > 0 ||
			len(advancedTest.Pipelines) > 0 ||
			len(advancedTest.Checkers) > 0 ||
			len(advancedTest.Analyzers) > 0 ||
			len(advancedTest.Categories) > 0 {
			return "advanced", nil
		}
	}

	// Try parsing as legacy config
	var legacyTest LegacyConfig
	if err := yaml.Unmarshal(data, &legacyTest); err == nil {
		// Check for legacy config markers
		if len(legacyTest.Repositories) > 0 {
			return "legacy", nil
		}
	}

	return "unknown", fmt.Errorf("unable to determine configuration format")
}

// loadLegacyConfig loads configuration from legacy format
func (m *ConfigMigrator) loadLegacyConfig(path string) (*LegacyConfig, error) {
	data, err := os.ReadFile(path) //nolint:gosec // Config path is from user input
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config LegacyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// convertToAdvanced converts legacy configuration to advanced format
func (m *ConfigMigrator) convertToAdvanced(legacy *LegacyConfig) *AdvancedConfig {
	advanced := &AdvancedConfig{
		Version:      "1.0",
		Engine:       m.convertEngineConfig(legacy),
		Checkers:     m.convertCheckers(legacy),
		Analyzers:    m.convertAnalyzers(legacy),
		Reporters:    m.convertReporters(legacy),
		Categories:   m.convertCategories(legacy),
		Profiles:     m.createDefaultProfiles(),
		Pipelines:    m.createDefaultPipelines(),
		Overrides:    []OverrideConfig{},
		Extensions:   ExtensionsConfig{},
		Integrations: IntegrationsConfig{},
	}

	return advanced
}

// convertEngineConfig converts engine configuration
func (m *ConfigMigrator) convertEngineConfig(legacy *LegacyConfig) core.EngineConfig {
	timeout := 5 * time.Minute
	if legacy.Checkers.Timeout > 0 {
		timeout = legacy.Checkers.Timeout
	}

	return core.EngineConfig{
		MaxConcurrency: 4, // Default reasonable value
		Timeout:        timeout,
		CacheEnabled:   true,
		CacheTTL:       time.Hour,
	}
}

// convertCheckers converts checker configuration
//
//nolint:gocyclo // Complex checker conversion logic requires high cyclomatic complexity
func (m *ConfigMigrator) convertCheckers(legacy *LegacyConfig) map[string]core.CheckerConfig {
	checkers := make(map[string]core.CheckerConfig)

	// Convert enabled checkers
	enabledSet := make(map[string]bool)
	for _, name := range legacy.Checkers.Enabled {
		enabledSet[name] = true
	}

	// Default timeout
	timeout := 30 * time.Second
	if legacy.Checkers.Timeout > 0 {
		timeout = legacy.Checkers.Timeout
	}

	// Git checker
	if enabledSet["git"] || len(legacy.Checkers.Enabled) == 0 {
		checkers["git-status"] = core.CheckerConfig{
			Enabled: true,
			Timeout: timeout,
			Options: map[string]interface{}{
				"check_uncommitted": legacy.Checkers.Git.CheckUncommitted,
				"check_unpushed":    legacy.Checkers.Git.CheckUnpushed,
			},
		}
	}

	// Security checkers
	if enabledSet["security"] || len(legacy.Checkers.Enabled) == 0 {
		checkers["branch-protection"] = core.CheckerConfig{
			Enabled: true,
			Timeout: timeout,
			Options: map[string]interface{}{
				"require_reviews": true,
			},
		}

		checkers["vulnerability-scan"] = core.CheckerConfig{
			Enabled: true,
			Timeout: 5 * time.Minute,
			Options: map[string]interface{}{
				"severity_threshold": "medium",
			},
		}
	}

	// Dependencies checker
	if enabledSet["dependencies"] || len(legacy.Checkers.Enabled) == 0 {
		checkers["outdated-dependencies"] = core.CheckerConfig{
			Enabled: true,
			Timeout: timeout,
			Options: map[string]interface{}{
				"check_security": true,
			},
		}
	}

	// Complexity checker
	if legacy.Complexity.Enabled || len(legacy.Checkers.Enabled) == 0 {
		options := map[string]interface{}{
			"default_threshold": legacy.Complexity.Threshold,
		}
		if legacy.Complexity.Threshold == 0 {
			options["default_threshold"] = 10
		}
		if len(legacy.Complexity.Languages) > 0 {
			options["language_thresholds"] = legacy.Complexity.Languages
		}

		checkers["cyclomatic-complexity"] = core.CheckerConfig{
			Enabled: true,
			Timeout: timeout,
			Options: options,
		}
	}

	return checkers
}

// convertAnalyzers converts analyzer configuration
func (m *ConfigMigrator) convertAnalyzers(legacy *LegacyConfig) map[string]core.AnalyzerConfig {
	analyzers := make(map[string]core.AnalyzerConfig)

	// Enable analyzers for languages that had complexity thresholds

	if len(legacy.Complexity.Languages) > 0 {
		for lang := range legacy.Complexity.Languages {
			analyzers[lang] = core.AnalyzerConfig{
				Enabled:           true,
				FileExtensions:    getExtensionsForLanguage(lang),
				ExcludePatterns:   []string{},
				ComplexityEnabled: true,
				FunctionLevel:     false,
			}
		}
	} else {
		// Default analyzers
		commonConfig := core.AnalyzerConfig{
			Enabled:           true,
			FileExtensions:    []string{},
			ExcludePatterns:   []string{},
			ComplexityEnabled: legacy.Complexity.Enabled,
			FunctionLevel:     false,
		}

		analyzers["go"] = commonConfig
		analyzers["python"] = commonConfig
		analyzers["javascript"] = commonConfig
		analyzers["java"] = commonConfig
	}

	return analyzers
}

// convertReporters converts reporter configuration
func (m *ConfigMigrator) convertReporters(legacy *LegacyConfig) map[string]core.ReporterConfig {
	reporters := make(map[string]core.ReporterConfig)

	format := "table"
	if legacy.Reporting.Format != "" {
		format = legacy.Reporting.Format
	}

	// Console reporter
	reporters["console"] = core.ReporterConfig{
		Enabled: true,
		Options: map[string]interface{}{
			"format":       format,
			"show_summary": legacy.Reporting.Summary,
		},
	}

	// File reporter if output file specified
	if legacy.Reporting.OutputFile != "" {
		reporters["file"] = core.ReporterConfig{
			Enabled: true,
			Options: map[string]interface{}{
				"output_file":  legacy.Reporting.OutputFile,
				"format":       format,
				"show_summary": legacy.Reporting.Summary,
			},
		}
	}

	return reporters
}

// convertCategories creates category configuration
func (m *ConfigMigrator) convertCategories(_ *LegacyConfig) map[string]CategoryConfig {
	categories := make(map[string]CategoryConfig)

	categories["git"] = CategoryConfig{
		Name:        "Git Management",
		Description: "Git repository management checks",
		Enabled:     true,
		Severity:    "low",
		Weight:      1.0,
		Checkers:    []string{"git-status"},
	}

	categories["security"] = CategoryConfig{
		Name:        "Security",
		Description: "Security-related checks",
		Enabled:     true,
		Severity:    "high",
		Weight:      1.5,
		Checkers:    []string{"branch-protection", "vulnerability-scan"},
	}

	categories["quality"] = CategoryConfig{
		Name:        "Code Quality",
		Description: "Code quality and maintainability checks",
		Enabled:     true,
		Severity:    "medium",
		Weight:      1.0,
		Checkers:    []string{"cyclomatic-complexity", "outdated-dependencies"},
	}

	return categories
}

// createDefaultProfiles creates default configuration profiles
func (m *ConfigMigrator) createDefaultProfiles() map[string]ProfileConfig {
	profiles := make(map[string]ProfileConfig)

	profiles["default"] = ProfileConfig{
		Name:        "Default Profile",
		Description: "Standard health check profile (migrated from legacy config)",
		Categories:  []string{"git", "security", "quality"},
		Exclusions:  []string{},
	}

	profiles["quick"] = ProfileConfig{
		Name:        "Quick Check",
		Description: "Fast health check for development",
		Categories:  []string{"git"},
		Exclusions:  []string{"vulnerability-scan"},
	}

	return profiles
}

// createDefaultPipelines creates default pipeline configuration
func (m *ConfigMigrator) createDefaultPipelines() map[string]PipelineConfig {
	pipelines := make(map[string]PipelineConfig)

	pipelines["default"] = PipelineConfig{
		Name:        "Default Pipeline",
		Description: "Standard health check pipeline (migrated from legacy config)",
		Steps: []StepConfig{
			{
				Name:    "analysis",
				Type:    "analysis",
				Enabled: true,
				Timeout: "5m",
				Config: map[string]interface{}{
					"include_complexity": true,
				},
			},
			{
				Name:         "checks",
				Type:         "checkers",
				Enabled:      true,
				Timeout:      "10m",
				Dependencies: []string{"analysis"},
				Config: map[string]interface{}{
					"categories": []string{"git", "security", "quality"},
				},
			},
			{
				Name:         "reporting",
				Type:         "reporting",
				Enabled:      true,
				Timeout:      "1m",
				Dependencies: []string{"checks"},
				Config: map[string]interface{}{
					"formats": []string{"console"},
				},
			},
		},
		Config: ExecutionConfig{
			MaxConcurrency:  3,
			Timeout:         "30m",
			FailFast:        false,
			RetryCount:      1,
			RetryDelay:      "10s",
			ContinueOnError: true,
			OutputFormats:   []string{"console"},
		},
	}

	return pipelines
}

// saveAdvancedConfig saves the advanced configuration to file
func (m *ConfigMigrator) saveAdvancedConfig(config *AdvancedConfig, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	err = os.WriteFile(path, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// getExtensionsForLanguage returns file extensions for a given language
func getExtensionsForLanguage(language string) []string {
	extensions := map[string][]string{
		"go":         {".go"},
		"python":     {".py"},
		"javascript": {".js", ".jsx", ".ts", ".tsx"},
		"java":       {".java"},
		"rust":       {".rs"},
		"cpp":        {".cpp", ".cc", ".cxx", ".c++"},
		"c":          {".c", ".h"},
		"csharp":     {".cs"},
		"php":        {".php"},
		"ruby":       {".rb"},
		"kotlin":     {".kt", ".kts"},
		"swift":      {".swift"},
		"scala":      {".scala", ".sc"},
	}

	if exts, exists := extensions[language]; exists {
		return exts
	}
	return []string{} // Return empty slice for unknown languages
}
