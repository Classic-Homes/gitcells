# Basic Concepts

Understanding how GitCells works will help you get the most out of it. This guide explains the key concepts in simple terms.

## The Problem GitCells Solves

Excel files are binary files, which means:
- Git can't show you what changed inside the file
- Multiple people editing the same file causes conflicts
- File history takes up lots of space
- You can't search for specific changes across versions

GitCells solves these problems by converting Excel files to a text format that Git understands.

## How GitCells Works

### 1. The Conversion Process

Think of GitCells as a translator between Excel and Git:

```
Excel File (.xlsx) → GitCells → JSON Chunks (.gitcells/data/) → Git
```

- **Excel File**: Your normal spreadsheet file
- **GitCells**: The translator that converts between formats
- **JSON Chunks**: Text versions of your spreadsheet (split by sheets) that Git can track
- **Git**: The version control system that tracks changes

### 2. What Gets Preserved

GitCells preserves everything important from your Excel files:

- **Cell Values**: All your data (numbers, text, dates)
- **Formulas**: Including complex formulas and references
- **Formatting**: Bold, italic, colors, borders, etc.
- **Comments**: Cell comments and notes
- **Structure**: Merged cells, hidden rows/columns
- **Charts**: Chart definitions and data
- **Pivot Tables**: Complete pivot table configurations

### 3. The JSON Format

JSON is a human-readable text format. Here's a simple example:

```json
{
  "sheets": [{
    "name": "Sales Data",
    "cells": {
      "A1": {
        "value": "Month",
        "style": { "bold": true }
      },
      "B1": {
        "value": "Revenue",
        "style": { "bold": true }
      },
      "B2": {
        "value": 5000,
        "formula": "=SUM(C2:E2)"
      }
    }
  }]
}
```

You can actually read and understand what's in your spreadsheet!

## Key Concepts

### File Watching

GitCells can watch folders for changes:
- You edit and save an Excel file
- GitCells detects the change immediately
- It automatically converts the file to JSON
- The JSON chunk files are saved in `.gitcells/data` directory

### Bidirectional Conversion

GitCells works both ways:
- **Excel → JSON**: For tracking changes in Git
- **JSON → Excel**: To restore or share spreadsheets

This means you can always get back to a regular Excel file.

### Git Integration

When integrated with Git, GitCells can:
- Automatically commit changes when you save Excel files
- Show detailed differences between versions
- Help merge changes from multiple people
- Maintain a complete history of all changes

### The .gitcells.yaml Configuration

This file tells GitCells how to behave:
- Which folders to watch
- What files to ignore
- How to format commit messages
- Whether to auto-commit changes

## Workflow Examples

### Solo User Workflow

1. Initialize GitCells in your Excel folder
2. Start the file watcher
3. Edit your Excel files normally
4. GitCells automatically tracks all changes
5. View history anytime with `gitcells status`

### Team Workflow

1. Set up a Git repository for the team
2. Everyone installs GitCells
3. Team members clone the repository
4. Each person runs `gitcells watch`
5. Changes are automatically tracked and can be shared
6. Git shows who changed what and when

### Manual Conversion Workflow

Sometimes you just want to convert files:
1. Convert Excel to JSON to see the content
2. Edit the JSON chunk files directly (advanced users)
3. Convert back to Excel
4. Share the file with others

## File Structure

After using GitCells, your folder might look like this:

```
MyExcelFiles/
├── .gitcells.yaml          # GitCells configuration
├── .git/                   # Git repository (if using Git)
├── .gitcells/
│   └── data/               # Centralized JSON storage
│       ├── Budget2024.xlsx_chunks/
│       │   ├── workbook.json
│       │   ├── sheet_Sheet1.json
│       │   └── .gitcells_chunks.json
│       └── Sales.xlsx_chunks/
│           ├── workbook.json
│           ├── sheet_Data.json
│           └── .gitcells_chunks.json
├── Budget2024.xlsx         # Your Excel file
└── Sales.xlsx              # Another Excel file
```

## Important Notes

### Your Excel Files Are Safe

- GitCells never modifies your original Excel files
- It only creates additional JSON chunk files
- You can delete JSON chunk files anytime (though you'll lose history)
- Excel files work normally with or without GitCells

### Storage Considerations

- JSON chunk files combined are usually larger than Excel files
- But Git compresses them efficiently
- Overall repository size is manageable
- Old versions are compressed even more

### Performance

- Small to medium Excel files: Instant conversion
- Large Excel files (>10MB): May take a few seconds
- Very large files (>100MB): Consider splitting them

## Common Terms

- **Repository (Repo)**: A folder tracked by Git
- **Commit**: A saved snapshot of your files
- **JSON**: JavaScript Object Notation - a text format for data
- **Binary File**: Files that aren't human-readable (like Excel files)
- **Text File**: Files you can read in a text editor
- **Watcher**: The GitCells component that monitors file changes

## Next Steps

Now that you understand the basics:
- Follow the [Quick Start Guide](quickstart.md) to try GitCells
- Learn about [Configuration Options](../user-guide/configuration.md)
- Explore [Git Integration](../user-guide/git-integration.md) for team collaboration