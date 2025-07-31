package adapter

import (
	"github.com/Classic-Homes/gitcells/internal/git"
	gogit "github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
)

// GitAdapter bridges the TUI with the git package
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

// GetBranches returns list of branches with their status
func (ga *GitAdapter) GetBranches() ([]BranchInfo, error) {
	// This would be implemented using the git client
	// For now, return mock data
	return []BranchInfo{
		{Name: "main", Current: true, HasChanges: false},
		{Name: "feature/updates", Current: false, HasChanges: true},
	}, nil
}

// GetStatus returns current repository status
func (ga *GitAdapter) GetStatus() (*RepoStatus, error) {
	// This would use the git client to get actual status
	// For now, return mock data
	return &RepoStatus{
		Branch:     "main",
		Clean:      true,
		FileCount:  0,
		LastCommit: "Initial commit",
	}, nil
}

// BranchInfo contains information about a git branch
type BranchInfo struct {
	Name       string
	Current    bool
	HasChanges bool
	Ahead      int
	Behind     int
}

// RepoStatus contains current repository status
type RepoStatus struct {
	Branch     string
	Clean      bool
	FileCount  int
	LastCommit string
}

func initializeGitRepo(directory string) error {
	_, err := gogit.PlainInit(directory, false)
	return err
}