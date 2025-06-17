package config

import (
	"fmt"
	"strings"
)

// ConfigValidator provides validation for configuration files
type ConfigValidator struct {
	rules []ValidationRule
}

// ValidationRule defines the interface for configuration validation rules
type ValidationRule interface {
	Validate(config *AdvancedConfig) error
	GetDescription() string
}

// BasicValidationRule defines validation for basic config
type BasicValidationRule interface {
	ValidateBasic(config *Config) error
	GetDescription() string
}

// NewConfigValidator creates a new configuration validator with default rules
func NewConfigValidator() *ConfigValidator {
	validator := &ConfigValidator{
		rules: make([]ValidationRule, 0),
	}

	// Add default validation rules
	validator.AddRule(&ProfileValidationRule{})
	validator.AddRule(&PipelineValidationRule{})
	validator.AddRule(&EngineValidationRule{})

	return validator
}

// AddRule adds a validation rule to the validator
func (v *ConfigValidator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// Validate validates the configuration against all rules
func (v *ConfigValidator) Validate(config *AdvancedConfig) error {
	var errors []string

	for _, rule := range v.rules {
		if err := rule.Validate(config); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", rule.GetDescription(), err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// ValidateBasicConfig validates a basic configuration
func (v *ConfigValidator) ValidateBasicConfig(config *Config) error {
	if len(config.Repositories) == 0 {
		return fmt.Errorf("no repositories configured")
	}

	if len(config.Repositories) > 1000 {
		return fmt.Errorf("too many repositories: %d (max: 1000)", len(config.Repositories))
	}

	for i, repo := range config.Repositories {
		if repo.Name == "" {
			return fmt.Errorf("repository %d: name cannot be empty", i+1)
		}
		if repo.URL == "" {
			return fmt.Errorf("repository '%s': URL cannot be empty", repo.Name)
		}
	}

	return nil
}

// EngineValidationRule validates engine configuration
type EngineValidationRule struct{}

func (r *EngineValidationRule) Validate(config *AdvancedConfig) error {
	if config.Engine.MaxConcurrency < 1 {
		return fmt.Errorf("engine max_concurrency must be at least 1")
	}
	if config.Engine.MaxConcurrency > 100 {
		return fmt.Errorf("engine max_concurrency too high: %d (max: 100)", config.Engine.MaxConcurrency)
	}
	return nil
}

func (r *EngineValidationRule) GetDescription() string {
	return "Engine Configuration Validation"
}

// ProfileValidationRule validates profile configurations
type ProfileValidationRule struct{}

func (r *ProfileValidationRule) Validate(config *AdvancedConfig) error {
	for name, profile := range config.Profiles {
		if name == "" {
			return fmt.Errorf("profile name cannot be empty")
		}
		if profile.Description == "" {
			return fmt.Errorf("profile '%s' missing description", name)
		}
	}
	return nil
}

func (r *ProfileValidationRule) GetDescription() string {
	return "Profile Configuration Validation"
}

// PipelineValidationRule validates pipeline configurations
type PipelineValidationRule struct{}

func (r *PipelineValidationRule) Validate(config *AdvancedConfig) error {
	for name, pipeline := range config.Pipelines {
		if name == "" {
			return fmt.Errorf("pipeline name cannot be empty")
		}
		if len(pipeline.Steps) == 0 {
			return fmt.Errorf("pipeline '%s' has no steps configured", name)
		}
		for i, step := range pipeline.Steps {
			if step.Name == "" {
				return fmt.Errorf("pipeline '%s' step %d missing name", name, i+1)
			}
		}
	}
	return nil
}

func (r *PipelineValidationRule) GetDescription() string {
	return "Pipeline Configuration Validation"
}
