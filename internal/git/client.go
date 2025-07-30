// Package git provides Git operations and conflict resolution functionality.
package git

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Classic-Homes/sheetsync/internal/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
	AutoPush       bool
	AutoPull       bool
	Branch         string
}

func NewClient(repoPath string, config *Config, logger *logrus.Logger) (*Client, error) {
	repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})

	if err == git.ErrRepositoryNotExists {
		// Initialize new repository
		repo, err = git.PlainInit(repoPath, false)
		if err != nil {
			return nil, utils.WrapError(err, utils.ErrorTypeGit, "initRepository", "failed to initialize git repository")
		}
		logger.Info("Initialized new git repository")
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

func (c *Client) AutoCommit(files []string, metadata map[string]string) error {
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

	// Create commit message
	message := c.formatCommitMessage(files, metadata)

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

	c.logger.Infof("Created commit: %s", commit.String())

	// Auto push if enabled
	if c.config.AutoPush {
		return c.Push()
	}

	return nil
}

func (c *Client) formatCommitMessage(files []string, metadata map[string]string) string {
	// Use template with variable substitution
	message := c.config.CommitTemplate

	// Replace variables
	replacements := map[string]string{
		"{timestamp}": time.Now().Format(time.RFC3339),
		"{files}":     fmt.Sprintf("%d file(s)", len(files)),
		"{branch}":    c.config.Branch,
	}

	for key, value := range metadata {
		replacements["{"+key+"}"] = value
	}

	for key, value := range replacements {
		message = strings.ReplaceAll(message, key, value)
	}

	return message
}

func (c *Client) Push() error {
	err := c.repo.Push(&git.PushOptions{
		RemoteName: "origin",
	})

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			c.logger.Debug("Repository already up to date")
			return nil
		}
		return utils.WrapError(err, utils.ErrorTypeGit, "push", "failed to push to remote repository")
	}

	c.logger.Info("Successfully pushed to origin")
	return nil
}

func (c *Client) Pull() error {
	err := c.worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
	})

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			c.logger.Debug("Repository already up to date")
			return nil
		}
		return utils.WrapError(err, utils.ErrorTypeGit, "pull", "failed to pull from remote repository")
	}

	c.logger.Info("Successfully pulled from origin")
	return nil
}

func (c *Client) GetStatus() (git.Status, error) {
	return c.worktree.Status()
}

func (c *Client) GetHistory(limit int) ([]*object.Commit, error) {
	ref, err := c.repo.Head()
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorTypeGit, "getHEAD", "failed to get HEAD reference")
	}

	iter, err := c.repo.Log(&git.LogOptions{
		From: ref.Hash(),
	})
	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorTypeGit, "getCommitHistory", "failed to get commit history")
	}
	defer iter.Close()

	var commits []*object.Commit
	count := 0

	err = iter.ForEach(func(commit *object.Commit) error {
		if limit > 0 && count >= limit {
			return nil
		}
		commits = append(commits, commit)
		count++
		return nil
	})

	if err != nil {
		return nil, utils.WrapError(err, utils.ErrorTypeGit, "iterateCommits", "failed to iterate commits")
	}

	return commits, nil
}

func (c *Client) GetCurrentBranch() (string, error) {
	ref, err := c.repo.Head()
	if err != nil {
		return "", utils.WrapError(err, utils.ErrorTypeGit, "getCurrentBranch", "failed to get HEAD reference")
	}

	if ref.Name().IsBranch() {
		return ref.Name().Short(), nil
	}

	return "", utils.NewError(utils.ErrorTypeGit, "getCurrentBranch", "HEAD is not a branch")
}

func (c *Client) CreateBranch(name string) error {
	headRef, err := c.repo.Head()
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeGit, "createBranch", "failed to get HEAD reference")
	}

	ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName(name), headRef.Hash())
	err = c.repo.Storer.SetReference(ref)
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeGit, "createBranch", "failed to create branch")
	}

	c.logger.Infof("Created branch: %s", name)
	return nil
}

func (c *Client) CheckoutBranch(name string) error {
	err := c.worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(name),
	})
	if err != nil {
		return utils.WrapError(err, utils.ErrorTypeGit, "checkoutBranch",
			fmt.Sprintf("failed to checkout branch %s", name))
	}

	c.logger.Infof("Checked out branch: %s", name)
	return nil
}
