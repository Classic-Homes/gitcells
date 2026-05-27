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
	cfg, err := Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "1.0", cfg.Version)
	assert.Equal(t, "main", cfg.Git.Branch)
	assert.Equal(t, "GitCells", cfg.Git.UserName)
	assert.Equal(t, true, cfg.Converter.PreserveFormulas)
	assert.Equal(t, 1000000, cfg.Converter.MaxCellsPerSheet)
	assert.Contains(t, cfg.Watcher.FileExtensions, ".xlsx")
}

func TestLoadConfigFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".gitcells.yaml")

	configContent := `version: "2.0"
git:
  branch: develop
  user_name: "Test User"
converter:
  preserve_formulas: false
  max_cells_per_sheet: 5000
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0600))

	cfg, err := Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "2.0", cfg.Version)
	assert.Equal(t, "develop", cfg.Git.Branch)
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

func TestConfigSaveRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-config.yaml")

	cfg := GetDefault()
	cfg.Version = "3.0"
	cfg.Git.UserName = "Custom User"
	cfg.Converter.PreserveFormulas = false
	cfg.Converter.MaxCellsPerSheet = 5000

	require.NoError(t, cfg.Save(configPath))

	_, err := os.Stat(configPath)
	require.NoError(t, err)

	loaded, err := Load(configPath)
	require.NoError(t, err)
	assert.Equal(t, "3.0", loaded.Version)
	assert.Equal(t, "Custom User", loaded.Git.UserName)
	assert.Equal(t, false, loaded.Converter.PreserveFormulas)
	assert.Equal(t, 5000, loaded.Converter.MaxCellsPerSheet)
}
