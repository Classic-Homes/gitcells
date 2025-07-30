# Resolving Conflicts

Learn how to handle merge conflicts when multiple people edit Excel files.

## Understanding Conflicts

### When Conflicts Occur

Conflicts happen when:
- Two people edit the same cell
- Formulas reference conflicting changes
- Sheets are reorganized differently
- Incompatible formatting changes

### Conflict Types

1. **Cell Value Conflicts**: Same cell, different values
2. **Formula Conflicts**: Formula references changed cells
3. **Structure Conflicts**: Sheets added/removed/renamed
4. **Format Conflicts**: Same cell, different formatting

## Detecting Conflicts

### Automatic Detection

GitCells detects conflicts during:
```bash
# Pull/merge operations
git pull
# GitCells: Conflict detected in budget.xlsx

# Status checks
gitcells status
# Output: 1 conflicted file
```

### Manual Check

Explicitly check for conflicts:
```bash
gitcells conflict --check
```

## Viewing Conflicts

### Basic Conflict View

```bash
gitcells conflict budget.xlsx
```

Output:
```
Conflicts in budget.xlsx:

Sheet: Summary
  Cell B5:
    Local:  1500 (yours)
    Remote: 2000 (theirs)
    Base:   1000 (original)
  
  Cell D10:
    Local:  =SUM(B1:B10)
    Remote: =SUM(B1:B15)
    Base:   =SUM(B1:B5)
```

### Detailed View

```bash
gitcells conflict budget.xlsx --detailed
```

Shows:
- Who made each change
- When changes were made
- Related cell dependencies
- Potential formula impacts

### Visual Comparison

```bash
gitcells conflict budget.xlsx --visual
```

Opens a side-by-side comparison view.

## Resolution Strategies

### 1. Interactive Resolution

Step through conflicts one by one:
```bash
gitcells resolve budget.xlsx --interactive
```

For each conflict, choose:
- **Keep Mine** (local version)
- **Keep Theirs** (remote version)
- **Keep Base** (original version)
- **Merge** (combine both changes)
- **Custom** (enter new value)

### 2. Strategy-Based Resolution

Apply resolution strategy:
```bash
# Always use remote version
gitcells resolve budget.xlsx --strategy theirs

# Always use local version
gitcells resolve budget.xlsx --strategy ours

# Use newer timestamp
gitcells resolve budget.xlsx --strategy newer

# Use higher value (for numbers)
gitcells resolve budget.xlsx --strategy max
```

### 3. Rule-Based Resolution

Configure automatic rules:
```yaml
# .gitcells.yml
conflict_resolution:
  rules:
    - pattern: "*revenue*"
      strategy: max
    - pattern: "*cost*"
      strategy: min
    - sheet: "Summary"
      strategy: theirs
    - cell_range: "A1:A10"
      strategy: ours
```

## Manual Resolution

### Using Excel

1. Open conflict file:
```bash
gitcells conflict budget.xlsx --open
```

2. GitCells creates a special sheet showing:
   - Conflicted cells highlighted
   - Both versions side-by-side
   - Resolution column for your choice

3. Save and mark resolved:
```bash
gitcells resolved budget.xlsx
```

### Using JSON

For complex conflicts, edit JSON directly:
```bash
# Open conflict markers in JSON
git status
# Conflicts in budget.json

# Edit JSON file
vim budget.json

# Look for conflict markers
<<<<<<< HEAD
  "B5": {"value": 1500}
=======
  "B5": {"value": 2000}
>>>>>>> feature-branch

# Resolve and save
# Then sync back to Excel
gitcells sync budget.json
```

## Formula Conflicts

### Understanding Formula Dependencies

When formulas conflict:
```
Cell D5: =B5+C5
  Conflict: B5 has different values
  Impact: D5 result will change
```

### Resolution Options

```bash
# Show formula impacts
gitcells conflict budget.xlsx --show-formulas

# Recalculate after resolution
gitcells resolve budget.xlsx --recalculate
```

## Advanced Conflict Handling

### Partial Resolution

Resolve specific conflicts:
```bash
# Resolve only cell B5
gitcells resolve budget.xlsx --cell B5 --use theirs

# Resolve entire sheet
gitcells resolve budget.xlsx --sheet Summary --use ours
```

### Batch Resolution

Handle multiple files:
```bash
# Resolve all with same strategy
gitcells resolve *.xlsx --strategy newer

# Interactive batch mode
gitcells resolve *.xlsx --interactive --save-choices
```

### Three-Way Merge

Use base version for smarter merging:
```bash
gitcells merge budget.xlsx --three-way
```

Shows:
- What changed in yours vs base
- What changed in theirs vs base
- Suggests resolution based on changes

## Conflict Prevention

### Best Practices

1. **Communicate**: Tell team what you're editing
2. **Pull Often**: `git pull` before starting work
3. **Commit Often**: Smaller commits = easier conflicts
4. **Use Branches**: Isolate major changes

### Locking Mechanism

Prevent conflicts with locks:
```bash
# Lock before editing
gitcells lock budget.xlsx --sheet Revenue

# Others see lock
gitcells status
# budget.xlsx: Sheet "Revenue" locked by alice

# Unlock when done
gitcells unlock budget.xlsx
```

### Protected Ranges

Define non-editable areas:
```yaml
# .gitcells.yml
protected:
  - file: budget.xlsx
    range: A1:A10
    users: [admin]
    message: "Headers protected"
```

## Conflict Workflows

### Daily Standup Flow

```bash
# Morning sync
git pull

# Check for conflicts
gitcells conflict --check

# Resolve before starting
gitcells resolve *.xlsx --interactive

# Begin work
gitcells watch
```

### Release Process

```bash
# Merge feature branch
git merge feature-branch

# Bulk resolve expected conflicts
gitcells resolve *.xlsx --config release-rules.yml

# Validate resolution
gitcells validate *.xlsx

# Commit resolution
git commit -m "Resolved conflicts for release"
```

## Troubleshooting

### Common Issues

**"Cannot auto-resolve"**
```bash
# Try three-way merge
gitcells merge file.xlsx --three-way

# Fall back to manual
gitcells conflict file.xlsx --export
# Resolve manually
gitcells resolved file.xlsx
```

**"Formula broken after resolve"**
```bash
# Recalculate formulas
gitcells recalculate file.xlsx

# Validate formulas
gitcells validate file.xlsx --formulas
```

**"Lost formatting"**
```bash
# Merge formatting separately
gitcells merge file.xlsx --formatting-only
```

### Recovery Options

If resolution goes wrong:
```bash
# Abort merge
git merge --abort
gitcells reset

# Restore from backup
gitcells restore budget.xlsx --from-backup

# Start over
git reset --hard HEAD
gitcells sync
```

## Integration

### CI/CD Conflict Detection

```yaml
# .github/workflows/conflict-check.yml
name: Conflict Check
on: [pull_request]
jobs:
  check:
    steps:
      - name: Check for conflicts
        run: |
          gitcells conflict --check
          gitcells validate *.xlsx
```

### Automated Resolution

```bash
# Scheduled conflict resolution
0 6 * * * gitcells resolve *.xlsx --auto --email conflicts@company.com
```

## Next Steps

- Set up [auto-sync](auto-sync.md) to minimize conflicts
- Learn about [team workflows](collaboration.md)
- Configure [conflict prevention](../reference/configuration.md#conflict-prevention)