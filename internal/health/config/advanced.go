package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/codcod/repos/internal/core"
)

// AdvancedConfig implements the Config interface with advanced features
type AdvancedConfig struct {
	Version    string                         `yaml:"version"`
	Engine     core.EngineConfig              `yaml:"engine"`
	Checkers   map[string]core.CheckerConfig  `yaml:"checkers"`
	Analyzers  map[string]core.AnalyzerConfig `yaml:"analyzers"`
	Reporters  map[string]core.ReporterConfig `yaml:"reporters"`
	Categories map[string]CategoryConfig      `yaml:"categories"`
	Overrides  []OverrideConfig               `yaml:"overrides"`
	// Future use - extension points not yet implemented
	// Extensions   ExtensionsConfig               `yaml:"extensions"`
	// Integrations IntegrationsConfig             `yaml:"integrations"`
}

// CategoryConfig defines configuration for a category of checks
type CategoryConfig struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Enabled     bool                   `yaml:"enabled"`
	Severity    string                 `yaml:"severity"`
	Weight      float64                `yaml:"weight"`
	Checkers    []string               `yaml:"checkers"`
	Options     map[string]interface{} `yaml:"options"`
}

// OverrideConfig defines conditional configuration overrides
type OverrideConfig struct {
	Name       string                         `yaml:"name"`
	Conditions []ConditionConfig              `yaml:"conditions"`
	Checkers   map[string]core.CheckerConfig  `yaml:"checkers"`
	Analyzers  map[string]core.AnalyzerConfig `yaml:"analyzers"`
	Engine     *core.EngineConfig             `yaml:"engine,omitempty"`
}

// ConditionConfig defines conditions for applying overrides
type ConditionConfig struct {
	Type     string        `yaml:"type"` // "repository", "language", "tag", "path"
	Field    string        `yaml:"field"`
	Operator string        `yaml:"operator"` // "equals", "contains", "matches", "in"
	Value    interface{}   `yaml:"value"`
	Values   []interface{} `yaml:"values,omitempty"`
}

// ExtensionsConfig configures extension points
type ExtensionsConfig struct {
	CustomCheckers []CustomCheckerConfig `yaml:"custom_checkers"`
	Hooks          []HookConfig          `yaml:"hooks"`
	Plugins        []PluginConfig        `yaml:"plugins"`
}

// CustomCheckerConfig defines a custom checker
type CustomCheckerConfig struct {
	ID       string                 `yaml:"id"`
	Name     string                 `yaml:"name"`
	Category string                 `yaml:"category"`
	Command  string                 `yaml:"command"`
	Args     []string               `yaml:"args"`
	Config   map[string]interface{} `yaml:"config"`
}

// HookConfig defines execution hooks
type HookConfig struct {
	Name    string   `yaml:"name"`
	Event   string   `yaml:"event"` // "pre_check", "post_check", "on_error"
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
}

// PluginConfig defines plugin configuration
type PluginConfig struct {
	Name    string                 `yaml:"name"`
	Path    string                 `yaml:"path"`
	Config  map[string]interface{} `yaml:"config"`
	Enabled bool                   `yaml:"enabled"`
}

// IntegrationsConfig configures external integrations
type IntegrationsConfig struct {
	GitHub GitHubConfig `yaml:"github"`
	Slack  SlackConfig  `yaml:"slack"`
	JIRA   JIRAConfig   `yaml:"jira"`
}

// GitHubConfig configures GitHub integration
type GitHubConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Token        string `yaml:"token"`
	BaseURL      string `yaml:"base_url"`
	CreateIssues bool   `yaml:"create_issues"`
	UpdatePRs    bool   `yaml:"update_prs"`
}

// SlackConfig configures Slack integration
type SlackConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
	Username   string `yaml:"username"`
}

// JIRAConfig configures JIRA integration
type JIRAConfig struct {
	Enabled  bool   `yaml:"enabled"`
	BaseURL  string `yaml:"base_url"`
	Username string `yaml:"username"`
	APIToken string `yaml:"api_token"`
	Project  string `yaml:"project"`
}

