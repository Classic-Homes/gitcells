# Command Reference

Complete reference for all GitCells commands.

## Global Options

Options available for all commands:

```bash
--config PATH      # Use custom config file
--verbose, -v      # Verbose output
--quiet, -q        # Suppress output
--json             # JSON output format
--no-color         # Disable colored output
--help, -h         # Show help
--version          # Show version
```

## Core Commands

### gitcells init

Initialize GitCells in a directory.

```bash
gitcells init [options]
```

**Options:**
- `--force`: Overwrite existing configuration
- `--team`: Use team-oriented defaults
- `--minimal`: Minimal configuration
- `--wizard`: Interactive setup wizard

**Examples:**
```bash
gitcells init
gitcells init --team
gitcells init --wizard
```

### gitcells convert

Convert between Excel and JSON formats.

```bash
gitcells convert <file|pattern> [options]
```

**Options:**
- `--output PATH`: Output location
- `--pretty`: Pretty-print JSON
- `--force`: Overwrite existing files
- `--recursive`: Process subdirectories
- `--parallel N`: Use N parallel workers
- `--sheets LIST`: Convert specific sheets only
- `--skip-hidden`: Skip hidden sheets
- `--data-only`: Skip formatting information
- `--compress`: Compress JSON output

**Examples:**
```bash
gitcells convert report.xlsx
gitcells convert *.xlsx --pretty
gitcells convert data/ --recursive --parallel 4
gitcells convert report.xlsx --sheets "Summary,Data"
```

### gitcells sync

Synchronize Excel and JSON file pairs.

```bash
gitcells sync [file|pattern] [options]
```

**Options:**
- `--all`: Sync all tracked files
- `--direction [excel-to-json|json-to-excel|auto]`: Sync direction
- `--force`: Force sync even if up-to-date
- `--dry-run`: Show what would be done

**Examples:**
```bash
gitcells sync
gitcells sync report.xlsx
gitcells sync --all
gitcells sync --direction json-to-excel
```

### gitcells status

Show status of Excel files.

```bash
gitcells status [file|pattern] [options]
```

**Options:**
- `--detailed`: Show detailed information
- `--format [table|json|csv]`: Output format
- `--filter [modified|synced|conflicted|untracked]`: Filter by status
- `--since TIME`: Show changes since time

**Examples:**
```bash
gitcells status
gitcells status --detailed
gitcells status --filter modified
gitcells status --since yesterday
```

### gitcells diff

Show differences between versions.

```bash
gitcells diff <file> [commit] [options]
```

**Options:**
- `--format [summary|detailed|json]`: Output format
- `--sheets LIST`: Diff specific sheets only
- `--filter [values|formulas|formatting]`: Filter changes
- `--threshold N`: Only show changes above threshold
- `--export PATH`: Export diff to file
- `--visual`: Open visual diff tool

**Examples:**
```bash
gitcells diff report.xlsx
gitcells diff report.xlsx HEAD~1
gitcells diff report.xlsx --filter formulas
gitcells diff report.xlsx abc123..def456 --visual
```

### gitcells watch

Watch for file changes and auto-sync.

```bash
gitcells watch [directory] [options]
```

**Options:**
- `--patterns LIST`: File patterns to watch
- `--ignore LIST`: Patterns to ignore
- `--debounce N`: Debounce delay in seconds
- `--daemon`: Run as background daemon
- `--stop`: Stop running watch
- `--status`: Show watch status
- `--log`: Show live log

**Examples:**
```bash
gitcells watch
gitcells watch reports/
gitcells watch --patterns "*.xlsx,*.xlsm"
gitcells watch --daemon
gitcells watch --stop
```

## Conflict Commands

### gitcells conflict

Show and manage conflicts.

```bash
gitcells conflict <file> [options]
```

**Options:**
- `--check`: Check for conflicts only
- `--show`: Show conflict details
- `--visual`: Open visual conflict viewer
- `--export PATH`: Export conflicts to file

**Examples:**
```bash
gitcells conflict budget.xlsx
gitcells conflict --check
gitcells conflict budget.xlsx --visual
```

### gitcells resolve

Resolve conflicts in files.

```bash
gitcells resolve <file> [options]
```

**Options:**
- `--interactive`: Interactive resolution
- `--strategy [ours|theirs|base|newer|max|min]`: Resolution strategy
- `--cell CELL`: Resolve specific cell
- `--sheet SHEET`: Resolve specific sheet
- `--auto`: Use configured auto-resolution rules

**Examples:**
```bash
gitcells resolve budget.xlsx --interactive
gitcells resolve budget.xlsx --strategy theirs
gitcells resolve budget.xlsx --cell B5 --use ours
```

### gitcells merge

Three-way merge for Excel files.

```bash
gitcells merge <base> <ours> <theirs> [options]
```

**Options:**
- `--output PATH`: Output file path
- `--strategy NAME`: Merge strategy
- `--visual`: Use visual merge tool

**Examples:**
```bash
gitcells merge base.xlsx mine.xlsx theirs.xlsx
gitcells merge base.xlsx mine.xlsx theirs.xlsx --output merged.xlsx
```

## Advanced Commands

### gitcells validate

Validate Excel files and conversions.

```bash
gitcells validate <file> [options]
```

