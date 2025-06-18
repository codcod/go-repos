# Migration Guide: v2.0 Improvements

This guide helps you understand and adopt the significant improvements made to the repos tool in v2.0. The changes focus on maintainability, observability, and user experience while maintaining backward compatibility.

## Overview of Changes

### 🎯 **What's New**
- **Structured Logging**: Rich, contextual logging with automatic operation tracking
- **Metrics Collection**: Performance monitoring and analytics
- **Enhanced Error Handling**: Detailed error context and recovery hints
- **Configuration Validation**: Comprehensive validation with helpful error messages
- **Improved Testing**: Builder patterns and test utilities for better coverage
- **Modular Architecture**: Clear separation of concerns and dependency injection

### 🔄 **Migration Impact**
- **Backward Compatible**: Existing configurations continue to work
- **Enhanced Logging**: More detailed and structured log output
- **Better Error Messages**: Clearer, actionable error reporting
- **Performance Insights**: Optional metrics collection and reporting

## Configuration Changes

### Basic Configuration (No Changes Required)

Your existing `config.yaml` files continue to work without modification:

```yaml
# This continues to work as before
repositories:
  - name: my-service
    url: git@github.com:org/my-service.git
    tags: [backend, go]
```

### Advanced Configuration (Enhanced)

New validation and enhanced configuration options:

```yaml
# Enhanced with validation and better error reporting
profiles:
  development:
    checkers:
      security:
        enabled: true
        severity_threshold: medium
      code_quality:
        enabled: true
        max_complexity: 15
  
  production:
    checkers:
      security:
        enabled: true
        severity_threshold: low  # Stricter for production
      code_quality:
        enabled: true
        max_complexity: 10
```

**Benefits:**
- Automatic validation with helpful error messages
- Profile support for different environments
- Better organization and documentation

## Command Line Changes

### Enhanced Verbose Output

The `--verbose` flag now provides structured, contextual information:

**Before:**
```
Running health checks...
Done.
```

**After:**
```
[2025-06-17 10:30:15] [INFO] [health-cmd] validating health command configuration
[2025-06-17 10:30:15] [INFO] [health-cmd] loading advanced configuration config_path=/path/to/config.yaml
[2025-06-17 10:30:15] [INFO] [health-cmd] loading repositories basic_config=/path/to/basic.yaml tag=backend
[2025-06-17 10:30:15] [INFO] [health-cmd] converted repositories count=5
[2025-06-17 10:30:15] [INFO] [health-cmd] executing comprehensive health checks repositories=5 profile=default
[2025-06-17 10:30:16] [INFO] [health-cmd] health checks completed successfully

=== Metrics Summary ===
Total Duration: 1.234s
Repositories Processed: 5 (4.05/sec)
Checks Executed: 25 (20.25/sec)

Counters:
  repositories_processed: 5
  checks_executed: 25
  checks_healthy: 20
  checks_warning: 3
  checks_critical: 2

Histograms:
  check_scores: count=25, min=65.00, max=98.00, mean=87.20, p95=95.00
  check_duration_ms: count=25, min=45.00, max=250.00, mean=125.30, p95=230.00
```

### New Options

```bash
# Enhanced logging levels (automatically determined by --verbose)
repos health --config config.yaml --verbose

# All existing options continue to work
repos health --config config.yaml --profile production --timeout 10m
```

## Error Handling Improvements

### Before: Basic Error Messages

```
Error: configuration invalid
Error: health check failed
```

### After: Contextual Error Messages

```
[ERROR] Configuration validation failed: repository name is required
  Context: field=name, repository_index=2
  Suggestion: Add a 'name' field to repository configuration

[ERROR] Health check execution failed: timeout exceeded
  Context: operation=security_scan, repository=my-service, timeout=5m0s
  Suggestion: Try increasing timeout with --timeout flag or check repository accessibility
```

**Benefits:**
- Clear identification of the problem
- Context showing where the error occurred
- Actionable suggestions for resolution

## Logging Improvements

### Output Format

**Before:**
```
my-repo | Starting analysis
my-repo | Analysis complete
```

**After (with structured logging):**
```
[2025-06-17 10:30:15] [INFO] [security-checker] starting security analysis context={repository=my-repo, language=go}
[2025-06-17 10:30:15] [INFO] [security-checker] analysis completed context={repository=my-repo} fields={duration=1.2s, vulnerabilities_found=0, dependencies_scanned=25}
```

### Operation Tracking

Operations are now automatically tracked with timing:

```
[2025-06-17 10:30:15] [DEBUG] [health-cmd] operation started context={operation=health_check}
[2025-06-17 10:30:16] [INFO] [health-cmd] operation completed context={operation=health_check} fields={duration=1.234s}
```

