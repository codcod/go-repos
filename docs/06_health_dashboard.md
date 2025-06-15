# Repository Health Dashboard

The Repository Health Dashboard is a comprehensive feature that analyzes multiple aspects of your repositories to provide insights into their maintenance status, security posture, and overall health.

## Features

### Health Checks

The dashboard performs the following health checks:

#### Git Status
- **Category**: `git`
- **Description**: Checks for uncommitted changes and git repository status
- **Healthy**: Clean working directory
- **Warning**: Uncommitted changes present
- **Critical**: Not a git repository

#### Last Commit
- **Category**: `git`
- **Description**: Analyzes when the last commit was made
- **Healthy**: Recent activity (< 90 days)
- **Warning**: Moderate inactivity (90-365 days)
- **Critical**: Very stale (> 365 days)

#### Dependencies
- **Category**: `dependencies`
- **Description**: Checks dependency files and their status
- **Supported**: Go (go.mod), Node.js (package.json), Python (requirements.txt)
- **Healthy**: Dependencies are up to date
- **Warning**: Dependencies need updating or files missing

#### Security
- **Category**: `security`
- **Description**: Scans for security vulnerabilities and policies
- **Features**: Go vulnerability scanning via `govulncheck`, security policy presence
- **Healthy**: No vulnerabilities, security policy present
- **Warning**: No security policy
- **Critical**: Security vulnerabilities found

#### License
- **Category**: `compliance`
- **Description**: Checks for license file presence
- **Healthy**: License file found
- **Warning**: No license file

#### CI/CD
- **Category**: `automation`
- **Description**: Checks for continuous integration configuration
- **Supported**: GitHub Actions, GitLab CI, Travis CI, Jenkins, CircleCI, Azure Pipelines
- **Healthy**: CI/CD configuration found
- **Warning**: No CI/CD configuration

#### Documentation
- **Category**: `documentation`
- **Description**: Analyzes documentation quality
- **Healthy**: Good README with multiple sections
- **Warning**: Short README or missing sections
- **Critical**: No README file

#### Branch Protection
- **Category**: `security`
- **Description**: Checks branch protection settings (requires GitHub API)
- **Status**: Currently placeholder - requires GitHub token

#### Cyclomatic Complexity
- **Category**: `code-quality`
- **Description**: Analyzes code complexity to identify overly complex functions
- **Supported Languages**: Go (gocyclo), Python (radon), JavaScript/TypeScript (eslint), Java (PMD)
- **Healthy**: Low complexity (< 10 cyclomatic complexity)
- **Warning**: Moderate complexity (10-15)
- **Critical**: High complexity (> 15)

## Usage

### Basic Health Check

```bash
# Check all repositories
repos health

# Check repositories with specific tag
repos health --tag backend

# Run checks in parallel
repos health --parallel
```

### Output Formats

```bash
# Table format (default)
repos health --format table

# JSON format
repos health --format json

# HTML report
repos health --format html

# Save to file
repos health --format html --output-file health-report.html
```

### Category Filtering

```bash
# List all available categories with descriptions
repos health --list-categories

# Check only specific categories
repos health --categories git,security,documentation

# Check code quality metrics
repos health --categories code-quality

# Exclude certain categories
repos health --exclude dependencies,automation

# Available categories:
# - git: Git repository status and commit history checks
# - dependencies: Dependency management and outdated package checks
# - security: Security vulnerabilities and policy checks
# - code-quality: Code quality metrics and complexity analysis
# - compliance: License and legal compliance checks
# - automation: CI/CD and automation configuration checks
# - documentation: Documentation completeness and quality checks
```

### Summary View

```bash
# Show compact summary table
repos health --summary

# Set score threshold (exit with error if below threshold)
repos health --threshold 80
```

## Exit Codes

The health command uses exit codes to indicate the overall status:

- **0**: All repositories healthy
- **1**: Warning issues found
- **2**: Critical issues found

This makes it suitable for use in CI/CD pipelines:

```bash
# In CI pipeline
repos health --threshold 75 || exit 1
```

## Configuration

### Environment Variables

- `GITHUB_TOKEN`: Required for GitHub API features (branch protection checks)

### Config File Integration

The health command respects the same configuration as other repos commands:

```yaml
repositories:
  - name: my-app
    url: https://github.com/owner/my-app.git
    tags: [backend, go]
    path: ./my-app
  - name: frontend
    url: https://github.com/owner/frontend.git
    tags: [frontend, react]
    path: ./frontend
```

## Examples

### CI/CD Integration

```yaml
# GitHub Actions example
name: Repository Health Check
on: [push, pull_request]

jobs:
  health-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Install analysis tools
        run: |
          go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
          pip install radon
          npm install -g eslint
      - name: Setup repos tool
        run: |
          go install github.com/codcod/repos/cmd/repos@latest
      - name: Check repository health
        run: |
          repos health --format json --output-file health-report.json
      - name: Check code quality specifically 
        run: |
          repos health --categories code-quality --threshold 80
      - name: Upload health report
        uses: actions/upload-artifact@v3
        with:
          name: health-report
          path: health-report.json
```

### Weekly Health Reports

```bash
#!/bin/bash
# weekly-health-check.sh

# Generate HTML health report
repos health --format html --output-file weekly-health-$(date +%Y%m%d).html

# Send email if critical issues found
if repos health --threshold 60; then
    echo "All repositories healthy"
else
    mail -s "Repository Health Alert" admin@company.com < weekly-health-$(date +%Y%m%d).html
fi
```

