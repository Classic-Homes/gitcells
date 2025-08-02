# Command Reference

This reference documents all GitCells commands, their options, and usage examples.

## Global Options

These options work with all commands:

```bash
gitcells [global options] command [command options] [arguments...]
```

### Global Flags

- `--config string` - Specify configuration file path (default: `.gitcells.yaml`)
- `--verbose` - Enable verbose logging
- `--tui` - Launch Terminal User Interface
- `--help, -h` - Show help
- `--version, -v` - Show version information

## Commands Overview

| Command | Description |
|---------|-------------|
| `init` | Initialize GitCells in a directory |
| `watch` | Watch directories for Excel file changes |
| `convert` | Convert between Excel and JSON formats |
| `sync` | Synchronize Excel files with their JSON representations |
| `status` | Show status of tracked files |
| `diff` | Show differences between file versions |
| `update` | Update GitCells to the latest version |
| `version` | Display version information |
| `tui` | Launch Terminal User Interface |

## init

Initialize GitCells configuration in a directory.

### Synopsis

```bash
gitcells init [directory] [flags]
```

### Description

Creates a `.gitcells.yaml` configuration file and optionally initializes a Git repository. If no directory is specified, uses the current directory.

### Flags

- `--force` - Overwrite existing configuration
- `--git` - Initialize Git repository (default: true)
- `--tui` - Use TUI setup wizard

### Examples

```bash
# Initialize in current directory
gitcells init

# Initialize in specific directory
gitcells init /path/to/excel/files

# Force overwrite existing config
gitcells init --force

# Skip Git initialization
gitcells init --git=false

# Use interactive setup wizard
gitcells init --tui
```

### Output

Creates:
- `.gitcells.yaml` - Configuration file
- `.gitignore` - Git ignore patterns (if Git enabled)

## watch

Monitor directories for Excel file changes.

### Synopsis

```bash
gitcells watch [directories...] [flags]
```

### Description

Starts file system monitoring for specified directories. Automatically converts Excel files to JSON when changes are detected.

### Flags

- `--auto-commit` - Automatically commit changes to Git (default: true)
- `--auto-push` - Automatically push commits to remote (default: false)

### Examples

```bash
# Watch current directory
gitcells watch .

# Watch multiple directories
gitcells watch ./reports ./data /shared/excel

# Watch without auto-commit
gitcells watch --auto-commit=false .

# Watch with auto-push
gitcells watch --auto-push=true .

# Watch with custom config
gitcells watch --config prod.yaml ./production
```

### Behavior

1. Monitors specified directories recursively
2. Detects create, modify, and delete events
3. Applies debounce delay from configuration
4. Converts modified Excel files to JSON
5. Optionally commits changes to Git

## convert

Convert between Excel and JSON formats.

### Synopsis

```bash
gitcells convert <file> [flags]
```

### Description

Converts Excel files to JSON chunks or JSON chunks back to Excel. Direction is determined by input type.

### Flags

- `-o, --output string` - Output file path (auto-generated if not specified)
- `--preserve-formulas` - Preserve Excel formulas (default: true)
- `--preserve-styles` - Preserve cell styles (default: true)
- `--preserve-comments` - Preserve cell comments (default: true)
- `--compact` - Output compact JSON (default: false)

### Examples

```bash
# Convert Excel to JSON chunks
gitcells convert Budget.xlsx

# Convert with specific output location
gitcells convert Budget.xlsx -o /backup/Budget.xlsx

# Convert JSON chunks back to Excel
gitcells convert .gitcells/data/Budget.xlsx_chunks/

# Compact JSON output
gitcells convert Report.xlsx --compact

# Minimal conversion (data only)
gitcells convert Data.xlsx \
  --preserve-formulas=false \
  --preserve-styles=false \
  --preserve-comments=false
```

### File Storage

- Excel → JSON: Creates chunks in `.gitcells/data/file.xlsx_chunks/`
- JSON → Excel: Reads chunks from `.gitcells/data/` to create Excel file

## sync

Synchronize Excel files with their JSON representations.

### Synopsis

```bash
gitcells sync [directories...] [flags]
```

### Description

Ensures all Excel files have up-to-date JSON representations and vice versa. Useful after pulling changes from Git.

### Flags

- `--direction string` - Sync direction: "both", "excel-to-json", "json-to-excel" (default: "both")
- `--force` - Force overwrite newer files

### Examples

```bash
# Sync current directory
gitcells sync .

# Sync specific directories
gitcells sync ./reports ./data

# Only update JSON from Excel
gitcells sync --direction excel-to-json .

# Only restore Excel from JSON
gitcells sync --direction json-to-excel .

# Force sync even if destination is newer
gitcells sync --force .
```

### Sync Logic

