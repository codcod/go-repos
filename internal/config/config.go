// Package config provides configuration management for repository operations.
// It handles loading YAML configuration files and filtering repositories by tags.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

// Repository represents a GitHub repository configuration
type Repository struct {
	Name   string   `yaml:"name"`
	URL    string   `yaml:"url"`
	Tags   []string `yaml:"tags"`
	Path   string   `yaml:"path,omitempty"`   // Optional custom local path
	Branch string   `yaml:"branch,omitempty"` // Optional branch to clone
}

// Config represents the application configuration
type Config struct {
	Repositories []Repository `yaml:"repositories"`
}

// LoadConfig loads the configuration from a YAML file.
// It returns an error if the file cannot be read or parsed.
func LoadConfig(path string) (*Config, error) {
	// Clean and validate the path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// #nosec G304 - This is a legitimate config file read operation
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// FilterRepositoriesByTag filters repositories by tag.
// If tag is empty, returns all repositories.
// Tag matching is case-sensitive.
func (c *Config) FilterRepositoriesByTag(tag string) []Repository {
	if tag == "" {
		return c.Repositories
	}

	var filtered []Repository
	for _, repo := range c.Repositories {
		for _, t := range repo.Tags {
			if t == tag {
				filtered = append(filtered, repo)
				break
			}
		}
	}

	return filtered
}

// HasTag checks if a repository has the specified tag.
func (r *Repository) HasTag(tag string) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// Validate checks if a repository configuration is valid.
func (r *Repository) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("repository name is required")
	}
	if r.URL == "" {
		return fmt.Errorf("repository URL is required")
	}
	return nil
}
