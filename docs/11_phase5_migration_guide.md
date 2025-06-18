# Migration Reference Guide

## Overview

This reference guide documents the migration features implemented in the modular architecture. These features provide tools and processes for configuration migration, feature flag control, and backward compatibility.

**Note**: The migration is complete and the system is production-ready. This guide serves as reference documentation for the migration features that remain available for ongoing configuration management.

## Key Features Available

### 1. Feature Flags System

Feature flags enable gradual migration by allowing components to be enabled/disabled individually:

```yaml
# Configuration example with feature flags
feature_flags:
  - name: "modular_engine"
    enabled: false
    description: "Enable the new modular execution engine"
  - name: "parallel_execution"
    enabled: true
    description: "Enable parallel execution of checks and analysis"
  - name: "config_migration"
    enabled: true
    description: "Enable automatic configuration migration"
```

Available feature flags:

- `modular_engine`: Enable the new modular execution engine
- `modular_checkers`: Enable the new modular checkers system
- `modular_analyzers`: Enable the new modular analyzers system
- `parallel_execution`: Enable parallel execution capabilities
- `advanced_config`: Enable advanced configuration format support
- `legacy_compatibility`: Enable legacy compatibility mode
- `config_migration`: Enable automatic configuration migration

### 2. Configuration Migration System

Automatic migration from legacy to advanced configuration formats:

```go
// Usage example
migrationManager := config.NewMigrationManager(logger)
migrationManager.InitializeFeatureFlags()

// Automatically detects format and migrates if needed
cfg, err := migrationManager.LoadConfigWithMigration("config.yaml")
```

The migration system:
- Detects configuration format automatically
- Migrates legacy configs to advanced format
- Preserves all existing functionality
- Creates backup of original configuration

### 3. Gradual Cutover Process

The migration manager enables gradual component activation:

```go
// Check which components are enabled
if migrationManager.IsModularEngineEnabled() {
    // Use new modular engine
} else {
    // Use legacy engine
}

if migrationManager.IsParallelExecutionEnabled() {
    // Enable parallel processing
}
```

### 4. Backward Compatibility

Maintains full backward compatibility through:
- Legacy configuration support with automatic migration
- Adapter patterns for old interfaces
- Feature flags to control transition pace
- CLI interface remains unchanged

## Configuration Management

The system supports both legacy and advanced configuration formats with automatic migration:

### Current Recommendation

Use the advanced configuration format for new deployments:

```yaml
# examples/advanced-config-sample.yaml
version: "1.0"

feature_flags:
  - name: "parallel_execution"
    enabled: true
  - name: "modular_engine"
    enabled: true

# ... rest of configuration
```

## Configuration Examples

### Legacy Configuration (Automatically Migrated)

Legacy configurations are automatically detected and migrated:

```yaml
# legacy-config.yaml
repositories:
  - name: "my-repo"
    url: "https://github.com/user/repo"
    branch: "main"

checkers:
  enabled: ["git", "dependencies"]
  timeout: 30s

complexity:
  threshold: 10
  languages:
    python: 8
    java: 12
```

### Advanced Configuration (Recommended)

Use the advanced format for new configurations:

```yaml
# advanced-config.yaml
version: "1.0"

feature_flags:
  - name: "parallel_execution"
    enabled: true
  - name: "config_migration"
    enabled: true

engine:
  max_concurrency: 4
  timeout: 5m
  cache_enabled: true

checkers:
  git-status:
    enabled: true
    severity: medium
    timeout: 30s

analyzers:
  python:
    enabled: true
    file_extensions: [".py"]
    complexity_enabled: true

profiles:
  default:
    name: "Default Profile"
    checkers: 
      git-status:
        enabled: true
        severity: medium
        timeout: 30s
    analyzers:
      python:
        enabled: true
        file_extensions: [".py"]
        complexity_enabled: true
```

## CLI Usage

### Standard Health Checks

```bash
# Basic usage - automatic config handling
./repos health --config config.yaml

# With specific profile
./repos health --config config.yaml --profile development
```

### Feature Flag Management

Feature flags can be managed through:

1. **Configuration file**: Define in YAML config
2. **Runtime**: Programmatically via `MigrationManager.SetFlag()`
3. **Environment**: Via environment variables (future enhancement)

## Training Materials

### For End Users

1. **Basic Migration**: Use existing CLI commands - migration is transparent
2. **Configuration Updates**: Learn new advanced configuration format
3. **Feature Flags**: Understand how to control component activation

### For Developers

1. **Architecture Understanding**: Study modular design patterns
2. **Extension Development**: Learn how to create new checkers/analyzers
3. **Testing Strategies**: Use new testing infrastructure

## Monitoring Migration

### Logging

The migration manager provides detailed logging:

```
INFO  Feature flags initialized for gradual migration count=7
INFO  Detected configuration format format=legacy path=config.yaml
INFO  Starting automatic configuration migration legacy_path=config.yaml advanced_path=config-advanced.yaml
INFO  Configuration migration completed successfully advanced_path=config-advanced.yaml
INFO  Enabling gradual cutover with current feature flags flags=map[...]
```

### Health Checks

Monitor system health through:
- Configuration validation
- Feature flag status
- Component activation logging
- Performance metrics

## Troubleshooting

### Common Issues

1. **Configuration Issues**
   - Check configuration syntax
   - Verify file permissions
   - Review system logs

2. **Feature Flags Not Working**
   - Ensure flags are properly defined in config
   - Check flag names match constants
   - Verify system initialization

3. **Performance Issues**
   - Monitor parallel execution settings
   - Adjust concurrency limits
   - Check cache configuration

### Support

- Review system logs for detailed error information
- Check feature flag status via logging output
- Validate configuration format

## Best Practices

1. **Use Advanced Config**: Prefer advanced configuration format for new setups
2. **Monitor Performance**: Watch for any performance issues
3. **Test Thoroughly**: Test configurations in non-production first
4. **Backup Configs**: Keep configuration backups
5. **Monitor Logs**: Watch system logs for issues

## Future Enhancements

The current system sets the foundation for:
- Environment-based feature flag control
- Web-based configuration management
- Advanced monitoring dashboards
- Plugin ecosystem expansion
