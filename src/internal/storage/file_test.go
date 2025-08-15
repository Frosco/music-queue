package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFileStorage(t *testing.T) {
	filePath := "/tmp/test.txt"
	storage := NewFileStorage(filePath)

	if storage == nil {
		t.Fatal("NewFileStorage returned nil")
	}

	if storage.filePath != filePath {
		t.Errorf("Expected filePath %s, got %s", filePath, storage.filePath)
	}
}

func TestFileStorage_ReadLines_FileNotExists(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

	storage := NewFileStorage(nonExistentFile)
	lines, err := storage.ReadLines()

	if err != nil {
		t.Errorf("ReadLines should not return error for non-existent file, got: %v", err)
	}

	if len(lines) != 0 {
		t.Errorf("Expected empty slice for non-existent file, got %d lines", len(lines))
	}
}

func TestFileStorage_ReadLines_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	emptyFile := filepath.Join(tempDir, "empty.txt")

	// Create empty file
	file, err := os.Create(emptyFile)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	storage := NewFileStorage(emptyFile)
	lines, err := storage.ReadLines()

	if err != nil {
		t.Errorf("ReadLines should not return error for empty file, got: %v", err)
	}

	if len(lines) != 0 {
		t.Errorf("Expected empty slice for empty file, got %d lines", len(lines))
	}
}

func TestFileStorage_ReadLines_WithEmptyLines(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Create test file with empty lines and whitespace
	content := "Artist 1 - Album 1\nArtist 2 - Album 2\n\n  \nArtist 3 - Album 3\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	storage := NewFileStorage(testFile)
	lines, err := storage.ReadLines()

	if err != nil {
		t.Errorf("ReadLines returned error: %v", err)
	}

	// Should return only non-empty lines, trimmed
	expected := []string{"Artist 1 - Album 1", "Artist 2 - Album 2", "Artist 3 - Album 3"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines, got %d", len(expected), len(lines))
	}

	for i, expectedLine := range expected {
		if i < len(lines) && lines[i] != expectedLine {
			t.Errorf("Line %d: expected %q, got %q", i, expectedLine, lines[i])
		}
	}
}

func TestFileStorage_WriteLines(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	storage := NewFileStorage(testFile)
	testLines := []string{"Artist 1 - Album 1", "Artist 2 - Album 2", "Artist 3 - Album 3"}

	err := storage.WriteLines(testLines)
	if err != nil {
		t.Errorf("WriteLines returned error: %v", err)
	}

	// Verify file was created and contains expected content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read test file: %v", err)
	}

	expectedContent := "Artist 1 - Album 1\nArtist 2 - Album 2\nArtist 3 - Album 3\n"
	if string(content) != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, string(content))
	}
}

func TestFileStorage_WriteLines_EmptySlice(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "empty_write_test.txt")

	storage := NewFileStorage(testFile)

	err := storage.WriteLines([]string{})
	if err != nil {
		t.Errorf("WriteLines returned error for empty slice: %v", err)
	}

	// Verify file was created but is empty
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("ReadLines returned error: %v", err)
	}

	if len(lines) != 0 {
		t.Errorf("Expected empty file, got %d lines", len(lines))
	}
}

func TestFileStorage_WriteLines_CreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nestedFile := filepath.Join(tempDir, "nested", "subdir", "test.txt")

	storage := NewFileStorage(nestedFile)
	testLines := []string{"Test Album"}

	err := storage.WriteLines(testLines)
	if err != nil {
		t.Errorf("WriteLines should create directories, got error: %v", err)
	}

	// Verify file was created in nested directory
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("ReadLines returned error: %v", err)
	}

	if len(lines) != 1 || lines[0] != "Test Album" {
		t.Errorf("Expected ['Test Album'], got %v", lines)
	}
}

func TestFileStorage_GetFilePath(t *testing.T) {
	filePath := "/tmp/test.txt"
	storage := NewFileStorage(filePath)

	retrievedPath := storage.GetFilePath()
	if retrievedPath != filePath {
		t.Errorf("Expected file path %s, got %s", filePath, retrievedPath)
	}

	// Test with different path
	anotherPath := "/home/user/music/queue.txt"
	anotherStorage := NewFileStorage(anotherPath)

	retrievedAnotherPath := anotherStorage.GetFilePath()
	if retrievedAnotherPath != anotherPath {
		t.Errorf("Expected file path %s, got %s", anotherPath, retrievedAnotherPath)
	}
}
