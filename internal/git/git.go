package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/util"
)

// CloneRepository clones a repository
func CloneRepository(repo config.Repository) error {
	logger := util.NewLogger()

	// Determine target directory
	targetDir := util.GetRepoDir(repo)

	// Check if directory already exists
	if _, err := os.Stat(targetDir); err == nil {
		logger.Warn(repo, "Repository directory already exists, skipping")
		return nil
	}

	// Clone the repository
	args := []string{"clone"}

	// Add branch flag if a branch is specified
	if repo.Branch != "" {
		args = append(args, "-b", repo.Branch)
	}

	// Add repository URL and target directory
	args = append(args, repo.URL, targetDir)

	cmd := exec.Command("git", args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Print which branch is being cloned
	if repo.Branch != "" {
		logger.Info(repo, "Cloning branch '%s' from %s", repo.Branch, repo.URL)
	} else {
		logger.Info(repo, "Cloning default branch from %s", repo.URL)
	}

	err := cmd.Run()

	if stdoutBuf.Len() > 0 {
		fmt.Printf("%s | %s", repo.Name, stdoutBuf.String())
	}

	if stderrBuf.Len() > 0 {
		logger.Error(repo, "%s", stderrBuf.String())
	}

	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Only log success if we actually cloned
	logger.Success(repo, "Successfully cloned")
	return nil
}

// RunGitCommand runs a git command in the repository directory
func RunGitCommand(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Output()
}

// HasChanges checks if the repository has any uncommitted changes
func HasChanges(dir string) (bool, error) {
	output, err := RunGitCommand(dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return len(output) > 0, nil
}

// BranchExists checks if a branch exists in the repository
func BranchExists(dir string, branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	cmd.Dir = dir
	return cmd.Run() == nil
}

// CreateAndCheckoutBranch creates and checks out a new branch
func CreateAndCheckoutBranch(dir string, branch string) error {
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = dir
	return cmd.Run()
}

// AddAllChanges stages all changes in the repository
func AddAllChanges(dir string) error {
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = dir
	return cmd.Run()
}

// CommitChanges commits staged changes with the given message
func CommitChanges(dir string, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	return cmd.Run()
}

// PushBranch pushes a branch to the remote repository
func PushBranch(dir string, branch string) error {
	cmd := exec.Command("git", "push", "--set-upstream", "origin", branch)
	cmd.Dir = dir
	return cmd.Run()
}
