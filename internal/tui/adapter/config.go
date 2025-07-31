package adapter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/Classic-Homes/gitcells/internal/tui/types"
	"github.com/spf13/viper"
)

// ConfigAdapter bridges the TUI setup wizard with the config package
type ConfigAdapter struct {
	configPath string
}

func NewConfigAdapter(directory string) *ConfigAdapter {
	return &ConfigAdapter{
		configPath: filepath.Join(directory, ".gitcells.yaml"),
	}
}

// SaveSetupConfig converts TUI setup config to internal config format and saves it
func (ca *ConfigAdapter) SaveSetupConfig(setup types.SetupConfig) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(setup.Directory, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a new viper instance for the config
	v := viper.New()
	v.SetConfigFile(ca.configPath)

	// Set configuration values
	v.Set("version", "1.0")

	// Git settings
	v.Set("git.branch", "main")
	v.Set("git.auto_push", setup.AutoPush)
	v.Set("git.auto_pull", true)
	v.Set("git.user_name", "GitCells")
	v.Set("git.user_email", "gitcells@localhost")
	v.Set("git.commit_template", setup.CommitTemplate)

	// Watcher settings
	v.Set("watcher.directories", []string{setup.Directory})
	v.Set("watcher.ignore_patterns", []string{"~$*", "*.tmp", ".~lock.*"})
	v.Set("watcher.debounce_delay", "2s")
	v.Set("watcher.file_extensions", []string{".xlsx", ".xls", ".xlsm"})

	// Converter settings
	v.Set("converter.preserve_formulas", true)
	v.Set("converter.preserve_styles", true)
	v.Set("converter.preserve_comments", true)
	v.Set("converter.compact_json", false)
	v.Set("converter.ignore_empty_cells", true)
	v.Set("converter.max_cells_per_sheet", 1000000)

	// Write the configuration file
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// LoadConfig loads existing configuration
func (ca *ConfigAdapter) LoadConfig() (*config.Config, error) {
	cfg, err := config.Load(filepath.Dir(ca.configPath))
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// CreateGitIgnore creates a .gitignore file with Excel patterns
func (ca *ConfigAdapter) CreateGitIgnore(directory string) error {
	gitignorePath := filepath.Join(directory, ".gitignore")
	gitignoreContent := `# Excel temporary files
~$*.xlsx
~$*.xls
~$*.xlsm
*.tmp

# OS files
.DS_Store
Thumbs.db

# GitCells
.gitcells.cache/
`
	return os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
}
