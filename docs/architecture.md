## **High-Level Architecture**

We will use a layered architecture that separates the command-line interface (CLI) from the core business logic, and the business logic from the data storage mechanism.

1.  **CLI Layer (`cmd`)**: This layer is responsible for parsing user commands and arguments.
2.  **Business Logic Layer (`internal/queue`)**: This layer contains the core application logic, such as how to add an album, check for duplicates, and select the next album. It knows nothing about the command line.
3.  **Storage Layer (`internal/storage`)**: This layer is responsible for reading from and writing to the `queue.txt` file. It knows nothing about albums or business rules.

Here's how these components will fit together in the project's directory structure:

```
go-music-queue/
├── cmd/
│   └── queue/
│       └── main.go         // Entry point, CLI command parsing
├── internal/
│   ├── queue/
│   │   └── queue.go        // Core business logic (Add, Next, List, etc.)
│   └── storage/
│       └── file.go         // File I/O operations
└── queue.txt               // The data file (can be placed elsewhere)
```

## **Component Breakdown**

### 1\. CLI Layer: `cmd/queue/main.go`

  * **Responsibility**: To be the main entry point of the application.
  * **Implementation**:
      * It will use Go's built-in `flag` package or a more robust third-party library like **`cobra`** to define and parse the commands (`add`, `import`, `next`, `list`, `count`).
      * Based on the command provided by the user, it will call the appropriate function in the `queue` package.
      * It will handle all `fmt.Println()` calls to display output to the user. It is the only part of the application that should directly interact with the console.

### 2\. Business Logic Layer: `internal/queue/queue.go`

  * **Responsibility**: To implement all the business rules defined in the PRD. This package will be the heart of the application.
  * **Proposed Functions**:
      * `NewQueue(storageService *storage.FileStorage) *QueueService`: Creates a new queue service instance.
      * `AddAlbum(albumTitle string) error`: Handles adding a single album, including the case-insensitive duplicate check.
      * `ImportAlbums(filename string) (added int, skipped int, err error)`: Logic for importing from a file.
      * `GetNextAlbum() (string, error)`: Logic to get a random album and trigger its removal.
      * `ListAlbums() ([]string, error)`: Gets the list of all albums.
      * `CountAlbums() (int, error)`: Counts the albums.
  * **Key Detail**: This layer will coordinate operations. For example, `AddAlbum` will first ask the `storage` layer for the current list, then perform its duplicate check, and finally tell the `storage` layer to write the new list back to the file.

### 3\. Storage Layer: `internal/storage/file.go`

  * **Responsibility**: To abstract away the file system operations. Its only job is to read and write slices of strings from/to a file.
  * **Proposed Functions**:
      * `NewFileStorage(filePath string) *FileStorage`: Creates a new storage manager for a specific file.
      * `ReadLines() ([]string, error)`: Reads all lines from the `queue.txt` file into a slice of strings.
      * `WriteLines(lines []string) error`: Overwrites the `queue.txt` file with a new slice of strings.
  * **Data Location**: I recommend the `queue.txt` file be stored in the user's home directory by default (e.g., `~/.go-music-queue/queue.txt`). This prevents cluttering the directory where the command is run and gives the queue a persistent, predictable location. The application can create this directory on its first run.

## **Data Flow Example: `queue next`**

1.  **User** runs `go run main.go next` in their terminal.
2.  **`main.go`** parses the `next` command.
3.  It calls the `queueService.GetNextAlbum()` function from the `queue` package.
4.  `GetNextAlbum()` calls `storage.ReadLines()` to get all albums from `queue.txt`.
5.  It checks if the list is empty. If not, it randomly selects one album.
6.  It creates a *new* list containing all albums *except* the one it selected.
7.  It calls `storage.WriteLines()` to save this new, shorter list back to `queue.txt`.
8.  It returns the selected album string to `main.go`.
9.  **`main.go`** prints the result to the console: `Now Listening: Artist - Album`.
