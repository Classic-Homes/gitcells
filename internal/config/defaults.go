package config

import "time"

// DefaultConfigYAML provides the default configuration template
const DefaultConfigYAML = `version: 1.0

git:
  branch: main
  auto_push: false
  auto_pull: true
  user_name: "SheetSync"
  user_email: "sheetsync@localhost"
  commit_template: "SheetSync: {action} {filename} at {timestamp}"

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
`

// GetDefault returns a Config struct with default values
func GetDefault() *Config {
	return &Config{
		Version: "1.0",
		Git: GitConfig{
			Branch:         "main",
			AutoPush:       false,
			AutoPull:       true,
			UserName:       "SheetSync",
			UserEmail:      "sheetsync@localhost",
			CommitTemplate: "SheetSync: {action} {filename} at {timestamp}",
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
	}
}
