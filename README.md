[![Release](https://img.shields.io/github/v/release/codcod/repos?style=flat-square)](https://github.com/codcod/repos/releases)

# Repos

A CLI tool to manage multiple GitHub repositories: clone them and run arbitrary
commands in each, with colored output and logging.

## Features

- Clone multiple repositories from a config file
- Filter repositories by tag
- Run commands in all or filtered repositories
- Create pull requests for changes
- Parallel execution support
- Real-time, colorized logs
- Per-repo log files

## Installation

```sh
brew tap codcod/taps
brew install repos
```

## Configuration

To work properly, `repos` requires information about Git repositories. To
provide it, create a `config.yaml` file in the root directory:

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
# Scan deeper directories (up to 5 levels)
repos init --depth 5

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

## License

MIT
