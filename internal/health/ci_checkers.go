package health

import (
	"github.com/codcod/repos/internal/config"
)

// CIStatusChecker checks for CI/CD configuration
type CIStatusChecker struct{}

func (c *CIStatusChecker) Name() string     { return "CI Status" }
func (c *CIStatusChecker) Category() string { return "ci" }

func (c *CIStatusChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)

	// Check for various CI configuration files
	ciFiles := []string{
		".github/workflows",    // GitHub Actions
		".gitlab-ci.yml",       // GitLab CI
		"Jenkinsfile",          // Jenkins
		".travis.yml",          // Travis CI
		"azure-pipelines.yml",  // Azure DevOps
		".circleci/config.yml", // CircleCI
		"buildkite.yml",        // Buildkite
	}

	found := fileExistsInPath(repoPath, ciFiles)
	if len(found) > 0 {
		return createHealthCheck(
			c.Name(),
			"ci",
			HealthStatusHealthy,
			"CI/CD configuration found",
			"Repository has continuous integration setup",
			0,
		)
	}

	return createHealthCheck(
		c.Name(),
		"ci",
		HealthStatusWarning,
		"No CI/CD configuration found",
		"Repository should have continuous integration setup for automated testing and deployment",
		1,
	)
}
