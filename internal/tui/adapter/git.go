package adapter

import (
	"github.com/Classic-Homes/gitcells/internal/git"
	"github.com/sirupsen/logrus"
)

// GitAdapter provides minimal git operations for GitCells TUI
// Users should use their preferred git tools for all git operations
type GitAdapter struct {
	client *git.Client
}

func NewGitAdapter(directory string) (*GitAdapter, error) {
	// Create a minimal git config for basic commit operations only
	gitConfig := &git.Config{
		UserName:  "GitCells",
		UserEmail: "gitcells@localhost",
	}

	// Create a simple logger
	logger := logrus.New()

	client, err := git.NewClient(directory, gitConfig, logger)
	// Client may be nil if no git repo exists - that's fine
	return &GitAdapter{client: client}, err
}

// InGitRepository returns true if we're in a git repository
func (ga *GitAdapter) InGitRepository() bool {
	return ga.client != nil && ga.client.InGitRepository()
}

// IsClean returns true if the working directory is clean
func (ga *GitAdapter) IsClean() (bool, error) {
	if ga.client == nil {
		return true, nil // No git = clean
	}
	return ga.client.IsClean()
}

// CommitFiles commits the specified files with a simple message
func (ga *GitAdapter) CommitFiles(files []string, message string) error {
	if ga.client == nil {
		return nil // No git repo - skip commit
	}
	return ga.client.AutoCommit(files, message)
}
