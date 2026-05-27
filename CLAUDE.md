# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GitCells is a Go CLI that bridges Excel and Git by converting Excel files to human-readable JSON for version control. Scope is deliberately small: the core Excel→JSON→Git pipeline must work perfectly on Windows, macOS, and Linux. Features: CLI surface (`init`, `convert`, `watch`, `sync`, `status`, `diff`, `update`, `version`), cross-platform file watching with atomic-save handling, smart per-sheet chunking for clean Git diffs.

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

The release process is now version-driven and automated:

### 1. Pre-release Creation (Automatic)
1. Update the VERSION file with your new version (e.g., `1.0.1`)
2. Commit and push to main branch
3. CI automatically detects VERSION change and:
   - Creates a version tag (e.g., `v1.0.1`)
   - Builds all platform artifacts
   - Creates a pre-release with all binaries

### 2. Production Release (Manual Promotion)
1. Test the pre-release thoroughly
2. Go to GitHub Actions → "Promote Release" workflow
3. Click "Run workflow" and enter the version number
4. The workflow promotes the pre-release to production release

### 3. Version Management
```bash
# Check current version
cat VERSION

# Update version for next release
echo "1.0.1" > VERSION
git add VERSION
git commit -m "bump version to 1.0.1"
git push

# The CI will automatically create v1.0.1 pre-release
```

### 4. Manual Release (if needed)
For manual releases outside the automated process:
1. Run `make release` to create artifacts
2. Test release binaries with `./scripts/test-releases.sh`
3. Create GitHub release with artifacts
4. Update documentation if needed

## Current Status

The project is implemented with:
- CLI commands: init, convert, watch, sync, status, diff, update, version
- Excel ↔ JSON conversion with per-sheet chunking and atomic staging-dir installs
- File watching with debouncing, Windows lock-retry on open
- Git integration via go-git
- Self-update check (no in-place self-replace; package-manager / release page install)
- Cross-platform CI matrix (Linux / macOS / Windows)

## Path Portability

Chunk metadata (`.gitcells_chunks.json`) stores chunk file paths as
forward slashes regardless of the writing OS. The reader applies
`filepath.FromSlash` before `filepath.Join` so chunk trees authored on
one OS read back correctly on another. Do not break this contract.