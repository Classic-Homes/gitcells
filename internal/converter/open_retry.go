package converter

import (
	"errors"
	"io/fs"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/xuri/excelize/v2"
)

// DefaultExcelOpenTimeout caps how long we'll keep retrying to open a
// locked Excel file. Excel on Windows holds the file open through the
// rename phase of its atomic-save; a couple of seconds is normal, ten
// is a generous ceiling that still bounds user-visible latency.
const DefaultExcelOpenTimeout = 10 * time.Second

// openExcelWithRetry wraps excelize.OpenFile with exponential-backoff
// retry on transient file-lock errors. Non-transient errors (file not
// found, malformed xlsx, etc.) return immediately.
func openExcelWithRetry(path string, timeout time.Duration) (*excelize.File, error) {
	var lastErr error
	for _, wait := range backoffSchedule(timeout) {
		f, err := excelize.OpenFile(path)
		if err == nil {
			return f, nil
		}
		lastErr = err
		if !isTransientLockError(err, path) {
			return nil, err
		}
		time.Sleep(wait)
	}
	return nil, lastErr
}

// openFileWithRetry is the os.Open analogue for the same retry policy.
// Used by checksum calculation, which races the rename phase too.
func openFileWithRetry(path string, timeout time.Duration) (*os.File, error) {
	var lastErr error
	for _, wait := range backoffSchedule(timeout) {
		f, err := os.Open(path) // #nosec G304 - path comes from watcher event
		if err == nil {
			return f, nil
		}
		lastErr = err
		if !isTransientLockError(err, path) {
			return nil, err
		}
		time.Sleep(wait)
	}
	return nil, lastErr
}

// backoffSchedule yields sleep durations whose cumulative sum stays at
// or below timeout. The first attempt is always made with zero delay,
// so a healthy file opens immediately.
func backoffSchedule(timeout time.Duration) []time.Duration {
	steps := []time.Duration{0, 50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond, 500 * time.Millisecond, 1 * time.Second}
	out := make([]time.Duration, 0, len(steps)+8)
	var total time.Duration
	for _, s := range steps {
		out = append(out, s)
		total += s
		if total >= timeout {
			return out
		}
	}
	for total < timeout {
		out = append(out, 1*time.Second)
		total += 1 * time.Second
	}
	return out
}

// isTransientLockError reports whether err is the kind of failure that
// can plausibly succeed if we wait a moment and try again. We require
// the file to still exist on disk — a true "not found" should not be
// retried.
func isTransientLockError(err error, path string) bool {
	if err == nil {
		return false
	}
	if _, statErr := os.Stat(path); statErr != nil {
		return false
	}

	if errors.Is(err, syscall.EBUSY) || errors.Is(err, syscall.EAGAIN) {
		return true
	}

	if runtime.GOOS == "windows" {
		// ERROR_SHARING_VIOLATION (32) and ERROR_LOCK_VIOLATION (33) both
		// surface as syscall.Errno on Windows. errors.Is doesn't bridge
		// to Errno values cleanly across platforms, so we check the
		// numeric code via type assertion on the wrapped PathError.
		var perr *os.PathError
		if errors.As(err, &perr) {
			if errno, ok := perr.Err.(syscall.Errno); ok {
				if errno == 32 || errno == 33 {
					return true
				}
			}
		}
	}

	if errors.Is(err, fs.ErrPermission) {
		return true
	}

	// excelize wraps lower-level errors in its own messages. The phrase
	// "being used by another process" is the canonical Windows wording
	// and shows up in excelize too.
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "being used by another process") ||
		strings.Contains(msg, "sharing violation") ||
		strings.Contains(msg, "resource temporarily unavailable") {
		return true
	}

	return false
}
