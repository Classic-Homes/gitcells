# JSON Format Reference

This reference documents the JSON format used by GitCells to represent Excel files. Understanding this format helps with debugging, custom processing, and conflict resolution.

## Format Overview

GitCells converts Excel files to a structured JSON format that preserves all important spreadsheet features while being human-readable and Git-friendly.

## JSON Structure

### Root Object

```json
{
  "version": "1.0",
  "metadata": { ... },
  "properties": { ... },
  "sheets": [ ... ],
  "defined_names": { ... },
  "vba_project": { ... }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | Yes | Format version (currently "1.0") |
| `metadata` | object | Yes | File metadata |
| `properties` | object | No | Document properties |
| `sheets` | array | Yes | Array of sheet objects |
| `defined_names` | object | No | Named ranges |
| `vba_project` | object | No | VBA macro information |

### Metadata Object

Contains information about the file and conversion process.

```json
{
  "metadata": {
    "created": "2024-01-15T10:30:00Z",
    "modified": "2024-01-15T14:45:00Z",
    "app_version": "gitcells-0.3.0",
    "original_file": "Budget2024.xlsx",
    "file_size": 125432,
    "checksum": "sha256:abcdef1234567890",
    "conversion_time": 1.234
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `created` | string | ISO 8601 creation timestamp |
| `modified` | string | ISO 8601 modification timestamp |
| `app_version` | string | GitCells version used |
| `original_file` | string | Original filename |
| `file_size` | integer | Size in bytes |
| `checksum` | string | SHA256 checksum |
| `conversion_time` | number | Conversion duration in seconds |

### Properties Object

Document properties from Excel.

```json
{
  "properties": {
    "title": "Annual Budget 2024",
    "subject": "Financial Planning",
    "author": "John Doe",
    "last_modified_by": "Jane Smith",
    "company": "Acme Corp",
    "category": "Finance",
    "keywords": ["budget", "2024", "finance"],
    "comments": "Approved by board on 2024-01-10",
    "custom": {
      "department": "Finance",
      "project_code": "FIN-2024-001"
    }
  }
}
```

### Sheet Object

Represents a single worksheet.

```json
{
  "name": "Summary",
  "index": 0,
  "visible": true,
  "protection": { ... },
  "cells": { ... },
  "merged_cells": [ ... ],
  "row_heights": { ... },
  "column_widths": { ... },
  "charts": [ ... ],
  "pivot_tables": [ ... ],
  "conditional_formats": [ ... ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Sheet name |
| `index` | integer | Sheet index (0-based) |
| `visible` | boolean | Sheet visibility |
| `protection` | object | Sheet protection settings |
| `cells` | object | Cell data (key: cell reference) |
| `merged_cells` | array | Merged cell ranges |
| `row_heights` | object | Custom row heights |
| `column_widths` | object | Custom column widths |
| `charts` | array | Chart definitions |
| `pivot_tables` | array | Pivot table definitions |
| `conditional_formats` | array | Conditional formatting rules |

### Cell Object

Represents a single cell.

```json
{
  "A1": {
    "value": "Annual Budget",
    "type": "string",
    "formula": "",
    "formula_r1c1": "",
    "array_formula": null,
    "style": { ... },
    "comment": { ... },
    "hyperlink": { ... },
    "data_validation": { ... }
  },
  "B2": {
    "value": 150000,
    "type": "number",
    "formula": "=SUM(B3:B10)",
    "formula_r1c1": "=SUM(R[1]C:R[8]C)",
    "number_format": "$#,##0.00",
    "style": { ... }
  }
}
```

#### Cell Fields

| Field | Type | Description |
|-------|------|-------------|
| `value` | any | Cell value (string, number, boolean, null) |
| `type` | string | Value type: "string", "number", "boolean", "date", "error", "formula" |
| `formula` | string | A1-style formula |
| `formula_r1c1` | string | R1C1-style formula |
| `array_formula` | object | Array formula information |
| `style` | object | Cell styling |
| `comment` | object | Cell comment |
| `hyperlink` | object | Hyperlink information |
| `data_validation` | object | Validation rules |
| `number_format` | string | Number format string |

### Style Object

Cell formatting information.

```json
{
  "style": {
    "font": {
      "name": "Calibri",
      "size": 11,
      "bold": true,
      "italic": false,
      "underline": false,
      "strike": false,
      "color": "#000000"
    },
    "fill": {
      "type": "solid",
      "color": "#FFFF00",
      "pattern": "none"
    },
    "border": {
      "top": { "style": "thin", "color": "#000000" },
      "right": { "style": "thin", "color": "#000000" },
      "bottom": { "style": "thick", "color": "#000000" },
      "left": { "style": "thin", "color": "#000000" }
    },
    "alignment": {
      "horizontal": "center",
      "vertical": "middle",
      "wrap_text": true,
      "text_rotation": 0
    }
  }
}
```

### Array Formula Object

For cells with array formulas.

```json
{
  "array_formula": {
    "range": "A1:C3",
    "formula": "{=MMULT(E1:F2,H1:I2)}",
    "is_master": true
  }
}
```

### Comment Object

Cell comments and notes.

```json
{
  "comment": {
    "author": "John Doe",
    "text": "Verify this calculation with finance team",
    "created": "2024-01-15T10:30:00Z",
    "visible": false,
    "width": 200,
    "height": 100
  }
}
```

### Hyperlink Object

```json
{
  "hyperlink": {
    "type": "url",
    "target": "https://example.com/report",
    "tooltip": "View detailed report"
  }
}
```

Types: "url", "email", "file", "cell"

### Data Validation Object

```json
{
  "data_validation": {
    "type": "list",
    "formula1": "=$A$1:$A$10",
    "formula2": "",
    "allow_blank": true,
    "show_dropdown": true,
    "error_title": "Invalid Entry",
    "error_message": "Please select from the list"
  }
}
```

### Merged Cells Array

```json
{
  "merged_cells": [
    { "range": "A1:D1" },
    { "range": "B5:C8" }
  ]
}
```

### Chart Object

```json
{
  "charts": [{
    "name": "Sales Chart",
    "type": "column",
    "position": {
      "from": { "col": 5, "row": 1 },
      "to": { "col": 10, "row": 15 }
    },
    "series": [{
      "name": "Q1 Sales",
      "categories": "Sheet1!$A$2:$A$10",
      "values": "Sheet1!$B$2:$B$10"
    }],
    "title": "Quarterly Sales",
    "legend": { "position": "bottom" }
  }]
}
```

### Pivot Table Object

```json
{
  "pivot_tables": [{
    "name": "SalesPivot",
    "source_range": "A1:D100",
    "target_cell": "F5",
    "rows": ["Region", "Product"],
    "columns": ["Quarter"],
    "values": [{
      "field": "Sales",
      "function": "sum"
    }],
    "filters": ["Year"]
  }]
}
```

## Data Types

### Cell Value Types

| Type | JSON Type | Example | Description |
|------|-----------|---------|-------------|
| string | string | `"Hello"` | Text values |
| number | number | `123.45` | Numeric values |
| boolean | boolean | `true` | TRUE/FALSE |
| date | string | `"2024-01-15T00:00:00Z"` | ISO 8601 dates |
| error | object | `{"error": "#DIV/0!"}` | Excel errors |
| empty | null | `null` | Empty cells |

### Number Formats

Common Excel number formats preserved:

- `"General"` - Default format
- `"0.00"` - Two decimal places
- `"$#,##0.00"` - Currency
- `"0.00%"` - Percentage
- `"mm/dd/yyyy"` - Date
- `"@"` - Text

### Formula Representation

Formulas are stored in both A1 and R1C1 notation:

```json
{
  "formula": "=SUM(A1:A10)",
  "formula_r1c1": "=SUM(R1C1:R10C1)"
}
```

## Working with the JSON

### Reading with jq

```bash
# Get all sheet names
cat file.xlsx.json | jq '.sheets[].name'

# Get value of cell A1 from first sheet
cat file.xlsx.json | jq '.sheets[0].cells.A1.value'

# Find all cells with formulas
cat file.xlsx.json | jq '.sheets[].cells | to_entries[] | select(.value.formula != "")'

# Extract all comments
cat file.xlsx.json | jq '.sheets[].cells | to_entries[] | select(.value.comment != null)'
```

### Python Example

```python
import json

# Load GitCells JSON
with open('Budget.xlsx.json', 'r') as f:
    workbook = json.load(f)

# Access sheet data
for sheet in workbook['sheets']:
    print(f"Sheet: {sheet['name']}")
    
    # Access cells
    for cell_ref, cell_data in sheet['cells'].items():
        value = cell_data.get('value')
        formula = cell_data.get('formula')
        
        if formula:
            print(f"  {cell_ref}: {formula} = {value}")
        else:
            print(f"  {cell_ref}: {value}")
```

### JavaScript Example

```javascript
const fs = require('fs');

// Load GitCells JSON
const workbook = JSON.parse(fs.readFileSync('Budget.xlsx.json', 'utf8'));

// Process sheets
workbook.sheets.forEach(sheet => {
    console.log(`Sheet: ${sheet.name}`);
    
    // Find all cells with values over 1000
    Object.entries(sheet.cells).forEach(([ref, cell]) => {
        if (cell.type === 'number' && cell.value > 1000) {
            console.log(`  ${ref}: ${cell.value}`);
        }
    });
});
```

## Compact vs Pretty Format

### Pretty Format (Default)

```json
{
  "version": "1.0",
  "sheets": [
    {
      "name": "Sheet1",
      "cells": {
        "A1": {
          "value": "Hello",
          "type": "string"
        }
      }
    }
  ]
}
```

### Compact Format

```json
{"version":"1.0","sheets":[{"name":"Sheet1","cells":{"A1":{"value":"Hello","type":"string"}}}]}
```

Enable with: `gitcells convert --compact file.xlsx`

## Handling Large Files

For large Excel files, GitCells may split the JSON into chunks:

### Chunked Structure

```
Budget.xlsx.json          # Main file with metadata
Budget.xlsx.chunk1.json   # First chunk of data
Budget.xlsx.chunk2.json   # Second chunk of data
```

### Main File with Chunks

```json
{
  "version": "1.0",
  "metadata": { ... },
  "chunks": [
    {
      "filename": "Budget.xlsx.chunk1.json",
      "sheets": ["Sheet1", "Sheet2"],
      "size": 5242880
    },
    {
      "filename": "Budget.xlsx.chunk2.json",
      "sheets": ["Sheet3"],
      "size": 3145728
    }
  ]
}
```

## Version Compatibility

### Version 1.0

Current version with full feature support.

### Future Versions

Future versions will maintain backward compatibility. New fields may be added but existing fields won't be removed or changed in breaking ways.

## Best Practices

1. **Preserve Formulas**: Always keep formula information for accurate reconstruction
2. **Use R1C1**: R1C1 notation helps with formula relocation
3. **Include Metadata**: Helps track file origin and conversion details
4. **Validate JSON**: Ensure JSON is valid before committing
5. **Consider Size**: Use compact format for large files

## Next Steps

- See [Converting Files](../user-guide/converting.md) for conversion examples
- Review [Git Integration](../user-guide/git-integration.md) for version control
- Check [API Reference](api.md) for programmatic access
- Read [Troubleshooting](../user-guide/troubleshooting.md) for JSON issues