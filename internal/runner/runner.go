// Package runner provides functionality to run commands in repositories
// and process their output in real-time.
package runner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/util"
	"github.com/fatih/color"
)

// OutputProcessor handles processing of command output
type OutputProcessor struct {
	RepoName  string
	LogFile   *os.File
	IsStderr  bool
	HeaderSet bool
}

// ProcessOutput reads from the given reader and processes the output
func (p *OutputProcessor) ProcessOutput(reader io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()

	scanner := bufio.NewScanner(reader)

	// Choose color based on output type
	var repoColor func(a ...interface{}) string
	if p.IsStderr {
		repoColor = color.New(color.FgRed, color.Bold).SprintFunc()
	} else {
		repoColor = color.New(color.FgCyan).SprintFunc()
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Print to stdout/stderr with colored repo name
		if p.IsStderr {
			fmt.Fprintf(os.Stderr, "%s | %s\n", repoColor(p.RepoName), line)
		} else {
			fmt.Printf("%s | %s\n", repoColor(p.RepoName), line)
		}

		// Save to log file if enabled
		if p.LogFile != nil {
			// Add stderr section header if needed
			if p.IsStderr && !p.HeaderSet {
				_, _ = p.LogFile.WriteString("\n=== STDERR ===\n")
				p.HeaderSet = true
			}

			_, _ = fmt.Fprintf(p.LogFile, "%s | %s\n", p.RepoName, line)
			_ = p.LogFile.Sync()
		}
	}
}

// PrepareLogFile creates and initializes a log file
func PrepareLogFile(repo config.Repository, logDir, command, repoDir string) (*os.File, string, error) {
	if logDir == "" {
		return nil, "", nil
	}

	if err := util.EnsureDirectoryExists(logDir); err != nil {
		return nil, "", fmt.Errorf("failed to create log directory: %w", err)
	}

	logFilePath := filepath.Join(logDir, fmt.Sprintf("%s_%s.log",
		repo.Name,
		time.Now().Format("20060102_150405")))

	// #nosec G304 - This is a legitimate log file creation in a controlled directory
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create log file: %w", err)
	}

	// Write header information
	_, _ = fmt.Fprintf(logFile, "Repository: %s\n", repo.Name)
	_, _ = fmt.Fprintf(logFile, "Command: %s\n", command)
	_, _ = fmt.Fprintf(logFile, "Directory: %s\n", repoDir)
	_, _ = fmt.Fprintf(logFile, "Timestamp: %s\n\n", time.Now().Format(time.RFC3339))
	_, _ = logFile.WriteString("=== STDOUT ===\n")

	return logFile, logFilePath, nil
}

// RunCommand runs a command in the repository directory
func RunCommand(repo config.Repository, command string, logDir string) error {
	logger := util.NewLogger()

	// Determine repository directory
	repoDir := util.GetRepoDir(repo)

	// Check if directory exists
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("repository directory does not exist: %s", repoDir)
	}

	// Prepare command - use shell to properly handle quotes and complex commands
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = repoDir

	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Prepare log file
	logFile, logFilePath, err := PrepareLogFile(repo, logDir, command, repoDir)
	if err != nil {
		return err
	}
	if logFile != nil {
		defer func() { _ = logFile.Close() }()
	}

	// Run the command
	logger.Info(repo, "Running '%s'", command)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Set up wait group to ensure both stdout and stderr are fully processed
	var wg sync.WaitGroup
	wg.Add(2)

	// Process stdout and stderr in real-time
	stdoutProcessor := &OutputProcessor{
		RepoName: repo.Name,
		LogFile:  logFile,
		IsStderr: false,
	}
	stderrProcessor := &OutputProcessor{
		RepoName: repo.Name,
		LogFile:  logFile,
		IsStderr: true,
	}

	go stdoutProcessor.ProcessOutput(stdoutPipe, &wg)
	go stderrProcessor.ProcessOutput(stderrPipe, &wg)

	// Wait for both stdout and stderr to be fully processed
	wg.Wait()

	// Wait for the command to complete
	err = cmd.Wait()

	if logFile != nil && err == nil {
		logger.Info(repo, "Log saved to %s", logFilePath)
	}

	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}
