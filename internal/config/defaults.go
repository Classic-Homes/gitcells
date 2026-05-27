package config

import (
	"time"

	"github.com/Classic-Homes/gitcells/internal/constants"
)

// GetDefault returns a Config struct with default values
func GetDefault() *Config {
	return &Config{
		Version: "1.0",
		Git: GitConfig{
			Branch:         "main",
			AutoPush:       false,
			AutoPull:       true,
			UserName:       constants.DefaultGitUserName,
			UserEmail:      constants.DefaultGitUserEmail,
			CommitTemplate: constants.DefaultCommitTemplate,
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
			MaxCellsPerSheet: 1000000,
		},
		Features: FeaturesConfig{
			EnableExperimentalFeatures: false,
			EnableBetaUpdates:          false,
			EnableTelemetry:            true,
		},
		Updates: UpdatesConfig{
			AutoCheckUpdates:    true,
			CheckInterval:       24 * time.Hour,
			IncludePrereleases:  false,
			AutoDownloadUpdates: false,
			NotifyOnUpdate:      true,
		},
	}
}
