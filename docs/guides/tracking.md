# Tracking Changes

Learn how to monitor and track changes to your Excel files using GitCells.

## Viewing Status

### Check Current Status

See which files have changed:
```bash
gitcells status
```

Output example:
```
Excel files status:
  Modified:    financial-report.xlsx (2 minutes ago)
  Synced:      budget-2024.xlsx
  Untracked:   new-analysis.xlsx
  Conflicted:  shared-data.xlsx
```

### Detailed Status

Get more information:
```bash
gitcells status --detailed
```

Shows:
- Last modified time
- File size changes
- Number of cells changed
- Sync state

## Viewing Differences

### Basic Diff

Compare Excel file with last commit:
```bash
gitcells diff report.xlsx
```

### Compare Versions

Compare with specific commit:
```bash
gitcells diff report.xlsx HEAD~1
```

Compare two commits:
```bash
gitcells diff report.xlsx abc123..def456
```

### Diff Output Formats

**Summary view** (default):
```
Sheet: Summary
  Cell B5: 1000 → 1500
  Cell C10: "Q3" → "Q4"
  Formula D15: =SUM(B:B) → =SUM(B:C)
```

**Detailed view**:
```bash
gitcells diff report.xlsx --detailed
```

**JSON diff**:
```bash
gitcells diff report.xlsx --json
```

## Change Types

### Cell Value Changes

Shows old and new values:
```
Cell A1: "Revenue" → "Total Revenue"
Cell B2: 50000 → 75000
```

### Formula Changes

Highlights formula modifications:
```
Cell D5: =SUM(A1:A10) → =SUM(A1:A20)
Cell E10: =B5*C5 → =B5*C5*1.1
```

### Formatting Changes

Track style updates:
```
Cell A1: Font changed from Arial to Calibri
Cell B2: Background color: none → yellow
Row 5: Hidden → Visible
```

### Structural Changes

Monitor sheet modifications:
```
Sheet "Data" renamed to "Raw Data"
Sheet "Analysis" added
Column C deleted
Rows 10-15 inserted
```

## Git Integration

### Commit History

View Excel-specific history:
```bash
# Show commits affecting Excel files
git log --oneline -- "*.xlsx"

# Show detailed changes
git log -p -- report.json
```

### Blame

Find who changed specific cells:
```bash
# Check JSON file
git blame report.json | grep "B5"

# Use GitCells blame
gitcells blame report.xlsx B5
```

## Advanced Tracking

### Watch Mode

Monitor changes in real-time:
```bash
gitcells watch --verbose
```

Shows:
```
[10:23:45] Detected change: budget.xlsx
[10:23:46] Converting to JSON...
[10:23:47] Changes detected in cells: B5, C10, D15
[10:23:48] Auto-commit: Update budget figures
```

### Change Notifications

Configure notifications:
```yaml
# .gitcells.yml
notifications:
  on_change: true
  webhook: "https://your-webhook.com"
  email: "team@company.com"
```

### Change Filters

Track specific changes:
```bash
# Only formula changes
gitcells diff report.xlsx --filter formulas

# Only specific sheets
gitcells diff report.xlsx --sheets "Summary,Totals"

# Value changes over threshold
gitcells diff report.xlsx --threshold 1000
```

## Change Reports

### Generate Summary

Create a change report:
```bash
gitcells report --from HEAD~10 --to HEAD
```

Output:
```markdown
# Change Report
Period: 2024-01-01 to 2024-01-15

## Files Modified
- budget.xlsx: 45 changes
- forecast.xlsx: 23 changes

## Top Changes
1. Cell Budget!B5: 15 modifications
2. Formula Forecast!D10: 8 modifications
```

### Export Changes

Export to CSV:
```bash
gitcells diff report.xlsx --export changes.csv
```

## Tracking Patterns

### Daily Standup

Quick morning check:
```bash
# What changed yesterday?
gitcells diff *.xlsx @{yesterday}

# Who made changes?
git log --since=yesterday --pretty=format:"%an: %s" -- "*.xlsx"
```

### Weekly Review

Comprehensive weekly analysis:
```bash
# Generate weekly report
gitcells report --from @{1.week.ago}

# Show major changes
gitcells diff *.xlsx @{1.week.ago} --threshold 10000
```

### Audit Trail

Create audit log:
```bash
# Full history for specific file
gitcells history budget.xlsx --full

# Export audit trail
gitcells audit budget.xlsx --export audit-log.pdf
```

## Best Practices

### 1. Regular Status Checks

Make it a habit:
```bash
# Before starting work
gitcells status

# After making changes
gitcells diff

# Before committing
gitcells status --detailed
```

### 2. Meaningful Commits

Group related changes:
```bash
# Bad: Auto-commit everything
gitcells watch --auto-commit

# Good: Logical commits
git add budget.json forecast.json
git commit -m "Update Q4 projections based on actuals"
```

### 3. Use Branches

Track experimental changes:
```bash
git checkout -b test-scenarios
# Make Excel changes
gitcells diff --summary
git checkout main
```

## Troubleshooting

### Missing Changes

If changes aren't detected:
```bash
# Force rescan
gitcells status --force

# Check watch status
gitcells watch --status

# Clear cache
gitcells cache clear
```

### Performance

For large files:
```bash
# Quick diff
gitcells diff large.xlsx --quick

# Cached diff
gitcells diff large.xlsx --cached
```

## Integration

### Slack Notifications

```.gitcells.yml
notifications:
  slack:
    webhook: "https://hooks.slack.com/..."
    channel: "#excel-changes"
    mentions: ["@finance-team"]
```

### Email Reports

```bash
# Daily email summary
gitcells report --email team@company.com --schedule daily
```

## Next Steps

- Learn about [team collaboration](collaboration.md)
- Set up [automated workflows](auto-sync.md)
- Configure [conflict resolution](conflicts.md)