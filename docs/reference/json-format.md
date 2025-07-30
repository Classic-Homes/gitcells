# JSON Format Specification

Detailed specification of the GitCells JSON format for Excel files.

## Overview

GitCells converts Excel files to a structured JSON format that preserves all data, formulas, formatting, and metadata while being Git-friendly and human-readable.

## Format Version

Current format version: **1.0**

```json
{
  "gitcells_version": "1.0",
  "metadata": {
    "format_version": "1.0"
  }
}
```

## Top-Level Structure

```json
{
  "gitcells_version": "1.0",
  "metadata": {},
  "properties": {},
  "sheets": [],
  "defined_names": {},
  "styles": {},
  "themes": {},
  "custom_xml": {}
}
```

## Metadata Section

Contains conversion and file metadata:

```json
{
  "metadata": {
    "format_version": "1.0",
    "created": "2024-01-15T10:30:00Z",
    "modified": "2024-01-15T14:45:00Z",
    "converted_at": "2024-01-15T15:00:00Z",
    "converter_version": "gitcells-1.5.0",
    "source_file": {
      "name": "report.xlsx",
      "size": 524288,
      "hash": "sha256:abc123..."
    },
    "excel_version": "16.0",
    "features_used": [
      "formulas",
      "conditional_formatting",
      "charts",
      "pivot_tables"
    ]
  }
}
```

## Properties Section

Document-level properties:

```json
{
  "properties": {
    "title": "Q4 Financial Report",
    "subject": "Quarterly Financials",
    "author": "John Doe",
    "manager": "Jane Smith",
    "company": "Acme Corp",
    "category": "Finance",
    "keywords": ["finance", "quarterly", "report"],
    "comments": "Final version for board review",
    "last_modified_by": "Jane Smith",
    "revision": 15,
    "version": "1.0",
    "created_date": "2024-01-01T09:00:00Z",
    "modified_date": "2024-01-15T14:45:00Z",
    "custom_properties": {
      "department": "Finance",
      "fiscal_year": "2024",
      "confidentiality": "Internal"
    }
  }
}
```

## Sheets Section

Array of sheet objects:

```json
{
  "sheets": [
    {
      "name": "Summary",
      "sheet_id": 1,
      "state": "visible",
      "type": "worksheet",
      "properties": {},
      "cells": {},
      "merges": [],
      "conditional_formatting": [],
      "data_validations": [],
      "charts": [],
      "images": [],
      "comments": {},
      "dimensions": {}
    }
  ]
}
```

### Sheet Properties

```json
{
  "properties": {
    "tab_color": "#FF0000",
    "zoom": 100,
    "gridlines": true,
    "headings": true,
    "protection": {
      "protected": true,
      "password": "hashed:...",
      "options": {
        "select_locked_cells": true,
        "select_unlocked_cells": true,
        "format_cells": false,
        "insert_columns": false,
        "insert_rows": false
      }
    },
    "print_setup": {
      "orientation": "landscape",
      "paper_size": "A4",
      "margins": {
        "top": 0.75,
        "bottom": 0.75,
        "left": 0.7,
        "right": 0.7,
        "header": 0.3,
        "footer": 0.3
      }
    }
  }
}
```

### Cell Representation

Cells are stored in a dictionary with cell address as key:

```json
{
  "cells": {
    "A1": {
      "value": "Revenue",
      "type": "string",
      "style": "s1",
      "hyperlink": null,
      "comment": null
    },
    "B1": {
      "value": 150000,
      "type": "number",
      "formula": "=SUM(B2:B10)",
      "style": "s2",
      "number_format": "#,##0.00"
    },
    "C1": {
      "value": "2024-01-15",
      "type": "date",
      "style": "s3",
      "number_format": "yyyy-mm-dd"
    }
  }
}
```

### Cell Types

Supported cell types:

