package health

import (
	"strings"
	"testing"
	"time"

	"github.com/codcod/repos/internal/config"
)

func TestPythonDependencyChecker(t *testing.T) {
	checker := &DependencyChecker{}

	// Test pyproject.toml project
	t.Run("Pyproject.toml Project", func(t *testing.T) {
		repo := config.Repository{
			Name: "jira-epic-timeline",
			Path: "/Users/nicos/Projects/private/repos/cloned_repos/jira-epic-timeline",
		}

		result := checker.Check(repo)

		// Should detect Python project
		if result.Name != "Dependencies" {
			t.Errorf("Expected name 'Dependencies', got %s", result.Name)
		}

		if result.Category != "dependencies" {
			t.Errorf("Expected category 'dependencies', got %s", result.Category)
		}

		// Should have some status (not empty)
		if result.Status == "" {
			t.Error("Expected non-empty status")
		}

		// Should have a reasonable timestamp
		if time.Since(result.LastChecked) > time.Minute {
			t.Error("LastChecked timestamp seems too old")
		}

		t.Logf("Python pyproject.toml check result: %s - %s", result.Status, result.Message)
		if result.Details != "" {
			t.Logf("Details: %s", result.Details)
		}
	})

	// Test command existence for Python tools
	t.Run("Python Tools Detection", func(t *testing.T) {
		pip := checker.commandExists("pip")
		pip3 := checker.commandExists("pip3")
		python := checker.commandExists("python")
		python3 := checker.commandExists("python3")

		t.Logf("pip available: %v", pip)
		t.Logf("pip3 available: %v", pip3)
		t.Logf("python available: %v", python)
		t.Logf("python3 available: %v", python3)

		if !pip && !pip3 {
			t.Log("Warning: Neither pip nor pip3 is available")
		}
	})
}

func TestPyprojectTomlCheckerDirect(t *testing.T) {
	checker := &DependencyChecker{}

	// Test directly calling the pyproject.toml checker
	result := checker.checkPyprojectToml("/Users/nicos/Projects/private/repos/cloned_repos/jira-epic-timeline")

	t.Logf("Direct pyproject.toml check: %s - %s", result.Status, result.Message)
	if result.Details != "" {
		t.Logf("Details:\n%s", result.Details)
	}

	// Basic validation
	if result.Name != "Dependencies" {
		t.Errorf("Expected name 'Dependencies', got %s", result.Name)
	}

	if result.Status == "" {
		t.Error("Expected non-empty status")
	}
}

func TestRequirementsTxtChecker(t *testing.T) {
	checker := &DependencyChecker{}

	// Test directly calling the requirements.txt checker
	result := checker.checkRequirementsTxt("/tmp/python-test-project")

	t.Logf("Requirements.txt check: %s - %s", result.Status, result.Message)
	if result.Details != "" {
		t.Logf("Details:\n%s", result.Details)
	}

	// Basic validation
	if result.Name != "Dependencies" {
		t.Errorf("Expected name 'Dependencies', got %s", result.Name)
	}

	if result.Status == "" {
		t.Error("Expected non-empty status")
	}

	// Should detect unpinned dependencies
	if !strings.Contains(result.Message, "warning") && !strings.Contains(result.Details, "not pinned") {
		t.Log("Note: Expected to detect unpinned dependencies")
	}
}
