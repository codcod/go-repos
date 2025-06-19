// Package common provides shared utilities for CLI commands
package common

import (
	"fmt"
	"os"

	"github.com/codcod/repos/internal/config"
	"github.com/fatih/color"
)

// CLIError represents a CLI error with an exit code
type CLIError struct {
	Message  string
	ExitCode int
}

func (e *CLIError) Error() string {
	return e.Message
}

// ExitWithError prints an error message and exits with the specified code
func ExitWithError(msg string, code int) {
	color.Red("Error: %s", msg)
	os.Exit(code)
}

// ExitWithErrorf prints a formatted error message and exits with the specified code
func ExitWithErrorf(format string, args ...interface{}) {
	color.Red("Error: "+format, args...)
	os.Exit(1)
}

// PrintSuccess prints a success message in green
func PrintSuccess(format string, args ...interface{}) {
	color.Green(format, args...)
}

// PrintWarning prints a warning message in yellow
func PrintWarning(format string, args ...interface{}) {
	color.Yellow(format, args...)
}

// PrintInfo prints an info message in cyan
func PrintInfo(format string, args ...interface{}) {
	color.Cyan(format, args...)
}

// LoadConfig loads configuration and returns error
func LoadConfig(configFile string) (*config.Config, error) {
	return config.LoadConfig(configFile)
}

// LoadConfigOrExit loads configuration and exits on error
func LoadConfigOrExit(configFile string) *config.Config {
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		ExitWithErrorf("Failed to load config: %v", err)
	}
	return cfg
}

// FilterRepositoriesOrExit filters repositories by tag and exits if none found
func FilterRepositoriesOrExit(cfg *config.Config, tag string) []config.Repository {
	repositories := cfg.FilterRepositoriesByTag(tag)
	if len(repositories) == 0 {
		PrintWarning("No repositories found with tag: %s", tag)
		os.Exit(0)
	}
	return repositories
}

// ProcessRepos processes repositories with clean error handling
func ProcessRepos(repositories []config.Repository, parallel bool, processor func(config.Repository) error) error {
	if parallel {
		return processReposParallel(repositories, processor)
	}
	return processReposSequential(repositories, processor)
}

func processReposSequential(repositories []config.Repository, processor func(config.Repository) error) error {
	for _, repo := range repositories {
		if err := processor(repo); err != nil {
			return fmt.Errorf("processing %s: %w", repo.Name, err)
		}
	}
	return nil
}

func processReposParallel(repositories []config.Repository, processor func(config.Repository) error) error {
	// Implementation would be similar to main.go's processRepos function
	// For now, fall back to sequential processing
	return processReposSequential(repositories, processor)
}

// GetEnvOrDefault returns environment variable value or default
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
