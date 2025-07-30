# Quick Start Guide

Get up and running with GitCells in 5 minutes!

## Step 1: Initialize GitCells

Navigate to your Excel project directory and initialize GitCells:

```bash
cd /path/to/your/excel/project
gitcells init
```

This creates:
- `.gitcells/` directory for GitCells data
- `.gitcells.yml` configuration file
- `.gitignore` entries for Excel temporary files

## Step 2: Convert Your First Excel File

Convert an Excel file to JSON:

```bash
gitcells convert myspreadsheet.xlsx
```

This creates `myspreadsheet.json` with all your Excel data, formulas, and formatting preserved.

## Step 3: Track Changes

View what's changed in your Excel files:

```bash
gitcells status
```

See detailed differences:

```bash
gitcells diff myspreadsheet.xlsx
```

## Step 4: Enable Auto-sync

Start watching for changes and auto-commit them:

```bash
gitcells watch
```

Now any changes to Excel files are automatically:
- Converted to JSON
- Committed to Git
- Tracked with descriptive messages

## Common Workflows

### Team Collaboration

1. Team member A makes changes in Excel
2. GitCells auto-converts and commits
3. Push changes: `git push`
4. Team member B pulls: `git pull`
5. GitCells auto-converts JSON back to Excel

### Review Changes

```bash
# See recent changes
git log --oneline -10

# View specific change
gitcells diff myspreadsheet.xlsx HEAD~1

# Revert changes
git checkout HEAD~1 -- myspreadsheet.json
gitcells sync
```

### Working with Multiple Files

```bash
# Convert all Excel files
gitcells convert *.xlsx

# Watch specific directory
gitcells watch --dir reports/

# Sync all files
gitcells sync
```

## Tips for Success

1. **Commit Often**: Use Git's commit history to track incremental changes
2. **Use Branches**: Create branches for major changes or experiments
3. **Write Good Messages**: GitCells auto-generates messages, but you can amend them
4. **Close Excel**: Close Excel files before pulling changes to avoid conflicts

## Example: Financial Report Workflow

```bash
# Start a new feature branch
git checkout -b update-q4-figures

# Make changes in Excel
# GitCells auto-tracks them

# Review changes
gitcells status
gitcells diff financial-report.xlsx

# Commit with custom message
git add .
git commit -m "Update Q4 revenue figures and projections"

# Merge back
git checkout main
git merge update-q4-figures
```

## Next Steps

- Learn about [Basic Concepts](concepts.md)
- Explore [Advanced Features](../guides/converting.md)
- Configure [Auto-sync](../guides/auto-sync.md)
- Read about [Conflict Resolution](../guides/conflicts.md)