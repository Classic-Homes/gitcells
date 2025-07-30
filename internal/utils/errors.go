package utils

import (
	"fmt"
	"strings"
)

// SheetSyncError represents a custom error type for SheetSync operations
type SheetSyncError struct {
	Type        ErrorType
	Operation   string
	File        string
	Cause       error
	Message     string
	Recoverable bool
}

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeConverter    ErrorType = "CONVERTER"
	ErrorTypeGit          ErrorType = "GIT"
	ErrorTypeWatcher      ErrorType = "WATCHER"
	ErrorTypeConfig       ErrorType = "CONFIG"
	ErrorTypeValidation   ErrorType = "VALIDATION"
	ErrorTypeFileSystem   ErrorType = "FILESYSTEM"
	ErrorTypeNetwork      ErrorType = "NETWORK"
	ErrorTypeConflict     ErrorType = "CONFLICT"
	ErrorTypePermission   ErrorType = "PERMISSION"
	ErrorTypeCorruption   ErrorType = "CORRUPTION"
)

// Error implements the error interface
func (e *SheetSyncError) Error() string {
	var parts []string
	
	if e.Type != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Type))
	}
	
	if e.Operation != "" {
		parts = append(parts, fmt.Sprintf("operation=%s", e.Operation))
	}
	
	if e.File != "" {
		parts = append(parts, fmt.Sprintf("file=%s", e.File))
	}
	
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	
	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("cause: %v", e.Cause))
	}
	
	return strings.Join(parts, " ")
}

// Unwrap returns the underlying error
func (e *SheetSyncError) Unwrap() error {
	return e.Cause
}

// IsRecoverable returns whether the error is recoverable
func (e *SheetSyncError) IsRecoverable() bool {
	return e.Recoverable
}

// NewError creates a new SheetSyncError
func NewError(errorType ErrorType, operation string, message string) *SheetSyncError {
	return &SheetSyncError{
		Type:        errorType,
		Operation:   operation,
		Message:     message,
		Recoverable: false,
	}
}

// WrapError wraps an existing error with SheetSync context
func WrapError(err error, errorType ErrorType, operation string, message string) *SheetSyncError {
	if err == nil {
		return nil
	}
	
	return &SheetSyncError{
		Type:      errorType,
		Operation: operation,
		Message:   message,
		Cause:     err,
		Recoverable: isRecoverableError(err),
	}
}

// WrapFileError wraps an error with file context
func WrapFileError(err error, errorType ErrorType, operation string, file string, message string) *SheetSyncError {
	if err == nil {
		return nil
	}
	
	return &SheetSyncError{
		Type:        errorType,
		Operation:   operation,
		File:        file,
		Message:     message,
		Cause:       err,
		Recoverable: isRecoverableError(err),
	}
}

// isRecoverableError determines if an error is potentially recoverable
func isRecoverableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	
	// Network-related errors are often recoverable
	if strings.Contains(errStr, "network") || strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") || strings.Contains(errStr, "dns") {
		return true
	}
	
	// File system temporary issues
	if strings.Contains(errStr, "temporary") || strings.Contains(errStr, "busy") ||
		strings.Contains(errStr, "locked") {
		return true
	}
	
	// Git-related recoverable errors
	if strings.Contains(errStr, "merge conflict") || strings.Contains(errStr, "diverged") {
		return true
	}
	
	return false
}

// ErrorCollector collects multiple errors and provides a summary
type ErrorCollector struct {
	errors []error
	limit  int
}

// NewErrorCollector creates a new error collector with optional limit
func NewErrorCollector(limit int) *ErrorCollector {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	return &ErrorCollector{
		errors: make([]error, 0),
		limit:  limit,
	}
}

// Add adds an error to the collection
func (ec *ErrorCollector) Add(err error) {
	if err != nil && len(ec.errors) < ec.limit {
		ec.errors = append(ec.errors, err)
	}
}

// AddContext adds an error with additional context
func (ec *ErrorCollector) AddContext(err error, context string) {
	if err != nil {
		contextErr := fmt.Errorf("%s: %w", context, err)
		ec.Add(contextErr)
	}
}

// HasErrors returns true if any errors were collected
func (ec *ErrorCollector) HasErrors() bool {
	return len(ec.errors) > 0
}

// Count returns the number of collected errors
func (ec *ErrorCollector) Count() int {
	return len(ec.errors)
}

// Errors returns all collected errors
func (ec *ErrorCollector) Errors() []error {
	return ec.errors
}

// First returns the first error, or nil if none
func (ec *ErrorCollector) First() error {
	if len(ec.errors) == 0 {
		return nil
	}
	return ec.errors[0]
}

// Error implements the error interface, providing a summary of all errors
func (ec *ErrorCollector) Error() string {
	if len(ec.errors) == 0 {
		return "no errors"
	}
	
	if len(ec.errors) == 1 {
		return ec.errors[0].Error()
	}
	
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("%d errors occurred:\n", len(ec.errors)))
	
	for i, err := range ec.errors {
		if i >= 5 { // Limit detailed output
			summary.WriteString(fmt.Sprintf("... and %d more errors\n", len(ec.errors)-i))
			break
		}
		summary.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
	}
	
	return summary.String()
}

// RetryableOperation represents an operation that can be retried
type RetryableOperation func() error

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts int
	ShouldRetry func(error) bool
	OnRetry     func(error, int)
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts: 3,
		ShouldRetry: func(err error) bool {
			if ssErr, ok := err.(*SheetSyncError); ok {
				return ssErr.IsRecoverable()
			}
			return isRecoverableError(err)
		},
		OnRetry: func(err error, attempt int) {
			// Default: do nothing
		},
	}
}

// Retry executes an operation with retry logic
func Retry(operation RetryableOperation, config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
	}
	
	var lastErr error
	
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil // Success
		}
		
		lastErr = err
		
		// Don't retry on the last attempt or if error is not retryable
		if attempt >= config.MaxAttempts || !config.ShouldRetry(err) {
			break
		}
		
		// Call retry callback
		if config.OnRetry != nil {
			config.OnRetry(err, attempt)
		}
	}
	
	// Wrap the final error with retry context
	return WrapError(lastErr, ErrorTypeConverter, "retry", 
		fmt.Sprintf("operation failed after %d attempts", config.MaxAttempts))
}

// ValidationError represents a validation error with field context
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

// Error implements the error interface
func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s' (value: %v): %s", v.Field, v.Value, v.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}
	
	if len(ve) == 1 {
		return ve[0].Error()
	}
	
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("%d validation errors:\n", len(ve)))
	
	for _, err := range ve {
		summary.WriteString(fmt.Sprintf("  - %s\n", err.Error()))
	}
	
	return summary.String()
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field string, value interface{}, message string) {
	*ve = append(*ve, ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	})
}

// HasErrors returns true if there are validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}