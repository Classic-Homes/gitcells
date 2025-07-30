package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newSyncCommand(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync Excel files with their JSON representations",
		Long:  "Synchronize Excel files with their JSON representations and commit changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Sync command will scan for changes and update files")

			// TODO: Implement sync functionality
			// This will:
			// 1. Scan for Excel files
			// 2. Compare with existing JSON files
			// 3. Convert changed files
			// 4. Optionally commit changes

			return nil
		},
	}

	cmd.Flags().Bool("commit", true, "commit changes to git")
	cmd.Flags().Bool("push", false, "push commits to remote")
	cmd.Flags().StringSlice("include", []string{"*.xlsx", "*.xls", "*.xlsm"}, "file patterns to include")
	cmd.Flags().StringSlice("exclude", []string{"~$*", "*.tmp"}, "file patterns to exclude")

	return cmd
}
