package queue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"music-queue/src/internal/storage"
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

	added, duplicates, formatErrors, err := queue.ImportAlbums(nonExistentFile)

	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	if added != 0 || duplicates != 0 || formatErrors != 0 {
		t.Errorf("Expected 0 added, 0 duplicates, 0 formatErrors for non-existent file, got added=%d, duplicates=%d, formatErrors=%d", added, duplicates, formatErrors)
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

	added, duplicates, formatErrors, err := queue.ImportAlbums(emptyImportFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error for empty file: %v", err)
	}

	if added != 0 || duplicates != 0 || formatErrors != 0 {
		t.Errorf("Expected 0 added, 0 duplicates, 0 formatErrors for empty file, got added=%d, duplicates=%d, formatErrors=%d", added, duplicates, formatErrors)
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

	added, duplicates, formatErrors, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	if added != 3 {
		t.Errorf("Expected 3 added, got %d", added)
	}

	if duplicates != 0 {
		t.Errorf("Expected 0 duplicates, got %d", duplicates)
	}

	if formatErrors != 0 {
		t.Errorf("Expected 0 formatErrors, got %d", formatErrors)
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
	existingContent := "Existing Artist - Existing Album\n"
	err := os.WriteFile(queueFile, []byte(existingContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create import file with new albums
	importContent := "Artist 1 - Album 1\nArtist 2 - Album 2\n"
	err = os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	added, duplicates, formatErrors, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	if added != 2 {
		t.Errorf("Expected 2 added, got %d", added)
	}

	if duplicates != 0 {
		t.Errorf("Expected 0 duplicates, got %d", duplicates)
	}

	if formatErrors != 0 {
		t.Errorf("Expected 0 formatErrors, got %d", formatErrors)
	}

	// Verify albums were added to existing queue
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	expected := []string{"Existing Artist - Existing Album", "Artist 1 - Album 1", "Artist 2 - Album 2"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines in queue, got %d", len(expected), len(lines))
	}

	for i, expectedAlbum := range expected {
		if i < len(lines) && lines[i] != expectedAlbum {
			t.Errorf("Line %d: expected %q, got %q", i, expectedAlbum, lines[i])
		}
	}
}

func TestQueueService_ImportAlbums_CaseInsensitiveDuplicates(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	importFile := filepath.Join(tempDir, "import.txt")

	// Create existing queue with mixed case
	existingContent := "Pink Floyd - Dark Side of the Moon\nThe Beatles - Abbey Road\n"
	err := os.WriteFile(queueFile, []byte(existingContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create import file with case variations and one new album
	importContent := "PINK FLOYD - DARK SIDE OF THE MOON\nthe beatles - abbey road\nLed Zeppelin - IV\n"
	err = os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	added, duplicates, formatErrors, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	if added != 1 {
		t.Errorf("Expected 1 added, got %d", added)
	}

	if duplicates != 2 {
		t.Errorf("Expected 2 duplicates, got %d", duplicates)
	}

	if formatErrors != 0 {
		t.Errorf("Expected 0 formatErrors, got %d", formatErrors)
	}
}

func TestQueueService_ImportAlbums_MalformedInput(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	importFile := filepath.Join(tempDir, "import.txt")

	// Create import file with some malformed entries
	importContent := "Pink Floyd - Dark Side of the Moon\n\n   \nThe Beatles - Abbey Road\n"
	err := os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	added, duplicates, formatErrors, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	if added != 2 {
		t.Errorf("Expected 2 added, got %d", added)
	}

	if duplicates != 0 {
		t.Errorf("Expected 0 duplicates, got %d", duplicates)
	}

	if formatErrors != 0 {
		t.Errorf("Expected 0 formatErrors, got %d", formatErrors)
	}

	// Verify only valid albums were added
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	expected := []string{"Pink Floyd - Dark Side of the Moon", "The Beatles - Abbey Road"}
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

	// Create import file with duplicates within the file itself
	importContent := "Pink Floyd - Dark Side of the Moon\nThe Beatles - Abbey Road\npink floyd - dark side of the moon\nLed Zeppelin - IV\n"
	err := os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	added, duplicates, formatErrors, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	if added != 3 {
		t.Errorf("Expected 3 added, got %d", added)
	}

	if duplicates != 1 {
		t.Errorf("Expected 1 duplicate, got %d", duplicates)
	}

	if formatErrors != 0 {
		t.Errorf("Expected 0 formatErrors, got %d", formatErrors)
	}

	// Verify correct albums were added (first occurrence wins)
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	expected := []string{"Pink Floyd - Dark Side of the Moon", "The Beatles - Abbey Road", "Led Zeppelin - IV"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines in queue, got %d", len(expected), len(lines))
	}

	for i, expectedAlbum := range expected {
		if i < len(lines) && lines[i] != expectedAlbum {
			t.Errorf("Line %d: expected %q, got %q", i, expectedAlbum, lines[i])
		}
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

	added, duplicates, formatErrors, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	// Should add 3 valid albums, skip 6 invalid ones
	if added != 3 {
		t.Errorf("Expected 3 added, got %d", added)
	}

	if duplicates != 0 {
		t.Errorf("Expected 0 duplicates, got %d", duplicates)
	}

	if formatErrors != 6 {
		t.Errorf("Expected 6 formatErrors, got %d", formatErrors)
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

	// Create existing queue
	existingContent := "Pink Floyd - The Dark Side of the Moon\n"
	err := os.WriteFile(queueFile, []byte(existingContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create import file with valid, invalid, and duplicate entries
	importContent := `Pink Floyd - The Dark Side of the Moon
Invalid Format Without Dash
The Beatles - Abbey Road
- Missing Artist Name
Led Zeppelin - IV
pink floyd - the dark side of the moon
`
	err = os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	added, duplicates, formatErrors, err := queue.ImportAlbums(importFile)

	if err != nil {
		t.Errorf("ImportAlbums returned error: %v", err)
	}

	// Should add 2 valid new albums, skip 2 duplicates, and 2 format errors
	if added != 2 {
		t.Errorf("Expected 2 added, got %d", added)
	}

	if duplicates != 2 {
		t.Errorf("Expected 2 duplicates, got %d", duplicates)
	}

	if formatErrors != 2 {
		t.Errorf("Expected 2 formatErrors, got %d", formatErrors)
	}

	// Verify correct albums were added
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

// TestQueueService_AddAlbum_Success tests successfully adding a new album
func TestQueueService_AddAlbum_Success(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	// Add first album to empty queue
	err := queue.AddAlbum("Pink Floyd - The Wall")
	if err != nil {
		t.Errorf("AddAlbum returned error: %v", err)
	}

	// Verify album was added
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	if len(lines) != 1 {
		t.Errorf("Expected 1 line in queue, got %d", len(lines))
	}

	if lines[0] != "Pink Floyd - The Wall" {
		t.Errorf("Expected 'Pink Floyd - The Wall', got %q", lines[0])
	}

	// Add second album to existing queue
	err = queue.AddAlbum("The Beatles - Abbey Road")
	if err != nil {
		t.Errorf("AddAlbum returned error: %v", err)
	}

	// Verify both albums are present
	lines, err = storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	expected := []string{"Pink Floyd - The Wall", "The Beatles - Abbey Road"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines in queue, got %d", len(expected), len(lines))
	}

	for i, expectedAlbum := range expected {
		if i < len(lines) && lines[i] != expectedAlbum {
			t.Errorf("Line %d: expected %q, got %q", i, expectedAlbum, lines[i])
		}
	}
}

// TestQueueService_AddAlbum_Duplicate tests adding a duplicate album (case-insensitive)
func TestQueueService_AddAlbum_Duplicate(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	// Add initial album
	err := queue.AddAlbum("Pink Floyd - The Wall")
	if err != nil {
		t.Errorf("AddAlbum returned error: %v", err)
	}

	// Try to add exact duplicate
	err = queue.AddAlbum("Pink Floyd - The Wall")
	if err == nil {
		t.Error("Expected error for duplicate album")
	}

	expectedErr := "album 'Pink Floyd - The Wall' already exists"
	if err == nil || err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got: %v", expectedErr, err)
	}

	// Try to add case-insensitive duplicate
	err = queue.AddAlbum("PINK FLOYD - THE WALL")
	if err == nil {
		t.Error("Expected error for case-insensitive duplicate album")
	}

	// The error message will reflect the input, not the original
	expectedErr = "album 'PINK FLOYD - THE WALL' already exists"
	if err == nil || err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got: %v", expectedErr, err)
	}

	// Try to add another case variation
	err = queue.AddAlbum("pink floyd - the wall")
	if err == nil {
		t.Error("Expected error for lowercase duplicate album")
	}

	// Verify only one album is in the queue
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	if len(lines) != 1 {
		t.Errorf("Expected 1 line in queue after duplicates, got %d", len(lines))
	}

	if lines[0] != "Pink Floyd - The Wall" {
		t.Errorf("Expected 'Pink Floyd - The Wall', got %q", lines[0])
	}
}

// TestQueueService_AddAlbum_InvalidFormat tests adding albums with invalid format
func TestQueueService_AddAlbum_InvalidFormat(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	testCases := []string{
		"",                     // empty string
		"No Dash Here",         // no dash
		"- Missing Artist",     // missing artist
		"Missing Album -",      // missing album
		"   - Whitespace Only", // whitespace before dash
		"Artist -   ",          // whitespace after dash
	}

	for _, invalidAlbum := range testCases {
		err := queue.AddAlbum(invalidAlbum)
		if err == nil {
			t.Errorf("Expected error for invalid album format: %q", invalidAlbum)
		}

		if !strings.Contains(err.Error(), "invalid album format") {
			t.Errorf("Expected 'invalid album format' error for %q, got: %v", invalidAlbum, err)
		}
	}

	// Verify no albums were added
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	if len(lines) != 0 {
		t.Errorf("Expected 0 lines in queue after invalid formats, got %d", len(lines))
	}
}

// TestQueueService_AddAlbum_WithExistingQueue tests adding to an existing queue
func TestQueueService_AddAlbum_WithExistingQueue(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Create existing queue
	storage := storage.NewFileStorage(queueFile)
	err := storage.WriteLines([]string{"Existing Album 1 - Title 1", "Existing Album 2 - Title 2"})
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)

	// Add new album
	err = queue.AddAlbum("New Artist - New Album")
	if err != nil {
		t.Errorf("AddAlbum returned error: %v", err)
	}

	// Verify album was appended
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	expected := []string{"Existing Album 1 - Title 1", "Existing Album 2 - Title 2", "New Artist - New Album"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d lines in queue, got %d", len(expected), len(lines))
	}

	for i, expectedAlbum := range expected {
		if i < len(lines) && lines[i] != expectedAlbum {
			t.Errorf("Line %d: expected %q, got %q", i, expectedAlbum, lines[i])
		}
	}

	// Try to add duplicate of existing album
	err = queue.AddAlbum("existing album 1 - title 1") // case insensitive
	if err == nil {
		t.Error("Expected error for duplicate of existing album")
	}

	expectedErr := "album 'existing album 1 - title 1' already exists"
	if err == nil || err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got: %v", expectedErr, err)
	}

	// Verify queue wasn't modified
	lines, err = storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	if len(lines) != len(expected) {
		t.Errorf("Expected queue to remain unchanged after duplicate, got %d lines", len(lines))
	}
}

// TestQueueService_AddAlbum_WhitespaceHandling tests proper whitespace handling
func TestQueueService_AddAlbum_WhitespaceHandling(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	// Add album with extra whitespace
	err := queue.AddAlbum("  Pink Floyd  -  The Wall  ")
	if err != nil {
		t.Errorf("AddAlbum returned error: %v", err)
	}

	// Verify album was trimmed and stored correctly
	lines, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue: %v", err)
	}

	if len(lines) != 1 {
		t.Errorf("Expected 1 line in queue, got %d", len(lines))
	}

	// Note: The exact whitespace handling depends on the implementation
	// The album should be stored with trimmed overall whitespace
	expected := "Pink Floyd  -  The Wall" // Whitespace around dash preserved, but overall trimmed
	if lines[0] != expected {
		t.Errorf("Expected %q, got %q", expected, lines[0])
	}
}

func TestQueueService_GetNextAlbum_Success(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Create queue with test albums
	queueStorage := storage.NewFileStorage(queueFile)
	testAlbums := []string{"Artist 1 - Album 1", "Artist 2 - Album 2", "Artist 3 - Album 3"}
	err := queueStorage.WriteLines(testAlbums)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(queueStorage)

	// Get next album
	selectedAlbum, err := queue.GetNextAlbum()
	if err != nil {
		t.Errorf("GetNextAlbum returned error: %v", err)
	}

	// Verify selected album is one of the test albums
	found := false
	for _, album := range testAlbums {
		if album == selectedAlbum {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Selected album %q not found in original list", selectedAlbum)
	}

	// Verify queue size decreased by 1
	remainingAlbums, err := queueStorage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read remaining albums: %v", err)
	}

	if len(remainingAlbums) != len(testAlbums)-1 {
		t.Errorf("Expected %d remaining albums, got %d", len(testAlbums)-1, len(remainingAlbums))
	}

	// Verify selected album was removed
	for _, remainingAlbum := range remainingAlbums {
		if remainingAlbum == selectedAlbum {
			t.Errorf("Selected album %q was not removed from queue", selectedAlbum)
		}
	}
}

func TestQueueService_GetNextAlbum_EmptyQueue(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Create empty queue
	storage := storage.NewFileStorage(queueFile)
	err := storage.WriteLines([]string{})
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)

	// Get next album from empty queue
	selectedAlbum, err := queue.GetNextAlbum()
	if err == nil {
		t.Error("Expected error for empty queue")
	}

	if selectedAlbum != "" {
		t.Errorf("Expected empty album string for empty queue, got %q", selectedAlbum)
	}

	// Check error message
	if !strings.Contains(err.Error(), "queue is empty") {
		t.Errorf("Expected 'queue is empty' error message, got: %v", err)
	}
}

func TestQueueService_GetNextAlbum_SingleAlbum(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Create queue with single album
	storage := storage.NewFileStorage(queueFile)
	testAlbum := "Pink Floyd - The Wall"
	err := storage.WriteLines([]string{testAlbum})
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)

	// Get next album
	selectedAlbum, err := queue.GetNextAlbum()
	if err != nil {
		t.Errorf("GetNextAlbum returned error: %v", err)
	}

	// Verify correct album was selected
	if selectedAlbum != testAlbum {
		t.Errorf("Expected %q, got %q", testAlbum, selectedAlbum)
	}

	// Verify queue is now empty
	remainingAlbums, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read remaining albums: %v", err)
	}

	if len(remainingAlbums) != 0 {
		t.Errorf("Expected empty queue after selecting last album, got %d albums", len(remainingAlbums))
	}
}

func TestQueueService_GetNextAlbum_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "nonexistent.txt")

	// Use non-existent file (ReadLines should return empty slice, not error)
	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	// Get next album from non-existent file
	selectedAlbum, err := queue.GetNextAlbum()
	if err == nil {
		t.Error("Expected error for non-existent queue file")
	}

	if selectedAlbum != "" {
		t.Errorf("Expected empty album string for non-existent file, got %q", selectedAlbum)
	}

	// Should get "queue is empty" error since ReadLines returns empty slice for non-existent files
	if !strings.Contains(err.Error(), "queue is empty") {
		t.Errorf("Expected 'queue is empty' error message, got: %v", err)
	}
}

func TestQueueService_ListAlbums_EmptyQueue(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	albums, err := queue.ListAlbums()

	if err != nil {
		t.Errorf("ListAlbums returned error for empty queue: %v", err)
	}

	if len(albums) != 0 {
		t.Errorf("Expected empty slice for empty queue, got %d albums", len(albums))
	}
}

func TestQueueService_ListAlbums_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "nonexistent.txt")

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	albums, err := queue.ListAlbums()

	if err != nil {
		t.Errorf("ListAlbums returned error for non-existent file: %v", err)
	}

	if len(albums) != 0 {
		t.Errorf("Expected empty slice for non-existent file, got %d albums", len(albums))
	}
}

