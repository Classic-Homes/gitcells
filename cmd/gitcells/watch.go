package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/Classic-Homes/gitcells/internal/converter"
	"github.com/Classic-Homes/gitcells/internal/git"
	"github.com/Classic-Homes/gitcells/internal/utils"
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
				return utils.WrapFileError(err, utils.ErrorTypeConfig, "watch", configPath, "failed to load config")
			}

			// Override config with command flags
			autoCommit, _ := cmd.Flags().GetBool("auto-commit")

			// Initialize components
			conv := converter.NewConverter(logger)

			// Sweep stale staging/backup chunk dirs from a previous crash.
			// Best-effort; safe to run before any writes start. Resolve
			// relative to cwd so the sweep targets the repo's data dir,
			// not whatever `.gitcells/data` happens to mean from the
			// process's working directory.
			converter.SweepOrphans(resolveDataDir("."), converter.DefaultOrphanMaxAge, logger)

			gitConfig := &git.Config{
				UserName:  cfg.Git.UserName,
				UserEmail: cfg.Git.UserEmail,
			}

			gitClient, err := git.NewClient(".", gitConfig, logger)
			if err != nil {
				return utils.WrapError(err, utils.ErrorTypeGit, "watch", "failed to initialize git client")
			}

			// Create event handler
			handler := func(event watcher.FileEvent) error {
				logger.Infof("Processing %s: %s", event.Type, event.Path)

				if !autoCommit {
					logger.Infof("Auto-commit disabled, skipping git operations")
					return nil
				}

				// Convert Excel to JSON using chunking
				convertOptions := converter.ConvertOptions{
					PreserveFormulas: cfg.Converter.PreserveFormulas,
					PreserveStyles:   cfg.Converter.PreserveStyles,
					PreserveComments: cfg.Converter.PreserveComments,
					CompactJSON:      cfg.Converter.CompactJSON,
					IgnoreEmptyCells: cfg.Converter.IgnoreEmptyCells,
					MaxCellsPerSheet: cfg.Converter.MaxCellsPerSheet,
					ChunkingStrategy: "sheet-based",
				}

				// The converter will automatically save to .gitcells/data directory
				if convertErr := conv.ExcelToJSONFile(event.Path, event.Path, convertOptions); convertErr != nil {
					return utils.WrapFileError(convertErr, utils.ErrorTypeConverter, "watch", event.Path, "failed to convert Excel to JSON")
				}

				// Commit changes if git repository exists
				if gitClient != nil {
					// Get the chunk paths that were created
					chunkPaths, err := conv.GetChunkPaths(event.Path)
					if err != nil {
						logger.Warnf("Failed to get chunk paths for git commit: %v", err)
						// Fall back to committing the entire .gitcells/data directory
						chunkPaths = []string{filepath.Join(".gitcells", "data")}
					}

					message := fmt.Sprintf("GitCells: %s %s", event.Type.String(), filepath.Base(event.Path))
					return gitClient.AutoCommit(chunkPaths, message)
				}
				return nil
			}

			// Setup watcher
			watcherConfig := &watcher.Config{
				IgnorePatterns: cfg.Watcher.IgnorePatterns,
				DebounceDelay:  cfg.Watcher.DebounceDelay,
				FileExtensions: cfg.Watcher.FileExtensions,
			}

			fw, err := watcher.NewFileWatcher(watcherConfig, handler, logger)
			if err != nil {
				return utils.WrapError(err, utils.ErrorTypeWatcher, "watch", "failed to create file watcher")
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
				return utils.WrapError(err, utils.ErrorTypeWatcher, "watch", "failed to start file watcher")
			}

			// Drain failures from the watcher onto a durable log file. Without
			// this, a git commit failure or a corrupted source file goes silent
			// and the user thinks everything is fine.
			failureWg := &sync.WaitGroup{}
			failureWg.Add(1)
			go func() {
				defer failureWg.Done()
				logPath := filepath.Join(constants.GitCellsDir, "failures.log")
				for we := range fw.Errors() {
					line := fmt.Sprintf("%s\t%s\t%s\t%v\n",
						we.Timestamp.UTC().Format(time.RFC3339),
						we.EventType.String(),
						we.Path,
						we.Err,
					)
					if appendErr := appendFailureLog(logPath, line); appendErr != nil {
						logger.Warnf("Failed to write failures.log: %v", appendErr)
					}
				}
			}()

			logger.Info("Watching for changes... Press Ctrl+C to stop")

			// Wait for interrupt signal
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			defer signal.Stop(sigChan)
			<-sigChan

			logger.Info("Shutting down (draining in-flight conversions)...")
			stopErr := fw.Stop(watcher.DefaultShutdownGrace)
			failureWg.Wait()
			return stopErr
		},
	}

	cmd.Flags().Bool("auto-commit", true, "automatically commit changes to git")

	return cmd
}

// appendFailureLog appends a single line to the failures log, creating
// the parent directory if needed. Each line is timestamp \t event \t path \t error.
func appendFailureLog(path, line string) error {
	if err := os.MkdirAll(filepath.Dir(path), constants.DirPermissions); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, constants.FilePermissions) // #nosec G304 - fixed path under .gitcells/
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line)
	return err
}
