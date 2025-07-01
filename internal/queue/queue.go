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

// addAlbumCheck validates a single album, checks for duplicates against the given map,
// and returns an error if it fails. This is a helper for AddAlbum and ImportAlbums.
func addAlbumCheck(albumTitle string, existingAlbumsMap map[string]bool) error {
	trimmedTitle := strings.TrimSpace(albumTitle)

	if !validateAlbumFormat(trimmedTitle) {
		return fmt.Errorf("invalid format") // Generic error, handled by callers
	}

	albumLower := strings.ToLower(trimmedTitle)
	if existingAlbumsMap[albumLower] {
		return fmt.Errorf("album '%s' already exists", trimmedTitle)
	}

	return nil
}

// AddAlbum adds a single album to the queue with duplicate checking
// Returns an error if the album format is invalid or if there's a storage error
func (qs *QueueService) AddAlbum(albumTitle string) error {
	// Read existing queue
	existingAlbums, err := qs.storage.ReadLines()
	if err != nil {
		return fmt.Errorf("failed to read existing queue: %w", err)
	}

	// Create a map for case-insensitive duplicate checking
	existingAlbumsMap := make(map[string]bool)
	for _, album := range existingAlbums {
		existingAlbumsMap[strings.ToLower(strings.TrimSpace(album))] = true
	}

	// Validate and check for duplicates using the helper
	err = addAlbumCheck(albumTitle, existingAlbumsMap)
	if err != nil {
		if strings.Contains(err.Error(), "invalid format") {
			return fmt.Errorf("invalid album format: must be 'Artist - Album' format")
		}
		// Return the "already exists" error directly
		return err
	}

	// Add album to queue
	updatedAlbums := append(existingAlbums, strings.TrimSpace(albumTitle))

	// Save updated queue
	err = qs.storage.WriteLines(updatedAlbums)
	if err != nil {
		return fmt.Errorf("failed to save updated queue: %w", err)
	}

	return nil
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

	// Process import albums using the helper function
	addedCount := 0
	skippedCount := 0
	currentAlbums := existingAlbums

	for _, album := range importAlbums {
		// Skip empty lines
		if strings.TrimSpace(album) == "" {
			continue
		}

		// Validate album format and check for duplicates using helper
		err := addAlbumCheck(album, existingAlbumsMap)
		if err != nil {
			skippedCount++
			continue // Skip invalid format or duplicate
		}

		// Add album
		processedAlbum := strings.TrimSpace(album)
		currentAlbums = append(currentAlbums, processedAlbum)
		existingAlbumsMap[strings.ToLower(processedAlbum)] = true
		addedCount++
	}

	// If we have new albums, save the updated queue
	if addedCount > 0 {
		err = qs.storage.WriteLines(currentAlbums)
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
		return ".music-queue/queue.txt"
	}
	return filepath.Join(homeDir, ".music-queue", "queue.txt")
}