func TestQueueService_ListAlbums_WithAlbums(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Create queue with albums
	storage := storage.NewFileStorage(queueFile)
	expectedAlbums := []string{
		"Pink Floyd - Dark Side of the Moon",
		"The Beatles - Abbey Road",
		"Led Zeppelin - IV",
		"Queen - A Night at the Opera",
	}
	err := storage.WriteLines(expectedAlbums)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)

	albums, err := queue.ListAlbums()

	if err != nil {
		t.Errorf("ListAlbums returned error: %v", err)
	}

	if len(albums) != len(expectedAlbums) {
		t.Errorf("Expected %d albums, got %d", len(expectedAlbums), len(albums))
	}

	// Verify albums are returned in the same order
	for i, expected := range expectedAlbums {
		if i < len(albums) && albums[i] != expected {
			t.Errorf("Album %d: expected %q, got %q", i, expected, albums[i])
		}
	}
}

func TestQueueService_ListAlbums_WithSingleAlbum(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Create queue with single album
	storage := storage.NewFileStorage(queueFile)
	expectedAlbums := []string{"Pink Floyd - Dark Side of the Moon"}
	err := storage.WriteLines(expectedAlbums)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)

	albums, err := queue.ListAlbums()

	if err != nil {
		t.Errorf("ListAlbums returned error: %v", err)
	}

	if len(albums) != 1 {
		t.Errorf("Expected 1 album, got %d", len(albums))
	}

	if albums[0] != expectedAlbums[0] {
		t.Errorf("Expected %q, got %q", expectedAlbums[0], albums[0])
	}
}

