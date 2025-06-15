package health

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codcod/repos/internal/config"
)

// DependencyChecker checks for outdated dependencies
type DependencyChecker struct{}

func (c *DependencyChecker) Name() string     { return "Dependencies" }
func (c *DependencyChecker) Category() string { return "dependencies" }

func (c *DependencyChecker) Check(repo config.Repository) HealthCheck {
	repoPath := GetRepoPath(repo)
	foundFiles := c.findDependencyFiles(repoPath)

	if len(foundFiles) == 0 {
		return c.createDependencyHealthCheck(HealthStatusHealthy, "No dependency files found", "", 1)
	}

	return c.checkDependenciesByType(repoPath, foundFiles)
}

func (c *DependencyChecker) findDependencyFiles(repoPath string) []string {
	depFiles := []string{
		"go.mod", "package.json", "requirements.txt", "pyproject.toml",
		"Gemfile", "Cargo.toml", "pom.xml", "build.gradle", "build.gradle.kts",
	}
	var foundFiles []string

	for _, file := range depFiles {
		if _, err := os.Stat(filepath.Join(repoPath, file)); err == nil {
			foundFiles = append(foundFiles, file)
		}
	}
	return foundFiles
}

func (c *DependencyChecker) checkDependenciesByType(repoPath string, foundFiles []string) HealthCheck {
	if contains(foundFiles, "go.mod") {
		return c.checkGoMod(repoPath)
	}
	if contains(foundFiles, "package.json") {
		return c.checkPackageJSON(repoPath)
	}
	if contains(foundFiles, "pyproject.toml") {
		return c.checkPyprojectToml(repoPath)
	}
	if contains(foundFiles, "requirements.txt") {
		return c.checkRequirementsTxt(repoPath)
	}
	if contains(foundFiles, "pom.xml") {
		return c.checkMavenPom(repoPath)
	}
	if contains(foundFiles, "build.gradle") || contains(foundFiles, "build.gradle.kts") {
		return c.checkGradleBuild(repoPath)
	}

	return c.createDependencyHealthCheck(HealthStatusWarning,
		fmt.Sprintf("Found dependency files: %s", strings.Join(foundFiles, ", ")),
		"Dependency checking not implemented for this project type", 1)
}

func (c *DependencyChecker) createDependencyHealthCheck(status HealthStatus, message, details string, severity int) HealthCheck {
	return HealthCheck{
		Name:        c.Name(),
		Status:      status,
		Message:     message,
		Details:     details,
		Severity:    severity,
		Category:    c.Category(),
		LastChecked: time.Now(),
	}
}

func (c *DependencyChecker) checkGoMod(repoPath string) HealthCheck {
	// Check if go mod tidy would make changes
	cmd := exec.Command("go", "mod", "tidy", "-diff")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return HealthCheck{
			Name:        c.Name(),
			Status:      HealthStatusWarning,
			Message:     "Unable to check go.mod status",
			Details:     stderr.String(),
			Severity:    2,
			Category:    c.Category(),
			LastChecked: time.Now(),
		}
	}

	if stdout.Len() > 0 {
		return HealthCheck{
			Name:        c.Name(),
			Status:      HealthStatusWarning,
			Message:     "go.mod needs tidying",
			Details:     "Run 'go mod tidy' to fix",
			Severity:    1,
			Category:    c.Category(),
			LastChecked: time.Now(),
		}
	}

	return HealthCheck{
		Name:        c.Name(),
		Status:      HealthStatusHealthy,
		Message:     "Go dependencies are up to date",
		Severity:    1,
		Category:    c.Category(),
		LastChecked: time.Now(),
	}
}

func (c *DependencyChecker) checkPackageJSON(repoPath string) HealthCheck {
	// Basic package.json existence check
	packageFile := filepath.Join(repoPath, "package.json")
	lockFile := filepath.Join(repoPath, "package-lock.json")

	if _, err := os.Stat(packageFile); err != nil {
		return HealthCheck{
			Name:        c.Name(),
			Status:      HealthStatusWarning,
			Message:     "package.json not found",
			Severity:    2,
			Category:    c.Category(),
			LastChecked: time.Now(),
		}
	}

	status := HealthStatusHealthy
	message := "Node.js dependencies found"

	if _, err := os.Stat(lockFile); err != nil {
		status = HealthStatusWarning
		message = "package-lock.json missing"
	}

	return HealthCheck{
		Name:        c.Name(),
		Status:      status,
		Message:     message,
		Severity:    1,
		Category:    c.Category(),
		LastChecked: time.Now(),
	}
}

