package ci

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codcod/repos/internal/checkers/base"
	"github.com/codcod/repos/internal/core"
)

// CIConfigChecker checks for CI/CD configuration
type CIConfigChecker struct {
	*base.BaseChecker
}

// NewCIConfigChecker creates a new CI configuration checker
func NewCIConfigChecker() *CIConfigChecker {
	config := core.CheckerConfig{
		Enabled:    true,
		Severity:   "medium",
		Timeout:    30 * time.Second,
		Categories: []string{"ci", "automation"},
	}

	return &CIConfigChecker{
		BaseChecker: base.NewBaseChecker(
			"ci-config",
			"CI/CD Configuration",
			"ci",
			config,
		),
	}
}

// Check performs the CI configuration check
func (c *CIConfigChecker) Check(ctx context.Context, repoCtx core.RepositoryContext) (core.CheckResult, error) {
	return c.Execute(ctx, repoCtx, func() (core.CheckResult, error) {
		return c.checkCIConfig(repoCtx)
	})
}

// checkCIConfig performs the actual CI configuration check
func (c *CIConfigChecker) checkCIConfig(repoCtx core.RepositoryContext) (core.CheckResult, error) {
	builder := base.NewResultBuilder(c.ID(), c.Name(), c.Category())

	// Look for CI configuration files
	ciConfigs := c.findCIConfigs(repoCtx.Repository.Path)
	builder.AddMetric("ci_configs_found", len(ciConfigs))

	if len(ciConfigs) == 0 {
		builder.WithStatus(core.StatusWarning)
		builder.WithScore(30, 100)
		builder.AddIssue(base.NewIssueWithSuggestion(
			"no_ci_config",
			core.SeverityMedium,
			"No CI/CD configuration found",
			"Add CI/CD configuration (e.g., GitHub Actions, Travis CI, Jenkins) to automate testing and deployment",
		))
		return builder.Build(), nil
	}

	// Analyze CI configurations
	score, issues, warnings := c.analyzeCIConfigs(repoCtx.Repository.Path, ciConfigs)
	builder.WithScore(score, 100)

	// Add found configs to metrics
	for i, config := range ciConfigs {
		builder.AddMetric(fmt.Sprintf("ci_config_%d", i), config.Path)
		builder.AddMetric(fmt.Sprintf("ci_type_%d", i), config.Type)
	}

	// Add issues and warnings
	for _, issue := range issues {
		builder.AddIssue(issue)
	}
	for _, warning := range warnings {
		builder.AddWarning(warning)
	}

	// Set overall status
	if score >= 80 {
		builder.WithStatus(core.StatusHealthy)
	} else if score >= 50 {
		builder.WithStatus(core.StatusWarning)
	} else {
		builder.WithStatus(core.StatusCritical)
	}

	return builder.Build(), nil
}

// CIConfig represents a CI configuration file
type CIConfig struct {
	Path string
	Type string
}

// findCIConfigs finds CI configuration files in the repository
func (c *CIConfigChecker) findCIConfigs(repoPath string) []CIConfig {
	var configs []CIConfig

	// GitHub Actions
	githubWorkflowsPath := filepath.Join(repoPath, ".github", "workflows")
	if entries, err := os.ReadDir(githubWorkflowsPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yml") || strings.HasSuffix(entry.Name(), ".yaml")) {
				configs = append(configs, CIConfig{
					Path: filepath.Join(".github", "workflows", entry.Name()),
					Type: "GitHub Actions",
				})
			}
		}
	}

	// Travis CI
	travisFiles := []string{".travis.yml", ".travis.yaml"}
	for _, file := range travisFiles {
		if _, err := os.Stat(filepath.Join(repoPath, file)); err == nil {
			configs = append(configs, CIConfig{
				Path: file,
				Type: "Travis CI",
			})
		}
	}

	// CircleCI
	circleciFiles := []string{".circleci/config.yml", ".circleci/config.yaml"}
	for _, file := range circleciFiles {
		if _, err := os.Stat(filepath.Join(repoPath, file)); err == nil {
			configs = append(configs, CIConfig{
				Path: file,
				Type: "CircleCI",
			})
		}
	}

	// GitLab CI
	gitlabFiles := []string{".gitlab-ci.yml", ".gitlab-ci.yaml"}
	for _, file := range gitlabFiles {
		if _, err := os.Stat(filepath.Join(repoPath, file)); err == nil {
			configs = append(configs, CIConfig{
				Path: file,
				Type: "GitLab CI",
			})
		}
	}

	// Jenkins
	jenkinsFiles := []string{"Jenkinsfile", "jenkins.yml", "jenkins.yaml"}
	for _, file := range jenkinsFiles {
		if _, err := os.Stat(filepath.Join(repoPath, file)); err == nil {
			configs = append(configs, CIConfig{
				Path: file,
				Type: "Jenkins",
			})
		}
	}

	// Azure Pipelines
	azureFiles := []string{"azure-pipelines.yml", "azure-pipelines.yaml", ".azure/pipelines.yml"}
	for _, file := range azureFiles {
		if _, err := os.Stat(filepath.Join(repoPath, file)); err == nil {
			configs = append(configs, CIConfig{
				Path: file,
				Type: "Azure Pipelines",
			})
		}
	}

	// Buildkite
	if _, err := os.Stat(filepath.Join(repoPath, ".buildkite")); err == nil {
		configs = append(configs, CIConfig{
			Path: ".buildkite",
			Type: "Buildkite",
		})
	}

	return configs
}

