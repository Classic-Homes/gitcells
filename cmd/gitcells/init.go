package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Classic-Homes/gitcells/internal/tui"
	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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

// initTimeout is the maximum time to wait for initialization operations
const initTimeout = 10 * time.Second

// checkDirectoryAccess verifies if we can access and write to a directory
func checkDirectoryAccess(dir string) error {
	// First check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create parent to test write access
			parent := filepath.Dir(dir)
			if parent != dir { // Avoid infinite recursion
				if err := checkDirectoryAccess(parent); err != nil {
					return fmt.Errorf("cannot access parent directory %s: %w", parent, err)
				}
			}
			return nil // Directory doesn't exist but parent is writable
		}
		return fmt.Errorf("cannot access directory %s: %w", dir, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	// Test write access by creating a temporary file
	testFile := filepath.Join(dir, ".gitcells-write-test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		// Check for common permission-related error messages
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "permission denied") ||
			strings.Contains(errStr, "access denied") ||
			strings.Contains(errStr, "operation not permitted") {
			return utils.WrapFileError(err, utils.ErrorTypePermission, "init", dir,
				"insufficient permissions to write to directory")
		}
		return utils.WrapFileError(err, utils.ErrorTypeFileSystem, "init", dir,
			"cannot write to directory")
	}

	// Clean up test file
	os.Remove(testFile)
	return nil
}

// timeoutOperation runs an operation with a timeout and provides user feedback
func timeoutOperation(ctx context.Context, logger *logrus.Logger, operation string, fn func() error) error {
	done := make(chan error, 1)

	go func() {
		done <- fn()
	}()

	// Provide user feedback after 2 seconds
	feedbackTimer := time.NewTimer(2 * time.Second)
	defer feedbackTimer.Stop()

	select {
	case err := <-done:
		return err
	case <-feedbackTimer.C:
		logger.Infof("Still working on %s (this may take a moment if directories require elevated privileges)...", operation)
		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return fmt.Errorf("%s operation timed out after %v - this may be due to insufficient permissions",
				operation, initTimeout)
		}
	case <-ctx.Done():
		return fmt.Errorf("%s operation timed out after %v - this may be due to insufficient permissions",
			operation, initTimeout)
	}
}

