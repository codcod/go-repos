package health

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codcod/repos/internal/config"
	"github.com/codcod/repos/internal/health/patterns"
)

// DeprecatedComponentsChecker checks for usage of deprecated components, APIs, and patterns
type DeprecatedComponentsChecker struct{}

func (c *DeprecatedComponentsChecker) Name() string     { return "Deprecated Components" }
func (c *DeprecatedComponentsChecker) Category() string { return "quality" }

func (c *DeprecatedComponentsChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)

	issues := []string{}

	// Check for deprecated patterns based on project type
	if err := c.checkDeprecatedPatterns(repoPath, &issues); err != nil {
		return createHealthCheck(
			c.Name(),
			"quality",
			HealthStatusWarning,
			"Could not scan for deprecated components",
			err.Error(),
			1,
		)
	}

	if len(issues) > 0 {
		return createHealthCheck(
			c.Name(),
			"quality",
			HealthStatusWarning,
			"Deprecated components found",
			strings.Join(issues, "; "),
			2,
		)
	}

	return createHealthCheck(
		c.Name(),
		"quality",
		HealthStatusHealthy,
		"No deprecated components found",
		"No deprecated patterns or APIs detected",
		0,
	)
}

func (c *DeprecatedComponentsChecker) checkDeprecatedPatterns(repoPath string, issues *[]string) error {
	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files that can't be read
		}

		// Skip certain directories and binary files
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == "target" || name == ".vscode" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only check text files based on extension
		ext := strings.ToLower(filepath.Ext(path))
		supportedExts := map[string]bool{
			".go": true, ".js": true, ".ts": true, ".py": true, ".java": true,
			".rb": true, ".php": true, ".cs": true, ".cpp": true, ".c": true,
			".h": true, ".hpp": true, ".md": true, ".txt": true, ".yml": true, ".yaml": true,
		}

		if !supportedExts[ext] {
			return nil
		}

		// Check file content for deprecated patterns
		if err := c.scanFileForDeprecatedPatterns(path, ext, issues); err != nil {
			// Log error but don't fail the entire check
			return nil
		}

		return nil
	})
}

func (c *DeprecatedComponentsChecker) scanFileForDeprecatedPatterns(filePath, ext string, issues *[]string) error {
	file, err := os.Open(filePath) // #nosec G304 - filePath is validated by caller
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Ignore close error for read-only file operations
			_ = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Get deprecated patterns for this file type
		patterns := c.getDeprecatedPatternsForExtension(ext)

		for pattern, depPattern := range patterns {
			if strings.Contains(strings.ToLower(line), strings.ToLower(pattern)) {
				relPath, _ := filepath.Rel("/", filePath)
				*issues = append(*issues,
					"deprecated pattern '"+pattern+"' found in "+relPath+":"+
						strconv.Itoa(lineNum)+" - "+depPattern.Description)
			}
		}
	}

	return scanner.Err()
}

func (c *DeprecatedComponentsChecker) getDeprecatedPatternsForExtension(ext string) map[string]patterns.DeprecatedPattern {
	switch ext {
	case ".go":
		return patterns.GetGoDeprecatedPatterns()
	case ".js", ".ts":
		return patterns.GetJavaScriptDeprecatedPatterns()
	case ".py":
		return patterns.GetPythonDeprecatedPatterns()
	case ".java":
		return patterns.GetJavaDeprecatedPatterns()
	default:
		return make(map[string]patterns.DeprecatedPattern)
	}
}
