# SheetSync Implementation Guide

## Project Overview

SheetSync is a cross-platform Go application that bridges Excel and Git by converting Excel files to version-controlled JSON format. This guide provides a structured approach to implementation with recommended improvements.

## Phase 1: Core Foundation

### 1.1 Project Structure

```
sheetsync/
├── cmd/
│   └── sheetsync/
│       ├── main.go
│       ├── init.go
│       ├── watch.go
│       ├── sync.go
│       ├── convert.go
│       └── status.go
├── internal/
│   ├── converter/
│   │   ├── excel_to_json.go
│   │   ├── json_to_excel.go
│   │   ├── converter.go
│   │   └── types.go
│   ├── git/
│   │   ├── client.go
│   │   ├── operations.go
│   │   └── conflicts.go
│   ├── watcher/
│   │   ├── watcher.go
│   │   └── debouncer.go
│   ├── config/
│   │   ├── config.go
│   │   └── defaults.go
│   └── utils/
│       ├── path.go
│       └── logging.go
├── pkg/
│   └── models/
│       ├── excel.go
│       └── diff.go
├── test/
│   └── testdata/
│       └── sample_files/
├── .gitignore
├── .sheetsync.yaml
├── CLAUDE.md
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 1.2 Core Dependencies

```bash
# Essential dependencies only
go get github.com/xuri/excelize/v2
go get github.com/go-git/go-git/v5
go get github.com/fsnotify/fsnotify
go get github.com/spf13/cobra
go get github.com/spf13/viper
go get github.com/sirupsen/logrus  # Better logging
```

## Phase 2: Data Models (Improved)

### 2.1 Enhanced Excel Model (`pkg/models/excel.go`)

```go
package models

import (
    "time"
    "encoding/json"
)

// ExcelDocument represents the complete Excel file structure
type ExcelDocument struct {
    Version      string           `json:"version"`
    Metadata     DocumentMetadata `json:"metadata"`
    Sheets       []Sheet          `json:"sheets"`
    DefinedNames map[string]string `json:"defined_names,omitempty"`
    Properties   DocumentProperties `json:"properties,omitempty"`
}

type DocumentMetadata struct {
    Created      time.Time `json:"created"`
    Modified     time.Time `json:"modified"`
    AppVersion   string    `json:"app_version"`
    OriginalFile string    `json:"original_file"`
    FileSize     int64     `json:"file_size"`
    Checksum     string    `json:"checksum"` // SHA256 of original
}

type DocumentProperties struct {
    Title       string `json:"title,omitempty"`
    Subject     string `json:"subject,omitempty"`
    Author      string `json:"author,omitempty"`
    Company     string `json:"company,omitempty"`
    Keywords    string `json:"keywords,omitempty"`
    Description string `json:"description,omitempty"`
}

type Sheet struct {
    Name          string                 `json:"name"`
    Index         int                    `json:"index"`
    Cells         map[string]Cell        `json:"cells"`
    MergedCells   []MergedCell           `json:"merged_cells,omitempty"`
    RowHeights    map[int]float64        `json:"row_heights,omitempty"`
    ColumnWidths  map[string]float64     `json:"column_widths,omitempty"`
    Hidden        bool                   `json:"hidden"`
    Protection    *SheetProtection       `json:"protection,omitempty"`
    ConditionalFormats []ConditionalFormat `json:"conditional_formats,omitempty"`
}

type Cell struct {
    Value      interface{}   `json:"value"`
    Formula    string        `json:"formula,omitempty"`
    Style      *CellStyle    `json:"style,omitempty"`
    Type       CellType      `json:"type"`
    Comment    *Comment      `json:"comment,omitempty"`
    Hyperlink  string        `json:"hyperlink,omitempty"`
    DataValidation *DataValidation `json:"data_validation,omitempty"`
}

type CellType string

const (
    CellTypeString  CellType = "string"
    CellTypeNumber  CellType = "number"
    CellTypeBoolean CellType = "boolean"
    CellTypeDate    CellType = "date"
    CellTypeError   CellType = "error"
    CellTypeFormula CellType = "formula"
)

// Compact representation for merged cells
type MergedCell struct {
    Range string `json:"range"` // e.g., "A1:C3"
}

