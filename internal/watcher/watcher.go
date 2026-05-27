// Package watcher provides cross-platform file system monitoring for
// Excel files. It is the entry point for the watch → convert → commit
// pipeline and must behave consistently on Windows, macOS, and Linux.
//
// Filesystem assumptions this package relies on (and that the next
// maintainer should NOT regress):
//
//   - Excel saves atomically on every OS. The full sequence is roughly:
//     write `~$tmp.xlsx` → rename over the target → unlink the temp.
//     fsnotify emits Create + Write + Rename + Remove for what the user
//     experiences as a single "save". The debounce window (cfg.Watcher.
//     DebounceDelay, default 2s) must exceed typical save duration so all
//     of those events collapse into one handler call.
//
//   - Classifying by the *last* fsnotify Op bit produces false deletes
//     (an atomic-replace Rename arrives last). We instead os.Stat the
//     path at the end of the debounce window: present → Modify, gone →
//     Delete. See processEvents.
//
//   - fsnotify quirks per OS:
//
//   - Linux (inotify) coalesces rapid events; ordering is preserved
//     but counts collapse.
//
//   - macOS (FSEvents) is lossy — fsnotify masks this by polling, but
//     counts collapse and timestamps lag. Debounce window absorbs it.
//
//   - Windows (ReadDirectoryChangesW) splits a Rename into a CREATE on
//     the destination and a REMOVE on the source. We treat REMOVE on
//     a still-present path as a no-op via the stat check above.
//
//   - Excel lock files (`~$*.xlsx`, `.~lock.*`) are ignored at filter
//     time. Dot-files and anything under `.git/` are also skipped.
package watcher

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Classic-Homes/gitcells/internal/utils"
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
		return nil, utils.WrapError(err, utils.ErrorTypeWatcher, "NewFileWatcher", "failed to create fsnotify watcher")
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
	// Cancel pending debounced callbacks first. Otherwise an in-flight
	// time.AfterFunc may fire after Close() and call the event handler against
	// a torn-down dashboard / closed fsnotify channels.
	fw.debouncer.CancelAll()
	fw.cancel()
	err := fw.watcher.Close()
	fw.logger.Info("File watcher stopped")
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeWatcher, "Stop", "failed to close watcher")
	}
	return nil
}

func (fw *FileWatcher) AddDirectory(path string) error {
	// Walk directory tree and add all subdirectories
	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return utils.WrapFileError(err, utils.ErrorTypeFileSystem, "AddDirectory", walkPath, "failed to access path during walk")
		}

		if info.IsDir() {
			if fw.shouldIgnorePath(walkPath) {
				return filepath.SkipDir
			}

			// LoadOrStore makes the recursive walk idempotent. fsnotify itself
			// tolerates a duplicate Add, but we don't want to redundantly call
			// it or emit a duplicate debug line for the same directory.
			if _, loaded := fw.watchedDirs.LoadOrStore(walkPath, true); loaded {
				return nil
			}
			if err := fw.watcher.Add(walkPath); err != nil {
				fw.watchedDirs.Delete(walkPath)
				fw.logger.Warnf("Failed to watch %s: %v", walkPath, err)
			} else {
				fw.logger.Debugf("Watching directory: %s", walkPath)
			}
		}

		return nil
	})

	if err != nil {
		return utils.WrapFileError(err, utils.ErrorTypeWatcher, "AddDirectory", path, "failed to walk directory tree")
	}
	return nil
}

func (fw *FileWatcher) RemoveDirectory(path string) error {
	err := fw.watcher.Remove(path)
	if err != nil {
		return utils.WrapFileError(err, utils.ErrorTypeWatcher, "RemoveDirectory", path, "failed to remove directory from watcher")
	}
	return nil
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

					// Classify by current filesystem state, not by the last
					// fsnotify op bit. Excel saves atomically on every OS
					// (write-to-temp, rename-over-target, remove-temp), which
					// fires Create+Write+Rename+Remove for the *same* logical
					// "modify". Trusting the last op produces false deletes.
					//
					// If the file exists now (after the debounce window), it
					// was created or modified — treat as Modify (Create is just
					// a special-case Modify from a converter's perspective).
					// If gone, it was deleted or renamed away.
					info, statErr := os.Stat(event.Name)
					switch {
					case statErr == nil && info.IsDir():
						// New directory appeared inside a watched tree.
						_ = fw.AddDirectory(event.Name)
						return
					case statErr == nil:
						fileEvent.Type = EventTypeModify
					case os.IsNotExist(statErr):
						// Only emit Delete if the originating op was actually
						// a remove/rename. A spurious Stat miss on a write
						// (rare, but possible mid-rename on Windows) should
						// not nuke chunks.
						if event.Op&(fsnotify.Remove|fsnotify.Rename) == 0 {
							return
						}
						fileEvent.Type = EventTypeDelete
					default:
						fw.logger.Debugf("Skipping event for %s: stat error: %v", event.Name, statErr)
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

	// Check if path contains .git directory (normalize path separators for cross-platform)
	cleanPath := filepath.ToSlash(path)
	if strings.Contains(cleanPath, "/.git/") {
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
		if strings.Contains(pattern, "/") || strings.Contains(pattern, string(filepath.Separator)) {
			// Normalize both pattern and path to forward slashes for consistent comparison
			cleanPattern := filepath.ToSlash(pattern)

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
