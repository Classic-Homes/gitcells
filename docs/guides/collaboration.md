# Working with Teams

Learn how to use GitCells effectively in a team environment.

## Team Setup

### Initial Repository Setup

1. **Create shared repository**:
```bash
git init excel-projects
cd excel-projects
gitcells init --team
```

2. **Configure team settings**:
```yaml
# .gitcells.yml
team:
  auto_sync: true
  conflict_strategy: prompt
  commit_style: descriptive
  protected_branches: [main, production]
```

3. **Share with team**:
```bash
git remote add origin https://github.com/company/excel-projects
git push -u origin main
```

## Team Workflows

### Basic Collaboration Flow

1. **Clone and setup**:
```bash
git clone https://github.com/company/excel-projects
cd excel-projects
gitcells init
```

2. **Daily workflow**:
```bash
# Start of day
git pull
gitcells sync

# Work in Excel
gitcells watch

# End of day
git push
```

### Feature Branch Workflow

Working on major changes:

```bash
# Create feature branch
git checkout -b update-q4-forecasts

# Make changes in Excel
# GitCells tracks automatically

# Review changes
gitcells status
gitcells diff *.xlsx

# Commit and push
git commit -am "Update Q4 forecasts with latest data"
git push origin update-q4-forecasts

# Create pull request
```

### Parallel Editing

Multiple people editing different sheets:

```yaml
# .gitcells.yml
collaboration:
  sheet_locking: true
  owner_tracking: true
```

```bash
# Alice works on Sheet1
gitcells lock budget.xlsx --sheet "Revenue"

# Bob works on Sheet2  
gitcells lock budget.xlsx --sheet "Expenses"

# Both can work simultaneously
```

## Conflict Prevention

### 1. Communication

Use commit messages effectively:
```bash
git commit -m "WIP: Updating formulas in column D - DO NOT EDIT"
```

### 2. File Organization

Structure files to minimize conflicts:
```
/reports
  /alice
    - sales-analysis.xlsx
  /bob
    - inventory-report.xlsx
  /shared
    - master-budget.xlsx (coordinate edits)
```

### 3. Time-based Editing

Schedule edit windows:
```yaml
# .gitcells.yml
schedule:
  maintenance_window: "Sunday 2-4 AM"
  readonly_hours: "Friday 5PM - Monday 8AM"
```

## Merge Strategies

### Auto-merge Safe Changes

Configure auto-merge rules:
```yaml
# .gitcells.yml
merge:
  auto_merge:
    - different_sheets: true
    - different_cells: true
    - formatting_only: true
  require_review:
    - formula_changes: true
    - same_cell: true
```

### Manual Merge Process

When conflicts occur:

```bash
# Pull changes
git pull

# GitCells detects conflict
gitcells status
# Output: Conflict in budget.xlsx

# Review conflicts
gitcells conflict budget.xlsx --show

# Resolve
gitcells conflict budget.xlsx --resolve
```

## Code Review for Excel

### Pull Request Best Practices

1. **Create descriptive PRs**:
```markdown
## Summary
Updated Q4 revenue projections based on October actuals

## Changes
- Sheet "Revenue": Cells B15-B27 updated with new figures
- Sheet "Summary": Formula in D5 now includes October data
- Added new "October Actuals" sheet

## Testing
- [x] Formulas verified
- [x] Totals match expected values
- [x] No circular references
```

2. **Review checklist**:
- [ ] Formulas are correct
- [ ] No hardcoded values that should be formulas
- [ ] Consistent formatting
- [ ] Protected cells remain protected
- [ ] No sensitive data exposed

### Review Tools

```bash
# Generate review report
gitcells review PR-123 --report

# Compare before/after
gitcells compare main..feature-branch --visual

# Validate formulas
gitcells validate *.xlsx --check-formulas
```

## Access Control

### Role-based Permissions

```yaml
# .gitcells.yml
permissions:
  admins:
    - can_edit: "*"
    - can_merge: true
    - can_delete: true
  
  analysts:
    - can_edit: ["reports/*.xlsx"]
    - can_merge: false
    - protected_cells: ["*!A1:A10"]
  
  viewers:
    - can_edit: []
    - readonly: true
```

### Protected Elements

Prevent accidental changes:
```bash
# Protect formulas
gitcells protect budget.xlsx --formulas

# Protect specific ranges
gitcells protect budget.xlsx --range "A1:D10"

# Protect structure
gitcells protect budget.xlsx --structure
```

## Team Communication

### Change Notifications

Configure team alerts:
```yaml
# .gitcells.yml
notifications:
  email:
    on_change: ["team@company.com"]
    on_conflict: ["lead@company.com"]
  
  slack:
    webhook: "https://hooks.slack.com/..."
    channels:
      changes: "#excel-updates"
      conflicts: "#excel-urgent"
```

### Change Summaries

Daily digest for team:
```bash
# Generate daily summary
gitcells summary --since yesterday --email team@company.com

# Post to Slack
gitcells summary --since @{8.hours.ago} --slack
```

## Best Practices

### 1. Establish Conventions

Document team standards:
- Naming conventions for files and sheets
- Formula standards
- Comment requirements
- Commit message format

### 2. Regular Syncs

Schedule team syncs:
```bash
# Monday morning sync
git pull
gitcells sync --all
gitcells status --team
```

### 3. Backup Strategy

Implement redundancy:
```yaml
# .gitcells.yml
backup:
  enabled: true
  frequency: hourly
  destination: "s3://backups/excel/"
  retention: 30
```

### 4. Training

Ensure team knows:
- Basic Git commands
- GitCells workflow
- Conflict resolution
- Best practices

## Common Scenarios

### Scenario 1: Quarterly Close

```bash
# Lock master files
gitcells lock quarter-end/*.xlsx --message "Quarterly close in progress"

# Create working branch
git checkout -b q4-close

# Make updates
# ... work in Excel ...

# Review all changes
gitcells diff --summary > q4-changes.txt

# Merge after approval
git checkout main
git merge q4-close --no-ff
```

### Scenario 2: Template Updates

```bash
# Update template
gitcells convert template.xlsx
git commit -m "Update template with new categories"

# Apply to all files
gitcells apply-template template.xlsx reports/*.xlsx
```

### Scenario 3: Audit Preparation

```bash
# Generate audit trail
gitcells audit --from 2024-01-01 --to 2024-12-31

# Lock files
gitcells lock *.xlsx --message "Audit in progress"

# Create read-only copies
gitcells export *.xlsx --readonly --to audit/
```

## Troubleshooting

### Common Team Issues

**"File is locked"**
```bash
# Check lock status
gitcells lock --status

# Force unlock (admin only)
gitcells unlock budget.xlsx --force
```

**"Merge conflicts"**
```bash
# Show conflict details
gitcells conflict --details

# Use theirs/ours
git checkout --theirs budget.json
gitcells sync
```

**"Out of sync"**
```bash
# Force sync
gitcells sync --force --all

# Reset to remote
git reset --hard origin/main
gitcells sync
```

## Next Steps

- Set up [conflict resolution](conflicts.md)
- Configure [auto-sync](auto-sync.md)
- Learn about [advanced workflows](../reference/commands.md)