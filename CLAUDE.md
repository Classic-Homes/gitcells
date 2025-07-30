# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SheetSync is a Go application that bridges Excel and Git by converting Excel files to human-readable JSON for version control. The project is in early development stage with only the initial module setup completed.

## Build and Development Commands

```bash
# Initialize project dependencies
go mod tidy

# Build the application (once main.go exists)
go build -o dist/sheetsync cmd/sheetsync/main.go

# Cross-platform builds (from planned Makefile)
make build    # Builds for Mac, Windows, and Linux

# Run tests
go test -v ./...

# Clean build artifacts
make clean
```

## Architecture and Structure

The planned architecture follows Go best practices with clear separation of concerns:

- **cmd/sheetsync/**: Entry point and CLI commands using Cobra framework
- **internal/**: Private application code
  - **converter/**: Excel↔JSON conversion logic (core functionality)
  - **git/**: Git operations wrapper using go-git
  - **watcher/**: File system monitoring with fsnotify
  - **config/**: Configuration management with Viper
- **pkg/models/**: Public data models for Excel document representation

## Key Dependencies

- **github.com/xuri/excelize/v2**: Excel file manipulation
- **github.com/go-git/go-git/v5**: Git operations
- **github.com/fsnotify/fsnotify**: File system watching
- **github.com/spf13/cobra**: CLI framework
- **github.com/spf13/viper**: Configuration management

## Implementation Guidelines

1. **Excel Handling**: The converter must preserve formulas, styles, merged cells, and other Excel features in the JSON representation
2. **Git Integration**: Automatic commits should use descriptive messages and handle merge conflicts intelligently
3. **File Watching**: Must ignore Excel temporary files (~$*.xlsx) and implement debouncing for rapid changes
4. **Cross-Platform**: Use filepath package for path handling to ensure Windows/Mac/Linux compatibility

## Current Status

The project has been initialized with go.mod but no code has been implemented yet. Follow the implementation guide in excel-git-sync-guide.md for the development roadmap.