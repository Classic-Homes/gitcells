package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveDirArg(t *testing.T) {
	t.Run("no arg falls back to cwd", func(t *testing.T) {
		dir, err := resolveDirArg(nil)
		require.NoError(t, err)
		assert.Equal(t, ".", dir)
	})

	t.Run("valid directory is returned as-is", func(t *testing.T) {
		tmp := t.TempDir()
		dir, err := resolveDirArg([]string{tmp})
		require.NoError(t, err)
		assert.Equal(t, tmp, dir)
	})

	t.Run("missing path errors with 'does not exist'", func(t *testing.T) {
		_, err := resolveDirArg([]string{"/definitely/not/a/real/path/abc123"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("path that isn't a directory errors with 'not a directory'", func(t *testing.T) {
		// Create a regular file in a tempdir, point resolveDirArg at it.
		tmp := t.TempDir()
		fpath := filepath.Join(tmp, "notes.txt")
		require.NoError(t, os.WriteFile(fpath, []byte("hi"), 0o644))

		_, err := resolveDirArg([]string{fpath})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})
}
