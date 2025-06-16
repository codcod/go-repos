package health

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// CheckerConfig represents the configuration for health checkers
type CheckerConfig struct {
	CyclomaticComplexity ComplexityConfig      `yaml:"cyclomatic_complexity"`
	DeprecatedComponents DeprecatedConfig      `yaml:"deprecated_components"`
	General              GeneralConfig         `yaml:"general"`
	Languages            map[string]LangConfig `yaml:"languages"`
}

// ComplexityConfig represents configuration for cyclomatic complexity checking
type ComplexityConfig struct {
	DefaultThreshold int            `yaml:"default_threshold"`
	DetailedReport   bool           `yaml:"detailed_report"`
	LanguageSpecific map[string]int `yaml:"language_specific"`
	Exclusions       []string       `yaml:"exclusions"`
}

// DeprecatedConfig represents configuration for deprecated component checking
type DeprecatedConfig struct {
	SeverityLevels map[string][]string `yaml:"severity_levels"`
	CustomPatterns []PatternConfig     `yaml:"custom_patterns"`
}

// GeneralConfig represents general configuration options
type GeneralConfig struct {
	Timeout        time.Duration `yaml:"timeout"`
	MaxConcurrency int           `yaml:"max_concurrency"`
	CacheResults   bool          `yaml:"cache_results"`
	CacheTTL       time.Duration `yaml:"cache_ttl"`
}

// LangConfig represents language-specific configuration
type LangConfig struct {
	Patterns            []string `yaml:"patterns"`
	Exclusions          []string `yaml:"exclusions"`
	ComplexityThreshold int      `yaml:"complexity_threshold"`
	EnableFunctionLevel bool     `yaml:"enable_function_level"`
}

// PatternConfig represents a custom pattern configuration
type PatternConfig struct {
	Name       string   `yaml:"name"`
	Patterns   []string `yaml:"patterns"`
	Severity   string   `yaml:"severity"`
	Message    string   `yaml:"message"`
	Suggestion string   `yaml:"suggestion"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *CheckerConfig {
	return &CheckerConfig{
		CyclomaticComplexity: ComplexityConfig{
			DefaultThreshold: 10,
			DetailedReport:   false,
			LanguageSpecific: map[string]int{
				"python":     8,
				"java":       12,
				"javascript": 10,
				"typescript": 10,
				"go":         10,
			},
			Exclusions: []string{
				"*_test.go",
				"test_*.py",
				"*.spec.js",
				"*.test.js",
			},
		},
		DeprecatedComponents: DeprecatedConfig{
			SeverityLevels: map[string][]string{
				"high":   {"security", "breaking_change"},
				"medium": {"performance", "deprecated_api"},
				"low":    {"style", "legacy"},
			},
		},
		General: GeneralConfig{
			Timeout:        30 * time.Second,
			MaxConcurrency: 4,
			CacheResults:   true,
			CacheTTL:       5 * time.Minute,
		},
		Languages: map[string]LangConfig{
			"python": {
				Patterns:            []string{"*.py"},
				Exclusions:          []string{".venv/", "__pycache__/", ".pytest_cache/", "venv/", "env/"},
				ComplexityThreshold: 8,
				EnableFunctionLevel: true,
			},
			"java": {
				Patterns:            []string{"*.java"},
				Exclusions:          []string{"target/", "build/", ".gradle/"},
				ComplexityThreshold: 12,
				EnableFunctionLevel: true,
			},
			"javascript": {
				Patterns:            []string{"*.js", "*.jsx"},
				Exclusions:          []string{"node_modules/", "dist/", "build/", ".next/"},
				ComplexityThreshold: 10,
				EnableFunctionLevel: true,
			},
			"typescript": {
				Patterns:            []string{"*.ts", "*.tsx"},
				Exclusions:          []string{"node_modules/", "dist/", "build/", ".next/"},
				ComplexityThreshold: 10,
				EnableFunctionLevel: true,
			},
			"go": {
				Patterns:            []string{"*.go"},
				Exclusions:          []string{"vendor/", "*_test.go"},
				ComplexityThreshold: 10,
				EnableFunctionLevel: true,
			},
		},
	}
}

// LoadConfig loads configuration from a file or returns default if file doesn't exist
func LoadConfig(configPath string) (*CheckerConfig, error) {
	// If no config path provided or file doesn't exist, use defaults
	if configPath == "" || !fileExists(configPath) {
		return DefaultConfig(), nil
	}

	data, err := safeReadConfigFile(configPath)
	if err != nil {
		return nil, NewCheckerError("config", "load", err, ErrorCodeFileNotFound)
	}

	var config CheckerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, NewCheckerError("config", "parse", err, ErrorCodeInvalidInput)
	}

	// Merge with defaults to ensure all fields are populated
	defaultConfig := DefaultConfig()
	mergeConfigs(&config, defaultConfig)

	return &config, nil
}

// SaveConfig saves the configuration to a file
func (c *CheckerConfig) SaveConfig(configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return NewCheckerError("config", "create_dir", err, ErrorCodePermission)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return NewCheckerError("config", "marshal", err, ErrorCodeInvalidInput)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return NewCheckerError("config", "write", err, ErrorCodePermission)
	}

	return nil
}

// GetComplexityThreshold returns the complexity threshold for a language
func (c *CheckerConfig) GetComplexityThreshold(language string) int {
	if threshold, exists := c.CyclomaticComplexity.LanguageSpecific[language]; exists {
		return threshold
	}
	return c.CyclomaticComplexity.DefaultThreshold
}

// GetLanguageConfig returns the configuration for a specific language
func (c *CheckerConfig) GetLanguageConfig(language string) LangConfig {
	if config, exists := c.Languages[language]; exists {
		return config
	}
	// Return empty config if language not found
	return LangConfig{}
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// mergeConfigs merges two configurations, using non-zero values from source
func mergeConfigs(dest, src *CheckerConfig) {
	if dest.CyclomaticComplexity.DefaultThreshold == 0 {
		dest.CyclomaticComplexity.DefaultThreshold = src.CyclomaticComplexity.DefaultThreshold
	}
	if dest.CyclomaticComplexity.LanguageSpecific == nil {
		dest.CyclomaticComplexity.LanguageSpecific = src.CyclomaticComplexity.LanguageSpecific
	}
	if dest.General.Timeout == 0 {
		dest.General.Timeout = src.General.Timeout
	}
	if dest.General.MaxConcurrency == 0 {
		dest.General.MaxConcurrency = src.General.MaxConcurrency
	}
	if dest.Languages == nil {
		dest.Languages = src.Languages
	}
}

// safeReadConfigFile reads a config file with path validation
func safeReadConfigFile(configPath string) ([]byte, error) {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(configPath)

	// Ensure the path doesn't contain directory traversal patterns
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid config path: contains directory traversal")
	}

	// Ensure it's a YAML file
	ext := strings.ToLower(filepath.Ext(cleanPath))
	if ext != ".yaml" && ext != ".yml" {
		return nil, fmt.Errorf("invalid config file extension: %s (expected .yaml or .yml)", ext)
	}

	return os.ReadFile(cleanPath)
}
