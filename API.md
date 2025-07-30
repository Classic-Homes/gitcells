# GitCells API Documentation

This document describes the internal API architecture and data structures used by GitCells.

## Architecture Overview

GitCells follows a clean architecture pattern with clear separation of concerns:

```
cmd/gitcells/          # CLI commands and entry points
├── main.go            # Application entry point
├── init.go            # Initialize command
├── convert.go         # Convert command
├── watch.go           # Watch command  
├── sync.go            # Sync command
└── status.go          # Status command

internal/              # Private application code
├── config/            # Configuration management
├── converter/         # Excel ↔ JSON conversion
├── git/               # Git operations
├── watcher/           # File system monitoring
└── utils/             # Utilities and helpers

pkg/                   # Public API packages
└── models/            # Data models and types
```

## Core Data Models

### ExcelDocument

The `ExcelDocument` represents the complete structure of an Excel file in JSON format:

```go
type ExcelDocument struct {
    Version      string            `json:"version"`
    Metadata     DocumentMetadata  `json:"metadata"`
    Sheets       []Sheet           `json:"sheets"`
    DefinedNames map[string]string `json:"defined_names,omitempty"`
    Properties   DocumentProperties `json:"properties,omitempty"`
}
```

**Fields:**
- `Version`: GitCells format version (e.g., "1.0")
- `Metadata`: File metadata including checksums and timestamps
- `Sheets`: Array of worksheet data
- `DefinedNames`: Excel named ranges and definitions
- `Properties`: Document properties (title, author, etc.)

### DocumentMetadata

Contains metadata about the Excel file:

```go
type DocumentMetadata struct {
    Created      time.Time `json:"created"`
    Modified     time.Time `json:"modified"`
    AppVersion   string    `json:"app_version"`
    OriginalFile string    `json:"original_file"`
    FileSize     int64     `json:"file_size"`
    Checksum     string    `json:"checksum"`
}
```

**Fields:**
- `Created`: When the JSON was first created
- `Modified`: Last modification time of the Excel file
- `AppVersion`: Version of GitCells that created this JSON
- `OriginalFile`: Path to the original Excel file
- `FileSize`: Size of the original Excel file in bytes
- `Checksum`: SHA256 hash of the original Excel file

### Sheet

Represents a single worksheet within an Excel file:

```go
type Sheet struct {
    Name            string                 `json:"name"`
    Index           int                    `json:"index"`
    Cells           map[string]Cell        `json:"cells"`
    MergedCells     []MergedCell           `json:"merged_cells,omitempty"`
    RowHeights      map[int]float64        `json:"row_heights,omitempty"`
    ColumnWidths    map[string]float64     `json:"column_widths,omitempty"`
    Hidden          bool                   `json:"hidden"`
    Protection      *SheetProtection       `json:"protection,omitempty"`
    ConditionalFormats []ConditionalFormat `json:"conditional_formats,omitempty"`
}
```

**Fields:**
- `Name`: Worksheet name as it appears in Excel
- `Index`: Zero-based position of the sheet in the workbook
- `Cells`: Map of cell references (e.g., "A1") to cell data
- `MergedCells`: Array of merged cell ranges
- `RowHeights`: Custom row heights (if different from default)
- `ColumnWidths`: Custom column widths (if different from default)
- `Hidden`: Whether the sheet is hidden in Excel
- `Protection`: Sheet protection settings
- `ConditionalFormats`: Conditional formatting rules

### Cell

Represents a single cell's data and formatting:

```go
type Cell struct {
    Value          interface{}      `json:"value"`
    Formula        string           `json:"formula,omitempty"`
    Style          *CellStyle       `json:"style,omitempty"`
    Type           CellType         `json:"type"`
    Comment        *Comment         `json:"comment,omitempty"`
    Hyperlink      string           `json:"hyperlink,omitempty"`
    DataValidation *DataValidation  `json:"data_validation,omitempty"`
}
```

**Fields:**
- `Value`: The computed value of the cell (string, number, boolean, etc.)
- `Formula`: Excel formula if the cell contains one (includes the = prefix)
- `Style`: Formatting information (fonts, colors, borders, etc.)
- `Type`: The type of cell content (string, number, formula, etc.)
- `Comment`: Cell comment/annotation
- `Hyperlink`: URL or cell reference if cell contains a hyperlink
- `DataValidation`: Data validation rules applied to the cell

### CellType

Enumeration of possible cell types:

```go
type CellType string

const (
    CellTypeString  CellType = "string"
    CellTypeNumber  CellType = "number"
    CellTypeBoolean CellType = "boolean"
    CellTypeDate    CellType = "date"
    CellTypeError   CellType = "error"
    CellTypeFormula CellType = "formula"
)
```

## Converter Interface

The converter package provides the core Excel ↔ JSON conversion functionality:

```go
type Converter interface {
    ExcelToJSON(filePath string, options ConvertOptions) (*models.ExcelDocument, error)
    JSONToExcel(doc *models.ExcelDocument, outputPath string, options ConvertOptions) error
}
```

### ConvertOptions

