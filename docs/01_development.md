## Development

> **Quick Start**: New to the project? Start with the [README](../README.md) for an overview, then return here for development setup.

### Prerequisites

- Go 1.24 or later
- Node.js and npm (for commit message validation)
- Git

### Setup

1. **Clone the repository**:
   ```sh
   git clone https://github.com/codcod/repos.git
   cd repos
   ```

2. **Install development tools**:
   ```sh
   make install-tools
   ```

3. **Install commitlint dependencies**:
   ```sh
   make install-commitlint
   ```

### Build and Test

```sh
# Build the application
make build

# Run all tests
make test-all

# Run linter
make lint

# Run pre-commit checks
make pre-commit
```

### Commit Messages

This project uses [commitlint](https://commitlint.js.org/) to enforce consistent
commit messages following [Conventional Commits](https://www.conventionalcommits.org/).
Please format your commit messages as follows:

```
<type>(<scope>): <subject>
```

**Examples:**
- `feat(api): add user authentication endpoint`
- `fix(ui): resolve navigation menu overflow issue`
- `fix(git): handle repositories without remote origin`
- `docs: update installation instructions`
- `chore: upgrade dependencies`

Valid types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`, `build`, `revert`

### Commit types

- **feat**: A new feature (triggers a minor release)
- **fix**: A bug fix (triggers a patch release)
- **docs**: Documentation only changes
- **style**: Changes that do not affect the meaning of the code
- **refactor**: A code change that neither fixes a bug nor adds a feature
- **perf**: A code change that improves performance
- **test**: Adding missing tests or correcting existing tests
- **chore**: Changes to the build process or auxiliary tools

### Breaking Changes

To trigger a major release, include `BREAKING CHANGE:` in the commit footer or add `!` after the type:

```
feat!: remove deprecated API endpoint

BREAKING CHANGE: The /v1/old-endpoint has been removed. Use /v2/new-endpoint instead.
```

### New release (manual - deprecated)

> **Note**: This project now uses automated semantic releases. See [Semantic Release Migration](02_semantic_release.md) for details.

Releases are now done automatically via semantic-release. The following commands were used previously for manual releases:

```sh
git tag v1.0.x
git push origin v1.0.x

# To create a release go to github.com or use gh tool:
gh release create v1.0.x --title "Repos v1.0.x" --notes "Release notes for version 1.0.x"
```

## Related Documentation

- [Testing Guide](04_testing.md) - Comprehensive testing strategy and practices
- [Semantic Release Migration](02_semantic_release.md) - Automated release process
- [Code Improvements](05_code_improvements.md) - Recent maintainability enhancements
- [Scripts Documentation](../scripts/README.md) - Utility scripts reference