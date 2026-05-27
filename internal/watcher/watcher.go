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
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// DefaultShutdownGrace is how long Stop will wait for in-flight conversions
// to finish before force-cancelling. Excel files routinely take >2s to
// convert; 30s is a conservative ceiling that still bounds shutdown.
const DefaultShutdownGrace = 30 * time.Second

// WatcherError carries a failure surfaced from the event-handling pipeline.
// Consumers should drain Errors() in a separate goroutine; the channel is
// buffered but not unbounded.
type WatcherError struct {
	Path      string
	EventType EventType
	Err       error
	Timestamp time.Time
}

type FileWatcher struct {
	watcher      *fsnotify.Watcher
	watchedDirs  sync.Map
	eventHandler EventHandler
	debouncer    *Debouncer
	config       *Config
	logger       *logrus.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	errors       chan WatcherError
	stopOnce     sync.Once
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
		errors:       make(chan WatcherError, 64),
	}

	fw.debouncer.SetPanicLogger(func(r interface{}, stack []byte) {
		logger.Errorf("panic in debounced handler: %v\n%s", r, stack)
	})

	return fw, nil
}

// Errors returns a channel of failures surfaced from the event pipeline.
// Consumers must drain it; the channel is buffered (64) and additional
// failures past that point will be logged but dropped from the channel.
func (fw *FileWatcher) Errors() <-chan WatcherError {
	return fw.errors
}

func (fw *FileWatcher) Start() error {
	go fw.processEvents()
	fw.logger.Info("File watcher started")
	return nil
}

// Stop closes the watcher gracefully. It waits up to grace for any
// in-flight debounced conversions to complete, then force-cancels the
// rest. Dropped paths are logged at warning level. Pass 0 for an
// immediate stop (used by tests).
func (fw *FileWatcher) Stop(grace time.Duration) error {
	var (
		dropped  []string
		closeErr error
	)
	fw.stopOnce.Do(func() {
		dropped = fw.debouncer.Drain(grace)
		fw.cancel()
		closeErr = fw.watcher.Close()
		close(fw.errors)
		fw.logger.Info("File watcher stopped")
	})

	for _, path := range dropped {
		fw.logger.Warnf("Shutdown grace expired before processing: %s", path)
	}

	if closeErr != nil {
		return utils.WrapError(closeErr, utils.ErrorTypeWatcher, "Stop", "failed to close watcher")
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
	defer func() {
		if r := recover(); r != nil {
			fw.logger.Errorf("panic in processEvents: %v\n%s", r, debug.Stack())
		}
	}()

	for {
		select {
		case <-fw.ctx.Done():
			return

		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			if fw.shouldProcessFile(event.Name) {
				op := event.Op
				path := event.Name
				fw.debouncer.Debounce(path, func() {
					fileEvent := FileEvent{
						Path:      path,
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
					info, statErr := os.Stat(path)
					switch {
					case statErr == nil && info.IsDir():
						// New directory appeared inside a watched tree.
						_ = fw.AddDirectory(path)
						return
					case statErr == nil:
						fileEvent.Type = EventTypeModify
					case os.IsNotExist(statErr):
						// Only emit Delete if the originating op was actually
						// a remove/rename. A spurious Stat miss on a write
						// (rare, but possible mid-rename on Windows) should
						// not nuke chunks.
						if op&(fsnotify.Remove|fsnotify.Rename) == 0 {
							return
						}
						fileEvent.Type = EventTypeDelete
					default:
						fw.logger.Debugf("Skipping event for %s: stat error: %v", path, statErr)
						return
					}

					fw.logger.Debugf("Processing file event: %s %s", fileEvent.Type, fileEvent.Path)

					if err := fw.eventHandler(fileEvent); err != nil {
						fw.surfaceError(fileEvent.Path, fileEvent.Type, err)
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

// surfaceError publishes a failure to the Errors channel and always
// logs it. If the channel is full (consumer fell behind), the error
// is logged but dropped from the channel to keep the watcher
// non-blocking. The channel is closed only by Stop, so a closed-channel
// send after Stop is prevented by the ctx.Done check.
func (fw *FileWatcher) surfaceError(path string, eventType EventType, err error) {
	fw.logger.Errorf("Event handler error for %s: %v", path, err)
	we := WatcherError{
		Path:      path,
		EventType: eventType,
		Err:       err,
		Timestamp: time.Now(),
	}
	select {
	case <-fw.ctx.Done():
		return
	case fw.errors <- we:
	default:
		fw.logger.Warnf("Errors channel full; dropped error for %s", path)
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

	// Normalize once so pattern matching is identical on Windows and Unix.
	// fsnotify hands us backslashes on Windows; config patterns are
	// conventionally forward-slash style.
	cleanPath := filepath.ToSlash(path)

	if strings.Contains(cleanPath, "/.git/") {
		return true
	}

	for _, pattern := range fw.config.IgnorePatterns {
		cleanPattern := filepath.ToSlash(pattern)
		if matched, _ := filepath.Match(cleanPattern, base); matched {
			return true
		}
		if matched, _ := filepath.Match(cleanPattern, cleanPath); matched {
			return true
		}
		// Substring fall-through for patterns that name a directory
		// component rather than a glob (e.g. "build/" or "vendor").
		if strings.ContainsAny(cleanPattern, "/") && strings.Contains(cleanPath, cleanPattern) {
			return true
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
