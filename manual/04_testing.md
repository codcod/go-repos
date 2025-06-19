# Testing Guide

This document describes the comprehensive testing strategy for the repos CLI tool, including unit tests, integration tests, and best practices for maintaining test quality.

## Overview

The testing suite includes:
- **Unit Tests**: Test individual components and functions
- **Integration Tests**: Test end-to-end CLI functionality
- **Benchmark Tests**: Performance testing for critical operations
- **Coverage Reports**: Ensure adequate test coverage

## Test Structure

```
repos/
├── internal/
│   ├── config/
│   │   └── config_test.go          # Config loading and filtering tests
│   ├── git/
│   │   ├── git_test.go             # Git operations tests
│   │   └── remove_test.go          # Repository removal tests
│   ├── github/
│   │   └── pr_test.go              # GitHub API and PR creation tests
│   ├── runner/
│   │   └── runner_test.go          # Command execution tests
│   └── util/
│       ├── logging_test.go         # Logging functionality tests
│       ├── repo_test.go            # Repository utilities tests
│       └── scan_test.go            # Repository scanning tests
├── cmd/repos/
│   └── main_test.go                # Main package and CLI tests
└── integration_test.go             # End-to-end integration tests
```

## Running Tests

### All Tests
```bash
make test-all
```

### Unit Tests Only
```bash
make test-unit
# or
go test -v ./internal/... ./cmd/...
```

### Integration Tests
```bash
make test-integration
# or
go test -v -tags=integration .
```

### Coverage Report
```bash
make test-coverage
```
This generates `coverage.html` with a visual coverage report.

### Benchmarks
```bash
make test-bench
# or
go test -v -bench=. -benchmem ./...
```

### Quick Test (Skip Long-Running Tests)
```bash
go test -short ./...
```

## Test Categories

### Unit Tests

#### Config Package Tests (`internal/config/config_test.go`)
- **TestLoadConfig**: Configuration file loading and parsing
- **TestFilterRepositoriesByTag**: Repository filtering by tags
- **TestRepositoryFields**: Validation of repository field parsing
- **BenchmarkFilterRepositoriesByTag**: Performance testing for filtering

#### Git Package Tests (`internal/git/git_test.go`, `internal/git/remove_test.go`)
- **TestHasChanges**: Git status checking
- **TestBranchExists**: Branch existence validation
- **TestCreateAndCheckoutBranch**: Branch creation and checkout
- **TestCommitChanges**: Git commit operations
- **TestRemoveRepository**: Repository deletion with various scenarios
- **TestRemoveRepositoryWithNestedFiles**: Complex directory structure removal

#### GitHub Package Tests (`internal/github/pr_test.go`)
- **TestCreatePullRequest**: PR creation with various options
- **TestCreateGitHubPullRequest**: GitHub API interaction testing
- Mock HTTP servers for API testing without real GitHub calls

#### Runner Package Tests (`internal/runner/runner_test.go`)
- **TestRunCommand**: Command execution in repositories
- **TestPrepareLogFile**: Log file creation and formatting
- **TestOutputProcessor**: Real-time output processing

#### Util Package Tests
- **Logging Tests** (`logging_test.go`): Colored output and formatting
- **Repository Tests** (`repo_test.go`): Path utilities and URL parsing
- **Scanning Tests** (`scan_test.go`): Git repository discovery

#### Main Package Tests (`cmd/repos/main_test.go`)
- **TestGetEnvOrDefault**: Environment variable handling
- **TestVersionVariables**: Version information management

### Integration Tests (`integration_test.go`)

End-to-end testing of the complete CLI:
- **TestCLIVersion**: Version command functionality
- **TestCLIHelp**: Help system testing
- **TestCLIInitCommand**: Repository discovery and config generation
- **TestCLIRunCommandWithConfig**: Command execution with configuration
- **TestCLIParallelExecution**: Parallel processing verification
- **TestCLILogging**: Log file generation and content
- **TestCLIErrorHandling**: Error scenarios and recovery

## Test Patterns and Best Practices

### Temporary Directories
All tests use `t.TempDir()` for isolated test environments:
```go
func TestSomething(t *testing.T) {
    tmpDir := t.TempDir()
    // tmpDir is automatically cleaned up after test
}
```

### Git Repository Mocking
Tests create minimal git repositories for testing:
```go
gitDir := filepath.Join(tmpDir, ".git")
os.MkdirAll(gitDir, 0755)
createGitConfig(t, gitDir, "git@github.com:owner/repo.git")
```

### HTTP Server Mocking
GitHub API tests use `httptest.NewServer()`:
```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Mock GitHub API responses
}))
defer server.Close()
```

### Environment Variable Testing
Tests properly clean up environment variables:
```go
originalValue := os.Getenv("TEST_VAR")
defer func() {
    if originalValue != "" {
        os.Setenv("TEST_VAR", originalValue)
    } else {
        os.Unsetenv("TEST_VAR")
    }
}()
```

