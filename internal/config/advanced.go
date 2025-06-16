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
	Version      string                         `yaml:"version"`
	Engine       core.EngineConfig              `yaml:"engine"`
	Checkers     map[string]core.CheckerConfig  `yaml:"checkers"`
	Analyzers    map[string]core.AnalyzerConfig `yaml:"analyzers"`
	Reporters    map[string]core.ReporterConfig `yaml:"reporters"`
	Categories   map[string]CategoryConfig      `yaml:"categories"`
	Profiles     map[string]ProfileConfig       `yaml:"profiles"`
	Pipelines    map[string]PipelineConfig      `yaml:"pipelines"`
	Overrides    []OverrideConfig               `yaml:"overrides"`
	Extensions   ExtensionsConfig               `yaml:"extensions"`
	Integrations IntegrationsConfig             `yaml:"integrations"`
	FeatureFlags []FeatureFlag                  `yaml:"feature_flags,omitempty"`
}

// PipelineConfig represents a pipeline configuration
type PipelineConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Steps       []StepConfig      `yaml:"steps"`
	Config      ExecutionConfig   `yaml:"config"`
	Metadata    map[string]string `yaml:"metadata,omitempty"`
}

// StepConfig represents a step configuration
type StepConfig struct {
	Name         string                 `yaml:"name"`
	Type         string                 `yaml:"type"`
	Config       map[string]interface{} `yaml:"config"`
	Dependencies []string               `yaml:"dependencies,omitempty"`
	Enabled      bool                   `yaml:"enabled"`
	Timeout      string                 `yaml:"timeout,omitempty"`
}

// ExecutionConfig represents execution configuration
type ExecutionConfig struct {
	MaxConcurrency  int                    `yaml:"max_concurrency"`
	Timeout         string                 `yaml:"timeout"`
	FailFast        bool                   `yaml:"fail_fast"`
	RetryCount      int                    `yaml:"retry_count"`
	RetryDelay      string                 `yaml:"retry_delay"`
	ContinueOnError bool                   `yaml:"continue_on_error"`
	OutputFormats   []string               `yaml:"output_formats"`
	ReportingConfig map[string]interface{} `yaml:"reporting_config"`
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

// ProfileConfig defines a configuration profile
type ProfileConfig struct {
	Name        string                         `yaml:"name"`
	Description string                         `yaml:"description"`
	Base        string                         `yaml:"base,omitempty"` // Inherit from another profile
	Checkers    map[string]core.CheckerConfig  `yaml:"checkers"`
	Analyzers   map[string]core.AnalyzerConfig `yaml:"analyzers"`
	Categories  []string                       `yaml:"categories"`
	Exclusions  []string                       `yaml:"exclusions"`
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

// FeatureFlag defines a feature flag configuration
type FeatureFlag struct {
	Name        string `yaml:"name" json:"name"`
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
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
	if c.Profiles == nil {
		c.Profiles = make(map[string]ProfileConfig)
	}
}

// validate validates the configuration
func (c *AdvancedConfig) validate() error {
	// Validate profiles don't have circular dependencies
	if err := c.validateProfileDependencies(); err != nil {
		return err
	}

	// Validate override conditions
	for _, override := range c.Overrides {
		if err := c.validateOverrideConditions(override); err != nil {
			return fmt.Errorf("invalid override '%s': %w", override.Name, err)
		}
	}

	return nil
}

// validateProfileDependencies checks for circular profile dependencies
func (c *AdvancedConfig) validateProfileDependencies() error {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for profileName := range c.Profiles {
		if err := c.checkProfileDependency(profileName, visited, recursionStack); err != nil {
			return err
		}
	}

	return nil
}

// checkProfileDependency performs DFS to detect circular dependencies
func (c *AdvancedConfig) checkProfileDependency(profileName string, visited, recursionStack map[string]bool) error {
	if recursionStack[profileName] {
		return fmt.Errorf("circular dependency detected in profile '%s'", profileName)
	}

	if visited[profileName] {
		return nil
	}

	profile, exists := c.Profiles[profileName]
	if !exists {
		return fmt.Errorf("profile '%s' not found", profileName)
	}

	visited[profileName] = true
	recursionStack[profileName] = true

	if profile.Base != "" {
		if err := c.checkProfileDependency(profile.Base, visited, recursionStack); err != nil {
			return err
		}
	}

	recursionStack[profileName] = false
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

// ApplyProfile applies a configuration profile
func (c *AdvancedConfig) ApplyProfile(profileName string, profile ProfileConfig) error {
	// Apply base profile first if specified
	if profile.Base != "" {
		baseProfile, exists := c.Profiles[profile.Base]
		if !exists {
			return fmt.Errorf("base profile '%s' not found", profile.Base)
		}
		if err := c.ApplyProfile(profile.Base, baseProfile); err != nil {
			return fmt.Errorf("failed to apply base profile '%s': %w", profile.Base, err)
		}
	}

	// Apply profile configurations
	for checkerID, checkerConfig := range profile.Checkers {
		c.Checkers[checkerID] = checkerConfig
	}

	for language, analyzerConfig := range profile.Analyzers {
		c.Analyzers[language] = analyzerConfig
	}

	return nil
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

	// Merge profiles
	for name, config := range other.Profiles {
		c.Profiles[name] = config
	}

	// Append overrides
	c.Overrides = append(c.Overrides, other.Overrides...)
}
