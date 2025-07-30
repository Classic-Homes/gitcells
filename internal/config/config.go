// Package config provides configuration management for SheetSync application.
package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Version   string          `yaml:"version"`
	Git       GitConfig       `yaml:"git"`
	Watcher   WatcherConfig   `yaml:"watcher"`
	Converter ConverterConfig `yaml:"converter"`
}

type GitConfig struct {
	Remote         string `yaml:"remote"`
	Branch         string `yaml:"branch"`
	AutoPush       bool   `yaml:"auto_push"`
	AutoPull       bool   `yaml:"auto_pull"`
	UserName       string `yaml:"user_name"`
	UserEmail      string `yaml:"user_email"`
	CommitTemplate string `yaml:"commit_template"`
}

type WatcherConfig struct {
	Directories    []string      `yaml:"directories"`
	IgnorePatterns []string      `yaml:"ignore_patterns"`
	DebounceDelay  time.Duration `yaml:"debounce_delay"`
	FileExtensions []string      `yaml:"file_extensions"`
}

type ConverterConfig struct {
	PreserveFormulas bool `yaml:"preserve_formulas"`
	PreserveStyles   bool `yaml:"preserve_styles"`
	PreserveComments bool `yaml:"preserve_comments"`
	CompactJSON      bool `yaml:"compact_json"`
	IgnoreEmptyCells bool `yaml:"ignore_empty_cells"`
	MaxCellsPerSheet int  `yaml:"max_cells_per_sheet"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("version", "1.0")
	v.SetDefault("git.branch", "main")
	v.SetDefault("git.auto_push", false)
	v.SetDefault("git.auto_pull", true)
	v.SetDefault("git.user_name", "SheetSync")
	v.SetDefault("git.user_email", "sheetsync@localhost")
	v.SetDefault("git.commit_template", "SheetSync: {action} {filename} at {timestamp}")
	v.SetDefault("watcher.debounce_delay", "1s")
	v.SetDefault("watcher.file_extensions", []string{".xlsx", ".xls", ".xlsm"})
	v.SetDefault("watcher.ignore_patterns", []string{"~$*", "*.tmp"})
	v.SetDefault("converter.preserve_formulas", true)
	v.SetDefault("converter.preserve_styles", true)
	v.SetDefault("converter.preserve_comments", true)
	v.SetDefault("converter.compact_json", false)
	v.SetDefault("converter.ignore_empty_cells", true)
	v.SetDefault("converter.max_cells_per_sheet", 1000000)

	// Load config file
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	} else {
		v.SetConfigName(".sheetsync")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		// Try to read config, but don't error if not found (use defaults)
		_ = v.ReadInConfig()
	}

	// Manually construct config from viper values
	cfg := &Config{
		Version: v.GetString("version"),
		Git: GitConfig{
			Remote:         v.GetString("git.remote"),
			Branch:         v.GetString("git.branch"),
			AutoPush:       v.GetBool("git.auto_push"),
			AutoPull:       v.GetBool("git.auto_pull"),
			UserName:       v.GetString("git.user_name"),
			UserEmail:      v.GetString("git.user_email"),
			CommitTemplate: v.GetString("git.commit_template"),
		},
		Watcher: WatcherConfig{
			Directories:    v.GetStringSlice("watcher.directories"),
			IgnorePatterns: v.GetStringSlice("watcher.ignore_patterns"),
			DebounceDelay:  v.GetDuration("watcher.debounce_delay"),
			FileExtensions: v.GetStringSlice("watcher.file_extensions"),
		},
		Converter: ConverterConfig{
			PreserveFormulas: v.GetBool("converter.preserve_formulas"),
			PreserveStyles:   v.GetBool("converter.preserve_styles"),
			PreserveComments: v.GetBool("converter.preserve_comments"),
			CompactJSON:      v.GetBool("converter.compact_json"),
			IgnoreEmptyCells: v.GetBool("converter.ignore_empty_cells"),
			MaxCellsPerSheet: v.GetInt("converter.max_cells_per_sheet"),
		},
	}

	return cfg, nil
}

// ConvertOptions defines options for conversion (will be moved to converter package)
type ConvertOptions struct {
	PreserveFormulas bool
	PreserveStyles   bool
	PreserveComments bool
	CompactJSON      bool
	IgnoreEmptyCells bool
	MaxCellsPerSheet int
}

// ToOptions converts config to converter options
func (c *ConverterConfig) ToOptions() ConvertOptions {
	return ConvertOptions{
		PreserveFormulas: c.PreserveFormulas,
		PreserveStyles:   c.PreserveStyles,
		PreserveComments: c.PreserveComments,
		CompactJSON:      c.CompactJSON,
		IgnoreEmptyCells: c.IgnoreEmptyCells,
		MaxCellsPerSheet: c.MaxCellsPerSheet,
	}
}
