package adapter

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/Classic-Homes/gitcells/internal/converter"
	"github.com/Classic-Homes/gitcells/internal/git"
	"github.com/Classic-Homes/gitcells/internal/watcher"
	"github.com/sirupsen/logrus"
)

// WatcherAdapter bridges the TUI with the watcher package
type WatcherAdapter struct {
	watcher   *watcher.FileWatcher
	config    *config.Config
	logger    *logrus.Logger
	converter converter.Converter
	gitClient *git.Client

	// State tracking
	isRunning     bool
	startTime     time.Time
	filesWatched  int
	lastEvent     string
	lastEventTime time.Time

	// Event callback for TUI updates
	onEvent func(WatcherEvent)
}

// WatcherEvent represents a watcher event for the TUI
type WatcherEvent struct {
	Type      string // "started", "stopped", "file_changed", "error"
	Message   string
	Details   string
	Timestamp time.Time
	FilePath  string
}

// WatcherStatus represents the current status of the watcher
type WatcherStatus struct {
	IsRunning     bool
	StartTime     time.Time
	FilesWatched  int
	LastEvent     string
	LastEventTime time.Time
	Directories   []string
}

// NewWatcherAdapter creates a new watcher adapter
func NewWatcherAdapter(cfg *config.Config, logger *logrus.Logger, onEvent func(WatcherEvent)) (*WatcherAdapter, error) {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel) // Reduce noise in TUI
	}

	conv := converter.NewConverter(logger)

	return &WatcherAdapter{
		config:    cfg,
		logger:    logger,
		converter: conv,
		onEvent:   onEvent,
	}, nil
}

// Start starts the file watcher
func (wa *WatcherAdapter) Start() error {
	if wa.isRunning {
		return fmt.Errorf("watcher is already running")
	}

	// Initialize git client
	gitConfig := &git.Config{
		UserName:  wa.config.Git.UserName,
		UserEmail: wa.config.Git.UserEmail,
	}

	gitClient, err := git.NewClient(".", gitConfig, wa.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize git client: %w", err)
	}
	wa.gitClient = gitClient

	// Create event handler
	handler := func(event watcher.FileEvent) error {
		wa.lastEvent = fmt.Sprintf("%s: %s", event.Type, filepath.Base(event.Path))
		wa.lastEventTime = event.Timestamp

		// Notify TUI
		if wa.onEvent != nil {
			wa.onEvent(WatcherEvent{
				Type:      "file_changed",
				Message:   fmt.Sprintf("File %s: %s", event.Type, filepath.Base(event.Path)),
				Details:   event.Path,
				Timestamp: event.Timestamp,
				FilePath:  event.Path,
			})
		}

		// Convert Excel to JSON
		convertOptions := converter.ConvertOptions{
			PreserveFormulas: wa.config.Converter.PreserveFormulas,
			PreserveStyles:   wa.config.Converter.PreserveStyles,
			PreserveComments: wa.config.Converter.PreserveComments,
			CompactJSON:      wa.config.Converter.CompactJSON,
			IgnoreEmptyCells: wa.config.Converter.IgnoreEmptyCells,
			MaxCellsPerSheet: wa.config.Converter.MaxCellsPerSheet,
			ChunkingStrategy: "sheet-based",
		}

		if err := wa.converter.ExcelToJSONFile(event.Path, event.Path, convertOptions); err != nil {
			if wa.onEvent != nil {
				wa.onEvent(WatcherEvent{
					Type:      "error",
					Message:   fmt.Sprintf("Conversion failed: %s", filepath.Base(event.Path)),
					Details:   err.Error(),
					Timestamp: time.Now(),
					FilePath:  event.Path,
				})
			}
			return fmt.Errorf("failed to convert Excel to JSON: %w", err)
		}

		// Auto-commit to git
		if wa.gitClient != nil {
			chunkPaths, err := wa.converter.GetChunkPaths(event.Path)
			if err != nil {
				wa.logger.Warnf("Failed to get chunk paths for git commit: %v", err)
				chunkPaths = []string{filepath.Join(".gitcells", "data")}
			}

			message := fmt.Sprintf("GitCells: %s %s", event.Type.String(), filepath.Base(event.Path))
			if err := wa.gitClient.AutoCommit(chunkPaths, message); err != nil {
				if wa.onEvent != nil {
					wa.onEvent(WatcherEvent{
						Type:      "error",
						Message:   "Git commit failed",
						Details:   err.Error(),
						Timestamp: time.Now(),
						FilePath:  event.Path,
					})
				}
				return fmt.Errorf("failed to auto-commit: %w", err)
			}
		}

		return nil
	}

	// Setup watcher
	watcherConfig := &watcher.Config{
		IgnorePatterns: wa.config.Watcher.IgnorePatterns,
		DebounceDelay:  wa.config.Watcher.DebounceDelay,
		FileExtensions: wa.config.Watcher.FileExtensions,
	}

	fw, err := watcher.NewFileWatcher(watcherConfig, handler, wa.logger)
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	wa.watcher = fw

	// Add watch directories from config
	watchDirs := wa.config.Watcher.Directories
	if len(watchDirs) == 0 {
		// Default to current directory if none specified
		watchDirs = []string{"."}
	}

	for _, dir := range watchDirs {
		if err := fw.AddDirectory(dir); err != nil {
			wa.logger.Warnf("Failed to add directory %s: %v", dir, err)
		}
	}

	// Start watching
	if err := fw.Start(); err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}

	wa.isRunning = true
	wa.startTime = time.Now()
	wa.filesWatched = len(fw.GetWatchedDirectories())

	// Notify TUI
	if wa.onEvent != nil {
		wa.onEvent(WatcherEvent{
			Type:      "started",
			Message:   fmt.Sprintf("Watcher started, monitoring %d directories", wa.filesWatched),
			Timestamp: wa.startTime,
		})
	}

	return nil
}

// Stop stops the file watcher
func (wa *WatcherAdapter) Stop() error {
	if !wa.isRunning {
		return fmt.Errorf("watcher is not running")
	}

	if wa.watcher != nil {
		if err := wa.watcher.Stop(); err != nil {
			return fmt.Errorf("failed to stop watcher: %w", err)
		}
	}

	wa.isRunning = false
	wa.lastEvent = "Stopped"
	wa.lastEventTime = time.Now()

	// Notify TUI
	if wa.onEvent != nil {
		wa.onEvent(WatcherEvent{
			Type:      "stopped",
			Message:   "Watcher stopped",
			Timestamp: wa.lastEventTime,
		})
	}

	return nil
}

// GetStatus returns the current watcher status
func (wa *WatcherAdapter) GetStatus() WatcherStatus {
	status := WatcherStatus{
		IsRunning:     wa.isRunning,
		StartTime:     wa.startTime,
		FilesWatched:  wa.filesWatched,
		LastEvent:     wa.lastEvent,
		LastEventTime: wa.lastEventTime,
	}

	if wa.watcher != nil {
		status.Directories = wa.watcher.GetWatchedDirectories()
		status.FilesWatched = len(status.Directories)
	}

	return status
}

// IsRunning returns whether the watcher is currently running
func (wa *WatcherAdapter) IsRunning() bool {
	return wa.isRunning
}