Configuration options for conversion operations:

```go
type ConvertOptions struct {
    PreserveFormulas   bool  // Keep Excel formulas in JSON
    PreserveStyles     bool  // Keep cell formatting
    PreserveComments   bool  // Keep cell comments
    CompactJSON        bool  // Generate compact vs. pretty JSON
    IgnoreEmptyCells   bool  // Skip empty cells in JSON
    MaxCellsPerSheet   int   // Memory protection limit
}
```

### Usage Examples

#### Converting Excel to JSON

```go
package main

import (
    "github.com/sirupsen/logrus"
    "github.com/Classic-Homes/gitcells/internal/converter"
)

func main() {
    logger := logrus.New()
    conv := converter.NewConverter(logger)
    
    options := converter.ConvertOptions{
        PreserveFormulas: true,
        PreserveStyles:   true,
        PreserveComments: true,
        IgnoreEmptyCells: true,
    }
    
    doc, err := conv.ExcelToJSON("data.xlsx", options)
    if err != nil {
        logger.Fatal(err)
    }
    
    // Use the document...
}
```

#### Converting JSON back to Excel

```go
err = conv.JSONToExcel(doc, "output.xlsx", options)
if err != nil {
    logger.Fatal(err)
}
```

## File Watcher

The watcher package provides file system monitoring capabilities:

```go
type FileWatcher struct {
    // Internal fields...
}

type EventHandler func(event FileEvent) error

type FileEvent struct {
    Path      string
    Type      EventType
    Timestamp time.Time
}

type EventType int

const (
    EventTypeCreate EventType = iota
    EventTypeModify
    EventTypeDelete
)
```

### Watcher Configuration

```go
type Config struct {
    IgnorePatterns []string      // File patterns to ignore
    DebounceDelay  time.Duration // Delay before processing events
    FileExtensions []string      // Extensions to watch (.xlsx, .xls, etc.)
}
```

### Usage Example

```go
package main

import (
    "time"
    "github.com/sirupsen/logrus"
    "github.com/Classic-Homes/gitcells/internal/watcher"
)

func main() {
    config := &watcher.Config{
        IgnorePatterns: []string{"~$*", "*.tmp"},
        DebounceDelay:  2 * time.Second,
        FileExtensions: []string{".xlsx", ".xls", ".xlsm"},
    }
    
    handler := func(event watcher.FileEvent) error {
        logger.Infof("File %s was %s", event.Path, event.Type)
        // Process the file change...
        return nil
    }
    
    logger := logrus.New()
    fw, err := watcher.NewFileWatcher(config, handler, logger)
    if err != nil {
        logger.Fatal(err)
    }
    
    // Add directories to watch
    err = fw.AddDirectory("./data")
    if err != nil {
        logger.Fatal(err)
    }
    
    // Start watching
    err = fw.Start()
    if err != nil {
        logger.Fatal(err)
    }
    
    // ... wait for events ...
    
    // Stop watching
    fw.Stop()
}
```

## Git Integration

The git package provides version control operations:

```go
type Client struct {
    // Internal fields...
}

type Config struct {
    UserName        string
    UserEmail       string
    CommitTemplate  string
    AutoPush        bool
    AutoPull        bool
    Branch          string
}
```

### Git Operations

```go
func NewClient(repoPath string, config *Config, logger *logrus.Logger) (*Client, error)

func (c *Client) AutoCommit(files []string, metadata map[string]string) error
func (c *Client) Push() error
func (c *Client) Pull() error
func (c *Client) GetStatus() (*Status, error)
```

### Usage Example

```go
gitConfig := &git.Config{
    UserName:       "GitCells",
    UserEmail:      "gitcells@example.com",
    CommitTemplate: "GitCells: {action} {filename} at {timestamp}",
    AutoPush:       false,
    AutoPull:       true,
    Branch:         "main",
}

client, err := git.NewClient(".", gitConfig, logger)
if err != nil {
    logger.Fatal(err)
}

// Commit changes
metadata := map[string]string{
    "filename": "data.xlsx",
    "action":   "modify",
}

err = client.AutoCommit([]string{"data.xlsx.json"}, metadata)
if err != nil {
    logger.Error(err)
}
```

## Configuration Management

The config package handles application configuration:

```go
type Config struct {
    Version   string          `yaml:"version"`
    Git       GitConfig       `yaml:"git"`
    Watcher   WatcherConfig   `yaml:"watcher"`
    Converter ConverterConfig `yaml:"converter"`
}

func Load(configPath string) (*Config, error)
```

### Configuration Structure

```yaml
version: 1.0

git:
  branch: main
  auto_push: false
  auto_pull: true
  user_name: "GitCells"
  user_email: "gitcells@example.com"
  commit_template: "GitCells: {action} {filename} at {timestamp}"

watcher:
  directories: ["./data"]
  ignore_patterns: ["~$*", "*.tmp"]
  debounce_delay: 2s
  file_extensions: [".xlsx", ".xls", ".xlsm"]

converter:
  preserve_formulas: true
  preserve_styles: true
  preserve_comments: true
  compact_json: false
  ignore_empty_cells: true
  max_cells_per_sheet: 1000000
```

