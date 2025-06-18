// Package config provides configuration management specifically for the health command.
//
// This package contains the AdvancedConfig and related types that are used exclusively
// by the health command to configure checkers, analyzers, and orchestration engine settings.
//
// The package was moved from internal/config to internal/health/config to better reflect
// its specific purpose and to encapsulate health-specific configuration under the health
// package hierarchy.
//
// Key types:
//   - AdvancedConfig: Main configuration struct for health checks
//   - CategoryConfig: Configuration for check categories
//   - OverrideConfig: Conditional configuration overrides
//   - ConfigValidator: Validates advanced configuration
//
// Example usage:
//
//	config := healthconfig.NewDefaultAdvancedConfig()
//	config, err := healthconfig.LoadAdvancedConfigOrDefault("health.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	validator := healthconfig.NewConfigValidator()
//	if err := validator.Validate(config); err != nil {
//		log.Fatal(err)
//	}
package config
