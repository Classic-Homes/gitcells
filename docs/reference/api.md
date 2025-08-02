# API Reference

GitCells can be used as a Go library in your own applications. This reference documents the public API for programmatic access to GitCells functionality.

## Installation

```bash
go get github.com/Classic-Homes/gitcells
```

## Package Overview

GitCells provides several packages for different functionality:

- `converter` - Excel/JSON conversion
- `watcher` - File system monitoring
- `git` - Git integration
- `config` - Configuration management
- `models` - Data structures

## Converter Package

The converter package handles conversion between Excel and JSON formats.

### Import

```go
import "github.com/Classic-Homes/gitcells/internal/converter"
```

### Types

#### Converter Interface

```go
type Converter interface {
    // Convert Excel to JSON
    ExcelToJSON(filePath string, options ConvertOptions) (*models.ExcelDocument, error)
    ExcelToJSONFile(inputPath, outputPath string, options ConvertOptions) error
    
    // Convert JSON to Excel
    JSONToExcel(doc *models.ExcelDocument, outputPath string, options ConvertOptions) error
    JSONFileToExcel(inputPath, outputPath string, options ConvertOptions) error
}
```

#### ConvertOptions

```go
type ConvertOptions struct {
    PreserveFormulas    bool
    PreserveStyles      bool
    PreserveComments    bool
    PreserveCharts      bool
    PreservePivotTables bool
    CompactJSON         bool
    IgnoreEmptyCells    bool
    MaxCellsPerSheet    int
    ChunkingStrategy    string
    ProgressCallback    func(status string, current, total int)
}
```

### Functions

#### NewConverter

```go
func NewConverter(logger *logrus.Logger) Converter
```

Creates a new converter instance.

**Parameters:**
- `logger`: Logger instance for output

**Returns:**
- Converter interface implementation

### Example Usage

```go
package main

import (
    "log"
    "github.com/Classic-Homes/gitcells/internal/converter"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Create converter
    conv := converter.NewConverter(logger)
    
    // Set options
    opts := converter.ConvertOptions{
        PreserveFormulas: true,
        PreserveStyles:   true,
        CompactJSON:      false,
        ProgressCallback: func(status string, current, total int) {
            log.Printf("%s: %d/%d\n", status, current, total)
        },
    }
    
    // Convert Excel to JSON
    doc, err := conv.ExcelToJSON("Budget.xlsx", opts)
    if err != nil {
        log.Fatal(err)
    }
    
    // Save to file
    err = conv.ExcelToJSONFile("Budget.xlsx", "Budget.xlsx", opts)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Watcher Package

The watcher package provides file system monitoring capabilities.

### Import

```go
import "github.com/Classic-Homes/gitcells/internal/watcher"
```

### Types

#### FileWatcher Interface

```go
type FileWatcher interface {
    Start() error
    Stop() error
    AddDirectory(path string) error
    RemoveDirectory(path string) error
    GetWatchedDirectories() []string
}
```

#### Config

```go
type Config struct {
    IgnorePatterns []string
    DebounceDelay  time.Duration
    FileExtensions []string
    Recursive      bool
    FollowSymlinks bool
}
```

#### FileEvent

```go
type FileEvent struct {
    Path      string
    Type      EventType
    Timestamp time.Time
}

type EventType int

const (
    Created EventType = iota
    Modified
    Deleted
    Renamed
)
```

### Functions

#### NewFileWatcher

```go
func NewFileWatcher(
    config *Config,
    handler func(FileEvent) error,
    logger *logrus.Logger,
) (FileWatcher, error)
```

Creates a new file watcher instance.

**Parameters:**
- `config`: Watcher configuration
- `handler`: Callback for file events
- `logger`: Logger instance

**Returns:**
- FileWatcher interface
- Error if creation fails

### Example Usage

```go
package main

import (
    "log"
    "time"
    "github.com/Classic-Homes/gitcells/internal/watcher"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Configure watcher
    config := &watcher.Config{
        IgnorePatterns: []string{"~$*", "*.tmp"},
        DebounceDelay:  2 * time.Second,
        FileExtensions: []string{".xlsx", ".xls"},
        Recursive:      true,
    }
    
    // Event handler
    handler := func(event watcher.FileEvent) error {
        log.Printf("File %s was %s\n", event.Path, event.Type)
        // Process the file here
        return nil
    }
    
    // Create watcher
    fw, err := watcher.NewFileWatcher(config, handler, logger)
    if err != nil {
        log.Fatal(err)
    }
    
    // Add directories
    fw.AddDirectory("./excel-files")
    
    // Start watching
    if err := fw.Start(); err != nil {
        log.Fatal(err)
    }
    
    // Wait...
    time.Sleep(1 * time.Hour)
    
    // Stop watching
    fw.Stop()
}
```

## Git Package

The git package provides Git repository operations.

### Import

```go
import "github.com/Classic-Homes/gitcells/internal/git"
```

### Types

#### Client Interface

```go
type Client interface {
    AutoCommit(files []string, message string) error
    GetStatus() (*Status, error)
    Pull() error
    Push() error
    GetCurrentBranch() (string, error)
    Checkout(branch string) error
}
```

#### Config

```go
type Config struct {
    UserName  string
    UserEmail string
    SignKey   string
}
```

#### Status

```go
type Status struct {
    Branch     string
    Clean      bool
    Modified   []string
    Untracked  []string
    Staged     []string
}
```

### Functions

#### NewClient

```go
func NewClient(
    repoPath string,
    config *Config,
    logger *logrus.Logger,
) (Client, error)
```

Creates a new Git client.

### Example Usage

```go
package main

