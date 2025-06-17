package core

import (
	"testing"
	"time"
)

// Test HealthStatus constants
func TestHealthStatus_Constants(t *testing.T) {
	tests := []struct {
		status   HealthStatus
		expected string
	}{
		{StatusHealthy, "healthy"},
		{StatusWarning, "warning"},
		{StatusCritical, "critical"},
		{StatusUnknown, "unknown"},
	}

	for _, test := range tests {
		if string(test.status) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, string(test.status))
		}
	}
}

// Test Severity constants
func TestSeverity_Constants(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityLow, "low"},
		{SeverityMedium, "medium"},
		{SeverityHigh, "high"},
		{SeverityCritical, "critical"},
	}

	for _, test := range tests {
		if string(test.severity) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, string(test.severity))
		}
	}
}

// Test Repository struct
func TestRepository_Creation(t *testing.T) {
	repo := Repository{
		Name:      "test-repo",
		Path:      "/path/to/repo",
		URL:       "https://github.com/user/repo.git",
		Branch:    "main",
		Tags:      []string{"backend", "go"},
		Metadata:  map[string]string{"owner": "team-a"},
		Language:  "go",
		Framework: "gin",
	}

	if repo.Name != "test-repo" {
		t.Errorf("Expected Name to be 'test-repo', got %s", repo.Name)
	}
	if repo.Path != "/path/to/repo" {
		t.Errorf("Expected Path to be '/path/to/repo', got %s", repo.Path)
	}
	if repo.URL != "https://github.com/user/repo.git" {
		t.Errorf("Expected URL to be 'https://github.com/user/repo.git', got %s", repo.URL)
	}
	if repo.Branch != "main" {
		t.Errorf("Expected Branch to be 'main', got %s", repo.Branch)
	}
	if len(repo.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(repo.Tags))
	}
	if repo.Tags[0] != "backend" || repo.Tags[1] != "go" {
		t.Errorf("Expected tags ['backend', 'go'], got %v", repo.Tags)
	}
	if repo.Metadata["owner"] != "team-a" {
		t.Errorf("Expected metadata owner to be 'team-a', got %s", repo.Metadata["owner"])
	}
	if repo.Language != "go" {
		t.Errorf("Expected Language to be 'go', got %s", repo.Language)
	}
	if repo.Framework != "gin" {
		t.Errorf("Expected Framework to be 'gin', got %s", repo.Framework)
	}
}

// Test CheckResult struct
func TestCheckResult_Creation(t *testing.T) {
	issues := []Issue{
		{
			Type:     "security",
			Message:  "Vulnerable dependency found",
			Severity: SeverityHigh,
			Location: &Location{
				File: "go.mod",
				Line: 5,
			},
			Suggestion: "Update dependency to fix vulnerability",
		},
	}

	warnings := []Warning{
		{
			Type:    "style",
			Message: "Consider using consistent naming",
			Location: &Location{
				File: "main.go",
				Line: 10,
			},
		},
	}

	metrics := map[string]interface{}{
		"complexity": 15,
		"coverage":   85.5,
	}

	metadata := map[string]string{
		"checker_version": "1.0.0",
		"execution_time":  "2.5s",
	}

	result := CheckResult{
		ID:       "security-check-001",
		Name:     "Security Vulnerability Check",
		Category: "security",
		Status:   StatusWarning,
		Score:    75,
		MaxScore: 100,
		Issues:   issues,
		Warnings: warnings,
		Metrics:  metrics,
		Metadata: metadata,
	}

	if result.ID != "security-check-001" {
		t.Errorf("Expected ID to be 'security-check-001', got %s", result.ID)
	}
	if result.Name != "Security Vulnerability Check" {
		t.Errorf("Expected Name to be 'Security Vulnerability Check', got %s", result.Name)
	}
	if result.Category != "security" {
		t.Errorf("Expected Category to be 'security', got %s", result.Category)
	}
	if result.Status != StatusWarning {
		t.Errorf("Expected Status to be StatusWarning, got %s", result.Status)
	}
	if result.Score != 75 {
		t.Errorf("Expected Score to be 75, got %d", result.Score)
	}
	if result.MaxScore != 100 {
		t.Errorf("Expected MaxScore to be 100, got %d", result.MaxScore)
	}
	if len(result.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(result.Issues))
	}
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
	if result.Metrics["complexity"] != 15 {
		t.Errorf("Expected complexity metric to be 15, got %v", result.Metrics["complexity"])
	}
	if result.Metadata["checker_version"] != "1.0.0" {
		t.Errorf("Expected checker_version to be '1.0.0', got %s", result.Metadata["checker_version"])
	}
}

