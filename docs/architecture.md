# Music Queue Architecture Document

## Introduction

This document outlines the overall project architecture for **Music Queue**, including backend systems, shared services, and non-UI specific concerns. Its primary goal is to serve as the guiding architectural blueprint for AI-driven development, ensuring consistency and adherence to chosen patterns and technologies.

**Relationship to Frontend Architecture:**
This project is a command-line interface (CLI) application with no user interface components. All user interaction occurs through terminal commands, making this document the complete architectural reference for the project.

### Starter Template or Existing Project

**N/A** - This is a greenfield Go project built from scratch without using any starter template or boilerplate. The project follows standard Go conventions and project structure.

### Change Log

| Date | Version | Description | Author |
| :--- | :------ | :---------- | :----- |
| 2024-12-19 | 2.0 | Complete architecture redesign following template | Winston (Architect) |
| 2024-10-27 | 1.0 | Initial architecture document | Original Team |

## High Level Architecture

### Technical Summary

The Music Queue application follows a **clean layered architecture** that separates the command-line interface from core business logic and data storage. The system uses a **monolithic single-binary approach** with **file-based persistence**, providing a simple yet robust solution for personal music queue management. The architecture emphasizes **simplicity, maintainability, and zero external dependencies** while supporting core operations like adding albums, importing from files, and random album selection.

### High Level Overview

1. **Architectural Style**: Layered monolithic architecture with clear separation of concerns
2. **Repository Structure**: Single repository (monorepo approach) with organized package structure
3. **Service Architecture**: Single-service monolith compiled to standalone binary
4. **Primary User Interaction Flow**: User executes CLI commands → CLI layer parses arguments → Business logic layer processes request → Storage layer handles file I/O → Results returned to user
5. **Key Architectural Decisions**:
   - **File-based storage** for simplicity and zero setup requirements
   - **Standard library only** to minimize dependencies and ensure reliability
   - **Package-based layering** to maintain clean boundaries between concerns
   - **Case-insensitive duplicate detection** for better user experience

### High Level Project Diagram

*(High-level architecture diagram showing the layered structure with User Terminal → CLI Layer → Business Logic Layer → Storage Layer → File System)*

### Architectural and Design Patterns

- **Layered Architecture:** Clear separation between CLI, business logic, and storage layers - _Rationale:_ Enables independent testing, maintainability, and future enhancement without tight coupling
- **Dependency Injection:** Storage service injected into business logic layer - _Rationale:_ Allows for easy mocking in tests and potential future storage backend changes
- **Repository Pattern:** Abstract data access through storage layer interface - _Rationale:_ Isolates file I/O operations and enables future migration to different storage mechanisms
- **Command Pattern:** CLI commands mapped to business logic operations - _Rationale:_ Provides clear separation between user interface and business operations
- **Single Responsibility Principle:** Each layer has one clear responsibility - _Rationale:_ Improves code maintainability and reduces coupling between components

## Tech Stack

### Cloud Infrastructure

- **Provider:** Local Development Environment (No Cloud Required)
- **Key Services:** Local File System
- **Deployment Regions:** User's Local Machine

### Technology Stack Table

| Category           | Technology         | Version     | Purpose     | Rationale      |
| :----------------- | :----------------- | :---------- | :---------- | :------------- |
| **Language**       | Go                 | 1.24.4      | Primary development language | Memory efficient, fast compilation, excellent CLI tooling, single binary output |
| **Runtime**        | Go Runtime         | 1.24.4      | Application runtime | Built-in with Go, cross-platform support, no external dependencies |
| **Framework**      | Standard Library   | 1.24.4      | CLI and core functionality | Zero dependencies, reliable, well-tested, sufficient for project needs |
| **Database**       | File System        | N/A         | Data persistence | Simple, no setup required, human-readable, version control friendly |
| **Cache**          | In-Memory Maps     | N/A         | Duplicate detection | Fast lookups, suitable for small datasets, no external dependencies |
| **Message Queue**  | N/A                | N/A         | Not required | Single-user application with immediate processing |
| **API Style**      | CLI Commands       | N/A         | User interface | Appropriate for developer tooling, fast, scriptable |
| **Authentication** | File Permissions   | OS Level    | Access control | Leverages OS security model, no additional complexity |
| **Testing**        | Go Testing         | 1.24.4      | Unit and integration tests | Built-in, no external dependencies, good tooling support |
| **Build Tool**     | Go Build           | 1.24.4      | Compilation and build | Native Go tooling, cross-compilation support |
| **IaC Tool**       | N/A                | N/A         | No infrastructure required | Local application, no cloud resources needed |
| **Monitoring**     | Exit Codes/Logs    | N/A         | Basic error reporting | Sufficient for CLI tool, integrates with shell scripting |
| **Logging**        | Standard Output    | N/A         | User feedback | Direct console output, appropriate for CLI tools |

