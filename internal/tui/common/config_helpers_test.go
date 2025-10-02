package common

import (
	"testing"
	"time"

	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetStringValue(t *testing.T) {
	cfg := &config.Config{}

	tests := []struct {
		name       string
		key        string
		value      string
		wantErr    bool
		checkValue func(*testing.T, *config.Config)
	}{
		{
			name:    "set git remote",
			key:     "git.remote",
			value:   "origin",
			wantErr: false,
			checkValue: func(t *testing.T, c *config.Config) {
				assert.Equal(t, "origin", c.Git.Remote)
			},
		},
		{
			name:    "set git branch",
			key:     "git.branch",
			value:   "main",
			wantErr: false,
			checkValue: func(t *testing.T, c *config.Config) {
				assert.Equal(t, "main", c.Git.Branch)
			},
		},
		{
			name:    "set user name",
			key:     "git.user_name",
			value:   "Test User",
			wantErr: false,
			checkValue: func(t *testing.T, c *config.Config) {
				assert.Equal(t, "Test User", c.Git.UserName)
			},
		},
		{
			name:    "unknown key",
			key:     "unknown.key",
			value:   "value",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetStringValue(cfg, tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkValue != nil {
					tt.checkValue(t, cfg)
				}
			}
		})
	}
}

func TestSetBoolValue(t *testing.T) {
	cfg := &config.Config{}

	tests := []struct {
		name       string
		key        string
		value      bool
		wantErr    bool
		checkValue func(*testing.T, *config.Config)
	}{
		{
			name:    "enable auto push",
			key:     "git.auto_push",
			value:   true,
			wantErr: false,
			checkValue: func(t *testing.T, c *config.Config) {
				assert.True(t, c.Git.AutoPush)
			},
		},
		{
			name:    "disable preserve formulas",
			key:     "converter.preserve_formulas",
			value:   false,
			wantErr: false,
			checkValue: func(t *testing.T, c *config.Config) {
				assert.False(t, c.Converter.PreserveFormulas)
			},
		},
		{
			name:    "enable telemetry",
			key:     "features.telemetry",
			value:   true,
			wantErr: false,
			checkValue: func(t *testing.T, c *config.Config) {
				assert.True(t, c.Features.EnableTelemetry)
			},
		},
		{
			name:    "unknown key",
			key:     "unknown.bool",
			value:   true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetBoolValue(cfg, tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkValue != nil {
					tt.checkValue(t, cfg)
				}
			}
		})
	}
}

func TestSetDurationValue(t *testing.T) {
	cfg := &config.Config{}

	tests := []struct {
		name       string
		key        string
		value      string
		wantErr    bool
		checkValue func(*testing.T, *config.Config)
	}{
		{
			name:    "set debounce delay",
			key:     "watcher.debounce_delay",
			value:   "2s",
			wantErr: false,
			checkValue: func(t *testing.T, c *config.Config) {
				assert.Equal(t, 2*time.Second, c.Watcher.DebounceDelay)
			},
		},
		{
			name:    "set check interval",
			key:     "updates.check_interval",
			value:   "24h",
			wantErr: false,
			checkValue: func(t *testing.T, c *config.Config) {
				assert.Equal(t, 24*time.Hour, c.Updates.CheckInterval)
			},
		},
		{
			name:    "invalid duration",
			key:     "watcher.debounce_delay",
			value:   "invalid",
			wantErr: true,
		},
		{
			name:    "unknown key",
			key:     "unknown.duration",
			value:   "1s",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetDurationValue(cfg, tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkValue != nil {
					tt.checkValue(t, cfg)
				}
			}
		})
	}
}

func TestSetIntValue(t *testing.T) {
	cfg := &config.Config{}

	tests := []struct {
		name       string
		key        string
		value      string
		wantErr    bool
		checkValue func(*testing.T, *config.Config)
	}{
		{
			name:    "set max cells",
			key:     "converter.max_cells_per_sheet",
			value:   "1000000",
			wantErr: false,
			checkValue: func(t *testing.T, c *config.Config) {
				assert.Equal(t, 1000000, c.Converter.MaxCellsPerSheet)
			},
		},
		{
			name:    "invalid int",
			key:     "converter.max_cells_per_sheet",
			value:   "abc",
			wantErr: true,
		},
		{
			name:    "negative value",
			key:     "converter.max_cells_per_sheet",
			value:   "0",
			wantErr: true,
		},
		{
			name:    "unknown key",
			key:     "unknown.int",
			value:   "123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetIntValue(cfg, tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkValue != nil {
					tt.checkValue(t, cfg)
				}
			}
		})
	}
}

func TestSetStringSliceValue(t *testing.T) {
	cfg := &config.Config{}

	tests := []struct {
		name       string
		key        string
		value      string
		wantErr    bool
		checkValue func(*testing.T, *config.Config)
	}{
		{
			name:    "set directories",
			key:     "watcher.directories",
			value:   "./dir1, ./dir2, ./dir3",
			wantErr: false,
			checkValue: func(t *testing.T, c *config.Config) {
				assert.Equal(t, []string{"./dir1", "./dir2", "./dir3"}, c.Watcher.Directories)
			},
		},
		{
			name:    "set ignore patterns",
			key:     "watcher.ignore_patterns",
			value:   "*.tmp, ~$*",
			wantErr: false,
			checkValue: func(t *testing.T, c *config.Config) {
				assert.Equal(t, []string{"*.tmp", "~$*"}, c.Watcher.IgnorePatterns)
			},
		},
		{
			name:    "set file extensions",
			key:     "watcher.file_extensions",
			value:   ".xlsx,.xls,.xlsm",
			wantErr: false,
			checkValue: func(t *testing.T, c *config.Config) {
				assert.Equal(t, []string{".xlsx", ".xls", ".xlsm"}, c.Watcher.FileExtensions)
			},
		},
		{
			name:    "unknown key",
			key:     "unknown.slice",
			value:   "a,b,c",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetStringSliceValue(cfg, tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkValue != nil {
					tt.checkValue(t, cfg)
				}
			}
		})
	}
}

func TestToggleBoolValue(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitConfig{
			AutoPush: false,
		},
	}

	// Toggle from false to true
	err := ToggleBoolValue(cfg, "git.auto_push")
	require.NoError(t, err)
	assert.True(t, cfg.Git.AutoPush)

	// Toggle from true to false
	err = ToggleBoolValue(cfg, "git.auto_push")
	require.NoError(t, err)
	assert.False(t, cfg.Git.AutoPush)

	// Unknown key
	err = ToggleBoolValue(cfg, "unknown.key")
	assert.Error(t, err)
}

func TestGetBoolValue(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitConfig{
			AutoPush: true,
		},
		Features: config.FeaturesConfig{
			EnableTelemetry: false,
		},
	}

	value, err := GetBoolValue(cfg, "git.auto_push")
	require.NoError(t, err)
	assert.True(t, value)

	value, err = GetBoolValue(cfg, "features.telemetry")
	require.NoError(t, err)
	assert.False(t, value)

	_, err = GetBoolValue(cfg, "unknown.key")
	assert.Error(t, err)
}

func TestGetStringValue(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitConfig{
			Branch:   "main",
			UserName: "Test User",
		},
	}

	value, err := GetStringValue(cfg, "git.branch")
	require.NoError(t, err)
	assert.Equal(t, "main", value)

	value, err = GetStringValue(cfg, "git.user_name")
	require.NoError(t, err)
	assert.Equal(t, "Test User", value)

	_, err = GetStringValue(cfg, "unknown.key")
	assert.Error(t, err)
}