// checkMavenPom checks Maven dependencies and project health
func (c *DependencyChecker) checkMavenPom(repoPath string) HealthCheck {
	// Validate that pom.xml exists
	pomFile := filepath.Join(repoPath, "pom.xml")
	if _, err := os.Stat(pomFile); os.IsNotExist(err) {
		return c.createDependencyHealthCheck(HealthStatusCritical, "Maven POM validation failed",
			fmt.Sprintf("pom.xml not found in %s", repoPath), 3)
	}

	if !c.commandExists("mvn") {
		return c.createDependencyHealthCheck(HealthStatusWarning, "Maven not available",
			"Install Maven for detailed dependency analysis", 2)
	}

	var issues, warnings []string
	c.checkMavenDependencyAnalysis(repoPath, &issues, &warnings)
	c.checkMavenDependencyResolution(repoPath, &issues, &warnings)

	return c.createMavenHealthCheckResult(issues, warnings)
}

func (c *DependencyChecker) checkMavenDependencyAnalysis(repoPath string, issues, warnings *[]string) {
	ctx, cancel := CreateHealthContext()
	defer cancel()

	cmd := exec.CommandContext(ctx, "mvn", "dependency:analyze", "-q")
	cmd.Dir = repoPath
	output, err := cmd.Output()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			*warnings = append(*warnings, "Maven dependency check timed out")
		} else {
			*warnings = append(*warnings, "Unable to analyze dependencies")
		}
		return
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "Used undeclared dependencies") {
		*issues = append(*issues, "Has undeclared dependencies")
	}
	if strings.Contains(outputStr, "Unused declared dependencies") {
		*warnings = append(*warnings, "Has unused declared dependencies")
	}
}

func (c *DependencyChecker) checkMavenDependencyResolution(repoPath string, issues, warnings *[]string) {
	ctx, cancel := CreateHealthContext()
	defer cancel()

	cmd := exec.CommandContext(ctx, "mvn", "dependency:resolve", "-q")
	cmd.Dir = repoPath
	err := cmd.Run()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			*warnings = append(*warnings, "Maven dependency resolution timed out")
		} else {
			*issues = append(*issues, "Dependencies cannot be resolved")
		}
	}
}

func (c *DependencyChecker) createMavenHealthCheckResult(issues, warnings []string) HealthCheck {
	status := HealthStatusHealthy
	message := "Maven dependencies are healthy"
	severity := 1

	if len(issues) > 0 {
		status = HealthStatusCritical
		message = fmt.Sprintf("Maven dependency issues: %s", strings.Join(issues, ", "))
		severity = 3
	} else if len(warnings) > 0 {
		status = HealthStatusWarning
		message = fmt.Sprintf("Maven dependency warnings: %s", strings.Join(warnings, ", "))
		severity = 2
	}

	details := ""
	if len(issues) > 0 || len(warnings) > 0 {
		allItems := append(issues, warnings...)
		details = strings.Join(allItems, "\n")
	}

	return c.createDependencyHealthCheck(status, message, details, severity)
}

// checkGradleBuild checks Gradle dependencies and project health
func (c *DependencyChecker) checkGradleBuild(repoPath string) HealthCheck {
	foundFile, err := c.findGradleBuildFile(repoPath)
	if err != nil {
		return c.createDependencyHealthCheck(HealthStatusCritical, err.Error(), "", 3)
	}

	gradleCmd, useWrapper, err := c.determineGradleCommand(repoPath)
	if err != nil {
		return c.createDependencyHealthCheck(HealthStatusWarning, err.Error(),
			"Install Gradle or use Gradle wrapper (gradlew) for detailed dependency analysis", 2)
	}

	var issues, warnings []string
	c.checkGradleDependencyResolution(repoPath, gradleCmd, useWrapper, &issues, &warnings)
	c.checkGradleOutdatedDependencies(repoPath, gradleCmd, useWrapper, &warnings)
	c.checkGradleSecurityScan(repoPath, gradleCmd, useWrapper, &warnings)
	c.checkGradleBestPractices(useWrapper, &warnings)

	return c.createGradleHealthCheckResult(foundFile, issues, warnings)
}