import (
    "log"
    "github.com/Classic-Homes/gitcells/internal/git"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Configure Git
    config := &git.Config{
        UserName:  "GitCells Bot",
        UserEmail: "gitcells@example.com",
    }
    
    // Create client
    client, err := git.NewClient(".", config, logger)
    if err != nil {
        log.Fatal(err)
    }
    
    // Check status
    status, err := client.GetStatus()
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Branch: %s, Clean: %v\n", status.Branch, status.Clean)
    
    // Commit files
    files := []string{"Budget.xlsx", ".gitcells/data/Budget.xlsx_chunks/"}
    err = client.AutoCommit(files, "Updated budget spreadsheet")
    if err != nil {
        log.Fatal(err)
    }
}
```

## Models Package

The models package contains data structures used throughout GitCells.

### Import

```go
import "github.com/Classic-Homes/gitcells/pkg/models"
```

### Core Types

#### ExcelDocument

```go
type ExcelDocument struct {
    Version      string                 `json:"version"`
    Metadata     DocumentMetadata       `json:"metadata"`
    Properties   DocumentProperties     `json:"properties,omitempty"`
    Sheets       []Sheet               `json:"sheets"`
    DefinedNames map[string]string     `json:"defined_names,omitempty"`
    VBAProject   *VBAProject           `json:"vba_project,omitempty"`
}
```

#### Sheet

```go
type Sheet struct {
    Name               string                    `json:"name"`
    Index              int                      `json:"index"`
    Visible            bool                     `json:"visible"`
    Protection         *SheetProtection         `json:"protection,omitempty"`
    Cells              map[string]Cell          `json:"cells"`
    MergedCells        []MergedCell            `json:"merged_cells,omitempty"`
    RowHeights         map[int]float64         `json:"row_heights,omitempty"`
    ColumnWidths       map[string]float64      `json:"column_widths,omitempty"`
    Charts             []Chart                 `json:"charts,omitempty"`
    PivotTables        []PivotTable           `json:"pivot_tables,omitempty"`
    ConditionalFormats []ConditionalFormat    `json:"conditional_formats,omitempty"`
}
```

#### Chart

Chart objects represent detected chart patterns from tabular data.

```go
type Chart struct {
    ID       string        `json:"id"`
    Type     string        `json:"type"` // "pie", "line", "column", "unknown"
    Title    string        `json:"title,omitempty"`
    Position ChartPosition `json:"position"`
    Series   []ChartSeries `json:"series"`
    Legend   *ChartLegend  `json:"legend,omitempty"`
    Axes     *ChartAxes    `json:"axes,omitempty"`
    Style    *ChartStyle   `json:"style,omitempty"`
}

type ChartPosition struct {
    X      float64 `json:"x"`
    Y      float64 `json:"y"`
    Width  float64 `json:"width"`
    Height float64 `json:"height"`
}

type ChartSeries struct {
    Name       string `json:"name,omitempty"`
    Categories string `json:"categories,omitempty"` // Range reference like "A1:A10"
    Values     string `json:"values,omitempty"`     // Range reference like "B1:B10"
    Color      string `json:"color,omitempty"`
}
```

**Chart Detection**: Charts are detected through intelligent pattern analysis of tabular data. When GitCells finds headers with multiple numeric columns, it creates chart metadata representing the likely chart structure.

**Chart Types**: 
- `"pie"` - Single numeric column
- `"line"` - Multiple columns with many rows (time-series)
- `"column"` - Multiple numeric columns (comparison data)
- `"unknown"` - Patterns detected but type unclear

**Limitations**: Chart objects represent detected data patterns, not actual embedded chart objects from Excel files.

#### Cell

```go
type Cell struct {
    Value           interface{}      `json:"value"`
    Type            CellType         `json:"type"`
    Formula         string           `json:"formula,omitempty"`
    FormulaR1C1     string           `json:"formula_r1c1,omitempty"`
    ArrayFormula    *ArrayFormula    `json:"array_formula,omitempty"`
    Style           *CellStyle       `json:"style,omitempty"`
    Comment         *Comment         `json:"comment,omitempty"`
    Hyperlink       *Hyperlink       `json:"hyperlink,omitempty"`
    DataValidation  *DataValidation  `json:"data_validation,omitempty"`
    NumberFormat    string           `json:"number_format,omitempty"`
}
```

## Config Package

The config package handles configuration management.

### Import

```go
import "github.com/Classic-Homes/gitcells/internal/config"
```

### Types

#### Config

```go
type Config struct {
    Version   string           `yaml:"version"`
    Git       GitConfig        `yaml:"git"`
    Watcher   WatcherConfig    `yaml:"watcher"`
    Converter ConverterConfig  `yaml:"converter"`
    Advanced  AdvancedConfig   `yaml:"advanced"`
}
```

### Functions

#### Load

```go
func Load(path string) (*Config, error)
```

Loads configuration from file.

#### LoadWithDefaults

```go
func LoadWithDefaults() *Config
```

Returns default configuration.

### Example Usage

```go
package main

