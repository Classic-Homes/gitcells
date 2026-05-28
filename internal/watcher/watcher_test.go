package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileWatcher(t *testing.T) {
	config := &Config{
		IgnorePatterns: []string{"*.tmp", "~$*"},
		DebounceDelay:  100 * time.Millisecond,
		FileExtensions: []string{".xlsx", ".xls"},
	}

	eventChan := make(chan FileEvent)
	handler := func(event FileEvent) error {
		eventChan <- event
		return nil
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	fw, err := NewFileWatcher(config, handler, logger)
	require.NoError(t, err)
	assert.NotNil(t, fw)

	err = fw.Stop(0)
	assert.NoError(t, err)
}

func TestFileWatcher_AddDirectory(t *testing.T) {
	config := &Config{
		IgnorePatterns: []string{"*.tmp"},
		DebounceDelay:  100 * time.Millisecond,
		FileExtensions: []string{".xlsx"},
	}

	handler := func(event FileEvent) error { return nil }
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	fw, err := NewFileWatcher(config, handler, logger)
	require.NoError(t, err)
	defer func() { _ = fw.Stop(0) }()

	// Create a temporary directory
	tempDir := t.TempDir()

	// Create subdirectories
	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0750)
	require.NoError(t, err)

	// Add the directory to watcher
	err = fw.AddDirectory(tempDir)
	assert.NoError(t, err)

	// Test adding non-existent directory
	err = fw.AddDirectory("/path/that/does/not/exist")
	assert.Error(t, err)
}

// TestFileWatcher_AddDirectory_RootExempt confirms the watch root the
// user explicitly passes is never skipped by shouldIgnorePath, while
// hidden descendants (.git, .gitcells) still are. Without the exemption,
// `gitcells watch .` would silently watch nothing because base(".") is
// "." which trips the dot-prefix rule.
func TestFileWatcher_AddDirectory_RootExempt(t *testing.T) {
	config := &Config{
		IgnorePatterns: []string{},
		DebounceDelay:  100 * time.Millisecond,
		FileExtensions: []string{".xlsx"},
	}
	handler := func(event FileEvent) error { return nil }
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	fw, err := NewFileWatcher(config, handler, logger)
	require.NoError(t, err)
	defer func() { _ = fw.Stop(0) }()

	tempDir := t.TempDir()
	// Layout:
	//   tempDir/reports         (should be watched)
	//   tempDir/.git/objects    (should NOT be watched — hidden descendant)
	//   tempDir/.gitcells/data  (should NOT be watched — hidden descendant)
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "reports"), 0o750))
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, ".git", "objects"), 0o750))
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, ".gitcells", "data"), 0o750))

	require.NoError(t, fw.AddDirectory(tempDir))

	watched := fw.GetWatchedDirectories()
	hasRoot, hasReports := false, false
	for _, w := range watched {
		switch {
		case w == tempDir:
			hasRoot = true
		case w == filepath.Join(tempDir, "reports"):
			hasReports = true
		case w == filepath.Join(tempDir, ".git") || w == filepath.Join(tempDir, ".git", "objects"):
			t.Errorf(".git descendants must not be watched, but got: %s", w)
		case w == filepath.Join(tempDir, ".gitcells") || w == filepath.Join(tempDir, ".gitcells", "data"):
			t.Errorf(".gitcells descendants must not be watched, but got: %s", w)
		}
	}
	assert.True(t, hasRoot, "the explicitly-watched root must always be added")
	assert.True(t, hasReports, "non-hidden descendants must be added")
}

