package adapter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWatcherAdapter(t *testing.T) {
	t.Run("creates new watcher adapter", func(t *testing.T) {
		cfg := &config.Config{
			Watcher: config.WatcherConfig{
				Directories:    []string{"."},
				IgnorePatterns: []string{"~$*"},
				DebounceDelay:  time.Second,
				FileExtensions: []string{".xlsx"},
			},
		}
		logger := logrus.New()
		logger.SetLevel(logrus.WarnLevel)

		eventReceived := false
		onEvent := func(event WatcherEvent) {
			eventReceived = true
		}

		adapter, err := NewWatcherAdapter(cfg, logger, onEvent)
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Equal(t, cfg, adapter.config)
		assert.False(t, adapter.isRunning)
		assert.False(t, eventReceived)
	})

	t.Run("creates adapter with nil logger", func(t *testing.T) {
		cfg := &config.Config{
			Watcher: config.WatcherConfig{
				Directories:    []string{"."},
				IgnorePatterns: []string{"~$*"},
				DebounceDelay:  time.Second,
				FileExtensions: []string{".xlsx"},
			},
		}

		adapter, err := NewWatcherAdapter(cfg, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.NotNil(t, adapter.logger) // Should create default logger
	})
}

func TestWatcherAdapter_GetStatus(t *testing.T) {
	t.Run("returns correct status when stopped", func(t *testing.T) {
		cfg := &config.Config{
			Watcher: config.WatcherConfig{
				Directories:    []string{"."},
				IgnorePatterns: []string{"~$*"},
				DebounceDelay:  time.Second,
				FileExtensions: []string{".xlsx"},
			},
		}
		logger := logrus.New()
		logger.SetLevel(logrus.WarnLevel)

		adapter, err := NewWatcherAdapter(cfg, logger, nil)
		require.NoError(t, err)

		status := adapter.GetStatus()
		assert.False(t, status.IsRunning)
		assert.Zero(t, status.StartTime)
		assert.Zero(t, status.DirectoriesWatched)
		assert.Empty(t, status.LastEvent)
	})
}

func TestWatcherAdapter_IsRunning(t *testing.T) {
	t.Run("returns false when not started", func(t *testing.T) {
		cfg := &config.Config{
			Watcher: config.WatcherConfig{
				Directories:    []string{"."},
				IgnorePatterns: []string{"~$*"},
				DebounceDelay:  time.Second,
				FileExtensions: []string{".xlsx"},
			},
		}
		logger := logrus.New()
		logger.SetLevel(logrus.WarnLevel)

		adapter, err := NewWatcherAdapter(cfg, logger, nil)
		require.NoError(t, err)

		assert.False(t, adapter.IsRunning())
	})
}

func TestWatcherAdapter_StartStop(t *testing.T) {
	t.Run("stops watcher when not running", func(t *testing.T) {
		cfg := &config.Config{
			Watcher: config.WatcherConfig{
				Directories:    []string{"."},
				IgnorePatterns: []string{"~$*"},
				DebounceDelay:  time.Second,
				FileExtensions: []string{".xlsx"},
			},
		}
		logger := logrus.New()
		logger.SetLevel(logrus.WarnLevel)

		adapter, err := NewWatcherAdapter(cfg, logger, nil)
		require.NoError(t, err)

		err = adapter.Stop()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not running")
	})

	t.Run("start and stop in git repository", func(t *testing.T) {
		// Create temporary directory with git init
		tmpDir, err := os.MkdirTemp("", "watcher_test_")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Initialize git repository
		gitDir := filepath.Join(tmpDir, ".git")
		err = os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		// Change to temp directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			_ = os.Chdir(originalDir)
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		cfg := &config.Config{
			Watcher: config.WatcherConfig{
				Directories:    []string{tmpDir},
				IgnorePatterns: []string{"~$*"},
				DebounceDelay:  time.Millisecond * 100,
				FileExtensions: []string{".xlsx"},
			},
			Git: config.GitConfig{
				UserName:  "Test User",
				UserEmail: "test@example.com",
			},
			Converter: config.ConverterConfig{
				PreserveFormulas: true,
				PreserveStyles:   true,
				PreserveComments: true,
				CompactJSON:      false,
				IgnoreEmptyCells: true,
			},
		}

		logger := logrus.New()
		logger.SetLevel(logrus.ErrorLevel)

		events := []WatcherEvent{}
		onEvent := func(event WatcherEvent) {
			events = append(events, event)
		}

		adapter, err := NewWatcherAdapter(cfg, logger, onEvent)
		require.NoError(t, err)

		// Start watcher
		err = adapter.Start()
		require.NoError(t, err)
		assert.True(t, adapter.IsRunning())

		// Wait a bit for initialization
		time.Sleep(time.Millisecond * 200)

		// Stop watcher
		err = adapter.Stop()
		require.NoError(t, err)
		assert.False(t, adapter.IsRunning())

		// Should have received start and stop events
		assert.GreaterOrEqual(t, len(events), 2)
	})
}

func TestWatcherEvent(t *testing.T) {
	t.Run("watcher event struct", func(t *testing.T) {
		event := WatcherEvent{
			Type:      "test",
			Message:   "test message",
			Details:   "test details",
			Timestamp: time.Now(),
			FilePath:  "/test/path",
		}

		assert.Equal(t, "test", event.Type)
		assert.Equal(t, "test message", event.Message)
		assert.Equal(t, "test details", event.Details)
		assert.Equal(t, "/test/path", event.FilePath)
		assert.NotZero(t, event.Timestamp)
	})
}

func TestWatcherStatus(t *testing.T) {
	t.Run("watcher status struct", func(t *testing.T) {
		status := WatcherStatus{
			IsRunning:          true,
			StartTime:          time.Now(),
			DirectoriesWatched: 5,
			LastEvent:          "test event",
			LastEventTime:      time.Now(),
			Directories:        []string{"/test/dir"},
		}

		assert.True(t, status.IsRunning)
		assert.Equal(t, 5, status.DirectoriesWatched)
		assert.Equal(t, "test event", status.LastEvent)
		assert.Len(t, status.Directories, 1)
	})
}