## Data Models

### Album

**Purpose:** Represents a music album entry in the queue with artist and album information

**Key Attributes:**

- Title: string - Complete album string in "Artist - Album" format
- Artist: string (derived) - Artist name portion before the dash
- Album: string (derived) - Album name portion after the dash

**Relationships:**

- Part of Albums collection in queue
- No relationships to other entities (simple flat structure)

**Validation Rules:**

- Must contain at least one dash character
- Must have non-empty content before and after the dash
- Case-insensitive duplicate detection

### Queue

**Purpose:** Represents the collection of albums awaiting listening

**Key Attributes:**

- Albums: []string - Ordered list of album titles
- FilePath: string - Location of queue.txt file
- Size: int (computed) - Number of albums in queue

**Relationships:**

- Contains multiple Album entries
- Persisted to file system as queue.txt

## Components

### CLI Layer (cmd/queue)

**Responsibility:** Command-line interface parsing, user interaction, and output formatting

**Key Interfaces:**

- Command parsing and validation
- Flag handling for optional parameters
- User feedback and error reporting
- Help text and usage information

**Dependencies:** Business Logic Layer (queue service)

**Technology Stack:** Go flag package, os package, fmt package for output

### Business Logic Layer (internal/queue)

**Responsibility:** Core application logic, business rules, and data validation

**Key Interfaces:**

- AddAlbum(albumTitle string) error
- ImportAlbums(filename string) (added int, skipped int, err error)
- GetNextAlbum() (string, error)
- ListAlbums() ([]string, error)
- CountAlbums() (int, error)

**Dependencies:** Storage Layer (file storage service)

**Technology Stack:** Go standard library, custom business logic

### Storage Layer (internal/storage)

**Responsibility:** File I/O operations and data persistence abstraction

**Key Interfaces:**

- ReadLines() ([]string, error)
- WriteLines(lines []string) error
- File path management and directory creation

**Dependencies:** Go file system packages (os, filepath)

**Technology Stack:** Go os package, filepath package, io utilities

### Component Diagrams

*(Component interaction diagram created above showing the relationships between CLI, Business Logic, and Storage layers)*

## External APIs

**No External APIs Required** - This application operates entirely locally with no external service dependencies. All data is stored and retrieved from the local file system.

## Core Workflows

### Add Album Workflow

*(Sequence diagram created above showing the complete flow of adding an album to the queue)*

### Get Next Album Workflow

The `queue next` command follows this flow:
1. User executes `queue next` command
2. CLI layer parses command and calls business logic
3. Business logic reads current queue from storage
4. If queue is empty, return error message
5. If queue has albums, randomly select one using secure random number generation
6. Remove selected album from queue and update storage
7. Return selected album to user with "Now Listening:" message

### Import Albums Workflow

The `queue import` command processes bulk album additions:
1. User executes `queue import filename.txt`
2. CLI validates import file exists
3. Business logic reads both import file and existing queue
4. For each album in import file:
   - Validate format (Artist - Album)
   - Check for case-insensitive duplicates
   - Add to queue if valid and unique
   - Track added vs skipped counts
5. Update storage with new albums
6. Display summary of results to user

## REST API Spec

**Not Applicable** - This is a command-line application with no REST API endpoints.

## Database Schema

### File-Based Storage Schema

The application uses a simple text file format for data persistence:

**File: `queue.txt`**
- **Location:** `~/.music-queue/queue.txt` (default) or user-specified path
- **Format:** Plain text, one album per line
- **Encoding:** UTF-8
- **Structure:**
  ```
  Artist Name - Album Title
  Another Artist - Another Album
  Third Artist - Third Album Title
  ```

