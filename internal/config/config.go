// Package config provides configuration management for GitCells application.
package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	// DefaultMaxCellsPerSheet is the default maximum number of cells per sheet
	DefaultMaxCellsPerSheet = 1000000
)

type Config struct {
	Version   string          `yaml:"version"`
	Git       GitConfig       `yaml:"git"`
	Watcher   WatcherConfig   `yaml:"watcher"`
	Converter ConverterConfig `yaml:"converter"`
	Features  FeaturesConfig  `yaml:"features"`
	Updates   UpdatesConfig   `yaml:"updates"`
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
	PreserveFormulas bool   `yaml:"preserve_formulas"`
	PreserveStyles   bool   `yaml:"preserve_styles"`
	PreserveComments bool   `yaml:"preserve_comments"`
	CompactJSON      bool   `yaml:"compact_json"`
	IgnoreEmptyCells bool   `yaml:"ignore_empty_cells"`
	MaxCellsPerSheet int    `yaml:"max_cells_per_sheet"`
	ChunkingStrategy string `yaml:"chunking_strategy"`
}

type FeaturesConfig struct {
	EnableExperimentalFeatures bool `yaml:"enable_experimental_features"`
	EnableBetaUpdates          bool `yaml:"enable_beta_updates"`
	EnableTelemetry            bool `yaml:"enable_telemetry"`
}

type UpdatesConfig struct {
	AutoCheckUpdates    bool          `yaml:"auto_check_updates"`
	CheckInterval       time.Duration `yaml:"check_interval"`
	IncludePrereleases  bool          `yaml:"include_prereleases"`
	AutoDownloadUpdates bool          `yaml:"auto_download_updates"`
	NotifyOnUpdate      bool          `yaml:"notify_on_update"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("version", "1.0")
	v.SetDefault("git.branch", "main")
	v.SetDefault("git.auto_push", false)
	v.SetDefault("git.auto_pull", true)
	v.SetDefault("git.user_name", "GitCells")
	v.SetDefault("git.user_email", "gitcells@localhost")
	v.SetDefault("git.commit_template", "GitCells: {action} {filename} at {timestamp}")
	v.SetDefault("watcher.debounce_delay", "1s")
	v.SetDefault("watcher.file_extensions", []string{".xlsx", ".xls", ".xlsm"})
	v.SetDefault("watcher.ignore_patterns", []string{"~$*", "*.tmp"})
	v.SetDefault("converter.preserve_formulas", true)
	v.SetDefault("converter.preserve_styles", true)
	v.SetDefault("converter.preserve_comments", true)
	v.SetDefault("converter.compact_json", false)
	v.SetDefault("converter.ignore_empty_cells", true)
	v.SetDefault("converter.max_cells_per_sheet", DefaultMaxCellsPerSheet)
	v.SetDefault("converter.chunking_strategy", "sheet-based")
	v.SetDefault("features.enable_experimental_features", false)
	v.SetDefault("features.enable_beta_updates", false)
	v.SetDefault("features.enable_telemetry", true)
	v.SetDefault("updates.auto_check_updates", true)
	v.SetDefault("updates.check_interval", "24h")
	v.SetDefault("updates.include_prereleases", false)
	v.SetDefault("updates.auto_download_updates", false)
	v.SetDefault("updates.notify_on_update", true)

	// Load config file
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	} else {
		v.SetConfigName(".gitcells")
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
			ChunkingStrategy: v.GetString("converter.chunking_strategy"),
		},
		Features: FeaturesConfig{
			EnableExperimentalFeatures: v.GetBool("features.enable_experimental_features"),
			EnableBetaUpdates:          v.GetBool("features.enable_beta_updates"),
			EnableTelemetry:            v.GetBool("features.enable_telemetry"),
		},
		Updates: UpdatesConfig{
			AutoCheckUpdates:    v.GetBool("updates.auto_check_updates"),
			CheckInterval:       v.GetDuration("updates.check_interval"),
			IncludePrereleases:  v.GetBool("updates.include_prereleases"),
			AutoDownloadUpdates: v.GetBool("updates.auto_download_updates"),
			NotifyOnUpdate:      v.GetBool("updates.notify_on_update"),
		},
	}

	return cfg, nil
}

// Save saves the configuration to a file using YAML marshaling.
// This approach is simpler and more maintainable than manually setting
// each field in viper, as new fields are automatically handled.
func (c *Config) Save(configPath string) error {
	if configPath == "" {
		configPath = ".gitcells.yaml"
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// ConvertOptions defines options for conversion (will be moved to converter package)
type ConvertOptions struct {
	PreserveFormulas bool
	PreserveStyles   bool
	PreserveComments bool
	CompactJSON      bool
	IgnoreEmptyCells bool
	MaxCellsPerSheet int
	ChunkingStrategy string
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
		ChunkingStrategy: c.ChunkingStrategy,
	}
}
