package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureManagedFile(t *testing.T) {
	t.Run("creates file when missing", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".gitignore")

		require.NoError(t, ensureManagedFile(path, "# block\nfoo/\n"))

		raw, err := os.ReadFile(path) // #nosec G304 - test path
		require.NoError(t, err)
		s := string(raw)
		assert.Contains(t, s, managedBlockMarker)
		assert.Contains(t, s, "foo/")
		assert.Contains(t, s, managedBlockEnd)
	})

	t.Run("preserves user content and appends block", func(t *testing.T) {
		// A user-authored .gitignore must NOT be overwritten — its lines
		// stay verbatim and the GitCells block is appended at the end,
		// bracketed by the managed-block markers.
		dir := t.TempDir()
		path := filepath.Join(dir, ".gitignore")

		userContent := "node_modules/\n*.log\nsecrets.env\n"
		require.NoError(t, os.WriteFile(path, []byte(userContent), 0o644))

		require.NoError(t, ensureManagedFile(path, "# gitcells\n.gitcells/logs/\n"))

		raw, err := os.ReadFile(path) // #nosec G304 - test path
		require.NoError(t, err)
		s := string(raw)
		// User content survives.
		assert.Contains(t, s, "node_modules/")
		assert.Contains(t, s, "*.log")
		assert.Contains(t, s, "secrets.env")
		// Managed block landed.
		assert.Contains(t, s, managedBlockMarker)
		assert.Contains(t, s, ".gitcells/logs/")
		// Block lands AFTER user content.
		assert.Greater(t, strings.Index(s, managedBlockMarker), strings.Index(s, "secrets.env"))
	})

	t.Run("re-run is idempotent when marker is present", func(t *testing.T) {
		// Once the managed block is in place, a second call must not
		// double-append. The file's byte count stays identical.
		dir := t.TempDir()
		path := filepath.Join(dir, ".gitignore")

		require.NoError(t, ensureManagedFile(path, "# block\nfoo/\n"))
		first, err := os.ReadFile(path) // #nosec G304 - test path
		require.NoError(t, err)

		require.NoError(t, ensureManagedFile(path, "# block\nfoo/\n"))
		second, err := os.ReadFile(path) // #nosec G304 - test path
		require.NoError(t, err)

		assert.Equal(t, string(first), string(second), "re-run must be a no-op once marker is in file")
	})

	t.Run("adds leading newline when prior content lacks one", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".gitignore")
		// Note: no trailing newline.
		require.NoError(t, os.WriteFile(path, []byte("custom-line"), 0o644))

		require.NoError(t, ensureManagedFile(path, "# gitcells\nfoo/\n"))

		raw, err := os.ReadFile(path) // #nosec G304 - test path
		require.NoError(t, err)
		s := string(raw)
		// The original line is still there, immediately followed by a
		// newline then the marker.
		assert.True(t, strings.HasPrefix(s, "custom-line\n"+managedBlockMarker),
			"managed block must start on its own line; got: %q", s)
	})
}

// TestEnableGitLongPaths verifies that initializing a repository writes
// `core.longpaths = true` to the on-disk git config. The setting is a
// no-op on macOS/Linux but lifts Windows' default 260-char path limit so
// deeply nested `.gitcells/data/...` chunk dirs don't trigger silent
// `git add` failures. Asserting on the config file makes this OS-agnostic.
func TestEnableGitLongPaths(t *testing.T) {
	dir := t.TempDir()

	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)

	require.NoError(t, enableGitLongPaths(repo))

	configPath := filepath.Join(dir, ".git", "config")
	raw, err := os.ReadFile(configPath) // #nosec G304 - test path
	require.NoError(t, err)

	got := string(raw)
	assert.Contains(t, got, "longpaths = true",
		"core.longpaths must be true; got config:\n%s", got)

	// Re-applying is idempotent and must not duplicate the option.
	require.NoError(t, enableGitLongPaths(repo))
	raw2, err := os.ReadFile(configPath) // #nosec G304 - test path
	require.NoError(t, err)
	assert.Equal(t, strings.Count(string(raw), "longpaths"),
		strings.Count(string(raw2), "longpaths"),
		"re-applying must not duplicate the longpaths entry")
}