// Test Issue struct
func TestIssue_Creation(t *testing.T) {
	issue := Issue{
		Type:     "security",
		Message:  "SQL injection vulnerability",
		Severity: SeverityCritical,
		Location: &Location{
			File:     "database.go",
			Line:     42,
			Column:   15,
			Function: "getUserData",
		},
		Suggestion: "Use parameterized queries to prevent SQL injection",
		Context: map[string]interface{}{
			"function": "getUserData",
			"variable": "query",
		},
	}

	if issue.Type != "security" {
		t.Errorf("Expected Type to be 'security', got %s", issue.Type)
	}
	if issue.Message != "SQL injection vulnerability" {
		t.Errorf("Expected Message to be 'SQL injection vulnerability', got %s", issue.Message)
	}
	if issue.Location.File != "database.go" {
		t.Errorf("Expected File to be 'database.go', got %s", issue.Location.File)
	}
	if issue.Location.Line != 42 {
		t.Errorf("Expected Line to be 42, got %d", issue.Location.Line)
	}
	if issue.Location.Column != 15 {
		t.Errorf("Expected Column to be 15, got %d", issue.Location.Column)
	}
	if issue.Severity != SeverityCritical {
		t.Errorf("Expected Severity to be SeverityCritical, got %s", issue.Severity)
	}
	if issue.Suggestion != "Use parameterized queries to prevent SQL injection" {
		t.Errorf("Expected Suggestion to match, got %s", issue.Suggestion)
	}
	if issue.Context["function"] != "getUserData" {
		t.Errorf("Expected Context function to be 'getUserData', got %v", issue.Context["function"])
	}
}

// Test Warning struct
func TestWarning_Creation(t *testing.T) {
	warning := Warning{
		Type:    "style",
		Message: "Function name should start with uppercase letter",
		Location: &Location{
			File:     "utils.go",
			Line:     23,
			Column:   1,
			Function: "validateInput",
		},
	}

	if warning.Type != "style" {
		t.Errorf("Expected Type to be 'style', got %s", warning.Type)
	}
	if warning.Message != "Function name should start with uppercase letter" {
		t.Errorf("Expected Message to match, got %s", warning.Message)
	}
	if warning.Location.File != "utils.go" {
		t.Errorf("Expected File to be 'utils.go', got %s", warning.Location.File)
	}
	if warning.Location.Line != 23 {
		t.Errorf("Expected Line to be 23, got %d", warning.Location.Line)
	}
	if warning.Location.Column != 1 {
		t.Errorf("Expected Column to be 1, got %d", warning.Location.Column)
	}
	if warning.Location.Function != "validateInput" {
		t.Errorf("Expected Function to be 'validateInput', got %s", warning.Location.Function)
	}
}

// Test Location struct
func TestLocation_Creation(t *testing.T) {
	location := Location{
		File:     "main.go",
		Line:     100,
		Column:   25,
		Function: "processData",
	}

	if location.File != "main.go" {
		t.Errorf("Expected File to be 'main.go', got %s", location.File)
	}
	if location.Line != 100 {
		t.Errorf("Expected Line to be 100, got %d", location.Line)
	}
	if location.Column != 25 {
		t.Errorf("Expected Column to be 25, got %d", location.Column)
	}
	if location.Function != "processData" {
		t.Errorf("Expected Function to be 'processData', got %s", location.Function)
	}
}