### Custom Health Dashboard

```bash
# List all available categories and their descriptions
repos health --list-categories

# Generate data for custom dashboard
repos health --format json | jq '.repositories[] | select(.status == "critical")'

# Get summary statistics
repos health --format json | jq '.summary'

# Check only code quality across all repositories
repos health --categories code-quality --format json

# Export specific category results for analysis
repos health --categories security,dependencies --format yaml --output-file security-deps-report.yaml
```

## Implementation Details

### Architecture

The health dashboard is implemented with a modular architecture:

```
internal/health/
├── health.go      # Core types and orchestration
├── checkers.go    # Individual health checker implementations
├── formatter.go   # Output formatting (table, JSON, HTML)
└── health_test.go # Comprehensive tests
```

### Extending Health Checks

To add a new health checker:

1. Implement the `HealthChecker` interface:

```go
type CustomChecker struct{}

func (c *CustomChecker) Name() string     { return "Custom Check" }
func (c *CustomChecker) Category() string { return "custom" }

func (c *CustomChecker) Check(repo config.Repository) HealthCheck {
    // Implement your check logic here
    return HealthCheck{
        Name:        c.Name(),
        Status:      HealthStatusHealthy,
        Message:     "Check passed",
        Category:    c.Category(),
        LastChecked: time.Now(),
    }
}
```

2. Add it to the `GetHealthCheckers` function in `health.go`

### Performance Considerations

- Health checks can be run in parallel using the `--parallel` flag
- Individual checkers are designed to be fast and non-destructive
- Large repositories may take longer for dependency analysis
- GitHub API checks require network access and may be rate-limited

## Troubleshooting

### Common Issues

1. **"govulncheck not found"**
   ```bash
   go install golang.org/x/vuln/cmd/govulncheck@latest
   ```

2. **"gocyclo not found"** (for Go complexity analysis)
   ```bash
   go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
   ```

3. **"radon not found"** (for Python complexity analysis)
   ```bash
   pip install radon
   ```

4. **"eslint not found"** (for JavaScript/TypeScript complexity analysis)
   ```bash
   npm install -g eslint
   # Configure eslint with complexity rules in your project
   ```

5. **"PMD not found"** (for Java complexity analysis)
   ```bash
   # Download from https://pmd.github.io/
   # Or install via package manager (brew install pmd on macOS)
   ```

6. **"Permission denied" errors**
   - Ensure the repos tool has read access to repository directories
   - Check file permissions on .git directories

7. **GitHub API rate limiting**
   - Set `GITHUB_TOKEN` environment variable
   - Use personal access token with appropriate permissions

8. **Slow performance**
   - Use `--parallel` flag for faster execution
   - Use `--categories` to limit checks to specific areas
   - Consider excluding large repositories from certain checks

### Debug Mode

Enable verbose logging to debug health check issues:

```bash
# Set log level (if implemented)
REPOS_LOG_LEVEL=debug repos health

# Check individual repositories
repos health --tag specific-repo
```

## Future Enhancements

Planned improvements for the health dashboard:

1. **Additional Checkers**
   - Code coverage analysis
   - Dependency vulnerability scanning for more languages
   - ✅ ~~Code quality metrics integration~~ (Implemented: Cyclomatic complexity)
   - Performance benchmarking results
   - Test coverage metrics
   - Code duplication detection

2. **GitHub Integration**
   - Branch protection rule verification
   - Issue and PR analysis
   - GitHub security alerts integration
   - Repository settings compliance

3. **Reporting**
   - Trend analysis over time
   - Health score tracking
   - Custom report templates
   - Integration with monitoring systems

4. **Configuration**
   - Custom health check profiles
   - Severity level customization
   - Ignore patterns for specific checks
   - Team-specific health policies

5. **Enhanced Code Quality**
   - Additional complexity metrics (cognitive complexity, maintainability index)
   - Integration with SonarQube, CodeClimate
   - Custom complexity thresholds per language
   - Code style and formatting checks

## Code Quality Analysis

### Cyclomatic Complexity

The cyclomatic complexity checker analyzes code complexity across multiple programming languages. It helps identify overly complex functions that may be difficult to maintain, test, or debug.

#### Supported Languages and Tools

- **Go**: Uses `gocyclo` tool
  ```bash
  go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
  ```

- **Python**: Uses `radon` tool
  ```bash
  pip install radon
  ```

- **JavaScript/TypeScript**: Uses `eslint` with complexity rules
  ```bash
  npm install -g eslint
  # Requires eslint configuration with complexity rules
  ```

- **Java**: Uses PMD static analyzer
  ```bash
  # Download PMD from https://pmd.github.io/
  # Or install via package manager
  ```

#### Complexity Thresholds

- **Healthy** (Green): Cyclomatic complexity < 10
- **Warning** (Yellow): Cyclomatic complexity 10-15
- **Critical** (Red): Cyclomatic complexity > 15

#### Usage Examples

```bash
# Check code quality for all repositories
repos health --categories code-quality

# Check code quality for specific language projects
repos health --categories code-quality --tag go
repos health --categories code-quality --tag python

# Generate detailed report
repos health --categories code-quality --format json --output-file complexity-report.json
```

The complexity checker will automatically detect the appropriate tool based on the project type and provide detailed feedback about functions or methods that exceed the complexity thresholds.
