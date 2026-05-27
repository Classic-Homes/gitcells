// Package git provides minimal Git operations for GitCells.
// For advanced git operations, users should use dedicated git tools.
package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
)

type Client struct {
	repo     *git.Repository
	worktree *git.Worktree
	config   *Config
	logger   *logrus.Logger
}

type Config struct {
	UserName       string
	UserEmail      string
	CommitTemplate string
}

func NewClient(repoPath string, config *Config, logger *logrus.Logger) (*Client, error) {
	repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})

	if err == git.ErrRepositoryNotExists {
		// Git repository not found - this is fine, we'll just skip git operations
		logger.Debug("No git repository found - git operations will be skipped")
		return nil, nil
	} else if err != nil {
		return nil, utils.WrapError(err, utils.ErrorTypeGit, "openRepository", "failed to open git repository")
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorTypeGit, "getWorktree", "failed to get git worktree")
	}

	return &Client{
		repo:     repo,
		worktree: worktree,
		config:   config,
		logger:   logger,
	}, nil
}

// AutoCommit commits converted JSON files with a simple message.
// Not safe for concurrent calls from multiple goroutines on the same Client.
func (c *Client) AutoCommit(files []string, message string) error {
	if c == nil {
		// No git repository - skip commit
		return nil
	}

	// Stage files
	worktreeRoot := c.worktree.Filesystem.Root()
	for _, file := range files {
		relPath, err := filepath.Rel(worktreeRoot, file)
		if err != nil {
			return utils.WrapFileError(err, utils.ErrorTypeGit, "stageFile", file,
				"file is not under git worktree root")
		}
		if strings.HasPrefix(relPath, "..") {
			c.logger.Warnf("Skipping file outside worktree: %s", file)
			continue
		}
		if _, err := c.worktree.Add(relPath); err != nil {
			return utils.WrapFileError(err, utils.ErrorTypeGit, "stageFile", file, "failed to stage file")
		}
	}

	// Check if there are changes to commit
	status, err := c.worktree.Status()
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeGit, "getStatus", "failed to get git status")
	}

	if status.IsClean() {
		c.logger.Debug("No changes to commit")
		return nil
	}

	// Use provided message, fall back to template, then to default.
	if message == "" {
		message = c.renderCommitTemplate(len(files))
	}

	// Create commit
	commit, err := c.worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  c.config.UserName,
			Email: c.config.UserEmail,
			When:  time.Now(),
		},
	})

	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeGit, "createCommit", "failed to create git commit")
	}

	c.logger.Infof("Committed JSON updates: %s", commit.String()[:8])
	return nil
}

// renderCommitTemplate substitutes {action}, {filename}, {timestamp} in the
// configured CommitTemplate. Returns a sensible default if the template is
// empty or substitution produced no recognised tokens.
func (c *Client) renderCommitTemplate(fileCount int) string {
	defaultMsg := fmt.Sprintf("GitCells: Update JSON representations (%d files)", fileCount)
	if c.config == nil || c.config.CommitTemplate == "" {
		return defaultMsg
	}
	msg := c.config.CommitTemplate
	msg = strings.ReplaceAll(msg, "{action}", "sync")
	msg = strings.ReplaceAll(msg, "{filename}", fmt.Sprintf("%d files", fileCount))
	msg = strings.ReplaceAll(msg, "{timestamp}", time.Now().Format("2006-01-02 15:04:05"))
	if msg == c.config.CommitTemplate {
		// Template had no recognised tokens — fall back rather than commit
		// an opaque static string.
		return defaultMsg
	}
	return msg
}

// IsClean returns true if the working directory is clean
func (c *Client) IsClean() (bool, error) {
	if c == nil {
		return true, nil // No git repo = clean
	}
	status, err := c.worktree.Status()
	if err != nil {
		return false, err
	}
	return status.IsClean(), nil
}

// InGitRepository returns true if the current directory is in a git repository
func (c *Client) InGitRepository() bool {
	return c != nil && c.repo != nil
}

// gitRootCache caches git root lookups for performance
var gitRootCache sync.Map

// FindRepositoryRoot finds the root directory of the Git repository
// containing the given path. If no Git repository is found, returns
// an error.
func FindRepositoryRoot(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	dir := absPath
	for {
		gitPath := filepath.Join(dir, ".git")
		info, err := os.Stat(gitPath)
		if err == nil {
			// Check if it's a directory (not a .git file from submodules)
			if info.IsDir() {
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", utils.NewError(utils.ErrorTypeGit, "findRepositoryRoot", "not a git repository (or any parent)")
		}
		dir = parent
	}
}

// FindRepositoryRootCached finds the git repository root with caching for performance.
// This is useful when repeatedly checking the same paths.
func FindRepositoryRootCached(startPath string) (string, error) {
	if cached, ok := gitRootCache.Load(startPath); ok {
		return cached.(string), nil
	}

	root, err := FindRepositoryRoot(startPath)
	if err == nil {
		gitRootCache.Store(startPath, root)
	}
	return root, err
}
