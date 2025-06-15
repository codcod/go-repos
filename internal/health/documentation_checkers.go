package health

import (
	"github.com/codcod/repos/internal/config"
)

// DocumentationChecker checks for documentation files
type DocumentationChecker struct{}

func (c *DocumentationChecker) Name() string     { return "Documentation" }
func (c *DocumentationChecker) Category() string { return "documentation" }

func (c *DocumentationChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)

	// Check for documentation files
	docFiles := []string{
		"README.md", "README.txt", "README",
		"docs/", "doc/", "documentation/",
		"CONTRIBUTING.md", "CONTRIBUTING.txt",
		"CHANGELOG.md", "CHANGELOG.txt", "HISTORY.md",
	}

	found := fileExistsInPath(repoPath, docFiles)
	if len(found) > 0 {
		return createHealthCheck(
			c.Name(),
			"documentation",
			HealthStatusHealthy,
			"Documentation found",
			"Repository has proper documentation",
			0,
		)
	}

	return createHealthCheck(
		c.Name(),
		"documentation",
		HealthStatusWarning,
		"No documentation found",
		"Repository should include documentation files like README.md",
		1,
	)
}
