package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/codcod/repos/internal/config"
)

// FindGitRepositories scans a directory for git repositories (non-recursive)
func FindGitRepositories(rootPath string) ([]config.Repository, error) {
	var repos []config.Repository

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entryPath := filepath.Join(rootPath, entry.Name())

		// Check if this directory is a git repository
		if IsGitRepository(entryPath) {
			repoName := entry.Name()

			// Get remote URL
			gitRemoteURL, _ := GetRemoteURL(entryPath)

			// Skip if no remote URL found
			if gitRemoteURL == "" {
				continue
			}

			// Create repository entry
			repo := config.Repository{
				Name: repoName,
				URL:  gitRemoteURL,
				Path: entryPath,
				Tags: []string{"auto-discovered"},
			}

			repos = append(repos, repo)
		}
	}

	return repos, nil
}

// GetRemoteURL retrieves the origin remote URL of a git repository
func GetRemoteURL(repoPath string) (string, error) {
	gitConfigPath := filepath.Join(repoPath, ".git", "config")

	data, err := os.ReadFile(gitConfigPath)
	if err != nil {
		return "", err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	inOrigin := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[remote \"origin\"]" {
			inOrigin = true
			continue
		}

		if inOrigin && strings.HasPrefix(line, "url = ") {
			return strings.TrimPrefix(line, "url = "), nil
		}

		// We're out of the origin section
		if inOrigin && strings.HasPrefix(line, "[") {
			break
		}
	}

	return "", nil
}
