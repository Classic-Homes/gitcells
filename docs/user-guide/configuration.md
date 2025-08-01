# Configuration Guide

GitCells uses a YAML configuration file (`.gitcells.yaml`) to control its behavior. This guide explains all configuration options and how to use them.

## Configuration File Location

GitCells looks for `.gitcells.yaml` in the following order:
1. Current directory
2. Parent directories (up to the Git repository root)
3. Home directory (`~/.gitcells.yaml`)

You can also specify a config file using the `--config` flag:
```bash
gitcells watch --config /path/to/config.yaml .
```

## Basic Configuration

Here's a simple configuration to get started:

```yaml
version: 1.0

watcher:
  directories: 
    - "."
  file_extensions:
    - ".xlsx"
    - ".xls"
```

This tells GitCells to watch the current directory for Excel files.

## Complete Configuration Reference

Here's a fully configured `.gitcells.yaml` with all available options:

```yaml
version: 1.0

# Git integration settings
git:
  # Branch to use for commits (default: main)
  branch: main
  
  # Automatically push commits to remote (default: false)
  auto_push: false
  
  # Automatically pull before operations (default: true)
  auto_pull: true
  
  # Git user name for commits
  user_name: "GitCells"
  
  # Git user email for commits
  user_email: "gitcells@localhost"
  
  # Commit message template
  # Available variables: {action}, {filename}, {timestamp}
  commit_template: "GitCells: {action} {filename} at {timestamp}"

# File watcher settings
watcher:
  # Directories to watch (can be relative or absolute paths)
  directories:
    - "."
    - "./spreadsheets"
    - "/absolute/path/to/excel/files"
  
  # Patterns to ignore (glob patterns)
  ignore_patterns:
    - "~$*"              # Excel temporary files
    - "*.tmp"            # Temporary files
    - ".~lock.*"         # LibreOffice lock files
    - "**/archive/**"    # Ignore archive folders
    - "test_*.xlsx"      # Ignore test files
  
  # Delay before processing changes (prevents multiple triggers)
  debounce_delay: 2s
  
  # File extensions to watch
  file_extensions:
    - ".xlsx"
    - ".xls"
    - ".xlsm"
    - ".xlsb"

# Converter settings
converter:
  # Preserve Excel formulas in JSON (default: true)
  preserve_formulas: true
  
  # Preserve cell styles and formatting (default: true)
  preserve_styles: true
  
  # Preserve cell comments (default: true)
  preserve_comments: true
  
  # Output compact JSON (smaller files, less readable)
  compact_json: false
  
  # Skip empty cells in JSON output
  ignore_empty_cells: true
  
  # Maximum cells per sheet (prevents memory issues)
  max_cells_per_sheet: 1000000
  
  # Preserve charts (default: true)
  preserve_charts: true
  
  # Preserve pivot tables (default: true)
  preserve_pivot_tables: true
  
  # Chunking strategy for large files
  # Options: "sheet-based", "row-based", "size-based"
  chunking_strategy: "sheet-based"
  
  # Maximum chunk size (for size-based chunking)
  max_chunk_size: "10MB"

# Advanced settings
advanced:
  # Number of worker threads for processing
  worker_threads: 4
  
  # Cache converted files for performance
  enable_cache: true
  
  # Cache directory location
  cache_dir: ".gitcells.cache"
  
  # Log level (debug, info, warn, error)
  log_level: "info"
  
  # Enable performance profiling
  enable_profiling: false
```

## Common Configuration Scenarios

### Basic Personal Use

For tracking your personal Excel files:

```yaml
version: 1.0

watcher:
  directories: ["."]
  debounce_delay: 5s

converter:
  compact_json: true
  ignore_empty_cells: true
```

### Team Collaboration

For teams working with shared Excel files:

```yaml
version: 1.0

git:
  auto_push: true
  auto_pull: true
  commit_template: "{user}: Updated {filename}"

watcher:
  directories: ["./shared"]
  debounce_delay: 10s

converter:
  preserve_formulas: true
  preserve_styles: true
  preserve_comments: true
```

### Large Files

For working with large Excel files:

```yaml
version: 1.0

converter:
  ignore_empty_cells: true
  max_cells_per_sheet: 500000
  chunking_strategy: "size-based"
  max_chunk_size: "5MB"

advanced:
  worker_threads: 8
  enable_cache: true
```

### Development Environment

For development and testing:

```yaml
version: 1.0

watcher:
  ignore_patterns:
    - "~$*"
    - "*.tmp"
    - "test/**"
    - "temp/**"

advanced:
  log_level: "debug"
  enable_profiling: true
```

## Environment Variables

You can override configuration using environment variables:

```bash
# Override log level
export GITCELLS_LOG_LEVEL=debug

# Override auto-push setting
export GITCELLS_GIT_AUTO_PUSH=true

# Override worker threads
export GITCELLS_WORKER_THREADS=8
```

## Configuration Tips

### Performance Optimization

1. **For many small files**: Increase `worker_threads`
2. **For large files**: Use `chunking_strategy: "size-based"`
3. **For slow systems**: Increase `debounce_delay`
4. **For fast processing**: Enable `cache_dir`

### Storage Optimization

1. **Reduce file size**: Set `compact_json: true`
2. **Skip empty cells**: Set `ignore_empty_cells: true`
3. **Limit cell count**: Reduce `max_cells_per_sheet`
4. **Selective preservation**: Disable unnecessary features (styles, comments)

### Reliability

1. **Prevent conflicts**: Increase `debounce_delay` for shared files
2. **Ensure updates**: Set `auto_pull: true` for team environments
3. **Track everything**: Keep all `preserve_*` options enabled
4. **Monitor issues**: Set `log_level: "info"` or `"debug"`

## Validating Configuration

To check if your configuration is valid:

```bash
gitcells init --validate
```

This will report any errors in your `.gitcells.yaml` file.

## Examples by Use Case

### Financial Reports

```yaml
version: 1.0

watcher:
  directories: ["./reports"]
  ignore_patterns: ["draft_*", "temp_*"]

converter:
  preserve_formulas: true
  preserve_styles: true
  max_cells_per_sheet: 2000000
```

### Data Analysis

```yaml
version: 1.0

converter:
  preserve_formulas: true
  preserve_pivot_tables: true
  preserve_charts: true
  ignore_empty_cells: false  # Keep structure intact
```

### Shared Templates

```yaml
version: 1.0

git:
  auto_push: true
  commit_template: "Template update: {filename}"

watcher:
  directories: ["./templates"]
  file_extensions: [".xlsx", ".xltx"]
```

## Troubleshooting Configuration

### Common Issues

1. **GitCells not detecting changes**
   - Check `file_extensions` includes your file type
   - Verify `ignore_patterns` isn't excluding your files
   - Increase `debounce_delay` if changes happen too quickly

2. **Performance problems**
   - Reduce `max_cells_per_sheet`
   - Enable `compact_json`
   - Increase `worker_threads`

3. **Git conflicts**
   - Enable `auto_pull`
   - Increase `debounce_delay`
   - Use unique `commit_template` with `{user}`

## Next Steps

- Learn about [File Watching](watching.md) to monitor Excel files
- Explore [Git Integration](git-integration.md) for version control
- Use the [Terminal UI](tui.md) for easier configuration