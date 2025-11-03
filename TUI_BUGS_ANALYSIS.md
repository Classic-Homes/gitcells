# TUI Dashboard Issues - Windows User Feedback

## Executive Summary

Three critical bugs were identified in the GitCells TUI dashboard based on Windows user testing. All three issues stem from incomplete/stub implementations in the dashboard models.

**Status**: All issues confirmed and root causes identified  
**Priority**: High - Core TUI functionality is broken  
**Affected Files**: 
- `internal/tui/models/dashboard.go`
- `internal/tui/models/unified_dashboard.go`
- `internal/tui/adapter/watcher.go`

---

## Issue #1: Inconsistent File/Directory Count Display

### User Report
> In the TUI dashboard, the overview specifies that it is watching "2 files" but then in the Activity page, it claims it is watching "2 directories". It is configured to watch only 1 directory via the YAML file.

### Root Cause
Confusing and inconsistent terminology throughout the dashboard displays.

### Technical Details

**File**: `internal/tui/models/dashboard.go`

**Problem Locations**:
1. **Line 239** - Header displays: `Tracking: %d files` using `m.totalFiles`
2. **Line 290** - Overview shows: `%d directories` using `len(m.watching)`
3. **Line 291** - Overview shows: `%d Excel files` using `m.totalFiles`

**Additional Issue**: `internal/tui/adapter/watcher.go:187`
```go
wa.filesWatched = len(fw.GetWatchedDirectories())
```
The variable `filesWatched` is misleadingly named - it actually contains a count of **directories**, not files.

### Why This Matters
- Users see "2 files" in the header but "2 directories" in content
- The variable naming (`filesWatched`) suggests files when it means directories
- Creates confusion about what the application is actually monitoring

### Proposed Fix

#### Option A: Show Both Metrics (Recommended)
Change `dashboard.go:239` to display both:
```go
status := statusStyle.Render(fmt.Sprintf("%s %s | Watching: %d dirs | Tracking: %d files", 
    statusIcon, statusText, len(m.watching), m.totalFiles))
```

#### Option B: Consistent Terminology
Update all references to use "directories" consistently:
```go
status := statusStyle.Render(fmt.Sprintf("%s %s | Watching: %d directories", 
    statusIcon, statusText, len(m.watching)))
```

#### Option C: Rename Variables for Clarity
In `internal/tui/adapter/watcher.go`:
- Line 26: Rename `filesWatched int` to `directoriesWatched int`
- Line 47: Rename `FilesWatched int` to `DirectoriesWatched int`
- Update all references throughout the file

**Recommendation**: Implement all three options for maximum clarity.

---

## Issue #2: Watch Shortcut Works Only Twice, Then All Shortcuts Break

### User Report
> The Watch shortcut appears to work only twice, one to enable watching then once to disable watching. After using that shortcut once, no other shortcuts work.
> Attempting to use any other shortcut does nothing, and causes none of them to work afterward, including Watch.

### Root Cause
The `toggleWatcher()` function in `DashboardModel` is a **non-functional stub** that creates fake operations instead of actually controlling the file watcher.

### Technical Details

**File**: `internal/tui/models/dashboard.go:528-542`

**Current Implementation**:
```go
func (m *DashboardModel) toggleWatcher() tea.Cmd {
    return func() tea.Msg {
        op := FileOperation{
            ID:        fmt.Sprintf("watch-%d", time.Now().UnixNano()),
            Type:      OpWatch,
            FileName:  "File watching",
            Status:    StatusInProgress,
            Progress:  0,
            StartTime: time.Now(),
        }

        // In a real implementation, this would toggle the file watcher
        return operationUpdateMsg{operation: op}
    }
}
```

Note the comment on **line 539**: *"In a real implementation, this would toggle the file watcher"*

### What's Happening

1. **First 'w' press**: Creates a fake "File watching" operation that never completes
2. **Second 'w' press**: Creates another fake operation
3. **Subsequent presses**: Operations accumulate, potentially causing:
   - Memory growth from unbounded operations list
   - State corruption in the dashboard model
   - Event processing queue overflow
   - UI becoming unresponsive to key events

### Why All Shortcuts Break

The accumulated fake operations likely:
- Clog the message queue in the Bubble Tea event loop
- Cause the dashboard model's state to become corrupted
- Prevent proper handling of subsequent `tea.KeyMsg` events
- May trigger infinite loops in the `dashboardTick()` function (line 125-140) trying to process operations that never complete

### Proposed Fix

