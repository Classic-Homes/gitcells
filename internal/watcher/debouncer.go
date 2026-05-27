// Package watcher provides file system monitoring functionality for the GitCells application.
package watcher

import (
	"runtime/debug"
	"sync"
	"time"
)

// Debouncer prevents rapid successive calls by delaying execution.
//
// Lifecycle invariant: every accepted Debounce call increments wg by 1.
// That increment is decremented exactly once — either when the timer's
// callback completes, or when the timer is cancelled (by a replacement
// call for the same key, or by Drain timing out). This is what lets
// Drain wait for in-flight work without racing.
type Debouncer struct {
	delay    time.Duration
	timers   map[string]*time.Timer
	mu       sync.Mutex
	wg       sync.WaitGroup
	closed   bool
	panicLog func(r interface{}, stack []byte)
}

// NewDebouncer creates a new debouncer with the specified delay.
func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{
		delay:  delay,
		timers: make(map[string]*time.Timer),
	}
}

// SetPanicLogger installs a callback invoked when a debounced fn panics.
// If unset, a recovered panic is silently swallowed (still preferable to
// taking down the watcher goroutine).
func (d *Debouncer) SetPanicLogger(f func(r interface{}, stack []byte)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.panicLog = f
}

// Debounce delays the execution of fn until delay has passed since the last call.
// After Drain has been called, further Debounce calls are dropped and fn is
// never run — this is required so shutdown is bounded.
func (d *Debouncer) Debounce(key string, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}

	// Cancel any prior timer for this key. If Stop returns true the
	// callback won't run, so we must release the wg slot we added for
	// it. If Stop returns false the callback has already fired or is
	// running — it will Done() itself.
	if existing, ok := d.timers[key]; ok {
		if existing.Stop() {
			d.wg.Done()
		}
		delete(d.timers, key)
	}

	d.wg.Add(1)

	var timer *time.Timer
	timer = time.AfterFunc(d.delay, func() {
		defer d.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				d.mu.Lock()
				log := d.panicLog
				d.mu.Unlock()
				if log != nil {
					log(r, debug.Stack())
				}
			}
		}()

		fn()

		d.mu.Lock()
		if cur, ok := d.timers[key]; ok && cur == timer {
			delete(d.timers, key)
		}
		d.mu.Unlock()
	})

	d.timers[key] = timer
}

// Drain stops accepting new work, then waits up to maxWait for all
// pending and in-flight callbacks to finish. Returns the keys of any
// callbacks that were force-cancelled because they didn't complete in
// time. Safe to call multiple times; subsequent calls are no-ops.
func (d *Debouncer) Drain(maxWait time.Duration) []string {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil
	}
	d.closed = true
	d.mu.Unlock()

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(maxWait):
	}

	// Grace expired — cancel anything still pending. Already-running
	// callbacks keep running; we wait for them below.
	dropped := make([]string, 0)
	d.mu.Lock()
	for key, t := range d.timers {
		if t.Stop() {
			d.wg.Done()
			dropped = append(dropped, key)
		}
		delete(d.timers, key)
	}
	d.mu.Unlock()

	d.wg.Wait()
	return dropped
}

// PendingCount returns the number of keys with pending or in-flight
// callbacks. Best-effort: a value read here can be stale by the time
// the caller acts on it. Useful for status displays, not for control flow.
func (d *Debouncer) PendingCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.timers)
}
