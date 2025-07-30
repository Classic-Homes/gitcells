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

	err = fw.Stop()
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
	defer fw.Stop()

	// Create a temporary directory
	tempDir := t.TempDir()

	// Create subdirectories
	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	// Add the directory to watcher
	err = fw.AddDirectory(tempDir)
	assert.NoError(t, err)

	// Test adding non-existent directory
	err = fw.AddDirectory("/path/that/does/not/exist")
	assert.Error(t, err)
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
	defer fw.Stop()

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
	defer fw.Stop()

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
	defer fw.Stop()

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
	err = os.WriteFile(testFile, []byte("test content"), 0644)
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
	err = os.WriteFile(testFile, []byte("modified content"), 0644)
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
	err = fw.Stop()
	assert.NoError(t, err)

	// Second stop should also work
	err = fw.Stop()
	assert.NoError(t, err)
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
	defer fw.Stop()

	tempDir := t.TempDir()
	err = fw.AddDirectory(tempDir)
	require.NoError(t, err)

	err = fw.Start()
	require.NoError(t, err)

	// Create a file that should trigger the error handler
	testFile := filepath.Join(tempDir, "test.xlsx")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Wait a bit for processing
	time.Sleep(500 * time.Millisecond)

	// The watcher should continue running even if handler returns error
	// This is hard to test directly, but we can verify the watcher is still responsive
	err = fw.Stop()
	assert.NoError(t, err)
}