### Error Testing
All error cases are thoroughly tested:
```go
err := SomeFunction()
if err == nil {
    t.Error("Expected error but got none")
}
if !strings.Contains(err.Error(), "expected message") {
    t.Errorf("Error should contain 'expected message', got: %v", err)
}
```

## Benchmarking

Performance-critical functions include benchmarks:
- `BenchmarkFilterRepositoriesByTag`: Repository filtering performance
- `BenchmarkGetRepoDir`: Path resolution performance
- `BenchmarkHasChanges`: Git status checking performance
- `BenchmarkRunCommand`: Command execution overhead

Run benchmarks with memory profiling:
```bash
go test -bench=. -benchmem -benchtime=5s ./...
```

## Coverage Requirements

Target coverage levels:
- **Overall**: >85%
- **Critical paths**: >95% (config loading, git operations, command execution)
- **Error handling**: 100% (all error paths must be tested)

### Viewing Coverage
1. Generate coverage report: `make test-coverage`
2. Open `coverage.html` in a browser
3. Review uncovered lines and add tests as needed

## CI/CD Integration

### GitHub Actions
The release workflow includes:
1. Unit test execution
2. Coverage report generation
3. Integration test execution
4. Benchmark execution
5. Coverage upload to Codecov

### Local Development
Before committing:
```bash
make test-all
make lint
```

## Test Data Management

### Configuration Files
Tests use YAML strings for configuration:
```go
configYAML := `repositories:
  - name: test-repo
    url: git@github.com:owner/test-repo.git
    tags: [go, backend]`
```

### Mock Git Repositories
Helper functions create realistic git repository structures:
```go
func createGitConfig(t testing.TB, gitDir, remoteURL string) {
    gitConfig := fmt.Sprintf(`[remote "origin"]
    url = %s`, remoteURL)
    os.WriteFile(filepath.Join(gitDir, "config"), []byte(gitConfig), 0644)
}
```

## Troubleshooting Tests

### Common Issues

1. **Git Not Available**
   ```
   Tests that require git will be skipped with:
   t.Skip("git not available, skipping test")
   ```

2. **Symlink Support**
   ```
   Platform-specific tests check for symlink support:
   if err := os.Symlink(src, dst); err != nil {
       t.Skipf("Symlinks not supported: %v", err)
   }
   ```

3. **Integration Test Failures**
   ```
   Ensure the binary is built before running integration tests:
   make build
   ```

### Debugging Test Failures

1. **Increase Verbosity**
   ```bash
   go test -v ./...
   ```

2. **Run Specific Test**
   ```bash
   go test -v -run TestSpecificFunction ./internal/config
   ```

3. **Enable Race Detection**
   ```bash
   go test -race ./...
   ```

4. **Profile Tests**
   ```bash
   go test -cpuprofile=cpu.prof -memprofile=mem.prof ./...
   ```

## Adding New Tests

### For New Features
1. Add unit tests for the core functionality
2. Add integration tests for CLI behavior
3. Add benchmarks for performance-critical code
4. Update this documentation

### Test Naming Convention
- `TestFunctionName`: Basic functionality
- `TestFunctionNameError`: Error cases
- `TestFunctionNameEdgeCases`: Edge cases and boundary conditions
- `BenchmarkFunctionName`: Performance tests

### Test Structure Template
```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name        string
        input       InputType
        expected    ExpectedType
        expectError bool
    }{
        {
            name:        "valid input",
            input:       validInput,
            expected:    expectedOutput,
            expectError: false,
        },
        {
            name:        "invalid input",
            input:       invalidInput,
            expected:    zeroValue,
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := NewFeature(tt.input)
            
            if tt.expectError {
                if err == nil {
                    t.Error("Expected error but got none")
                }
                return
            }
            
            if err != nil {
                t.Errorf("Unexpected error: %v", err)
                return
            }
            
            if result != tt.expected {
                t.Errorf("Expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

## Resources

- [Go Testing Documentation](https://golang.org/doc/tutorial/add-a-test)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testing with Temporary Files](https://golang.org/pkg/testing/#T.TempDir)
- [HTTP Testing](https://golang.org/pkg/net/http/httptest/)
- [Benchmarking](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [Development Guide](01_development.md) - For commit message validation setup

## Commit Message Validation

This project uses [commitlint](https://commitlint.js.org/) to enforce consistent commit message formatting following the [Conventional Commits](https://www.conventionalcommits.org/) specification.

> **Note**: For complete commit message guidelines and setup instructions, see the [Development Guide](01_development.md#commit-messages).

### Quick Reference

**Format**: `<type>(<scope>): <subject>`

**Common types**: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`

**Setup**: 
```bash
make install-commitlint
make setup-commitlint
```