## Performance Monitoring

### New Metrics Collection

When running with `--verbose`, you'll see performance metrics:

```bash
repos health --config config.yaml --verbose
```

**Output includes:**
- Processing rates (repositories/second, checks/second)
- Duration histograms with percentiles
- Success/failure counters
- Resource utilization metrics

### Metrics Categories

1. **Throughput**: Repositories and checks processed per second
2. **Latency**: Distribution of operation durations
3. **Success Rates**: Counters for successful vs failed operations
4. **Resource Usage**: Memory and processing metrics

## Adoption Guide

### 1. Immediate Benefits (No Changes Required)

Simply update to v2.0 to get:
- Better error messages
- Enhanced logging
- Performance improvements
- More reliable execution

### 2. Enhanced Experience (Recommended)

Enable verbose logging for development and troubleshooting:

```bash
# Add --verbose for enhanced output
repos health --config config.yaml --verbose
```

### 3. Advanced Usage (Optional)

Take advantage of new features:

#### Configuration Validation
```bash
# Validate your configuration
repos health --config config.yaml --dry-run --verbose
```

#### Profile Usage
```yaml
# In your advanced config
profiles:
  development:
    checkers:
      security:
        enabled: true
        severity_threshold: medium
  
  production:
    checkers:
      security:
        enabled: true
        severity_threshold: low
```

```bash
# Use specific profiles
repos health --config config.yaml --profile production
```

#### Custom Timeouts
```bash
# Set appropriate timeouts for your environment
repos health --config config.yaml --timeout 10m
```

## Troubleshooting

### Common Issues and Solutions

#### 1. Configuration Validation Errors

**Issue**: Configuration validation fails with detailed error messages

**Solution**: The new validation system provides specific guidance:
```
[ERROR] Configuration validation failed: checker 'security' is missing required field 'enabled'
  Context: checker=security, profile=default
  Suggestion: Add 'enabled: true/false' to the security checker configuration
```

#### 2. Performance Issues

**Issue**: Health checks taking longer than expected

**Solution**: Use verbose mode to identify bottlenecks:
```bash
repos health --config config.yaml --verbose
```

Look for:
- High p95 values in duration histograms
- Low processing rates
- Error counters indicating failures

#### 3. Log Volume

**Issue**: Too much log output in verbose mode

**Solution**: Use verbose mode selectively:
- Use `--verbose` only when troubleshooting
- Normal mode provides concise, actionable output
- Consider adjusting timeout values to reduce retry operations

## Migration Checklist

### ✅ **Phase 1: Update and Verify**
- [ ] Update to repos v2.0
- [ ] Run existing configurations to verify compatibility
- [ ] Test with `--dry-run` to validate configurations

### ✅ **Phase 2: Enhanced Usage**
- [ ] Try `--verbose` mode for enhanced visibility
- [ ] Review error messages for improvement opportunities
- [ ] Identify performance bottlenecks using metrics

### ✅ **Phase 3: Optimization** (Optional)
- [ ] Create environment-specific profiles
- [ ] Adjust timeouts based on metrics
- [ ] Implement configuration validation in CI/CD

## Examples

### Basic Migration Example

**Before:**
```bash
# Simple health check
repos health --config config.yaml
```

**After (same command, enhanced output):**
```bash
# Same command, better output and error handling
repos health --config config.yaml

# Enhanced troubleshooting
repos health --config config.yaml --verbose
```

### Advanced Migration Example

**Before:**
```yaml
# Basic configuration
checkers:
  security:
    enabled: true
```

**After:**
```yaml
# Enhanced with profiles and validation
profiles:
  default:
    checkers:
      security:
        enabled: true
        severity_threshold: medium
      code_quality:
        enabled: true
        max_complexity: 15

checkers:
  security:
    enabled: true
    config:
      scan_dependencies: true
      check_secrets: true
  code_quality:
    enabled: true
    config:
      max_complexity: 15
      min_coverage: 80
```

## Support and Resources

### Documentation
- [Architecture Guide](architecture.md): Detailed technical overview
- [Maintainability Improvements](09_maintainability_improvements.md): Implementation details
- [Code Structure Analysis](12_code_structure_analysis.md): Design decisions

### Getting Help

1. **Configuration Issues**: Use `--dry-run --verbose` for detailed validation
2. **Performance Issues**: Use `--verbose` to identify bottlenecks
3. **Error Messages**: The new contextual errors provide specific guidance

### Community

The improvements maintain full backward compatibility while providing enhanced capabilities. You can adopt new features gradually as your needs evolve.