func TestQueueService_CountAlbums_EmptyQueue(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	count, err := queue.CountAlbums()
	if err != nil {
		t.Errorf("CountAlbums returned error for empty queue: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 albums in empty queue, got %d", count)
	}
}

func TestQueueService_CountAlbums_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "nonexistent.txt")

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	count, err := queue.CountAlbums()
	if err != nil {
		t.Errorf("CountAlbums returned error for non-existent file: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 albums for non-existent file, got %d", count)
	}
}

func TestQueueService_CountAlbums_WithAlbums(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	// Add multiple albums
	albums := []string{
		"Artist 1 - Album 1",
		"Artist 2 - Album 2",
		"Artist 3 - Album 3",
	}

	for _, album := range albums {
		err := queue.AddAlbum(album)
		if err != nil {
			t.Fatalf("AddAlbum failed: %v", err)
		}
	}

	count, err := queue.CountAlbums()
	if err != nil {
		t.Errorf("CountAlbums returned error: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 albums, got %d", count)
	}
}

func TestQueueService_CountAlbums_WithSingleAlbum(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	storage := storage.NewFileStorage(queueFile)
	queue := NewQueue(storage)

	// Add one album
	err := queue.AddAlbum("Single Artist - Single Album")
	if err != nil {
		t.Fatalf("AddAlbum failed: %v", err)
	}

	count, err := queue.CountAlbums()
	if err != nil {
		t.Errorf("CountAlbums returned error: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 album, got %d", count)
	}
}

