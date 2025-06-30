package queue

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"music-queue/internal/storage"
)

// QueueService handles business logic for the music queue
type QueueService struct {
	storage *storage.FileStorage
}

// NewQueue creates a new QueueService instance with the provided storage service
func NewQueue(storageService *storage.FileStorage) *QueueService {
	return &QueueService{
		storage: storageService,
	}
}

// validateAlbumFormat checks if an album entry follows the "Artist Name - Album Title" format
// Returns true if valid, false otherwise
func validateAlbumFormat(album string) bool {
	album = strings.TrimSpace(album)

	// Must contain at least one dash
	dashIndex := strings.Index(album, "-")
	if dashIndex == -1 {
		return false
	}

	// Must have at least one character before the dash
	if dashIndex == 0 {
		return false
	}

	// Must have at least one character after the dash
	if dashIndex == len(album)-1 {
		return false
	}

	// Check that there's content before and after the dash (not just whitespace)
	beforeDash := strings.TrimSpace(album[:dashIndex])
	afterDash := strings.TrimSpace(album[dashIndex+1:])

	return len(beforeDash) > 0 && len(afterDash) > 0
}

// ImportAlbums imports albums from a text file, skipping duplicates (case-insensitive)
// Returns the number of albums added, number skipped, and any error encountered
func (qs *QueueService) ImportAlbums(filename string) (added int, skipped int, err error) {
	// Check if import file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return 0, 0, fmt.Errorf("file not found: %s", filename)
	}

	// Read import file
	importStorage := storage.NewFileStorage(filename)
	importAlbums, err := importStorage.ReadLines()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read import file: %w", err)
	}

	// Handle empty file gracefully
	if len(importAlbums) == 0 {
		return 0, 0, nil
	}

	// Read existing queue
	existingAlbums, err := qs.storage.ReadLines()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read existing queue: %w", err)
	}

	// Create a map for case-insensitive duplicate checking
	existingAlbumsMap := make(map[string]bool)
	for _, album := range existingAlbums {
		existingAlbumsMap[strings.ToLower(strings.TrimSpace(album))] = true
	}

	// Process import albums
	var newAlbums []string
	addedCount := 0
	skippedCount := 0

	for _, album := range importAlbums {
		album = strings.TrimSpace(album)

		// Skip empty lines (already handled by storage layer, but being explicit)
		if album == "" {
			continue
		}

		// Validate album format (Artist Name - Album Title)
		if !validateAlbumFormat(album) {
			skippedCount++
			continue
		}

		albumLower := strings.ToLower(album)

		// Check for duplicates (case-insensitive)
		if existingAlbumsMap[albumLower] {
			skippedCount++
			continue
		}

		// Add to new albums list and mark as existing to prevent duplicates within import file
		newAlbums = append(newAlbums, album)
		existingAlbumsMap[albumLower] = true
		addedCount++
	}

	// If we have new albums, append them to the existing queue and save
	if len(newAlbums) > 0 {
		allAlbums := append(existingAlbums, newAlbums...)
		err = qs.storage.WriteLines(allAlbums)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to save updated queue: %w", err)
		}
	}

	return addedCount, skippedCount, nil
}

// GetDefaultQueuePath returns the default queue file path
func GetDefaultQueuePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if can't get home directory
		return ".go-music-queue/queue.txt"
	}
	return filepath.Join(homeDir, ".go-music-queue", "queue.txt")
}
