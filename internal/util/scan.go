package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/codcod/repos/internal/config"
)

// FindGitRepositories recursively scans a directory for git repositories
func FindGitRepositories(rootPath string, maxDepth int) ([]config.Repository, error) {
	var repos []config.Repository

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == rootPath {
			return nil
		}

		// Calculate current depth
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		depth := len(strings.Split(relPath, string(os.PathSeparator)))

		// Skip if we've exceeded max depth
		if maxDepth > 0 && depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Look for .git directory
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			repoName := filepath.Base(repoPath)

			// Get remote URL
			gitRemoteURL, _ := GetRemoteURL(repoPath)

			// Skip if no remote URL found
			if gitRemoteURL == "" {
				return nil
			}

			// Create repository entry
			repo := config.Repository{
				Name: repoName,
				URL:  gitRemoteURL,
				Path: repoPath,
				Tags: []string{"auto-discovered"},
			}

			repos = append(repos, repo)

			// Skip deeper traversal of this directory
			return filepath.SkipDir
		}
		return nil
	})

	return repos, err
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
