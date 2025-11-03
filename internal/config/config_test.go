package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Test loading with empty config path (should use defaults)
	cfg, err := Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check default values
	assert.Equal(t, "1.0", cfg.Version)
	assert.Equal(t, "main", cfg.Git.Branch)
	assert.Equal(t, false, cfg.Git.AutoPush)
	assert.Equal(t, true, cfg.Git.AutoPull)
	assert.Equal(t, "GitCells", cfg.Git.UserName)
	assert.Equal(t, true, cfg.Converter.PreserveFormulas)
	assert.Equal(t, 1000000, cfg.Converter.MaxCellsPerSheet)
	assert.Contains(t, cfg.Watcher.FileExtensions, ".xlsx")
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gitcells.yaml")

	configContent := `version: "2.0"
git:
  branch: develop
  auto_push: true
  user_name: "Test User"
converter:
  preserve_formulas: false
  max_cells_per_sheet: 5000
`

	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Load config
	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check loaded values
	assert.Equal(t, "2.0", cfg.Version)
	assert.Equal(t, "develop", cfg.Git.Branch)
	assert.Equal(t, true, cfg.Git.AutoPush)
	assert.Equal(t, "Test User", cfg.Git.UserName)
	assert.Equal(t, false, cfg.Converter.PreserveFormulas)
	assert.Equal(t, 5000, cfg.Converter.MaxCellsPerSheet)
}

func TestGetDefault(t *testing.T) {
	cfg := GetDefault()
	require.NotNil(t, cfg)

	assert.Equal(t, "1.0", cfg.Version)
	assert.Equal(t, "main", cfg.Git.Branch)
	assert.Equal(t, 2*time.Second, cfg.Watcher.DebounceDelay)
	assert.Len(t, cfg.Watcher.FileExtensions, 3)
	assert.Contains(t, cfg.Watcher.FileExtensions, ".xlsx")
}

func TestToConverterOptions(t *testing.T) {
	cfg := &ConverterConfig{
		PreserveFormulas: true,
		PreserveStyles:   false,
		PreserveComments: true,
		CompactJSON:      false,
		IgnoreEmptyCells: true,
		MaxCellsPerSheet: 10000,
	}

	opts := cfg.ToOptions()

	assert.Equal(t, cfg.PreserveFormulas, opts.PreserveFormulas)
	assert.Equal(t, cfg.PreserveStyles, opts.PreserveStyles)
	assert.Equal(t, cfg.PreserveComments, opts.PreserveComments)
	assert.Equal(t, cfg.CompactJSON, opts.CompactJSON)
	assert.Equal(t, cfg.IgnoreEmptyCells, opts.IgnoreEmptyCells)
	assert.Equal(t, cfg.MaxCellsPerSheet, opts.MaxCellsPerSheet)
}

