# Condensed TUI Overview

The new condensed TUI provides a more streamlined and efficient user experience:

## Key Improvements

### 1. **Unified Dashboard**
- Combines Status Dashboard and Watcher into a single screen with tabs
- Three tabs: Overview, Files, Activity
- Real-time updates every 5 seconds

### 2. **Global Quick Actions**
Always visible at the bottom of the screen:
- `w` - Toggle watcher on/off
- `c` - Convert a file
- `d` - Open diff viewer
- `s` - Open settings
- `?` - Show help/error logs

### 3. **Consistent Navigation**
- `Tab` / `Shift+Tab` - Switch between tabs
- `↑↓` or `j/k` - Navigate lists
- `Enter` - Select/confirm
- `Esc` - Go back to dashboard (from any screen)
- `q` - Quit application

### 4. **Streamlined Workflow**
Instead of navigating through multiple menu levels:
- Start directly in the dashboard
- Use quick actions to access features instantly
- Tab navigation for related information
- No need to return to main menu between tasks

## Usage

Launch the new condensed TUI (now default):
```bash
gitcells tui
```

Or explicitly use v2:
```bash
gitcells tui --v2
```

To use the classic menu-based TUI:
```bash
gitcells tui --v2=false
```

## Benefits

1. **Faster Access** - Quick actions available from any screen
2. **Better Context** - See status, files, and activity together
3. **Reduced Navigation** - Fewer screens to navigate through
4. **Consistent UX** - Same keys work across all screens
5. **Live Updates** - Dashboard auto-refreshes with latest data