// Archive functionality tests

func TestQueueService_GetNextAlbum_ArchivesAlbum(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	archiveFile := filepath.Join(tempDir, "archive.txt")

	// Create queue with test albums
	storage := storage.NewFileStorage(queueFile)
	testAlbums := []string{"Artist 1 - Album 1", "Artist 2 - Album 2", "Artist 3 - Album 3"}
	err := storage.WriteLines(testAlbums)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)

	// Get next album
	selectedAlbum, err := queue.GetNextAlbum()
	if err != nil {
		t.Errorf("GetNextAlbum returned error: %v", err)
	}

	// Verify selected album is one of the test albums
	found := false
	for _, album := range testAlbums {
		if album == selectedAlbum {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Selected album %q not found in original list", selectedAlbum)
	}

	// Verify queue size decreased by 1
	remainingAlbums, err := storage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read remaining albums: %v", err)
	}

	if len(remainingAlbums) != len(testAlbums)-1 {
		t.Errorf("Expected %d remaining albums, got %d", len(testAlbums)-1, len(remainingAlbums))
	}

	// Verify selected album was removed from queue
	for _, remainingAlbum := range remainingAlbums {
		if remainingAlbum == selectedAlbum {
			t.Errorf("Selected album %q was not removed from queue", selectedAlbum)
		}
	}

	// Verify album was added to archive
	archiveStorageInstance := storage.NewFileStorage(archiveFile)
	archivedAlbums, err := archiveStorageInstance.ReadLines()
	if err != nil {
		t.Errorf("Failed to read archive: %v", err)
	}

	if len(archivedAlbums) != 1 {
		t.Errorf("Expected 1 album in archive, got %d", len(archivedAlbums))
	}

	if archivedAlbums[0] != selectedAlbum {
		t.Errorf("Expected archived album %q, got %q", selectedAlbum, archivedAlbums[0])
	}
}

