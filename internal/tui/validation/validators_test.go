package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDirectory(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (string, func())
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty path",
			setup: func() (string, func()) {
				return "", func() {}
			},
			expectError: true,
			errorMsg:    "directory path cannot be empty",
		},
		{
			name: "writable directory",
			setup: func() (string, func()) {
				dir, err := os.MkdirTemp("", "gitcells_test_*")
				require.NoError(t, err)
				return dir, func() { os.RemoveAll(dir) }
			},
			expectError: false,
		},
		{
			name: "non-writable directory",
			setup: func() (string, func()) {
				dir, err := os.MkdirTemp("", "gitcells_test_*")
				require.NoError(t, err)
				err = os.Chmod(dir, 0555) // read-only
				require.NoError(t, err)
				return dir, func() {
					_ = os.Chmod(dir, 0755)
					os.RemoveAll(dir)
				}
			},
			expectError: true,
			errorMsg:    "directory is not writable",
		},
		{
			name: "non-existent directory with writable parent",
			setup: func() (string, func()) {
				parent, err := os.MkdirTemp("", "gitcells_test_*")
				require.NoError(t, err)
				newDir := filepath.Join(parent, "new_dir")
				return newDir, func() { os.RemoveAll(parent) }
			},
			expectError: false,
		},
		{
			name: "non-existent directory with non-writable parent",
			setup: func() (string, func()) {
				parent, err := os.MkdirTemp("", "gitcells_test_*")
				require.NoError(t, err)
				err = os.Chmod(parent, 0555) // read-only
				require.NoError(t, err)
				newDir := filepath.Join(parent, "new_dir")
				return newDir, func() {
					_ = os.Chmod(parent, 0755)
					os.RemoveAll(parent)
				}
			},
			expectError: true,
			errorMsg:    "cannot create directory (parent not writable)",
		},
		{
			name: "file instead of directory",
			setup: func() (string, func()) {
				file, err := os.CreateTemp("", "gitcells_test_*.txt")
				require.NoError(t, err)
				file.Close()
				return file.Name(), func() { os.Remove(file.Name()) }
			},
			expectError: true,
			errorMsg:    "path exists but is not a directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := tt.setup()
			defer cleanup()
			err := ValidateDirectory(path)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestInspectDirectory(t *testing.T) {
	t.Run("writable directory", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "gitcells_test_*")
		require.NoError(t, err)
		defer os.RemoveAll(dir)
		info, err := InspectDirectory(dir)
		assert.NoError(t, err)
		assert.True(t, info.Exists)
		assert.True(t, info.IsDirectory)
		assert.True(t, info.IsWritable)
		assert.False(t, info.IsGitRepo)
		assert.False(t, info.HasGitCells)
	})
	t.Run("non-writable directory", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "gitcells_test_*")
		require.NoError(t, err)
		defer func() {
			_ = os.Chmod(dir, 0755)
			os.RemoveAll(dir)
		}()
		err = os.Chmod(dir, 0555) // read-only
		require.NoError(t, err)
		info, err := InspectDirectory(dir)
		assert.NoError(t, err)
		assert.True(t, info.Exists)
		assert.True(t, info.IsDirectory)
		assert.False(t, info.IsWritable)
	})
	t.Run("non-existent directory", func(t *testing.T) {
		info, err := InspectDirectory("/non/existent/path")
		assert.NoError(t, err)
		assert.False(t, info.Exists)
		assert.False(t, info.IsDirectory)
		assert.False(t, info.IsWritable)
	})
}
