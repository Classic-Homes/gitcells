package git

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := &Config{
		UserName:  "Test User",
		UserEmail: "test@example.com",
	}

	t.Run("non-existent repository returns nil", func(t *testing.T) {
		tempDir := t.TempDir()
		nonExistentPath := filepath.Join(tempDir, "nonexistent")

		client, err := NewClient(nonExistentPath, config, logger)
		assert.NoError(t, err)
		assert.Nil(t, client)
	})

	t.Run("valid git repository", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)
		require.NotNil(t, repo)

		client, err := NewClient(tempDir, config, logger)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, config, client.config)
		assert.Equal(t, logger, client.logger)
		assert.NotNil(t, client.repo)
		assert.NotNil(t, client.worktree)
	})

	t.Run("invalid path", func(t *testing.T) {
		client, err := NewClient("/invalid\x00path", config, logger)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestClient_AutoCommit(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := &Config{
		UserName:  "Test User",
		UserEmail: "test@example.com",
	}

	t.Run("nil client returns no error", func(t *testing.T) {
		var client *Client
		err := client.AutoCommit([]string{"test.json"}, "test commit")
		assert.NoError(t, err)
	})

	t.Run("successful commit with files", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		client, err := NewClient(tempDir, config, logger)
		require.NoError(t, err)
		require.NotNil(t, client)

		// Create a test file to commit
		testFile := filepath.Join(tempDir, "test.json")
		err = os.WriteFile(testFile, []byte(`{"test": "data"}`), 0600)
		require.NoError(t, err)

		// Commit the file
		err = client.AutoCommit([]string{testFile}, "Add test JSON file")
		assert.NoError(t, err)

		// Verify commit exists
		ref, err := repo.Head()
		assert.NoError(t, err)
		assert.NotNil(t, ref)

		commit, err := repo.CommitObject(ref.Hash())
		assert.NoError(t, err)
		assert.Equal(t, "Add test JSON file", commit.Message)
		assert.Equal(t, "Test User", commit.Author.Name)
		assert.Equal(t, "test@example.com", commit.Author.Email)
	})

	t.Run("commit with default message", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		repo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		client, err := NewClient(tempDir, config, logger)
		require.NoError(t, err)
		require.NotNil(t, client)

		// Create test files
		testFile1 := filepath.Join(tempDir, "test1.json")
		testFile2 := filepath.Join(tempDir, "test2.json")
		err = os.WriteFile(testFile1, []byte(`{"test": "data1"}`), 0600)
		require.NoError(t, err)
		err = os.WriteFile(testFile2, []byte(`{"test": "data2"}`), 0600)
		require.NoError(t, err)

		// Commit with empty message
		err = client.AutoCommit([]string{testFile1, testFile2}, "")
		assert.NoError(t, err)

		// Verify default message was used
		ref, err := repo.Head()
		assert.NoError(t, err)
		commit, err := repo.CommitObject(ref.Hash())
		assert.NoError(t, err)
		assert.Equal(t, "GitCells: Update JSON representations (2 files)", commit.Message)
	})

	t.Run("no changes to commit", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		client, err := NewClient(tempDir, config, logger)
		require.NoError(t, err)
		require.NotNil(t, client)

		// Try to commit non-existent files
		err = client.AutoCommit([]string{}, "test commit")
		assert.NoError(t, err)
	})

	t.Run("stage non-existent file", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		client, err := NewClient(tempDir, config, logger)
		require.NoError(t, err)
		require.NotNil(t, client)

		// Try to commit non-existent file
		nonExistentFile := filepath.Join(tempDir, "nonexistent.json")
		err = client.AutoCommit([]string{nonExistentFile}, "test commit")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to stage file")
	})
}

func TestClient_IsClean(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := &Config{
		UserName:  "Test User",
		UserEmail: "test@example.com",
	}

	t.Run("nil client is clean", func(t *testing.T) {
		var client *Client
		clean, err := client.IsClean()
		assert.NoError(t, err)
		assert.True(t, clean)
	})

	t.Run("clean repository", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		client, err := NewClient(tempDir, config, logger)
		require.NoError(t, err)
		require.NotNil(t, client)

		clean, err := client.IsClean()
		assert.NoError(t, err)
		assert.True(t, clean)
	})

	t.Run("dirty repository", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		client, err := NewClient(tempDir, config, logger)
		require.NoError(t, err)
		require.NotNil(t, client)

		// Create an untracked file
		testFile := filepath.Join(tempDir, "test.json")
		err = os.WriteFile(testFile, []byte(`{"test": "data"}`), 0600)
		require.NoError(t, err)

		clean, err := client.IsClean()
		assert.NoError(t, err)
		assert.False(t, clean)
	})
}

func TestClient_InGitRepository(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := &Config{
		UserName:  "Test User",
		UserEmail: "test@example.com",
	}

	t.Run("nil client not in git repository", func(t *testing.T) {
		var client *Client
		assert.False(t, client.InGitRepository())
	})

	t.Run("valid client in git repository", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize git repository
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		client, err := NewClient(tempDir, config, logger)
		require.NoError(t, err)
		require.NotNil(t, client)

		assert.True(t, client.InGitRepository())
	})
}

func TestConfig(t *testing.T) {
	t.Run("config struct fields", func(t *testing.T) {
		config := &Config{
			UserName:       "John Doe",
			UserEmail:      "john@example.com",
			CommitTemplate: "GitCells: ${files}",
		}

		assert.Equal(t, "John Doe", config.UserName)
		assert.Equal(t, "john@example.com", config.UserEmail)
		assert.Equal(t, "GitCells: ${files}", config.CommitTemplate)
	})
}

func TestIntegrationWorkflow(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := &Config{
		UserName:  "Integration Test",
		UserEmail: "integration@example.com",
	}

	tempDir := t.TempDir()

	// Initialize git repository
	repo, err := git.PlainInit(tempDir, false)
	require.NoError(t, err)

	client, err := NewClient(tempDir, config, logger)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify initially clean
	clean, err := client.IsClean()
	assert.NoError(t, err)
	assert.True(t, clean)

	// Create multiple JSON files
	files := []string{}
	for i := 0; i < 3; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("test%d.json", i))
		content := fmt.Sprintf(`{"test": "data%d"}`, i)
		err = os.WriteFile(filename, []byte(content), 0600)
		require.NoError(t, err)
		files = append(files, filename)
	}

	// Repository should now be dirty
	clean, err = client.IsClean()
	assert.NoError(t, err)
	assert.False(t, clean)

	// Commit all files
	err = client.AutoCommit(files, "Add test JSON files")
	assert.NoError(t, err)

	// Repository should be clean again
	clean, err = client.IsClean()
	assert.NoError(t, err)
	assert.True(t, clean)

	// Verify commit exists in history
	ref, err := repo.Head()
	assert.NoError(t, err)
	commit, err := repo.CommitObject(ref.Hash())
	assert.NoError(t, err)
	assert.Equal(t, "Add test JSON files", commit.Message)
	assert.Equal(t, "Integration Test", commit.Author.Name)
	assert.Equal(t, "integration@example.com", commit.Author.Email)
}