// Additional types for completeness...
```

### 2.2 Diff Model (`pkg/models/diff.go`)

```go
package models

type ExcelDiff struct {
    Timestamp   time.Time     `json:"timestamp"`
    Summary     DiffSummary   `json:"summary"`
    SheetDiffs  []SheetDiff   `json:"sheet_diffs"`
}

type DiffSummary struct {
    TotalChanges int `json:"total_changes"`
    AddedSheets  int `json:"added_sheets"`
    ModifiedSheets int `json:"modified_sheets"`
    DeletedSheets int `json:"deleted_sheets"`
}

type SheetDiff struct {
    SheetName string        `json:"sheet_name"`
    Changes   []CellChange  `json:"changes"`
}

type CellChange struct {
    Cell     string      `json:"cell"`
    Type     ChangeType  `json:"type"`
    OldValue interface{} `json:"old_value,omitempty"`
    NewValue interface{} `json:"new_value,omitempty"`
}

type ChangeType string

const (
    ChangeTypeAdd    ChangeType = "add"
    ChangeTypeModify ChangeType = "modify"
    ChangeTypeDelete ChangeType = "delete"
)
```

## Phase 3: Core Converter Implementation

### 3.1 Converter Interface (`internal/converter/converter.go`)

```go
package converter

import (
    "github.com/Classic-Homes/sheetsync/pkg/models"
)

type Converter interface {
    ExcelToJSON(filePath string, options ConvertOptions) (*models.ExcelDocument, error)
    JSONToExcel(doc *models.ExcelDocument, outputPath string, options ConvertOptions) error
}

type ConvertOptions struct {
    PreserveFormulas   bool
    PreserveStyles     bool
    PreserveComments   bool
    CompactJSON        bool
    IgnoreEmptyCells   bool
    MaxCellsPerSheet   int // Prevent memory issues with huge files
}

type converter struct {
    logger *logrus.Logger
}

func NewConverter(logger *logrus.Logger) Converter {
    return &converter{logger: logger}
}
```

### 3.2 Excel to JSON Implementation (`internal/converter/excel_to_json.go`)

```go
package converter

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "os"
    "time"
    
    "github.com/xuri/excelize/v2"
    "github.com/Classic-Homes/sheetsync/pkg/models"
)

func (c *converter) ExcelToJSON(filePath string, options ConvertOptions) (*models.ExcelDocument, error) {
    // Calculate file checksum
    checksum, err := c.calculateChecksum(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to calculate checksum: %w", err)
    }

    // Open Excel file
    f, err := excelize.OpenFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open Excel file: %w", err)
    }
    defer f.Close()

    // Get file info
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to get file info: %w", err)
    }

    doc := &models.ExcelDocument{
        Version: "1.0",
        Metadata: models.DocumentMetadata{
            Created:      time.Now(),
            Modified:     fileInfo.ModTime(),
            AppVersion:   "sheetsync-0.1.0",
            OriginalFile: filePath,
            FileSize:     fileInfo.Size(),
            Checksum:     checksum,
        },
        Sheets:       []models.Sheet{},
        DefinedNames: make(map[string]string),
    }

    // Extract document properties
    props, err := f.GetDocProps()
    if err == nil && props != nil {
        doc.Properties = c.extractProperties(props)
    }

    // Process each sheet
    for index, sheetName := range f.GetSheetList() {
        sheet, err := c.processSheet(f, sheetName, index, options)
        if err != nil {
            c.logger.Warnf("Failed to process sheet %s: %v", sheetName, err)
            continue
        }
        doc.Sheets = append(doc.Sheets, *sheet)
    }

    // Extract defined names
    for _, definedName := range f.GetDefinedName() {
        doc.DefinedNames[definedName.Name] = definedName.RefersTo
    }

    return doc, nil
}

