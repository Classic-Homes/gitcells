# Git Integration Guide

GitCells seamlessly integrates with Git to provide version control for your Excel files. This guide explains how to set up and use Git features effectively.

## Prerequisites

Before using Git integration:
1. Install Git (version 2.0 or higher)
2. Initialize a Git repository in your project folder
3. Configure Git user name and email

```bash
# Check if Git is installed
git --version

# Initialize repository (if needed)
git init

# Configure Git user
git config user.name "Your Name"
git config user.email "your.email@example.com"
```

## Basic Git Integration

### Automatic Commits

GitCells can automatically commit changes when Excel files are modified:

```bash
# Watch with auto-commit enabled (default)
gitcells watch .

# Watch without auto-commit
gitcells watch --auto-commit=false .
```

### Commit Messages

GitCells uses configurable commit message templates:

```yaml
# In .gitcells.yaml
git:
  commit_template: "GitCells: {action} {filename} at {timestamp}"
```

Available variables:
- `{action}` - The type of change (created, modified, deleted)
- `{filename}` - Name of the Excel file
- `{timestamp}` - When the change occurred
- `{user}` - System username

Example messages:
- "GitCells: modified Budget2024.xlsx at 2024-01-15 10:30:45"
- "GitCells: created NewReport.xlsx at 2024-01-15 14:22:10"

## Configuration Options

### Complete Git Configuration

```yaml
git:
  # Branch to use for commits
  branch: main
  
  # Automatically push after commits
  auto_push: false
  
  # Pull before operations
  auto_pull: true
  
  # Git user for commits
  user_name: "GitCells Bot"
  user_email: "gitcells@company.com"
  
  # Custom commit template
  commit_template: "[GitCells] {user} {action} {filename}"
  
  # Add co-authors to commits
  co_authors:
    - "name:Co Author <coauthor@example.com>"
```

### Auto-Push Setup

To enable automatic pushing to remote:

1. Set up remote repository:
```bash
git remote add origin https://github.com/username/repo.git
```

2. Configure GitCells:
```yaml
git:
  auto_push: true
  branch: main
```

3. Ensure credentials are configured:
```bash
# For HTTPS
git config credential.helper store

# For SSH
ssh-add ~/.ssh/id_rsa
```

## Working with Branches

### Branch Configuration

Specify which branch to use:
```yaml
git:
  branch: develop  # Use develop branch instead of main
```

### Multi-Branch Workflow

For different branches per directory:
```bash
# Create branch-specific configs
# config-main.yaml
git:
  branch: main
  
# config-develop.yaml  
git:
  branch: develop

# Run separate watchers
gitcells watch --config config-main.yaml ./production
gitcells watch --config config-develop.yaml ./development
```

## Team Collaboration

### Shared Repository Setup

1. **Central Repository**: Create a shared Git repository
2. **Clone Locally**: Each team member clones the repository
3. **Configure GitCells**: Everyone uses the same `.gitcells.yaml`
4. **Start Watching**: Each person runs `gitcells watch`

### Handling Conflicts

GitCells helps prevent conflicts by:
- Auto-pulling before operations
- Using timestamps in commits
- Converting to mergeable JSON format

When conflicts occur:
```bash
# Check status
gitcells status

# Pull latest changes
git pull

# Resolve conflicts in JSON chunk files
git mergetool

# Convert back to Excel from chunks
gitcells convert .gitcells/data/resolved.xlsx_chunks/
```

### Best Practices for Teams

1. **Consistent Configuration**:
```yaml
git:
  auto_pull: true  # Always pull first
  commit_template: "{user}: Updated {filename}"
```

2. **Clear Commit Messages**:
```yaml
git:
  commit_template: "[{action}] {filename} - {user} at {timestamp}"
```

3. **Regular Syncing**:
```bash
# Create sync script
#!/bin/bash
git pull
gitcells sync
git push
```

## Advanced Git Features

### Git Hooks Integration

Create custom Git hooks for Excel files:

#### Pre-commit Hook
`.git/hooks/pre-commit`:
```bash
#!/bin/bash
# Validate Excel files before commit

for file in $(git diff --cached --name-only | grep '\.xlsx$'); do
  # Check file size
  size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file")
  if [ $size -gt 10485760 ]; then  # 10MB
    echo "Error: $file is too large (>10MB)"
    exit 1
  fi
  
  # Ensure JSON chunks exist
  chunk_dir=".gitcells/data/${file}_chunks"
  if [ ! -d "$chunk_dir" ]; then
    echo "Converting $file to JSON chunks..."
    gitcells convert "$file"
    git add "$chunk_dir/"
  fi
done
```

