# GitCells Bubble Tea Migration Plan

This document outlines the steps required to convert GitCells from a Cobra CLI application to support a Bubble Tea TUI (Terminal User Interface) for enhanced user interaction during setup, branching, and conflict resolution operations.

## Overview

GitCells currently uses Cobra for CLI commands. While this works well for simple operations, complex workflows like conflict resolution and branch management would benefit from an interactive TUI. Bubble Tea provides a model-view-update architecture perfect for building rich terminal applications.

## Phase 1: Foundation Setup

### 1.1 Add Dependencies

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/lipgloss
```

### 1.2 Create TUI Package Structure

```
internal/tui/
├── app.go              # Main TUI application and router
├── models/
│   ├── setup.go        # Setup wizard model
│   ├── branch.go       # Branch management model
│   ├── conflict.go     # Conflict resolution model
│   └── dashboard.go    # Status dashboard model
├── views/
│   ├── setup.go        # Setup wizard views
│   ├── branch.go       # Branch management views
│   ├── conflict.go     # Conflict resolution views
│   └── dashboard.go    # Status dashboard views
├── components/
│   ├── filepicker.go   # File browser component
│   ├── table.go        # Data table component
│   ├── progress.go     # Progress bar component
│   └── diff.go         # Excel diff viewer component
└── styles/
    └── theme.go        # Consistent styling across TUI
```

## Phase 2: Core TUI Features

### 2.1 Setup Wizard (Replaces `gitcells init`)

**Features:**
- Interactive directory selection with file browser
- Excel file pattern configuration with live preview
- Git settings configuration:
  - Auto-commit toggle
  - Auto-push toggle
  - Commit message templates
- Watch directory setup with pattern testing
- Configuration preview before saving
- Validation with helpful error messages

**User Flow:**
1. Welcome screen with GitCells overview
2. Directory selection (current dir or browse)
3. Excel pattern configuration (*.xlsx, specific folders)
4. Git integration settings
5. Review and confirm configuration
6. Initialize with progress indicator

### 2.2 Branch Management Interface

**Features:**
- List all branches with Excel file status
- Visual indicators for:
  - Current branch
  - Branches with uncommitted changes
  - Branches with conflicts
- Create new branch with guided workflow
- Switch branches with safety checks
- Merge branches with conflict preview
- Delete branches with confirmation

**User Flow:**
1. Main branch list view
2. Keyboard navigation (j/k or arrows)
3. Actions menu (n=new, s=switch, m=merge, d=delete)
4. Context-sensitive help at bottom

### 2.3 Conflict Resolution Interface

**Features:**
- Side-by-side diff view of conflicting cells
- Cell-by-cell navigation
- Multiple resolution strategies:
  - Accept theirs
  - Accept ours
  - Manual edit
  - Skip cell
- Formula conflict special handling
- Batch operations (accept all from source)
- Preview resolved file before committing

**User Flow:**
1. Conflict summary screen
2. Navigate through conflicts
3. Resolve each conflict
4. Review all resolutions
5. Apply and commit

### 2.4 Status Dashboard

**Features:**
- Real-time file monitoring display
- Conversion queue and progress
- Git sync status indicators
- Recent activity log
- Quick actions menu
- System resource usage

**Layout:**
```
┌─────────────────────────────────────────────┐
│ GitCells Status Dashboard                   │
├─────────────────────────────────────────────┤
│ Watching: 3 directories, 15 Excel files     │
│ Status: ✓ Synced | Last commit: 2 min ago  │
├─────────────────────────────────────────────┤
│ File Operations:                            │
│ ► Converting: Budget2024.xlsx (45%)         │
│ ✓ Completed: Report.xlsx → Report.json      │
│ ⚠ Skipped: ~$TempFile.xlsx (temp file)     │
├─────────────────────────────────────────────┤
│ [w]atch [c]onvert [s]ync [q]uit [?]help    │
└─────────────────────────────────────────────┘
```

## Phase 3: Hybrid CLI/TUI Implementation

### 3.1 Command Line Integration

Modify `cmd/gitcells/main.go`:

```go
// Add global TUI flag
rootCmd.PersistentFlags().Bool("tui", false, "Launch interactive TUI mode")

