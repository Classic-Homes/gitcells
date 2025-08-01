# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GitCells is a production-ready Go application that bridges Excel and Git by converting Excel files to human-readable JSON for version control. The project includes a full CLI, interactive TUI, automatic file watching, smart chunking for large files, and self-updating capabilities.

## Build and Development Commands

```bash
# Initialize project dependencies
go mod tidy

# Build the application
make build            # Build for current platform
make build-all        # Build for all platforms
make release          # Create release artifacts

# Run tests
make test             # Run all tests
make test-short       # Skip integration tests
make test-coverage    # Generate coverage report

# Development tools
make lint             # Run linters
make fmt              # Format code
make check            # Run all checks (lint, fmt, test)

# Installation
make install          # Install to /usr/local/bin
make install-dev      # Install with development flags

# Docker operations
make docker-build     # Build Docker image
make docker-push      # Push to registry

# Documentation
./scripts/serve-docs.sh           # Start MkDocs server
./scripts/build-docs.sh           # Build static docs
docker-compose up docs            # Alternative docs server

# Clean build artifacts
make clean

# GitCells runtime commands
gitcells tui                      # Launch interactive TUI
gitcells init                     # Initialize in current directory
gitcells watch                    # Watch for Excel file changes
gitcells convert file.xlsx        # Convert Excel to JSON
gitcells status                   # Show sync status
gitcells diff file.xlsx           # Show differences
gitcells sync                     # Sync with Git
gitcells update                   # Self-update to latest version
gitcells version --check-update   # Check for updates
```

## Architecture and Structure

The architecture follows Go best practices with clear separation of concerns:

- **cmd/gitcells/**: Entry point and CLI commands using Cobra framework
  - `main.go`: Application entry point with command registration
  - `init.go`: Repository initialization
  - `convert.go`: Excel to JSON conversion command
  - `watch.go`: File system watcher command
  - `sync.go`: Git synchronization
  - `status.go`: Status display
  - `diff.go`: Difference viewer
  - `update.go`: Self-update functionality
  - `constants.go`: Command constants
- **internal/**: Private application code
  - **converter/**: Excel↔JSON conversion logic
    - `excel_to_json.go`: Excel parsing and JSON generation
    - `json_to_excel.go`: JSON to Excel reconstruction
    - `chunking.go`: Smart file chunking for large workbooks
    - `types.go`: Converter data structures
  - **git/**: Git operations wrapper using go-git
    - `client.go`: Git client implementation
  - **watcher/**: File system monitoring with fsnotify
    - `watcher.go`: File system event handling
    - `debouncer.go`: Intelligent debouncing for rapid changes
  - **config/**: Configuration management with Viper
    - `config.go`: Configuration loading and validation
    - `defaults.go`: Default configuration values
  - **updater/**: Self-update functionality
    - `updater.go`: GitHub releases integration
  - **tui/**: Terminal User Interface
    - `app.go`: Main TUI application
    - `models/`: TUI models (dashboard, settings, setup, error_log)
    - `components/`: Reusable UI components
    - `views/`: Screen views
    - `styles/`: Theme and styling
    - `adapter/`: Business logic adapters
  - **utils/**: Utility functions
    - `errors.go`: Error handling utilities
    - `logging.go`: Logging configuration
  - **constants/**: Application constants
    - `version.go`: Version information
- **pkg/models/**: Public data models
  - `excel.go`: Excel document representation
  - `diff.go`: Diff data structures
- **docs/**: MkDocs documentation
- **scripts/**: Build and deployment scripts
- **test/**: Integration tests and test data

## Key Dependencies

- **github.com/xuri/excelize/v2**: Excel file manipulation
- **github.com/go-git/go-git/v5**: Git operations
- **github.com/fsnotify/fsnotify**: File system watching
- **github.com/spf13/cobra**: CLI framework
- **github.com/spf13/viper**: Configuration management
- **github.com/charmbracelet/bubbletea**: Terminal UI framework
- **github.com/charmbracelet/bubbles**: TUI components
- **github.com/charmbracelet/lipgloss**: Terminal styling
- **github.com/sirupsen/logrus**: Structured logging
- **github.com/stretchr/testify**: Testing assertions
- **github.com/Masterminds/semver/v3**: Semantic version comparison

## Implementation Guidelines

1. **Excel Handling**: 
   - Preserves formulas, styles, merged cells, comments, and other Excel features
   - Implements smart chunking for large files to optimize Git performance
   - Stores JSON representations in `.gitcells/data/` directory
   - Maintains bidirectional conversion fidelity

2. **Git Integration**: 
   - Automatic commits with customizable templates
   - Smart conflict resolution for concurrent edits
   - Integration with existing Git workflows
   - Support for hooks and CI/CD pipelines

3. **File Watching**: 
   - Ignores Excel temporary files (~$*.xlsx)
   - Implements configurable debouncing (default 2s)
   - Supports multiple watch directories
   - Handles file system events efficiently

4. **Cross-Platform**: 
   - Uses filepath package for path handling
   - Platform-specific installers and update mechanisms
   - Consistent behavior across Windows/Mac/Linux

5. **Error Handling**:
   - Comprehensive error logging with context
   - User-friendly error messages
   - Error recovery mechanisms
   - Debug mode for troubleshooting

6. **Testing**:
   - Unit tests for all major components
   - Integration tests for end-to-end workflows
   - Test data in `test/testdata/`
   - Coverage reporting

## Development Workflow

1. **Feature Development**:
   ```bash
   git checkout -b feature/your-feature
   make test
   make lint
   git commit -m "feat: your feature description"
   ```

2. **Testing Changes**:
   ```bash
   make test              # Run all tests
   make test-short        # Quick tests only
   go test ./internal/converter -v  # Test specific package
   ```

3. **Building**:
   ```bash
   make build             # Build for current platform
   make build-all         # Build all platforms
   ./dist/gitcells --help # Test the binary
   ```

4. **Documentation**:
   - Update relevant .md files in docs/
   - Run `./scripts/serve-docs.sh` to preview
   - Ensure examples are current

## Release Process

1. Update version in Makefile
2. Run `make release` to create artifacts
3. Test release binaries with `./scripts/test-releases.sh`
4. Create GitHub release with artifacts
5. Update documentation if needed

## Current Status

The project is fully implemented with:
- Complete CLI with all planned commands
- Interactive TUI for user-friendly operations  
- Excel ↔ JSON conversion with chunking
- File watching with debouncing
- Git integration
- Self-update system
- Comprehensive documentation
- Cross-platform support