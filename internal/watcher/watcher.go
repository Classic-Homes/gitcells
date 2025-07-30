package watcher

import (
	"context"
	"os"
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

func (et EventType) String() string {
	switch et {
	case EventTypeCreate:
		return "create"
	case EventTypeModify:
		return "modify"
	case EventTypeDelete:
		return "delete"
	default:
		return "unknown"
	}
}

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
	fw.logger.Info("File watcher started")
	return nil
}

func (fw *FileWatcher) Stop() error {
	fw.cancel()
	err := fw.watcher.Close()
	fw.logger.Info("File watcher stopped")
	return err
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

func (fw *FileWatcher) RemoveDirectory(path string) error {
	return fw.watcher.Remove(path)
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
						// If a directory was created, watch it
						if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
							fw.AddDirectory(event.Name)
						}
					case event.Op&fsnotify.Write == fsnotify.Write:
						fileEvent.Type = EventTypeModify
					case event.Op&fsnotify.Remove == fsnotify.Remove:
						fileEvent.Type = EventTypeDelete
					case event.Op&fsnotify.Rename == fsnotify.Rename:
						fileEvent.Type = EventTypeDelete // Treat rename as delete for simplicity
					default:
						return
					}

					fw.logger.Debugf("Processing file event: %s %s", fileEvent.Type, fileEvent.Path)

					if err := fw.eventHandler(fileEvent); err != nil {
						fw.logger.Errorf("Event handler error for %s: %v", fileEvent.Path, err)
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

	// Always ignore hidden files and directories
	if strings.HasPrefix(base, ".") {
		return true
	}

	// Check if path contains .git directory
	if strings.Contains(path, "/.git/") || strings.Contains(path, "\\.git\\") {
		return true
	}

	// Check against ignore patterns
	for _, pattern := range fw.config.IgnorePatterns {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
		// Also check the full path
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		// Check if pattern contains directory separators and match path components
		if strings.Contains(pattern, "/") || strings.Contains(pattern, "\\") {
			// Convert pattern to work with paths
			cleanPattern := strings.ReplaceAll(pattern, "\\", "/")
			cleanPath := strings.ReplaceAll(path, "\\", "/")

			// Check if path matches pattern or contains pattern as a component
			if strings.Contains(cleanPath, cleanPattern) {
				return true
			}
		}
	}

	return false
}

func (fw *FileWatcher) GetWatchedDirectories() []string {
	var dirs []string
	fw.watchedDirs.Range(func(key, value interface{}) bool {
		if dir, ok := key.(string); ok {
			dirs = append(dirs, dir)
		}
		return true
	})
	return dirs
}

func (fw *FileWatcher) IsWatching(path string) bool {
	_, exists := fw.watchedDirs.Load(path)
	return exists
}