func (c *DependencyChecker) findGradleBuildFile(repoPath string) (string, error) {
	buildFile := filepath.Join(repoPath, "build.gradle")
	buildFileKts := filepath.Join(repoPath, "build.gradle.kts")

	if _, err := os.Stat(buildFile); err == nil {
		return "build.gradle", nil
	}
	if _, err := os.Stat(buildFileKts); err == nil {
		return "build.gradle.kts", nil
	}

	return "", fmt.Errorf("gradle build file not found")
}

func (c *DependencyChecker) determineGradleCommand(repoPath string) (string, bool, error) {
	gradlewPath := filepath.Join(repoPath, "gradlew")
	if _, err := os.Stat(gradlewPath); err == nil {
		return "./gradlew", true, nil
	}

	if !c.commandExists("gradle") {
		return "", false, fmt.Errorf("gradle command not available")
	}
	return "gradle", false, nil
}

func (c *DependencyChecker) checkGradleDependencyResolution(repoPath, gradleCmd string, useWrapper bool, issues, warnings *[]string) {
	ctx, cancel := CreateHealthContext()
	defer cancel()

	var cmd *exec.Cmd
	if useWrapper {
		cmd = exec.CommandContext(ctx, "./gradlew", "dependencies", "--configuration", "compileClasspath", "-q")
	} else {
		cmd = exec.CommandContext(ctx, gradleCmd, "dependencies", "--configuration", "compileClasspath", "-q")
	}
	cmd.Dir = repoPath

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			*warnings = append(*warnings, "Gradle dependency resolution timed out")
		} else {
			*issues = append(*issues, "Dependencies cannot be resolved")
		}
	}
}

func (c *DependencyChecker) checkGradleOutdatedDependencies(repoPath, gradleCmd string, useWrapper bool, warnings *[]string) {
	ctx, cancel := CreateHealthContext()
	defer cancel()

	var cmd *exec.Cmd
	if useWrapper {
		cmd = exec.CommandContext(ctx, "./gradlew", "dependencyUpdates", "-q")
	} else {
		cmd = exec.CommandContext(ctx, gradleCmd, "dependencyUpdates", "-q")
	}
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err == nil {
		if strings.Contains(string(output), "outdated dependencies") {
			*warnings = append(*warnings, "Has outdated dependencies")
		}
	} else if ctx.Err() == context.DeadlineExceeded {
		*warnings = append(*warnings, "Gradle dependency update check timed out")
	}
}

func (c *DependencyChecker) checkGradleSecurityScan(repoPath, gradleCmd string, useWrapper bool, warnings *[]string) {
	var cmd *exec.Cmd
	if useWrapper {
		cmd = exec.Command("./gradlew", "dependencyCheckAnalyze", "-q")
	} else {
		cmd = exec.Command(gradleCmd, "dependencyCheckAnalyze", "-q")
	}
	cmd.Dir = repoPath

	if err := cmd.Run(); err == nil {
		*warnings = append(*warnings, "Security scan completed")
	}
}

func (c *DependencyChecker) checkGradleBestPractices(useWrapper bool, warnings *[]string) {
	if !useWrapper {
		*warnings = append(*warnings, "Gradle wrapper not found - consider using gradlew for reproducible builds")
	}
}

func (c *DependencyChecker) createGradleHealthCheckResult(foundFile string, issues, warnings []string) HealthCheck {
	status := HealthStatusHealthy
	message := fmt.Sprintf("Gradle dependencies are healthy (%s)", foundFile)
	severity := 1

	if len(issues) > 0 {
		status = HealthStatusCritical
		message = fmt.Sprintf("Gradle dependency issues: %s", strings.Join(issues, ", "))
		severity = 3
	} else if len(warnings) > 0 {
		status = HealthStatusWarning
		message = fmt.Sprintf("Gradle dependency warnings: %s", strings.Join(warnings, ", "))
		severity = 2
	}

	details := ""
	if len(issues) > 0 || len(warnings) > 0 {
		allItems := append(issues, warnings...)
		details = strings.Join(allItems, "\n")
	}

	return c.createDependencyHealthCheck(status, message, details, severity)
}

