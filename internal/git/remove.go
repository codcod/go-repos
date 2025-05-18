package git

import (
	"fmt"
	"os"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/util"
)

// RemoveRepository removes a cloned repository
func RemoveRepository(repo config.Repository) error {
	// Determine repository directory
	repoDir := util.GetRepoDir(repo)

	// Check if directory exists
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("repository directory does not exist: %s", repoDir)
	}

	// Check if it's a git repository
	if !util.IsGitRepository(repoDir) {
		return fmt.Errorf("not a git repository: %s", repoDir)
	}

	// Remove the directory
	if err := os.RemoveAll(repoDir); err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}

	return nil
}
