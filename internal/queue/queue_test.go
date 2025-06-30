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
	importContent := "Artist 1 - Album 1\nArtist 2 - Album 2\nArtist 3 - Album 3\n"
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

	expected := []string{"Artist 1 - Album 1", "Artist 2 - Album 2", "Artist 3 - Album 3"}
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
	err := storage.WriteLines([]string{"Existing Artist 1 - Existing Album 1", "Existing Artist 2 - Existing Album 2"})
	if err != nil {
		t.Fatal(err)
	}

	// Create import file with mix of new and existing albums
	importContent := "Artist 1 - Album 1\nExisting Artist 1 - Existing Album 1\nArtist 2 - Album 2\n"
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

	expected := []string{"Existing Artist 1 - Existing Album 1", "Existing Artist 2 - Existing Album 2", "Artist 1 - Album 1", "Artist 2 - Album 2"}
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
	err := storage.WriteLines([]string{"Pink Floyd - Dark Side of the Moon", "The Beatles - Abbey Road"})
	if err != nil {
		t.Fatal(err)
	}

	// Create import file with case variations
	importContent := "PINK FLOYD - DARK SIDE OF THE MOON\nthe beatles - abbey road\nPink Floyd - The Wall\npink floyd - dark side of the moon\n"
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
	importContent := "Artist 1 - Album 1\n\n  \n\t\nArtist 2 - Album 2\n   Artist 3 - Album 3   \n\n"
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

	expected := []string{"Artist 1 - Album 1", "Artist 2 - Album 2", "Artist 3 - Album 3"}
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
	importContent := "Artist 1 - Album 1\nArtist 2 - Album 2\nARTIST 1 - ALBUM 1\nartist 2 - album 2\nArtist 3 - Album 3\n"
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

	expected := []string{"Artist 1 - Album 1", "Artist 2 - Album 2", "Artist 3 - Album 3"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines in queue, got %d", len(expected), len(lines))
	}
}

func TestValidateAlbumFormat(t *testing.T) {
	tests := []struct {
		name     string
		album    string
		expected bool
	}{
		// Valid formats
		{"valid simple", "Artist - Album", true},
		{"valid with spaces", "Artist Name - Album Title", true},
		{"valid with extra spaces", "  Artist Name  -  Album Title  ", true},
		{"valid with special chars", "Artist & Band - Album: Subtitle", true},
		{"valid with numbers", "Artist 123 - Album 456", true},
		{"valid with multiple words", "Pink Floyd - The Dark Side of the Moon", true},

		// Invalid formats
		{"no dash", "Artist Album", false},
		{"dash at start", "- Album Title", false},
		{"dash at end", "Artist Name -", false},
		{"only dash", "-", false},
		{"empty before dash", " - Album Title", false},
		{"empty after dash", "Artist Name - ", false},
		{"whitespace only before dash", "   - Album Title", false},
		{"whitespace only after dash", "Artist Name -   ", false},
		{"empty string", "", false},
		{"whitespace only", "   ", false},
		{"tab only", "\t", false},
		{"newline only", "\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateAlbumFormat(tt.album)
			if result != tt.expected {
				t.Errorf("validateAlbumFormat(%q) = %v, want %v", tt.album, result, tt.expected)
			}
		})
	}
}

func TestQueueService_ImportAlbums_FormatValidation(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	importFile := filepath.Join(tempDir, "import.txt")

	// Create import file with mix of valid and invalid formats
	importContent := `Pink Floyd - The Dark Side of the Moon
Invalid Format
The Beatles - Abbey Road
No Dash Here
Led Zeppelin - IV
- Missing Artist
Artist Missing -
   - Whitespace Before Dash
Artist -   
`
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

	// Should add 3 valid albums, skip 6 invalid ones
	if added != 3 {
		t.Errorf("Expected 3 added, got %d", added)
	}

	if skipped != 6 {
		t.Errorf("Expected 6 skipped, got %d", skipped)
	}

	// Verify only valid albums were added
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	expected := []string{
		"Pink Floyd - The Dark Side of the Moon",
		"The Beatles - Abbey Road",
		"Led Zeppelin - IV",
	}

	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines in queue, got %d", len(expected), len(lines))
	}

	for i, expectedAlbum := range expected {
		if i < len(lines) && lines[i] != expectedAlbum {
			t.Errorf("Line %d: expected %q, got %q", i, expectedAlbum, lines[i])
		}
	}
}

func TestQueueService_ImportAlbums_FormatValidationWithDuplicates(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	importFile := filepath.Join(tempDir, "import.txt")

	// Create existing queue with valid format
	storage := storage.NewFileStorage(queueFile)
	err := storage.WriteLines([]string{"Pink Floyd - The Wall", "The Beatles - Let It Be"})
	if err != nil {
		t.Fatal(err)
	}

	// Create import file with mix of valid/invalid formats and duplicates
	importContent := `Pink Floyd - The Dark Side of the Moon
Invalid Format
The Beatles - Let It Be
No Dash Here
Pink Floyd - The Wall
- Missing Artist
`
	err = os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)
	added, skipped, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	// Should add 1 new valid album, skip 5 (2 duplicates + 3 invalid format)
	if added != 1 {
		t.Errorf("Expected 1 added, got %d", added)
	}

	if skipped != 5 {
		t.Errorf("Expected 5 skipped, got %d", skipped)
	}

	// Verify final queue content
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	expected := []string{
		"Pink Floyd - The Wall",
		"The Beatles - Let It Be",
		"Pink Floyd - The Dark Side of the Moon",
	}

	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines in queue, got %d", len(expected), len(lines))
	}
}