// LoadAdvancedConfig loads configuration from a YAML file with advanced features
func LoadAdvancedConfig(configPath string) (*AdvancedConfig, error) {
	data, err := os.ReadFile(configPath) //nolint:gosec // Config path is from user input
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AdvancedConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	config.setDefaults()

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// NewDefaultAdvancedConfig creates a default advanced configuration with sane defaults
func NewDefaultAdvancedConfig() *AdvancedConfig {
	config := &AdvancedConfig{
		Version: "1.0",
		Engine: core.EngineConfig{
			MaxConcurrency: 4,
			Timeout:        30 * time.Minute,
			CacheTTL:       1 * time.Hour,
		},
		Checkers:  make(map[string]core.CheckerConfig),
		Analyzers: make(map[string]core.AnalyzerConfig),
		Reporters: make(map[string]core.ReporterConfig),
		Categories: map[string]CategoryConfig{
			"security": {
				Name:        "Security",
				Description: "Security-related checks",
				Weight:      30,
				Enabled:     true,
			},
			"quality": {
				Name:        "Code Quality",
				Description: "Code quality and maintainability checks",
				Weight:      25,
				Enabled:     true,
			},
			"compliance": {
				Name:        "Compliance",
				Description: "Compliance and licensing checks",
				Weight:      20,
				Enabled:     true,
			},
			"ci": {
				Name:        "CI/CD",
				Description: "Continuous integration and deployment checks",
				Weight:      15,
				Enabled:     true,
			},
			"docs": {
				Name:        "Documentation",
				Description: "Documentation completeness checks",
				Weight:      10,
				Enabled:     true,
			},
		},
		Overrides: []OverrideConfig{},
		// Extensions and Integrations will be added when implemented
	}

	// Set defaults to ensure consistency
	config.setDefaults()

	return config
}

// LoadAdvancedConfigOrDefault loads configuration from a file, or returns default config if file doesn't exist
func LoadAdvancedConfigOrDefault(configPath string) (*AdvancedConfig, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default configuration if file doesn't exist
		return NewDefaultAdvancedConfig(), nil
	}

	// Load from file if it exists
	return LoadAdvancedConfig(configPath)
}

// setDefaults sets default values for configuration
func (c *AdvancedConfig) setDefaults() {
	if c.Version == "" {
		c.Version = "1.0"
	}

	// Set engine defaults
	if c.Engine.MaxConcurrency == 0 {
		c.Engine.MaxConcurrency = 4
	}
	if c.Engine.Timeout == 0 {
		c.Engine.Timeout = 30 * time.Minute
	}
	if c.Engine.CacheTTL == 0 {
		c.Engine.CacheTTL = 1 * time.Hour
	}

	// Initialize maps if nil
	if c.Checkers == nil {
		c.Checkers = make(map[string]core.CheckerConfig)
	}
	if c.Analyzers == nil {
		c.Analyzers = make(map[string]core.AnalyzerConfig)
	}
	if c.Reporters == nil {
		c.Reporters = make(map[string]core.ReporterConfig)
	}
	if c.Categories == nil {
		c.Categories = make(map[string]CategoryConfig)
	}
}

// validate validates the configuration
func (c *AdvancedConfig) validate() error {
	// Validate override conditions
	for _, override := range c.Overrides {
		if err := c.validateOverrideConditions(override); err != nil {
			return fmt.Errorf("invalid override '%s': %w", override.Name, err)
		}
	}

	return nil
}

// validateOverrideConditions validates override conditions
func (c *AdvancedConfig) validateOverrideConditions(override OverrideConfig) error {
	validTypes := map[string]bool{
		"repository": true,
		"language":   true,
		"tag":        true,
		"path":       true,
	}

	validOperators := map[string]bool{
		"equals":   true,
		"contains": true,
		"matches":  true,
		"in":       true,
	}

	for _, condition := range override.Conditions {
		if !validTypes[condition.Type] {
			return fmt.Errorf("invalid condition type '%s'", condition.Type)
		}

		if !validOperators[condition.Operator] {
			return fmt.Errorf("invalid condition operator '%s'", condition.Operator)
		}

		if condition.Operator == "in" && len(condition.Values) == 0 {
			return fmt.Errorf("'in' operator requires 'values' field")
		}
	}

	return nil
}

// GetCheckerConfig implements core.Config
func (c *AdvancedConfig) GetCheckerConfig(checkerID string) (core.CheckerConfig, bool) {
	config, exists := c.Checkers[checkerID]
	return config, exists
}

// GetAnalyzerConfig implements core.Config
func (c *AdvancedConfig) GetAnalyzerConfig(language string) (core.AnalyzerConfig, bool) {
	config, exists := c.Analyzers[language]
	return config, exists
}

// GetReporterConfig implements core.Config
func (c *AdvancedConfig) GetReporterConfig(reporterID string) (core.ReporterConfig, bool) {
	config, exists := c.Reporters[reporterID]
	return config, exists
}

// GetEngineConfig implements core.Config
func (c *AdvancedConfig) GetEngineConfig() core.EngineConfig {
	return c.Engine
}

// ApplyOverrides applies configuration overrides based on repository context
func (c *AdvancedConfig) ApplyOverrides(repo core.Repository) error {
	for _, override := range c.Overrides {
		if c.matchesConditions(override.Conditions, repo) {
			// Apply override configurations
			for checkerID, checkerConfig := range override.Checkers {
				c.Checkers[checkerID] = checkerConfig
			}

			for language, analyzerConfig := range override.Analyzers {
				c.Analyzers[language] = analyzerConfig
			}

			if override.Engine != nil {
				c.Engine = *override.Engine
			}
		}
	}

	return nil
}

