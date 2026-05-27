
# GitCells

<p align="center">
  <img src="docs/assets/logo.png" alt="GitCells Logo" width="200">
</p>

<p align="center">
  <strong>Version Control for Excel Files</strong><br>
  <em>Track, diff, and collaborate on spreadsheets with Git</em>
</p>

<p align="center">
  <a href="https://github.com/Classic-Homes/gitcells/releases"><img src="https://img.shields.io/github/v/release/Classic-Homes/gitcells" alt="Release"></a>
  <a href="https://github.com/Classic-Homes/gitcells/blob/main/LICENSE"><img src="https://img.shields.io/github/license/Classic-Homes/gitcells" alt="License"></a>
  <a href="https://github.com/Classic-Homes/gitcells/actions"><img src="https://img.shields.io/github/actions/workflow/status/Classic-Homes/gitcells/ci.yml" alt="Build Status"></a>
</p>

GitCells bridges Excel and Git, enabling version control for spreadsheets. It converts Excel files to human-readable JSON for diffing and merging, then restores them to native Excel format for editing. No more binary conflicts, lost formulas, or overwritten work.

GitCells pairs well with GUI Git clients like GitHub Desktop — it focuses on the watch → convert → commit pipeline and leaves general Git operations (push, pull, branch, merge) to dedicated tools.

## Scope

**Preserved on round-trip:**

- Cell values, formulas (A1 style)
- Basic styling: fonts, fills, borders, alignment, number formats
- Merged cells, comments, hyperlinks, data validation
- Multi-sheet workbooks, unicode sheet names

**NOT preserved (logged as a warning per sheet at conversion time):**

- Charts, pivot tables, conditional formatting, tables (ListObjects)
- Rich-text-per-run formatting, R1C1 formulas, array formula spill ranges
- Embedded pictures, macros, VBA

If your workbook depends on those, GitCells commits the cells and skips the rest — the conversion log names what was dropped.

## Features

- **Excel ↔ JSON Conversion**: Converts Excel files to structured JSON
- **Smart Chunking**: Splits workbooks per-sheet for clean Git diffs; sheet-name collisions are disambiguated with a content hash
- **File Watching**: Monitors directories with cross-platform-aware debouncing — handles Excel's atomic save sequence (write-temp / rename-over / unlink) on Windows, macOS, and Linux
- **Git Integration**: Auto-commits converted JSON; concurrent saves are serialized to prevent `index.lock` corruption
- **Minimal TUI**: Single-screen dashboard with watcher start/stop, sync status, files list, activity log
- **Diff Generation**: Text or JSON diffs between Excel files
- **Update Check**: `gitcells version --check-update` reports new releases — install via your package manager or the release page (no in-place self-replace)
- **Cross-Platform**: Windows, macOS, Linux all CI-tested per commit

## Installation

### Pre-built Binaries

