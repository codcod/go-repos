package health

import (
	"testing"
	"time"

	"github.com/codcod/repos/internal/config"
)

func TestJavaDependencyChecker(t *testing.T) {
	checker := &DependencyChecker{}

	// Test Maven project
	t.Run("Maven Project", func(t *testing.T) {
		repo := config.Repository{
			Name: "docusign-connectors",
			Path: "/Users/nicos/Projects/private/repos/cloned_repos/docusign-connectors",
		}

		result := checker.Check(repo)

		// Should detect Maven project
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

		t.Logf("Maven check result: %s - %s", result.Status, result.Message)
		if result.Details != "" {
			t.Logf("Details: %s", result.Details)
		}
	})

	// Test Gradle project (even though we don't have one in the workspace)
	t.Run("Gradle Project Detection", func(t *testing.T) {
		// This will test the checkGradleBuild method indirectly
		// by checking if it can handle a non-existent Gradle project gracefully
		repo := config.Repository{
			Name: "non-existent-gradle",
			Path: "../non-existent-path",
		}

		result := checker.Check(repo)

		// Should handle gracefully
		if result.Name != "Dependencies" {
			t.Errorf("Expected name 'Dependencies', got %s", result.Name)
		}

		t.Logf("Non-existent project result: %s - %s", result.Status, result.Message)
	})
}

func TestCommandExists(t *testing.T) {
	checker := &DependencyChecker{}

	// Test with a command that should exist
	if !checker.commandExists("echo") {
		t.Error("Expected 'echo' command to exist")
	}

	// Test with a command that should not exist
	if checker.commandExists("definitely-not-a-real-command-12345") {
		t.Error("Expected non-existent command to return false")
	}

	// Test Maven command availability
	mvnExists := checker.commandExists("mvn")
	t.Logf("Maven available: %v", mvnExists)

	// Test Gradle command availability
	gradleExists := checker.commandExists("gradle")
	t.Logf("Gradle available: %v", gradleExists)
}
