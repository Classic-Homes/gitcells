# Converting Excel Files

Learn how to convert between Excel and JSON formats using GitCells.

## Basic Conversion

### Excel to JSON

Convert a single file:
```bash
gitcells convert report.xlsx
```

Convert multiple files:
```bash
gitcells convert *.xlsx
```

Convert with options:
```bash
# Pretty print JSON
gitcells convert report.xlsx --pretty

# Custom output location
gitcells convert report.xlsx --output data/report.json

# Force overwrite
gitcells convert report.xlsx --force
```

### JSON to Excel

Reverse conversion happens automatically with sync:
```bash
# After pulling changes
git pull
gitcells sync
```

Or explicitly:
```bash
gitcells sync report.json
```

## Batch Operations

### Convert Directory

Convert all Excel files in a directory:
```bash
gitcells convert --recursive reports/
```

### Pattern Matching

Use glob patterns:
```bash
# All 2024 reports
gitcells convert *2024*.xlsx

# Specific sheets
gitcells convert finance/*.xlsx
```

## Conversion Options

### Pretty Printing

Make JSON human-readable:
```bash
gitcells convert report.xlsx --pretty
```

Result:
```json
{
  "metadata": {
    "created": "2024-01-15T10:30:00Z",
    "gitcells_version": "1.0.0"
  },
  "sheets": [
    {
      "name": "Summary",
      "cells": {
        "A1": {
          "value": "Total Revenue",
          "type": "string"
        }
      }
    }
  ]
}
```

### Selective Conversion

Convert specific sheets only:
```bash
gitcells convert report.xlsx --sheets "Summary,Details"
```

Skip hidden sheets:
```bash
gitcells convert report.xlsx --skip-hidden
```

## Understanding the JSON Format

### Structure Overview

```json
{
  "metadata": {
    "created": "timestamp",
    "modified": "timestamp",
    "gitcells_version": "version"
  },
  "properties": {
    "title": "Document Title",
    "author": "Author Name"
  },
  "sheets": [
    {
      "name": "Sheet1",
      "cells": {},
      "merges": [],
      "formatting": {}
    }
  ]
}
```

### Cell Representation

Each cell contains:
```json
{
  "A1": {
    "value": 100,
    "type": "number",
    "formula": "=SUM(B1:B10)",
    "format": "0.00",
    "style": {
      "font": "Arial",
      "size": 12,
      "bold": true
    }
  }
}
```

## Advanced Scenarios

### Large Files

For files over 10MB:
```bash
# Use streaming mode
gitcells convert large-file.xlsx --stream

# Compress JSON output
gitcells convert large-file.xlsx --compress
```

### Formula Preservation

Formulas are preserved exactly:

Excel: `=VLOOKUP(A2,Sheet2!A:B,2,FALSE)`

JSON:
```json
{
  "formula": "=VLOOKUP(A2,Sheet2!A:B,2,FALSE)",
  "value": "Result"
}
```

### Maintaining Links

External links are preserved:
```json
{
  "formula": "='[Budget.xlsx]Summary'!A1",
  "value": 50000,
  "external_link": true
}
```

## Validation

### Check Conversion

Verify conversion accuracy:
```bash
gitcells validate report.xlsx report.json
```

### Round-trip Test

Test Excel → JSON → Excel:
```bash
gitcells test-roundtrip report.xlsx
```

## Performance Tips

### 1. Use Batch Mode

Instead of:
```bash
gitcells convert file1.xlsx
gitcells convert file2.xlsx
gitcells convert file3.xlsx
```

Do:
```bash
gitcells convert *.xlsx
```

### 2. Skip Unnecessary Data

```bash
# Skip empty cells
gitcells convert report.xlsx --skip-empty

# Skip formatting
gitcells convert report.xlsx --data-only
```

### 3. Parallel Processing

For many files:
```bash
gitcells convert *.xlsx --parallel 4
```

## Troubleshooting

### Common Issues

**Protected Files**
```bash
# Provide password
gitcells convert protected.xlsx --password "secret"
```

**Corrupted Files**
```bash
# Try recovery mode
gitcells convert damaged.xlsx --recover
```

**Memory Issues**
```bash
# Use low memory mode
gitcells convert huge.xlsx --low-memory
```

### Debugging

Enable verbose output:
```bash
gitcells convert report.xlsx -v
```

Check conversion log:
```bash
cat .gitcells/logs/conversion.log
```

## Integration Examples

### CI/CD Pipeline

```yaml
# .github/workflows/excel-check.yml
name: Validate Excel Files
on: [push]
jobs:
  validate:
    steps:
      - uses: actions/checkout@v2
      - name: Convert Excel files
        run: gitcells convert *.xlsx
      - name: Check for changes
        run: git diff --exit-code
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit
gitcells convert --check *.xlsx
```

## Next Steps

- Learn about [tracking changes](tracking.md)
- Set up [auto-sync](auto-sync.md)
- Configure [advanced options](../reference/configuration.md)