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
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
)

type Client struct {
	// mu serializes worktree operations. go-git's worktree is not
	// safe for concurrent use, and the watcher fires handlers in
	// parallel goroutines (one per debounced file key).
	mu       sync.Mutex
	repo     *git.Repository
	worktree *git.Worktree
	config   *Config
	logger   *logrus.Logger
}

type Config struct {
	UserName  string
	UserEmail string
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
// Safe for concurrent callers; serialized via c.mu. Transient errors
// (index lock, "resource busy") get one retry with a short backoff.
func (c *Client) AutoCommit(files []string, message string) error {
	if c == nil {
		// No git repository - skip commit
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Stage files. Track how many actually made it into the index so we
	// can distinguish "caller asked for nothing" (fine) from "caller
	// asked for several files and every one was outside the worktree"
	// (failure that previously presented as silent success).
	worktreeRoot := c.worktree.Filesystem.Root()
	stagedCount := 0
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
		if err := retryTransient(func() error {
			_, addErr := c.worktree.Add(relPath)
			return addErr
		}); err != nil {
			return utils.WrapFileError(err, utils.ErrorTypeGit, "stageFile", file, "failed to stage file")
		}
		stagedCount++
	}

	if len(files) > 0 && stagedCount == 0 {
		return utils.NewError(utils.ErrorTypeGit, "AutoCommit",
			"none of the provided files were under the git worktree; nothing was committed")
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

	if message == "" {
		message = fmt.Sprintf("GitCells: Update JSON representations (%d files)", len(files))
	}

	var commit plumbing.Hash
	commitErr := retryTransient(func() error {
		h, err := c.worktree.Commit(message, &git.CommitOptions{
			Author: &object.Signature{
				Name:  c.config.UserName,
				Email: c.config.UserEmail,
				When:  time.Now(),
			},
		})
		if err != nil {
			return err
		}
		commit = h
		return nil
	})
	if commitErr != nil {
		return utils.WrapError(commitErr, utils.ErrorTypeGit, "createCommit", "failed to create git commit")
	}

	c.logger.Infof("Committed JSON updates: %s", commit.String()[:8])
	return nil
}

// retryTransient runs op once, and on a transient-looking failure runs
// it a second time after a 200ms pause. "Transient" here means anything
// that suggests a temporary lock or fs contention — index.lock, EBUSY,
// etc. Non-transient errors return immediately.
func retryTransient(op func() error) error {
	err := op()
	if err == nil || !isTransientGitError(err) {
		return err
	}
	time.Sleep(200 * time.Millisecond)
	return op()
}

func isTransientGitError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "index.lock") ||
		strings.Contains(msg, "resource temporarily unavailable") ||
		strings.Contains(msg, "resource busy") ||
		strings.Contains(msg, "file is in use") ||
		strings.Contains(msg, "sharing violation")
}

// IsClean returns true if the working directory is clean
func (c *Client) IsClean() (bool, error) {
	if c == nil {
		return true, nil // No git repo = clean
	}
	c.mu.Lock()
	defer c.mu.Unlock()
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

// CurrentBranch returns the short HEAD branch name, or "" if HEAD is
// detached, the repo is uninitialized, or no repo is open.
func (c *Client) CurrentBranch() (string, error) {
	if c == nil || c.repo == nil {
		return "", nil
	}
	head, err := c.repo.Head()
	if err != nil {
		return "", nil
	}
	ref := head.Name()
	if !ref.IsBranch() {
		return "", nil
	}
	return ref.Short(), nil
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