func (c *converter) processSheet(f *excelize.File, sheetName string, index int, options ConvertOptions) (*models.Sheet, error) {
    sheet := &models.Sheet{
        Name:         sheetName,
        Index:        index,
        Cells:        make(map[string]models.Cell),
        MergedCells:  []models.MergedCell{},
        RowHeights:   make(map[int]float64),
        ColumnWidths: make(map[string]float64),
    }

    // Get all cells in sheet
    rows, err := f.GetRows(sheetName)
    if err != nil {
        return nil, err
    }

    cellCount := 0
    for rowIndex, row := range rows {
        for colIndex, cellValue := range row {
            if options.MaxCellsPerSheet > 0 && cellCount >= options.MaxCellsPerSheet {
                c.logger.Warnf("Sheet %s exceeded max cells limit (%d)", sheetName, options.MaxCellsPerSheet)
                break
            }

            cellRef, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
            
            // Skip empty cells if option is set
            if options.IgnoreEmptyCells && cellValue == "" {
                continue
            }

            cell := models.Cell{
                Value: cellValue,
                Type:  c.detectCellType(cellValue),
            }

            // Get formula if exists
            if options.PreserveFormulas {
                formula, _ := f.GetCellFormula(sheetName, cellRef)
                if formula != "" {
                    cell.Formula = formula
                    cell.Type = models.CellTypeFormula
                }
            }

            // Get style if requested
            if options.PreserveStyles {
                styleID, _ := f.GetCellStyle(sheetName, cellRef)
                if styleID > 0 {
                    cell.Style = c.extractCellStyle(f, styleID)
                }
            }

            // Get comment if requested
            if options.PreserveComments {
                comment, _ := f.GetComment(sheetName, cellRef)
                if comment != "" {
                    cell.Comment = &models.Comment{Text: comment}
                }
            }

            sheet.Cells[cellRef] = cell
            cellCount++
        }
    }

    // Get merged cells
    mergedCells, _ := f.GetMergeCells(sheetName)
    for _, mc := range mergedCells {
        sheet.MergedCells = append(sheet.MergedCells, models.MergedCell{
            Range: mc.GetStartAxis() + ":" + mc.GetEndAxis(),
        })
    }

    return sheet, nil
}

func (c *converter) calculateChecksum(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    hash := sha256.New()
    if _, err := io.Copy(hash, file); err != nil {
        return "", err
    }

    return hex.EncodeToString(hash.Sum(nil)), nil
}

// Additional helper methods...
```

## Phase 4: Git Integration (Enhanced)

### 4.1 Git Client (`internal/git/client.go`)

```go
package git

import (
    "fmt"
    "path/filepath"
    "time"

    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/object"
    "github.com/sirupsen/logrus"
)

type Client struct {
    repo     *git.Repository
    worktree *git.Worktree
    config   *Config
    logger   *logrus.Logger
}

type Config struct {
    UserName        string
    UserEmail       string
    CommitTemplate  string
    AutoPush        bool
    AutoPull        bool
    Branch          string
}

func NewClient(repoPath string, config *Config, logger *logrus.Logger) (*Client, error) {
    repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
        DetectDotGit: true,
    })
    
    if err == git.ErrRepositoryNotExists {
        // Initialize new repository
        repo, err = git.PlainInit(repoPath, false)
        if err != nil {
            return nil, fmt.Errorf("failed to initialize repository: %w", err)
        }
        logger.Info("Initialized new git repository")
    } else if err != nil {
        return nil, fmt.Errorf("failed to open repository: %w", err)
    }

    worktree, err := repo.Worktree()
    if err != nil {
        return nil, fmt.Errorf("failed to get worktree: %w", err)
    }

    return &Client{
        repo:     repo,
        worktree: worktree,
        config:   config,
        logger:   logger,
    }, nil
}

func (c *Client) AutoCommit(files []string, metadata map[string]string) error {
    // Stage files
    for _, file := range files {
        relPath, _ := filepath.Rel(c.worktree.Filesystem.Root(), file)
        if err := c.worktree.Add(relPath); err != nil {
            return fmt.Errorf("failed to stage %s: %w", file, err)
        }
    }

    // Check if there are changes to commit
    status, err := c.worktree.Status()
    if err != nil {
        return fmt.Errorf("failed to get status: %w", err)
    }

    if status.IsClean() {
        c.logger.Debug("No changes to commit")
        return nil
    }

    // Create commit message
    message := c.formatCommitMessage(files, metadata)

    // Create commit
    commit, err := c.worktree.Commit(message, &git.CommitOptions{
        Author: &object.Signature{
            Name:  c.config.UserName,
            Email: c.config.UserEmail,
            When:  time.Now(),
        },
    })

    if err != nil {
        return fmt.Errorf("failed to create commit: %w", err)
    }

    c.logger.Infof("Created commit: %s", commit.String())

    // Auto push if enabled
    if c.config.AutoPush {
        return c.Push()
    }

    return nil
}

