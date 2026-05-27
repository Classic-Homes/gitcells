package config

import (
	"time"

	"github.com/Classic-Homes/gitcells/internal/constants"
)

// GetDefault returns a Config populated with built-in defaults. Used when
// no config file exists or the loader errored.
func GetDefault() *Config {
	return &Config{
		Version: "1.0",
		Git: GitConfig{
			Branch:    constants.DefaultGitBranch,
			UserName:  constants.DefaultGitUserName,
			UserEmail: constants.DefaultGitUserEmail,
		},
		Watcher: WatcherConfig{
			Directories:    []string{},
			IgnorePatterns: constants.DefaultIgnorePatterns,
			DebounceDelay:  2 * time.Second,
			FileExtensions: constants.ExcelExtensions,
		},
		Converter: ConverterConfig{
			PreserveFormulas: true,
			PreserveStyles:   true,
			PreserveComments: true,
			CompactJSON:      false,
			IgnoreEmptyCells: true,
			MaxCellsPerSheet: DefaultMaxCellsPerSheet,
			ChunkingStrategy: "sheet-based",
		},
	}
}
