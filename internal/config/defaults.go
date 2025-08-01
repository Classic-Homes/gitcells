package config

import "time"

// DefaultConfigYAML provides the default configuration template
const DefaultConfigYAML = `version: 1.0

git:
  branch: main
  auto_push: false
  auto_pull: true
  user_name: "GitCells"
  user_email: "gitcells@localhost"
  commit_template: "GitCells: {action} {filename} at {timestamp}"

watcher:
  directories: []
  ignore_patterns:
    - "~$*"
    - "*.tmp"
    - ".~lock.*"
  debounce_delay: 2s
  file_extensions:
    - ".xlsx"
    - ".xls"
    - ".xlsm"

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
			UserName:       "GitCells",
			UserEmail:      "gitcells@localhost",
			CommitTemplate: "GitCells: {action} {filename} at {timestamp}",
		},
		Watcher: WatcherConfig{
			Directories:    []string{},
			IgnorePatterns: []string{"~$*", "*.tmp", ".~lock.*"},
			DebounceDelay:  2 * time.Second,
			FileExtensions: []string{".xlsx", ".xls", ".xlsm"},
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
