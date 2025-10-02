package common

import (
	"sync"
)

// ModelState provides common state management for TUI models
type ModelState struct {
	Width   int
	Height  int
	Cursor  int
	Status  string
	Loading bool
	Error   error
	mu      sync.RWMutex
}

// NewModelState creates a new model state
func NewModelState() *ModelState {
	return &ModelState{
		Width:  80,
		Height: 24,
	}
}

// SetSize updates the model dimensions
func (s *ModelState) SetSize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Width = width
	s.Height = height
}

// GetSize returns the current model dimensions
func (s *ModelState) GetSize() (int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Width, s.Height
}

// SetCursor updates the cursor position
func (s *ModelState) SetCursor(cursor int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Cursor = cursor
}

// GetCursor returns the current cursor position
func (s *ModelState) GetCursor() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Cursor
}

// MoveCursor moves the cursor with bounds checking
func (s *ModelState) MoveCursor(delta int, max int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newCursor := s.Cursor + delta
	if newCursor < 0 {
		newCursor = 0
	}
	if newCursor >= max {
		newCursor = max - 1
	}
	s.Cursor = newCursor
}

// SetStatus updates the status message
func (s *ModelState) SetStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
}

// GetStatus returns the current status message
func (s *ModelState) GetStatus() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status
}

// SetLoading updates the loading state
func (s *ModelState) SetLoading(loading bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Loading = loading
}

// IsLoading returns whether the model is in loading state
func (s *ModelState) IsLoading() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Loading
}

// SetError sets the error state
func (s *ModelState) SetError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Error = err
	if err != nil {
		s.Status = err.Error()
	}
}

// GetError returns the current error
func (s *ModelState) GetError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Error
}

// ClearError clears the error state
func (s *ModelState) ClearError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Error = nil
}

// Reset resets the state to defaults
func (s *ModelState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Cursor = 0
	s.Status = ""
	s.Loading = false
	s.Error = nil
}
