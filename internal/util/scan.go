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

		// Look for .git directory first
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			
			// Calculate repository depth (the depth of the repository directory itself)
			repoRelPath, err := filepath.Rel(rootPath, repoPath)
			if err != nil {
				return err
			}
			
			repoDepth := 0
			if repoRelPath != "." {
				repoDepth = strings.Count(repoRelPath, string(os.PathSeparator)) + 1
			}
			
			// Only include repository if it's within the max depth
			if maxDepth > 0 && repoDepth > maxDepth {
				return filepath.SkipDir
			}
			
			repoName := filepath.Base(repoPath)

			// Get remote URL
			gitRemoteURL, _ := GetRemoteURL(repoPath)

			// Skip if no remote URL found - but don't skip the directory, just continue
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

		// Calculate current depth relative to root for non-.git directories
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		
		// Count directory separators to determine depth
		depth := 0
		if relPath != "." {
			depth = strings.Count(relPath, string(os.PathSeparator)) + 1
		}

		// Skip if we've exceeded max depth
		if maxDepth > 0 && depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
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
