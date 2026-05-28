//go:build windows

package converter

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// TestOpenFileWithRetry_WindowsSharingViolation verifies the retry loop
// recovers from a real ERROR_SHARING_VIOLATION — the failure mode Excel
// produces during its atomic save on Windows when it holds the target
// file open through the rename phase.
//
// Setup: open the file with FILE_SHARE_NONE so any concurrent open
// returns ERROR_SHARING_VIOLATION (errno 32). Then call
// openFileWithRetry in another goroutine; it must keep retrying until
// we release the lock.
func TestOpenFileWithRetry_WindowsSharingViolation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "locked.xlsx")
	require.NoError(t, os.WriteFile(path, []byte("payload"), 0o644))

	pathPtr, err := syscall.UTF16PtrFromString(path)
	require.NoError(t, err)

	// Open with no sharing — concurrent opens must fail until we close.
	h, err := windows.CreateFile(
		pathPtr,
		windows.GENERIC_READ,
		0, // dwShareMode = 0 → exclusive
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	require.NoError(t, err)

	var (
		wg          sync.WaitGroup
		retryErr    error
		retryHandle *os.File
		retryStart  = time.Now()
		retryEnd    time.Time
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		retryHandle, retryErr = openFileWithRetry(path, 3*time.Second)
		retryEnd = time.Now()
	}()

	// Hold the lock long enough that openFileWithRetry must back off at
	// least once before succeeding.
	time.Sleep(300 * time.Millisecond)
	require.NoError(t, windows.CloseHandle(h))

	wg.Wait()

	require.NoError(t, retryErr, "openFileWithRetry must recover after the lock is released")
	require.NotNil(t, retryHandle)
	defer retryHandle.Close()

	elapsed := retryEnd.Sub(retryStart)
	assert.GreaterOrEqual(t, elapsed, 250*time.Millisecond,
		"retry should have waited at least one backoff before succeeding")
	assert.Less(t, elapsed, 3*time.Second,
		"retry should not have exhausted the full timeout budget")
}
