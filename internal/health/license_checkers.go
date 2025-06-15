package health

import (
	"github.com/codcod/repos/internal/config"
)

// LicenseChecker checks for license files
type LicenseChecker struct{}

func (c *LicenseChecker) Name() string     { return "License" }
func (c *LicenseChecker) Category() string { return "compliance" }

func (c *LicenseChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)

	// Check for license files (both LICENSE and international spelling variations are valid)
	licenseFiles := []string{"LICENSE", "LICENSE.txt", "LICENSE.md", "LICENCE", "LICENCE.txt", "LICENCE.md"} //nolint:misspell // international spellings

	found := fileExistsInPath(repoPath, licenseFiles)
	if len(found) > 0 {
		return createHealthCheck(
			c.Name(),
			"license",
			HealthStatusHealthy,
			"License file found",
			"Repository has proper licensing",
			0,
		)
	}

	return createHealthCheck(
		c.Name(),
		"license",
		HealthStatusWarning,
		"No license file found",
		"Repository should include a LICENSE file for legal clarity",
		1,
	)
}
