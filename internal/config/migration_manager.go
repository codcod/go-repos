package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codcod/repos/internal/core"
)

// MigrationManager handles the gradual migration from legacy to modular architecture
type MigrationManager struct {
	featureFlags *FeatureFlags
	migrator     *ConfigMigrator
	logger       core.Logger
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(logger core.Logger) *MigrationManager {
	return &MigrationManager{
		featureFlags: NewFeatureFlags(),
		migrator:     NewConfigMigrator(logger),
		logger:       logger,
	}
}

// InitializeFeatureFlags initializes feature flags with default values
func (m *MigrationManager) InitializeFeatureFlags() {
	flags := GetDefaultFlags()
	m.featureFlags.LoadFlags(flags)

	m.logger.Info("Feature flags initialized for gradual migration",
		core.Int("total_flags", len(flags)))
}

// LoadConfigWithMigration loads configuration with automatic migration support
func (m *MigrationManager) LoadConfigWithMigration(configPath string) (core.Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	// Detect configuration format
	format, err := m.migrator.DetectConfigFormat(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect config format: %w", err)
	}

	m.logger.Info("Detected configuration format",
		core.String("format", format),
		core.String("path", configPath))

	switch format {
	case "legacy":
		return m.handleLegacyConfig(configPath)
	case "advanced":
		return m.handleAdvancedConfig(configPath)
	default:
		return nil, fmt.Errorf("unsupported configuration format: %s", format)
	}
}

// handleLegacyConfig processes legacy configuration
func (m *MigrationManager) handleLegacyConfig(configPath string) (core.Config, error) {
	// Check if automatic migration is enabled
	if !m.featureFlags.IsEnabled(FlagConfigMigration) {
		m.logger.Warn("Legacy configuration detected but migration is disabled",
			core.String("path", configPath))
		// For now, return an error. In the future, we could load legacy config directly
		return nil, fmt.Errorf("legacy configuration detected but migration is disabled")
	}

	// Perform automatic migration
	advancedPath := m.generateAdvancedConfigPath(configPath)

	m.logger.Info("Starting automatic configuration migration",
		core.String("legacy_path", configPath),
		core.String("advanced_path", advancedPath))

	err := m.migrator.MigrateConfig(configPath, advancedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate configuration: %w", err)
	}

	// Load the migrated configuration
	return m.handleAdvancedConfig(advancedPath)
}

// handleAdvancedConfig processes advanced configuration
func (m *MigrationManager) handleAdvancedConfig(configPath string) (core.Config, error) {
	config, err := LoadAdvancedConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load advanced config: %w", err)
	}

	// Load feature flags from config if present
	if len(config.FeatureFlags) > 0 {
		m.featureFlags.LoadFlags(config.FeatureFlags)
		m.logger.Info("Loaded feature flags from configuration",
			core.Int("count", len(config.FeatureFlags)))
	}

	return config, nil
}

// generateAdvancedConfigPath generates a path for the migrated advanced configuration
func (m *MigrationManager) generateAdvancedConfigPath(legacyPath string) string {
	dir := filepath.Dir(legacyPath)
	filename := filepath.Base(legacyPath)
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]

	return filepath.Join(dir, fmt.Sprintf("%s-advanced%s", name, ext))
}

// GetFeatureFlags returns the current feature flags manager
func (m *MigrationManager) GetFeatureFlags() *FeatureFlags {
	return m.featureFlags
}

// EnableGradualCutover enables components based on feature flags for gradual migration
func (m *MigrationManager) EnableGradualCutover() {
	flags := m.featureFlags.GetAllFlags()

	m.logger.Info("Enabling gradual cutover with current feature flags",
		core.Any("flags", flags))

	// Log which components are enabled/disabled
	for flagName, enabled := range flags {
		status := "disabled"
		if enabled {
			status = "enabled"
		}
		m.logger.Info("Feature flag status",
			core.String("flag", flagName),
			core.String("status", status))
	}
}

// IsModularEngineEnabled checks if modular engine should be used
func (m *MigrationManager) IsModularEngineEnabled() bool {
	return m.featureFlags.IsEnabled(FlagModularEngine)
}

// IsModularCheckersEnabled checks if modular checkers should be used
func (m *MigrationManager) IsModularCheckersEnabled() bool {
	return m.featureFlags.IsEnabled(FlagModularCheckers)
}

// IsModularAnalyzersEnabled checks if modular analyzers should be used
func (m *MigrationManager) IsModularAnalyzersEnabled() bool {
	return m.featureFlags.IsEnabled(FlagModularAnalyzers)
}

// IsParallelExecutionEnabled checks if parallel execution should be used
func (m *MigrationManager) IsParallelExecutionEnabled() bool {
	return m.featureFlags.IsEnabled(FlagParallelExecution)
}

// IsLegacyCompatibilityEnabled checks if legacy compatibility mode is enabled
func (m *MigrationManager) IsLegacyCompatibilityEnabled() bool {
	return m.featureFlags.IsEnabled(FlagLegacyCompatibility)
}

// SetFlag updates a feature flag (useful for runtime toggle)
func (m *MigrationManager) SetFlag(flagName string, enabled bool) {
	m.featureFlags.SetFlag(flagName, enabled)
	m.logger.Info("Feature flag updated",
		core.String("flag", flagName),
		core.Bool("enabled", enabled))
}
