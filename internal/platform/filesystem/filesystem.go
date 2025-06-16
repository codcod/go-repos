package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codcod/repos/internal/core"
)

// OSFileSystem implements core.FileSystem using the OS file system
type OSFileSystem struct{}

// NewOSFileSystem creates a new OS file system
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// ReadFile reads a file from the file system
func (f *OSFileSystem) ReadFile(path string) ([]byte, error) {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// Basic security check
	if strings.Contains(cleanPath, "..") {
		return nil, fs.ErrInvalid
	}

	return os.ReadFile(cleanPath)
}

// WriteFile writes data to a file
func (f *OSFileSystem) WriteFile(path string, data []byte) error {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// Basic security check
	if strings.Contains(cleanPath, "..") {
		return fs.ErrInvalid
	}

	// Ensure directory exists
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(cleanPath, data, 0644)
}

// Exists checks if a file or directory exists
func (f *OSFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDir checks if the path is a directory
func (f *OSFileSystem) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ListFiles lists files matching a pattern
func (f *OSFileSystem) ListFiles(path string, pattern string) ([]string, error) {
	var files []string

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		if info.IsDir() {
			// Skip common directories that shouldn't be analyzed
			name := filepath.Base(filePath)
			skipDirs := []string{
				".git", ".svn", ".hg",
				"node_modules", "vendor", "target", "build", "dist",
				".venv", "venv", "env", "__pycache__",
				".gradle", ".next", ".nuxt",
			}

			for _, skipDir := range skipDirs {
				if name == skipDir {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check if file matches pattern
		if matched, err := filepath.Match(pattern, filepath.Base(filePath)); err == nil && matched {
			files = append(files, filePath)
		}

		return nil
	})

	return files, err
}

// Walk walks the file tree
func (f *OSFileSystem) Walk(path string, walkFn func(path string, info core.FileInfo) error) error {
	return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		fileInfo := core.FileInfo{
			Name:    info.Name(),
			Size:    info.Size(),
			Mode:    uint32(info.Mode()),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		}

		return walkFn(filePath, fileInfo)
	})
}

// MemoryFileSystem implements core.FileSystem in memory (for testing)
type MemoryFileSystem struct {
	files map[string][]byte
	dirs  map[string]bool
}

// NewMemoryFileSystem creates a new in-memory file system
func NewMemoryFileSystem() *MemoryFileSystem {
	return &MemoryFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

// ReadFile reads a file from memory
func (f *MemoryFileSystem) ReadFile(path string) ([]byte, error) {
	cleanPath := filepath.Clean(path)

	if data, exists := f.files[cleanPath]; exists {
		return data, nil
	}

	return nil, fs.ErrNotExist
}

// WriteFile writes data to memory
func (f *MemoryFileSystem) WriteFile(path string, data []byte) error {
	cleanPath := filepath.Clean(path)

	// Ensure parent directories exist
	dir := filepath.Dir(cleanPath)
	f.dirs[dir] = true

	f.files[cleanPath] = data
	return nil
}

// Exists checks if a file or directory exists in memory
func (f *MemoryFileSystem) Exists(path string) bool {
	cleanPath := filepath.Clean(path)

	if _, exists := f.files[cleanPath]; exists {
		return true
	}

	if _, exists := f.dirs[cleanPath]; exists {
		return true
	}

	return false
}

// IsDir checks if the path is a directory in memory
func (f *MemoryFileSystem) IsDir(path string) bool {
	cleanPath := filepath.Clean(path)
	_, exists := f.dirs[cleanPath]
	return exists
}

// ListFiles lists files matching a pattern in memory
func (f *MemoryFileSystem) ListFiles(path string, pattern string) ([]string, error) {
	var files []string

	for filePath := range f.files {
		if strings.HasPrefix(filePath, path) {
			if matched, err := filepath.Match(pattern, filepath.Base(filePath)); err == nil && matched {
				files = append(files, filePath)
			}
		}
	}

	return files, nil
}

// Walk walks the memory file tree
func (f *MemoryFileSystem) Walk(path string, walkFn func(path string, info core.FileInfo) error) error {
	// Walk directories first
	for dirPath := range f.dirs {
		if strings.HasPrefix(dirPath, path) {
			info := core.FileInfo{
				Name:    filepath.Base(dirPath),
				Size:    0,
				Mode:    0755,
				ModTime: time.Now(),
				IsDir:   true,
			}

			if err := walkFn(dirPath, info); err != nil {
				return err
			}
		}
	}

	// Walk files
	for filePath, data := range f.files {
		if strings.HasPrefix(filePath, path) {
			info := core.FileInfo{
				Name:    filepath.Base(filePath),
				Size:    int64(len(data)),
				Mode:    0644,
				ModTime: time.Now(),
				IsDir:   false,
			}

			if err := walkFn(filePath, info); err != nil {
				return err
			}
		}
	}

	return nil
}

// AddFile adds a file to memory file system (helper for testing)
func (f *MemoryFileSystem) AddFile(path string, content string) {
	f.files[filepath.Clean(path)] = []byte(content)
}

// AddDir adds a directory to memory file system (helper for testing)
func (f *MemoryFileSystem) AddDir(path string) {
	f.dirs[filepath.Clean(path)] = true
}
