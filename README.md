# updateCursor

A modern, configurable Go implementation of the Cursor editor updater with beautiful progress bars and flexible configuration options.

## Features

- 🚀 **Fast & Efficient**: Written in Go for optimal performance
- 📁 **Configurable Paths**: Customize download locations, filenames, and symlinks
- 📊 **Beautiful Progress Bars**: Real-time download progress with speed and size information
- 🔄 **Smart Version Management**: Automatic version detection and symlink management
- 📝 **Comprehensive Logging**: Track all updates in a TSV-formatted ledger
- 🧪 **Test-Driven Development**: Built with comprehensive test coverage
- 🐧 **Linux Optimized**: Designed specifically for Linux AppImage management

## What It Does

updateCursor automatically downloads and manages Cursor editor versions on Linux systems. It:

- Detects the latest available Cursor version
- Downloads AppImage files with progress tracking
- Manages version-specific files and symlinks
- Tracks update history in a configurable ledger
- Provides flexible configuration for different use cases

## Installation

### Prerequisites

- Go 1.23.3 or later
- Linux system with AppImage support

### Build from Source

```bash
# Clone the repository
git clone https://github.com/CoGorm/updateCursor.git
cd updateCursor

# Build the binary
go build -o updatecursor ./cmd/updatecursor

# Make it executable
chmod +x updatecursor

# Move to a directory in your PATH (optional)
sudo mv updatecursor /usr/local/bin/
```

## Usage

### Basic Commands

```bash
# Check current versions (exits with status 10 if update needed)
./updatecursor check

# Download latest version if newer (default command)
./updatecursor update

# Force re-download latest version
./updatecursor force

# List update history
./updatecursor list

# Switch to specific version
./updatecursor switch 1.4.5

# Show help
./updatecursor --help
```

### Default Behavior

By default, updateCursor:
- Downloads to `~/Downloads/Cursor/`
- Names files as `Cursor-<version>-x86_64.AppImage`
- Creates symlink at `~/Downloads/Cursor/Cursor.AppImage`
- Stores ledger at `~/.config/updateCursor/cursor-versions.log`

## Configuration

updateCursor supports flexible configuration through a YAML file. The configuration file is located at:
```
~/.config/updateCursor/config.yaml
```

If the file doesn't exist, updateCursor will create it with default values on first run.

### Configuration Options

| Setting | Description | Default Value |
|---------|-------------|---------------|
| `download_dir` | Directory where Cursor AppImage files are stored | `~/Downloads/Cursor` |
| `file_name_pattern` | Pattern for downloaded filenames (use `<version>` placeholder) | `Cursor-<version>-x86_64.AppImage` |
| `latest_symlink` | Path to symlink pointing to current version | `~/Downloads/Cursor/Cursor.AppImage` |
| `ledger_path` | Path to update history log file | `~/.config/updateCursor/cursor-versions.log` |

### Example Configurations

#### Custom User Setup
```yaml
download_dir: "~/Applications/Cursor"
file_name_pattern: "Cursor_<version>.AppImage"
latest_symlink: "~/.local/bin/Cursor.AppImage"
ledger_path: "~/.config/updateCursor/cursor-versions.log"
```

#### System-wide Installation
```yaml
download_dir: "/opt/cursor/versions"
file_name_pattern: "cursor-<version>-linux-x64.AppImage"
latest_symlink: "/usr/local/bin/cursor"
ledger_path: "/var/log/cursor-updates.log"
```

#### Portable Setup
```yaml
download_dir: "./cursor-versions"
file_name_pattern: "Cursor-<version>.AppImage"
latest_symlink: "./cursor"
ledger_path: "./cursor-versions.log"
```

See `config.example.yaml` for more examples and detailed explanations.

## Building

### Using Makefile (Recommended)
```bash
# Build the binary
make build

# Build for release (stripped binary)
make build-release

# Cross-compile for Linux
make build-linux

# Clean build artifacts
make clean

# See all available targets
make help
```

### Manual Build
```bash
# Development build
go build -o updatecursor ./cmd/updatecursor

# Release build (stripped binary)
go build -ldflags="-s -w" -o updatecursor ./cmd/updatecursor

# Cross-compile for other platforms
GOOS=linux GOARCH=amd64 go build -o updatecursor-linux-amd64 ./cmd/updatecursor
```

## Development

### Testing

The project follows Test-Driven Development (TDD) principles with comprehensive test coverage.

```bash
# Using Makefile
make test              # Run all tests
make test-verbose      # Run tests with verbose output
make test-race         # Run tests with race detection
make test-coverage     # Run tests with coverage report

# Manual testing
go test ./...          # Run all tests
go test -v ./...       # Run tests with verbose output
go test -v ./internal/updater  # Run tests for specific package
go test -v -run TestDownloadCursor ./internal/updater  # Run specific test
```

### Code Quality

```bash
# Linting (requires golangci-lint)
make lint

# Format code
go fmt ./...
goimports -w .

# Run all quality checks
make all  # test + lint + build
```

### Test Coverage

- **internal/config**: Configuration loading, validation, and path expansion
- **internal/ledger**: TSV ledger management and entry operations
- **internal/updater**: Download, version management, and symlink operations
- **internal/version**: Semantic version parsing and comparison
- **cmd/updatecursor**: CLI command parsing and execution

## Project Structure

```
.
├── cmd/
│   └── updatecursor/
│       ├── main.go          # Thin CLI entrypoint
│       └── main_test.go     # Main package tests
├── internal/
│   ├── cli/                 # CLI logic and command handling
│   │   ├── run.go
│   │   └── run_test.go
│   ├── config/              # Configuration management
│   │   ├── config.go
│   │   ├── config_test.go
│   │   └── example_test.go
│   ├── ledger/              # Version history tracking
│   │   ├── ledger.go
│   │   └── ledger_test.go
│   ├── updater/             # Core update logic
│   │   ├── updater.go
│   │   └── updater_test.go
│   └── version/             # Semantic versioning utilities
│       ├── version.go
│       └── version_test.go
├── config.example.yaml       # Example configuration file
├── Makefile                  # Common development tasks
├── .golangci.yml            # Linting configuration
├── go.mod                   # Go module definition
├── LICENSE                  # MIT License
└── README.md                # This file
```

### Key Design Principles

- **Separation of Concerns**: CLI logic separated from main entrypoint
- **Testability**: All packages have comprehensive unit tests
- **Configuration**: Flexible YAML-based configuration system
- **Error Handling**: Proper error propagation and user-friendly messages
- **Progress Feedback**: Real-time download progress with beautiful UI

## Migration from Bash Scripts

This Go implementation replaces the original bash scripts:
- `updateCursor.sh` - Comprehensive updater with ledger tracking
- `updateCursor_simple.sh` - Simple update functionality

The Go version provides:
- Better error handling and user feedback
- Configurable paths and naming
- Progress bars during downloads
- Cross-platform compatibility
- Easier maintenance and extension

## Contributing

1. Fork the repository
2. Create a feature branch
3. Follow TDD principles (write tests first)
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Original bash script authors for the concept and functionality
- Go community for excellent tooling and libraries
- Cursor team for providing the excellent editor