// checkPyprojectToml checks Python projects using pyproject.toml
func (c *DependencyChecker) checkPyprojectToml(repoPath string) HealthCheck {
	pyprojectFile := filepath.Join(repoPath, "pyproject.toml")

	if _, err := os.Stat(pyprojectFile); err != nil {
		return c.createDependencyHealthCheck(HealthStatusCritical, "pyproject.toml not found", "", 3)
	}

	content, err := os.ReadFile(pyprojectFile) // #nosec G304 - file path is controlled
	if err != nil {
		return c.createDependencyHealthCheck(HealthStatusCritical, "Cannot read pyproject.toml", err.Error(), 3)
	}

	var warnings, issues []string
	pipAvailable := c.commandExists("pip") || c.commandExists("pip3")

	c.checkPyprojectPipAvailability(pipAvailable, &warnings)
	c.checkPyprojectStructure(string(content), &warnings)
	c.checkPyprojectVirtualEnv(repoPath, &warnings)
	c.checkPyprojectDependencyManagement(repoPath, string(content), pipAvailable, &warnings)

	return c.createPythonHealthCheckResult(issues, warnings, "Python pyproject.toml")
}

func (c *DependencyChecker) checkPyprojectPipAvailability(pipAvailable bool, warnings *[]string) {
	if !pipAvailable {
		*warnings = append(*warnings, "pip not available - install pip for dependency management")
	}
}

func (c *DependencyChecker) checkPyprojectStructure(content string, warnings *[]string) {
	hasProjectSection := strings.Contains(content, "[project")
	hasBuildSystem := strings.Contains(content, "[build-system")
	hasDependencies := strings.Contains(content, "dependencies")

	if !hasProjectSection && !hasBuildSystem {
		*warnings = append(*warnings, "pyproject.toml may be incomplete - missing [project] or [build-system] sections")
	}
	if !hasDependencies {
		*warnings = append(*warnings, "No dependencies declared in pyproject.toml")
	}
}

func (c *DependencyChecker) checkPyprojectVirtualEnv(repoPath string, warnings *[]string) {
	venvPaths := []string{
		filepath.Join(repoPath, "venv"),
		filepath.Join(repoPath, ".venv"),
		filepath.Join(repoPath, "env"),
	}

	for _, venvPath := range venvPaths {
		if _, err := os.Stat(venvPath); err == nil {
			return // Found virtual environment
		}
	}
	*warnings = append(*warnings, "No virtual environment found - consider using venv for dependency isolation")
}

func (c *DependencyChecker) checkPyprojectDependencyManagement(repoPath, content string, pipAvailable bool, warnings *[]string) {
	if !pipAvailable {
		return
	}

	c.checkPyprojectDevDependencies(repoPath, content, warnings)
	c.checkPyprojectPipTools(repoPath, warnings)
}

func (c *DependencyChecker) checkPyprojectDevDependencies(repoPath, content string, warnings *[]string) {
	devFiles := []string{"requirements-dev.txt", "requirements-test.txt", "dev-requirements.txt"}
	for _, devFile := range devFiles {
		if _, err := os.Stat(filepath.Join(repoPath, devFile)); err == nil {
			return // Found dev dependencies file
		}
	}

	if strings.Contains(content, "test") {
		*warnings = append(*warnings, "Consider separating development/test dependencies")
	}
}

func (c *DependencyChecker) checkPyprojectPipTools(repoPath string, warnings *[]string) {
	requirementsIn := filepath.Join(repoPath, "requirements.in")
	requirementsTxt := filepath.Join(repoPath, "requirements.txt")

	_, hasRequirementsIn := os.Stat(requirementsIn)
	_, hasRequirementsTxt := os.Stat(requirementsTxt)

	if hasRequirementsIn == nil && hasRequirementsTxt != nil {
		*warnings = append(*warnings, "Found requirements.in but no requirements.txt - run pip-compile to generate lock file")
	}
}

func (c *DependencyChecker) createPythonHealthCheckResult(issues, warnings []string, baseMessage string) HealthCheck {
	status := HealthStatusHealthy
	message := fmt.Sprintf("%s is healthy", baseMessage)
	severity := 1

	if len(issues) > 0 {
		status = HealthStatusCritical
		message = fmt.Sprintf("Python project issues: %s", strings.Join(issues, ", "))
		severity = 3
	} else if len(warnings) > 0 {
		status = HealthStatusWarning
		message = fmt.Sprintf("Python project warnings: %s", strings.Join(warnings, ", "))
		severity = 2
	}

	details := ""
	if len(issues) > 0 || len(warnings) > 0 {
		allItems := append(issues, warnings...)
		details = strings.Join(allItems, "\n")
	}

	return c.createDependencyHealthCheck(status, message, details, severity)
}