// suggestSolution provides helpful suggestions based on the error and platform
func suggestSolution(err error, dir string, logger *logrus.Logger) {
	if err == nil {
		return
	}

	errStr := strings.ToLower(err.Error())
	isPermissionError := strings.Contains(errStr, "permission") ||
		strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "operation not permitted")

	if !isPermissionError {
		return
	}

	logger.Error("Initialization failed due to insufficient permissions.")
	logger.Infof("Directory: %s", dir)

	switch runtime.GOOS {
	case "windows":
		logger.Info("Try one of the following solutions:")
		logger.Info("  1. Run the command prompt as Administrator")
		logger.Info("  2. Choose a different directory where you have write permissions")
		logger.Info("  3. Use a directory in your user profile (e.g., $env:USERPROFILE\\gitcells)")
	case "darwin", "linux":
		logger.Info("Try one of the following solutions:")
		logger.Info("  1. Use sudo: sudo gitcells init")
		logger.Info("  2. Change directory permissions: chmod 755 <directory>")
		logger.Info("  3. Choose a different directory where you have write permissions")
		logger.Info("  4. Use a directory in your home folder: ~/gitcells")
	default:
		logger.Info("Try choosing a different directory where you have write permissions")
	}
}

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

			// Convert to absolute path for better error messages
			absDir, err := filepath.Abs(dir)
			if err != nil {
				return utils.WrapFileError(err, utils.ErrorTypeFileSystem, "init", dir, "failed to resolve directory path")
			}

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), initTimeout)
			defer cancel()

			// Check directory access first
			logger.Info("Checking directory permissions...")
			if err := checkDirectoryAccess(filepath.Dir(absDir)); err != nil {
				suggestSolution(err, absDir, logger)
				return err
			}

			// Ensure directory exists with timeout
			err = timeoutOperation(ctx, logger, "directory creation", func() error {
				return os.MkdirAll(absDir, dirPermissions)
			})
			if err != nil {
				suggestSolution(err, absDir, logger)
				return utils.WrapFileError(err, utils.ErrorTypeFileSystem, "init", absDir, "failed to create directory")
			}

			// Create config file
			configPath := filepath.Join(absDir, ".gitcells.yaml")
			if _, err := os.Stat(configPath); err == nil {
				// Config already exists
				overwrite, _ := cmd.Flags().GetBool("force")
				if !overwrite {
					return utils.NewError(utils.ErrorTypeConfig, "init", "config file already exists. Use --force to overwrite")
				}
			}

			err = timeoutOperation(ctx, logger, "config file creation", func() error {
				return os.WriteFile(configPath, []byte(defaultConfig), filePermissions)
			})
			if err != nil {
				suggestSolution(err, absDir, logger)
				return utils.WrapFileError(err, utils.ErrorTypeFileSystem, "init", configPath, "failed to write config file")
			}

			logger.Infof("Created GitCells configuration at %s", configPath)

			// Create .gitignore
			gitignorePath := filepath.Join(absDir, ".gitignore")
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
			err = timeoutOperation(ctx, logger, ".gitignore creation", func() error {
				return os.WriteFile(gitignorePath, []byte(gitignoreContent), filePermissions)
			})
			if err != nil {
				logger.Warnf("Failed to create .gitignore: %v", err)
				suggestSolution(err, absDir, logger)
			} else {
				logger.Info("Created .gitignore file")
			}

			// Initialize git repo if requested
			initGit, _ := cmd.Flags().GetBool("git")
			if initGit {
				var repo *git.Repository
				var worktree *git.Worktree

				// Check if directory is already a git repository with timeout
				err = timeoutOperation(ctx, logger, "git repository check", func() error {
					var err error
					repo, err = git.PlainOpen(absDir)
					return err
				})

				switch err {
				case nil:
					logger.Info("Directory is already a git repository")
				case git.ErrRepositoryNotExists:
					// Initialize new git repository with timeout
					err = timeoutOperation(ctx, logger, "git repository initialization", func() error {
						var err error
						repo, err = git.PlainInit(absDir, false)
						return err
					})
					if err != nil {
						suggestSolution(err, absDir, logger)
						return utils.WrapError(err, utils.ErrorTypeGit, "init", "failed to initialize git repository")
					}

					// Get worktree with timeout
					err = timeoutOperation(ctx, logger, "git worktree access", func() error {
						var err error
						worktree, err = repo.Worktree()
						return err
					})
					if err != nil {
						suggestSolution(err, absDir, logger)
						return utils.WrapError(err, utils.ErrorTypeGit, "init", "failed to get worktree")
					}

					// Add files and create initial commit with timeout
					err = timeoutOperation(ctx, logger, "git initial commit", func() error {
						// Add .gitcells.yaml to git
						if _, err := worktree.Add(".gitcells.yaml"); err != nil {
							logger.Warnf("Failed to add .gitcells.yaml to git: %v", err)
						}

						// Add .gitignore to git
						if _, err := worktree.Add(".gitignore"); err != nil {
							logger.Warnf("Failed to add .gitignore to git: %v", err)
						}

						// Check if there are changes to commit
						status, err := worktree.Status()
						if err != nil {
							return fmt.Errorf("failed to get git status: %w", err)
						}

						if !status.IsClean() {
							// Create initial commit
							_, err := worktree.Commit("Initial GitCells setup", &git.CommitOptions{
								Author: &object.Signature{
									Name:  "GitCells",
									Email: "gitcells@localhost",
									When:  time.Now(),
								},
							})
							if err != nil {
								return fmt.Errorf("failed to create initial commit: %w", err)
							}
						}
						return nil
					})

					if err != nil {
						logger.Warnf("Failed to create initial commit: %v", err)
						suggestSolution(err, absDir, logger)
					} else {
						logger.Info("Git repository initialized successfully")
					}
				default:
					suggestSolution(err, absDir, logger)
					return utils.WrapError(err, utils.ErrorTypeGit, "init", "failed to check git repository")
				}
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