func TestFileWatcher_ShouldProcessFile(t *testing.T) {
	config := &Config{
		IgnorePatterns: []string{"*.tmp", "~$*"},
		DebounceDelay:  100 * time.Millisecond,
		FileExtensions: []string{".xlsx", ".xls", ".xlsm"},
	}

	handler := func(event FileEvent) error { return nil }
	logger := logrus.New()
	fw, err := NewFileWatcher(config, handler, logger)
	require.NoError(t, err)
	defer func() { _ = fw.Stop(0) }()

	tests := []struct {
		name     string
		filepath string
		expected bool
	}{
		{
			name:     "valid xlsx file",
			filepath: "/path/to/document.xlsx",
			expected: true,
		},
		{
			name:     "valid xls file",
			filepath: "/path/to/document.xls",
			expected: true,
		},
		{
			name:     "valid xlsm file",
			filepath: "/path/to/document.xlsm",
			expected: true,
		},
		{
			name:     "invalid extension",
			filepath: "/path/to/document.txt",
			expected: false,
		},
		{
			name:     "excel temp file",
			filepath: "/path/to/~$document.xlsx",
			expected: false,
		},
		{
			name:     "temp file pattern",
			filepath: "/path/to/document.tmp",
			expected: false,
		},
		{
			name:     "case insensitive extension",
			filepath: "/path/to/document.XLSX",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fw.shouldProcessFile(tt.filepath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileWatcher_ShouldIgnorePath(t *testing.T) {
	config := &Config{
		IgnorePatterns: []string{"*.tmp", "~$*", ".git/*"},
		DebounceDelay:  100 * time.Millisecond,
		FileExtensions: []string{".xlsx"},
	}

	handler := func(event FileEvent) error { return nil }
	logger := logrus.New()
	fw, err := NewFileWatcher(config, handler, logger)
	require.NoError(t, err)
	defer func() { _ = fw.Stop(0) }()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "normal file",
			path:     "/path/to/document.xlsx",
			expected: false,
		},
		{
			name:     "excel temp file",
			path:     "/path/to/~$document.xlsx",
			expected: true,
		},
		{
			name:     "tmp file",
			path:     "/path/to/backup.tmp",
			expected: true,
		},
		{
			name:     "git directory",
			path:     "/path/.git/config",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fw.shouldIgnorePath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileWatcher_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &Config{
		IgnorePatterns: []string{"*.tmp"},
		DebounceDelay:  200 * time.Millisecond,
		FileExtensions: []string{".xlsx"},
	}

	eventChan := make(chan FileEvent, 10)
	handler := func(event FileEvent) error {
		eventChan <- event
		return nil
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	fw, err := NewFileWatcher(config, handler, logger)
	require.NoError(t, err)
	defer func() { _ = fw.Stop(0) }()

	// Create a temporary directory
	tempDir := t.TempDir()

	// Add directory to watcher
	err = fw.AddDirectory(tempDir)
	require.NoError(t, err)

	// Start watching
	err = fw.Start()
	require.NoError(t, err)

	// Give watcher time to initialize
	time.Sleep(100 * time.Millisecond)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.xlsx")
	err = os.WriteFile(testFile, []byte("test content"), 0600)
	require.NoError(t, err)

	// Wait for event (with timeout)
	select {
	case event := <-eventChan:
		assert.Equal(t, testFile, event.Path)
		// In some environments, WriteFile triggers Write instead of Create
		assert.True(t, event.Type == EventTypeCreate || event.Type == EventTypeModify,
			"Expected Create or Modify event, got %v", event.Type)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for file creation event")
	}

	// Modify the file
	err = os.WriteFile(testFile, []byte("modified content"), 0600)
	require.NoError(t, err)

	// Wait for modify event
	select {
	case event := <-eventChan:
		assert.Equal(t, testFile, event.Path)
		assert.Equal(t, EventTypeModify, event.Type)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for file modification event")
	}

	// Delete the file
	err = os.Remove(testFile)
	require.NoError(t, err)

	// Wait for delete event
	select {
	case event := <-eventChan:
		assert.Equal(t, testFile, event.Path)
		assert.Equal(t, EventTypeDelete, event.Type)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for file deletion event")
	}
}

func TestFileWatcher_ContextCancellation(t *testing.T) {
	config := &Config{
		IgnorePatterns: []string{},
		DebounceDelay:  100 * time.Millisecond,
		FileExtensions: []string{".xlsx"},
	}

	handler := func(event FileEvent) error { return nil }
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	fw, err := NewFileWatcher(config, handler, logger)
	require.NoError(t, err)

	// Start watching
	err = fw.Start()
	require.NoError(t, err)

	// Stop should work without errors
	err = fw.Stop(0)
	assert.NoError(t, err)

	// Second stop should also work
	err = fw.Stop(0)
	assert.NoError(t, err)
}

// TestFileWatcher_AtomicSaveSequence simulates the event sequence Excel
// produces during a save: write a lock file (~$book.xlsx), write the
// payload to a temp path, rename it over the target, then remove the
// lock file. The watcher must:
//   - ignore the ~$ lock file at filter time (no event),
//   - coalesce the temp-write + rename + lock-remove into a single
//     EventTypeModify for book.xlsx (debounce + stat-based classifier).
//
// Runs on Linux/macOS/Windows. Excel's exact sequence varies per OS
// (FSEvents coalesces on macOS, Windows splits Rename); the post-debounce
// stat classifier at watcher.go normalizes them.
func TestFileWatcher_AtomicSaveSequence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration-style watcher test in short mode")
	}

	config := &Config{
		IgnorePatterns: []string{"~$*"},
		DebounceDelay:  150 * time.Millisecond,
		FileExtensions: []string{".xlsx"},
	}

	eventChan := make(chan FileEvent, 8)
	handler := func(event FileEvent) error {
		eventChan <- event
		return nil
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	fw, err := NewFileWatcher(config, handler, logger)
	require.NoError(t, err)
	defer func() { _ = fw.Stop(0) }()

	tempDir := t.TempDir()
	require.NoError(t, fw.AddDirectory(tempDir))
	require.NoError(t, fw.Start())
	time.Sleep(100 * time.Millisecond) // let fsnotify warm up

	target := filepath.Join(tempDir, "book.xlsx")
	lock := filepath.Join(tempDir, "~$book.xlsx")
	staging := filepath.Join(tempDir, ".book.tmp.xlsx")

	// Excel atomic-save sequence:
	// 1. Create lock file. Must produce no event (ignored at filter time).
	require.NoError(t, os.WriteFile(lock, []byte("lock"), 0600))
	// 2. Write the new content to a staging path. Hidden-dot prefix means
	//    shouldIgnorePath drops it; no event expected.
	require.NoError(t, os.WriteFile(staging, []byte("payload"), 0600))
	// 3. Rename staging over the target. This is the visible publish step.
	require.NoError(t, os.Rename(staging, target))
	// 4. Remove the lock file. Filtered.
	require.NoError(t, os.Remove(lock))

	// Collect events for ~3× debounce so coalescing has a fair window.
	deadline := time.After(3 * config.DebounceDelay)
	var got []FileEvent
collect:
	for {
		select {
		case ev := <-eventChan:
			got = append(got, ev)
		case <-deadline:
			break collect
		}
	}

	// Exactly one event on the target file, classified as Modify.
	var targetEvents []FileEvent
	for _, ev := range got {
		assert.NotContains(t, filepath.Base(ev.Path), "~$",
			"lock file events must be filtered: got %s", ev.Path)
		assert.NotContains(t, filepath.Base(ev.Path), ".tmp.",
			"staging file events must be filtered: got %s", ev.Path)
		if ev.Path == target {
			targetEvents = append(targetEvents, ev)
		}
	}
	require.Len(t, targetEvents, 1, "expected exactly one event for %s, got %d (all events: %+v)", target, len(targetEvents), got)
	assert.Equal(t, EventTypeModify, targetEvents[0].Type,
		"atomic-save sequence must classify as Modify, not Delete")

	// Second pass: explicit Remove of the target should classify as Delete.
	for len(eventChan) > 0 {
		<-eventChan
	}
	require.NoError(t, os.Remove(target))

	select {
	case ev := <-eventChan:
		assert.Equal(t, target, ev.Path)
		assert.Equal(t, EventTypeDelete, ev.Type)
	case <-time.After(3 * config.DebounceDelay):
		t.Fatal("timeout waiting for delete event")
	}
}

func TestFileWatcher_HandlerError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &Config{
		IgnorePatterns: []string{},
		DebounceDelay:  100 * time.Millisecond,
		FileExtensions: []string{".xlsx"},
	}

	// Handler that returns an error
	errorHandler := func(event FileEvent) error {
		return assert.AnError
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	fw, err := NewFileWatcher(config, errorHandler, logger)
	require.NoError(t, err)
	defer func() { _ = fw.Stop(0) }()

	tempDir := t.TempDir()
	err = fw.AddDirectory(tempDir)
	require.NoError(t, err)

	err = fw.Start()
	require.NoError(t, err)

	// Create a file that should trigger the error handler
	testFile := filepath.Join(tempDir, "test.xlsx")
	err = os.WriteFile(testFile, []byte("test"), 0600)
	require.NoError(t, err)

	// Wait a bit for processing
	time.Sleep(500 * time.Millisecond)

	// The watcher should continue running even if handler returns error
	// This is hard to test directly, but we can verify the watcher is still responsive
	err = fw.Stop(0)
	assert.NoError(t, err)
}
