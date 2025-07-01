# Music Queue - System Architecture

## **Overview**

The Music Queue application is a command-line tool designed to manage a personal music listening queue. It follows a clean, layered architecture that separates concerns and promotes maintainability.

## **High-Level Architecture**

We use a layered architecture that separates the command-line interface (CLI) from the core business logic, and the business logic from the data storage mechanism.

1. **CLI Layer (`src/cmd`)**: This layer is responsible for parsing user commands and arguments.
2. **Business Logic Layer (`src/internal/queue`)**: This layer contains the core application logic, such as how to add an album, check for duplicates, and select the next album. It knows nothing about the command line.
3. **Storage Layer (`src/internal/storage`)**: This layer is responsible for reading from and writing to the `queue.txt` file. It knows nothing about albums or business rules.

## **Source Tree**

```
music-queue/
├── src/
│   ├── cmd/
│   │   └── queue/
│   │       ├── main.go         // Entry point, CLI command parsing
│   │       └── main_test.go    // CLI layer tests
│   └── internal/
│       ├── queue/
│       │   ├── queue.go        // Core business logic (Add, Next, List, etc.)
│       │   └── queue_test.go   // Business logic tests
│       └── storage/
│           ├── file.go         // File I/O operations
│           └── file_test.go    // Storage layer tests
├── docs/
│   ├── architecture.md     // This document
│   ├── prd.md             // Product Requirements Document
│   └── project-brief.md   // Project overview
├── go.mod                 // Go module definition
├── go.sum                 // Go module checksums
├── LICENSE                // Project license
└── queue.txt              // The data file (can be placed elsewhere)
```

## **Tech Stack**

### **Core Technology**
- **Language**: Go 1.21+
- **Module System**: Go modules
- **CLI Framework**: Standard library `flag` package (or Cobra for advanced CLI features)
- **Testing**: Go's built-in testing framework with `testing` package

### **Dependencies**
- **Standard Library Only**: Minimize external dependencies for simplicity and reliability
- **Potential Additions**:
  - `github.com/spf13/cobra` - For advanced CLI features (optional)
  - `github.com/spf13/viper` - For configuration management (future enhancement)

### **Development Tools**
- **Linting**: `golangci-lint` for code quality
- **Formatting**: `gofmt` for consistent code formatting
- **Documentation**: Go's built-in `godoc` for API documentation

### **Platform Support**
- **Operating Systems**: Linux, macOS, Windows
- **Architecture**: x86_64, ARM64
- **Distribution**: Single binary executable

## **Component Breakdown**

### 1. CLI Layer: `src/cmd/queue/main.go`

**Responsibility**: To be the main entry point of the application.

**Implementation**:
- Uses Go's built-in `flag` package or a more robust third-party library like **`cobra`** to define and parse the commands (`add`, `import`, `next`, `list`, `count`).
- Based on the command provided by the user, it calls the appropriate function in the `queue` package.
- Handles all `fmt.Println()` calls to display output to the user. It is the only part of the application that should directly interact with the console.
- Provides clear error messages and usage instructions.

### 2. Business Logic Layer: `src/internal/queue/queue.go`

**Responsibility**: To implement all the business rules defined in the PRD. This package is the heart of the application.

**Proposed Functions**:
- `NewQueue(storageService *storage.FileStorage) *QueueService`: Creates a new queue service instance.
- `AddAlbum(albumTitle string) error`: Handles adding a single album, including the case-insensitive duplicate check.
- `ImportAlbums(filename string) (added int, skipped int, err error)`: Logic for importing from a file.
- `GetNextAlbum() (string, error)`: Logic to get a random album and trigger its removal.
- `ListAlbums() ([]string, error)`: Gets the list of all albums.
- `CountAlbums() (int, error)`: Counts the albums.

**Key Detail**: This layer coordinates operations. For example, `AddAlbum` will first ask the `storage` layer for the current list, then perform its duplicate check, and finally tell the `storage` layer to write the new list back to the file.

