package adapter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitAdapter(t *testing.T) {
	t.Run("creates git adapter in non-git directory", func(t *testing.T) {
		tempDir := t.TempDir()

		adapter, err := NewGitAdapter(tempDir)
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.Nil(t, adapter.client) // No git repo, so client should be nil
	})

	t.Run("creates git adapter in git directory", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		adapter, err := NewGitAdapter(tempDir)
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.NotNil(t, adapter.client) // Git repo exists, so client should be present
	})

	t.Run("handles invalid directory", func(t *testing.T) {
		adapter, err := NewGitAdapter("/invalid\x00path")
		// Should still create adapter even if directory is invalid
		assert.NotNil(t, adapter)
		// Error behavior may vary based on implementation
		if err != nil {
			assert.NotEmpty(t, err.Error())
		}
	})
}

func TestGitAdapter_InGitRepository(t *testing.T) {
	t.Run("returns false for non-git directory", func(t *testing.T) {
		tempDir := t.TempDir()

		adapter, err := NewGitAdapter(tempDir)
		require.NoError(t, err)

		assert.False(t, adapter.InGitRepository())
	})

	t.Run("returns true for git directory", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		adapter, err := NewGitAdapter(tempDir)
		require.NoError(t, err)

		assert.True(t, adapter.InGitRepository())
	})
}

func TestGitAdapter_IsClean(t *testing.T) {
	t.Run("returns true for non-git directory", func(t *testing.T) {
		tempDir := t.TempDir()

		adapter, err := NewGitAdapter(tempDir)
		require.NoError(t, err)

		clean, err := adapter.IsClean()
		assert.NoError(t, err)
		assert.True(t, clean)
	})

	t.Run("returns true for clean git repository", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		adapter, err := NewGitAdapter(tempDir)
		require.NoError(t, err)

		clean, err := adapter.IsClean()
		assert.NoError(t, err)
		assert.True(t, clean)
	})

	t.Run("returns false for dirty git repository", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		// Create an untracked file
		testFile := filepath.Join(tempDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0600)
		require.NoError(t, err)

		adapter, err := NewGitAdapter(tempDir)
		require.NoError(t, err)

		clean, err := adapter.IsClean()
		assert.NoError(t, err)
		assert.False(t, clean)
	})
}

func TestGitAdapter_CommitFiles(t *testing.T) {
	t.Run("does nothing for non-git directory", func(t *testing.T) {
		tempDir := t.TempDir()

		adapter, err := NewGitAdapter(tempDir)
		require.NoError(t, err)

		err = adapter.CommitFiles([]string{"test.txt"}, "test commit")
		assert.NoError(t, err) // Should not error when no git repo
	})

	t.Run("commits files in git repository", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		// Create a test file
		testFile := filepath.Join(tempDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0600)
		require.NoError(t, err)

		adapter, err := NewGitAdapter(tempDir)
		require.NoError(t, err)

		err = adapter.CommitFiles([]string{testFile}, "Add test file")
		assert.NoError(t, err)

		// Verify commit was created
		ref, err := repo.Head()
		assert.NoError(t, err)

		commit, err := repo.CommitObject(ref.Hash())
		assert.NoError(t, err)
		assert.Equal(t, "Add test file", commit.Message)
		assert.Equal(t, "GitCells", commit.Author.Name)
		assert.Equal(t, "gitcells@localhost", commit.Author.Email)
	})

	t.Run("handles commit with non-existent file", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		adapter, err := NewGitAdapter(tempDir)
		require.NoError(t, err)

		err = adapter.CommitFiles([]string{filepath.Join(tempDir, "nonexistent.txt")}, "test commit")
		assert.Error(t, err) // Should error when trying to stage non-existent file
	})

	t.Run("handles empty file list", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		adapter, err := NewGitAdapter(tempDir)
		require.NoError(t, err)

		err = adapter.CommitFiles([]string{}, "empty commit")
		assert.NoError(t, err) // Should not error with empty file list
	})
}

func TestGitAdapter_Integration(t *testing.T) {
	t.Run("full workflow with git operations", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		adapter, err := NewGitAdapter(tempDir)
		require.NoError(t, err)

		// Verify initial state
		assert.True(t, adapter.InGitRepository())

		clean, err := adapter.IsClean()
		require.NoError(t, err)
		assert.True(t, clean)

		// Create and commit first file
		file1 := filepath.Join(tempDir, "file1.txt")
		err = os.WriteFile(file1, []byte("content 1"), 0600)
		require.NoError(t, err)

		// Repository should now be dirty
		clean, err = adapter.IsClean()
		require.NoError(t, err)
		assert.False(t, clean)

		// Commit the file
		err = adapter.CommitFiles([]string{file1}, "Add file1")
		require.NoError(t, err)

		// Repository should be clean again
		clean, err = adapter.IsClean()
		require.NoError(t, err)
		assert.True(t, clean)

		// Create and commit second file
		file2 := filepath.Join(tempDir, "file2.txt")
		err = os.WriteFile(file2, []byte("content 2"), 0600)
		require.NoError(t, err)

		err = adapter.CommitFiles([]string{file2}, "Add file2")
		require.NoError(t, err)

		// Verify we have 2 commits
		iter, err := repo.Log(&git.LogOptions{})
		require.NoError(t, err)

		commitCount := 0
		err = iter.ForEach(func(c *object.Commit) error {
			commitCount++
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 2, commitCount)
	})
}
