# Converting Files

GitCells can convert between Excel and JSON formats in both directions. This guide explains how to use the conversion features effectively.

## Basic Conversion

### Excel to JSON

Convert an Excel file to JSON:
```bash
gitcells convert Budget2024.xlsx
```

This creates `Budget2024.xlsx.json` in the same directory.

### JSON to Excel

Convert a JSON file back to Excel:
```bash
gitcells convert Budget2024.xlsx.json
```

This creates `Budget2024.xlsx` (or restores it if it was deleted).

## Conversion Options

### Specify Output File

Use the `-o` flag to specify the output location:
```bash
# Excel to JSON with custom output
gitcells convert Budget.xlsx -o /path/to/output/Budget.json

# JSON to Excel with custom output
gitcells convert Budget.json -o RestoresBudget.xlsx
```

### Conversion Flags

Control what gets preserved during conversion:

```bash
# Preserve everything (default)
gitcells convert file.xlsx

# Minimize file size
gitcells convert file.xlsx --compact --no-preserve-styles

# Custom preservation options
gitcells convert file.xlsx \
  --preserve-formulas \
  --preserve-styles \
  --preserve-comments \
  --compact
```

Available flags:
- `--preserve-formulas` - Keep Excel formulas (default: true)
- `--preserve-styles` - Keep cell formatting (default: true)
- `--preserve-comments` - Keep cell comments (default: true)
- `--compact` - Output compact JSON (default: false)
- `-o, --output` - Specify output file path

## Understanding the Conversion Process

### What Gets Converted

GitCells converts these Excel elements:

1. **Cell Data**
   - Values (text, numbers, dates, booleans)
   - Formulas (with R1C1 notation preserved)
   - Array formulas
   - Data validation rules

2. **Formatting**
   - Font styles (bold, italic, size, color)
   - Cell styles (background, borders)
   - Number formats
   - Conditional formatting

3. **Structure**
   - Multiple sheets
   - Merged cells
   - Row heights and column widths
   - Hidden rows/columns
   - Sheet protection settings

4. **Objects**
   - Charts (definitions and data)
   - Pivot tables
   - Images (as base64)
   - Comments and notes
   - Defined names (named ranges)

### The JSON Structure

Here's what the JSON output looks like:

```json
{
  "version": "1.0",
  "metadata": {
    "created": "2024-01-15T10:30:00Z",
    "modified": "2024-01-15T14:45:00Z",
    "app_version": "gitcells-0.3.0",
    "original_file": "Budget2024.xlsx",
    "file_size": 125432,
    "checksum": "sha256:abc123..."
  },
  "properties": {
    "title": "Annual Budget 2024",
    "author": "John Doe",
    "company": "Acme Corp"
  },
  "sheets": [{
    "name": "Summary",
    "index": 0,
    "cells": {
      "A1": {
        "value": "Annual Budget 2024",
        "type": "string",
        "style": {
          "font": {
            "bold": true,
            "size": 14
          }
        }
      },
      "B2": {
        "value": 150000,
        "type": "number",
        "formula": "=SUM(Details!B:B)",
        "number_format": "$#,##0.00"
      }
    },
    "merged_cells": [{
      "range": "A1:D1"
    }]
  }]
}
```

## Batch Conversion

### Convert Multiple Files

Convert all Excel files in a directory:
```bash
# Using bash
for file in *.xlsx; do
  gitcells convert "$file"
done

# Using find
find . -name "*.xlsx" -exec gitcells convert {} \;
```

### Convert with Pattern Matching

Convert specific files:
```bash
# Convert all budget files
for file in Budget*.xlsx; do
  gitcells convert "$file" -o "json/$file.json"
done
```

## Advanced Conversion

### Handling Large Files

For large Excel files, GitCells uses chunking:

```yaml
# In .gitcells.yaml
converter:
  chunking_strategy: "size-based"
  max_chunk_size: "10MB"
  max_cells_per_sheet: 500000
```

The chunking process:
1. **Sheet-based**: Each sheet becomes a separate chunk
2. **Size-based**: Splits when size limit is reached
3. **Row-based**: Splits after certain number of rows

### Partial Conversion