## Diff Generation

The models package includes diff functionality for comparing Excel documents:

```go
type ExcelDiff struct {
    Timestamp   time.Time     `json:"timestamp"`
    Summary     DiffSummary   `json:"summary"`
    SheetDiffs  []SheetDiff   `json:"sheet_diffs"`
}

type DiffSummary struct {
    TotalChanges   int `json:"total_changes"`
    AddedSheets    int `json:"added_sheets"`
    ModifiedSheets int `json:"modified_sheets"`
    DeletedSheets  int `json:"deleted_sheets"`
}

func ComputeDiff(oldDoc, newDoc *ExcelDocument) *ExcelDiff
```

### Usage Example

```go
// Load two versions of a document
oldDoc, _ := conv.ExcelToJSON("data_v1.xlsx", options)
newDoc, _ := conv.ExcelToJSON("data_v2.xlsx", options)

// Compute the diff
diff := models.ComputeDiff(oldDoc, newDoc)

// Display the changes
fmt.Printf("Total changes: %d\n", diff.Summary.TotalChanges)
for _, sheetDiff := range diff.SheetDiffs {
    fmt.Printf("Sheet %s has %d changes\n", sheetDiff.SheetName, len(sheetDiff.Changes))
}
```

## Error Handling

GitCells uses structured error handling with custom error types:

```go
type GitCellsError struct {
    Type      ErrorType
    Operation string
    File      string
    Cause     error
    Message   string
    Retryable bool
}

func (e *GitCellsError) Error() string
func (e *GitCellsError) Unwrap() error
func (e *GitCellsError) IsRecoverable() bool
```

### Error Types

```go
type ErrorType string

const (
    ErrorTypeValidation ErrorType = "validation"
    ErrorTypeIO         ErrorType = "io"
    ErrorTypeConversion ErrorType = "conversion"
    ErrorTypeGit        ErrorType = "git"
    ErrorTypeConfig     ErrorType = "config"
    ErrorTypeNetwork    ErrorType = "network"
)
```

### Usage Example

```go
import "github.com/Classic-Homes/gitcells/internal/utils"

// Create a structured error
err := utils.NewError(
    utils.ErrorTypeConversion,
    "excel_to_json",
    "data.xlsx",
    "failed to parse formula in cell B2",
)

// Wrap an existing error
wrappedErr := utils.WrapError(originalErr, utils.ErrorTypeIO, "file_read", "data.xlsx")

// Check if error is recoverable
if utils.IsRecoverableError(err) {
    // Retry the operation
}
```

## Testing

GitCells includes comprehensive testing utilities:

### Test Data

Test Excel files are located in `test/testdata/sample_files/`:
- `simple.xlsx`: Basic spreadsheet with text, numbers, and formulas
- `complex.xlsx`: Multi-sheet workbook with merged cells, comments, and formulas
- `empty.xlsx`: Empty Excel file for testing edge cases

### Test Utilities

```go
// Create test converter
logger := logrus.New()
logger.SetLevel(logrus.WarnLevel)
conv := converter.NewConverter(logger)

// Test conversion
doc, err := conv.ExcelToJSON("test/testdata/sample_files/simple.xlsx", options)
require.NoError(t, err)
assert.NotNil(t, doc)
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.txt ./...

# View coverage report
go tool cover -html=coverage.txt
```

## Performance Considerations

### Memory Management

- Use `MaxCellsPerSheet` to limit memory usage with large files
- Enable `IgnoreEmptyCells` for sparse spreadsheets
- Process files in batches for bulk operations

### File System Optimization

- Configure appropriate `DebounceDelay` to avoid excessive processing
- Use specific file extension filters to reduce monitoring overhead
- Exclude temporary files and directories from watching

### Git Operations

- Batch multiple file changes into single commits
- Use shallow clones for better performance with large repositories
- Configure appropriate Git user credentials to avoid authentication prompts

## Extensibility

GitCells is designed to be extensible:

### Custom Converters

Implement the `Converter` interface to add support for other formats:

```go
type CustomConverter struct {
    logger *logrus.Logger
}

func (c *CustomConverter) ExcelToJSON(filePath string, options ConvertOptions) (*models.ExcelDocument, error) {
    // Custom implementation
}

func (c *CustomConverter) JSONToExcel(doc *models.ExcelDocument, outputPath string, options ConvertOptions) error {
    // Custom implementation
}
```

### Custom Event Handlers

Create specialized file event handlers:

```go
func DatabaseSyncHandler(event watcher.FileEvent) error {
    // Sync changes to database
    return nil
}

func NotificationHandler(event watcher.FileEvent) error {
    // Send notifications about changes
    return nil
}
```

### Plugin Architecture

While not yet implemented, GitCells is designed to support plugins for:
- Custom file format support
- Additional version control systems
- Integration with external services
- Custom conflict resolution strategies

This API documentation provides a comprehensive overview of GitCells's internal architecture and public interfaces. For more specific implementation details, refer to the source code and inline documentation.