#### Post-merge Hook
`.git/hooks/post-merge`:
```bash
#!/bin/bash
# Regenerate Excel files after merge

for chunk_dir in $(find .gitcells/data -name "*_chunks" -newer .git/MERGE_HEAD); do
  excel_name=$(basename "$chunk_dir" | sed 's/_chunks$//')
  echo "Regenerating $excel_name from JSON chunks..."
  gitcells convert "$chunk_dir" -o "$excel_name"
done
```

### Git Attributes

Configure Git attributes for Excel files:

`.gitattributes`:
```
# Excel files
*.xlsx binary
*.xls binary
*.xlsm binary

# JSON chunk files - ensure LF line endings
.gitcells/data/**/*.json text eol=lf

# Mark generated files
.gitcells/data/** linguist-generated=true
```

### Git LFS for Large Files

For large Excel files, use Git LFS:

1. Install Git LFS:
```bash
git lfs install
```

2. Track Excel files:
```bash
git lfs track "*.xlsx"
git lfs track "*.xlsm"
```

3. Configure GitCells to handle LFS:
```yaml
converter:
  max_file_size: "50MB"  # Larger limit for LFS
```

## Viewing History

### Git Commands for Excel Files

View Excel file history:
```bash
# See all changes to an Excel file
git log --follow Budget.xlsx

# See what changed in each commit
git log -p .gitcells/data/Budget.xlsx_chunks/

# View specific version (requires checking out chunks)
git checkout HEAD~3 -- .gitcells/data/Budget.xlsx_chunks/
gitcells convert .gitcells/data/Budget.xlsx_chunks/ -o OldBudget.xlsx
git checkout HEAD -- .gitcells/data/Budget.xlsx_chunks/
```

### GitCells Status Command

Check current status:
```bash
gitcells status

# Output:
# Repository Status: Clean
# Tracked Excel Files: 12
# Recent Changes:
#   - Budget2024.xlsx (modified 2 hours ago)
#   - Report.xlsx (modified yesterday)
```

### Diff Command

Compare Excel file versions:
```bash
# Compare with last commit
gitcells diff Budget.xlsx

# Compare two versions
gitcells diff --from HEAD~3 --to HEAD Budget.xlsx
```

## Workflows

### Individual Workflow

```bash
# 1. Initialize
cd my-excel-files
git init
gitcells init

# 2. Configure
# Edit .gitcells.yaml

# 3. Start tracking
gitcells watch .

# 4. Work normally
# Edit Excel files as usual
# GitCells auto-commits changes

# 5. View history
git log --oneline
gitcells status
```

### Team Workflow

```bash
# 1. Setup shared repository
git clone https://github.com/team/excel-files.git
cd excel-files
gitcells init

# 2. Configure for team
# Edit .gitcells.yaml with team settings

# 3. Start collaborating
gitcells watch .

# 4. Sync regularly
git pull
git push
```

### CI/CD Workflow

```yaml
# .github/workflows/excel-validation.yml
name: Validate Excel Files

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Install GitCells
        run: |
          curl -L $GITCELLS_URL -o gitcells
          chmod +x gitcells
          sudo mv gitcells /usr/local/bin/
      
      - name: Validate Excel files
        run: |
          for file in $(find . -name "*.xlsx"); do
            gitcells convert "$file" --validate
          done
```

## Security Considerations

### Sensitive Data

Protect sensitive Excel files:

1. **Use .gitignore**:
```
# .gitignore
confidential/
*_private.xlsx
*_secret.xlsx
```

2. **Encrypt repositories**:
```bash
# Use git-crypt
git-crypt init
git-crypt add-gpg-user YOUR_GPG_ID
```

3. **Separate repositories**:
```yaml
# public-files/.gitcells.yaml
git:
  remote: https://github.com/company/public-excel

# private-files/.gitcells.yaml  
git:
  remote: https://private-git.company.com/sensitive-excel
```

### Access Control

- Use Git repository permissions
- Configure branch protection rules
- Set up code review requirements
- Enable audit logging

## Troubleshooting

### Common Git Issues

1. **"Not a git repository" error**:
```bash
git init
```

2. **"No remote configured" error**:
```bash
git remote add origin YOUR_REMOTE_URL
```

3. **Authentication failures**:
```bash
# HTTPS
git config credential.helper store

# SSH
ssh-keygen -t rsa -b 4096
```

4. **Merge conflicts**:
- Pull latest changes first
- Resolve conflicts in JSON files
- Convert back to Excel
- Test the merged file

## Next Steps

- Explore the [Terminal UI](tui.md) for visual Git status
- Learn about [Troubleshooting](troubleshooting.md) Git issues
- Read about [JSON Format](../reference/json-format.md) for understanding conflicts
- Check [Command Reference](../reference/commands.md) for Git-related commands