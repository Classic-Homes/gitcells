// Package git provides minimal Git operations for GitCells.
// For advanced git operations, users should use dedicated git tools.
package git

import (
	"fmt"
	"path/filepath"
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

// AutoCommit commits converted JSON files with a simple message
func (c *Client) AutoCommit(files []string, message string) error {
	if c == nil {
		// No git repository - skip commit
		return nil
	}

	// Stage files
	for _, file := range files {
		relPath, _ := filepath.Rel(c.worktree.Filesystem.Root(), file)
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

	// Use provided message or default
	if message == "" {
		message = fmt.Sprintf("GitCells: Update JSON representations (%d files)", len(files))
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