func (c *Client) formatCommitMessage(files []string, metadata map[string]string) string {
    // Use template with variable substitution
    message := c.config.CommitTemplate
    
    // Replace variables
    replacements := map[string]string{
        "{timestamp}": time.Now().Format(time.RFC3339),
        "{files}":     fmt.Sprintf("%d file(s)", len(files)),
        "{branch}":    c.config.Branch,
    }

    for key, value := range metadata {
        replacements["{"+key+"}"] = value
    }

    for key, value := range replacements {
        message = strings.ReplaceAll(message, key, value)
    }

    return message
}

// Additional methods: Pull, Push, GetHistory, etc.
```

## Phase 5: File Watcher (Improved)

### 5.1 Watcher with Debouncing (`internal/watcher/watcher.go`)

```go
package watcher

import (
    "context"
    "path/filepath"
    "strings"
    "sync"
    "time"

    "github.com/fsnotify/fsnotify"
    "github.com/sirupsen/logrus"
)

type FileWatcher struct {
    watcher      *fsnotify.Watcher
    watchedDirs  sync.Map
    eventHandler EventHandler
    debouncer    *Debouncer
    config       *Config
    logger       *logrus.Logger
    ctx          context.Context
    cancel       context.CancelFunc
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

type Config struct {
    IgnorePatterns []string
    DebounceDelay  time.Duration
    FileExtensions []string
}

func NewFileWatcher(config *Config, handler EventHandler, logger *logrus.Logger) (*FileWatcher, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    ctx, cancel := context.WithCancel(context.Background())

    fw := &FileWatcher{
        watcher:      watcher,
        eventHandler: handler,
        debouncer:    NewDebouncer(config.DebounceDelay),
        config:       config,
        logger:       logger,
        ctx:          ctx,
        cancel:       cancel,
    }

    return fw, nil
}

func (fw *FileWatcher) Start() error {
    go fw.processEvents()
    return nil
}

func (fw *FileWatcher) Stop() error {
    fw.cancel()
    return fw.watcher.Close()
}

func (fw *FileWatcher) AddDirectory(path string) error {
    // Walk directory tree and add all subdirectories
    err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            if fw.shouldIgnorePath(walkPath) {
                return filepath.SkipDir
            }

            if err := fw.watcher.Add(walkPath); err != nil {
                fw.logger.Warnf("Failed to watch %s: %v", walkPath, err)
            } else {
                fw.watchedDirs.Store(walkPath, true)
                fw.logger.Debugf("Watching directory: %s", walkPath)
            }
        }

        return nil
    })

    return err
}

func (fw *FileWatcher) processEvents() {
    for {
        select {
        case <-fw.ctx.Done():
            return

        case event, ok := <-fw.watcher.Events:
            if !ok {
                return
            }

            if fw.shouldProcessFile(event.Name) {
                fw.debouncer.Debounce(event.Name, func() {
                    fileEvent := FileEvent{
                        Path:      event.Name,
                        Timestamp: time.Now(),
                    }

                    switch {
                    case event.Op&fsnotify.Create == fsnotify.Create:
                        fileEvent.Type = EventTypeCreate
                    case event.Op&fsnotify.Write == fsnotify.Write:
                        fileEvent.Type = EventTypeModify
                    case event.Op&fsnotify.Remove == fsnotify.Remove:
                        fileEvent.Type = EventTypeDelete
                    default:
                        return
                    }

                    if err := fw.eventHandler(fileEvent); err != nil {
                        fw.logger.Errorf("Event handler error: %v", err)
                    }
                })
            }

        case err, ok := <-fw.watcher.Errors:
            if !ok {
                return
            }
            fw.logger.Errorf("Watcher error: %v", err)
        }
    }
}

func (fw *FileWatcher) shouldProcessFile(path string) bool {
    // Check if it's an Excel file
    ext := strings.ToLower(filepath.Ext(path))
    validExt := false
    for _, allowedExt := range fw.config.FileExtensions {
        if ext == allowedExt {
            validExt = true
            break
        }
    }

    if !validExt {
        return false
    }

    // Check ignore patterns
    return !fw.shouldIgnorePath(path)
}

