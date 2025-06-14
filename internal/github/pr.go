package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/git"
	"github.com/codcod/repos/internal/util"
)

// PROptions configures how the pull request is created
type PROptions struct {
	Title      string
	Body       string
	BaseBranch string
	BranchName string
	CommitMsg  string
	Draft      bool
	Token      string // GitHub API token
	CreateOnly bool   // Only create PR, don't make changes
}

// CreatePullRequest creates a PR for changes in the repository
func CreatePullRequest(repo config.Repository, options PROptions) error {
	logger := util.NewLogger()

	// Determine repository directory
	repoDir := util.GetRepoDir(repo)

	// Check if directory exists
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("repository directory does not exist: %s", repoDir)
	}

	// Change to repository directory for Git operations
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer func() {
		// Ensure we always change back to the original directory
		if err := os.Chdir(originalDir); err != nil {
			logger.Error(repo, "Failed to change back to original directory: %v", err)
		}
	}()

	if err := os.Chdir(repoDir); err != nil {
		return fmt.Errorf("failed to change to repository directory: %w", err)
	}

	// Create changes unless "create only" mode is enabled
	if !options.CreateOnly {
		// Check for changes
		hasChanges, err := git.HasChanges(".") // Use "." instead of repoDir since we've already changed to that dir
		if err != nil {
			return fmt.Errorf("failed to check for changes: %w", err)
		}

		if !hasChanges {
			return fmt.Errorf("no changes detected in repository")
		}

		// Create a new branch if one wasn't specified
		if options.BranchName == "" {
			options.BranchName = fmt.Sprintf("automated-changes-%d", os.Getpid())
		}

		// Create and checkout the branch
		if err := git.CreateAndCheckoutBranch(".", options.BranchName); err != nil { // Use "." instead of repoDir
			return fmt.Errorf("failed to create branch: %w", err)
		}

		// Add all changes
		if err := git.AddAllChanges("."); err != nil { // Use "." instead of repoDir
			return fmt.Errorf("failed to add changes: %w", err)
		}

		// Commit the changes
		commitMsg := options.CommitMsg
		if commitMsg == "" {
			commitMsg = options.Title
		}
		if commitMsg == "" {
			commitMsg = "Automated changes"
		}

		if err := git.CommitChanges(".", commitMsg); err != nil { // Use "." instead of repoDir
			return fmt.Errorf("failed to commit changes: %w", err)
		}

		// Push the branch
		if err := git.PushBranch(".", options.BranchName); err != nil { // Use "." instead of repoDir
			return fmt.Errorf("failed to push branch: %w", err)
		}
	}

	// Extract owner and repo name from URL
	owner, repoName, err := util.ExtractOwnerAndRepo(repo.URL)
	if err != nil {
		return fmt.Errorf("failed to extract owner and repo: %w", err)
	}

	// Determine base branch if not specified
	baseBranch := options.BaseBranch
	if baseBranch == "" {
		// Default to 'main' if not specified
		baseBranch = "main"

		// Check if 'main' exists, otherwise use 'master'
		if !git.BranchExists(".", baseBranch) && git.BranchExists(".", "master") { // Use "." instead of repoDir
			baseBranch = "master"
		}
	}

	// Create the PR
	return createGitHubPullRequest(owner, repoName, options, baseBranch)
}

// createGitHubPullRequestFunc is the function type for creating GitHub pull requests
type createGitHubPullRequestFunc func(owner, repo string, options PROptions, baseBranch string) error

// createGitHubPullRequest is a variable that can be overridden for testing
var createGitHubPullRequest createGitHubPullRequestFunc = createGitHubPullRequestImpl

// createGitHubPullRequestImpl creates a pull request via the GitHub API
func createGitHubPullRequestImpl(owner, repo string, options PROptions, baseBranch string) error {
	// GitHub API endpoint
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)

	// Check if token is provided
	if options.Token == "" {
		options.Token = os.Getenv("GITHUB_TOKEN")
		if options.Token == "" {
			return fmt.Errorf("GitHub token not provided and GITHUB_TOKEN environment variable not set")
		}
	}

	// Create PR request body
	data := map[string]interface{}{
		"title": options.Title,
		"body":  options.Body,
		"head":  options.BranchName,
		"base":  baseBranch,
		"draft": options.Draft,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+options.Token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response
	if resp.StatusCode != http.StatusCreated {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return fmt.Errorf("failed to create PR, status: %d", resp.StatusCode)
		}
		return fmt.Errorf("failed to create PR: %v", errorResponse)
	}

	// Decode response
	var prResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&prResponse); err != nil {
		return err
	}

	// Print PR URL
	fmt.Printf("Pull request created: %s\n", prResponse["html_url"])
	return nil
}
