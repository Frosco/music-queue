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
