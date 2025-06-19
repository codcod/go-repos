package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	return os.WriteFile(cleanPath, data, 0600)
}

// Exists checks if a file or directory exists
func (f *OSFileSystem) Exists(path string) bool {
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return false
	}

	_, err := os.Stat(cleanPath)
	return err == nil
}

// IsDir checks if a path is a directory
func (f *OSFileSystem) IsDir(path string) bool {
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return false
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ListFiles lists files in a directory matching a pattern
func (f *OSFileSystem) ListFiles(dir string, pattern string) ([]string, error) {
	cleanDir := filepath.Clean(dir)
	if strings.Contains(cleanDir, "..") {
		return nil, fs.ErrInvalid
	}

	var files []string
	err := filepath.Walk(cleanDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if pattern == "" {
				files = append(files, path)
			} else {
				matched, matchErr := filepath.Match(pattern, filepath.Base(path))
				if matchErr != nil {
					return matchErr
				}
				if matched {
					files = append(files, path)
				}
			}
		}
		return nil
	})

	return files, err
}

// Walk walks the file tree rooted at root
func (f *OSFileSystem) Walk(root string, walkFn func(path string, info core.FileInfo) error) error {
	cleanRoot := filepath.Clean(root)
	if strings.Contains(cleanRoot, "..") {
		return fs.ErrInvalid
	}

	return filepath.Walk(cleanRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fileInfo := core.FileInfo{
			Name:    info.Name(),
			Size:    info.Size(),
			Mode:    uint32(info.Mode()),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		}

		return walkFn(path, fileInfo)
	})
}
