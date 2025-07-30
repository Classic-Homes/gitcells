# Configuration Reference

Complete reference for GitCells configuration options.

## Configuration File

GitCells uses a YAML configuration file named `.gitcells.yml` in your project root.

### File Locations (in order of precedence)

1. Command line: `--config /path/to/config.yml`
2. Current directory: `./.gitcells.yml`
3. Parent directories: `../.gitcells.yml` (recursively)
4. User home: `~/.gitcells/config.yml`
5. System: `/etc/gitcells/config.yml`

## Core Configuration

### Basic Settings

```yaml
# .gitcells.yml
version: "1.0"

# Project settings
project:
  name: "My Excel Project"
  description: "Financial reports and analysis"
  
# File handling
files:
  patterns:
    - "*.xlsx"
    - "*.xlsm"
    - "*.xlsb"
  
  ignore:
    - "~$*"        # Excel temp files
    - "*.tmp"
    - ".~lock.*"   # LibreOffice locks
    - "backup/*"
    - "archive/*"
  
  max_size: "50MB"
  encoding: "utf-8"
```

### Conversion Settings

```yaml
conversion:
  # JSON output options
  json:
    pretty: true
    indent: 2
    compress: false
    include_metadata: true
    include_formatting: true
    include_empty_cells: false
  
  # Performance options
  performance:
    parallel: true
    workers: 4
    cache: true
    cache_ttl: "24h"
    streaming: auto  # auto|true|false
  
  # Sheet handling
  sheets:
    include_hidden: false
    include_very_hidden: false
    preserve_order: true
    
  # Cell handling
  cells:
    preserve_formulas: true
    evaluate_formulas: false
    include_comments: true
    include_validation: true
```

## Watch Configuration

### Basic Watch Settings

```yaml
watch:
  enabled: true
  
  # Directories to watch
  directories:
    - "."
    - "reports/"
    - "data/"
  
  # Watch behavior
  recursive: true
  follow_symlinks: false
  polling: false  # Use polling instead of native events
  interval: "1s"  # Polling interval
  
  # Debouncing
  debounce: "2s"
  batch_changes: true
  batch_timeout: "5m"
```

### Auto-commit Configuration

```yaml
auto_commit:
  enabled: true
  
  # Commit strategy
  strategy: "immediate"  # immediate|batch|scheduled
  
  # Commit message template
  message:
    template: "{action} {filename} - {summary}"
    
    # Available variables:
    # {filename} - Base filename
    # {filepath} - Full file path
    # {action} - create|update|delete
    # {timestamp} - ISO timestamp
    # {user} - System username
    # {hostname} - Machine name
    # {summary} - Change summary
    # {sheets} - Changed sheet names
    # {cells} - Number of cells changed
    
    prefix: "[GitCells]"
    suffix: ""
    max_length: 72
    
  # Commit options
  sign_commits: false
  gpg_key: ""
  author: "{user} <{user}@{hostname}>"
```

## Sync Configuration

```yaml
sync:
  # Sync direction
  direction: "auto"  # auto|excel-to-json|json-to-excel
  
  # Conflict handling
  conflict_resolution:
    strategy: "prompt"  # prompt|ours|theirs|newer|merge
    
    # Automatic resolution rules
    rules:
      - pattern: "*revenue*"
        strategy: "max"
      - pattern: "*cost*"
        strategy: "min"
      - sheet: "Summary"
        strategy: "theirs"
      - range: "A1:A10"
        strategy: "ours"
  
  # Validation
  validation:
    before_sync: true
    after_sync: true
    fail_on_warning: false
    fail_on_error: true
```

## Team Configuration

```yaml
team:
  # Locking
  locking:
    enabled: true
    timeout: "30m"
    allow_break: ["admin", "manager"]
    
  # Permissions
  permissions:
    default: "read"
    
    roles:
      admin:
        - all: "*"
      
      editor:
        - edit: ["reports/*", "data/*"]
        - create: true
        - delete: false
      
      viewer:
        - read: "*"
        - edit: []
  
  # Notifications
  notifications:
    email:
      enabled: true
      smtp:
        host: "smtp.company.com"
        port: 587
        username: "gitcells@company.com"
        password: "${SMTP_PASSWORD}"
      
      recipients:
        changes: ["team@company.com"]
        conflicts: ["admin@company.com"]
        errors: ["support@company.com"]
    
    slack:
      enabled: true
      webhook: "${SLACK_WEBHOOK}"
      channels:
        changes: "#excel-updates"
        conflicts: "#excel-alerts"
    
    webhooks:
      - url: "https://api.company.com/gitcells"
        events: ["change", "conflict", "error"]
        headers:
          Authorization: "Bearer ${API_TOKEN}"
```

## Security Configuration

```yaml
security:
  # Encryption
  encryption:
    enabled: false
    algorithm: "AES-256"
    key_file: "~/.gitcells/key"
  
  # File protection
  protection:
    remove_macros: false
    remove_external_links: false
    remove_personal_info: true
    
  # Audit
  audit:
    enabled: true
    log_file: ".gitcells/audit.log"
    include_cell_values: false
    retention: "90d"
```