// Add tui subcommand for direct TUI access
tuiCmd := &cobra.Command{
    Use:   "tui",
    Short: "Launch interactive TUI mode",
    Run: func(cmd *cobra.Command, args []string) {
        tui.Run()
    },
}

// Support --tui flag on specific commands
initCmd.Flags().Bool("tui", false, "Use TUI setup wizard")
```

### 3.2 Progressive Enhancement

1. Keep all existing CLI commands functional
2. Add `--tui` flag to commands that benefit from interactivity
3. Auto-detect when TUI would be helpful (e.g., conflicts detected)
4. Provide `GITCELLS_PREFER_TUI` environment variable

## Phase 4: Component Implementation

### 4.1 Reusable Components

**File Picker Component:**
- Directory tree navigation
- File filtering (Excel files only)
- Multi-select support
- Path preview
- Recent selections

**Excel Diff Viewer:**
- Text-based cell representation
- Syntax highlighting for formulas
- Side-by-side comparison
- Navigate by sheets/cells
- Export diff to file

**Progress Component:**
- Multiple concurrent operations
- Time estimates
- Cancel support
- Error handling display

### 4.2 Integration with Existing Code

Reuse internal packages without modification:
- `internal/converter` - Excel↔JSON conversion
- `internal/git` - Git operations
- `internal/config` - Configuration management
- `internal/watcher` - File monitoring

Create adapter layer in TUI package:
- Wrap operations with progress reporting
- Convert errors to user-friendly messages
- Handle async operations with tea.Cmd

## Phase 5: Enhanced TUI Features

### 5.1 Excel Preview Mode

- Show sheet names and dimensions
- Display cell statistics
- Preview formulas and values
- Identify potential issues

### 5.2 Keyboard Shortcuts

Global:
- `?` - Context-sensitive help
- `q` - Quit with confirmation
- `Ctrl+C` - Cancel operation
- `/` - Search/filter

Navigation:
- `Tab`/`Shift+Tab` - Focus navigation
- `Enter` - Select/confirm
- `Esc` - Cancel/back

### 5.3 Theme Support

- Light/dark theme toggle
- Customizable colors
- Accessible mode (high contrast)
- Save preferences

## Phase 6: Testing Strategy

### 6.1 Unit Tests

- Test each model independently
- Mock terminal for view testing
- Component isolation tests
- Key binding tests

### 6.2 Integration Tests

- Full workflow tests
- File system interaction
- Git operation verification
- Configuration persistence

### 6.3 Manual Testing Checklist

- [ ] Keyboard navigation works smoothly
- [ ] All shortcuts documented and functional
- [ ] Resize handling works correctly
- [ ] Color themes display properly
- [ ] Error messages are helpful
- [ ] Progress indicators accurate
- [ ] File operations are safe

## Phase 7: Migration Path

### 7.1 Release Strategy

1. **v2.0-beta**: TUI features behind `--tui` flag
2. **v2.1**: TUI features promoted but CLI default
3. **v2.2**: TUI default with `--cli` flag for legacy
4. **v3.0**: Full TUI with CLI compatibility layer

### 7.2 Documentation Updates

- Update README with TUI screenshots
- Create TUI user guide
- Add keyboard shortcut reference
- Video tutorials for complex workflows

## Implementation Priority

1. **High Priority** (Core functionality):
   - Basic TUI app structure
   - Setup wizard
   - Status dashboard

2. **Medium Priority** (Enhanced workflows):
   - Branch management
   - Conflict resolution
   - File picker component

3. **Low Priority** (Nice to have):
   - Themes and customization
   - Excel preview
   - Advanced keyboard shortcuts

## Technical Considerations

### Performance
- Lazy loading for large file lists
- Efficient diff rendering
- Debounced file system events
- Background operation queuing

### Error Handling
- Graceful degradation to CLI
- Clear error messages
- Recovery suggestions
- Operation rollback support

### Platform Compatibility
- Test on Windows Terminal
- Verify on macOS Terminal/iTerm2
- Support for Linux terminals
- Handle limited color terminals

## Success Metrics

- User can complete setup without documentation
- Conflict resolution time reduced by 50%
- Branch operations are more intuitive
- Error rate decreased
- User satisfaction improved

## Next Steps

1. Create proof of concept for setup wizard
2. Gather user feedback on TUI mockups
3. Implement core TUI structure
4. Iteratively add features based on usage
5. Comprehensive testing before release