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
	assert.Equal(t, "SheetSync", cfg.Git.UserName)
	assert.Equal(t, true, cfg.Converter.PreserveFormulas)
	assert.Equal(t, 1000000, cfg.Converter.MaxCellsPerSheet)
	assert.Contains(t, cfg.Watcher.FileExtensions, ".xlsx")
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sheetsync.yaml")
	
	configContent := `version: "2.0"
git:
  branch: develop
  auto_push: true
  user_name: "Test User"
converter:
  preserve_formulas: false
  max_cells_per_sheet: 5000
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
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