package config

import (
	"testing"

	"github.com/codcod/repos/internal/core"
)

func TestConfigValidator_ValidateBasicConfig(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Repositories: []Repository{
					{Name: "test-repo", URL: "https://github.com/test/repo.git"},
				},
			},
			wantErr: false,
		},
		{
			name: "no repositories",
			config: &Config{
				Repositories: []Repository{},
			},
			wantErr: true,
		},
		{
			name: "repository with empty name",
			config: &Config{
				Repositories: []Repository{
					{Name: "", URL: "https://github.com/test/repo.git"},
				},
			},
			wantErr: true,
		},
		{
			name: "repository with empty URL",
			config: &Config{
				Repositories: []Repository{
					{Name: "test-repo", URL: ""},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateBasicConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBasicConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAdvancedConfigValidation(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name    string
		config  *AdvancedConfig
		wantErr bool
	}{
		{
			name: "valid advanced config",
			config: &AdvancedConfig{
				Version: "1.0",
				Engine: core.EngineConfig{
					MaxConcurrency: 4,
				},
				Profiles: map[string]ProfileConfig{
					"dev": {Description: "Development profile"},
				},
				Pipelines: map[string]PipelineConfig{
					"default": {
						Name:        "default",
						Description: "Default pipeline",
						Steps:       []StepConfig{{Name: "test-step", Type: "check"}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid engine config",
			config: &AdvancedConfig{
				Version: "1.0",
				Engine: core.EngineConfig{
					MaxConcurrency: 0, // Invalid
				},
			},
			wantErr: true,
		},
		{
			name: "profile without description",
			config: &AdvancedConfig{
				Version: "1.0",
				Engine: core.EngineConfig{
					MaxConcurrency: 4,
				},
				Profiles: map[string]ProfileConfig{
					"dev": {Description: ""}, // Invalid
				},
			},
			wantErr: true,
		},
		{
			name: "pipeline without steps",
			config: &AdvancedConfig{
				Version: "1.0",
				Engine: core.EngineConfig{
					MaxConcurrency: 4,
				},
				Pipelines: map[string]PipelineConfig{
					"default": {
						Name:        "default",
						Description: "Default pipeline",
						Steps:       []StepConfig{}, // Invalid
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationRules(t *testing.T) {
	t.Run("ProfileValidationRule", func(t *testing.T) {
		rule := &ProfileValidationRule{}

		// Valid profile
		config := &AdvancedConfig{
			Profiles: map[string]ProfileConfig{
				"dev": {Description: "Development profile"},
			},
		}
		if err := rule.Validate(config); err != nil {
			t.Errorf("Expected no error for valid profile, got: %v", err)
		}

		// Invalid profile
		config.Profiles["invalid"] = ProfileConfig{Description: ""}
		if err := rule.Validate(config); err == nil {
			t.Error("Expected error for profile without description")
		}
	})

	t.Run("PipelineValidationRule", func(t *testing.T) {
		rule := &PipelineValidationRule{}

		// Valid pipeline
		config := &AdvancedConfig{
			Pipelines: map[string]PipelineConfig{
				"test": {
					Name:        "test",
					Description: "Test pipeline",
					Steps:       []StepConfig{{Name: "step1", Type: "check"}},
				},
			},
		}
		if err := rule.Validate(config); err != nil {
			t.Errorf("Expected no error for valid pipeline, got: %v", err)
		}

		// Invalid pipeline - no steps
		config.Pipelines["invalid"] = PipelineConfig{
			Name:        "invalid",
			Description: "Invalid pipeline",
			Steps:       []StepConfig{},
		}
		if err := rule.Validate(config); err == nil {
			t.Error("Expected error for pipeline without steps")
		}
	})

	t.Run("EngineValidationRule", func(t *testing.T) {
		rule := &EngineValidationRule{}

		// Valid engine config
		config := &AdvancedConfig{
			Engine: core.EngineConfig{MaxConcurrency: 4},
		}
		if err := rule.Validate(config); err != nil {
			t.Errorf("Expected no error for valid engine config, got: %v", err)
		}

		// Invalid engine config - too low
		config.Engine.MaxConcurrency = 0
		if err := rule.Validate(config); err == nil {
			t.Error("Expected error for max_concurrency = 0")
		}

		// Invalid engine config - too high
		config.Engine.MaxConcurrency = 101
		if err := rule.Validate(config); err == nil {
			t.Error("Expected error for max_concurrency > 100")
		}
	})
}
