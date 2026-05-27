package git

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// TestAutoCommit_Concurrent exercises the mutex added to Client.AutoCommit.
// The watcher fires its handler in a separate goroutine per debounced file
// key, so when two Excel files are saved in quick succession the converter
// calls AutoCommit twice in parallel. go-git's worktree operations are not
// safe for that without serialization — historically this surfaced as
// "index.lock" or corrupted commits.
//
// This test forces the race: N goroutines each stage a distinct file and
// commit. After the storm, the repo must be clean, exactly N commits must
// have landed, and no .git/index.lock should remain.
func TestAutoCommit_Concurrent(t *testing.T) {
	const goroutines = 10

	tmp := t.TempDir()
	_, err := gogit.PlainInit(tmp, false)
	require.NoError(t, err)

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	client, err := NewClient(tmp, &Config{UserName: "test", UserEmail: "test@example.com"}, logger)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Pre-create N distinct files so each goroutine has its own to stage.
	files := make([]string, goroutines)
	for i := 0; i < goroutines; i++ {
		p := filepath.Join(tmp, fmt.Sprintf("file%d.txt", i))
		require.NoError(t, os.WriteFile(p, []byte(fmt.Sprintf("content %d\n", i)), 0o600))
		files[i] = p
	}

	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			if err := client.AutoCommit([]string{files[i]}, fmt.Sprintf("commit %d", i)); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent AutoCommit returned error: %v", err)
	}

	// No leftover lock file.
	_, statErr := os.Stat(filepath.Join(tmp, ".git", "index.lock"))
	require.Truef(t, os.IsNotExist(statErr), "stale .git/index.lock present after concurrent commits")

	// Repo must end clean.
	clean, err := client.IsClean()
	require.NoError(t, err)
	require.True(t, clean, "repo not clean after concurrent commits")

	// All N commits should be on HEAD's history.
	repo, err := gogit.PlainOpen(tmp)
	require.NoError(t, err)
	head, err := repo.Head()
	require.NoError(t, err)
	iter, err := repo.Log(&gogit.LogOptions{From: head.Hash()})
	require.NoError(t, err)
	count := 0
	require.NoError(t, iter.ForEach(func(_ *object.Commit) error {
		count++
		return nil
	}))
	require.Equalf(t, goroutines, count, "expected %d commits, got %d", goroutines, count)
}