import (
    "log"
    "github.com/Classic-Homes/gitcells/internal/config"
)

func main() {
    // Load configuration
    cfg, err := config.Load(".gitcells.yaml")
    if err != nil {
        // Use defaults
        cfg = config.LoadWithDefaults()
    }
    
    log.Printf("Debounce delay: %s\n", cfg.Watcher.DebounceDelay)
    log.Printf("Auto-push: %v\n", cfg.Git.AutoPush)
}
```

## Complete Example

Here's a complete example that combines multiple packages:

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/Classic-Homes/gitcells/internal/config"
    "github.com/Classic-Homes/gitcells/internal/converter"
    "github.com/Classic-Homes/gitcells/internal/git"
    "github.com/Classic-Homes/gitcells/internal/watcher"
    "github.com/sirupsen/logrus"
)

func main() {
    // Setup logger
    logger := logrus.New()
    logger.SetLevel(logrus.InfoLevel)
    
    // Load configuration
    cfg, err := config.Load(".gitcells.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create converter
    conv := converter.NewConverter(logger)
    
    // Create Git client
    gitConfig := &git.Config{
        UserName:  cfg.Git.UserName,
        UserEmail: cfg.Git.UserEmail,
    }
    
    gitClient, err := git.NewClient(".", gitConfig, logger)
    if err != nil {
        log.Fatal(err)
    }
    
    // Event handler
    handler := func(event watcher.FileEvent) error {
        logger.Infof("Processing %s: %s", event.Type, event.Path)
        
        // Convert to JSON
        opts := converter.ConvertOptions{
            PreserveFormulas: cfg.Converter.PreserveFormulas,
            PreserveStyles:   cfg.Converter.PreserveStyles,
        }
        
        // Convert to chunks in .gitcells/data/
        err := conv.ExcelToJSONFile(event.Path, jsonPath, opts)
        if err != nil {
            return err
        }
        
        // Commit to Git
        if cfg.Git.AutoCommit {
            message := fmt.Sprintf("Updated %s", filepath.Base(event.Path))
            return gitClient.AutoCommit([]string{jsonPath}, message)
        }
        
        return nil
    }
    
    // Create watcher
    watcherConfig := &watcher.Config{
        IgnorePatterns: cfg.Watcher.IgnorePatterns,
        DebounceDelay:  cfg.Watcher.DebounceDelay,
        FileExtensions: cfg.Watcher.FileExtensions,
    }
    
    fw, err := watcher.NewFileWatcher(watcherConfig, handler, logger)
    if err != nil {
        log.Fatal(err)
    }
    
    // Add directories
    for _, dir := range cfg.Watcher.Directories {
        fw.AddDirectory(dir)
    }
    
    // Start watching
    if err := fw.Start(); err != nil {
        log.Fatal(err)
    }
    
    logger.Info("Watching for changes... Press Ctrl+C to stop")
    
    // Wait for interrupt
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan
    
    // Cleanup
    logger.Info("Shutting down...")
    fw.Stop()
}
```

## Error Handling

All GitCells packages follow consistent error handling patterns:

```go
// Check for specific errors
if err != nil {
    switch {
    case errors.Is(err, converter.ErrUnsupportedFormat):
        // Handle unsupported format
    case errors.Is(err, watcher.ErrPathNotFound):
        // Handle missing path
    case errors.Is(err, git.ErrNotRepository):
        // Handle missing repository
    default:
        // Handle generic error
    }
}
```

## Thread Safety

- Converter: Thread-safe, can be used concurrently
- Watcher: Thread-safe after Start()
- Git Client: NOT thread-safe, use mutex for concurrent access
- Config: Read-only after loading, safe for concurrent reads

## Performance Considerations

1. **Converter**: Use chunking for large files
2. **Watcher**: Limit watched directories
3. **Git**: Batch operations when possible
4. **Memory**: Set GOGC environment variable for aggressive GC

## Next Steps

- See [Architecture](../development/architecture.md) for system design
- Review [Contributing](../development/contributing.md) to contribute
- Check [Building](../development/building.md) for development setup
- Read source code for additional details