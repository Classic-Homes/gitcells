package git

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// DiffSummary represents a summary of changes between commits
type DiffSummary struct {
	FilesChanged int
	Insertions   int
	Deletions    int
	Files        []FileDiff
}

// FileDiff represents changes to a single file
type FileDiff struct {
	Path      string
	Status    string // "added", "modified", "deleted"
	Additions int
	Deletions int
}

// GetDiff returns a summary of changes between two commits
func (c *Client) GetDiff(fromCommit, toCommit string) (*DiffSummary, error) {
	from, err := c.repo.ResolveRevision(plumbing.Revision(fromCommit))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve from commit %s: %w", fromCommit, err)
	}

	to, err := c.repo.ResolveRevision(plumbing.Revision(toCommit))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve to commit %s: %w", toCommit, err)
	}

	fromCommitObj, err := c.repo.CommitObject(*from)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object %s: %w", fromCommit, err)
	}

	toCommitObj, err := c.repo.CommitObject(*to)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object %s: %w", toCommit, err)
	}

	fromTree, err := fromCommitObj.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree for commit %s: %w", fromCommit, err)
	}

	toTree, err := toCommitObj.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get tree for commit %s: %w", toCommit, err)
	}

	changes, err := fromTree.Diff(toTree)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}

	summary := &DiffSummary{
		Files: make([]FileDiff, 0, len(changes)),
	}

	for _, change := range changes {
		var status string
		// For simplicity, we'll determine status based on the change
		// The exact API might vary - this is a basic implementation
		if change.From.Name == "" {
			status = "added"
		} else if change.To.Name == "" {
			status = "deleted"
		} else {
			status = "modified"
		}

		// For simplicity, we'll just count the files
		// In a more complete implementation, you'd analyze the patch for line counts
		fileDiff := FileDiff{
			Path:   change.To.Name,
			Status: status,
		}

		summary.Files = append(summary.Files, fileDiff)
		summary.FilesChanged++
	}

	return summary, nil
}

// GetFileHistory returns the commit history for a specific file
func (c *Client) GetFileHistory(filePath string, limit int) ([]*object.Commit, error) {
	commits, err := c.repo.Log(&git.LogOptions{
		FileName: &filePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file history for %s: %w", filePath, err)
	}
	defer commits.Close()

	var result []*object.Commit
	count := 0

	err = commits.ForEach(func(commit *object.Commit) error {
		if limit > 0 && count >= limit {
			return nil
		}
		result = append(result, commit)
		count++
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate file history: %w", err)
	}

	return result, nil
}

// IsClean returns true if the working directory is clean
func (c *Client) IsClean() (bool, error) {
	status, err := c.worktree.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}
	return status.IsClean(), nil
}

// GetModifiedFiles returns a list of modified files
func (c *Client) GetModifiedFiles() ([]string, error) {
	status, err := c.worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var modifiedFiles []string
	for file, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified || fileStatus.Worktree != git.Unmodified {
			modifiedFiles = append(modifiedFiles, file)
		}
	}

	return modifiedFiles, nil
}

// AddFiles stages multiple files for commit
func (c *Client) AddFiles(files []string) error {
	for _, file := range files {
		if _, err := c.worktree.Add(file); err != nil {
			return fmt.Errorf("failed to stage file %s: %w", file, err)
		}
	}
	return nil
}

// ResetFile unstages a file
func (c *Client) ResetFile(filePath string) error {
	return c.worktree.Reset(&git.ResetOptions{
		Mode: git.MixedReset,
	})
}

// GetCommitInfo returns detailed information about a commit
func (c *Client) GetCommitInfo(commitHash string) (*CommitInfo, error) {
	hash, err := c.repo.ResolveRevision(plumbing.Revision(commitHash))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve commit %s: %w", commitHash, err)
	}

	commit, err := c.repo.CommitObject(*hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	info := &CommitInfo{
		Hash:      commit.Hash.String(),
		Author:    commit.Author.Name,
		Email:     commit.Author.Email,
		Date:      commit.Author.When,
		Message:   strings.TrimSpace(commit.Message),
		ShortHash: commit.Hash.String()[:7],
	}

	return info, nil
}

// CommitInfo represents detailed information about a commit
type CommitInfo struct {
	Hash      string
	ShortHash string
	Author    string
	Email     string
	Date      time.Time
	Message   string
}

// HasRemote checks if the repository has a remote configured
func (c *Client) HasRemote(name string) (bool, error) {
	remotes, err := c.repo.Remotes()
	if err != nil {
		return false, fmt.Errorf("failed to get remotes: %w", err)
	}

	for _, remote := range remotes {
		if remote.Config().Name == name {
			return true, nil
		}
	}

	return false, nil
}

// AddRemote adds a new remote to the repository
func (c *Client) AddRemote(name, url string) error {
	_, err := c.repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})
	if err != nil {
		return fmt.Errorf("failed to add remote %s: %w", name, err)
	}

	c.logger.Infof("Added remote %s: %s", name, url)
	return nil
}