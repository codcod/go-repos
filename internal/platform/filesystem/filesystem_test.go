package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codcod/repos/internal/core"
)

func TestOSFileSystem_NewOSFileSystem(t *testing.T) {
	fs := NewOSFileSystem()
	if fs == nil {
		t.Fatal("NewOSFileSystem() returned nil")
	}
}

func TestOSFileSystem_ReadWriteFile(t *testing.T) {
	fs := NewOSFileSystem()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	testData := []byte("test content")

	// Test writing
	err = fs.WriteFile(testFile, testData)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Test reading
	readData, err := fs.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(readData))
	}
}

func TestOSFileSystem_ReadFile_NonExistent(t *testing.T) {
	fs := NewOSFileSystem()

	_, err := fs.ReadFile("/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestOSFileSystem_WriteFile_DirectoryTraversal(t *testing.T) {
	fs := NewOSFileSystem()

	// Test directory traversal protection
	err := fs.WriteFile("../../../etc/passwd", []byte("malicious content"))
	if err == nil {
		t.Error("Expected error for directory traversal attempt")
	}
}

func TestOSFileSystem_ReadFile_DirectoryTraversal(t *testing.T) {
	fs := NewOSFileSystem()

	// Test directory traversal protection
	_, err := fs.ReadFile("../../../etc/passwd")
	if err == nil {
		t.Error("Expected error for directory traversal attempt")
	}
}

func TestOSFileSystem_Exists(t *testing.T) {
	fs := NewOSFileSystem()

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "exists-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Test existing file
	if !fs.Exists(tempFile.Name()) {
		t.Error("Expected file to exist")
	}

	// Test non-existent file
	if fs.Exists("/non/existent/file.txt") {
		t.Error("Expected non-existent file to not exist")
	}
}

func TestOSFileSystem_IsDir(t *testing.T) {
	fs := NewOSFileSystem()

	// Create temp directory and file
	tempDir, err := os.MkdirTemp("", "isdir-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(tempFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Test directory
	if !fs.IsDir(tempDir) {
		t.Error("Expected directory to be identified as directory")
	}

	// Test file
	if fs.IsDir(tempFile) {
		t.Error("Expected file to not be identified as directory")
	}

	// Test non-existent path
	if fs.IsDir("/non/existent/path") {
		t.Error("Expected non-existent path to not be identified as directory")
	}
}

func TestOSFileSystem_ListFiles(t *testing.T) {
	fs := NewOSFileSystem()

	// Create temp directory with test files
	tempDir, err := os.MkdirTemp("", "listfiles-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{"file1.txt", "file2.txt", "file3.log", "subdir/file4.txt"}
	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		err = os.WriteFile(fullPath, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Test listing *.txt files
	files, err := fs.ListFiles(tempDir, "*.txt")
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	// Should find 3 .txt files (including the one in subdir)
	if len(files) != 3 {
		t.Errorf("Expected 3 .txt files, got %d: %v", len(files), files)
	}

	// Verify all files end with .txt
	for _, file := range files {
		if !strings.HasSuffix(file, ".txt") {
			t.Errorf("Expected file to end with .txt, got %s", file)
		}
	}
}

func TestOSFileSystem_Walk(t *testing.T) {
	fs := NewOSFileSystem()

	// Create temp directory with test structure
	tempDir, err := os.MkdirTemp("", "walk-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test structure
	testPaths := []string{
		"file1.txt",
		"subdir1/file2.txt",
		"subdir1/subdir2/file3.txt",
	}

	for _, path := range testPaths {
		fullPath := filepath.Join(tempDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		err = os.WriteFile(fullPath, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	// Walk and collect paths
	var visitedPaths []string
	err = fs.Walk(tempDir, func(path string, info core.FileInfo) error {
		// Make path relative to temp dir for easier testing
		relPath, _ := filepath.Rel(tempDir, path)
		if relPath != "." { // Skip the root directory
			visitedPaths = append(visitedPaths, relPath)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should visit at least the files we created
	expectedFiles := []string{"file1.txt", "subdir1", "subdir1/file2.txt", "subdir1/subdir2", "subdir1/subdir2/file3.txt"}
	for _, expected := range expectedFiles {
		found := false
		for _, visited := range visitedPaths {
			if visited == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to visit %s, but it was not found in: %v", expected, visitedPaths)
		}
	}
}

func TestMemoryFileSystem_NewMemoryFileSystem(t *testing.T) {
	fs := NewMemoryFileSystem()
	if fs == nil {
		t.Fatal("NewMemoryFileSystem() returned nil")
	}

	if fs.files == nil {
		t.Fatal("files map is nil")
	}

	if fs.dirs == nil {
		t.Fatal("dirs map is nil")
	}
}

func TestMemoryFileSystem_ReadWriteFile(t *testing.T) {
	fs := NewMemoryFileSystem()

	testPath := "/test/file.txt"
	testData := []byte("test content")

	// Test writing
	err := fs.WriteFile(testPath, testData)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Test reading
	readData, err := fs.ReadFile(testPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(readData))
	}
}

func TestMemoryFileSystem_ReadFile_NonExistent(t *testing.T) {
	fs := NewMemoryFileSystem()

	_, err := fs.ReadFile("/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestMemoryFileSystem_Exists(t *testing.T) {
	fs := NewMemoryFileSystem()

	// Test non-existent file
	if fs.Exists("/non/existent/file.txt") {
		t.Error("Expected non-existent file to not exist")
	}

	// Add file and test
	fs.AddFile("/test/file.txt", "content")
	if !fs.Exists("/test/file.txt") {
		t.Error("Expected file to exist after adding")
	}

	// Add directory and test
	fs.AddDir("/test/dir")
	if !fs.Exists("/test/dir") {
		t.Error("Expected directory to exist after adding")
	}
}

func TestMemoryFileSystem_IsDir(t *testing.T) {
	fs := NewMemoryFileSystem()

	// Test non-existent path
	if fs.IsDir("/non/existent/path") {
		t.Error("Expected non-existent path to not be directory")
	}

	// Add file and test
	fs.AddFile("/test/file.txt", "content")
	if fs.IsDir("/test/file.txt") {
		t.Error("Expected file to not be directory")
	}

	// Add directory and test
	fs.AddDir("/test/dir")
	if !fs.IsDir("/test/dir") {
		t.Error("Expected directory to be identified as directory")
	}
}

func TestMemoryFileSystem_ListFiles(t *testing.T) {
	fs := NewMemoryFileSystem()

	// Add test files
	fs.AddFile("/test/file1.txt", "content1")
	fs.AddFile("/test/file2.txt", "content2")
	fs.AddFile("/test/file3.log", "content3")
	fs.AddFile("/other/file4.txt", "content4")

	// Test listing *.txt files in /test
	files, err := fs.ListFiles("/test", "*.txt")
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 .txt files, got %d: %v", len(files), files)
	}

	// Verify all files end with .txt
	for _, file := range files {
		if !strings.HasSuffix(file, ".txt") {
			t.Errorf("Expected file to end with .txt, got %s", file)
		}
	}
}

func TestMemoryFileSystem_Walk(t *testing.T) {
	fs := NewMemoryFileSystem()

	// Add test structure
	fs.AddDir("/test")
	fs.AddDir("/test/subdir")
	fs.AddFile("/test/file1.txt", "content1")
	fs.AddFile("/test/subdir/file2.txt", "content2")

	// Walk and collect paths
	var visitedPaths []string
	err := fs.Walk("/test", func(path string, info core.FileInfo) error {
		visitedPaths = append(visitedPaths, path)
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should visit directories and files
	expectedPaths := []string{"/test", "/test/subdir", "/test/file1.txt", "/test/subdir/file2.txt"}
	for _, expected := range expectedPaths {
		found := false
		for _, visited := range visitedPaths {
			if visited == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to visit %s, but it was not found in: %v", expected, visitedPaths)
		}
	}
}

func TestMemoryFileSystem_AddFileAndDir(t *testing.T) {
	fs := NewMemoryFileSystem()

	// Test AddFile
	fs.AddFile("/test/file.txt", "test content")

	data, err := fs.ReadFile("/test/file.txt")
	if err != nil {
		t.Fatalf("Failed to read added file: %v", err)
	}

	if string(data) != "test content" {
		t.Errorf("Expected 'test content', got %s", string(data))
	}

	// Test AddDir
	fs.AddDir("/test/directory")

	if !fs.IsDir("/test/directory") {
		t.Error("Expected added directory to be identified as directory")
	}
}
