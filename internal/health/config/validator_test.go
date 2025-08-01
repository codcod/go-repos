package config

import (
	"testing"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/core"
)

func TestConfigValidator_ValidateBasicConfig(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				Repositories: []config.Repository{
					{Name: "test-repo", URL: "https://github.com/test/repo.git"},
				},
			},
			wantErr: false,
		},
		{
			name: "no repositories",
			config: &config.Config{
				Repositories: []config.Repository{},
			},
			wantErr: true,
		},
		{
			name: "repository with empty name",
			config: &config.Config{
				Repositories: []config.Repository{
					{Name: "", URL: "https://github.com/test/repo.git"},
				},
			},
			wantErr: true,
		},
		{
			name: "repository with empty URL",
			config: &config.Config{
				Repositories: []config.Repository{
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
