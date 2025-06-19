// Package main provides the CLI entry point for the repos tool.
// All command logic has been modularized into the cmd/repos/commands package.
package main

import (
	"fmt"
	"os"

	"github.com/codcod/repos/cmd/repos/commands/clone"
	"github.com/codcod/repos/cmd/repos/commands/health"
	initcmd "github.com/codcod/repos/cmd/repos/commands/init"
	"github.com/codcod/repos/cmd/repos/commands/pr"
	"github.com/codcod/repos/cmd/repos/commands/rm"
	"github.com/codcod/repos/cmd/repos/commands/run"

	"github.com/spf13/cobra"
)

var (
	// Version information - will be set via build flags, with environment variable fallback
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// getEnvOrDefault returns the environment variable value or default if empty
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// init function to handle environment variable fallback for version info
func init() {
	// Use environment variables as fallback when build-time flags weren't set
	if version == "dev" {
		version = getEnvOrDefault("VERSION", version)
	}
	if commit == "unknown" {
		commit = getEnvOrDefault("COMMIT", commit)
	}
	if date == "unknown" {
		date = getEnvOrDefault("BUILD_DATE", date)
	}
}

var rootCmd = &cobra.Command{
	Use:   "repos",
	Short: "A tool to manage multiple GitHub repositories",
	Long:  `Clone multiple GitHub repositories and run arbitrary commands inside them.`,
}

func init() {
	// Set up persistent flags that are available to all commands
	rootCmd.PersistentFlags().StringP("config", "c", "config.yaml", "config file path")
	rootCmd.PersistentFlags().StringP("tag", "t", "", "filter repositories by tag")
	rootCmd.PersistentFlags().BoolP("parallel", "p", false, "execute operations in parallel")

	// Register all modular commands
	rootCmd.AddCommand(clone.NewCommand())
	rootCmd.AddCommand(run.NewCommand())
	rootCmd.AddCommand(pr.NewCommand())
	rootCmd.AddCommand(rm.NewCommand())
	rootCmd.AddCommand(initcmd.NewCommand())
	rootCmd.AddCommand(health.NewCommand())

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("repos %s (%s) built on %s\n", version, commit, date)
		},
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
