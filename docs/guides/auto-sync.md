# Auto-sync Setup

Configure GitCells to automatically track and sync your Excel files.

## Quick Start

### Basic Auto-sync

Start watching files:
```bash
gitcells watch
```

This monitors all Excel files in the current directory and:
- Converts changes to JSON immediately
- Creates automatic commits
- Syncs file pairs

### Stop Watching

```bash
# Ctrl+C in the terminal
# Or
gitcells watch --stop
```

## Configuration

### Watch Settings

Configure in `.gitcells.yml`:
```yaml
watch:
  enabled: true
  patterns:
    - "*.xlsx"
    - "*.xlsm"
  ignore:
    - "~$*"  # Temporary Excel files
    - "*.tmp"
    - "archived/*"
  
  debounce: 2s  # Wait 2 seconds after changes stop
  
  auto_commit: true
  commit_message: "Auto-update: {filename} at {timestamp}"
```

### Directory Watching

Watch specific directories:
```bash
# Single directory
gitcells watch reports/

# Multiple directories
gitcells watch --dir reports --dir budgets

# Recursive watching
gitcells watch --recursive
```

## Auto-commit Options

### Commit Strategies

```yaml
# .gitcells.yml
auto_commit:
  enabled: true
  
  # Commit after each file change
  strategy: immediate
  
  # Or batch changes
  strategy: batch
  batch_timeout: 5m
  
  # Or on schedule
  strategy: scheduled
  schedule: "*/15 * * * *"  # Every 15 minutes
```

### Commit Messages

Customize commit messages:
```yaml
commit:
  template: "[GitCells] {action} {filename}"
  
  variables:
    - filename: basename of changed file
    - filepath: full path
    - timestamp: ISO 8601 timestamp
    - user: system username
    - action: update/create/delete
    - sheets: changed sheet names
    - cells: number of cells changed
```

Examples:
- `[GitCells] Update budget.xlsx at 2024-01-15T10:30:45Z`
- `[GitCells] Alice updated 5 cells in revenue.xlsx`
- `[GitCells] Modified sheets: Summary, Details in report.xlsx`

## Advanced Watching

### Selective Watching

Watch only specific aspects:
```bash
# Only watch for formula changes
gitcells watch --filter formulas

# Only specific sheets
gitcells watch --sheets "Summary,Totals"

# Only when values change significantly
gitcells watch --threshold 100
```

### Performance Options

For large directories:
```yaml
watch:
  performance:
    max_files: 1000
    parallel_conversions: 4
    cache_enabled: true
    low_memory_mode: false
```

### Network Shares

Watch network locations:
```yaml
watch:
  network:
    retry_attempts: 3
    retry_delay: 5s
    offline_mode: true  # Queue changes when offline
```

## Triggers and Hooks

### Pre-conversion Hooks

Run scripts before conversion:
```yaml
hooks:
  pre_convert:
    - script: "./scripts/validate.sh"
      on_error: abort
    
    - script: "./scripts/backup.sh"
      on_error: continue
```

### Post-conversion Hooks

After successful conversion:
```yaml
hooks:
  post_convert:
    - script: "./scripts/notify-team.sh"
      args: ["{filename}", "{changes}"]
    
    - script: "./scripts/update-dashboard.py"
      async: true
```

### Custom Triggers

Define custom conditions:
```yaml
triggers:
  - name: "Large change alert"
    condition: "cells_changed > 100"
    action: 
      script: "./alert-manager.sh"
      email: "managers@company.com"
  
  - name: "Formula modification"
    condition: "formulas_changed > 0"
    action:
      require_approval: true
      approvers: ["lead@company.com"]
```

## Integration Patterns

### Development Workflow

```yaml
# .gitcells.yml
profiles:
  development:
    watch:
      auto_commit: true
      commit_prefix: "WIP: "
      branch_pattern: "dev-{user}-{date}"
  
  production:
    watch:
      auto_commit: false
      require_approval: true
      protected_files: ["master-*.xlsx"]
```

### CI/CD Integration