1. Compares timestamps of Excel and JSON chunk files
2. Updates older files from newer files
3. Creates missing JSON chunks from Excel
4. Recreates missing Excel files from JSON chunks

## status

Show status of tracked Excel files.

### Synopsis

```bash
gitcells status [flags]
```

### Description

Displays information about tracked Excel files, their JSON representations, and Git status.

### Flags

- `--detailed` - Show detailed file information
- `--format string` - Output format: "table", "json", "yaml" (default: "table")

### Examples

```bash
# Show basic status
gitcells status

# Show detailed information
gitcells status --detailed

# Output as JSON
gitcells status --format json

# Output as YAML
gitcells status --format yaml
```

### Output Information

- File count and total size
- Last modification times
- Sync status (up-to-date, needs update)
- Git status (committed, modified, untracked)
- Conversion errors

## diff

Show differences between Excel file versions.

### Synopsis

```bash
gitcells diff [file] [flags]
```

### Description

Compares Excel files by examining their JSON representations. Can compare with Git history or between specific versions.

### Flags

- `--from string` - Source version (Git ref or file path)
- `--to string` - Target version (Git ref or file path) (default: "working")
- `--format string` - Output format: "unified", "side-by-side", "json" (default: "unified")
- `--sheets string` - Comma-separated list of sheets to compare

### Examples

```bash
# Compare with last commit
gitcells diff Budget.xlsx

# Compare with specific commit
gitcells diff Budget.xlsx --from HEAD~3

# Compare two commits
gitcells diff Budget.xlsx --from HEAD~3 --to HEAD

# Compare specific sheets only
gitcells diff Report.xlsx --sheets "Summary,Data"

# Side-by-side diff
gitcells diff Sales.xlsx --format side-by-side

# JSON output for processing
gitcells diff Data.xlsx --format json
```

### Diff Output

Shows:
- Changed cell values
- Modified formulas
- Style changes
- Added/removed cells
- Sheet structure changes

## update

Update GitCells to the latest version.

### Synopsis

```bash
gitcells update [flags]
```

### Description

Checks for and installs updates from GitHub releases.

### Flags

- `--check` - Only check for updates, don't install
- `--force` - Skip confirmation prompt
- `--prerelease` - Include pre-release versions

### Examples

```bash
# Check for updates
gitcells update --check

# Update to latest stable
gitcells update

# Update without confirmation
gitcells update --force

# Update to latest including pre-releases
gitcells update --prerelease

# Check for pre-releases
gitcells update --prerelease --check
```

### Update Process

1. Checks GitHub for latest release
2. Compares with current version
3. Downloads appropriate binary
4. Verifies checksum
5. Replaces current binary
6. Preserves configuration

## version

Display version information.

### Synopsis

```bash
gitcells version [flags]
```

### Description

Shows current GitCells version and optionally checks for updates.

### Flags

- `--check-update` - Check for available updates
- `--prerelease` - Include pre-release versions when checking

### Examples

```bash
# Show version
gitcells version

# Show version and check for updates
gitcells version --check-update

# Check including pre-releases
gitcells version --check-update --prerelease
```

### Output

```
GitCells version 0.3.0 (built 2024-01-15)
✅ You are running the latest version
```

## tui

Launch the Terminal User Interface.

### Synopsis

```bash
gitcells tui
```

### Description

Opens an interactive terminal interface for managing GitCells operations.

### Examples

```bash
# Launch TUI
gitcells tui

# Launch TUI with custom config
gitcells --config custom.yaml tui
```

### TUI Features

- Setup wizard
- Status dashboard
- Error log viewer
- Settings management
- Interactive configuration

## Environment Variables

GitCells respects these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `GITCELLS_CONFIG` | Configuration file path | `.gitcells.yaml` |
| `GITCELLS_LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `GITCELLS_NO_COLOR` | Disable colored output | `false` |
| `GITCELLS_WORKER_THREADS` | Number of worker threads | `4` |
| `GITCELLS_CACHE_DIR` | Cache directory | `.gitcells.cache` |

## Exit Codes

GitCells uses standard exit codes:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | File not found |
| 4 | Permission denied |
| 5 | Git error |
| 6 | Conversion error |
| 7 | Network error |

## Shell Completion

Enable command completion for your shell:

### Bash

```bash
gitcells completion bash > /etc/bash_completion.d/gitcells
```

### Zsh

```bash
gitcells completion zsh > "${fpath[1]}/_gitcells"
```

### Fish

```bash
gitcells completion fish > ~/.config/fish/completions/gitcells.fish
```

### PowerShell

```powershell
gitcells completion powershell | Out-String | Invoke-Expression
```

## Next Steps

- Review [Configuration Reference](configuration.md) for all options
- Understand the [JSON Format](json-format.md)
- Learn about [API Usage](api.md) for programmatic access
- Check [Troubleshooting](../user-guide/troubleshooting.md) for issues