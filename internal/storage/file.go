package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileStorage handles file-based storage operations
type FileStorage struct {
	filePath string
}

// NewFileStorage creates a new FileStorage instance with the specified file path
func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		filePath: filePath,
	}
}

// ReadLines reads all lines from the file and returns them as a slice of strings
func (fs *FileStorage) ReadLines() ([]string, error) {
	// Check if file exists
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		// Return empty slice if file doesn't exist (not an error for our use case)
		return []string{}, nil
	}

	file, err := os.Open(fs.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", fs.filePath, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and whitespace-only lines
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fs.filePath, err)
	}

	return lines, nil
}

// WriteLines writes a slice of strings to the file, one line per string
func (fs *FileStorage) WriteLines(lines []string) error {
	// Ensure the directory exists
	dir := filepath.Dir(fs.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.Create(fs.filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fs.filePath, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write to file %s: %w", fs.filePath, err)
		}
	}

	return nil
}
