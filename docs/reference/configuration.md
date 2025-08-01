# Configuration Reference

This reference documents all configuration options available in GitCells. Configuration is stored in YAML format in `.gitcells.yaml`.

## Configuration File

### File Locations

GitCells searches for configuration in this order:
1. Path specified by `--config` flag
2. `.gitcells.yaml` in current directory
3. `.gitcells.yaml` in parent directories (up to Git root)
4. `~/.gitcells.yaml` (user home directory)
5. `/etc/gitcells/config.yaml` (system-wide)

### File Structure

```yaml
version: 1.0  # Configuration version (required)

git:          # Git integration settings
watcher:      # File watching settings
converter:    # Conversion settings
advanced:     # Advanced settings
```

## Complete Configuration Options

### Root Level

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `version` | string | required | Configuration version (must be "1.0") |

### git

Git integration settings.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `branch` | string | `"main"` | Git branch to use for commits |
| `auto_push` | boolean | `false` | Automatically push after commits |
| `auto_pull` | boolean | `true` | Pull before operations |
| `user_name` | string | `"GitCells"` | Git user name for commits |
| `user_email` | string | `"gitcells@localhost"` | Git user email for commits |
| `commit_template` | string | `"GitCells: {action} {filename} at {timestamp}"` | Commit message template |
| `co_authors` | []string | `[]` | Co-authors to add to commits |
| `gpg_sign` | boolean | `false` | Sign commits with GPG |
| `remote` | string | `"origin"` | Remote name for push/pull |

#### Commit Template Variables

- `{action}` - Action performed (created, modified, deleted)
- `{filename}` - Name of the Excel file
- `{timestamp}` - ISO 8601 timestamp
- `{user}` - System username
- `{hostname}` - Machine hostname
- `{branch}` - Current Git branch

### watcher

File system watching configuration.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `directories` | []string | `["."]` | Directories to watch |
| `ignore_patterns` | []string | `["~$*", "*.tmp", ".~lock.*"]` | Glob patterns to ignore |
| `debounce_delay` | duration | `"2s"` | Delay before processing changes |
| `file_extensions` | []string | `[".xlsx", ".xls", ".xlsm"]` | File extensions to watch |
| `recursive` | boolean | `true` | Watch directories recursively |
| `follow_symlinks` | boolean | `false` | Follow symbolic links |
| `max_depth` | integer | `10` | Maximum directory depth |
| `poll_interval` | duration | `"0s"` | Polling interval (0 = use native events) |

#### Duration Format

Durations use Go duration format:
- `"300ms"` - 300 milliseconds
- `"1.5s"` - 1.5 seconds
- `"2m"` - 2 minutes
- `"1h30m"` - 1 hour 30 minutes

### converter

Excel/JSON conversion settings.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `preserve_formulas` | boolean | `true` | Preserve Excel formulas |
| `preserve_styles` | boolean | `true` | Preserve cell styles |
| `preserve_comments` | boolean | `true` | Preserve cell comments |
| `preserve_charts` | boolean | `true` | Preserve charts |
| `preserve_pivot_tables` | boolean | `true` | Preserve pivot tables |
| `preserve_images` | boolean | `true` | Preserve embedded images |
| `preserve_macros` | boolean | `false` | Preserve VBA macros |
| `compact_json` | boolean | `false` | Output compact JSON |
| `ignore_empty_cells` | boolean | `true` | Skip empty cells in output |
| `ignore_hidden_sheets` | boolean | `false` | Skip hidden sheets |
| `max_cells_per_sheet` | integer | `1000000` | Maximum cells per sheet |
| `chunking_strategy` | string | `"sheet-based"` | Strategy for large files |
| `max_chunk_size` | string | `"10MB"` | Maximum chunk size |
| `number_precision` | integer | `15` | Decimal precision for numbers |
| `date_format` | string | `"2006-01-02T15:04:05Z07:00"` | Date format (Go format) |
| `formula_r1c1` | boolean | `true` | Include R1C1 formula notation |
| `include_metadata` | boolean | `true` | Include file metadata |

#### Chunking Strategies

- `"sheet-based"` - One chunk per sheet
- `"row-based"` - Split by row count
- `"size-based"` - Split by file size
- `"disabled"` - No chunking

### advanced

