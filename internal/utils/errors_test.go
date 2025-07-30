package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSheetSyncError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *SheetSyncError
		expected string
	}{
		{
			name: "minimal error",
			err: &SheetSyncError{
				Message: "test error",
			},
			expected: "test error",
		},
		{
			name: "error with type and operation",
			err: &SheetSyncError{
				Type:      ErrorTypeConverter,
				Operation: "excel_to_json",
				Message:   "conversion failed",
			},
			expected: "[CONVERTER] operation=excel_to_json conversion failed",
		},
		{
			name: "error with file",
			err: &SheetSyncError{
				Type:      ErrorTypeFileSystem,
				Operation: "read_file",
				File:      "test.xlsx",
				Message:   "file not found",
			},
			expected: "[FILESYSTEM] operation=read_file file=test.xlsx file not found",
		},
		{
			name: "error with cause",
			err: &SheetSyncError{
				Type:      ErrorTypeGit,
				Operation: "commit",
				Message:   "commit failed",
				Cause:     errors.New("permission denied"),
			},
			expected: "[GIT] operation=commit commit failed cause: permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSheetSyncError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	err := &SheetSyncError{
		Message: "wrapped",
		Cause:   cause,
	}

	assert.Equal(t, cause, err.Unwrap())
}

func TestSheetSyncError_IsRecoverable(t *testing.T) {
	tests := []struct {
		name        string
		recoverable bool
		expected    bool
	}{
		{
			name:        "recoverable error",
			recoverable: true,
			expected:    true,
		},
		{
			name:        "non-recoverable error",
			recoverable: false,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &SheetSyncError{
				Recoverable: tt.recoverable,
			}
			assert.Equal(t, tt.expected, err.IsRecoverable())
		})
	}
}

func TestNewError(t *testing.T) {
	err := NewError(ErrorTypeConverter, "test_operation", "test message")

	assert.Equal(t, ErrorTypeConverter, err.Type)
	assert.Equal(t, "test_operation", err.Operation)
	assert.Equal(t, "test message", err.Message)
	assert.False(t, err.Recoverable)
	assert.Nil(t, err.Cause)
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original")
	wrapped := WrapError(originalErr, ErrorTypeGit, "push", "failed to push")

	assert.Equal(t, ErrorTypeGit, wrapped.Type)
	assert.Equal(t, "push", wrapped.Operation)
	assert.Equal(t, "failed to push", wrapped.Message)
	assert.Equal(t, originalErr, wrapped.Cause)
}

func TestWrapError_NilError(t *testing.T) {
	wrapped := WrapError(nil, ErrorTypeGit, "push", "failed to push")
	assert.Nil(t, wrapped)
}

func TestWrapFileError(t *testing.T) {
	originalErr := errors.New("file error")
	wrapped := WrapFileError(originalErr, ErrorTypeFileSystem, "read", "test.xlsx", "failed to read")

	assert.Equal(t, ErrorTypeFileSystem, wrapped.Type)
	assert.Equal(t, "read", wrapped.Operation)
	assert.Equal(t, "test.xlsx", wrapped.File)
	assert.Equal(t, "failed to read", wrapped.Message)
	assert.Equal(t, originalErr, wrapped.Cause)
}

func TestIsRecoverableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "network error",
			err:      errors.New("network timeout"),
			expected: true,
		},
		{
			name:     "connection error",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "dns error",
			err:      errors.New("dns lookup failed"),
			expected: true,
		},
		{
			name:     "temporary error",
			err:      errors.New("temporary failure"),
			expected: true,
		},
		{
			name:     "busy error",
			err:      errors.New("resource busy"),
			expected: true,
		},
		{
			name:     "locked error",
			err:      errors.New("file locked"),
			expected: true,
		},
		{
			name:     "merge conflict",
			err:      errors.New("merge conflict detected"),
			expected: true,
		},
		{
			name:     "diverged branches",
			err:      errors.New("branches have diverged"),
			expected: true,
		},
		{
			name:     "non-recoverable error",
			err:      errors.New("syntax error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRecoverableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorCollector(t *testing.T) {
	collector := NewErrorCollector(5)

	assert.False(t, collector.HasErrors())
	assert.Equal(t, 0, collector.Count())
	assert.Nil(t, collector.First())

	// Add some errors
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	collector.Add(err1)
	collector.Add(err2)
	collector.AddContext(err3, "context")

	assert.True(t, collector.HasErrors())
	assert.Equal(t, 3, collector.Count())
	assert.Equal(t, err1, collector.First())

	errors := collector.Errors()
	assert.Len(t, errors, 3)
	assert.Equal(t, err1, errors[0])
	assert.Equal(t, err2, errors[1])
	assert.Contains(t, errors[2].Error(), "context: error 3")
}

func TestErrorCollector_Limit(t *testing.T) {
	collector := NewErrorCollector(2)

	collector.Add(errors.New("error 1"))
	collector.Add(errors.New("error 2"))
	collector.Add(errors.New("error 3")) // Should be ignored due to limit

	assert.Equal(t, 2, collector.Count())
}

func TestErrorCollector_Error(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		expected string
	}{
		{
			name:     "no errors",
			errors:   []error{},
			expected: "no errors",
		},
		{
			name:     "single error",
			errors:   []error{errors.New("single error")},
			expected: "single error",
		},
		{
			name: "multiple errors",
			errors: []error{
				errors.New("error 1"),
				errors.New("error 2"),
				errors.New("error 3"),
			},
			expected: "3 errors occurred:\n  1. error 1\n  2. error 2\n  3. error 3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewErrorCollector(10)
			for _, err := range tt.errors {
				collector.Add(err)
			}

			result := collector.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetry_Success(t *testing.T) {
	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary failure")
		}
		return nil
	}

	config := DefaultRetryConfig()
	err := Retry(operation, config)

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestRetry_AllAttemptsFail(t *testing.T) {
	attempts := 0
	operation := func() error {
		attempts++
		return errors.New("persistent failure")
	}

	config := &RetryConfig{
		MaxAttempts: 3,
		ShouldRetry: func(error) bool { return true },
	}

	err := Retry(operation, config)

	assert.Error(t, err)
	assert.Equal(t, 3, attempts)
	
	// Should be wrapped with retry context
	ssErr, ok := err.(*SheetSyncError)
	require.True(t, ok)
	assert.Equal(t, ErrorTypeConverter, ssErr.Type)
	assert.Equal(t, "retry", ssErr.Operation)
}

func TestRetry_NonRetryableError(t *testing.T) {
	attempts := 0
	operation := func() error {
		attempts++
		return errors.New("syntax error") // Non-retryable
	}

	config := DefaultRetryConfig()
	err := Retry(operation, config)

	assert.Error(t, err)
	assert.Equal(t, 1, attempts) // Should not retry
}

func TestRetry_OnRetryCallback(t *testing.T) {
	var retryCallbacks []int
	operation := func() error {
		return errors.New("network timeout") // Retryable
	}

	config := &RetryConfig{
		MaxAttempts: 3,
		ShouldRetry: func(error) bool { return true },
		OnRetry: func(err error, attempt int) {
			retryCallbacks = append(retryCallbacks, attempt)
		},
	}

	Retry(operation, config)

	assert.Equal(t, []int{1, 2}, retryCallbacks)
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "name",
		Value:   "",
		Message: "cannot be empty",
	}

	expected := "validation failed for field 'name' (value: ): cannot be empty"
	assert.Equal(t, expected, err.Error())
}

func TestValidationErrors(t *testing.T) {
	var errors ValidationErrors

	assert.False(t, errors.HasErrors())
	assert.Equal(t, "no validation errors", errors.Error())

	errors.Add("name", "", "cannot be empty")
	errors.Add("age", -1, "must be positive")

	assert.True(t, errors.HasErrors())
	
	errorStr := errors.Error()
	assert.Contains(t, errorStr, "2 validation errors:")
	assert.Contains(t, errorStr, "validation failed for field 'name'")
	assert.Contains(t, errorStr, "validation failed for field 'age'")
}

func TestValidationErrors_SingleError(t *testing.T) {
	var errors ValidationErrors
	errors.Add("test", "value", "invalid")

	result := errors.Error()
	expected := "validation failed for field 'test' (value: value): invalid"
	assert.Equal(t, expected, result)
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxAttempts)
	assert.NotNil(t, config.ShouldRetry)
	assert.NotNil(t, config.OnRetry)

	// Test ShouldRetry function
	assert.True(t, config.ShouldRetry(errors.New("network timeout")))
	assert.False(t, config.ShouldRetry(errors.New("syntax error")))

	// Test with SheetSyncError
	recoverableErr := &SheetSyncError{Recoverable: true}
	nonRecoverableErr := &SheetSyncError{Recoverable: false}
	
	assert.True(t, config.ShouldRetry(recoverableErr))
	assert.False(t, config.ShouldRetry(nonRecoverableErr))
}

// Benchmark tests
func BenchmarkErrorCollector_Add(b *testing.B) {
	collector := NewErrorCollector(1000)
	err := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.Add(err)
	}
}

func BenchmarkSheetSyncError_Error(b *testing.B) {
	err := &SheetSyncError{
		Type:      ErrorTypeConverter,
		Operation: "test_operation",
		File:      "test_file.xlsx",
		Message:   "test error message",
		Cause:     errors.New("underlying cause"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}