Download the latest release for your platform from [GitHub Releases](https://github.com/Classic-Homes/gitcells/releases).

### Quick Install Scripts

**macOS/Linux:**
```bash
curl -sSL https://raw.githubusercontent.com/Classic-Homes/gitcells/main/scripts/install.sh | sh
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/Classic-Homes/gitcells/main/scripts/install.ps1 | iex
```

### Manual Download

Download the latest release from the [releases page](https://github.com/Classic-Homes/gitcells/releases):

- **Windows**: `gitcells-windows.exe`
- **macOS Intel**: `gitcells-macos-intel.tar.gz`
- **macOS Apple Silicon**: `gitcells-macos-apple-silicon.tar.gz`
- **Linux x64**: `gitcells-linux.tar.gz`
- **Linux ARM64**: `gitcells-linux-arm64.tar.gz`

### Build from Repository

```bash
git clone https://github.com/Classic-Homes/gitcells.git
cd gitcells
make build

# Install locally
make install
```

## Updating GitCells

GitCells does NOT self-replace its binary (Windows file locks make that unreliable). Use:

```bash
# Check for a newer release
gitcells version --check-update

# Or: gitcells update (prints download URL + release notes)
gitcells update
```

Install the new version via your package manager, the install script, or by downloading the binary from [Releases](https://github.com/Classic-Homes/gitcells/releases).

## Quick Start

### 1. Initialize GitCells in your project

```bash
cd your-excel-project
gitcells init
```

This creates a `.gitcells.yaml` configuration file and sets up Git if needed.

### 2. Convert Excel files to JSON

```bash
# Convert a single file
gitcells convert spreadsheet.xlsx

# Convert multiple files
gitcells convert *.xlsx

# Convert with options
gitcells convert --preserve-styles --preserve-comments data.xlsx
```

### 3. Watch directories for automatic conversion

```bash
# Watch current directory
gitcells watch .

# Watch specific directories
gitcells watch ./data ./reports

# Watch with auto-commit to Git
gitcells watch --auto-commit ./spreadsheets
```

### 4. Check synchronization status

```bash
gitcells status
```

### 5. Manually sync with Git

```bash
gitcells sync
```

### Check for updates

```bash
# Print download URL if newer release available
gitcells version --check-update

# Same, but include pre-releases
gitcells version --check-update --prerelease

# `gitcells update` is equivalent + prints release notes
gitcells update
```

GitCells does not replace its own binary — install the new release via your package manager or by downloading from GitHub Releases.

## Configuration

GitCells uses a `.gitcells.yaml` configuration file:

```yaml
version: "1.0"

git:
  branch: main
  user_name: "GitCells"
  user_email: "gitcells@yourcompany.com"

watcher:
  directories:
    - "./data"
    - "./reports"
  ignore_patterns:
    - "~$*"           # Excel temp files
    - "*.tmp"         # Temporary files
    - ".~lock.*"      # LibreOffice lock files
  debounce_delay: 2s  # Wait before processing changes (must exceed Excel save time)
  file_extensions:
    - ".xlsx"
    - ".xls"
    - ".xlsm"

converter:
  preserve_formulas: true     # Keep Excel formulas
  preserve_styles: true       # Keep cell formatting
  preserve_comments: true     # Keep cell comments
  compact_json: false         # Use pretty-printed JSON
  ignore_empty_cells: true    # Skip empty cells in JSON
  max_cells_per_sheet: 1000000  # Memory protection
  chunking_strategy: sheet-based
```

## Command Reference

### Global Options

- `--config`: Specify config file path (default: `.gitcells.yaml`)
- `--verbose`: Enable verbose logging
- `--tui`: Launch the dashboard TUI
- `--help`: Show help information

### Commands

#### `gitcells tui`

Launch the dashboard TUI. Single screen showing sync status, tracked files, recent activity, and a watcher start/stop toggle.

```bash
gitcells tui
```

Keys: `w` toggle watcher, `r` refresh, `tab` switch tab, `q` quit.

#### `gitcells init`

Initialize GitCells in the current directory.

```bash
gitcells init [flags]
```

**Options:**
- `--force`: Overwrite existing configuration

#### `gitcells convert`

Convert Excel files to JSON format.

```bash
gitcells convert [files...] [flags]
```

**Options:**
- `--output-dir`: Output directory for JSON files
- `--preserve-formulas`: Preserve Excel formulas (default: true)
- `--preserve-styles`: Preserve cell styles and formatting
- `--preserve-comments`: Preserve cell comments
- `--compact`: Generate compact JSON (no pretty printing)

**Examples:**
```bash
gitcells convert data.xlsx
gitcells convert --preserve-styles *.xlsx
gitcells convert --output-dir ./json-output reports.xlsx
```

#### `gitcells watch`

Watch directories for Excel file changes and automatically convert them.

```bash
gitcells watch [directories...] [flags]
```

**Options:**
- `--auto-commit`: Automatically commit changes to Git
- `--debounce`: Debounce delay (e.g., "2s", "500ms")

**Examples:**
```bash
gitcells watch .
gitcells watch --auto-commit ./data ./reports
gitcells watch --debounce 5s ./spreadsheets
```

#### `gitcells sync`

Manually synchronize Excel files with Git repository.

```bash
gitcells sync [flags]
```

**Options:**
- `--push`: Push changes to remote repository
- `--pull`: Pull changes from remote repository

#### `gitcells status`

Show the current synchronization status.

```bash
gitcells status [flags]
```

#### `gitcells diff`

Show differences between Excel file versions.

```bash
gitcells diff [file] [flags]
```

**Options:**
- `--from`: Compare from specific commit/version
- `--to`: Compare to specific commit/version

#### `gitcells update`

Print latest release info + download URL. Does not self-replace.

```bash
gitcells update [flags]
```

**Options:**
- `--prerelease`: Include pre-release versions

#### `gitcells version`

Show version information.

```bash
gitcells version [flags]
```

**Options:**
- `--check-update`: Check for available updates
- `--prerelease`: Include pre-release versions when checking

## Excel File Support

GitCells supports the following Excel features:

### ✅ Supported Features

- **File Formats**: `.xlsx`, `.xls`, `.xlsm`
- **Cell Values**: Text, numbers, booleans, dates
- **Formulas**: All Excel formulas including references between sheets
- **Cell Formatting**: Fonts, colors, borders, number formats
- **Merged Cells**: Cell ranges merged across rows/columns
- **Multiple Sheets**: Workbooks with multiple worksheets
- **Comments**: Cell comments and annotations
- **Named Ranges**: Defined names and ranges
- **Data Validation**: Cell validation rules
- **Hyperlinks**: Links to URLs or other cells

### ⚠️ Not Preserved (logged as warnings per sheet)

- **Charts and Graphs**: Skipped
- **Pivot Tables**: Skipped (structure was previously extracted but never rebuilt — silent data loss)
- **Conditional Formatting**: Skipped (style portion never round-tripped)
- **Tables (ListObjects)**: Skipped (column definitions incomplete in excelize)
- **Rich-Text-Per-Run**: Cell-level styles preserved, per-run formatting dropped
- **Macros / VBA**: Not converted
- **R1C1 Formulas**: Not yet supported by excelize
- **Embedded Images**: Skipped
- **External Links**: May break on rebuild

## JSON Format

GitCells uses an intelligent chunking system that automatically splits Excel files for optimal Git performance:

### Storage Organization

**Original Excel files**: Remain in their original locations
**JSON representations**: Stored in `.gitcells/data/` directory

This separation keeps your working directory clean while maintaining full version control.

### Chunking Strategy

For large workbooks, GitCells automatically splits data into manageable chunks:

```
.gitcells/data/
└── myworkbook_chunks/
    ├── workbook.json           # Metadata and workbook properties
    ├── sheet_Sheet1.json       # Individual sheet data
    ├── sheet_Sheet2.json       # Individual sheet data
    └── .gitcells_chunks.json   # Chunk tracking metadata
```

**Benefits:**
- Faster Git operations
- Reduced memory usage
- Better diff visibility
- Easier conflict resolution

### Main Workbook File (workbook.json)
```json
{
  "version": "1.0",
  "metadata": {
    "created": "2024-01-15T10:30:00Z",
    "modified": "2024-01-15T15:45:00Z",
    "app_version": "gitcells-0.1.0",
    "original_file": "data.xlsx",
    "file_size": 45678,
    "checksum": "abc123def456..."
  },
  "sheets": [
    {"name": "Sheet1", "index": 0, "hidden": false},
    {"name": "Sheet2", "index": 1, "hidden": false}
  ],
  "defined_names": {},
  "properties": {}
}
```

### Individual Sheet File (sheet_Sheet1.json)
```json
{
  "version": "1.0",
  "workbook_checksum": "abc123def456...",
  "sheet": {
    "name": "Sheet1",
    "index": 0,
    "cells": {
      "A1": {
        "value": "Product Name",
        "type": "string"
      },
      "A2": {
        "value": 123.45,
        "type": "number"
      },
      "B2": {
        "value": "=A2*1.1",
        "formula": "=A2*1.1",
        "type": "formula"
      }
    },
    "merged_cells": [
      {"range": "A1:C1"}
    ]
  }
}
```

## Git Integration

GitCells provides seamless Git integration for version control:

### Automatic Commits

When watching directories, GitCells can automatically commit changes:

```bash
gitcells watch --auto-commit ./data
```

### Commit Messages

Commit messages are auto-generated as `GitCells: <event> <filename>` (for watch) or `GitCells: synchronized N Excel file(s)` (for sync). Customization is intentionally not supported — pair GitCells with a Git GUI if you want to rewrite messages.

### Conflict Resolution

GitCells does not include built-in merge tools. Use your Git client (GitHub Desktop, VS Code, `git mergetool`) to resolve JSON chunk conflicts the same way you would any text file. Per-sheet chunk files keep conflicts narrow.

## Integration with Development Workflows

### CI/CD Pipelines

Use GitCells in GitHub Actions:

```yaml
name: Excel Sync
on:
  push:
    paths:
      - '**/*.xlsx'
      - '**/*.xls'

jobs:
  excel-sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Download GitCells
        run: |
          curl -L https://github.com/Classic-Homes/gitcells/releases/latest/download/gitcells-linux.tar.gz -o gitcells-linux.tar.gz
          tar -xzf gitcells-linux.tar.gz
          mv gitcells-linux gitcells
          chmod +x gitcells
      - name: Convert Excel files
        run: ./gitcells convert *.xlsx
      - name: Commit changes
        run: |
          git config --local user.email "action@github.com"
          git config --local user.name "GitHub Action"
          git add *.json
          git diff --staged --quiet || git commit -m "Auto-sync Excel files"
          git push
```

### Pre-commit Hooks

Add GitCells to your pre-commit configuration:

```yaml
repos:
  - repo: local
    hooks:
      - id: gitcells
        name: GitCells Excel Converter
        entry: gitcells convert
        language: system
        files: \.(xlsx|xls|xlsm)$
```

## Best Practices

### 1. File Organization

```
project/
├── .gitcells.yaml
├── data/
│   ├── sales.xlsx
│   ├── sales.xlsx.json      # Auto-generated
│   └── inventory.xlsx
├── reports/
│   └── monthly.xlsx
└── templates/
    └── template.xlsx
```

### 2. Git Configuration

- Add `*.json` files to your Git repository
- Consider `.gitignore` for Excel temp files:
  ```
  ~$*
  *.tmp
  .~lock.*
  ```

### 3. Team Workflows

1. **Single Source of Truth**: Keep Excel files as the primary source
2. **Review JSON Changes**: Use Git to review what actually changed
3. **Merge Conflicts**: Let GitCells handle automatic resolution
4. **Regular Syncing**: Use `gitcells sync` before major changes

### 4. Performance Tips

- Use `ignore_empty_cells: true` for large, sparse spreadsheets
- Set appropriate `max_cells_per_sheet` limits for memory management
- Use `debounce_delay` to avoid excessive processing during active editing

## Troubleshooting

### Common Issues

**1. Permission Denied**
```bash
# Solution: Ensure gitcells is executable
chmod +x gitcells
```

**2. Excel File Locked**
```
Error: failed to open Excel file: file is locked
```
```bash
# Solution: Close Excel application or wait for auto-save
# GitCells automatically ignores temp files like ~$filename.xlsx
```

**3. Large File Memory Issues**
```
Error: out of memory processing large file
```
```yaml
# Solution: GitCells automatically chunks large files
# You can adjust limits in .gitcells.yaml:
converter:
  max_cells_per_sheet: 100000
  ignore_empty_cells: true
```

**4. Git Conflicts**

Resolve JSON chunk conflicts the same way you would any text file — use your Git client (GitHub Desktop, VS Code, `git mergetool`).

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
gitcells --verbose watch .
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
git clone https://github.com/Classic-Homes/gitcells.git
cd gitcells
go mod tidy
make test
```

### Running Tests

```bash
make test                    # Run all tests
make test-short             # Skip integration tests
make test-coverage          # Generate coverage report
```

### Documentation

#### Online Documentation

Visit [gitcells.com/docs](https://gitcells.com/docs) for the complete documentation.

#### Local Documentation Server

Run the documentation locally using Docker:

```bash
# Start the documentation server
./scripts/serve-docs.sh

# Or use Docker Compose directly
docker-compose up docs

# Stop the server
./scripts/serve-docs.sh stop
```

The documentation will be available at [http://localhost:8000](http://localhost:8000).

**Available commands:**
- `./scripts/serve-docs.sh` - Start the documentation server
- `./scripts/serve-docs.sh stop` - Stop the server
- `./scripts/serve-docs.sh restart` - Restart the server
- `./scripts/serve-docs.sh logs` - View server logs

#### Building Documentation

```bash
# Build static documentation site
./scripts/build-docs.sh

# Output will be in site/ directory
```

## License

GitCells is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Architecture

GitCells is built with a modular architecture:

- **CLI Layer**: Cobra-based command interface with TUI support
- **Converter Engine**: Handles Excel ↔ JSON transformations with chunking
- **Git Integration**: Automated version control operations
- **File Watcher**: FSNotify-based monitoring with intelligent debouncing
- **Update System**: Self-updating with GitHub releases integration

## Support

- **Documentation**: [Full Documentation](https://gitcells.com/docs) | Run locally: `./scripts/serve-docs.sh`
- **Issues**: [GitHub Issues](https://github.com/Classic-Homes/gitcells/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Classic-Homes/gitcells/discussions)
- **Quick Start**: Run `gitcells init` then `gitcells watch .`

---

*Built with ❤️ for teams who need Excel files under version control*