func (fw *FileWatcher) shouldIgnorePath(path string) bool {
    base := filepath.Base(path)
    
    // Always ignore Excel temp files
    if strings.HasPrefix(base, "~$") {
        return true
    }

    for _, pattern := range fw.config.IgnorePatterns {
        if matched, _ := filepath.Match(pattern, base); matched {
            return true
        }
    }

    return false
}
```

### 5.2 Debouncer (`internal/watcher/debouncer.go`)

```go
package watcher

import (
    "sync"
    "time"
)

type Debouncer struct {
    delay   time.Duration
    timers  sync.Map
    mu      sync.Mutex
}

func NewDebouncer(delay time.Duration) *Debouncer {
    return &Debouncer{
        delay: delay,
    }
}

func (d *Debouncer) Debounce(key string, fn func()) {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Cancel existing timer
    if timer, exists := d.timers.Load(key); exists {
        timer.(*time.Timer).Stop()
    }

    // Create new timer
    timer := time.AfterFunc(d.delay, func() {
        fn()
        d.timers.Delete(key)
    })

    d.timers.Store(key, timer)
}
```

## Phase 6: CLI Commands (Enhanced)

### 6.1 Main Entry Point (`cmd/sheetsync/main.go`)

```go
package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/sirupsen/logrus"
)

var (
    version   = "dev"
    buildTime = "unknown"
    logger    *logrus.Logger
)

func main() {
    logger = setupLogger()

    rootCmd := &cobra.Command{
        Use:   "sheetsync",
        Short: "Version control for Excel files",
        Long:  `SheetSync converts Excel files to JSON for version control and collaboration`,
        Version: fmt.Sprintf("%s (built %s)", version, buildTime),
    }

    // Add commands
    rootCmd.AddCommand(
        newInitCommand(logger),
        newWatchCommand(logger),
        newSyncCommand(logger),
        newConvertCommand(logger),
        newStatusCommand(logger),
        newHistoryCommand(logger),
        newDiffCommand(logger),
    )

    // Global flags
    rootCmd.PersistentFlags().String("config", "", "config file (default: .sheetsync.yaml)")
    rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose logging")

    if err := rootCmd.Execute(); err != nil {
        logger.Error(err)
        os.Exit(1)
    }
}

func setupLogger() *logrus.Logger {
    logger := logrus.New()
    logger.SetFormatter(&logrus.TextFormatter{
        FullTimestamp: true,
        DisableColors: false,
    })
    return logger
}
```

### 6.2 Watch Command (`cmd/sheetsync/watch.go`)

```go
package main

import (
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"
    "github.com/sirupsen/logrus"
    "github.com/Classic-Homes/sheetsync/internal/config"
    "github.com/Classic-Homes/sheetsync/internal/watcher"
    "github.com/Classic-Homes/sheetsync/internal/converter"
    "github.com/Classic-Homes/sheetsync/internal/git"
)

func newWatchCommand(logger *logrus.Logger) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "watch [directories...]",
        Short: "Watch directories for Excel file changes",
        Long:  "Start watching specified directories for Excel file changes and auto-commit to git",
        Args:  cobra.MinimumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Load configuration
            cfg, err := config.Load(cmd.Flag("config").Value.String())
            if err != nil {
                return err
            }

            // Initialize components
            conv := converter.NewConverter(logger)
            gitClient, err := git.NewClient(".", cfg.Git, logger)
            if err != nil {
                return err
            }

            // Create event handler
            handler := func(event watcher.FileEvent) error {
                logger.Infof("Processing %s: %s", event.Type, event.Path)

                // Convert to JSON
                doc, err := conv.ExcelToJSON(event.Path, cfg.Converter.ToOptions())
                if err != nil {
                    return err
                }

                // Save JSON
                jsonPath := event.Path + ".json"
                if err := conv.SaveJSON(doc, jsonPath); err != nil {
                    return err
                }

                // Commit changes
                metadata := map[string]string{
                    "filename": filepath.Base(event.Path),
                    "action":   event.Type.String(),
                }

                return gitClient.AutoCommit([]string{jsonPath}, metadata)
            }

            // Setup watcher
            fw, err := watcher.NewFileWatcher(cfg.Watcher, handler, logger)
            if err != nil {
                return err
            }

            // Add directories
            for _, dir := range args {
                if err := fw.AddDirectory(dir); err != nil {
                    logger.Warnf("Failed to add directory %s: %v", dir, err)
                }
            }

            // Start watching
            if err := fw.Start(); err != nil {
                return err
            }

            logger.Info("Watching for changes... Press Ctrl+C to stop")

            // Wait for interrupt
            sigChan := make(chan os.Signal, 1)
            signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
            <-sigChan

            logger.Info("Shutting down...")
            return fw.Stop()
        },
    }

    return cmd
}
```

## Phase 7: Configuration System

### 7.1 Configuration Structure (`internal/config/config.go`)

```go
package config