// Test FileSystem interface behavior (mock implementation)
type MockFileSystem struct {
	files map[string][]byte
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
	}
}

func (fs *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if data, exists := fs.files[path]; exists {
		return data, nil
	}
	return nil, &FileNotFoundError{Path: path}
}

func (fs *MockFileSystem) WriteFile(path string, data []byte) error {
	fs.files[path] = data
	return nil
}

func (fs *MockFileSystem) Exists(path string) bool {
	_, exists := fs.files[path]
	return exists
}

func (fs *MockFileSystem) IsDir(path string) bool {
	// Simple mock implementation
	return false
}

func (fs *MockFileSystem) ListFiles(path string, pattern string) ([]string, error) {
	var files []string
	for filepath := range fs.files {
		files = append(files, filepath)
	}
	return files, nil
}

func (fs *MockFileSystem) Walk(path string, walkFn func(path string, info FileInfo) error) error {
	for filepath := range fs.files {
		info := FileInfo{
			Name:    filepath,
			Size:    int64(len(fs.files[filepath])),
			Mode:    0644,
			ModTime: time.Now(),
			IsDir:   false,
		}
		if err := walkFn(filepath, info); err != nil {
			return err
		}
	}
	return nil
}

// Custom error type for testing
type FileNotFoundError struct {
	Path string
}

func (e *FileNotFoundError) Error() string {
	return "file not found: " + e.Path
}

// Test MockFileSystem
func TestMockFileSystem(t *testing.T) {
	fs := NewMockFileSystem()

	// Test WriteFile and ReadFile
	testData := []byte("test content")
	err := fs.WriteFile("test.txt", testData)
	if err != nil {
		t.Errorf("Expected no error writing file, got %v", err)
	}

	data, err := fs.ReadFile("test.txt")
	if err != nil {
		t.Errorf("Expected no error reading file, got %v", err)
	}
	if string(data) != "test content" {
		t.Errorf("Expected 'test content', got %s", string(data))
	}

	// Test Exists
	if !fs.Exists("test.txt") {
		t.Error("Expected file to exist")
	}
	if fs.Exists("nonexistent.txt") {
		t.Error("Expected file not to exist")
	}

	// Test ReadFile for non-existent file
	_, err = fs.ReadFile("nonexistent.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test ListFiles
	files, err := fs.ListFiles("", "*")
	if err != nil {
		t.Errorf("Expected no error listing files, got %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	// Test Walk
	walkCalled := false
	err = fs.Walk("", func(path string, info FileInfo) error {
		walkCalled = true
		if path != "test.txt" {
			t.Errorf("Expected path 'test.txt', got %s", path)
		}
		if info.Size != int64(len(testData)) {
			t.Errorf("Expected size %d, got %d", len(testData), info.Size)
		}
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error in Walk, got %v", err)
	}
	if !walkCalled {
		t.Error("Expected Walk function to be called")
	}
}

// Test FileInfo struct
func TestFileInfo_Creation(t *testing.T) {
	modTime := time.Now()
	info := FileInfo{
		Name:    "test.go",
		Size:    1024,
		Mode:    0644,
		ModTime: modTime,
		IsDir:   false,
	}

	if info.Name != "test.go" {
		t.Errorf("Expected Name to be 'test.go', got %s", info.Name)
	}
	if info.Size != 1024 {
		t.Errorf("Expected Size to be 1024, got %d", info.Size)
	}
	if info.Mode != 0644 {
		t.Errorf("Expected Mode to be 0644, got %o", info.Mode)
	}
	if !info.ModTime.Equal(modTime) {
		t.Errorf("Expected ModTime to match, got %v", info.ModTime)
	}
	if info.IsDir {
		t.Error("Expected IsDir to be false")
	}
}