// analyzeCIConfigs analyzes the quality of CI configurations
func (c *CIConfigChecker) analyzeCIConfigs(repoPath string, configs []CIConfig) (int, []core.Issue, []core.Warning) {
	var issues []core.Issue
	var warnings []core.Warning
	score := 40 // Base score for having CI

	// Bonus for multiple CI providers (redundancy)
	if len(configs) > 1 {
		score += 10
	}

	// Analyze each configuration
	for _, config := range configs {
		configScore, configIssues, configWarnings := c.analyzeIndividualConfig(repoPath, config)
		score += configScore / len(configs) // Average the scores

		issues = append(issues, configIssues...)
		warnings = append(warnings, configWarnings...)
	}

	// Check for common best practices
	bestPracticesScore, bestPracticesIssues, bestPracticesWarnings := c.checkCIBestPractices(repoPath, configs)
	score += bestPracticesScore

	issues = append(issues, bestPracticesIssues...)
	warnings = append(warnings, bestPracticesWarnings...)

	// Ensure score is within bounds
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score, issues, warnings
}

// analyzeIndividualConfig analyzes a single CI configuration file
func (c *CIConfigChecker) analyzeIndividualConfig(repoPath string, config CIConfig) (int, []core.Issue, []core.Warning) {
	var issues []core.Issue
	var warnings []core.Warning
	score := 0

	configPath := filepath.Join(repoPath, config.Path)
	content, err := os.ReadFile(configPath)
	if err != nil {
		issues = append(issues, base.NewIssueWithSuggestion(
			"ci_config_read_error",
			core.SeverityMedium,
			fmt.Sprintf("Unable to read CI config %s: %v", config.Path, err),
			"Check file permissions and ensure the CI configuration file is accessible",
		))
		return 0, issues, warnings
	}

	contentStr := strings.ToLower(string(content))

	// Check for basic CI features
	features := c.checkCIFeatures(contentStr)
	score += len(features) * 5 // 5 points per feature

	// Check for testing
	if c.hasTestingConfig(contentStr) {
		score += 20
	} else {
		warnings = append(warnings, core.Warning{
			Type:    "no_testing_in_ci",
			Message: fmt.Sprintf("CI configuration %s lacks testing steps", config.Path),
		})
	}

	// Check for build steps
	if c.hasBuildConfig(contentStr) {
		score += 15
	} else {
		warnings = append(warnings, core.Warning{
			Type:    "no_build_in_ci",
			Message: fmt.Sprintf("CI configuration %s lacks build steps", config.Path),
		})
	}

	// Check for deployment
	if c.hasDeploymentConfig(contentStr) {
		score += 10
	}

	// Check for common issues
	if len(content) < 50 {
		issues = append(issues, base.NewIssueWithSuggestion(
			"ci_config_too_short",
			core.SeverityMedium,
			fmt.Sprintf("CI configuration %s is very short and may be incomplete", config.Path),
			"Ensure the CI configuration includes necessary steps for testing and building",
		))
	}

	return score, issues, warnings
}

