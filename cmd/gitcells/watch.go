package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/Classic-Homes/gitcells/internal/converter"
	"github.com/Classic-Homes/gitcells/internal/git"
	"github.com/Classic-Homes/gitcells/internal/watcher"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newWatchCommand(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch [directories...]",
		Short: "Watch directories for Excel file changes",
		Long:  "Start watching specified directories for Excel file changes and auto-commit to git",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			configPath, _ := cmd.Flags().GetString("config")
			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Override config with command flags
			autoCommit, _ := cmd.Flags().GetBool("auto-commit")
			autoPush, _ := cmd.Flags().GetBool("auto-push")

			if cmd.Flags().Changed("auto-push") {
				cfg.Git.AutoPush = autoPush
			}

			// Initialize components
			conv := converter.NewConverter(logger)

			gitConfig := &git.Config{
				UserName:       cfg.Git.UserName,
				UserEmail:      cfg.Git.UserEmail,
				CommitTemplate: cfg.Git.CommitTemplate,
				AutoPush:       cfg.Git.AutoPush,
				AutoPull:       cfg.Git.AutoPull,
				Branch:         cfg.Git.Branch,
			}

			gitClient, err := git.NewClient(".", gitConfig, logger)
			if err != nil {
				return fmt.Errorf("failed to initialize git client: %w", err)
			}

			// Create event handler
			handler := func(event watcher.FileEvent) error {
				logger.Infof("Processing %s: %s", event.Type, event.Path)

				if !autoCommit {
					logger.Infof("Auto-commit disabled, skipping git operations")
					return nil
				}

				// Convert Excel to JSON
				convertOptions := converter.ConvertOptions{
					PreserveFormulas: cfg.Converter.PreserveFormulas,
					PreserveStyles:   cfg.Converter.PreserveStyles,
					PreserveComments: cfg.Converter.PreserveComments,
					CompactJSON:      cfg.Converter.CompactJSON,
					IgnoreEmptyCells: cfg.Converter.IgnoreEmptyCells,
					MaxCellsPerSheet: cfg.Converter.MaxCellsPerSheet,
				}
				doc, convertErr := conv.ExcelToJSON(event.Path, convertOptions)
				if convertErr != nil {
					return fmt.Errorf("failed to convert Excel to JSON: %w", convertErr)
				}

				// Save JSON file
				jsonPath := event.Path + ".json"
				jsonData, err := json.MarshalIndent(doc, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}

				if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
					return fmt.Errorf("failed to write JSON file: %w", err)
				}

				// Commit changes if auto-commit is enabled
				metadata := map[string]string{
					"filename": filepath.Base(event.Path),
					"action":   event.Type.String(),
				}

				return gitClient.AutoCommit([]string{jsonPath}, metadata)
			}

			// Setup watcher
			watcherConfig := &watcher.Config{
				IgnorePatterns: cfg.Watcher.IgnorePatterns,
				DebounceDelay:  cfg.Watcher.DebounceDelay,
				FileExtensions: cfg.Watcher.FileExtensions,
			}

			fw, err := watcher.NewFileWatcher(watcherConfig, handler, logger)
			if err != nil {
				return fmt.Errorf("failed to create file watcher: %w", err)
			}

			// Add directories to watch
			for _, dir := range args {
				if err := fw.AddDirectory(dir); err != nil {
					logger.Warnf("Failed to add directory %s: %v", dir, err)
				} else {
					logger.Infof("Watching directory: %s", dir)
				}
			}

			// Start watching
			if err := fw.Start(); err != nil {
				return fmt.Errorf("failed to start file watcher: %w", err)
			}

			logger.Info("Watching for changes... Press Ctrl+C to stop")

			// Wait for interrupt signal
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			<-sigChan

			logger.Info("Shutting down...")
			return fw.Stop()
		},
	}

	cmd.Flags().Bool("auto-commit", true, "automatically commit changes to git")
	cmd.Flags().Bool("auto-push", false, "automatically push commits to remote")

	return cmd
}
