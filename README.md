# GitCells

GitCells seamlessly bridges Excel and Git, enabling true version control for spreadsheets. It automatically converts Excel files to human-readable JSON for diffing and merging, then restores them to native Excel format for editing. No more binary conflicts, lost formulas, or overwritten work.

## Features

- **Excel ↔ JSON Conversion**: Converts Excel files to structured JSON that preserves formulas, styles, comments, and merged cells
- **Version Control**: Full Git integration with automatic commits and conflict resolution
- **File Watching**: Automatically monitors directories for Excel file changes
- **Diff Generation**: Human-readable diffs showing exactly what changed between Excel versions
- **Cross-Platform**: Works on Windows, macOS, and Linux
- **Formula Preservation**: Maintains Excel formulas and calculations during conversion
- **Smart Conflict Resolution**: Intelligent merging strategies for concurrent edits

## Installation

### Pre-built Binaries

Download the latest release for your platform from [GitHub Releases](https://github.com/Classic-Homes/gitcells/releases).

### From Source

```bash
go install github.com/Classic-Homes/gitcells/cmd/gitcells@latest
```

### Build from Repository

```bash
git clone https://github.com/Classic-Homes/gitcells.git
cd gitcells
make build

# Install locally
make install
```

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

## Configuration

GitCells uses a `.gitcells.yaml` configuration file. Here's a comprehensive example:

```yaml
version: 1.0

git:
  branch: main
  auto_push: false
  auto_pull: true
  user_name: "GitCells"
  user_email: "gitcells@yourcompany.com"
  commit_template: "GitCells: {action} {filename} at {timestamp}"

watcher:
  directories:
    - "./data"
    - "./reports"
  ignore_patterns:
    - "~$*"           # Excel temp files
    - "*.tmp"         # Temporary files
    - ".~lock.*"      # LibreOffice lock files
  debounce_delay: 2s  # Wait 2 seconds before processing changes
  file_extensions:
    - ".xlsx"
    - ".xls"
    - ".xlsm"

converter:
  preserve_formulas: true     # Keep Excel formulas
  preserve_styles: true      # Keep cell formatting
  preserve_comments: true    # Keep cell comments
  compact_json: false        # Use pretty-printed JSON
  ignore_empty_cells: true   # Skip empty cells in JSON
  max_cells_per_sheet: 1000000  # Memory protection
```

## Command Reference

### Global Options

- `--config`: Specify config file path (default: `.gitcells.yaml`)
- `--verbose`: Enable verbose logging
- `--help`: Show help information

### Commands

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

### ⚠️ Limitations

- **Charts and Graphs**: Not preserved in JSON format
- **Pivot Tables**: Structure preserved, but may need refresh
- **Macros**: VBA macros are not converted
- **External Links**: Links to other files may break
- **Images**: Embedded images are not preserved

## JSON Format

GitCells converts Excel files to a structured JSON format that preserves all sheet data:

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
    {
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
  ]
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

Customize commit messages using templates in your config:

```yaml
git:
  commit_template: "GitCells: {action} {filename} at {timestamp}"
```

Available variables:
- `{action}`: Type of change (create, modify, delete)
- `{filename}`: Name of the changed file
- `{timestamp}`: ISO timestamp of the change
- `{branch}`: Current Git branch

### Conflict Resolution

GitCells includes intelligent conflict resolution for concurrent edits:

1. **Smart Merge**: Attempts to merge non-conflicting changes automatically
2. **Timestamp Resolution**: Uses the most recent version when smart merge fails
3. **Manual Resolution**: Provides clear conflict markers for manual resolution

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
          curl -L https://github.com/Classic-Homes/gitcells/releases/latest/download/gitcells-linux-amd64 -o gitcells
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
# Solution: Adjust limits in .gitcells.yaml
converter:
  max_cells_per_sheet: 100000
  ignore_empty_cells: true
```

**4. Git Conflicts**
```bash
# Solution: Use GitCells's conflict resolution
gitcells sync --resolve-conflicts
```

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
gitcells --verbose watch .
```

### Log Files

GitCells logs are written to:
- Linux/macOS: `~/.local/share/gitcells/logs/`
- Windows: `%APPDATA%/gitcells/logs/`

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

## License

GitCells is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Support

- **Documentation**: [GitHub Wiki](https://github.com/Classic-Homes/gitcells/wiki)
- **Issues**: [GitHub Issues](https://github.com/Classic-Homes/gitcells/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Classic-Homes/gitcells/discussions)

---

*Built with ❤️ for teams who need Excel files under version control*