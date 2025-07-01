package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"music-queue/internal/queue"
	"music-queue/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "import":
		handleImportCommand()
	case "add":
		handleAddCommand()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleImportCommand() {
	// Set up flag parsing for import command
	importFlags := flag.NewFlagSet("import", flag.ExitOnError)
	queuePath := importFlags.String("queue", queue.GetDefaultQueuePath(), "Path to queue file")

	importFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s import [flags] <import-file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Import albums from a text file to the queue.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <import-file>  Path to text file containing album names (one per line)\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		importFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s import albums.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s import --queue /custom/path/queue.txt albums.txt\n", os.Args[0])
	}

	// Parse import command arguments
	err := importFlags.Parse(os.Args[2:])
	if err != nil {
		os.Exit(1)
	}

	// Check if import file was provided
	if importFlags.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Error: Import file not specified\n\n")
		importFlags.Usage()
		os.Exit(1)
	}

	importFile := importFlags.Arg(0)

	// Validate import file exists
	if _, err := os.Stat(importFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Import file '%s' not found\n", importFile)
		os.Exit(1)
	}

	// Get absolute path for better error messages
	absImportFile, err := filepath.Abs(importFile)
	if err != nil {
		absImportFile = importFile // fallback to original path
	}

	// Create storage and queue service
	queueStorage := storage.NewFileStorage(*queuePath)
	queueService := queue.NewQueue(queueStorage)

	// Perform import
	fmt.Printf("Importing albums from '%s'...\n", absImportFile)

	added, skipped, err := queueService.ImportAlbums(importFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Display results with clear formatting
	if added == 0 && skipped == 0 {
		fmt.Println("No albums found in import file.")
	} else {
		fmt.Printf("Import complete! Added %d albums, Skipped %d duplicates\n", added, skipped)

		// Show queue file location
		absQueuePath, err := filepath.Abs(*queuePath)
		if err != nil {
			absQueuePath = *queuePath
		}
		fmt.Printf("Queue saved to: %s\n", absQueuePath)
	}
}

func handleAddCommand() {
	// Set up flag parsing for add command
	addFlags := flag.NewFlagSet("add", flag.ExitOnError)
	queuePath := addFlags.String("queue", queue.GetDefaultQueuePath(), "Path to queue file")

	addFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s add [flags] \"Artist - Album\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Add a single album to the queue.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  \"Artist - Album\"  Album to add in 'Artist - Album' format\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		addFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s add \"The Beatles - Abbey Road\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s add --queue /custom/path/queue.txt \"Pink Floyd - The Wall\"\n", os.Args[0])
	}

	// Parse add command arguments
	err := addFlags.Parse(os.Args[2:])
	if err != nil {
		os.Exit(1)
	}

	// Check if album argument was provided
	if addFlags.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Error: Album not specified\n\n")
		addFlags.Usage()
		os.Exit(1)
	}

	albumTitle := addFlags.Arg(0)

	// Create storage and queue service
	queueStorage := storage.NewFileStorage(*queuePath)
	queueService := queue.NewQueue(queueStorage)

	// Add the album
	err = queueService.AddAlbum(albumTitle)
	if err != nil {
		// Handle duplicate album as an informational message, not an error
		if strings.Contains(err.Error(), "already exists") {
			// Capitalize first letter for better output and print to stdout
			fmt.Printf("Info: %s\n", strings.ToUpper(string(err.Error()[0]))+err.Error()[1:])
			os.Exit(0)
		}

		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Success message
	fmt.Printf("Successfully added album: '%s'\n", albumTitle)

	// Show queue file location
	absQueuePath, err := filepath.Abs(*queuePath)
	if err != nil {
		absQueuePath = *queuePath
	}
	fmt.Printf("Queue saved to: %s\n", absQueuePath)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Go Music Queue - Manage your music listening queue\n\n")
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [arguments]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  add \"Artist - Album\"  Add a single album to the queue\n")
	fmt.Fprintf(os.Stderr, "  import <file>         Import albums from a text file\n")
	fmt.Fprintf(os.Stderr, "  help                  Show this help message\n\n")
	fmt.Fprintf(os.Stderr, "For command-specific help:\n")
	fmt.Fprintf(os.Stderr, "  %s <command> --help\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  %s add \"The Beatles - Abbey Road\"\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s import my-albums.txt\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s add --help\n", os.Args[0])
}