Advanced settings for performance and debugging.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `worker_threads` | integer | `4` | Number of worker threads |
| `enable_cache` | boolean | `true` | Enable conversion cache |
| `cache_dir` | string | `".gitcells.cache"` | Cache directory location |
| `cache_ttl` | duration | `"24h"` | Cache time-to-live |
| `log_level` | string | `"info"` | Log level |
| `log_file` | string | `""` | Log file path (empty = stdout) |
| `log_format` | string | `"text"` | Log format (text, json) |
| `enable_profiling` | boolean | `false` | Enable performance profiling |
| `profile_dir` | string | `".gitcells.profile"` | Profile output directory |
| `memory_limit` | string | `"0"` | Memory limit (0 = unlimited) |
| `temp_dir` | string | system temp | Temporary file directory |
| `parallel_conversions` | integer | `2` | Max parallel conversions |
| `http_timeout` | duration | `"30s"` | HTTP timeout for updates |
| `disable_telemetry` | boolean | `false` | Disable anonymous usage stats |

#### Log Levels

- `"debug"` - Detailed debugging information
- `"info"` - General information
- `"warn"` - Warnings only
- `"error"` - Errors only

## Configuration Examples

### Minimal Configuration

```yaml
version: 1.0

watcher:
  directories: ["."]
```

### Development Configuration

```yaml
version: 1.0

git:
  branch: develop
  commit_template: "[{branch}] {user}: {action} {filename}"

watcher:
  directories: ["./src", "./tests"]
  ignore_patterns: ["~$*", "*.tmp", "test_*", "debug_*"]
  debounce_delay: 1s

converter:
  compact_json: true
  ignore_empty_cells: true

advanced:
  log_level: debug
  enable_profiling: true
```

### Production Configuration

```yaml
version: 1.0

git:
  auto_push: true
  auto_pull: true
  gpg_sign: true
  commit_template: "Excel Update: {filename} by {user}"

watcher:
  directories: ["/shared/excel/files"]
  debounce_delay: 10s
  follow_symlinks: true

converter:
  preserve_formulas: true
  preserve_styles: true
  preserve_comments: true
  preserve_charts: true
  preserve_pivot_tables: true
  chunking_strategy: size-based
  max_chunk_size: 25MB

advanced:
  worker_threads: 8
  enable_cache: true
  cache_ttl: 48h
  log_level: warn
  log_file: /var/log/gitcells/gitcells.log
  memory_limit: 2GB
```

### Network Drive Configuration

```yaml
version: 1.0

watcher:
  directories: ["//fileserver/shared/excel"]
  debounce_delay: 15s
  poll_interval: 5s  # Use polling for network drives

converter:
  chunking_strategy: size-based
  max_chunk_size: 5MB

advanced:
  worker_threads: 2
  parallel_conversions: 1
  http_timeout: 60s
```

## Environment Variable Overrides

Configuration values can be overridden using environment variables:

```bash
# Override pattern: GITCELLS_<SECTION>_<KEY>
export GITCELLS_GIT_AUTO_PUSH=true
export GITCELLS_WATCHER_DEBOUNCE_DELAY=5s
export GITCELLS_CONVERTER_COMPACT_JSON=true
export GITCELLS_ADVANCED_LOG_LEVEL=debug
```

## Validation

Validate configuration file:

```bash
gitcells init --validate
```

Common validation errors:
- Unknown configuration keys
- Invalid data types
- Invalid duration formats
- Mutually exclusive options

## Migration

### From v0.x to v1.0

```yaml
# Old format (v0.x)
watch_dirs: [".", "./reports"]
auto_commit: true

# New format (v1.0)
version: 1.0
watcher:
  directories: [".", "./reports"]
git:
  auto_commit: true
```

## Best Practices

1. **Version Control**: Always include `.gitcells.yaml` in Git
2. **Environment-Specific**: Use separate configs for dev/prod
3. **Documentation**: Comment complex configurations
4. **Security**: Don't store sensitive data in config
5. **Performance**: Tune based on file sizes and system resources

## Next Steps

- See [Command Reference](commands.md) for using configuration
- Review [User Guide](../user-guide/configuration.md) for examples
- Check [JSON Format](json-format.md) for output configuration
- Read [Troubleshooting](../user-guide/troubleshooting.md) for issues