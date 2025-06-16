package config

import (
	"sync"
)

// FeatureFlags manages feature flags for gradual migration
type FeatureFlags struct {
	mu    sync.RWMutex
	flags map[string]bool
}

// NewFeatureFlags creates a new feature flags manager
func NewFeatureFlags() *FeatureFlags {
	return &FeatureFlags{
		flags: make(map[string]bool),
	}
}

// LoadFlags loads feature flags from configuration
func (f *FeatureFlags) LoadFlags(flags []FeatureFlag) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, flag := range flags {
		f.flags[flag.Name] = flag.Enabled
	}
}

// IsEnabled checks if a feature flag is enabled
func (f *FeatureFlags) IsEnabled(flagName string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	enabled, exists := f.flags[flagName]
	return exists && enabled
}

// SetFlag sets a feature flag value
func (f *FeatureFlags) SetFlag(flagName string, enabled bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.flags[flagName] = enabled
}

// GetAllFlags returns all feature flags
func (f *FeatureFlags) GetAllFlags() map[string]bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make(map[string]bool)
	for name, enabled := range f.flags {
		result[name] = enabled
	}
	return result
}

// Default feature flag names for migration
const (
	// FlagModularEngine enables the new modular engine
	FlagModularEngine = "modular_engine"

	// FlagModularCheckers enables the new modular checkers
	FlagModularCheckers = "modular_checkers"

	// FlagModularAnalyzers enables the new modular analyzers
	FlagModularAnalyzers = "modular_analyzers"

	// FlagParallelExecution enables parallel execution
	FlagParallelExecution = "parallel_execution"

	// FlagAdvancedConfig enables advanced configuration format
	FlagAdvancedConfig = "advanced_config"

	// FlagLegacyCompatibility enables legacy compatibility mode
	FlagLegacyCompatibility = "legacy_compatibility"

	// FlagConfigMigration enables automatic config migration
	FlagConfigMigration = "config_migration"
)

// GetDefaultFlags returns the default feature flags for migration
func GetDefaultFlags() []FeatureFlag {
	return []FeatureFlag{
		{
			Name:        FlagModularEngine,
			Enabled:     false,
			Description: "Enable the new modular execution engine",
		},
		{
			Name:        FlagModularCheckers,
			Enabled:     false,
			Description: "Enable the new modular checkers system",
		},
		{
			Name:        FlagModularAnalyzers,
			Enabled:     false,
			Description: "Enable the new modular analyzers system",
		},
		{
			Name:        FlagParallelExecution,
			Enabled:     true,
			Description: "Enable parallel execution of checks and analysis",
		},
		{
			Name:        FlagAdvancedConfig,
			Enabled:     true,
			Description: "Enable advanced configuration format support",
		},
		{
			Name:        FlagLegacyCompatibility,
			Enabled:     true,
			Description: "Enable legacy compatibility mode",
		},
		{
			Name:        FlagConfigMigration,
			Enabled:     true,
			Description: "Enable automatic configuration migration",
		},
	}
}