func TestQueueService_GetNextAlbum_ArchivesMultipleAlbums(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	archiveFile := filepath.Join(tempDir, "archive.txt")

	// Create queue with test albums
	queueStorage := storage.NewFileStorage(queueFile)
	testAlbums := []string{"Artist 1 - Album 1", "Artist 2 - Album 2", "Artist 3 - Album 3"}
	err := queueStorage.WriteLines(testAlbums)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(queueStorage)

	// Get next album multiple times
	selectedAlbums := make([]string, 0)
	for i := 0; i < 2; i++ {
		selectedAlbum, err := queue.GetNextAlbum()
		if err != nil {
			t.Errorf("GetNextAlbum returned error on iteration %d: %v", i, err)
		}
		selectedAlbums = append(selectedAlbums, selectedAlbum)
	}

	// Verify queue size decreased by 2
	remainingAlbums, err := queueStorage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read remaining albums: %v", err)
	}

	if len(remainingAlbums) != len(testAlbums)-2 {
		t.Errorf("Expected %d remaining albums, got %d", len(testAlbums)-2, len(remainingAlbums))
	}

	// Verify archive contains both selected albums
	archiveStorageInstance := storage.NewFileStorage(archiveFile)
	archivedAlbums, err := archiveStorageInstance.ReadLines()
	if err != nil {
		t.Errorf("Failed to read archive: %v", err)
	}

	if len(archivedAlbums) != 2 {
		t.Errorf("Expected 2 albums in archive, got %d", len(archivedAlbums))
	}

	// Verify both selected albums are in archive
	for _, selectedAlbum := range selectedAlbums {
		found := false
		for _, archivedAlbum := range archivedAlbums {
			if archivedAlbum == selectedAlbum {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Selected album %q not found in archive", selectedAlbum)
		}
	}
}