1. **string**: Text values
2. **number**: Numeric values (integer or float)
3. **boolean**: True/False values
4. **date**: Date/time values (ISO 8601 format)
5. **error**: Error values (#DIV/0!, #N/A, etc.)
6. **formula**: Cells with formulas

### Cell Properties

Complete cell object:

```json
{
  "B5": {
    "value": 42.5,
    "type": "number",
    "formula": "=A5*1.5",
    "style": "s10",
    "number_format": "0.00",
    "hyperlink": {
      "type": "url",
      "target": "https://example.com",
      "tooltip": "Click for details"
    },
    "comment": {
      "text": "Verified by accounting",
      "author": "John Doe",
      "timestamp": "2024-01-15T10:00:00Z",
      "visible": false
    },
    "validation": {
      "type": "list",
      "formula1": "Valid,Invalid,Pending",
      "allow_blank": true,
      "show_dropdown": true,
      "error_title": "Invalid Entry",
      "error_message": "Please select from the list"
    }
  }
}
```

### Merged Cells

```json
{
  "merges": [
    {
      "range": "A1:C1",
      "value_cell": "A1"
    },
    {
      "range": "D5:F8",
      "value_cell": "D5"
    }
  ]
}
```

### Conditional Formatting

```json
{
  "conditional_formatting": [
    {
      "range": "B2:B100",
      "rules": [
        {
          "type": "cell_value",
          "operator": "greater_than",
          "value": 1000,
          "format": {
            "fill": {
              "type": "solid",
              "color": "#00FF00"
            },
            "font": {
              "bold": true
            }
          }
        }
      ]
    }
  ]
}
```

## Styles Section

Centralized style definitions:

```json
{
  "styles": {
    "s1": {
      "font": {
        "name": "Calibri",
        "size": 11,
        "bold": true,
        "italic": false,
        "underline": "none",
        "color": "#000000"
      },
      "fill": {
        "type": "solid",
        "color": "#FFFF00"
      },
      "borders": {
        "top": {"style": "thin", "color": "#000000"},
        "bottom": {"style": "thin", "color": "#000000"},
        "left": {"style": "thin", "color": "#000000"},
        "right": {"style": "thin", "color": "#000000"}
      },
      "alignment": {
        "horizontal": "center",
        "vertical": "middle",
        "wrap_text": true,
        "indent": 0,
        "rotation": 0
      }
    }
  }
}
```

## Charts Section

Chart definitions within sheets:

```json
{
  "charts": [
    {
      "id": "chart1",
      "type": "column",
      "title": "Revenue by Quarter",
      "position": {
        "from": {"col": 5, "row": 2},
        "to": {"col": 12, "row": 15}
      },
      "series": [
        {
          "name": "Revenue",
          "categories": "A2:A5",
          "values": "B2:B5"
        }
      ],
      "axes": {
        "primary_category": {
          "title": "Quarter"
        },
        "primary_value": {
          "title": "Revenue ($)",
          "min": 0,
          "max": 200000
        }
      }
    }
  ]
}
```

## Defined Names

Named ranges and formulas:

```json
{
  "defined_names": {
    "Revenue_Total": {
      "scope": "workbook",
      "reference": "Summary!$B$10",
      "comment": "Total revenue for the year"
    },
    "Tax_Rate": {
      "scope": "workbook",
      "reference": "0.25",
      "comment": "Corporate tax rate"
    },
    "Quarterly_Data": {
      "scope": "Summary",
      "reference": "Summary!$A$1:$D$10"
    }
  }
}
```

## Data Validation

```json
{
  "data_validations": [
    {
      "range": "D2:D100",
      "type": "list",
      "formula1": "$Z$1:$Z$10",
      "allow_blank": true,
      "show_dropdown": true,
      "show_input_message": true,
      "input_title": "Select Status",
      "input_message": "Choose from the list",
      "show_error_message": true,
      "error_style": "stop",
      "error_title": "Invalid Entry",
      "error_message": "Please select a valid status"
    }
  ]
}
```

## Images and Objects

```json
{
  "images": [
    {
      "id": "img1",
      "name": "Company Logo",
      "position": {
        "from": {"col": 0, "row": 0, "col_offset": 0, "row_offset": 0},
        "to": {"col": 2, "row": 3, "col_offset": 0, "row_offset": 0}
      },
      "data": "base64:iVBORw0KGgo...",
      "mime_type": "image/png",
      "alt_text": "Acme Corp Logo"
    }
  ]
}
```

## Pivot Tables

```json
{
  "pivot_tables": [
    {
      "name": "Sales Summary",
      "source_range": "Data!A1:F1000",
      "location": "G5",
      "rows": ["Region", "Product"],
      "columns": ["Quarter"],
      "values": [
        {
          "field": "Revenue",
          "function": "sum",
          "name": "Total Revenue"
        }
      ],
      "filters": ["Year"],
      "style": "PivotStyleLight16"
    }
  ]
}
```

## Formula Representation

### Basic Formulas

```json
{
  "formula": "=A1+B1",
  "formula": "=SUM(A1:A10)",
  "formula": "=IF(A1>100,\"High\",\"Low\")"
}
```

### Array Formulas

```json
{
  "formula": "=SUM(A1:A10*B1:B10)",
  "array_formula": true,
  "array_range": "C1:C1"
}
```

### External References

```json
{
  "formula": "='[Budget.xlsx]Summary'!A1",
  "external_references": [
    {
      "workbook": "Budget.xlsx",
      "sheet": "Summary",
      "range": "A1"
    }
  ]
}
```

## Compression

For large files, the JSON can be compressed:

```json
{
  "compression": {
    "enabled": true,
    "algorithm": "gzip",
    "cells": "compressed:H4sIAAAAAAAA..."
  }
}
```

## Extensibility

Custom extensions can be added:

```json
{
  "extensions": {
    "com.company.custom": {
      "version": "1.0",
      "data": {}
    }
  }
}
```

## Best Practices

1. **Minimize Size**: Omit default values
2. **Preserve Precision**: Use full numeric precision
3. **Maintain Order**: Keep sheet order
4. **Handle Special Characters**: Properly escape JSON
5. **Version Compatibility**: Include format version

## Migration

### From Version 0.9

```json
{
  "migration": {
    "from_version": "0.9",
    "to_version": "1.0",
    "changes": [
      "Renamed 'formatting' to 'conditional_formatting'",
      "Added 'themes' section",
      "Changed date format to ISO 8601"
    ]
  }
}
```

## Validation

Validate JSON against schema:

```bash
gitcells validate-json file.json --schema
```

Schema available at: https://gitcells.com/schema/v1.0/gitcells.json

## Next Steps

- Learn about [conversion process](../guides/converting.md)
- Review [configuration options](configuration.md)
- Understand [troubleshooting](troubleshooting.md)