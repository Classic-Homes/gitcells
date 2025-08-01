package config

import (
	"time"

	"github.com/Classic-Homes/gitcells/internal/constants"
)

// DefaultConfigYAML provides the default configuration template
const DefaultConfigYAML = `version: 1.0

git:
  branch: main
  auto_push: false
  auto_pull: true
  user_name: "" + constants.DefaultGitUserName + ""
  user_email: "" + constants.DefaultGitUserEmail + ""
  commit_template: "" + constants.DefaultCommitTemplate + ""

watcher:
  directories: []
  ignore_patterns:
    - "" + constants.ExcelTempPrefix + "*"
    - "" + constants.TempFilePattern + ""
    - "" + constants.LockFilePattern + ""
  debounce_delay: 2s
  file_extensions:
    - "" + constants.ExtXLSX + ""
    - "" + constants.ExtXLS + ""
    - "" + constants.ExtXLSM + ""

converter:
  preserve_formulas: true
  preserve_styles: true
  preserve_comments: true
  compact_json: false
  ignore_empty_cells: true
  max_cells_per_sheet: 1000000

features:
  enable_experimental_features: false
  enable_beta_updates: false
  enable_telemetry: true

updates:
  auto_check_updates: true
  check_interval: 24h
  include_prereleases: false
  auto_download_updates: false
  notify_on_update: true
`

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
