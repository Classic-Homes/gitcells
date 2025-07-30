package watcher

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDebouncer(t *testing.T) {
	delay := 100 * time.Millisecond
	debouncer := NewDebouncer(delay)

	assert.NotNil(t, debouncer)
	assert.Equal(t, delay, debouncer.delay)
}

func TestDebouncer_SingleCall(t *testing.T) {
	debouncer := NewDebouncer(100 * time.Millisecond)

	var called bool
	var mu sync.Mutex
	debouncer.Debounce("test", func() {
		mu.Lock()
		called = true
		mu.Unlock()
	})

	// Should not be called immediately
	mu.Lock()
	assert.False(t, called)
	mu.Unlock()

	// Wait for debounce delay
	time.Sleep(150 * time.Millisecond)
	mu.Lock()
	assert.True(t, called)
	mu.Unlock()
}

func TestDebouncer_MultipleCalls(t *testing.T) {
	debouncer := NewDebouncer(200 * time.Millisecond)

	callCount := 0
	var mu sync.Mutex

	fn := func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	// Make multiple rapid calls with same key
	debouncer.Debounce("test", fn)
	time.Sleep(50 * time.Millisecond)
	debouncer.Debounce("test", fn)
	time.Sleep(50 * time.Millisecond)
	debouncer.Debounce("test", fn)

	// Should not be called yet
	mu.Lock()
	assert.Equal(t, 0, callCount)
	mu.Unlock()

	// Wait for debounce delay
	time.Sleep(300 * time.Millisecond)

	// Should be called only once (last call)
	mu.Lock()
	assert.Equal(t, 1, callCount)
	mu.Unlock()
}

func TestDebouncer_DifferentKeys(t *testing.T) {
	debouncer := NewDebouncer(100 * time.Millisecond)

	call1Count := 0
	call2Count := 0
	var mu sync.Mutex

	fn1 := func() {
		mu.Lock()
		call1Count++
		mu.Unlock()
	}

	fn2 := func() {
		mu.Lock()
		call2Count++
		mu.Unlock()
	}

	// Make calls with different keys
	debouncer.Debounce("key1", fn1)
	debouncer.Debounce("key2", fn2)

	// Wait for debounce delay
	time.Sleep(150 * time.Millisecond)

	// Both should be called once
	mu.Lock()
	assert.Equal(t, 1, call1Count)
	assert.Equal(t, 1, call2Count)
	mu.Unlock()
}

func TestDebouncer_Cancellation(t *testing.T) {
	debouncer := NewDebouncer(200 * time.Millisecond)

	var firstCalled bool
	var secondCalled bool
	var mu sync.Mutex

	// First call
	debouncer.Debounce("test", func() {
		mu.Lock()
		firstCalled = true
		mu.Unlock()
	})

	// Wait a bit, then make second call (should cancel first)
	time.Sleep(100 * time.Millisecond)
	debouncer.Debounce("test", func() {
		mu.Lock()
		secondCalled = true
		mu.Unlock()
	})

	// Wait for debounce delay
	time.Sleep(250 * time.Millisecond)

	// Only second should be called
	mu.Lock()
	assert.False(t, firstCalled)
	assert.True(t, secondCalled)
	mu.Unlock()
}

func TestDebouncer_ConcurrentAccess(t *testing.T) {
	debouncer := NewDebouncer(50 * time.Millisecond)

	var callCount int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Concurrent calls with same key
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			debouncer.Debounce("concurrent", func() {
				mu.Lock()
				callCount++
				mu.Unlock()
			})
		}()
	}

	wg.Wait()

	// Wait for debounce delay
	time.Sleep(100 * time.Millisecond)

	// Should be called only once despite concurrent calls
	mu.Lock()
	assert.Equal(t, 1, callCount)
	mu.Unlock()
}

func TestDebouncer_ZeroDelay(t *testing.T) {
	debouncer := NewDebouncer(0)

	var called bool
	var mu sync.Mutex
	debouncer.Debounce("test", func() {
		mu.Lock()
		called = true
		mu.Unlock()
	})

	// With zero delay, should be called almost immediately
	time.Sleep(10 * time.Millisecond)
	mu.Lock()
	assert.True(t, called)
	mu.Unlock()
}

func TestDebouncer_MultipleKeysRapidFire(t *testing.T) {
	debouncer := NewDebouncer(100 * time.Millisecond)

	results := make(map[string]int)
	var mu sync.Mutex

	// Rapidly fire calls for multiple keys
	for i := 0; i < 5; i++ {
		for j := 0; j < 3; j++ {
			key := fmt.Sprintf("key%d", i)
			debouncer.Debounce(key, func() {
				mu.Lock()
				results[key]++
				mu.Unlock()
			})
		}
	}

	// Wait for all debounced calls
	time.Sleep(200 * time.Millisecond)

	// Each key should have been called exactly once
	mu.Lock()
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key%d", i)
		assert.Equal(t, 1, results[key], "Key %s should be called exactly once", key)
	}
	mu.Unlock()
}

func TestDebouncer_TimerCleanup(t *testing.T) {
	debouncer := NewDebouncer(50 * time.Millisecond)

	// Make calls that should complete
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		debouncer.Debounce(key, func() {
			// Do nothing
		})
	}

	// Wait for all timers to complete
	time.Sleep(100 * time.Millisecond)

	// Verify timers are cleaned up (indirectly by checking no panic occurs)
	// and that new calls still work
	var called bool
	var mu sync.Mutex
	debouncer.Debounce("new-key", func() {
		mu.Lock()
		called = true
		mu.Unlock()
	})

	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	assert.True(t, called)
	mu.Unlock()
}
