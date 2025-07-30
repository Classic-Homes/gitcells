// Package watcher provides file system monitoring functionality for the GitCells application.
package watcher

import (
	"sync"
	"time"
)

// Debouncer prevents rapid successive calls by delaying execution
type Debouncer struct {
	delay  time.Duration
	timers sync.Map
	mu     sync.Mutex
}

// NewDebouncer creates a new debouncer with the specified delay
func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{
		delay: delay,
	}
}

// Debounce delays the execution of fn until delay has passed since the last call
func (d *Debouncer) Debounce(key string, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Cancel existing timer if it exists
	if timer, exists := d.timers.Load(key); exists {
		timer.(*time.Timer).Stop()
	}

	// Create new timer
	timer := time.AfterFunc(d.delay, func() {
		fn()
		d.timers.Delete(key)
	})

	d.timers.Store(key, timer)
}

// Cancel cancels the debounced function for the given key
func (d *Debouncer) Cancel(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if timer, exists := d.timers.Load(key); exists {
		timer.(*time.Timer).Stop()
		d.timers.Delete(key)
	}
}

// CancelAll cancels all pending debounced functions
func (d *Debouncer) CancelAll() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.timers.Range(func(key, value interface{}) bool {
		timer := value.(*time.Timer)
		timer.Stop()
		d.timers.Delete(key)
		return true
	})
}

// HasPending returns true if there are any pending debounced functions
func (d *Debouncer) HasPending() bool {
	hasPending := false
	d.timers.Range(func(key, value interface{}) bool {
		hasPending = true
		return false // Stop iteration
	})
	return hasPending
}

// GetPendingKeys returns all keys that have pending debounced functions
func (d *Debouncer) GetPendingKeys() []string {
	var keys []string
	d.timers.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			keys = append(keys, keyStr)
		}
		return true
	})
	return keys
}

// SetDelay updates the delay for future debounce calls
func (d *Debouncer) SetDelay(delay time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.delay = delay
}

// GetDelay returns the current delay setting
func (d *Debouncer) GetDelay() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.delay
}
