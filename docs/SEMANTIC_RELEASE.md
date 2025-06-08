# Semantic Release Migration Guide

This document explains the migration from our homegrown semantic release solution to the official [semantic-release](https://semantic-release.gitbook.io/) project.

## What Changed

### Before (Homegrown Solution)
- Custom GitHub Actions workflow that parsed commit messages manually
- Manual version bumping logic based on conventional commits
- Manual generation of release notes
- Hardcoded version information in `cmd/repos/main.go`
- Custom logic for determining version bumps (major/minor/patch)

### After (semantic-release)
- Industry-standard semantic-release tooling
- Automatic version determination based on conventional commits
- Automatic changelog generation
- Environment variable-based version injection
- Standardized plugin ecosystem
- Better error handling and dry-run capabilities

## New Files Added

### Node.js Configuration
- `package.json` - Contains semantic-release dependencies
- `.releaserc.json` - Semantic-release configuration
- `CHANGELOG.md` - Auto-generated changelog (managed by semantic-release)

### Scripts and Documentation
- `scripts/test-release.sh` - Local testing script
- `docs/SEMANTIC_RELEASE.md` - This documentation

### Updated Files
- `.github/workflows/release.yaml` - Simplified workflow using semantic-release
- `cmd/repos/main.go` - Version info now uses environment variables
- `Makefile` - Added support for version injection via build flags
- `README.md` - Added commit convention documentation
- `.gitignore` - Added `node_modules/`

## How It Works

### 1. Commit Analysis
semantic-release analyzes commit messages to determine the type of release:

- `fix:` → Patch release (0.0.1)
- `feat:` → Minor release (0.1.0)
- `BREAKING CHANGE:` or `feat!:` → Major release (1.0.0)

### 2. Version Management
Instead of hardcoded versions, the application now uses environment variables:

```go
// Old approach
version = "0.2.1"
commit  = "f51155d"
date    = "2025-06-08"

// New approach
version = getEnvOrDefault("VERSION", "dev")
commit  = getEnvOrDefault("COMMIT", "unknown")
date    = getEnvOrDefault("BUILD_DATE", "unknown")
```

### 3. Build Process
The Makefile now supports version injection:

```bash
# Development build
make build

# Release build (done by semantic-release)
VERSION=1.0.0 COMMIT=abc123 BUILD_DATE=2024-12-19 make build
```

### 4. Release Process
The GitHub Actions workflow now:

1. Sets up Node.js and Go
2. Installs semantic-release dependencies
3. Runs tests
4. Executes `npx semantic-release`

semantic-release then:
1. Analyzes commits since last release
2. Determines new version number
3. Generates changelog
4. Builds the binary with version info
5. Creates GitHub release with binary attachment
6. Commits changelog back to repository

## Plugin Configuration

Our `.releaserc.json` uses these plugins:

- **@semantic-release/commit-analyzer**: Analyzes commits to determine release type
- **@semantic-release/release-notes-generator**: Generates release notes
- **@semantic-release/changelog**: Maintains CHANGELOG.md
- **@semantic-release/exec**: Builds Go binary with version injection
- **@semantic-release/github**: Creates GitHub releases and uploads assets
- **@semantic-release/git**: Commits changelog back to repository

## Commit Message Convention

We now follow [Conventional Commits](https://www.conventionalcommits.org/):

### Basic Format
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Examples
```bash
# Patch release
fix(cli): handle repositories without remote origin

# Minor release  
feat(pr): add support for draft pull requests

# Major release
feat!: remove deprecated init command

BREAKING CHANGE: The init command has been removed. Use 'repos scan' instead.
```

### Commit Types
- `feat`: New feature (minor release)
- `fix`: Bug fix (patch release)
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Test additions/changes
- `chore`: Build process/tooling changes

## Testing Locally

Use the provided test script:

```bash
./scripts/test-release.sh
```

This script:
- Validates the configuration
- Tests Go builds
- Checks version injection
- Validates recent commit messages

## Debugging

### Dry Run
Test semantic-release without making changes:

```bash
npx semantic-release --dry-run
```

Note: This will fail locally due to GitHub authentication, but it validates the configuration.

### Environment Variables
Test version injection manually:

```bash
VERSION=1.0.0 COMMIT=abc123 BUILD_DATE=2024-12-19 make build
./bin/repos version
```

### Check Configuration
Validate the semantic-release config:

```bash
npx semantic-release --dry-run --no-ci 2>&1 | grep -E "(Loaded plugin|✔|✘)"
```

## Migration Benefits

1. **Standardization**: Uses industry-standard tooling
2. **Reliability**: Battle-tested release automation
3. **Maintainability**: Less custom code to maintain
4. **Features**: Rich plugin ecosystem and better error handling
5. **Documentation**: Extensive community documentation
6. **Flexibility**: Easy to extend with additional plugins

## Troubleshooting

### Common Issues

1. **Invalid commit messages**: Ensure commits follow conventional format
2. **GitHub permissions**: Verify GITHUB_TOKEN has appropriate permissions
3. **Build failures**: Check that Go build succeeds before release
4. **Plugin errors**: Validate .releaserc.json syntax

### Recovery

If a release fails partway through:
1. Check the GitHub Actions logs
2. Fix the underlying issue
3. Push another commit (following conventional format)
4. The next push will trigger a new release attempt

## Future Enhancements

Potential improvements:
- Add `@semantic-release/npm` if we publish to npm registry
- Add commit message validation via commitlint
- Add automated dependency updates
- Add release notifications to Slack/Discord
- Add support for pre-releases and beta channels

## Resources

- [semantic-release documentation](https://semantic-release.gitbook.io/)
- [Conventional Commits specification](https://www.conventionalcommits.org/)
- [semantic-release plugins](https://semantic-release.gitbook.io/semantic-release/usage/plugins)
- [GitHub Actions documentation](https://docs.github.com/en/actions)