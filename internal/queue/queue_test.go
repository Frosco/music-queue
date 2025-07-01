package queue

import (
	"os"
	"path/filepath"
	"strings"
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
