package main

import (
	"fmt"
	"os"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/core"
	"gopkg.in/yaml.v3"
)

type simpleLogger struct{}

func (l *simpleLogger) Debug(msg string, fields ...core.Field) {
	fmt.Printf("[DEBUG] %s\n", msg)
}

func (l *simpleLogger) Info(msg string, fields ...core.Field) {
	fmt.Printf("[INFO] %s\n", msg)
}

func (l *simpleLogger) Warn(msg string, fields ...core.Field) {
	fmt.Printf("[WARN] %s\n", msg)
}

func (l *simpleLogger) Error(msg string, fields ...core.Field) {
	fmt.Printf("[ERROR] %s\n", msg)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: debug-config <config-file>")
		os.Exit(1)
	}

	configFile := os.Args[1]

	// Test format detection
	logger := &simpleLogger{}
	migrator := config.NewConfigMigrator(logger)

	// Read file manually
	data, err := os.ReadFile(configFile) //nolint:gosec // Config file path is from command line argument
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File size: %d bytes\n", len(data))

	// Try parsing as advanced config
	var advancedTest config.AdvancedConfig
	if err := yaml.Unmarshal(data, &advancedTest); err != nil {
		fmt.Printf("Error parsing as advanced config: %v\n", err)
	} else {
		fmt.Printf("Parsed successfully as advanced config\n")
		fmt.Printf("Version: '%s'\n", advancedTest.Version)
		fmt.Printf("FeatureFlags count: %d\n", len(advancedTest.FeatureFlags))
		fmt.Printf("Profiles count: %d\n", len(advancedTest.Profiles))
		fmt.Printf("Pipelines count: %d\n", len(advancedTest.Pipelines))
		fmt.Printf("Checkers count: %d\n", len(advancedTest.Checkers))
		fmt.Printf("Analyzers count: %d\n", len(advancedTest.Analyzers))
		fmt.Printf("Categories count: %d\n", len(advancedTest.Categories))
	}

	// Test using migrator
	format, err := migrator.DetectConfigFormat(configFile)
	if err != nil {
		fmt.Printf("Error detecting format: %v\n", err)
	} else {
		fmt.Printf("Detected format: %s\n", format)
	}
}