The `DashboardModel` needs a `WatcherAdapter` instance and proper integration. Here's the implementation:

#### Step 1: Add WatcherAdapter to DashboardModel

**File**: `internal/tui/models/dashboard.go`

Add to the `DashboardModel` struct (around line 17):
```go
type DashboardModel struct {
    width       int
    height      int
    config      *config.Config
    gitAdapter  *adapter.GitAdapter
    convAdapter *adapter.ConverterAdapter
    watcherAdapter *adapter.WatcherAdapter  // ADD THIS
    logger      *logrus.Logger              // ADD THIS

    // ... rest of fields
}
```

#### Step 2: Initialize WatcherAdapter

Update `NewDashboardModel()` function (around line 84):
```go
func NewDashboardModel() tea.Model {
    m := &DashboardModel{
        operations:   []FileOperation{},
        progressBars: components.NewMultiProgress(),
        lastUpdate:   time.Now(),
        logger:       logrus.New(),
    }

    m.logger.SetLevel(logrus.WarnLevel) // Reduce noise in TUI

    // Try to load configuration
    if cfg, err := config.Load("."); err == nil {
        m.config = cfg
        m.watching = cfg.Watcher.Directories
        
        // Initialize watcher adapter
        if watcherAdapter, err := adapter.NewWatcherAdapter(cfg, m.logger, m.handleWatcherEvent); err == nil {
            m.watcherAdapter = watcherAdapter
        }
    }

    // Initialize other adapters...
    return m
}
```

#### Step 3: Implement Watcher Event Handler

Add this method to `DashboardModel`:
```go
func (m *DashboardModel) handleWatcherEvent(event adapter.WatcherEvent) {
    // Add to operations list
    op := FileOperation{
        ID:        fmt.Sprintf("watch-event-%d", time.Now().UnixNano()),
        Type:      OpWatch,
        FileName:  event.Message,
        Status:    StatusCompleted,
        Progress:  100,
        StartTime: event.Timestamp,
    }
    
    if event.Type == "error" {
        op.Status = StatusFailed
        op.Error = fmt.Errorf("%s", event.Details)
    }
    
    m.operations = append([]FileOperation{op}, m.operations...)
    
    // Keep only last 50 operations
    if len(m.operations) > 50 {
        m.operations = m.operations[:50]
    }
}
```

#### Step 4: Implement Real toggleWatcher()

Replace the stub function (line 528) with:
```go
func (m *DashboardModel) toggleWatcher() tea.Cmd {
    return func() tea.Msg {
        if m.watcherAdapter == nil {
            return operationUpdateMsg{
                operation: FileOperation{
                    ID:        fmt.Sprintf("watch-error-%d", time.Now().UnixNano()),
                    Type:      OpWatch,
                    FileName:  "Watcher not available",
                    Status:    StatusFailed,
                    StartTime: time.Now(),
                    Error:     fmt.Errorf("watcher adapter not initialized"),
                },
            }
        }

        var err error
        var isRunning bool
        
        if m.watcherAdapter.IsRunning() {
            // Stop the watcher
            err = m.watcherAdapter.Stop()
            isRunning = false
        } else {
            // Start the watcher
            err = m.watcherAdapter.Start()
            isRunning = true
        }

        status := StatusCompleted
        statusText := "started"
        if !isRunning {
            statusText = "stopped"
        }
        
        var opError error
        if err != nil {
            status = StatusFailed
            opError = err
        }

        return operationUpdateMsg{
            operation: FileOperation{
                ID:        fmt.Sprintf("watch-%d", time.Now().UnixNano()),
                Type:      OpWatch,
                FileName:  fmt.Sprintf("Watcher %s", statusText),
                Status:    status,
                Progress:  100,
                StartTime: time.Now(),
                Error:     opError,
            },
        }
    }
}
```

#### Step 5: Add Cleanup on Quit

Update the `Update()` method to clean up the watcher when quitting:
```go
case tea.KeyMsg:
    switch msg.String() {
    case "ctrl+c", "q":
        // Stop watcher if running
        if m.watcherAdapter != nil && m.watcherAdapter.IsRunning() {
            _ = m.watcherAdapter.Stop()
        }
        return m, tea.Quit
```

---

## Issue #3: Tracked Files Tab Not Detecting Files

### User Report
> Tracked Files tab does not appear to be detecting any files within the watched directories, but the Activity tab does show actions when I modify a file. I tried on several random files.

### Root Cause
The `loadDashboardData()` function in `UnifiedDashboardModel` is an empty stub that never populates the `m.files` array.

