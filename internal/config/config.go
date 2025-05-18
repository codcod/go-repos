package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
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

// LoadConfig loads the configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// FilterRepositoriesByTag filters repositories by tag
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
