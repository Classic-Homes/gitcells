# File Watching Guide

GitCells can automatically monitor directories for Excel file changes and convert them to JSON in real-time. This guide explains how to use the file watching feature effectively.

## Starting the File Watcher

### Basic Usage

To watch the current directory:
```bash
gitcells watch .
```

To watch specific directories:
```bash
gitcells watch /path/to/excel/files /another/path
```

To watch with custom configuration:
```bash
gitcells watch --config my-config.yaml .
```

### Watch Command Options

```bash
gitcells watch [directories...] [flags]

Flags:
  --auto-commit    Automatically commit changes to git (default: true)
  --auto-push      Automatically push commits to remote (default: false)
  --config string  Custom config file path
  --verbose        Enable verbose logging
```

## How File Watching Works

1. **Detection**: GitCells monitors specified directories for file system events
2. **Filtering**: Only Excel files matching your configuration are processed
3. **Debouncing**: Multiple rapid changes are grouped together
4. **Conversion**: Excel files are converted to JSON
5. **Git Operations**: Changes are optionally committed to Git

### The Watching Process

```
Excel File Saved → GitCells Detects → Wait (debounce) → Convert to JSON → Commit to Git
```

## Configuring File Watching

### In .gitcells.yaml

```yaml
watcher:
  # Directories to watch
  directories:
    - "."
    - "./reports"
    - "/absolute/path/to/files"
  
  # Files to ignore
  ignore_patterns:
    - "~$*"           # Excel temp files
    - "*.tmp"         # Temporary files
    - "backup/**"     # Backup folders
    - "test_*.xlsx"   # Test files
  
  # Wait time before processing
  debounce_delay: 2s
  
  # File types to watch
  file_extensions:
    - ".xlsx"
    - ".xls"
    - ".xlsm"
```

### Understanding Debounce Delay

The debounce delay prevents GitCells from processing a file multiple times when:
- Excel auto-saves frequently
- You make rapid changes
- Multiple files are saved at once

Example delays:
- `1s` - Very responsive, may process multiple times
- `2s` - Good balance (default)
- `5s` - Better for slower systems or large files
- `10s` - Good for shared network drives

## Watching Patterns

### Watch Everything

Watch all Excel files in current directory and subdirectories:

```yaml
watcher:
  directories: ["."]
  file_extensions: [".xlsx", ".xls", ".xlsm", ".xlsb"]
```

### Watch Specific Folders

Watch only specific folders:

```yaml
watcher:
  directories:
    - "./finance/reports"
    - "./hr/records"
    - "./sales/data"
```

### Exclude Patterns

Common patterns to exclude:

```yaml
watcher:
  ignore_patterns:
    - "~$*"              # Excel temporary files
    - "*.tmp"            # Temp files
    - ".~lock.*"         # LibreOffice locks
    - "**/archive/**"    # Archive folders
    - "**/backup/**"     # Backup folders
    - "Book1.xlsx"       # Default Excel names
    - "Copy of *"        # Copies
    - "test_*"           # Test files
    - "draft_*"          # Draft files
```

## Monitoring the Watcher

### Console Output

When running, GitCells shows:
```
INFO[2024-01-15 10:30:45] Watching directory: ./spreadsheets
INFO[2024-01-15 10:30:45] Watching for changes... Press Ctrl+C to stop
INFO[2024-01-15 10:31:12] Processing modified: Budget2024.xlsx
INFO[2024-01-15 10:31:13] GitCells: modified Budget2024.xlsx at 2024-01-15 10:31:12
```

### Verbose Mode

For detailed information:
```bash
gitcells watch --verbose .
```

This shows:
- All file system events (including ignored files)
- Conversion details
- Git operations
- Performance metrics

### Using the TUI

The Terminal UI provides a visual way to monitor:
```bash
gitcells tui
```

Then select "Status Dashboard" to see:
- Active watchers
- Recent file changes
- Conversion status
- Error logs

## Best Practices

### 1. Directory Organization

Organize your Excel files for efficient watching:
```
project/
├── .gitcells.yaml
├── reports/          # Watch this
│   ├── monthly/
│   └── yearly/
├── templates/        # Watch this
└── archive/          # Don't watch this
```

### 2. Ignore Patterns

Always ignore temporary files:
```yaml
ignore_patterns:
  - "~$*"
  - "*.tmp"
  - ".~lock.*"
```

### 3. Performance Optimization

For better performance:
- Watch specific directories instead of entire drives
- Use appropriate debounce delays
- Exclude large archive folders
- Limit the number of watched directories

### 4. Network Drives

When watching network drives:
```yaml
watcher:
  debounce_delay: 10s  # Longer delay for network latency
  directories:
    - "//server/shared/excel"
```

## Troubleshooting

### Watcher Not Detecting Changes

1. **Check file extensions**:
   ```yaml
   file_extensions: [".xlsx", ".xls", ".xlsm"]
   ```

2. **Verify ignore patterns aren't too broad**:
   ```bash
   gitcells watch --verbose .  # See what's being ignored
   ```

3. **Ensure directory permissions**:
   ```bash
   ls -la /path/to/watch  # Check read permissions
   ```

### High CPU Usage

1. **Increase debounce delay**:
   ```yaml
   debounce_delay: 5s
   ```

2. **Watch fewer directories**:
   ```yaml
   directories: ["./active"]  # Not ["."]
   ```

3. **Exclude large folders**:
   ```yaml
   ignore_patterns: ["**/archive/**", "**/backup/**"]
   ```

### Files Processing Multiple Times

Increase the debounce delay:
```yaml
debounce_delay: 10s
```

### Memory Issues with Large Files

Configure chunking:
```yaml
converter:
  chunking_strategy: "size-based"
  max_chunk_size: "5MB"
```

## Advanced Watching

### Multiple Configurations

Run multiple watchers with different configs:

```bash
# Terminal 1: Watch reports with auto-commit
gitcells watch --config reports.yaml ./reports

# Terminal 2: Watch data without auto-commit
gitcells watch --config data.yaml --auto-commit=false ./data
```

### Scripted Watching

Create a watch script (`watch.sh`):
```bash
#!/bin/bash
echo "Starting GitCells watchers..."

# Watch different directories with different configs
gitcells watch --config finance.yaml ./finance &
gitcells watch --config hr.yaml ./hr &
gitcells watch --config sales.yaml ./sales &

# Wait for Ctrl+C
wait
```

### Systemd Service (Linux)

Create a systemd service for automatic watching:

```ini
[Unit]
Description=GitCells File Watcher
After=network.target

[Service]
Type=simple
User=youruser
WorkingDirectory=/path/to/excel/files
ExecStart=/usr/local/bin/gitcells watch .
Restart=always

[Install]
WantedBy=multi-user.target
```

## Security Considerations

### Sensitive Data

When watching sensitive files:
1. Use `.gitignore` to exclude sensitive data from Git
2. Configure ignore patterns for confidential files
3. Use local Git repositories only
4. Enable encryption for Git repositories

### Access Control

- GitCells respects file system permissions
- Only watches files the user can read
- Commits use configured Git credentials
- No data is sent to external services

## Next Steps

- Learn about [Converting Files](converting.md) manually
- Set up [Git Integration](git-integration.md) for version control
- Explore the [Terminal UI](tui.md) for visual monitoring
- Check [Troubleshooting](troubleshooting.md) for common issues