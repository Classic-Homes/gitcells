package common

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewModelState(t *testing.T) {
	state := NewModelState()

	assert.NotNil(t, state)
	assert.Equal(t, 80, state.Width)
	assert.Equal(t, 24, state.Height)
	assert.Equal(t, 0, state.Cursor)
	assert.Empty(t, state.Status)
	assert.False(t, state.Loading)
	assert.Nil(t, state.Error)
}

func TestModelState_SetSize(t *testing.T) {
	state := NewModelState()

	state.SetSize(100, 50)

	width, height := state.GetSize()
	assert.Equal(t, 100, width)
	assert.Equal(t, 50, height)
}

func TestModelState_SetCursor(t *testing.T) {
	state := NewModelState()

	state.SetCursor(5)
	assert.Equal(t, 5, state.GetCursor())

	state.SetCursor(10)
	assert.Equal(t, 10, state.GetCursor())
}

func TestModelState_MoveCursor(t *testing.T) {
	state := NewModelState()
	state.SetCursor(5)

	// Move up
	state.MoveCursor(-2, 10)
	assert.Equal(t, 3, state.GetCursor())

	// Move down
	state.MoveCursor(4, 10)
	assert.Equal(t, 7, state.GetCursor())

	// Try to move below 0
	state.MoveCursor(-10, 10)
	assert.Equal(t, 0, state.GetCursor())

	// Try to move above max
	state.MoveCursor(20, 10)
	assert.Equal(t, 9, state.GetCursor())
}

func TestModelState_Status(t *testing.T) {
	state := NewModelState()

	state.SetStatus("Loading...")
	assert.Equal(t, "Loading...", state.GetStatus())

	state.SetStatus("Complete")
	assert.Equal(t, "Complete", state.GetStatus())
}

func TestModelState_Loading(t *testing.T) {
	state := NewModelState()

	assert.False(t, state.IsLoading())

	state.SetLoading(true)
	assert.True(t, state.IsLoading())

	state.SetLoading(false)
	assert.False(t, state.IsLoading())
}

func TestModelState_Error(t *testing.T) {
	state := NewModelState()

	assert.Nil(t, state.GetError())

	err := errors.New("test error")
	state.SetError(err)
	assert.Equal(t, err, state.GetError())
	assert.Equal(t, "test error", state.GetStatus())

	state.ClearError()
	assert.Nil(t, state.GetError())
}

func TestModelState_Reset(t *testing.T) {
	state := NewModelState()

	// Set various state values
	state.SetCursor(10)
	state.SetStatus("Some status")
	state.SetLoading(true)
	state.SetError(errors.New("some error"))

	// Reset
	state.Reset()

	// Verify everything is reset
	assert.Equal(t, 0, state.GetCursor())
	assert.Empty(t, state.GetStatus())
	assert.False(t, state.IsLoading())
	assert.Nil(t, state.GetError())
}

func TestModelState_Concurrency(t *testing.T) {
	state := NewModelState()
	var wg sync.WaitGroup

	// Test concurrent reads and writes
	iterations := 100
	wg.Add(iterations * 4)

	// Concurrent cursor updates
	for i := 0; i < iterations; i++ {
		go func(val int) {
			defer wg.Done()
			state.SetCursor(val)
		}(i)
	}

	// Concurrent status updates
	for i := 0; i < iterations; i++ {
		go func(val int) {
			defer wg.Done()
			state.SetStatus("status")
			_ = state.GetStatus()
		}(i)
	}

	// Concurrent loading updates
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			state.SetLoading(true)
			_ = state.IsLoading()
		}()
	}

	// Concurrent error updates
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			state.SetError(errors.New("test"))
			_ = state.GetError()
		}()
	}

	wg.Wait()

	// Just verify no panics occurred
	assert.NotNil(t, state)
}