## Performance Configuration

```yaml
performance:
  # Memory management
  memory:
    max_heap: "2GB"
    gc_interval: "5m"
    low_memory_mode: false
  
  # Caching
  cache:
    enabled: true
    directory: ".gitcells/cache"
    max_size: "1GB"
    ttl: "24h"
    compression: true
  
  # Parallel processing
  parallel:
    enabled: true
    max_workers: 8
    queue_size: 100
    
  # File handling
  files:
    chunk_size: "10MB"
    buffer_size: "64KB"
    use_mmap: true
```

## Hooks Configuration

```yaml
hooks:
  # Pre-conversion hooks
  pre_convert:
    - name: "Validate data"
      script: "./scripts/validate.sh"
      args: ["{filepath}"]
      on_error: "abort"  # abort|continue|warn
      timeout: "30s"
    
    - name: "Backup file"
      script: "./scripts/backup.sh"
      on_error: "warn"
  
  # Post-conversion hooks
  post_convert:
    - name: "Update dashboard"
      script: "./scripts/update-dashboard.py"
      args: ["{filepath}", "{changes}"]
      async: true
  
  # Error hooks
  on_error:
    - name: "Send alert"
      script: "./scripts/alert.sh"
      args: ["{error}", "{filepath}"]
  
  # Custom triggers
  triggers:
    - name: "Large change detection"
      condition: "cells_changed > 100"
      script: "./scripts/review-required.sh"
    
    - name: "Formula change"
      condition: "formulas_changed > 0"
      script: "./scripts/validate-formulas.sh"
```

## Environment Variables

```yaml
# Use environment variables
database:
  host: "${DB_HOST}"
  port: "${DB_PORT:5432}"  # With default
  password: "${DB_PASSWORD}"
  
# Or use env section
env:
  - DB_HOST=localhost
  - DB_PORT=5432
  - LOG_LEVEL=info
```

## Profiles

```yaml
# Define multiple profiles
profiles:
  development:
    watch:
      auto_commit: true
      debounce: "1s"
    conversion:
      json:
        pretty: true
  
  production:
    watch:
      auto_commit: false
      debounce: "10s"
    security:
      audit:
        enabled: true
    performance:
      cache:
        enabled: true
  
  ci:
    watch:
      enabled: false
    validation:
      fail_on_warning: true
```

Activate profile:
```bash
gitcells --profile production watch
# Or
export GITCELLS_PROFILE=production
```

## Advanced Configuration

### Custom Converters

```yaml
converters:
  custom:
    - name: "Legacy Excel"
      pattern: "*.xls"
      handler: "./converters/legacy.py"
    
    - name: "Special Format"
      pattern: "*special*.xlsx"
      handler: "./converters/special.js"
```

### Plugins

```yaml
plugins:
  - name: "jira-integration"
    path: "~/.gitcells/plugins/jira"
    config:
      url: "https://jira.company.com"
      project: "EXCEL"
  
  - name: "s3-backup"
    path: "gitcells-plugin-s3"
    config:
      bucket: "excel-backups"
      region: "us-east-1"
```

### Logging

```yaml
logging:
  level: "info"  # debug|info|warn|error
  
  outputs:
    - type: "file"
      path: ".gitcells/logs/gitcells.log"
      rotation: "daily"
      retention: "7d"
      format: "json"
    
    - type: "console"
      format: "text"
      color: true
    
    - type: "syslog"
      facility: "local0"
      tag: "gitcells"
```

## Validation

Validate your configuration:

```bash
gitcells config --validate
```

Schema validation:
```bash
gitcells config --schema > schema.json
```

## Examples

### Minimal Configuration

```yaml
# .gitcells.yml
watch:
  enabled: true
  auto_commit: true
```

### Financial Team Configuration

```yaml
# .gitcells.yml
project:
  name: "Financial Reports"

files:
  patterns: ["reports/*.xlsx", "budgets/*.xlsx"]
  max_size: "25MB"

watch:
  enabled: true
  debounce: "5s"
  
auto_commit:
  message:
    template: "[Finance] {user}: {action} {filename}"
  
team:
  notifications:
    email:
      recipients:
        changes: ["finance@company.com"]
        
security:
  audit:
    enabled: true
  protection:
    remove_personal_info: true
```

### Research Team Configuration

```yaml
# .gitcells.yml
project:
  name: "Research Data"

conversion:
  json:
    include_metadata: true
    include_empty_cells: true

sync:
  validation:
    before_sync: true
    after_sync: true

hooks:
  pre_convert:
    - name: "Validate data integrity"
      script: "./validate-research-data.py"
      on_error: "abort"

security:
  encryption:
    enabled: true
  audit:
    enabled: true
    include_cell_values: true
```

## Next Steps

- Review [command reference](commands.md)
- Learn about [troubleshooting](troubleshooting.md)
- Explore [team workflows](../guides/collaboration.md)