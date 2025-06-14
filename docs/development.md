## Development

```sh
go mod tidy
go build -o repos ./cmd/repos
```

## Commit Message Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/) for automated releases. Please format your commit messages as follows:

### Commit Types

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

### Examples

```
feat(cli): add support for custom log directory
fix(git): handle repositories without remote origin
docs: update installation instructions
chore: upgrade dependencies
```

### New release (old, manual)

Releases are done automatically and follow semantic versioning. Prior to this
automatization the following commands were used to release a version:

```sh
git tag v1.0.x
git push origin v1.0.x

# To create a release go to github.com or use gh tool:
gh release create v1.0.x --title "Repos v1.0.x" --notes "Release notes for version 1.0.x"
```