### Technical Details

**File**: `internal/tui/models/unified_dashboard.go:476-481`

**Current Implementation**:
```go
func (m *UnifiedDashboardModel) loadDashboardData() tea.Cmd {
    return func() tea.Msg {
        // Load data from adapters
        // This would be implemented to actually fetch data
        return nil
    }
}
```

**Important Note**: The user is likely using the regular `DashboardModel` (set at `app.go:120`), which only has three tabs: "Overview", "Operations", and "Commits". There is **no** "Tracked Files" tab in `DashboardModel`.

However, if the application is somehow using `UnifiedDashboardModel`, or if the user is referring to the "Overview" tab's file list, the `m.files` array is never populated because `loadDashboardData()` returns `nil`.

### Proposed Fix

#### Option A: Implement loadDashboardData() in UnifiedDashboardModel

**File**: `internal/tui/models/unified_dashboard.go`

Replace the stub with a full implementation:

```go
func (m *UnifiedDashboardModel) loadDashboardData() tea.Cmd {
    return func() tea.Msg {
        // Create result message
        result := dashboardDataMsg{
            files:      []FileInfo{},
            activities: []Activity{},
        }

        // Load files from watched directories
        if m.config != nil && len(m.config.Watcher.Directories) > 0 {
            for _, dir := range m.config.Watcher.Directories {
                for _, ext := range m.config.Watcher.FileExtensions {
                    pattern := filepath.Join(dir, "*"+ext)
                    files, err := filepath.Glob(pattern)
                    if err != nil {
                        continue
                    }

                    for _, filePath := range files {
                        // Get file info
                        info, err := os.Stat(filePath)
                        if err != nil {
                            continue
                        }

                        // Check if tracked in git
                        isTracked := false
                        hasChanges := false
                        if m.gitAdapter != nil {
                            isTracked = m.gitAdapter.IsTracked(filePath)
                            if isTracked {
                                status, _ := m.gitAdapter.GetFileStatus(filePath)
                                hasChanges = status != "clean"
                            }
                        }

                        // Get JSON path
                        jsonPath := filepath.Join(".gitcells", "data", 
                            strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)) + ".json")

                        result.files = append(result.files, FileInfo{
                            Path:         filePath,
                            Size:         info.Size(),
                            LastModified: info.ModTime(),
                            IsTracked:    isTracked,
                            HasChanges:   hasChanges,
                            JSONPath:     jsonPath,
                        })
                    }
                }
            }
        }

        // Load git status
        if m.gitAdapter != nil {
            branch := "main"
            if currentBranch, err := m.gitAdapter.GetCurrentBranch(); err == nil {
                branch = currentBranch
            }

            isClean := true
            if clean, err := m.gitAdapter.IsClean(); err == nil {
                isClean = clean
            }

            result.syncStatus = SyncStatus{
                Branch:     branch,
                IsSynced:   isClean,
                HasChanges: !isClean,
            }
        }

        // Load watcher status
        if m.watcherAdapter != nil {
            status := m.watcherAdapter.GetStatus()
            result.watcherState = WatcherState{
                IsRunning:     status.IsRunning,
                StartTime:     status.StartTime,
                FilesWatched:  status.FilesWatched,
                LastEvent:     status.LastEvent,
                LastEventTime: status.LastEventTime,
            }
        }

        return result
    }
}

// Add message type
type dashboardDataMsg struct {
    files        []FileInfo
    activities   []Activity
    syncStatus   SyncStatus
    watcherState WatcherState
}
```

#### Step 2: Handle the Message in Update()

Add to the `Update()` method's switch statement:

```go
case dashboardDataMsg:
    m.files = msg.files
    if len(msg.activities) > 0 {
        m.activities = msg.activities
    }
    m.syncStatus = msg.syncStatus
    m.watcherState = msg.watcherState
    return m, nil
```

#### Option B: Switch to UnifiedDashboardModel (If Not Already Used)

**File**: `internal/tui/app.go:118-122`

Change from:
```go
case ModeDashboard:
    if m.dashModel == nil {
        m.dashModel = models.NewDashboardModel()
    }
    return m, m.dashModel.Init()
```

To:
```go
case ModeDashboard:
    if m.dashModel == nil {
        m.dashModel = models.NewUnifiedDashboardModel()
    }
    return m, m.dashModel.Init()
```

**Note**: This requires that `NewUnifiedDashboardModel()` returns `tea.Model` interface type.

#### Option C: Add "Files" Tab to Existing DashboardModel