// matchesConditions checks if repository matches override conditions
func (c *AdvancedConfig) matchesConditions(conditions []ConditionConfig, repo core.Repository) bool {
	for _, condition := range conditions {
		if !c.matchesCondition(condition, repo) {
			return false
		}
	}
	return true
}

// matchesCondition checks if repository matches a single condition
func (c *AdvancedConfig) matchesCondition(condition ConditionConfig, repo core.Repository) bool {
	var fieldValue string

	switch condition.Type {
	case "repository":
		switch condition.Field {
		case "name":
			fieldValue = repo.Name
		case "path":
			fieldValue = repo.Path
		case "url":
			fieldValue = repo.URL
		case "branch":
			fieldValue = repo.Branch
		}
	case "language":
		fieldValue = repo.Language
	case "tag":
		// Handle tags separately since it's a slice
		return c.matchesTagCondition(condition, repo.Tags)
	case "path":
		fieldValue = repo.Path
	}

	return c.evaluateCondition(condition, fieldValue)
}

// matchesTagCondition checks if tags match the condition
//
//nolint:gocyclo // Complex condition matching logic requires high cyclomatic complexity
func (c *AdvancedConfig) matchesTagCondition(condition ConditionConfig, tags []string) bool {
	switch condition.Operator {
	case "equals":
		targetTag := fmt.Sprintf("%v", condition.Value)
		for _, tag := range tags {
			if tag == targetTag {
				return true
			}
		}
		return false
	case "contains":
		targetTag := fmt.Sprintf("%v", condition.Value)
		for _, tag := range tags {
			if len(tag) >= len(targetTag) &&
				tag[:len(targetTag)] == targetTag {
				return true
			}
		}
		return false
	case "in":
		for _, tag := range tags {
			for _, value := range condition.Values {
				if tag == fmt.Sprintf("%v", value) {
					return true
				}
			}
		}
		return false
	}
	return false
}

// evaluateCondition evaluates a condition against a field value
func (c *AdvancedConfig) evaluateCondition(condition ConditionConfig, fieldValue string) bool {
	switch condition.Operator {
	case "equals":
		return fieldValue == fmt.Sprintf("%v", condition.Value)
	case "contains":
		return len(fieldValue) > 0 &&
			len(fieldValue) >= len(fmt.Sprintf("%v", condition.Value)) &&
			fieldValue == fmt.Sprintf("%v", condition.Value)
	case "in":
		for _, value := range condition.Values {
			if fieldValue == fmt.Sprintf("%v", value) {
				return true
			}
		}
		return false
	}
	return false
}

// SaveConfig saves the configuration to a file
func (c *AdvancedConfig) SaveConfig(configPath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// MergeConfig merges another configuration into this one
func (c *AdvancedConfig) MergeConfig(other *AdvancedConfig) {
	// Merge checkers
	for id, config := range other.Checkers {
		c.Checkers[id] = config
	}

	// Merge analyzers
	for lang, config := range other.Analyzers {
		c.Analyzers[lang] = config
	}

	// Merge reporters
	for id, config := range other.Reporters {
		c.Reporters[id] = config
	}

	// Merge categories
	for name, config := range other.Categories {
		c.Categories[name] = config
	}

	// Append overrides
	c.Overrides = append(c.Overrides, other.Overrides...)
}

// FilterByCategories creates a new AdvancedConfig with only checkers and analyzers
// that belong to the specified categories. If no categories are specified, returns
// the original config unchanged.
func (c *AdvancedConfig) FilterByCategories(categories []string) *AdvancedConfig {
	if len(categories) == 0 {
		return c
	}

	// Create a copy of the configuration
	filtered := &AdvancedConfig{
		Version:    c.Version,
		Engine:     c.Engine,
		Checkers:   make(map[string]core.CheckerConfig),
		Analyzers:  make(map[string]core.AnalyzerConfig),
		Reporters:  c.Reporters,  // Copy reporters as-is
		Categories: c.Categories, // Copy categories as-is
		Overrides:  c.Overrides,  // Copy overrides as-is
	}

	// Create a set of target categories for efficient lookup
	categorySet := make(map[string]bool)
	for _, cat := range categories {
		categorySet[cat] = true
	}

	// Filter checkers by categories
	for id, checker := range c.Checkers {
		// Check if any of the checker's categories match our target categories
		for _, checkerCategory := range checker.Categories {
			if categorySet[checkerCategory] {
				filtered.Checkers[id] = checker
				break
			}
		}
	}

	// Filter analyzers by categories (analyzers need a different approach since they don't have explicit categories)
	// For now, we'll include analyzers based on language categories or keep all analyzers
	// This can be enhanced later if analyzers get explicit category support
	for lang, analyzer := range c.Analyzers {
		// For now, we include all analyzers if any category is specified
		// This could be enhanced by adding category support to AnalyzerConfig
		filtered.Analyzers[lang] = analyzer
	}

	return filtered
}