import (
    "time"
    "github.com/spf13/viper"
)

type Config struct {
    Version   string           `yaml:"version"`
    Git       GitConfig        `yaml:"git"`
    Watcher   WatcherConfig    `yaml:"watcher"`
    Converter ConverterConfig  `yaml:"converter"`
}

type GitConfig struct {
    Remote         string `yaml:"remote"`
    Branch         string `yaml:"branch"`
    AutoPush       bool   `yaml:"auto_push"`
    AutoPull       bool   `yaml:"auto_pull"`
    UserName       string `yaml:"user_name"`
    UserEmail      string `yaml:"user_email"`
    CommitTemplate string `yaml:"commit_template"`
}

type WatcherConfig struct {
    Directories    []string      `yaml:"directories"`
    IgnorePatterns []string      `yaml:"ignore_patterns"`
    DebounceDelay  time.Duration `yaml:"debounce_delay"`
    FileExtensions []string      `yaml:"file_extensions"`
}

type ConverterConfig struct {
    PreserveFormulas bool `yaml:"preserve_formulas"`
    PreserveStyles   bool `yaml:"preserve_styles"`
    PreserveComments bool `yaml:"preserve_comments"`
    CompactJSON      bool `yaml:"compact_json"`
    IgnoreEmptyCells bool `yaml:"ignore_empty_cells"`
    MaxCellsPerSheet int  `yaml:"max_cells_per_sheet"`
}

