// Package config provides configuration management for GitCells.
package config

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// ResolveGitIdentity returns the user's git user.name and user.email by
// shelling out to `git config`. Empty strings on failure — callers decide
// whether to fall back to the GitCells placeholder identity.
func ResolveGitIdentity() (name, email string) {
	read := func(key string) string {
		out, err := exec.Command("git", "config", "--get", key).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	return read("user.name"), read("user.email")
}

const (
	// DefaultMaxCellsPerSheet caps cell ingestion per sheet during conversion.
	DefaultMaxCellsPerSheet = 1000000
)

type Config struct {
	Version   string          `yaml:"version"`
	Git       GitConfig       `yaml:"git"`
	Watcher   WatcherConfig   `yaml:"watcher"`
	Converter ConverterConfig `yaml:"converter"`
}

type GitConfig struct {
	Branch    string `yaml:"branch"`
	UserName  string `yaml:"user_name"`
	UserEmail string `yaml:"user_email"`
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

func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetDefault("version", "1.0")
	v.SetDefault("git.branch", "main")
	v.SetDefault("git.user_name", "GitCells")
	v.SetDefault("git.user_email", "gitcells@localhost")
	v.SetDefault("watcher.debounce_delay", "2s")
	v.SetDefault("watcher.file_extensions", []string{".xlsx", ".xls", ".xlsm"})
	v.SetDefault("watcher.ignore_patterns", []string{"~$*", "*.tmp", ".~lock.*"})
	v.SetDefault("converter.preserve_formulas", true)
	v.SetDefault("converter.preserve_styles", true)
	v.SetDefault("converter.preserve_comments", true)
	v.SetDefault("converter.compact_json", false)
	v.SetDefault("converter.ignore_empty_cells", true)
	v.SetDefault("converter.max_cells_per_sheet", DefaultMaxCellsPerSheet)
	v.SetDefault("converter.chunking_strategy", "sheet-based")

	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	} else {
		v.SetConfigName(".gitcells")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		// Missing config file = use defaults. Read errors for other reasons
		// (malformed YAML) propagate via the explicit-path branch.
		_ = v.ReadInConfig()
	}

	cfg := &Config{
		Version: v.GetString("version"),
		Git: GitConfig{
			Branch:    v.GetString("git.branch"),
			UserName:  v.GetString("git.user_name"),
			UserEmail: v.GetString("git.user_email"),
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
	}

	return cfg, nil
}

// Save writes the configuration to a YAML file.
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