### 3. Storage Layer: `src/internal/storage/file.go`

**Responsibility**: To abstract away the file system operations. Its only job is to read and write slices of strings from/to a file.

**Proposed Functions**:
- `NewFileStorage(filePath string) *FileStorage`: Creates a new storage manager for a specific file.
- `ReadLines() ([]string, error)`: Reads all lines from the `queue.txt` file into a slice of strings.
- `WriteLines(lines []string) error`: Overwrites the `queue.txt` file with a new slice of strings.

**Data Location**: The `queue.txt` file is stored in the user's home directory by default (e.g., `~/.music-queue/queue.txt`). This prevents cluttering the directory where the command is run and gives the queue a persistent, predictable location. The application creates this directory on its first run.

## **Data Flow Example: `queue next`**

1. **User** runs `go run main.go next` in their terminal.
2. **`main.go`** parses the `next` command.
3. It calls the `queueService.GetNextAlbum()` function from the `queue` package.
4. `GetNextAlbum()` calls `storage.ReadLines()` to get all albums from `queue.txt`.
5. It checks if the list is empty. If not, it randomly selects one album.
6. It creates a *new* list containing all albums *except* the one it selected.
7. It calls `storage.WriteLines()` to save this new, shorter list back to `queue.txt`.
8. It returns the selected album string to `main.go`.
9. **`main.go`** prints the result to the console: `Now Listening: Artist - Album`.

## **Coding Standards**

### **General Principles**
- **Simplicity**: Prefer simple, readable code over clever solutions
- **Consistency**: Follow Go conventions and idioms
- **Testability**: Write testable code with clear interfaces
- **Error Handling**: Always handle errors explicitly and provide meaningful error messages

### **Code Style**
- **Formatting**: Use `gofmt` for consistent formatting
- **Naming**: Follow Go naming conventions (camelCase for variables, PascalCase for exported names)
- **Comments**: Use clear, concise comments for exported functions and complex logic
- **Line Length**: Keep lines under 120 characters for readability

### **Package Structure**
- **Package Names**: Use descriptive, single-word package names
- **File Organization**: Group related functions in the same file
- **Exports**: Only export what's necessary for the public API

### **Error Handling**
- **Return Errors**: Always return errors from functions that can fail
- **Error Wrapping**: Use `fmt.Errorf` with `%w` verb for error wrapping
- **Error Messages**: Provide context in error messages
- **Panic Avoidance**: Avoid panics in production code; handle errors gracefully

### **Testing**
- **Test Coverage**: Aim for >80% test coverage
- **Test Naming**: Use descriptive test names that explain the scenario
- **Test Organization**: Group related tests using subtests
- **Mocking**: Use interfaces for testability and dependency injection

### **Documentation**
- **Package Comments**: Every package should have a package comment
- **Function Comments**: Document all exported functions
- **Examples**: Include usage examples in documentation
- **README**: Maintain an up-to-date README with installation and usage instructions

### **Performance Considerations**
- **Memory Usage**: Be mindful of memory allocations in hot paths
- **File I/O**: Minimize file operations and use buffered I/O when appropriate
- **Concurrency**: Use goroutines and channels when beneficial, but keep it simple

### **Security**
- **Input Validation**: Validate all user inputs
- **File Permissions**: Set appropriate file permissions for data files
- **Path Traversal**: Prevent path traversal attacks when handling file paths

## **Future Enhancements**

### **Potential Improvements**
- **Configuration Management**: Add support for configuration files
- **Multiple Storage Backends**: Support for databases or cloud storage
- **Web Interface**: Add a simple web UI for queue management
- **Statistics**: Track listening history and provide analytics
- **Integration**: Integrate with music streaming services

### **Scalability Considerations**
- **Concurrent Access**: Handle multiple processes accessing the queue
- **Large Queues**: Optimize for queues with thousands of albums
- **Performance**: Profile and optimize hot paths as the application grows

