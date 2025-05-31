# Repos

A CLI tool to manage multiple GitHub repositories: clone them and run arbitrary commands in each, with colored output and logging.

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

## Typical session

Example session with this tool can look like the following

```sh
# Clone a repo and use it as a starting point to adjust config.yaml
mkdir cloned_repos
cd cloned_repos
git clone http://github.com/myself/repo.git
repos init

# Clone java-based repositories in parallel
repos clone -t java -p

# Update Backstage catalog-info.yamls in all repos
repos run "fd 'catalog-info' -x sed -i '' 's/jira\/project-key: FOO/jira\/project-key: BAR/g' {}"

# Validate changes
rg 'jira/project-key: BAR' .

# Make sure that old entries were replaced
rg 'jira/project-key: FOO' .

# Create pull requests for all changes
repos pr --title "Update catalog-info.yaml" --body "Change jira/project-key to BAR"
```

See [Example commands](#example-commands) for more examples.

## Usage

### Repository Management

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

### Running Commands

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

```sh
# Create PRs for repositories with changes
export GITHUB_TOKEN=your_github_token
repos pr --title "My changes" --body "Description of changes"

# Create PRs with specific branch name
repos pr --branch feature/my-changes --title "My changes"

# Create draft pull requests
repos pr --draft

# Create PRs for specific repositories
repos pr -t backend
```

## Configuration

Create a `config.yaml` file in the root directory:

```yaml
repositories:
  - name: loan-pricing
    url: git@github.com:yourorg/loan-pricing.git
    tags: [java, backend]
    branch: develop  # Optional: Branch to clone

  - name: web-ui
    url: git@github.com:yourorg/web-ui.git
    tags: [frontend, react]
    # When branch is not specified, the default branch will be cloned
```

## Development

```sh
go mod tidy
go build -o repos ./cmd/repos
```

When ready for a new release:

```sh
git tag v1.0.x
git push origin v1.0.x

# To create a release go to github.com or use gh tool:
gh release create v1.0.x --title "Repos v1.0.x" --notes "Release notes for version 1.0.x"
```

## Alternatives

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