// checkRequirementsTxt checks Python projects using requirements.txt
func (c *DependencyChecker) checkRequirementsTxt(repoPath string) HealthCheck {
	requirementsFile := filepath.Join(repoPath, "requirements.txt")

	if _, err := os.Stat(requirementsFile); err != nil {
		return c.createDependencyHealthCheck(HealthStatusCritical, "requirements.txt not found", "", 3)
	}

	content, err := os.ReadFile(requirementsFile) // #nosec G304 - file path is controlled
	if err != nil {
		return c.createDependencyHealthCheck(HealthStatusCritical, "Cannot read requirements.txt", err.Error(), 3)
	}

	var warnings, issues []string
	pipAvailable := c.commandExists("pip") || c.commandExists("pip3")

	dependencies := c.parseRequirementsTxt(string(content), &warnings)

	c.checkPyprojectPipAvailability(pipAvailable, &warnings)
	c.checkPyprojectVirtualEnv(repoPath, &warnings)
	c.checkRequirementsTxtDevDependencies(repoPath, len(dependencies), &warnings)
	c.checkRequirementsTxtPipCheck(repoPath, pipAvailable, &warnings)
	c.checkRequirementsTxtPipTools(repoPath, requirementsFile, &warnings)

	message := fmt.Sprintf("Python requirements.txt (%d dependencies)", len(dependencies))
	return c.createPythonHealthCheckResult(issues, warnings, message)
}

func (c *DependencyChecker) parseRequirementsTxt(content string, warnings *[]string) []string {
	lines := strings.Split(content, "\n")
	var dependencies []string
	unpinnedCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		dependencies = append(dependencies, line)
		if !c.isDependencyPinned(line) {
			unpinnedCount++
		}
	}

	if len(dependencies) == 0 {
		*warnings = append(*warnings, "requirements.txt is empty")
	} else if unpinnedCount > 0 {
		*warnings = append(*warnings, fmt.Sprintf("%d dependencies are not pinned to specific versions", unpinnedCount))
	}

	return dependencies
}

func (c *DependencyChecker) isDependencyPinned(line string) bool {
	return strings.Contains(line, "==") || strings.Contains(line, ">=") ||
		strings.Contains(line, "<=") || strings.Contains(line, "~=")
}

func (c *DependencyChecker) checkRequirementsTxtDevDependencies(repoPath string, dependencyCount int, warnings *[]string) {
	devFiles := []string{"requirements-dev.txt", "requirements-test.txt", "dev-requirements.txt"}
	for _, devFile := range devFiles {
		if _, err := os.Stat(filepath.Join(repoPath, devFile)); err == nil {
			return // Found dev dependencies file
		}
	}

	if dependencyCount > 5 {
		*warnings = append(*warnings, "Consider separating development dependencies into requirements-dev.txt")
	}
}

func (c *DependencyChecker) checkRequirementsTxtPipCheck(repoPath string, pipAvailable bool, warnings *[]string) {
	if !pipAvailable {
		return
	}

	ctx, cancel := CreateHealthContext()
	defer cancel()

	cmd := exec.CommandContext(ctx, "pip", "check")
	cmd.Dir = repoPath

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			*warnings = append(*warnings, "pip check timed out")
		} else {
			*warnings = append(*warnings, "Potential dependency conflicts detected (pip check failed)")
		}
	}
}

func (c *DependencyChecker) checkRequirementsTxtPipTools(repoPath, requirementsFile string, warnings *[]string) {
	requirementsIn := filepath.Join(repoPath, "requirements.in")
	inStat, inErr := os.Stat(requirementsIn)
	if inErr != nil {
		return // requirements.in doesn't exist
	}

	txtStat, txtErr := os.Stat(requirementsFile)
	if txtErr == nil && inStat.ModTime().After(txtStat.ModTime()) {
		*warnings = append(*warnings, "requirements.in is newer than requirements.txt - run pip-compile to update")
	}
}

// commandExists checks if a command is available in the system PATH
func (c *DependencyChecker) commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
