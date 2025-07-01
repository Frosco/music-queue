# Go Music Queue

A simple, command-line application for managing a personal music album queue. Built with Go, this tool helps you organize albums you want to listen to and randomly selects your next listening choice.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Project Structure](#project-structure)
- [Development](#development)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Add Albums**: Add single albums manually or import from text files
- **Duplicate Detection**: Prevents duplicate albums with case-insensitive matching
- **Random Selection**: Get a random album from your queue and automatically remove it
- **Queue Management**: List all albums, count queue size, and manage your collection
- **File-based Storage**: Simple text file storage for portability and simplicity
- **Cross-platform**: Works on Linux, macOS, and Windows

## Installation

### Prerequisites

- Go 1.24.4 or later

### Build from Source

1. Clone the repository:
```bash
git clone <repository-url>
cd music-queue
```

2. Build the application:
```bash
go build -o queue src/cmd/queue/main.go
```

3. (Optional) Install globally:
```bash
# Linux/macOS
sudo mv queue /usr/local/bin/

# Or add to your PATH
export PATH=$PATH:$(pwd)
```

### Using Go Install

```bash
go install music-queue/src/cmd/queue@latest
```

## Quick Start

1. **Add your first album:**
```bash
./queue add "The Beatles - Abbey Road"
```

2. **Import albums from a file:**
```bash
# Create a text file with one album per line
echo "Pink Floyd - The Wall" > my-albums.txt
echo "Led Zeppelin - IV" >> my-albums.txt
./queue import my-albums.txt
```

3. **Get your next album to listen to:**
```bash
./queue next
```

4. **View your queue:**
```bash
./queue list
```

## Usage

### Commands

#### `import` - Import albums from a text file
```bash
./queue import [--queue /path/to/queue.txt] <import-file>
```

**Examples:**
```bash
./queue import albums.txt
./queue import --queue /custom/path/queue.txt albums.txt
```

The import file should contain one album per line in "Artist - Album" format:
```
The Beatles - Abbey Road
Pink Floyd - The Wall
Radiohead - OK Computer
```

#### `add` - Add a single album
```bash
./queue add [--queue /path/to/queue.txt] "Artist - Album"
```

**Examples:**
```bash
./queue add "Daft Punk - Discovery"
./queue add --queue /custom/path/queue.txt "King Gizzard & The Lizard Wizard - PetroDragonic Apocalypse"
```

#### `next` - Get next album (random selection)
```bash
./queue next [--queue /path/to/queue.txt]
```

Randomly selects an album from your queue, displays it, and removes it from the queue.

#### `list` - Display all albums in queue
```bash
./queue list [--queue /path/to/queue.txt]
```

Shows a numbered list of all albums currently in your queue.

#### `count` - Show queue size
```bash
./queue count [--queue /path/to/queue.txt]
```

Displays the total number of albums in your queue.

#### `help` - Show usage information
```bash
./queue help
```

### Album Format

Albums must follow the format: `Artist - Album Title`

**Valid examples:**
- `"The Beatles - Abbey Road"`
- `"Led Zeppelin - IV"`
- `"King Gizzard & The Lizard Wizard - PetroDragonic Apocalypse"`

**Invalid examples:**
- `"Abbey Road"` (missing artist)
- `"The Beatles"` (missing album)
- `"The Beatles Abbey Road"` (missing dash separator)

### Queue File Location

By default, the queue is stored in:
- **Linux/macOS**: `~/.local/share/music-queue/queue.txt`
- **Windows**: `%APPDATA%/music-queue/queue.txt`

You can specify a custom location using the `--queue` flag with any command.

## Project Structure

```
music-queue/
├── src/
│   ├── cmd/
│   │   └── queue/
│   │       ├── main.go           # CLI application entry point
│   │       └── main_test.go      # CLI integration tests
│   └── internal/
│       ├── queue/
│       │   ├── queue.go          # Core business logic
│       │   └── queue_test.go     # Queue service tests
│       └── storage/
│           ├── file.go           # File storage implementation
│           └── file_test.go      # Storage layer tests
├── docs/                         # Project documentation
├── go.mod                        # Go module definition
└── README.md                     # This file
```

### Architecture Overview

The application follows a clean architecture pattern with clear separation of concerns:

- **CLI Layer** (`src/cmd/queue/`): Handles command-line interface, argument parsing, and user interaction
- **Business Logic** (`src/internal/queue/`): Core queue operations, validation, and business rules
- **Storage Layer** (`src/internal/storage/`): File I/O operations and data persistence

**Key Components:**

- **QueueService**: Manages album operations (add, import, get next, list, count)
- **FileStorage**: Handles reading and writing to text files
- **Validation**: Ensures album format compliance and prevents duplicates

## Development

### Prerequisites

- Go 1.24.4 or later
- Git

### Setup Development Environment

1. Clone the repository:
```bash
git clone <repository-url>
cd music-queue
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run tests:
```bash
go test ./...
```

4. Build and test locally:
```bash
go build -o queue src/cmd/queue/main.go
./queue help
```

### Code Organization

- **Package Structure**: Follow Go package conventions with clear boundaries
- **Error Handling**: Comprehensive error handling with descriptive messages
- **Testing**: Unit tests for all major components
- **Documentation**: Inline documentation following Go conventions

## Testing

Run all tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

Run tests for specific package:
```bash
go test ./src/internal/queue/
go test ./src/internal/storage/
```

### Test Coverage

The project includes comprehensive tests for:
- Queue operations (add, import, get next, list, count)
- File storage operations
- CLI command handling
- Error conditions and edge cases

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Guidelines

- Follow Go conventions and best practices
- Write tests for new functionality
- Keep functions small and focused
- Use descriptive variable and function names
- Document public APIs

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

---

**Happy listening! 🎵**

For questions or support, please open an issue in the repository. 