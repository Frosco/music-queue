package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLI_Import_Success tests successful album import
func TestCLI_Import_Success(t *testing.T) {
	tempDir := t.TempDir()

	// Create import file
	importFile := filepath.Join(tempDir, "albums.txt")
	importContent := "Pink Floyd - Dark Side of the Moon\nThe Beatles - Abbey Road\nPink Floyd - The Wall\n"
	err := os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create queue file path
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Build and run the CLI
	cmd := exec.Command("go", "run", "main.go", "import", "--queue", queueFile, importFile)
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check output contains expected success message
	if !strings.Contains(outputStr, "Added 3 albums, Skipped 0 duplicates") {
		t.Errorf("Expected success message not found. Output: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Import complete!") {
		t.Errorf("Expected completion message not found. Output: %s", outputStr)
	}

	// Verify queue file was created with correct content
	queueContent, err := os.ReadFile(queueFile)
	if err != nil {
		t.Fatalf("Failed to read queue file: %v", err)
	}

	expectedAlbums := []string{"Pink Floyd - Dark Side of the Moon", "The Beatles - Abbey Road", "Pink Floyd - The Wall"}
	queueLines := strings.Split(strings.TrimSpace(string(queueContent)), "\n")

	if len(queueLines) != len(expectedAlbums) {
		t.Errorf("Expected %d albums in queue, got %d", len(expectedAlbums), len(queueLines))
	}

	for i, expected := range expectedAlbums {
		if i < len(queueLines) && queueLines[i] != expected {
			t.Errorf("Album %d: expected %q, got %q", i, expected, queueLines[i])
		}
	}
}

// TestCLI_Import_WithDuplicates tests import with existing queue and duplicates
func TestCLI_Import_WithDuplicates(t *testing.T) {
	tempDir := t.TempDir()

	// Create existing queue
	queueFile := filepath.Join(tempDir, "queue.txt")
	existingContent := "Pink Floyd - Dark Side of the Moon\nPink Floyd - Wish You Were Here\n"
	err := os.WriteFile(queueFile, []byte(existingContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create import file with duplicates
	importContent := "PINK FLOYD - DARK SIDE OF THE MOON\nThe Beatles - Abbey Road\npink floyd - wish you were here\nPink Floyd - The Wall\n"
	importFile := filepath.Join(tempDir, "albums.txt")
	err = os.WriteFile(importFile, []byte(importContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Build and run the CLI
	cmd := exec.Command("go", "run", "main.go", "import", "--queue", queueFile, importFile)
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check output shows correct counts
	if !strings.Contains(outputStr, "Added 2 albums, Skipped 2 duplicates") {
		t.Errorf("Expected duplicate handling message not found. Output: %s", outputStr)
	}
}

// TestCLI_Import_FileNotFound tests error handling for non-existent import file
func TestCLI_Import_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

	// Build and run the CLI
	cmd := exec.Command("go", "run", "main.go", "import", "--queue", queueFile, nonExistentFile)
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()

	// Should exit with non-zero code
	if err == nil {
		t.Error("Expected CLI to fail for non-existent file")
	}

	outputStr := string(output)

	// Check error message
	if !strings.Contains(outputStr, "not found") {
		t.Errorf("Expected 'not found' error message. Output: %s", outputStr)
	}
}

// TestCLI_Import_EmptyFile tests handling of empty import files
func TestCLI_Import_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Create empty import file
	emptyFile := filepath.Join(tempDir, "empty.txt")
	err := os.WriteFile(emptyFile, []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Build and run the CLI
	cmd := exec.Command("go", "run", "main.go", "import", "--queue", queueFile, emptyFile)
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check output handles empty file gracefully
	if !strings.Contains(outputStr, "No albums found") {
		t.Errorf("Expected empty file message. Output: %s", outputStr)
	}
}

// TestCLI_Import_MissingArguments tests error handling for missing arguments
func TestCLI_Import_MissingArguments(t *testing.T) {
	// Build and run the CLI without import file
	cmd := exec.Command("go", "run", "main.go", "import")
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()

	// Should exit with non-zero code
	if err == nil {
		t.Error("Expected CLI to fail for missing import file")
	}

	outputStr := string(output)

	// Check error message
	if !strings.Contains(outputStr, "Import file not specified") {
		t.Errorf("Expected missing argument error message. Output: %s", outputStr)
	}
}

// TestCLI_Help tests the help command
func TestCLI_Help(t *testing.T) {
	// Test help command
	cmd := exec.Command("go", "run", "main.go", "help")
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Help command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check help content
	if !strings.Contains(outputStr, "Go Music Queue") {
		t.Errorf("Expected help title. Output: %s", outputStr)
	}

	if !strings.Contains(outputStr, "add") {
		t.Errorf("Expected add command in help. Output: %s", outputStr)
	}

	if !strings.Contains(outputStr, "import") {
		t.Errorf("Expected import command in help. Output: %s", outputStr)
	}
}

// TestCLI_Import_Help tests the import help command
func TestCLI_Import_Help(t *testing.T) {
	// Test import help command
	cmd := exec.Command("go", "run", "main.go", "import", "--help")
	cmd.Dir = "." // Run from cmd/queue directory

	output, _ := cmd.CombinedOutput()

	// --help should exit with code 0 for flag package
	outputStr := string(output)

	// Check import-specific help content
	if !strings.Contains(outputStr, "Import albums from a text file") {
		t.Errorf("Expected import description in help. Output: %s", outputStr)
	}

	if !strings.Contains(outputStr, "--queue") {
		t.Errorf("Expected queue flag in help. Output: %s", outputStr)
	}
}

// TestCLI_UnknownCommand tests error handling for unknown commands
func TestCLI_UnknownCommand(t *testing.T) {
	// Build and run the CLI with unknown command
	cmd := exec.Command("go", "run", "main.go", "unknown")
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()

	// Should exit with non-zero code
	if err == nil {
		t.Error("Expected CLI to fail for unknown command")
	}

	outputStr := string(output)

	// Check error message
	if !strings.Contains(outputStr, "Unknown command") {
		t.Errorf("Expected unknown command error message. Output: %s", outputStr)
	}
}

// TestCLI_Add_Success tests successfully adding a single album
func TestCLI_Add_Success(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Add first album to empty queue
	cmd := exec.Command("go", "run", "main.go", "add", "--queue", queueFile, "The Beatles - Abbey Road")
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check success message
	if !strings.Contains(outputStr, "Successfully added album: 'The Beatles - Abbey Road'") {
		t.Errorf("Expected success message not found. Output: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Queue saved to:") {
		t.Errorf("Expected queue location message not found. Output: %s", outputStr)
	}

	// Verify queue file was created with correct content
	queueContent, err := os.ReadFile(queueFile)
	if err != nil {
		t.Fatalf("Failed to read queue file: %v", err)
	}

	queueLines := strings.Split(strings.TrimSpace(string(queueContent)), "\n")
	if len(queueLines) != 1 {
		t.Errorf("Expected 1 album in queue, got %d", len(queueLines))
	}

	if queueLines[0] != "The Beatles - Abbey Road" {
		t.Errorf("Expected 'The Beatles - Abbey Road', got %q", queueLines[0])
	}

	// Add second album to existing queue
	cmd = exec.Command("go", "run", "main.go", "add", "--queue", queueFile, "Pink Floyd - The Wall")
	cmd.Dir = "."

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, output)
	}

	outputStr = string(output)

	// Check success message for second album
	if !strings.Contains(outputStr, "Successfully added album: 'Pink Floyd - The Wall'") {
		t.Errorf("Expected success message for second album not found. Output: %s", outputStr)
	}

	// Verify both albums are in queue
	queueContent, err = os.ReadFile(queueFile)
	if err != nil {
		t.Fatalf("Failed to read queue file: %v", err)
	}

	expectedAlbums := []string{"The Beatles - Abbey Road", "Pink Floyd - The Wall"}
	queueLines = strings.Split(strings.TrimSpace(string(queueContent)), "\n")

	if len(queueLines) != len(expectedAlbums) {
		t.Errorf("Expected %d albums in queue, got %d", len(expectedAlbums), len(queueLines))
	}

	for i, expected := range expectedAlbums {
		if i < len(queueLines) && queueLines[i] != expected {
			t.Errorf("Album %d: expected %q, got %q", i, expected, queueLines[i])
		}
	}
}

// TestCLI_Add_Duplicate tests error handling for duplicate albums
func TestCLI_Add_Duplicate(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Add first album - should succeed
	cmd := exec.Command("go", "run", "main.go", "add", "--queue", queueFile, "The Beatles - Abbey Road")
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check success message for first add
	expectedMsg := "Successfully added album: 'The Beatles - Abbey Road'"
	if !strings.Contains(outputStr, expectedMsg) {
		t.Errorf("Expected success message '%s'. Output: %s", expectedMsg, outputStr)
	}

	// Now try to add the same album again - should detect duplicate
	cmd = exec.Command("go", "run", "main.go", "add", "--queue", queueFile, "The Beatles - Abbey Road")
	cmd.Dir = "."

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Expected CLI to succeed for duplicate album, but it failed: %v\nOutput: %s", err, output)
	}

	outputStr = string(output)

	// Check informational message for duplicate
	expectedMsg = "Info: Album 'The Beatles - Abbey Road' already exists"
	if !strings.Contains(outputStr, expectedMsg) {
		t.Errorf("Expected info message '%s'. Output: %s", expectedMsg, outputStr)
	}

	// Try to add case-insensitive duplicate
	cmd = exec.Command("go", "run", "main.go", "add", "--queue", queueFile, "the beatles - abbey road")
	cmd.Dir = "."

	output, err = cmd.CombinedOutput()

	// Should also exit with code 0
	if err != nil {
		t.Fatalf("Expected CLI to succeed for case-insensitive duplicate, but it failed: %v\nOutput: %s", err, output)
	}

	outputStr = string(output)

	// Check informational message
	expectedMsg = "Info: Album 'the beatles - abbey road' already exists"
	if !strings.Contains(outputStr, expectedMsg) {
		t.Errorf("Expected info message '%s'. Output: %s", expectedMsg, outputStr)
	}

	// Verify only one album is in queue
	queueContent, err := os.ReadFile(queueFile)
	if err != nil {
		t.Fatalf("Failed to read queue file: %v", err)
	}

	queueLines := strings.Split(strings.TrimSpace(string(queueContent)), "\n")
	if len(queueLines) != 1 {
		t.Errorf("Expected 1 album in queue after duplicates, got %d", len(queueLines))
	}

	if queueLines[0] != "The Beatles - Abbey Road" {
		t.Errorf("Expected 'The Beatles - Abbey Road', got %q", queueLines[0])
	}
}

// TestCLI_Add_InvalidFormat tests error handling for invalid album format
func TestCLI_Add_InvalidFormat(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	testCases := []struct {
		name  string
		album string
	}{
		{"no dash", "No Dash Here"},
		{"missing artist", "Missing Artist"},
		{"dash at end", "Missing Album -"},
		{"whitespace before dash", "   - Album"},
		{"whitespace after dash", "Artist -   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("go", "run", "main.go", "add", "--queue", queueFile, tc.album)
			cmd.Dir = "."

			output, err := cmd.CombinedOutput()

			// Should exit with non-zero code
			if err == nil {
				t.Errorf("Expected CLI to fail for invalid album format: %q", tc.album)
			}

			outputStr := string(output)

			// Check error message
			if !strings.Contains(outputStr, "invalid album format") {
				t.Errorf("Expected 'invalid album format' error message for %q. Output: %s", tc.album, outputStr)
			}
		})
	}

	// Verify no albums were added
	if _, err := os.Stat(queueFile); err == nil {
		queueContent, err := os.ReadFile(queueFile)
		if err == nil && len(strings.TrimSpace(string(queueContent))) > 0 {
			t.Errorf("Expected no albums in queue after invalid formats, but queue file has content: %s", string(queueContent))
		}
	}
}

// TestCLI_Add_MissingArgument tests error handling for missing album argument
func TestCLI_Add_MissingArgument(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Run add command without album argument
	cmd := exec.Command("go", "run", "main.go", "add", "--queue", queueFile)
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()

	// Should exit with non-zero code
	if err == nil {
		t.Error("Expected CLI to fail for missing album argument")
	}

	outputStr := string(output)

	// Check error message
	if !strings.Contains(outputStr, "Album not specified") {
		t.Errorf("Expected 'Album not specified' error message. Output: %s", outputStr)
	}
}

// TestCLI_Add_Help tests the add help command
func TestCLI_Add_Help(t *testing.T) {
	// Test add help command
	cmd := exec.Command("go", "run", "main.go", "add", "--help")
	cmd.Dir = "."

	output, _ := cmd.CombinedOutput()

	// --help should exit with code 0 for flag package
	outputStr := string(output)

	// Check add-specific help content
	if !strings.Contains(outputStr, "Add a single album to the queue") {
		t.Errorf("Expected add description in help. Output: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Artist - Album") {
		t.Errorf("Expected format description in help. Output: %s", outputStr)
	}

	if !strings.Contains(outputStr, "--queue") {
		t.Errorf("Expected queue flag in help. Output: %s", outputStr)
	}

	if !strings.Contains(outputStr, `"The Beatles - Abbey Road"`) {
		t.Errorf("Expected example in help. Output: %s", outputStr)
	}
}

// TestCLI_Add_WithExistingQueue tests adding to an existing queue file
func TestCLI_Add_WithExistingQueue(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Create existing queue
	existingContent := "Pink Floyd - Dark Side of the Moon\nLed Zeppelin - IV\n"
	err := os.WriteFile(queueFile, []byte(existingContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Add new album
	cmd := exec.Command("go", "run", "main.go", "add", "--queue", queueFile, "The Beatles - Abbey Road")
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check success message
	if !strings.Contains(outputStr, "Successfully added album: 'The Beatles - Abbey Road'") {
		t.Errorf("Expected success message not found. Output: %s", outputStr)
	}

	// Verify album was appended to existing queue
	queueContent, err := os.ReadFile(queueFile)
	if err != nil {
		t.Fatalf("Failed to read queue file: %v", err)
	}

	expectedAlbums := []string{"Pink Floyd - Dark Side of the Moon", "Led Zeppelin - IV", "The Beatles - Abbey Road"}
	queueLines := strings.Split(strings.TrimSpace(string(queueContent)), "\n")

	if len(queueLines) != len(expectedAlbums) {
		t.Errorf("Expected %d albums in queue, got %d", len(expectedAlbums), len(queueLines))
	}

	for i, expected := range expectedAlbums {
		if i < len(queueLines) && queueLines[i] != expected {
			t.Errorf("Album %d: expected %q, got %q", i, expected, queueLines[i])
		}
	}
}

// TestCLI_Next_Success tests successful next command with non-empty queue
func TestCLI_Next_Success(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Create queue with test albums
	queueContent := "Pink Floyd - Dark Side of the Moon\nThe Beatles - Abbey Road\nPink Floyd - The Wall\n"
	err := os.WriteFile(queueFile, []byte(queueContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Build and run the CLI
	cmd := exec.Command("go", "run", "main.go", "next", "--queue", queueFile)
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check output format matches "Now listening: [Artist] - [Album]"
	if !strings.Contains(outputStr, "Now listening:") {
		t.Errorf("Expected 'Now listening:' in output. Output: %s", outputStr)
	}

	// Verify one of the original albums was selected
	originalAlbums := []string{"Pink Floyd - Dark Side of the Moon", "The Beatles - Abbey Road", "Pink Floyd - The Wall"}
	foundSelectedAlbum := false
	var selectedAlbum string

	for _, album := range originalAlbums {
		if strings.Contains(outputStr, album) {
			foundSelectedAlbum = true
			selectedAlbum = album
			break
		}
	}

	if !foundSelectedAlbum {
		t.Errorf("Output doesn't contain any of the expected albums. Output: %s", outputStr)
	}

	// Verify queue file now has one less album
	updatedContent, err := os.ReadFile(queueFile)
	if err != nil {
		t.Fatalf("Failed to read updated queue file: %v", err)
	}

	updatedLines := strings.Split(strings.TrimSpace(string(updatedContent)), "\n")
	if len(updatedLines) != 2 {
		t.Errorf("Expected 2 albums remaining in queue, got %d", len(updatedLines))
	}

	// Verify selected album was removed
	for _, remainingAlbum := range updatedLines {
		if remainingAlbum == selectedAlbum {
			t.Errorf("Selected album %q was not removed from queue", selectedAlbum)
		}
	}
}

// TestCLI_Next_EmptyQueue tests next command with empty queue
func TestCLI_Next_EmptyQueue(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Create empty queue file
	err := os.WriteFile(queueFile, []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Build and run the CLI
	cmd := exec.Command("go", "run", "main.go", "next", "--queue", queueFile)
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()

	// Should exit with non-zero code
	if err == nil {
		t.Error("Expected CLI to fail for empty queue")
	}

	outputStr := string(output)

	// Check error message contains "queue is empty"
	if !strings.Contains(outputStr, "queue is empty") {
		t.Errorf("Expected 'queue is empty' error message. Output: %s", outputStr)
	}
}

// TestCLI_Next_NonExistentQueue tests next command with non-existent queue file
func TestCLI_Next_NonExistentQueue(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "nonexistent.txt")

	// Build and run the CLI
	cmd := exec.Command("go", "run", "main.go", "next", "--queue", queueFile)
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()

	// Should exit with non-zero code
	if err == nil {
		t.Error("Expected CLI to fail for non-existent queue")
	}

	outputStr := string(output)

	// Check error message contains "queue is empty" (since ReadLines returns empty slice for non-existent files)
	if !strings.Contains(outputStr, "queue is empty") {
		t.Errorf("Expected 'queue is empty' error message. Output: %s", outputStr)
	}
}

// TestCLI_Next_Help tests the next command help
func TestCLI_Next_Help(t *testing.T) {
	// Test next help command
	cmd := exec.Command("go", "run", "main.go", "next", "--help")
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Next help command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check help content
	if !strings.Contains(outputStr, "Usage:") {
		t.Errorf("Expected usage information in help. Output: %s", outputStr)
	}

	if !strings.Contains(outputStr, "next") {
		t.Errorf("Expected 'next' in help content. Output: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Get a random album from the queue") {
		t.Errorf("Expected command description in help. Output: %s", outputStr)
	}
}

// TestCLI_Next_SingleAlbum tests next command with single album in queue
func TestCLI_Next_SingleAlbum(t *testing.T) {
	tempDir := t.TempDir()
	queueFile := filepath.Join(tempDir, "queue.txt")

	// Create queue with single album
	testAlbum := "Pink Floyd - The Wall"
	err := os.WriteFile(queueFile, []byte(testAlbum+"\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Build and run the CLI
	cmd := exec.Command("go", "run", "main.go", "next", "--queue", queueFile)
	cmd.Dir = "." // Run from cmd/queue directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI command failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Check output contains the test album
	if !strings.Contains(outputStr, testAlbum) {
		t.Errorf("Expected output to contain %q. Output: %s", testAlbum, outputStr)
	}

	// Verify queue file is now empty
	updatedContent, err := os.ReadFile(queueFile)
	if err != nil {
		t.Fatalf("Failed to read updated queue file: %v", err)
	}

	updatedContentStr := strings.TrimSpace(string(updatedContent))
	if updatedContentStr != "" {
		t.Errorf("Expected empty queue file after selecting last album, got: %q", updatedContentStr)
	}
}
