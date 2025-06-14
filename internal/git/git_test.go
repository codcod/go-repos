package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codcod/repos/internal/config"
)

func TestHasChanges(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir := t.TempDir()

	// Initialize git repo
	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Configure git user for commits
	_ = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Test clean repository
	hasChanges, err := HasChanges(tmpDir)
	if err != nil {
		t.Fatalf("HasChanges() error = %v", err)
	}
	if hasChanges {
		t.Error("HasChanges() should return false for clean repo")
	}

	// Add a file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test untracked files
	hasChanges, err = HasChanges(tmpDir)
	if err != nil {
		t.Fatalf("HasChanges() error = %v", err)
	}
	if !hasChanges {
		t.Error("HasChanges() should return true for untracked files")
	}

	// Stage the file
	if err := exec.Command("git", "-C", tmpDir, "add", "test.txt").Run(); err != nil {
		t.Fatalf("Failed to stage file: %v", err)
	}

	// Test staged changes
	hasChanges, err = HasChanges(tmpDir)
	if err != nil {
		t.Fatalf("HasChanges() error = %v", err)
	}
	if !hasChanges {
		t.Error("HasChanges() should return true for staged files")
	}

	// Commit the file
	if err := exec.Command("git", "-C", tmpDir, "commit", "-m", "initial commit").Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Test clean repo after commit
	hasChanges, err = HasChanges(tmpDir)
	if err != nil {
		t.Fatalf("HasChanges() error = %v", err)
	}
	if hasChanges {
		t.Error("HasChanges() should return false for clean repo after commit")
	}

	// Modify the file
	err = os.WriteFile(testFile, []byte("modified content"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Test modified files
	hasChanges, err = HasChanges(tmpDir)
	if err != nil {
		t.Fatalf("HasChanges() error = %v", err)
	}
	if !hasChanges {
		t.Error("HasChanges() should return true for modified files")
	}
}

func TestHasChangesInvalidRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Test non-git directory
	_, err := HasChanges(tmpDir)
	if err == nil {
		t.Error("HasChanges() should return error for non-git directory")
	}
}

func TestBranchExists(t *testing.T) {
	// Create a temporary git repository
	tmpDir := t.TempDir()

	// Initialize git repo
	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Configure git user
	_ = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)
	_ = exec.Command("git", "-C", tmpDir, "add", "test.txt").Run()
	_ = exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	// Test existing branch (main or master)
	if BranchExists(tmpDir, "main") {
		// main branch exists
	} else if BranchExists(tmpDir, "master") {
		// master branch exists
	} else {
		t.Error("Neither main nor master branch exists")
	}

	// Test non-existent branch
	if BranchExists(tmpDir, "non-existent-branch") {
		t.Error("BranchExists() should return false for non-existent branch")
	}

	// Create a new branch
	if err := exec.Command("git", "-C", tmpDir, "checkout", "-b", "feature-branch").Run(); err != nil {
		t.Fatalf("Failed to create branch: %v", err)
	}

	// Test newly created branch
	if !BranchExists(tmpDir, "feature-branch") {
		t.Error("BranchExists() should return true for newly created branch")
	}
}

func TestCreateAndCheckoutBranch(t *testing.T) {
	// Create a temporary git repository
	tmpDir := t.TempDir()

	// Initialize git repo
	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Configure git user
	_ = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)
	_ = exec.Command("git", "-C", tmpDir, "add", "test.txt").Run()
	_ = exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	// Test creating new branch
	branchName := "test-branch"
	err := CreateAndCheckoutBranch(tmpDir, branchName)
	if err != nil {
		t.Fatalf("CreateAndCheckoutBranch() error = %v", err)
	}

	// Verify branch was created and checked out
	if !BranchExists(tmpDir, branchName) {
		t.Error("Branch should exist after CreateAndCheckoutBranch()")
	}

	// Verify we're on the new branch
	output, err := exec.Command("git", "-C", tmpDir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}
	currentBranch := strings.TrimSpace(string(output))
	if currentBranch != branchName {
		t.Errorf("Should be on branch %s, but on %s", branchName, currentBranch)
	}
}