// checkCIFeatures checks for common CI features in the configuration
func (c *CIConfigChecker) checkCIFeatures(content string) []string {
	var features []string

	featureKeywords := map[string][]string{
		"caching":               {"cache", "cached"},
		"matrix_builds":         {"matrix", "strategy"},
		"parallel_jobs":         {"parallel", "concurrent"},
		"artifacts":             {"artifact", "upload"},
		"notifications":         {"notify", "notification", "slack", "email"},
		"environment_variables": {"env:", "environment"},
		"secrets":               {"secret", "encrypted"},
		"docker":                {"docker", "container", "image"},
	}

	for feature, keywords := range featureKeywords {
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				features = append(features, feature)
				break
			}
		}
	}

	return features
}

// hasTestingConfig checks if the CI configuration includes testing
func (c *CIConfigChecker) hasTestingConfig(content string) bool {
	testKeywords := []string{"test", "spec", "check", "verify", "coverage", "junit", "pytest", "jest", "mocha"}
	for _, keyword := range testKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// hasBuildConfig checks if the CI configuration includes build steps
func (c *CIConfigChecker) hasBuildConfig(content string) bool {
	buildKeywords := []string{"build", "compile", "make", "gradle", "maven", "npm run build", "go build", "cargo build"}
	for _, keyword := range buildKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// hasDeploymentConfig checks if the CI configuration includes deployment
func (c *CIConfigChecker) hasDeploymentConfig(content string) bool {
	deployKeywords := []string{"deploy", "release", "publish", "docker push", "helm", "kubectl"}
	for _, keyword := range deployKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// checkCIBestPractices checks for CI/CD best practices
func (c *CIConfigChecker) checkCIBestPractices(repoPath string, configs []CIConfig) (int, []core.Issue, []core.Warning) {
	var issues []core.Issue
	var warnings []core.Warning
	score := 0

	// Check for branch protection integration
	hasMainBranchCI := c.hasMainBranchCI(repoPath, configs)
	if hasMainBranchCI {
		score += 10
	} else {
		warnings = append(warnings, core.Warning{
			Type:    "no_main_branch_ci",
			Message: "CI doesn't appear to run on main branch protection",
		})
	}

	// Check for PR/MR checks
	hasPRChecks := c.hasPRChecks(repoPath, configs)
	if hasPRChecks {
		score += 10
	} else {
		warnings = append(warnings, core.Warning{
			Type:    "no_pr_checks",
			Message: "CI doesn't appear to run on pull requests",
		})
	}

	// Check for status checks
	hasStatusChecks := c.hasStatusChecks(repoPath, configs)
	if hasStatusChecks {
		score += 5
	}

	return score, issues, warnings
}

// hasMainBranchCI checks if CI runs on the main branch
func (c *CIConfigChecker) hasMainBranchCI(repoPath string, configs []CIConfig) bool {
	for _, config := range configs {
		configPath := filepath.Join(repoPath, config.Path)
		if content, err := os.ReadFile(configPath); err == nil {
			contentStr := strings.ToLower(string(content))
			if strings.Contains(contentStr, "main") || strings.Contains(contentStr, "master") {
				return true
			}
		}
	}
	return false
}

// hasPRChecks checks if CI runs on pull requests
func (c *CIConfigChecker) hasPRChecks(repoPath string, configs []CIConfig) bool {
	for _, config := range configs {
		configPath := filepath.Join(repoPath, config.Path)
		if content, err := os.ReadFile(configPath); err == nil {
			contentStr := strings.ToLower(string(content))
			prKeywords := []string{"pull_request", "merge_request", "pr:", "mr:"}
			for _, keyword := range prKeywords {
				if strings.Contains(contentStr, keyword) {
					return true
				}
			}
		}
	}
	return false
}

// hasStatusChecks checks if CI provides status checks
func (c *CIConfigChecker) hasStatusChecks(repoPath string, configs []CIConfig) bool {
	for _, config := range configs {
		configPath := filepath.Join(repoPath, config.Path)
		if content, err := os.ReadFile(configPath); err == nil {
			contentStr := strings.ToLower(string(content))
			statusKeywords := []string{"status", "check", "badge"}
			for _, keyword := range statusKeywords {
				if strings.Contains(contentStr, keyword) {
					return true
				}
			}
		}
	}
	return false
}

// SupportsRepository checks if this checker supports the repository
func (c *CIConfigChecker) SupportsRepository(repo core.Repository) bool {
	// This checker supports all repositories
	return true
}
