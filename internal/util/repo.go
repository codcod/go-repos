package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codcod/repos/internal/config"
	"github.com/fatih/color"
)

// GetRepoDir returns the local directory for a repository
func GetRepoDir(repo config.Repository) string {
	if repo.Path != "" {
		return repo.Path
	}
	dir := filepath.Base(repo.URL)
	return strings.TrimSuffix(dir, ".git")
}

// IsGitRepository checks if the given directory is a git repository
func IsGitRepository(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// ExtractOwnerAndRepo extracts the owner and repository name from a GitHub URL
func ExtractOwnerAndRepo(url string) (owner string, repo string, err error) {
	// Handle SSH URLs: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		path := strings.TrimPrefix(url, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid GitHub URL format: %s", url)
		}
		return parts[0], parts[1], nil
	}

	// Handle HTTPS URLs: https://github.com/owner/repo.git
	if strings.HasPrefix(url, "https://github.com/") {
		path := strings.TrimPrefix(url, "https://github.com/")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid GitHub URL format: %s", url)
		}
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unsupported URL format: %s", url)
}

// ColoredRepoName returns the repository name formatted with the specified color
func ColoredRepoName(repo config.Repository, c *color.Color) string {
	return c.Sprint(repo.Name)
}

// EnsureDirectoryExists ensures that a directory exists, creating it if necessary
func EnsureDirectoryExists(path string) error {
	return os.MkdirAll(path, 0750)
}