func TestAddAllChanges(t *testing.T) {
	// Create a temporary git repository
	tmpDir := t.TempDir()

	// Initialize git repo
	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Configure git user
	_ = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create test files
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	testFile2 := filepath.Join(tmpDir, "test2.txt")
	_ = os.WriteFile(testFile1, []byte("test1"), 0644)
	_ = os.WriteFile(testFile2, []byte("test2"), 0644)

	// Add all changes
	err := AddAllChanges(tmpDir)
	if err != nil {
		t.Fatalf("AddAllChanges() error = %v", err)
	}

	// Verify files were staged
	output, err := exec.Command("git", "-C", tmpDir, "status", "--porcelain").Output()
	if err != nil {
		t.Fatalf("Failed to get git status: %v", err)
	}

	status := string(output)
	if !strings.Contains(status, "A  test1.txt") {
		t.Error("test1.txt should be staged")
	}
	if !strings.Contains(status, "A  test2.txt") {
		t.Error("test2.txt should be staged")
	}
}

func TestCommitChanges(t *testing.T) {
	// Create a temporary git repository
	tmpDir := t.TempDir()

	// Initialize git repo
	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Configure git user
	_ = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create and stage a file
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)
	_ = exec.Command("git", "-C", tmpDir, "add", "test.txt").Run()

	// Test commit
	commitMsg := "Test commit message"
	err := CommitChanges(tmpDir, commitMsg)
	if err != nil {
		t.Fatalf("CommitChanges() error = %v", err)
	}

	// Verify commit was created
	output, err := exec.Command("git", "-C", tmpDir, "log", "--oneline", "-1").Output()
	if err != nil {
		t.Fatalf("Failed to get git log: %v", err)
	}

	logOutput := string(output)
	if !strings.Contains(logOutput, commitMsg) {
		t.Errorf("Commit message should contain '%s', got: %s", commitMsg, logOutput)
	}
}

func TestRunGitCommand(t *testing.T) {
	// Create a temporary git repository
	tmpDir := t.TempDir()

	// Initialize git repo
	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		t.Skip("git not available, skipping test")
	}

	// Test valid git command
	output, err := RunGitCommand(tmpDir, "status", "--porcelain")
	if err != nil {
		t.Fatalf("RunGitCommand() error = %v", err)
	}

	// Should return empty for clean repo
	if len(output) != 0 {
		t.Errorf("Expected empty output for clean repo, got: %s", string(output))
	}

	// Test invalid git command
	_, err = RunGitCommand(tmpDir, "invalid-command")
	if err == nil {
		t.Error("RunGitCommand() should return error for invalid command")
	}
}

func TestCloneRepositoryMocking(t *testing.T) {
	// Note: This test would require mocking the git command or using a test repository
	// For now, we'll test the basic structure and error cases

	tmpDir := t.TempDir()

	repo := config.Repository{
		Name: "test-repo",
		URL:  "https://github.com/nonexistent/repo.git",
		Path: filepath.Join(tmpDir, "test-repo"),
	}

	// This should fail because the repository doesn't exist
	err := CloneRepository(repo)
	if err == nil {
		t.Error("CloneRepository() should fail for non-existent repository")
	}
}

func TestCloneRepositoryExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory that already exists
	repoDir := filepath.Join(tmpDir, "existing-repo")
	err := os.MkdirAll(repoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	repo := config.Repository{
		Name: "existing-repo",
		URL:  "https://github.com/owner/repo.git",
		Path: repoDir,
	}

	// Should skip cloning if directory already exists
	err = CloneRepository(repo)
	if err != nil {
		t.Errorf("CloneRepository() should not error for existing directory, got: %v", err)
	}
}

func TestPushBranch(t *testing.T) {
	// This test requires a remote repository, so we'll test error cases
	tmpDir := t.TempDir()

	// Test with non-git directory
	err := PushBranch(tmpDir, "main")
	if err == nil {
		t.Error("PushBranch() should return error for non-git directory")
	}
}

func BenchmarkHasChanges(b *testing.B) {
	// Create a temporary git repository
	tmpDir := b.TempDir()

	// Initialize git repo
	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		b.Skip("git not available, skipping benchmark")
	}

	// Configure git user
	_ = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)
	_ = exec.Command("git", "-C", tmpDir, "add", "test.txt").Run()
	_ = exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HasChanges(tmpDir)
		if err != nil {
			b.Fatalf("HasChanges() error = %v", err)
		}
	}
}

func BenchmarkBranchExists(b *testing.B) {
	// Create a temporary git repository
	tmpDir := b.TempDir()

	// Initialize git repo
	if err := exec.Command("git", "init", tmpDir).Run(); err != nil {
		b.Skip("git not available, skipping benchmark")
	}

	// Configure git user and create initial commit
	_ = exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	_ = exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)
	_ = exec.Command("git", "-C", tmpDir, "add", "test.txt").Run()
	_ = exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BranchExists(tmpDir, "main")
	}
}