**Options:**
- `--check-formulas`: Validate all formulas
- `--check-references`: Check external references
- `--check-circular`: Detect circular references
- `--round-trip`: Test Excel→JSON→Excel conversion

**Examples:**
```bash
gitcells validate report.xlsx
gitcells validate *.xlsx --check-formulas
gitcells validate report.xlsx --round-trip
```

### gitcells history

Show file history.

```bash
gitcells history <file> [options]
```

**Options:**
- `--limit N`: Limit number of entries
- `--since DATE`: Show history since date
- `--until DATE`: Show history until date
- `--format [table|json|timeline]`: Output format
- `--changes`: Include change details

**Examples:**
```bash
gitcells history budget.xlsx
gitcells history budget.xlsx --limit 10
gitcells history budget.xlsx --since 2024-01-01 --changes
```

### gitcells blame

Show who changed specific cells.

```bash
gitcells blame <file> <cell> [options]
```

**Options:**
- `--sheet SHEET`: Specify sheet
- `--range RANGE`: Blame cell range
- `--format [table|json]`: Output format

**Examples:**
```bash
gitcells blame budget.xlsx B5
gitcells blame budget.xlsx A1:D10 --sheet Summary
```

### gitcells lock

Lock files or ranges.

```bash
gitcells lock <file> [options]
```

**Options:**
- `--sheet SHEET`: Lock specific sheet
- `--range RANGE`: Lock cell range
- `--message MSG`: Lock message
- `--force`: Override existing lock

**Examples:**
```bash
gitcells lock budget.xlsx
gitcells lock budget.xlsx --sheet Revenue
gitcells lock budget.xlsx --range A1:D10 --message "Updating formulas"
```

### gitcells unlock

Unlock files or ranges.

```bash
gitcells unlock <file> [options]
```

**Options:**
- `--force`: Force unlock (admin only)
- `--all`: Unlock all locks

**Examples:**
```bash
gitcells unlock budget.xlsx
gitcells unlock budget.xlsx --force
gitcells unlock --all
```

## Utility Commands

### gitcells config

Manage configuration.

```bash
gitcells config [options]
```

**Options:**
- `--get KEY`: Get configuration value
- `--set KEY VALUE`: Set configuration value
- `--list`: List all configuration
- `--edit`: Open config in editor
- `--validate`: Validate configuration

**Examples:**
```bash
gitcells config --list
gitcells config --get watch.debounce
gitcells config --set watch.debounce 5s
gitcells config --edit
```

### gitcells cache

Manage cache.

```bash
gitcells cache [command] [options]
```

**Commands:**
- `clear`: Clear all cache
- `size`: Show cache size
- `prune`: Remove old entries

**Examples:**
```bash
gitcells cache clear
gitcells cache size
gitcells cache prune --older-than 7d
```

### gitcells export

Export Excel data.

```bash
gitcells export <file> [options]
```

**Options:**
- `--format [csv|pdf|html|markdown]`: Export format
- `--sheets LIST`: Export specific sheets
- `--output PATH`: Output location
- `--readonly`: Create read-only copy

**Examples:**
```bash
gitcells export report.xlsx --format pdf
gitcells export report.xlsx --format csv --sheets "Data"
gitcells export *.xlsx --readonly --output exports/
```

### gitcells import

Import data into Excel.

```bash
gitcells import <source> <target> [options]
```

**Options:**
- `--format [csv|json|xml]`: Source format
- `--sheet SHEET`: Target sheet
- `--append`: Append to existing data
- `--create`: Create file if not exists

**Examples:**
```bash
gitcells import data.csv report.xlsx
gitcells import data.json report.xlsx --sheet "Imported"
gitcells import api-response.json report.xlsx --create
```

## Diagnostic Commands

### gitcells doctor

Check system and configuration.

```bash
gitcells doctor [options]
```

**Options:**
- `--fix`: Attempt to fix issues
- `--verbose`: Detailed diagnosis

**Examples:**
```bash
gitcells doctor
gitcells doctor --fix
gitcells doctor --verbose
```

### gitcells debug

Debug information and logs.

```bash
gitcells debug [options]
```

**Options:**
- `--logs`: Show recent logs
- `--config`: Show active configuration
- `--env`: Show environment info
- `--trace`: Enable trace logging

**Examples:**
```bash
gitcells debug --logs
gitcells debug --config
gitcells debug --trace convert report.xlsx
```

## Exit Codes

- `0`: Success
- `1`: General error
- `2`: Command syntax error
- `3`: File not found
- `4`: Permission denied
- `5`: Conflict detected
- `6`: Validation failed
- `7`: Lock conflict
- `8`: Network error

## Environment Variables

- `GITCELLS_CONFIG`: Path to config file
- `GITCELLS_HOME`: GitCells home directory
- `GITCELLS_LOG_LEVEL`: Log level (debug|info|warn|error)
- `GITCELLS_NO_COLOR`: Disable colors
- `GITCELLS_EDITOR`: Editor for config editing

## Configuration File

Default location: `.gitcells.yml`

See [Configuration Reference](configuration.md) for details.

## Next Steps

- Learn about [configuration options](configuration.md)
- Read [troubleshooting guide](troubleshooting.md)
- Explore [advanced workflows](../guides/collaboration.md)