This is more complex and requires:
1. Adding a 4th tab to the dashboard
2. Implementing file scanning logic similar to Option A
3. Adding rendering logic for the files table

---

## Additional Recommendations

### 1. Add Operation Cleanup
The `operations` array in `DashboardModel` can grow unbounded. Add cleanup logic:

```go
func (m *DashboardModel) updateOperation(msg operationUpdateMsg) {
    // Add or update operation
    found := false
    for i, op := range m.operations {
        if op.ID == msg.operation.ID {
            m.operations[i] = msg.operation
            found = true
            break
        }
    }

    if !found {
        m.operations = append(m.operations, msg.operation)
        if msg.operation.Status == StatusInProgress {
            m.progressBars.AddBar(msg.operation.ID, msg.operation.FileName, 100)
        }
    }
    
    // Keep only last 100 operations
    if len(m.operations) > 100 {
        m.operations = m.operations[len(m.operations)-100:]
    }
}
```

### 2. Add Better Error Handling
Both dashboard models need better error messaging when:
- Configuration is missing
- No directories are configured
- Git repository is not initialized
- Watcher fails to start

### 3. Improve Variable Naming Throughout
Standardize terminology:
- `filesWatched` → `directoriesWatched`
- `watching []string` → `watchedDirectories []string`
- Be consistent with "files" vs "directories" everywhere

### 4. Add Unit Tests
None of the TUI models have adequate test coverage. Add tests for:
- `toggleWatcher()` functionality
- Operation list management
- File scanning and population
- Edge cases (empty config, no files, etc.)

---

## Testing Plan

### Manual Testing Steps

1. **Test Watch Toggle**:
   - Start TUI dashboard
   - Press 'w' to start watching
   - Verify watcher actually starts
   - Modify an Excel file
   - Verify event appears in operations
   - Press 'w' to stop watching
   - Verify watcher actually stops
   - Try other shortcuts (c, r, tab, etc.)
   - Verify all shortcuts still work

2. **Test File Detection**:
   - Configure watch directory with Excel files
   - Start TUI dashboard
   - Navigate to Files tab (if using UnifiedDashboardModel)
   - Verify all Excel files are listed
   - Verify correct file counts
   - Verify git status shows correctly

3. **Test Terminology Consistency**:
   - Check header shows correct counts
   - Check Overview tab shows correct counts
   - Verify "directories" vs "files" terminology is clear
   - Test with 0, 1, and multiple directories

### Automated Testing

Add tests to `internal/tui/models/dashboard_test.go`:

```go
func TestToggleWatcher(t *testing.T) {
    // Test that watcher actually starts/stops
    // Test that operations are created correctly
    // Test that shortcuts continue working after toggle
}

func TestFileScanning(t *testing.T) {
    // Test that files are detected in watched directories
    // Test that git status is correctly determined
    // Test that file info is accurate
}

func TestOperationCleanup(t *testing.T) {
    // Test that operations list doesn't grow unbounded
    // Test that old operations are removed
}
```

---

## Priority & Timeline

### Critical (Fix Immediately)
- **Issue #2**: Watch shortcut breaking all shortcuts
  - This makes the TUI nearly unusable
  - Estimated effort: 4-6 hours
  
### High (Fix Soon)
- **Issue #3**: Files not appearing in Files tab
  - Users can't see what's being tracked
  - Estimated effort: 2-4 hours
  
### Medium (Fix Next Release)
- **Issue #1**: Terminology consistency
  - Causes confusion but doesn't break functionality
  - Estimated effort: 1-2 hours

### Total Estimated Effort
- **10-12 hours** for complete fix of all three issues
- Additional 4-6 hours for comprehensive testing

---

## Files to Modify

| File | Changes Required | Estimated Lines |
|------|-----------------|-----------------|
| `internal/tui/models/dashboard.go` | Add WatcherAdapter integration, fix toggleWatcher() | ~100 lines |
| `internal/tui/models/unified_dashboard.go` | Implement loadDashboardData() | ~80 lines |
| `internal/tui/adapter/watcher.go` | Rename variables for clarity | ~10 lines |
| `internal/tui/models/dashboard_test.go` | Add test coverage | ~150 lines |

---

## Conclusion

All three reported issues are confirmed bugs caused by incomplete stub implementations in the TUI dashboard. The fixes are straightforward but require careful integration of the `WatcherAdapter` into the dashboard models and proper implementation of data loading logic.

The most critical issue is #2 (watch shortcut breaking), as it makes the TUI unusable after a few key presses. This should be addressed immediately before any Windows release.
