// Package health provides health command subcommands
package health

import (
	"fmt"
	"os"

	"github.com/codcod/repos/cmd/repos/common"
	"github.com/spf13/cobra"
)

// NewGenConfigCommand creates the genconfig subcommand
func NewGenConfigCommand() *cobra.Command {
	var outputFile string
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "genconfig",
		Short: "Generate a comprehensive health configuration template",
		Long:  `Generate a comprehensive health configuration template with all available options and detailed comments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateConfigTemplate(outputFile, overwrite)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&outputFile, "output", "o", "health-config.yaml", "Output file for the configuration template")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing configuration file")

	return cmd
}

// generateConfigTemplate generates a comprehensive configuration template
func generateConfigTemplate(outputFile string, overwrite bool) error {
	common.PrintInfo("🔧 Generating comprehensive health configuration template...")
	fmt.Println()

	template := `# Health Configuration Template
# This is a comprehensive template showing all available health check options.
# Uncomment and modify sections as needed for your project.

# Global health check settings
health:
  # Timeout for individual health checks (e.g., "30s", "1m", "2m30s")
  timeout: "30s"
  
  # Run health checks in parallel (true/false)
  parallel: false
  
  # Verbose output during health checks (true/false)
  verbose: false
  
  # Global categories to run (empty means all categories)
  # Available categories: ci, documentation, git, security, dependencies, compliance
  categories: []
  
  # Complexity analysis settings
  complexity:
    # Enable complexity analysis (true/false)
    enabled: false
    
    # Maximum allowed cyclomatic complexity (0 = no limit)
    max_complexity: 15
    
    # Report complexity even if within limits (true/false)
    report_all: true

  # Health checker configurations
  checkers:
    # CI/CD Configuration checks
    ci:
      enabled: true
      severity: "medium"
      config:
        # Check for common CI files (.github/workflows, .gitlab-ci.yml, etc.)
        check_ci_files: true
        # Validate CI configuration syntax
        validate_syntax: false

    # Documentation checks  
    documentation:
      enabled: true
      severity: "medium"
      config:
        # Require README.md file
        require_readme: true
        # Minimum README content length
        min_readme_length: 100
        # Check for common documentation sections
        required_sections: ["Description", "Installation", "Usage"]

    # Git repository checks
    git:
      enabled: true
      severity: "medium"  
      config:
        # Check for uncommitted changes
        check_clean_status: true
        # Check last commit recency (days)
        max_days_since_commit: 30
        # Validate commit message format
        validate_commit_messages: false

    # Security checks
    security:
      enabled: true
      severity: "high"
      config:
        # Check branch protection rules
        check_branch_protection: true
        # Scan for known vulnerabilities
        vulnerability_scan: true
        # Check for sensitive files
        check_sensitive_files: true
        # Patterns for sensitive content
        sensitive_patterns:
          - "password"
          - "secret"
          - "api_key"
          - "private_key"

    # Dependency checks
    dependencies:
      enabled: true
      severity: "medium"
      config:
        # Check for outdated dependencies
        check_outdated: true
        # Security vulnerability scanning
        security_scan: true
        # License compatibility checks
        license_check: false

    # Compliance checks
    compliance:
      enabled: true
      severity: "medium"
      config:
        # Check for required license file
        require_license: true
        # Accepted license types
        accepted_licenses: ["MIT", "Apache-2.0", "BSD-3-Clause"]
        # Check for code of conduct
        require_code_of_conduct: false

  # Code analyzer configurations
  analyzers:
    # JavaScript/TypeScript analyzer
    javascript:
      enabled: true
      extensions: [".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"]
      config:
        # ESLint integration
        use_eslint: true
        # Complexity analysis
        complexity_analysis: true

    # Go analyzer
    go:
      enabled: true
      extensions: [".go"]
      config:
        # Use go vet
        use_go_vet: true
        # Use golangci-lint if available
        use_golangci_lint: true
        # Complexity analysis
        complexity_analysis: true

    # Python analyzer
    python:
      enabled: true
      extensions: [".py"]
      config:
        # Use pylint if available
        use_pylint: false
        # Use flake8 if available  
        use_flake8: true
        # Complexity analysis
        complexity_analysis: true

    # Java analyzer
    java:
      enabled: true
      extensions: [".java"]
      config:
        # Use checkstyle if available
        use_checkstyle: false
        # Use PMD if available
        use_pmd: false
        # Complexity analysis
        complexity_analysis: true

# Repository filtering (inherited from main config if not specified)
# repositories:
#   - name: "my-repo"
#     url: "https://github.com/user/my-repo"
#     tags: ["backend", "critical"]
#   - name: "another-repo" 
#     url: "https://github.com/user/another-repo"
#     tags: ["frontend"]

# Example usage:
# 1. Save this template as 'health-config.yaml'
# 2. Uncomment and modify the sections you need
# 3. Run: repos health --config health-config.yaml
# 4. Use --category to filter by specific categories
# 5. Use --verbose for detailed output
# 6. Use --parallel for faster execution
# 7. Use --dry-run to preview what would be executed`

	// If no output file specified, print to console
	if outputFile == "" {
		fmt.Println(template)
		fmt.Println()

		common.PrintSuccess("📋 Comprehensive health configuration template generated!")
		fmt.Println()
		common.PrintInfo("💡 Usage Tips:")
		fmt.Println("  1. Copy the template above to a file (e.g., 'health-config.yaml')")
		fmt.Println("  2. Uncomment and modify sections as needed")
		fmt.Println("  3. Run: repos health --config health-config.yaml")
		fmt.Println("  4. Use --dry-run to preview configuration before execution")
		fmt.Println()
		common.PrintInfo("🔧 Quick Setup:")
		fmt.Println("  • repos health --list-categories      # See all available categories")
		fmt.Println("  • repos health --category git,security # Run specific categories")
		return nil
	}

	// Check if file exists and we're not overwriting
	if _, err := os.Stat(outputFile); err == nil && !overwrite {
		common.PrintWarning("Configuration file already exists: %s", outputFile)
		common.PrintInfo("Use --overwrite to replace it.")
		return nil
	}

	// Write template to file
	err := os.WriteFile(outputFile, []byte(template), 0600)
	if err != nil {
		return fmt.Errorf("failed to write configuration file: %v", err)
	}

	common.PrintSuccess("📋 Health configuration file created: %s", outputFile)
	common.PrintInfo("💡 Edit the file to customize health check settings.")

	return nil
}
