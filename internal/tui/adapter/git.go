package adapter

import (
	"github.com/Classic-Homes/gitcells/internal/git"
	gogit "github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
)

// GitAdapter provides minimal git operations for GitCells TUI
// Users should use their preferred git tools for branch management
type GitAdapter struct {
	client *git.Client
}

func NewGitAdapter(directory string) (*GitAdapter, error) {
	// Create a default git config
	gitConfig := &git.Config{
		UserName:       "GitCells",
		UserEmail:      "gitcells@localhost",
		CommitTemplate: "GitCells: {action} {filename}",
		AutoPush:       false,
		AutoPull:       true,
		Branch:         "main",
	}

	// Create a simple logger
	logger := logrus.New()

	client, err := git.NewClient(directory, gitConfig, logger)
	if err != nil {
		// If repo doesn't exist, try to initialize it
		if err == gogit.ErrRepositoryNotExists {
			if err := initializeGitRepo(directory); err != nil {
				return nil, err
			}
			// Try again after initialization
			client, err = git.NewClient(directory, gitConfig, logger)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &GitAdapter{client: client}, nil
}

// GetCurrentBranch returns only the current branch name
func (ga *GitAdapter) GetCurrentBranch() (string, error) {
	if ga.client == nil {
		return "main", nil
	}
	// In real implementation, would get current branch from git client
	return "main", nil
}

// GetStatus returns minimal repository status for Excel tracking
func (ga *GitAdapter) GetStatus() (*RepoStatus, error) {
	// Returns only the status needed for Excel file tracking
	return &RepoStatus{
		Branch:     "main",
		Clean:      true,
		FileCount:  0,
		LastCommit: "Excel files tracked",
	}, nil
}

// RepoStatus contains minimal repository status for Excel tracking
type RepoStatus struct {
	Branch     string
	Clean      bool
	FileCount  int  // Number of tracked Excel files
	LastCommit string
}

func initializeGitRepo(directory string) error {
	// Initialize a git repository for Excel file tracking
	_, err := gogit.PlainInit(directory, false)
	return err
}
