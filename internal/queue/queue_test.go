package queue

import (
	"os"
	"path/filepath"
	"testing"

	"music-queue/internal/storage"
)

func TestNewQueue(t *testing.T) {
	storage := storage.NewFileStorage("/tmp/test.txt")
	queue := NewQueue(storage)

	if queue == nil {
		t.Fatal("NewQueue returned nil")
	}

	if queue.storage != storage {
		t.Error("NewQueue did not set storage correctly")
	}
}

func TestGetDefaultQueuePath(t *testing.T) {
	path := GetDefaultQueuePath()

	if path == "" {
		t.Error("GetDefaultQueuePath returned empty string")
	}

	// Should end with the queue file name
	if filepath.Base(path) != "queue.txt" {
		t.Errorf("Expected path to end with queue.txt, got: %s", path)
	}
}

func TestQueueService_ImportAlbums_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	added, skipped, err := queue.ImportAlbums(nonExistentFile)

	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	if added != 0 || skipped != 0 {
		t.Errorf("Expected 0 added and 0 skipped for non-existent file, got added=%d, skipped=%d", added, skipped)
	}
}

func TestQueueService_ImportAlbums_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	emptyImportFile := filepath.Join(tempDir, "empty.txt")

	// Create empty import file
	err := os.WriteFile(emptyImportFile, []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	added, skipped, err := queue.ImportAlbums(emptyImportFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error for empty file: %v", err)
	}

	if added != 0 || skipped != 0 {
		t.Errorf("Expected 0 added and 0 skipped for empty file, got added=%d, skipped=%d", added, skipped)
	}
}

func TestQueueService_ImportAlbums_NewAlbums(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	importFile := filepath.Join(tempDir, "import.txt")

	// Create import file with new albums
	importContent := "Album 1\nAlbum 2\nAlbum 3\n"
	err := os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	added, skipped, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	if added != 3 {
		t.Errorf("Expected 3 added, got %d", added)
	}

	if skipped != 0 {
		t.Errorf("Expected 0 skipped, got %d", skipped)
	}

	// Verify albums were saved
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	expected := []string{"Album 1", "Album 2", "Album 3"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines in queue, got %d", len(expected), len(lines))
	}

	for i, expectedAlbum := range expected {
		if i < len(lines) && lines[i] != expectedAlbum {
			t.Errorf("Line %d: expected %q, got %q", i, expectedAlbum, lines[i])
		}
	}
}

func TestQueueService_ImportAlbums_WithExistingQueue(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	importFile := filepath.Join(tempDir, "import.txt")

	// Create existing queue
	storage := storage.NewFileStorage(queueFile)
	err := storage.WriteLines([]string{"Existing Album 1", "Existing Album 2"})
	if err != nil {
		t.Fatal(err)
	}

	// Create import file with mix of new and existing albums
	importContent := "Album 1\nExisting Album 1\nAlbum 2\n"
	err = os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)
	added, skipped, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	if added != 2 {
		t.Errorf("Expected 2 added, got %d", added)
	}

	if skipped != 1 {
		t.Errorf("Expected 1 skipped, got %d", skipped)
	}

	// Verify final queue content
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	expected := []string{"Existing Album 1", "Existing Album 2", "Album 1", "Album 2"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines in queue, got %d", len(expected), len(lines))
	}
}

func TestQueueService_ImportAlbums_CaseInsensitiveDuplicates(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	importFile := filepath.Join(tempDir, "import.txt")

	// Create existing queue with mixed case
	storage := storage.NewFileStorage(queueFile)
	err := storage.WriteLines([]string{"Dark Side of the Moon", "Abbey Road"})
	if err != nil {
		t.Fatal(err)
	}

	// Create import file with case variations
	importContent := "DARK SIDE OF THE MOON\nabbey road\nThe Wall\ndark side of the moon\n"
	err = os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)
	added, skipped, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	if added != 1 {
		t.Errorf("Expected 1 added, got %d", added)
	}

	if skipped != 3 {
		t.Errorf("Expected 3 skipped, got %d", skipped)
	}
}

func TestQueueService_ImportAlbums_MalformedInput(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	importFile := filepath.Join(tempDir, "import.txt")

	// Create import file with empty lines and whitespace
	importContent := "Album 1\n\n  \n\t\nAlbum 2\n   Album 3   \n\n"
	err := os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	added, skipped, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	if added != 3 {
		t.Errorf("Expected 3 added, got %d", added)
	}

	if skipped != 0 {
		t.Errorf("Expected 0 skipped, got %d", skipped)
	}

	// Verify trimmed content
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	expected := []string{"Album 1", "Album 2", "Album 3"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines in queue, got %d", len(expected), len(lines))
	}

	for i, expectedAlbum := range expected {
		if i < len(lines) && lines[i] != expectedAlbum {
			t.Errorf("Line %d: expected %q, got %q", i, expectedAlbum, lines[i])
		}
	}
}

func TestQueueService_ImportAlbums_DuplicatesWithinImportFile(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	importFile := filepath.Join(tempDir, "import.txt")

	// Create import file with duplicates within the same file
	importContent := "Album 1\nAlbum 2\nALBUM 1\nalbum 2\nAlbum 3\n"
	err := os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	added, skipped, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	if added != 3 {
		t.Errorf("Expected 3 added, got %d", added)
	}

	if skipped != 2 {
		t.Errorf("Expected 2 skipped, got %d", skipped)
	}

	// Verify no duplicates in final queue
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	expected := []string{"Album 1", "Album 2", "Album 3"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines in queue, got %d", len(expected), len(lines))
	}
}
