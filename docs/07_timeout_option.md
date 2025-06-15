# Timeout Option for Health Command

The health command now supports a `--timeout` option to control how long individual health checks can run before timing out.

## Usage

```bash
# Use default timeout (30 seconds)
repos health --config config.yaml

# Set custom timeout (10 seconds)
repos health --config config.yaml --timeout 10

# Set very short timeout for quick checks (5 seconds)
repos health --config config.yaml --timeout 5 --summary
```

## Examples

### Quick Health Check (5 second timeout)
```bash
repos health --timeout 5 --summary
```

### Comprehensive Health Check (60 second timeout)
```bash
repos health --timeout 60 --format json --output-file health-report.json
```

### Fast Parallel Health Check
```bash
repos health --timeout 10 --parallel --summary
```

## Timeout Behavior

The timeout applies to individual operations within health checks, such as:

- Maven dependency analysis (`mvn dependency:analyze`)
- Maven dependency resolution (`mvn dependency:resolve`)
- Gradle dependency checks
- Python pip dependency validation
- GitHub CLI branch protection checks

When a timeout occurs, the health check will:
1. Log a warning message indicating the operation timed out
2. Continue with other health checks
3. Not fail the entire health check process

## Default Values

- Default timeout: 30 seconds
- Minimum timeout: 1 second
- If timeout is set to 0 or negative, it uses the default value

## Benefits

- **Faster CI/CD**: Use shorter timeouts in CI environments where speed is important
- **Thorough Analysis**: Use longer timeouts for comprehensive health analysis
- **Responsive Interface**: Prevents health checks from hanging indefinitely
- **Configurable**: Adjust timeout based on network conditions and system performance
