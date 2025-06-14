// Package github provides functionality to create pull requests on GitHub
// and manage repository interactions.
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
	// Determine repository directory
	repoDir := util.GetRepoDir(repo)

	// Check if directory exists
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("repository directory does not exist: %s", repoDir)
	}

	// Execute within repository directory
	if err := executeInRepoDir(repoDir, func() error {
		return processPullRequest(repo, options)
	}); err != nil {
		return err
	}

	return nil
}

// executeInRepoDir executes a function within the repository directory
func executeInRepoDir(repoDir string, fn func() error) error {
	logger := util.NewLogger()

	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			logger.Error(config.Repository{Name: "unknown"}, "Failed to change back to original directory: %v", err)
		}
	}()

	if err := os.Chdir(repoDir); err != nil {
		return fmt.Errorf("failed to change to repository directory: %w", err)
	}

	return fn()
}

// processPullRequest handles the main PR creation logic
func processPullRequest(repo config.Repository, options PROptions) error {
	// Create changes unless "create only" mode is enabled
	if !options.CreateOnly {
		if err := createAndPushChanges(options); err != nil {
			return err
		}
	}

	// Extract owner and repo name from URL
	owner, repoName, err := util.ExtractOwnerAndRepo(repo.URL)
	if err != nil {
		return fmt.Errorf("failed to extract owner and repo: %w", err)
	}

	// Determine base branch
	baseBranch := determineBaseBranch(options.BaseBranch)

	// Create the PR
	return createGitHubPullRequest(owner, repoName, options, baseBranch)
}

// createAndPushChanges handles git operations for creating and pushing changes
func createAndPushChanges(options PROptions) error {
	// Check for changes
	hasChanges, err := git.HasChanges(".")
	if err != nil {
		return fmt.Errorf("failed to check for changes: %w", err)
	}

	if !hasChanges {
		return fmt.Errorf("no changes detected in repository")
	}

	// Create a new branch if one wasn't specified
	branchName := options.BranchName
	if branchName == "" {
		branchName = fmt.Sprintf("automated-changes-%d", os.Getpid())
		options.BranchName = branchName
	}

	// Create and checkout the branch
	if err := git.CreateAndCheckoutBranch(".", branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// Add all changes
	if err := git.AddAllChanges("."); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Commit the changes
	commitMsg := getCommitMessage(options)
	if err := git.CommitChanges(".", commitMsg); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Push the branch
	if err := git.PushBranch(".", branchName); err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	return nil
}

// getCommitMessage determines the commit message to use
func getCommitMessage(options PROptions) string {
	if options.CommitMsg != "" {
		return options.CommitMsg
	}
	if options.Title != "" {
		return options.Title
	}
	return "Automated changes"
}

// determineBaseBranch determines which base branch to use
func determineBaseBranch(baseBranch string) string {
	if baseBranch != "" {
		return baseBranch
	}

	// Default to 'main' if not specified
	defaultBranch := "main"

	// Check if 'main' exists, otherwise use 'master'
	if !git.BranchExists(".", defaultBranch) && git.BranchExists(".", "master") {
		return "master"
	}

	return defaultBranch
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
