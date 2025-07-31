package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Classic-Homes/gitcells/internal/tui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const defaultConfig = `version: 1.0

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
`

func newInitCommand(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "Initialize GitCells in a directory",
		Long:  "Initialize GitCells configuration and Git repository in the specified directory",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if TUI mode is requested
			useTUI, _ := cmd.Flags().GetBool("tui")
			if useTUI {
				logger.Info("Launching setup wizard in TUI mode...")
				return tui.Run()
			}

			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			// Ensure directory exists
			if err := os.MkdirAll(dir, dirPermissions); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			// Create config file
			configPath := filepath.Join(dir, ".gitcells.yaml")
			if _, err := os.Stat(configPath); err == nil {
				// Config already exists
				overwrite, _ := cmd.Flags().GetBool("force")
				if !overwrite {
					return fmt.Errorf("config file already exists. Use --force to overwrite")
				}
			}

			if err := os.WriteFile(configPath, []byte(defaultConfig), filePermissions); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			logger.Infof("Created GitCells configuration at %s", configPath)

			// Initialize git repo if requested
			initGit, _ := cmd.Flags().GetBool("git")
			if initGit {
				// TODO: Initialize git repository
				logger.Info("Git initialization will be implemented with git client")
			}

			// Create .gitignore
			gitignorePath := filepath.Join(dir, ".gitignore")
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
			if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), filePermissions); err != nil {
				logger.Warnf("Failed to create .gitignore: %v", err)
			}

			logger.Info("GitCells initialized successfully!")
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "overwrite existing configuration")
	cmd.Flags().Bool("git", true, "initialize git repository")
	cmd.Flags().Bool("tui", false, "use TUI setup wizard")

	return cmd
}