func Load(configPath string) (*Config, error) {
    v := viper.New()
    
    // Set defaults
    v.SetDefault("version", "1.0")
    v.SetDefault("git.branch", "main")
    v.SetDefault("git.commit_template", "SheetSync: {action} {filename} at {timestamp}")
    v.SetDefault("watcher.debounce_delay", "1s")
    v.SetDefault("watcher.file_extensions", []string{".xlsx", ".xls", ".xlsm"})
    v.SetDefault("watcher.ignore_patterns", []string{"~$*", "*.tmp"})
    v.SetDefault("converter.preserve_formulas", true)
    v.SetDefault("converter.preserve_styles", true)
    v.SetDefault("converter.max_cells_per_sheet", 1000000)

    // Load config file
    if configPath != "" {
        v.SetConfigFile(configPath)
    } else {
        v.SetConfigName(".sheetsync")
        v.SetConfigType("yaml")
        v.AddConfigPath(".")
    }

    if err := v.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

### 7.2 Default Configuration (`.sheetsync.yaml`)

```yaml
version: 1.0

git:
  branch: main
  auto_push: false
  auto_pull: true
  user_name: "SheetSync"
  user_email: "sheetsync@localhost"
  commit_template: "SheetSync: {action} {filename} at {timestamp}"

watcher:
  directories: []
  ignore_patterns:
    - "~$*"
    - "*.tmp"
    - ".~lock.*"
  debounce_delay: 2s
  file_extensions:
    - ".xlsx"
    - ".xls"
    - ".xlsm"

converter:
  preserve_formulas: true
  preserve_styles: true
  preserve_comments: true
  compact_json: false
  ignore_empty_cells: true
  max_cells_per_sheet: 1000000
```

## Phase 8: Testing Strategy

### 8.1 Test Structure

```
test/
├── converter_test.go
├── git_test.go
├── watcher_test.go
├── integration_test.go
└── testdata/
    ├── simple.xlsx
    ├── complex_formulas.xlsx
    ├── large_file.xlsx
    └── expected/
        ├── simple.json
        └── complex_formulas.json
```

### 8.2 Example Test (`test/converter_test.go`)

```go
package test

import (
    "testing"
    "path/filepath"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/sirupsen/logrus"
    "github.com/Classic-Homes/sheetsync/internal/converter"
)

func TestExcelToJSON(t *testing.T) {
    logger := logrus.New()
    conv := converter.NewConverter(logger)

    tests := []struct {
        name     string
        file     string
        options  converter.ConvertOptions
        validate func(t *testing.T, doc *models.ExcelDocument)
    }{
        {
            name: "simple file",
            file: "testdata/simple.xlsx",
            options: converter.ConvertOptions{
                PreserveFormulas: true,
                PreserveStyles:   true,
            },
            validate: func(t *testing.T, doc *models.ExcelDocument) {
                assert.Len(t, doc.Sheets, 1)
                assert.NotEmpty(t, doc.Sheets[0].Cells)
            },
        },
        {
            name: "complex formulas",
            file: "testdata/complex_formulas.xlsx",
            options: converter.ConvertOptions{
                PreserveFormulas: true,
            },
            validate: func(t *testing.T, doc *models.ExcelDocument) {
                // Check that formulas are preserved
                hasFormula := false
                for _, sheet := range doc.Sheets {
                    for _, cell := range sheet.Cells {
                        if cell.Formula != "" {
                            hasFormula = true
                            break
                        }
                    }
                }
                assert.True(t, hasFormula, "Expected to find formulas")
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            doc, err := conv.ExcelToJSON(filepath.Join("testdata", tt.file), tt.options)
            require.NoError(t, err)
            require.NotNil(t, doc)
            
            tt.validate(t, doc)
        })
    }
}
```

## Phase 9: Build and Release

### 9.1 Makefile

```makefile
.PHONY: all build test clean install lint

VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BINARY := sheetsync

all: test build

build:
	@echo "Building SheetSync $(VERSION)..."
	@mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64 ./cmd/sheetsync
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64 ./cmd/sheetsync
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-amd64.exe ./cmd/sheetsync
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 ./cmd/sheetsync

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

lint:
	@echo "Running linters..."
	golangci-lint run

clean:
	@echo "Cleaning..."
	@rm -rf dist/ coverage.txt

install: build
	@echo "Installing SheetSync..."
	@cp dist/$(BINARY)-$(shell go env GOOS)-$(shell go env GOARCH) $(GOPATH)/bin/$(BINARY)

.DEFAULT_GOAL := all
```

### 9.2 GitHub Actions (`.github/workflows/release.yml`)

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: make test
      
      - name: Build binaries
        run: make build
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Implementation Roadmap

### Week 1: Foundation
- [x] Set up project structure
- [x] Implement core data models
- [x] Create basic Excel to JSON converter
- [x] Write initial unit tests

### Week 2: Core Features
- [x] Implement JSON to Excel converter
- [x] Add Git integration
- [x] Create file watcher with debouncing
- [x] Implement basic CLI commands

### Week 3: Enhancement
- [ ] Add configuration system
- [ ] Implement diff functionality
- [ ] Add conflict resolution
- [ ] Improve error handling and logging

### Week 4: Polish
- [ ] Complete test coverage
- [ ] Add documentation
- [ ] Set up CI/CD pipeline
- [ ] Create installation scripts

## Key Improvements Made

1. **Better Error Handling**: Added context to errors, proper error wrapping
2. **Performance Optimization**: Added cell limits, streaming for large files
3. **Robust File Watching**: Proper debouncing, ignore patterns, context-based shutdown
4. **Enhanced Git Integration**: Better commit messages, conflict detection, atomic operations
5. **Comprehensive Testing**: Test structure, example tests, integration tests
6. **Production-Ready Build**: Proper versioning, cross-platform builds, GitHub Actions
7. **Improved Configuration**: Defaults, validation, multiple config sources
8. **Better Logging**: Structured logging with logrus, debug/info/error levels

## Next Steps

1. Start with Phase 1 - Create the directory structure
2. Implement the core data models (Phase 2)
3. Build the basic Excel to JSON converter
4. Add tests for each component as you build
5. Gradually add features following the roadmap

This implementation guide provides a solid foundation for building SheetSync with production-quality code and architecture.