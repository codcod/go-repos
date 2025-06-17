package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/codcod/repos/internal/core"
)

// ModularConfig implements core.Config interface
type ModularConfig struct {
	Checkers  map[string]core.CheckerConfig  `yaml:"checkers"`
	Analyzers map[string]core.AnalyzerConfig `yaml:"analyzers"`
	Reporters map[string]core.ReporterConfig `yaml:"reporters"`
	Engine    core.EngineConfig              `yaml:"engine"`
}

// GetCheckerConfig returns configuration for a checker
func (c *ModularConfig) GetCheckerConfig(checkerID string) (core.CheckerConfig, bool) {
	config, exists := c.Checkers[checkerID]
	return config, exists
}

// GetAnalyzerConfig returns configuration for an analyzer
func (c *ModularConfig) GetAnalyzerConfig(language string) (core.AnalyzerConfig, bool) {
	config, exists := c.Analyzers[language]
	return config, exists
}

// GetReporterConfig returns configuration for a reporter
func (c *ModularConfig) GetReporterConfig(reporterID string) (core.ReporterConfig, bool) {
	config, exists := c.Reporters[reporterID]
	return config, exists
}

// GetEngineConfig returns the engine configuration
func (c *ModularConfig) GetEngineConfig() core.EngineConfig {
	return c.Engine
}

// LoadModularConfig loads configuration from a file
func LoadModularConfig(configPath string) (*ModularConfig, error) {
	// If no config path provided or file doesn't exist, use defaults
	if configPath == "" || !fileExists(configPath) {
		return DefaultModularConfig(), nil
	}

	data, err := safeReadConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ModularConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Merge with defaults to ensure all fields are populated
	defaultConfig := DefaultModularConfig()
	mergeModularConfigs(&config, defaultConfig)

	return &config, nil
}

// DefaultModularConfig returns a default configuration
func DefaultModularConfig() *ModularConfig {
	return &ModularConfig{
		Checkers: map[string]core.CheckerConfig{
			"cyclomatic-complexity": {
				Enabled:  true,
				Severity: "medium",
				Timeout:  30 * time.Second,
				Options: map[string]interface{}{
					"default_threshold": 10,
					"detailed_report":   false,
					"thresholds": map[string]int{
						"python":     8,
						"java":       12,
						"javascript": 10,
						"typescript": 10,
						"go":         10,
					},
				},
				Categories: []string{"quality"},
				Exclusions: []string{
					"*_test.go",
					"test_*.py",
					"*.spec.js",
					"*.test.js",
				},
			},
		},
		Analyzers: map[string]core.AnalyzerConfig{
			"python": {
				Enabled:           true,
				FileExtensions:    []string{".py"},
				ExcludePatterns:   []string{".venv/", "__pycache__/", ".pytest_cache/", "venv/", "env/"},
				ComplexityEnabled: true,
				FunctionLevel:     true,
			},
			"go": {
				Enabled:           true,
				FileExtensions:    []string{".go"},
				ExcludePatterns:   []string{"vendor/", "*_test.go"},
				ComplexityEnabled: true,
				FunctionLevel:     true,
			},
		},
		Reporters: map[string]core.ReporterConfig{
			"console": {
				Enabled:    true,
				OutputFile: "",
				Template:   "table",
				Options: map[string]interface{}{
					"show_summary": true,
					"show_details": false,
				},
			},
		},
		Engine: core.EngineConfig{
			MaxConcurrency: 4,
			Timeout:        5 * time.Minute,
			CacheEnabled:   true,
			CacheTTL:       5 * time.Minute,
		},
	}
}

// safeReadConfigFile reads a config file with path validation
func safeReadConfigFile(configPath string) ([]byte, error) {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(configPath)

	// Basic security check
	if filepath.IsAbs(cleanPath) {
		return nil, &ConfigError{Code: "absolute_path", Message: "absolute paths not allowed"}
	}

	// Ensure it's a YAML file
	ext := filepath.Ext(cleanPath)
	if ext != ".yaml" && ext != ".yml" {
		return nil, &ConfigError{Code: "invalid_extension", Message: "configuration file must have .yaml or .yml extension"}
	}

	return os.ReadFile(cleanPath)
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// mergeModularConfigs merges two configurations
func mergeModularConfigs(dest, src *ModularConfig) {
	// Merge checkers
	for id, srcChecker := range src.Checkers {
		if _, exists := dest.Checkers[id]; !exists {
			dest.Checkers[id] = srcChecker
		}
	}

	// Merge analyzers
	for lang, srcAnalyzer := range src.Analyzers {
		if _, exists := dest.Analyzers[lang]; !exists {
			dest.Analyzers[lang] = srcAnalyzer
		}
	}

	// Merge reporters
	for id, srcReporter := range src.Reporters {
		if _, exists := dest.Reporters[id]; !exists {
			dest.Reporters[id] = srcReporter
		}
	}

	// Merge engine config
	if dest.Engine.MaxConcurrency == 0 {
		dest.Engine.MaxConcurrency = src.Engine.MaxConcurrency
	}
	if dest.Engine.Timeout == 0 {
		dest.Engine.Timeout = src.Engine.Timeout
	}
	if dest.Engine.CacheTTL == 0 {
		dest.Engine.CacheTTL = src.Engine.CacheTTL
	}
}

// ConfigError represents a configuration error
type ConfigError struct {
	Code    string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