Convert specific sheets only:
```bash
# Future feature - not yet implemented
gitcells convert file.xlsx --sheets "Summary,Data"
```

### Performance Options

Speed up conversion for large files:

```bash
# Compact JSON (smaller, faster)
gitcells convert large.xlsx --compact

# Skip empty cells
gitcells convert sparse.xlsx --ignore-empty
```

## Use Cases

### 1. Version Control

Convert before committing to Git:
```bash
# Convert to JSON
gitcells convert Report.xlsx

# Add both files to Git
git add Report.xlsx Report.xlsx.json
git commit -m "Updated report with Q4 data"
```

### 2. Data Recovery

Recover from JSON if Excel file is corrupted:
```bash
# Excel file corrupted? Restore from JSON
gitcells convert Report.xlsx.json -o Report_Restored.xlsx
```

### 3. Automation

Script conversions for automated workflows:
```bash
#!/bin/bash
# Convert all Excel files nightly

EXCEL_DIR="/path/to/excel/files"
JSON_DIR="/path/to/json/backup"

for file in "$EXCEL_DIR"/*.xlsx; do
  basename=$(basename "$file")
  gitcells convert "$file" -o "$JSON_DIR/$basename.json"
done
```

### 4. Data Analysis

Convert to JSON for processing with other tools:
```bash
# Convert to JSON
gitcells convert data.xlsx

# Process with jq
cat data.xlsx.json | jq '.sheets[0].cells | length'

# Extract specific data
cat data.xlsx.json | jq '.sheets[0].cells.A1.value'
```

## Validation and Testing

### Verify Conversion Integrity

Test round-trip conversion:
```bash
# Original to JSON
gitcells convert Original.xlsx -o Test1.json

# JSON back to Excel  
gitcells convert Test1.json -o Test2.xlsx

# Compare checksums
shasum Original.xlsx Test2.xlsx
```

### Check JSON Output

Validate JSON structure:
```bash
# Pretty print for inspection
cat file.json | jq '.' | less

# Validate JSON syntax
cat file.json | jq empty && echo "Valid JSON"
```

## Troubleshooting Conversion

### Common Issues

1. **"File too large" error**
   - Enable chunking in configuration
   - Increase memory limits
   - Split the Excel file

2. **"Unsupported format" error**
   - Ensure file is a valid Excel file
   - Check file extension matches content
   - Try opening in Excel first

3. **Missing formulas in JSON**
   - Check `--preserve-formulas` flag
   - Verify formulas aren't corrupted
   - Some complex formulas may need manual review

4. **Styles not preserved**
   - Ensure `--preserve-styles` is enabled
   - Some exotic styles may not convert
   - Check JSON for style data

### Performance Tips

1. **For faster conversion**:
   - Use `--compact` flag
   - Disable style preservation if not needed
   - Process files in parallel

2. **For smaller files**:
   - Use `--compact` flag
   - Enable `ignore_empty_cells`
   - Disable unnecessary preservation options

3. **For accuracy**:
   - Keep all preservation options enabled
   - Use verbose mode to see warnings
   - Verify critical formulas after conversion

## Integration Examples

### With Git Hooks

Create a pre-commit hook (`.git/hooks/pre-commit`):
```bash
#!/bin/bash
# Auto-convert Excel files before commit

for file in $(git diff --cached --name-only | grep -E '\.(xlsx|xls)$'); do
  gitcells convert "$file"
  git add "$file.json"
done
```

### With CI/CD

In your CI pipeline:
```yaml
# GitHub Actions example
- name: Convert Excel files
  run: |
    for file in $(find . -name "*.xlsx"); do
      gitcells convert "$file"
    done
```

### With File Watchers

Combine with system file watchers:
```bash
# Using fswatch (macOS)
fswatch -o *.xlsx | xargs -n1 -I{} gitcells convert {}

# Using inotifywait (Linux)
inotifywait -m -e close_write *.xlsx |
while read path action file; do
  gitcells convert "$file"
done
```

## Next Steps

- Set up [File Watching](watching.md) for automatic conversion
- Learn about [Git Integration](git-integration.md)
- Explore the [JSON Format](../reference/json-format.md) in detail
- Check [Troubleshooting](troubleshooting.md) for more solutions