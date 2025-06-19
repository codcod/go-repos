[![Release](https://img.shields.io/github/v/release/codcod/repos?style=flat-square)](https://github.com/codcod/repos/releases)

# Repos

A CLI tool to manage multiple GitHub repositories. Clone them, run commands
across all repositories, create pull requests, and more—all with colored output
and comprehensive logging.

## Features

- 🚀 **Multi-repository management**: Clone and manage multiple repositories from a single config file
- 🏷️ **Tag-based filtering**: Run commands on specific repository groups using tags
- ⚡ **Parallel execution**: Execute commands across repositories simultaneously for faster operations
- 🎨 **Colorized output**: Real-time colored logs with repository identification
- 📝 **Comprehensive logging**: Per-repository log files for detailed command history
- 🔄 **Pull request automation**: Create and manage pull requests across multiple repositories
- 🏥 **Health dashboard**: Comprehensive repository health analysis including security, dependencies, code quality, and documentation

## Installation:

```bash
brew tap codcod/taps
brew install repos
```

## Configuration

The `config.yaml` file defines which repositories to manage and how to organize them.

```yaml
repositories:
  - name: loan-pricing
    url: git@github.com:yourorg/loan-pricing.git
    tags: [java, backend]
    branch: develop                  # Optional: Branch to clone
    path: cloned_repos/loan-pricing  # Optional: Directory to place cloned repo

  - name: web-ui
    url: git@github.com:yourorg/web-ui.git
    tags: [frontend, react]
    # When branch is not specified, the default branch will be cloned
    # When path is not specified, the current directory will be used
```
**Tip:**  
You can clone repositories first and use these to generate your `config.yaml`:

```sh
mkdir cloned_repos && cd "$_"
git clone http://github.com/codcod/jira-epic-timeline.git
git clone http://github.com/codcod/repos.git
repos init
```

## Typical session

Once you have a configuration file in place, an example session can look like
the following:

```sh
# Remove existing repositories
repos rm

# Clone java-based repositories in parallel
repos clone -t java -p

# Run command to update Backstage's catalog-info.yamls in all repos
repos run "fd 'catalog-info' -x sed -i '' 's/jira\/project-key: FOO/jira\/project-key: BAR/g' {}"

# Validate changes to see if updates were applied properly
rg 'jira/project-key: BAR' .

# Make sure that old entries were replaced
rg 'jira/project-key: FOO' .

# Create pull requests for all changes
repos pr --title "Update catalog-info.yaml" --body "Change jira/project-key to BAR"
```

See [Example commands](#example-commands) for more examples of commands to run.

## Usage

### Repository management

To configure, clone and remove repositories:

```sh
# Scan current directory for git repositories
repos init

# Create a different output file
repos init -o my-repos-config.yaml

# Overwrite existing config file
repos init --overwrite

# Clone all repositories
repos clone

# Clone only repositories with tag "java"
repos clone -t java

# Clone in parallel
repos clone -p

# Use a custom config file
repos clone -c custom-config.yaml

# Remove cloned repositories
repos rm

# Remove only repositories with tag "java"
repos rm -t java

# Remove in parallel
repos rm -p
```

### Running commands

To run arbitrary commands in repositories:

```sh
# Run a command in all repositories
repos run "mvn clean compile"

# Run a command only in repositories with tag "java"
repos run -t java "mvn clean compile"

# Run in parallel
repos run -p "npm install"

# Specify a custom log directory
repos run -l custom/logs "make build"
```

#### Example commands

Example commands to run with `repos run ""`:

```sh
# Count the number of lines
find . -type f |wc -l

# Compile projects (consider using --parallel flag to execute)
mvn clean compile

# Update catalog-info.yaml files
fd 'catalog-info' -x sed -i '' 's/jira\/project-key: FOO/jira\/project-key: BAR/g' {}

# Create a report of the changes made in the previous month
git log --all --author='$(id -un)' --since='1 month ago' --pretty=format:'%h %an %ad %s' --date=short
```

### Creating Pull Requests

To submit changes made in the cloned repositories:

```sh
export GITHUB_TOKEN=your_github_token

# Create PRs for repositories with changes
repos pr --title "My changes" --body "Description of changes"

# Create PRs with specific branch name
repos pr --branch feature/my-changes --title "My changes"

# Create draft pull requests
repos pr --draft

# Create PRs for specific repositories
repos pr -t backend
```

### Repository Health Analysis

Analyze the health and maintenance status of your repositories using two available methods:

#### Simple Health Checks (Legacy)

#### Advanced Health Check Analysis (Recommended)

Use the health engine for comprehensive, configurable code analysis:

```sh
# Run with advanced configuration
repos health --config examples/advanced-config-sample.yaml

# Dry run to see what would be executed
repos health --config examples/advanced-config-sample.yaml --dry-run
```

Both health analysis methods provide comprehensive checks including:
- **Git**: Repository status and commit activity
- **Dependencies**: Package management and outdated dependencies
- **Security**: Vulnerabilities and security policies
- **Code Quality**: Cyclomatic complexity analysis
- **Documentation**: README quality and completeness
- **Compliance**: License files and legal requirements
- **Automation**: CI/CD configuration

The health-based approach offers additional benefits:
- **Feature Flags**: Control which components are enabled
- **Parallel Execution**: Improved performance
- **Advanced Configuration**: YAML-based configuration with inheritance

#### Cyclomatic Complexity Analysis

Generate detailed function-level cyclomatic complexity reports and comprehensive code analysis using the health engine:

```sh
# Use health checks for comprehensive analysis
repos health --config examples/advanced-config-sample.yaml --timeout 60

# Run only cyclomatic complexity analysis
repos health cyclomatic-complexity --max-complexity 10

# Run with verbose output for detailed analysis results
repos health --config examples/advanced-config-sample.yaml --verbose
```

#### Analysis Features

The health engine provides:
- **Multi-language support**: Go, Python, Java, JavaScript analysis
- **Security scanning**: Vulnerability detection and dependency analysis
- **Quality metrics**: Code quality and maintainability assessment
- **Advanced configuration**: Flexible YAML-based configuration system

**Flake8-style Format** (`--complexity-detailed`):
- One violation per line (similar to flake8 output)
- Format: `file:line:column: C901 'function_name' is too complex (complexity)`
- Easy to integrate with CI/CD systems and linters

**Note**: Use the `repos health cyclomatic-complexity` subcommand to generate **only** the complexity analysis and skip all other health checks for faster execution. To combine complexity reporting with other health checks, use the main `repos health` command with specific categories using `--category`.

The complexity report provides:
- **Function-level analysis**: Individual function complexity scores
- **Threshold filtering**: Only shows functions exceeding the specified complexity limit
- **Grouped output**: Functions organized by repository and file
- **Multi-language support**: Currently supports Go (more languages coming soon)
- **Actionable insights**: Identifies specific functions that may benefit from refactoring

**Example output:**
```
=== Cyclomatic Complexity Report (Threshold: 10) ===

Repository: loan-pricing
  internal/calculator/complex_pricing.go:
    CalculateLoanTerms: 12
    ProcessRiskAssessment: 15
    
  pkg/validation/validator.go:
    ValidateCompoundRules: 11

Repository: web-ui
  No functions exceed complexity threshold of 10

Total functions analyzed: 87
Functions above threshold: 3
```

**Flake8-style output example:**
```
src/timeline/timeline.py:77:1: C901 'calculate_epic_timeline' is too complex (12)
src/timeline/timeline.py:144:1: C901 'display_results' is too complex (13)
src/calculator/pricing.py:25:1: C901 'calculate_complex_pricing' is too complex (15)

📊 Total violations: 3 (threshold: 10)
```

### Documentation

- [Development Guide](docs/01_development.md) - Setup, building, and commit conventions
- [Testing Guide](docs/04_testing.md) - Comprehensive testing strategy and best practices
- [Scripts Documentation](scripts/README.md) - Utility scripts for development

## Alternatives

The following are the alternatives to `repos`:

* [gita](https://github.com/nosarthur/gita): A tool to manage multiple Git
repositories.
* [gr](http://mixu.net/gr): Another multi-repo management tool.
* [meta](https://github.com/mateodelnorte/meta): Helps in managing multiple
repositories.
* [mu-repo](https://fabioz.github.io/mu-repo): For managing many repositories.
* [myrepos](https://myrepos.branchable.com): A tool to manage multiple
repositories.
* [repo](https://android.googlesource.com/tools/repo): A repository management
tool often used for Android source code.

## License

MIT