func TestQueueService_GetNextAlbum_ArchiveFileCreatedInSameDirectory(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "custom_queue.txt")
	expectedArchiveFile := filepath.Join(tempDir, "custom_queue_archive.txt")

	// Create queue with custom filename
	storage := storage.NewFileStorage(queueFile)
	testAlbums := []string{"Artist 1 - Album 1"}
	err := storage.WriteLines(testAlbums)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)

	// Get next album
	_, err = queue.GetNextAlbum()
	if err != nil {
		t.Errorf("GetNextAlbum returned error: %v", err)
	}

	// Verify archive file was created in the same directory
	if _, err := os.Stat(expectedArchiveFile); os.IsNotExist(err) {
		t.Errorf("Archive file was not created at expected location: %s", expectedArchiveFile)
	}

	// Verify archive contains the album
	archiveStorageInstance := storage.NewFileStorage(expectedArchiveFile)
	archivedAlbums, err := archiveStorageInstance.ReadLines()
	if err != nil {
		t.Errorf("Failed to read archive: %v", err)
	}

	if len(archivedAlbums) != 1 {
		t.Errorf("Expected 1 album in archive, got %d", len(archivedAlbums))
	}
}

func TestQueueService_GetNextAlbum_ArchiveFileCreatedInNestedDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "subdir")
	queueFile := filepath.Join(nestedDir, "queue.txt")
	expectedArchiveFile := filepath.Join(nestedDir, "archive.txt")

	// Create queue in nested directory
	queueStorage := storage.NewFileStorage(queueFile)
	testAlbums := []string{"Artist 1 - Album 1"}
	err := queueStorage.WriteLines(testAlbums)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(queueStorage)

	// Get next album
	_, err = queue.GetNextAlbum()
	if err != nil {
		t.Errorf("GetNextAlbum returned error: %v", err)
	}

	// Verify archive file was created in the same nested directory
	if _, err := os.Stat(expectedArchiveFile); os.IsNotExist(err) {
		t.Errorf("Archive file was not created at expected location: %s", expectedArchiveFile)
	}
}

