# TUI Common Package

This package provides shared utilities for Terminal User Interface (TUI) models in GitCells. It was created to reduce code duplication across large TUI model files and provide reusable components.

## Purpose

The `common` package extracts frequently used patterns from TUI models into reusable utilities:

- **Configuration management** - Centralized config getters/setters
- **Rendering helpers** - Common UI rendering functions
- **State management** - Thread-safe model state handling

## Modules

### `config_helpers.go`

Provides centralized configuration value management with type-safe getters and setters.

**Key Functions:**
- `SetStringValue(cfg, key, value)` - Set string config values
- `SetBoolValue(cfg, key, value)` - Set boolean config values
- `SetDurationValue(cfg, key, value)` - Set duration config values
- `SetIntValue(cfg, key, value)` - Set integer config values
- `SetStringSliceValue(cfg, key, value)` - Set string slice config values
- `GetStringValue(cfg, key)` - Get string config values
- `GetBoolValue(cfg, key)` - Get boolean config values
- `ToggleBoolValue(cfg, key)` - Toggle boolean config values

**Example:**
```go
cfg := &config.Config{}

// Set values
err := SetBoolValue(cfg, "git.auto_push", true)
err = SetDurationValue(cfg, "watcher.debounce_delay", "2s")

// Get values
enabled, err := GetBoolValue(cfg, "git.auto_push")

// Toggle values
err = ToggleBoolValue(cfg, "converter.preserve_formulas")
```

### `rendering.go`

Provides common rendering functions for TUI components.

**Key Functions:**
- `RenderMenuItems(items, cursor, width, cursorStyle, descStyle)` - Render menu with cursor
- `RenderBooleanValue(value, enabledStyle, disabledStyle)` - Render boolean as "Enabled"/"Disabled"
- `RenderSettingItem(label, value, selected, ...)` - Render a settings row
- `RenderConfirmDialog(title, message, width, height, style)` - Render confirmation dialog
- `RenderStatus(status, isError, successStyle, errorStyle)` - Render status message
- `WrapText(text, width)` - Wrap text to fit width
- `TruncateString(s, maxLen)` - Truncate with ellipsis
- `RenderList(items, numbered, style)` - Render bulleted or numbered list

**Example:**
```go
items := []MenuItem{
    {Title: "Settings", Description: "Configure options", Key: "settings"},
    {Title: "Exit", Description: "Quit application", Key: "exit"},
}

content := RenderMenuItems(items, cursor, 80, cursorStyle, descStyle)
```

### `state.go`

Provides thread-safe state management for TUI models.

**Key Features:**
- Thread-safe access with RWMutex
- Common model state (size, cursor, status, loading, error)
- Cursor movement with bounds checking
- State reset functionality

**Example:**
```go
state := NewModelState()

// Set size
state.SetSize(100, 50)

// Move cursor safely
state.MoveCursor(1, maxItems)  // Move down with bounds

// Manage loading state
state.SetLoading(true)
defer state.SetLoading(false)

// Handle errors
if err != nil {
    state.SetError(err)
}
```

## Benefits

1. **Code Reuse** - Eliminates duplication across 900+ line TUI models
2. **Consistency** - Standardized rendering and config handling
3. **Maintainability** - Single location for common logic
4. **Testability** - Well-tested utilities with 100% coverage
5. **Type Safety** - Compile-time checking for config keys

## Migration Guide

When refactoring existing TUI models to use this package:

### Before:
```go
func (m Model) getCurrentValue(key string) string {
    switch key {
    case "git.branch":
        return m.config.Git.Branch
    case "git.remote":
        return m.config.Git.Remote
    // ... 50+ more cases
    }
}
```

### After:
```go
import "github.com/Classic-Homes/gitcells/internal/tui/common"

value, err := common.GetStringValue(m.config, "git.branch")
```

## Testing

Run tests with:
```bash
go test ./internal/tui/common/...
```

Coverage is maintained at 100% for all exported functions.

## Related Issues

This package was created as part of refactoring efforts discussed in the codebase review. See the commit history for details on the specific files that were refactored.
