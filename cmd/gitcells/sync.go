package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/Classic-Homes/gitcells/internal/converter"
	"github.com/Classic-Homes/gitcells/internal/git"
	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newSyncCommand(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync Excel files with their JSON representations",
		Long:  "Synchronize Excel files with their JSON representations",
		RunE: func(cmd *cobra.Command, args []string) error {
			commit, _ := cmd.Flags().GetBool("commit")
			includePatterns, _ := cmd.Flags().GetStringSlice("include")
			excludePatterns, _ := cmd.Flags().GetStringSlice("exclude")

			// Get current directory
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			// Load configuration
			cfg, err := config.Load(dir)
			if err != nil {
				logger.Warnf("Failed to load config, using defaults: %v", err)
				cfg = config.GetDefault()
			}

			// Find all Excel files
			excelFiles, err := findExcelFiles(dir, includePatterns)
			if err != nil {
				return utils.WrapError(err, utils.ErrorTypeFileSystem, "findExcelFiles", "failed to scan for Excel files")
			}

			// Filter out excluded patterns
			excelFiles = filterExcludedFiles(excelFiles, excludePatterns)

			if len(excelFiles) == 0 {
				fmt.Println("No Excel files found to sync")
				return nil
			}

			// Check status for each file
			var filesToSync []FileStatus
			for _, excelPath := range excelFiles {
				status, err := getFileStatus(excelPath, logger)
				if err != nil {
					logger.Warnf("Failed to get status for %s: %v", excelPath, err)
					continue
				}

				// Only sync modified or new files
				if status.Status == "modified" || status.Status == "new" {
					filesToSync = append(filesToSync, status)
				}
			}

			if len(filesToSync) == 0 {
				fmt.Println("✅ All files are already synced")
				return nil
			}

			fmt.Printf("\n🔄 Syncing %d files...\n", len(filesToSync))

			// Create converter
			conv := converter.NewConverter(logger)

			// Convert each file
			var convertedFiles []string
			successCount := 0
			for i, fileStatus := range filesToSync {
				fmt.Printf("[%d/%d] Converting %s... ", i+1, len(filesToSync), fileStatus.ExcelPath)

				// Ensure JSON directory exists
				jsonDir := filepath.Dir(fileStatus.JSONPath)
				if err := os.MkdirAll(jsonDir, constants.DirPermissions); err != nil {
					fmt.Printf("❌ Failed to create directory: %v\n", err)
					continue
				}

				// Convert Excel to JSON
				options := converter.ConvertOptions{
					PreserveFormulas: cfg.Converter.PreserveFormulas,
					PreserveStyles:   cfg.Converter.PreserveStyles,
					PreserveComments: cfg.Converter.PreserveComments,
					CompactJSON:      cfg.Converter.CompactJSON,
					IgnoreEmptyCells: cfg.Converter.IgnoreEmptyCells,
					MaxCellsPerSheet: cfg.Converter.MaxCellsPerSheet,
					ChunkingStrategy: cfg.Converter.ChunkingStrategy,
				}

				err := conv.ExcelToJSONFile(fileStatus.ExcelPath, fileStatus.JSONPath, options)
				if err != nil {
					fmt.Printf("❌ Error: %v\n", err)
					logger.Errorf("Failed to convert %s: %v", fileStatus.ExcelPath, err)
					continue
				}

				fmt.Println("✅")
				successCount++

				// Track converted chunk files for git commit via the
				// converter's own GetChunkPaths. The previous Walk-based
				// approach picked up everything ending in `.json` in the
				// chunk dir, which would include orphan staging leftovers
				// (`*.tmp.*/...`) that SweepOrphans is best-effort about
				// cleaning. GetChunkPaths returns only the files the
				// metadata claims, plus the metadata file itself.
				chunkPaths, err := conv.GetChunkPaths(fileStatus.ExcelPath)
				if err != nil {
					logger.Warnf("Failed to enumerate chunk paths for %s: %v", fileStatus.ExcelPath, err)
				} else {
					convertedFiles = append(convertedFiles, chunkPaths...)
				}
			}

			fmt.Printf("\n✅ Successfully synchronized %d/%d files\n", successCount, len(filesToSync))

			// Optionally commit to git
			if commit && len(convertedFiles) > 0 {
				gitRoot, err := git.FindRepositoryRoot(dir)
				if err != nil {
					logger.Debug("Not in a git repository, skipping commit")
					return nil
				}

				gitCfg := &git.Config{
					UserName:  cfg.Git.UserName,
					UserEmail: cfg.Git.UserEmail,
				}

				gitClient, err := git.NewClient(gitRoot, gitCfg, logger)
				if err != nil {
					return utils.WrapError(err, utils.ErrorTypeGit, "createGitClient", "failed to create git client")
				}

				if gitClient != nil {
					message := fmt.Sprintf("GitCells: synchronized %d Excel file(s)", len(convertedFiles))
					fmt.Printf("\nCommitting changes to git...\n")
					if err := gitClient.AutoCommit(convertedFiles, message); err != nil {
						return utils.WrapError(err, utils.ErrorTypeGit, "autoCommit", "failed to commit changes")
					}
					fmt.Println("Changes committed successfully")
				}
			}

			return nil
		},
	}

	cmd.Flags().Bool("commit", false, "commit JSON changes to git (if repository exists)")
	includePatterns := make([]string, len(constants.ExcelExtensions))
	for i, ext := range constants.ExcelExtensions {
		includePatterns[i] = "*" + ext
	}
	cmd.Flags().StringSlice("include", includePatterns, "file patterns to include")
	cmd.Flags().StringSlice("exclude", []string{constants.ExcelTempPrefix + "*", constants.TempFilePattern}, "file patterns to exclude")

	return cmd
}

func filterExcludedFiles(files []string, excludePatterns []string) []string {
	var filtered []string
	for _, file := range files {
		excluded := false
		fileName := filepath.Base(file)

		for _, pattern := range excludePatterns {
			matched, err := filepath.Match(pattern, fileName)
			if err != nil {
				continue
			}
			if matched {
				excluded = true
				break
			}
		}

		if !excluded {
			filtered = append(filtered, file)
		}
	}
	return filtered
}
