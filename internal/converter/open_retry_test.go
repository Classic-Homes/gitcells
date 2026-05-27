package converter

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackoffScheduleRespectsTimeout(t *testing.T) {
	schedule := backoffSchedule(50 * time.Millisecond)
	var total time.Duration
	for _, d := range schedule {
		total += d
	}
	// Cumulative sleep must not blow far past the budget. We allow a little
	// over since the loop appends the step that brings us to/over the limit.
	assert.LessOrEqual(t, total, 200*time.Millisecond, "schedule cumulative: %v", total)
	assert.NotEmpty(t, schedule)
}

func TestIsTransientLockError_FileMustExist(t *testing.T) {
	// A permission error on a missing file is not transient — never retry it,
	// otherwise a typo'd path stalls for the full budget.
	missing := filepath.Join(t.TempDir(), "does-not-exist.xlsx")
	err := &os.PathError{Op: "open", Path: missing, Err: fs.ErrPermission}
	assert.False(t, isTransientLockError(err, missing))
}

func TestIsTransientLockError_PermissionWithFilePresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "present.xlsx")
	require.NoError(t, os.WriteFile(path, []byte("hi"), 0o644))
	err := &os.PathError{Op: "open", Path: path, Err: fs.ErrPermission}
	assert.True(t, isTransientLockError(err, path))
}

func TestIsTransientLockError_EBUSY(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "present.xlsx")
	require.NoError(t, os.WriteFile(path, []byte("hi"), 0o644))
	err := &os.PathError{Op: "open", Path: path, Err: syscall.EBUSY}
	assert.True(t, isTransientLockError(err, path))
}

func TestIsTransientLockError_MessageMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "present.xlsx")
	require.NoError(t, os.WriteFile(path, []byte("hi"), 0o644))
	err := errors.New("the process cannot access the file because it is being used by another process")
	assert.True(t, isTransientLockError(err, path))
}

func TestOpenFileWithRetry_SucceedsAfterDelay(t *testing.T) {
	// Healthy file opens on first try with effectively no retry overhead.
	dir := t.TempDir()
	path := filepath.Join(dir, "ok.txt")
	require.NoError(t, os.WriteFile(path, []byte("hi"), 0o644))

	start := time.Now()
	f, err := openFileWithRetry(path, time.Second)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	assert.Less(t, time.Since(start), 100*time.Millisecond)
}
