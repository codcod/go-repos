package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codcod/repos/internal/core"
)

// MigrationManager handles configuration migration and loading
type MigrationManager struct {
	migrator *ConfigMigrator
	logger   core.Logger
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(logger core.Logger) *MigrationManager {
	return &MigrationManager{
		migrator: NewConfigMigrator(logger),
		logger:   logger,
	}
}

// LoadConfig loads configuration with automatic migration support
func (m *MigrationManager) LoadConfig(configPath string) (core.Config, error) {
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

// handleLegacyConfig processes legacy configuration with automatic migration
func (m *MigrationManager) handleLegacyConfig(configPath string) (core.Config, error) {
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
