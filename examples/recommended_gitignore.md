# Recommended .gitignore for GitCells Projects

When using GitCells, you'll want to configure your `.gitignore` file to:
1. **Ignore Excel files** (since they're binary and tracked via JSON)
2. **Track the JSON representations** in `.gitcells/data/`
3. **Ignore GitCells cache and temporary files**

## Recommended .gitignore entries

```gitignore
# Excel files (tracked via GitCells JSON format)
*.xlsx
*.xls
*.xlsm
*.xlsb

# GitCells cache and logs
.gitcells/cache/
.gitcells/logs/
.gitcells/tmp/

# GitCells lock files
.gitcells/*.lock

# Excel temporary files
~$*.xlsx
~$*.xls
~$*.xlsm

# OS-specific files
.DS_Store
Thumbs.db

# Keep GitCells data (JSON representations)
!.gitcells/data/
```

## Directory Structure Example

With this setup, your repository structure will look like:

```
project/
├── .git/
├── .gitignore
├── .gitcells.yml           # GitCells configuration
├── .gitcells/
│   ├── data/               # JSON representations (tracked)
│   │   ├── reports/
│   │   │   └── sales_2024_chunks/
│   │   │       ├── workbook.json
│   │   │       ├── sheet_Q1.json
│   │   │       └── sheet_Q2.json
│   │   └── budgets/
│   │       └── budget_2024_chunks/
│   │           ├── workbook.json
│   │           └── sheet_Annual.json
│   ├── cache/              # Temporary cache (ignored)
│   └── logs/               # Log files (ignored)
├── reports/
│   └── sales_2024.xlsx    # Original Excel (ignored)
└── budgets/
    └── budget_2024.xlsx    # Original Excel (ignored)
```

## Benefits

1. **Clean Working Directory**: Excel files stay where users expect them
2. **No Git Conflicts**: Binary Excel files aren't tracked
3. **Clear Separation**: JSON data is organized in `.gitcells/data/`
4. **Easy Recovery**: Can regenerate Excel files from JSON anytime

## Setup Commands

```bash
# Initialize GitCells
gitcells init

# Add recommended .gitignore entries
cat >> .gitignore << 'EOF'

# Excel files (tracked via GitCells)
*.xlsx
*.xls
*.xlsm
~$*.xlsx
~$*.xls

# GitCells temporary files
.gitcells/cache/
.gitcells/logs/
.gitcells/*.lock

# Keep GitCells data
!.gitcells/data/
EOF

# Convert existing Excel files
gitcells convert reports/sales_2024.xlsx
gitcells convert budgets/budget_2024.xlsx

# Commit the JSON representations
git add .gitcells/data/
git commit -m "Add GitCells JSON representations"
```

## Important Notes

1. **First Time Setup**: After adding `.gitignore`, you may need to untrack Excel files:
   ```bash
   git rm --cached *.xlsx *.xls *.xlsm
   git commit -m "Stop tracking Excel files directly"
   ```

2. **Team Collaboration**: Make sure all team members:
   - Have GitCells installed
   - Understand to commit JSON files, not Excel files
   - Know how to regenerate Excel files from JSON

3. **CI/CD Integration**: Your build scripts can regenerate Excel files:
   ```bash
   # Regenerate all Excel files
   find .gitcells/data -name "*_chunks" -type d | while read dir; do
     gitcells convert "$dir"
   done
   ```