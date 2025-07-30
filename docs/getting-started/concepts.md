# Basic Concepts

Understanding these core concepts will help you use GitCells effectively.

## Excel to JSON Conversion

GitCells converts Excel files into structured JSON that preserves:

- **Cell Values**: Numbers, text, dates, booleans
- **Formulas**: Complete formula expressions
- **Formatting**: Colors, fonts, borders, number formats
- **Structure**: Merged cells, hidden rows/columns
- **Metadata**: Sheet names, print settings, protection

### Why JSON?

- **Human-readable**: Easy to review changes
- **Git-friendly**: Line-based diffs work perfectly
- **Preserves everything**: No data loss during conversion
- **Mergeable**: Git can merge non-conflicting changes

## File Pairs

Each Excel file has a corresponding JSON file:

```
financial-report.xlsx  ←→  financial-report.json
sales-data.xlsx       ←→  sales-data.json
```

GitCells keeps these synchronized automatically.

## The GitCells Directory

`.gitcells/` stores:
- Conversion cache
- Temporary files
- Sync state
- Lock files

This directory should be in your `.gitignore`.

## Version Control Benefits

### Track Changes
```bash
# Who changed the formula in cell B5?
git blame financial-report.json | grep "B5"

# When was the budget updated?
git log -p -- budget.json | grep "total_budget"
```

### Collaboration
- Multiple people can work on different sheets
- Changes are merged automatically
- Conflicts are detected and can be resolved

### History
- Revert to any previous version
- Compare versions side-by-side
- Create audit trails

## Sync States

Files can be in different states:

1. **In Sync**: Excel and JSON match perfectly
2. **Excel Newer**: Excel changed, needs conversion
3. **JSON Newer**: JSON changed, needs reverse conversion
4. **Conflicted**: Both changed, requires resolution

## Automatic Operations

When watching is enabled, GitCells:

1. **Detects Changes**: Monitors Excel files
2. **Converts**: Updates JSON representation
3. **Commits**: Creates Git commit with changes
4. **Syncs**: Keeps file pairs synchronized

## Working with Git

GitCells enhances Git, it doesn't replace it:

```bash
# GitCells commands
gitcells convert   # Convert files
gitcells watch     # Monitor changes

# Regular Git commands
git add            # Stage changes
git commit         # Commit changes
git push/pull      # Share with team
```

## Best Practices

### 1. Close Before Pull
Always close Excel files before `git pull` to avoid conflicts.

### 2. Commit Logical Changes
Group related changes together:
```bash
git add budget.json forecast.json
git commit -m "Update Q4 financial projections"
```

### 3. Use Branches
Create branches for major changes:
```bash
git checkout -b new-product-line
# Make Excel changes
git checkout main
git merge new-product-line
```

### 4. Review JSON Changes
Before committing, review what changed:
```bash
git diff financial-report.json
```

## Common Patterns

### Daily Workflow
1. `git pull` - Get latest changes
2. Open Excel, make changes
3. `gitcells watch` - Auto-track changes
4. `git push` - Share with team

### Monthly Reporting
1. Create branch: `git checkout -b monthly-report-jan`
2. Update all reports in Excel
3. Review: `gitcells status`
4. Commit: `git commit -am "January monthly reports"`
5. Merge: `git checkout main && git merge monthly-report-jan`

## Next Steps

- Try the [hands-on tutorials](../guides/converting.md)
- Learn about [configuration options](../reference/configuration.md)
- Explore [advanced workflows](../guides/collaboration.md)