```yaml
# .github/workflows/excel-sync.yml
name: Excel Sync
on:
  push:
    paths:
      - '**.xlsx'
      - '**.json'

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Install GitCells
        run: |
          curl -sSL https://install.gitcells.com | bash
      
      - name: Sync files
        run: |
          gitcells sync --all
          gitcells validate *.xlsx
      
      - name: Commit changes
        run: |
          git config --global user.name "GitCells Bot"
          git add .
          git commit -m "Auto-sync Excel files" || true
          git push
```

### Docker Setup

Run auto-sync in container:
```dockerfile
FROM gitcells/gitcells:latest

WORKDIR /workspace
COPY .gitcells.yml .

CMD ["gitcells", "watch", "--daemon"]
```

```bash
docker run -d \
  -v $(pwd):/workspace \
  -v ~/.gitconfig:/root/.gitconfig \
  gitcells-watcher
```

## Monitoring

### Watch Status

Check what's being watched:
```bash
gitcells watch --status
```

Output:
```
Watching 5 files in 2 directories
Active since: 2024-01-15 09:00:00
Files watched: 5
Changes detected: 12
Commits created: 8
Last change: 2 minutes ago
```

### Logs

View watch logs:
```bash
# Live log
gitcells watch --log

# Historical logs
cat .gitcells/logs/watch.log

# Filter logs
gitcells logs --filter "ERROR|WARN"
```

### Metrics

Track sync performance:
```yaml
metrics:
  enabled: true
  export: prometheus
  endpoint: :9090
  
  track:
    - files_watched
    - changes_detected
    - conversion_time
    - commit_frequency
```

## Troubleshooting

### Common Issues

**"Changes not detected"**

Check ignore patterns:
```bash
gitcells watch --debug
# Shows which files are ignored and why
```

**"Too many commits"**

Adjust debouncing:
```yaml
watch:
  debounce: 10s  # Increase wait time
  batch_changes: true
  batch_timeout: 5m
```

**"High CPU usage"**

Optimize watching:
```yaml
watch:
  performance:
    polling: true  # Use polling instead of events
    interval: 30s  # Check every 30 seconds
    exclude_large: true  # Skip files > 10MB
```

### Recovery

Reset watch state:
```bash
# Stop all watchers
gitcells watch --stop-all

# Clear watch cache
gitcells cache clear --watch

# Restart fresh
gitcells watch --reset
```

## Best Practices

### 1. Appropriate Debouncing

Set based on workflow:
- Quick edits: 1-2 seconds
- Complex work: 5-10 seconds
- Batch processing: 1-5 minutes

### 2. Meaningful Commits

Configure descriptive messages:
```yaml
commit:
  template: "{user}: {action} {sheets} in {filename}"
  include_summary: true
  max_length: 72
```

### 3. Resource Management

Prevent resource exhaustion:
```yaml
watch:
  limits:
    max_files: 500
    max_file_size: 50MB
    max_memory: 1GB
    cpu_limit: 50%
```

### 4. Error Handling

Configure resilience:
```yaml
watch:
  error_handling:
    retry_failed: true
    quarantine_errors: true
    alert_on_failure: true
    fallback_mode: manual
```

## Examples

### Financial Team Setup

```yaml
# .gitcells.yml
watch:
  patterns: ["reports/*.xlsx", "budgets/*.xlsx"]
  
  schedules:
    - name: "Morning sync"
      time: "08:00"
      action: "full_sync"
    
    - name: "Hourly backup"
      time: "0 * * * *"
      action: "backup"
  
  notifications:
    changes: ["finance-team@company.com"]
    errors: ["it-support@company.com"]
```

### Research Team Setup

```yaml
# .gitcells.yml
watch:
  patterns: ["data/*.xlsx", "analysis/*.xlsx"]
  
  validation:
    required: true
    rules:
      - no_external_links
      - no_macros
      - formula_validation
  
  auto_commit:
    require_message: true
    require_approval: true
    approvers: ["pi@university.edu"]
```

## Next Steps

- Configure [team collaboration](collaboration.md)
- Set up [conflict resolution](conflicts.md)
- Learn about [advanced configuration](../reference/configuration.md)