func TestQueueService_GetNextAlbum_ArchiveAppendsToExistingArchive(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	archiveFile := filepath.Join(tempDir, "archive.txt")

	// Create existing archive with some albums
	archiveStorage := storage.NewFileStorage(archiveFile)
	existingArchived := []string{"Previously Archived - Album 1", "Previously Archived - Album 2"}
	err := archiveStorage.WriteLines(existingArchived)
	if err != nil {
		t.Fatal(err)
	}

	// Create queue with new albums
	storage := storage.NewFileStorage(queueFile)
	testAlbums := []string{"Artist 1 - Album 1", "Artist 2 - Album 2"}
	err = storage.WriteLines(testAlbums)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)

	// Get next album
	selectedAlbum, err := queue.GetNextAlbum()
	if err != nil {
		t.Errorf("GetNextAlbum returned error: %v", err)
	}

	// Verify archive contains both existing and new albums
	archivedAlbums, err := archiveStorage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read archive: %v", err)
	}

	expectedArchived := append(existingArchived, selectedAlbum)
	if len(archivedAlbums) != len(expectedArchived) {
		t.Errorf("Expected %d albums in archive, got %d", len(expectedArchived), len(archivedAlbums))
	}

	// Verify all expected albums are in archive
	for i, expected := range expectedArchived {
		if i < len(archivedAlbums) && archivedAlbums[i] != expected {
			t.Errorf("Archive line %d: expected %q, got %q", i, expected, archivedAlbums[i])
		}
	}
}

func TestQueueService_GetNextAlbum_ArchiveHandlesEmptyArchiveFile(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	archiveFile := filepath.Join(tempDir, "archive.txt")

	// Create empty archive file
	err := os.WriteFile(archiveFile, []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create queue with test albums
	storage := storage.NewFileStorage(queueFile)
	testAlbums := []string{"Artist 1 - Album 1"}
	err = storage.WriteLines(testAlbums)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)

	// Get next album
	selectedAlbum, err := queue.GetNextAlbum()
	if err != nil {
		t.Errorf("GetNextAlbum returned error: %v", err)
	}

	// Verify archive contains the album
	archiveStorageInstance := storage.NewFileStorage(archiveFile)
	archivedAlbums, err := archiveStorageInstance.ReadLines()
	if err != nil {
		t.Errorf("Failed to read archive: %v", err)
	}

	if len(archivedAlbums) != 1 {
		t.Errorf("Expected 1 album in archive, got %d", len(archivedAlbums))
	}

	if archivedAlbums[0] != selectedAlbum {
		t.Errorf("Expected archived album %q, got %q", selectedAlbum, archivedAlbums[0])
	}
}

func TestQueueService_GetNextAlbum_ArchiveErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Create queue with test albums
	storage := storage.NewFileStorage(queueFile)
	testAlbums := []string{"Artist 1 - Album 1"}
	err := storage.WriteLines(testAlbums)
	if err != nil {
		t.Fatal(err)
	}

	queue := NewQueue(storage)

	// Create a read-only directory to cause archive write failure
	archiveDir := filepath.Join(tempDir, "readonly")
	err = os.MkdirAll(archiveDir, 0444) // read-only
	if err != nil {
		t.Fatal(err)
	}

	// Temporarily change the queue to use a file in the read-only directory
	// This will cause the archive to be created in the read-only directory
	readonlyQueueFile := filepath.Join(archiveDir, "queue.txt")
	readonlyStorage := storage.NewFileStorage(readonlyQueueFile)
	err = readonlyStorage.WriteLines(testAlbums)
	if err != nil {
		t.Fatal(err)
	}

	readonlyQueue := NewQueue(readonlyStorage)

	// Get next album should fail due to archive write error
	_, err = readonlyQueue.GetNextAlbum()
	if err == nil {
		t.Error("Expected error when archive cannot be written")
	}

	if !strings.Contains(err.Error(), "failed to archive album") {
		t.Errorf("Expected archive error message, got: %v", err)
	}

	// Verify queue was not modified (transaction-like behavior)
	remainingAlbums, err := readonlyStorage.ReadLines()
	if err != nil {
		t.Errorf("Failed to read queue after error: %v", err)
	}

	if len(remainingAlbums) != 1 {
		t.Errorf("Expected queue to remain unchanged after archive error, got %d albums", len(remainingAlbums))
	}
}