**Schema Rules:**
- Each line represents one album entry
- Format must be "Artist - Album" with at least one character before and after the dash
- Empty lines are ignored during processing
- Case-insensitive duplicate detection during import/add operations
- No metadata or additional fields stored

**File Operations:**
- **Read:** Entire file loaded into memory as string slice
- **Write:** Complete file overwrite with updated album list
- **Backup:** No automatic backup (relies on user's file system backup strategy)

## Source Tree

```plaintext
music-queue/
├── .github/                    # CI/CD workflows (future)
│   └── workflows/
│       └── main.yml
├── docs/                       # Project documentation
│   ├── architecture.md         # This document
│   ├── prd.md                  # Product Requirements
│   └── project-brief.md        # Project overview
├── src/                        # Application source code
│   ├── cmd/                    # CLI entry points
│   │   └── queue/              # Main queue command
│   │       ├── main.go         # CLI implementation and command parsing
│   │       └── main_test.go    # CLI integration tests
│   └── internal/               # Private application packages
│       ├── queue/              # Core business logic
│       │   ├── queue.go        # Queue service implementation
│       │   └── queue_test.go   # Business logic unit tests
│       └── storage/            # Data persistence layer
│           ├── file.go         # File storage implementation
│           └── file_test.go    # Storage layer unit tests
├── .gitignore                  # Git ignore rules
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── LICENSE                     # Project license
└── README.md                   # Project documentation
```

**Key Design Decisions:**
- **`internal/` packages:** Prevents external imports, maintains encapsulation
- **Layered package structure:** Clear separation of concerns between cmd, business logic, and storage
- **Test files co-located:** Each `.go` file has corresponding `_test.go` for easy testing
- **Standard Go conventions:** Follows Go community standards for project layout

## Infrastructure and Deployment

### Infrastructure as Code

- **Tool:** Not Required
- **Location:** Local development environment only
- **Approach:** Single binary distribution, no infrastructure needed

### Deployment Strategy

- **Strategy:** Single binary compilation and distribution
- **CI/CD Platform:** GitHub Actions (planned for future automation)
- **Pipeline Configuration:** `.github/workflows/main.yml`

### Environments

- **Development:** Local development machine with Go installed
- **Production:** User's local machine running compiled binary
- **Testing:** Local testing during development phase

### Environment Promotion Flow

```text
Developer Machine → Build → Binary → User's Machine
```

### Rollback Strategy

- **Primary Method:** User can revert to previous binary version
- **Trigger Conditions:** User-initiated when issues are discovered
- **Recovery Time Objective:** Immediate (local binary replacement)

## Error Handling Strategy

### General Approach

- **Error Model:** Go's explicit error handling with wrapped errors
- **Exception Hierarchy:** Standard Go error interface with custom error types
- **Error Propagation:** Errors bubble up through layers with additional context

### Logging Standards

- **Library:** Standard `fmt` package for user output
- **Format:** Human-readable console output for CLI users
- **Levels:** Error messages to stderr, success/info messages to stdout
- **Required Context:**
  - Command being executed
  - File paths involved
  - Clear user-actionable error messages

### Error Handling Patterns

#### File I/O Errors

- **File Not Found:** Clear message with absolute path shown
- **Permission Errors:** Descriptive message suggesting permission fixes
- **Disk Space:** Graceful handling with helpful error messages
- **Corruption Prevention:** Atomic file operations where possible

#### Business Logic Errors

- **Invalid Format:** Clear format requirements shown to user
- **Duplicate Albums:** Informational message, not treated as error
- **Empty Queue:** Friendly message when no albums available

#### CLI Errors

- **Invalid Commands:** Show usage help automatically
- **Missing Arguments:** Clear indication of what's required
- **Flag Parsing:** Descriptive error messages with examples

## Coding Standards

### Core Standards

- **Language & Runtime:** Go 1.24.4 with standard library only
- **Style & Linting:** `gofmt` for formatting, `golangci-lint` for static analysis
- **Test Organization:** `_test.go` files co-located with source files

### Naming Conventions

| Element   | Convention           | Example           |
| :-------- | :------------------- | :---------------- |
| Variables | camelCase            | albumTitle        |
| Functions | camelCase (exported: PascalCase) | addAlbum, AddAlbum |
| Types     | PascalCase           | QueueService      |
| Files     | snake_case           | queue_test.go     |
| Packages  | lowercase            | queue, storage    |

### Critical Rules

- **Error Handling:** Always check and handle errors explicitly, never ignore
- **Input Validation:** All external inputs must be validated before processing
- **File Operations:** Use atomic operations where possible to prevent corruption
- **Testing:** All public functions must have corresponding unit tests
- **Documentation:** All exported functions must have clear godoc comments

## Test Strategy and Standards

### Testing Philosophy

- **Approach:** Test-driven development encouraged, comprehensive test coverage required
- **Coverage Goals:** >90% unit test coverage, 100% of critical paths covered
- **Test Pyramid:** Focus on unit tests, integration tests for file operations

### Test Types and Organization

#### Unit Tests

- **Framework:** Go testing package (`testing`)
- **File Convention:** `*_test.go` co-located with source
- **Location:** Same package as source code
- **Mocking Library:** Manual interface mocking (no external dependencies)
- **Coverage Requirement:** >90% for business logic packages

**AI Agent Requirements:**
- Generate tests for all public methods and critical private functions
- Cover edge cases including empty inputs, invalid formats, file errors
- Follow AAA pattern (Arrange, Act, Assert)
- Mock storage layer for business logic tests

#### Integration Tests

- **Scope:** End-to-end CLI command testing with real file system
- **Location:** `cmd/queue/main_test.go`
- **Test Infrastructure:** Temporary directories and files for isolated testing

#### Test Data Management

- **Strategy:** Generate test data programmatically in tests
- **Fixtures:** Minimal use, prefer generated data for flexibility
- **Cleanup:** Automatic cleanup of temporary files in tests

### Continuous Testing

- **CI Integration:** Run all tests on every commit (future GitHub Actions)
- **Performance Tests:** Basic performance validation for file operations
- **Security Tests:** Input validation and file permission testing

## Security

### Input Validation

- **Validation Approach:** Strict format validation for album entries
- **Location:** Business logic layer before any storage operations
- **Required Rules:**
  - All album inputs must match "Artist - Album" format
  - File paths must be validated against directory traversal attacks
  - Command arguments must be properly sanitized

### Authentication & Authorization

- **Auth Method:** Operating system file permissions
- **Session Management:** Not applicable (stateless CLI tool)
- **Required Patterns:**
  - Respect OS file permissions for queue file access
  - Use user's home directory for default storage location

### Secrets Management

- **Development:** No secrets required
- **Production:** No secrets required
- **Code Requirements:**
  - No sensitive data stored or processed
  - No network communications requiring authentication

### Data Protection

- **Encryption at Rest:** Relies on OS-level file system encryption
- **Encryption in Transit:** Not applicable (local file operations only)
- **PII Handling:** Music album titles are not considered sensitive data
- **Logging Restrictions:** No sensitive data logged (album titles are public information)

### Dependency Security

- **Scanning:** Go mod with standard library only minimizes attack surface
- **Update Policy:** Follow Go security updates and best practices
- **Approval Process:** No external dependencies to manage

## Next Steps

### Developer Handoff

**Prompt for Development Agent:**

"Begin implementing the Music Queue CLI application following the architecture document. Start with the Storage Layer (`internal/storage/file.go`) as it has no dependencies, then implement the Business Logic Layer (`internal/queue/queue.go`), and finally the CLI Layer (`cmd/queue/main.go`). Follow the coding standards strictly, ensure comprehensive test coverage, and implement all five core commands: add, import, next, list, and count. All business logic should be thoroughly tested with both unit and integration tests."

**Key Implementation Priorities:**
1. Set up proper Go module structure
2. Implement file storage with atomic operations
3. Create comprehensive business logic with proper error handling
4. Build intuitive CLI interface with clear help text
5. Ensure robust testing at all layers

**Reference Documents:**
- This architecture document for technical decisions and patterns
- `docs/prd.md` for functional requirements and acceptance criteria
- Coding standards section for mandatory development practices

