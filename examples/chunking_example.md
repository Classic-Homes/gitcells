# GitCells Chunking Example

This example demonstrates how GitCells' built-in sheet-based chunking improves file management and Git performance.

## Overview

GitCells automatically splits Excel files into multiple JSON files:
- One main `workbook.json` file containing metadata and workbook-level information
- Separate `sheet_<name>.json` files for each sheet's data
- A `.gitcells_chunks.json` metadata file to track all chunks

## Benefits

1. **Better Git Performance**: Smaller files mean faster diffs and commits
2. **Improved Readability**: Each sheet can be reviewed independently
3. **Reduced Merge Conflicts**: Changes to different sheets don't conflict
4. **Memory Efficiency**: Large workbooks can be processed sheet by sheet

## Usage

### Convert Excel to JSON

```bash
# Convert Excel to chunked JSON files (automatic)
gitcells convert myworkbook.xlsx

# This creates:
# .gitcells/data/myworkbook_chunks/
#   ├── workbook.json           # Main metadata file
#   ├── sheet_Sheet1.json       # First sheet data
#   ├── sheet_Sheet2.json       # Second sheet data
#   └── .gitcells_chunks.json   # Chunk metadata
```

### Configuration Options

You can configure the chunking strategy in `.gitcells.yaml`:

```yaml
converter:
  chunking_strategy: "sheet-based"  # Default, currently only option
```

### Convert Back to Excel

```bash
# GitCells automatically finds the JSON chunks based on the original Excel path
gitcells convert myworkbook.xlsx --output reconstructed.xlsx

# Or reference the chunk directory directly
gitcells convert .gitcells/data/myworkbook_chunks --output reconstructed.xlsx
```

## Example Structure

For an Excel file with 3 sheets (Sales, Inventory, Reports):

```
project/
├── data.xlsx                    # Original Excel file (in working directory)
├── .gitcells/
│   └── data/
│       └── data_chunks/         # Chunked JSON output
│           ├── workbook.json    # Contains:
│           │                    # - Document metadata
│           │                    # - Sheet list and properties
│           │                    # - Defined names
│           │                    # - Document properties
│           ├── sheet_Sales.json        # Sales sheet data only
│           ├── sheet_Inventory.json    # Inventory sheet data only
│           ├── sheet_Reports.json      # Reports sheet data only
│           └── .gitcells_chunks.json   # Chunking metadata
└── .gitignore                   # Excludes *.xlsx files
```

## Future Enhancements

The architecture supports future hybrid chunking strategies:
- Automatic detection of large sheets
- Row-based chunking for sheets exceeding size thresholds
- Configurable chunk size limits
- Smart chunking based on data patterns

## Best Practices

1. **Multi-Sheet Workbooks**: Chunking is especially beneficial for workbooks with many sheets
2. **Version Control**: Automatic chunking makes Git operations significantly more efficient
3. **Collaborative Workflows**: Team members can work on different sheets without conflicts

## Performance Considerations

- Minimal overhead during conversion
- Optimized reading - loads only needed sheets on demand
- Dramatically faster Git operations with smaller, focused files
- Perfect for collaborative workflows where different team members work on different sheets
- Reduces memory usage for large workbooks by processing sheets independently