package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewOSFileSystem(t *testing.T) {
	fs := NewOSFileSystem()
	if fs == nil {
		t.Fatal("Expected filesystem to be created")
	}
}

func TestOSFileSystem_ReadWrite(t *testing.T) {
	fs := NewOSFileSystem()

	// Create temp directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testData := []byte("test content")

	// Test WriteFile
	err := fs.WriteFile(testFile, testData)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Test ReadFile
	content, err := fs.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("Expected content '%s', got '%s'", testData, content)
	}
}

func TestOSFileSystem_Exists(t *testing.T) {
	fs := NewOSFileSystem()

	// Test with existing file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// File shouldn't exist yet
	if fs.Exists(testFile) {
		t.Error("File should not exist yet")
	}

	// Create file
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Now it should exist
	if !fs.Exists(testFile) {
		t.Error("File should exist now")
	}
}

func TestOSFileSystem_IsDir(t *testing.T) {
	fs := NewOSFileSystem()

	tempDir := t.TempDir()

	// Test directory
	if !fs.IsDir(tempDir) {
		t.Error("Should recognize directory")
	}

	// Test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if fs.IsDir(testFile) {
		t.Error("Should not recognize file as directory")
	}
}

func TestOSFileSystem_SecurityChecks(t *testing.T) {
	fs := NewOSFileSystem()

	// Test path traversal protection
	_, err := fs.ReadFile("../../../etc/passwd")
	if err == nil {
		t.Error("Should reject path traversal attempts")
	}

	err = fs.WriteFile("../../../tmp/test", []byte("test"))
	if err == nil {
		t.Error("Should reject path traversal attempts in WriteFile")
	}
}