func TestConfigSave(t *testing.T) {
	t.Run("saves config to default path", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }()

		require.NoError(t, os.Chdir(tmpDir))

		cfg := GetDefault()
		cfg.Version = "2.0"
		cfg.Git.Branch = "develop"
		cfg.Converter.MaxCellsPerSheet = 5000

		err = cfg.Save("")
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(".gitcells.yaml")
		require.NoError(t, err)

		// Load and verify
		loadedCfg, err := Load(".gitcells.yaml")
		require.NoError(t, err)
		assert.Equal(t, "2.0", loadedCfg.Version)
		assert.Equal(t, "develop", loadedCfg.Git.Branch)
		assert.Equal(t, 5000, loadedCfg.Converter.MaxCellsPerSheet)
	})

	t.Run("saves config to specified path", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "custom-config.yaml")

		cfg := GetDefault()
		cfg.Version = "3.0"
		cfg.Git.UserName = "Custom User"
		cfg.Converter.PreserveFormulas = false

		err := cfg.Save(configPath)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(configPath)
		require.NoError(t, err)

		// Load and verify
		loadedCfg, err := Load(configPath)
		require.NoError(t, err)
		assert.Equal(t, "3.0", loadedCfg.Version)
		assert.Equal(t, "Custom User", loadedCfg.Git.UserName)
		assert.Equal(t, false, loadedCfg.Converter.PreserveFormulas)
	})

	t.Run("saves all config fields correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "full-config.yaml")

		cfg := &Config{
			Version: "4.0",
			Git: GitConfig{
				Remote:         "origin",
				Branch:         "feature-branch",
				AutoPush:       true,
				AutoPull:       false,
				UserName:       "Full Test",
				UserEmail:      "full@test.com",
				CommitTemplate: "Custom: {action}",
			},
			Watcher: WatcherConfig{
				Directories:    []string{"dir1", "dir2"},
				IgnorePatterns: []string{"*.tmp", "~$*"},
				DebounceDelay:  3 * time.Second,
				FileExtensions: []string{".xlsx", ".xls"},
			},
			Converter: ConverterConfig{
				PreserveFormulas: false,
				PreserveStyles:   true,
				PreserveComments: false,
				CompactJSON:      true,
				IgnoreEmptyCells: false,
				MaxCellsPerSheet: 999,
				ChunkingStrategy: "size-based",
			},
			Features: FeaturesConfig{
				EnableExperimentalFeatures: true,
				EnableBetaUpdates:          true,
				EnableTelemetry:            false,
			},
			Updates: UpdatesConfig{
				AutoCheckUpdates:    false,
				CheckInterval:       12 * time.Hour,
				IncludePrereleases:  true,
				AutoDownloadUpdates: true,
				NotifyOnUpdate:      false,
			},
		}

		err := cfg.Save(configPath)
		require.NoError(t, err)

		// Load and verify all fields
		loadedCfg, err := Load(configPath)
		require.NoError(t, err)

		assert.Equal(t, cfg.Version, loadedCfg.Version)
		assert.Equal(t, cfg.Git.Remote, loadedCfg.Git.Remote)
		assert.Equal(t, cfg.Git.Branch, loadedCfg.Git.Branch)
		assert.Equal(t, cfg.Git.AutoPush, loadedCfg.Git.AutoPush)
		assert.Equal(t, cfg.Git.AutoPull, loadedCfg.Git.AutoPull)
		assert.Equal(t, cfg.Git.UserName, loadedCfg.Git.UserName)
		assert.Equal(t, cfg.Git.UserEmail, loadedCfg.Git.UserEmail)
		assert.Equal(t, cfg.Git.CommitTemplate, loadedCfg.Git.CommitTemplate)
		assert.Equal(t, cfg.Watcher.Directories, loadedCfg.Watcher.Directories)
		assert.Equal(t, cfg.Watcher.IgnorePatterns, loadedCfg.Watcher.IgnorePatterns)
		assert.Equal(t, cfg.Watcher.DebounceDelay, loadedCfg.Watcher.DebounceDelay)
		assert.Equal(t, cfg.Watcher.FileExtensions, loadedCfg.Watcher.FileExtensions)
		assert.Equal(t, cfg.Converter.PreserveFormulas, loadedCfg.Converter.PreserveFormulas)
		assert.Equal(t, cfg.Converter.PreserveStyles, loadedCfg.Converter.PreserveStyles)
		assert.Equal(t, cfg.Converter.PreserveComments, loadedCfg.Converter.PreserveComments)
		assert.Equal(t, cfg.Converter.CompactJSON, loadedCfg.Converter.CompactJSON)
		assert.Equal(t, cfg.Converter.IgnoreEmptyCells, loadedCfg.Converter.IgnoreEmptyCells)
		assert.Equal(t, cfg.Converter.MaxCellsPerSheet, loadedCfg.Converter.MaxCellsPerSheet)
		assert.Equal(t, cfg.Converter.ChunkingStrategy, loadedCfg.Converter.ChunkingStrategy)
		assert.Equal(t, cfg.Features.EnableExperimentalFeatures, loadedCfg.Features.EnableExperimentalFeatures)
		assert.Equal(t, cfg.Features.EnableBetaUpdates, loadedCfg.Features.EnableBetaUpdates)
		assert.Equal(t, cfg.Features.EnableTelemetry, loadedCfg.Features.EnableTelemetry)
		assert.Equal(t, cfg.Updates.AutoCheckUpdates, loadedCfg.Updates.AutoCheckUpdates)
		assert.Equal(t, cfg.Updates.CheckInterval, loadedCfg.Updates.CheckInterval)
		assert.Equal(t, cfg.Updates.IncludePrereleases, loadedCfg.Updates.IncludePrereleases)
		assert.Equal(t, cfg.Updates.AutoDownloadUpdates, loadedCfg.Updates.AutoDownloadUpdates)
		assert.Equal(t, cfg.Updates.NotifyOnUpdate, loadedCfg.Updates.NotifyOnUpdate)
	})
}
