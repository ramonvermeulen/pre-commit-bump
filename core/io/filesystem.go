package io

import "os"

// FileSystem abstracts file system operations for better testability
type FileSystem interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm int) error
}

// OSFileSystem implements FileSystem using the standard os package
type OSFileSystem struct{}

// NewOSFileSystem creates a new OSFileSystem instance
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// ReadFile reads a file from the file system
func (fs *OSFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// WriteFile writes data to a file in the file system
func (fs *OSFileSystem) WriteFile(filename string, data []byte, perm int) error {
	return os.WriteFile(filename, data, os.FileMode(perm))
}
