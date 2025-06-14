# Scripts

This directory contains utility scripts for development and maintenance.

## Available Scripts

### `setup-commitlint.sh`

Sets up Git hooks and configuration for commit message validation using commitlint.

**Usage:**
```bash
./scripts/setup-commitlint.sh
# or
make setup-commitlint
```

**What it does:**
- Creates a `commit-msg` Git hook to validate commit messages
- Sets up a commit message template with format guidelines  
- Configures Git to use the `.gitmessage` template

**Requirements:**
- Node.js and npm (for commit message validation to work)
- Git repository

**More Info:**
- See [Development Guide](../docs/01_development.md#commit-messages) for complete commit message guidelines

### `test-release.sh`

Script for testing semantic release functionality.

**Usage:**
```bash
./scripts/test-release.sh
# or
npm run release:test
```

**More Info:**
- See [Semantic Release Migration](../docs/02_semantic_release.md) for complete release automation details
