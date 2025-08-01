# Terminal User Interface (TUI) Guide

GitCells includes an interactive Terminal User Interface that makes it easy to manage Excel file tracking without memorizing commands. This guide explains how to use the TUI effectively.

## Starting the TUI

Launch the TUI from any directory:

```bash
gitcells tui
```

Or use the global flag with any command:
```bash
gitcells --tui
```

## Main Menu

When you start the TUI, you'll see the main menu with four options:

```
GitCells
Excel Version Control Management

▶ Setup Wizard
    Configure GitCells for your Excel tracking repository

  Status Dashboard
    Monitor Excel file tracking and conversion status

  Settings
    Update, uninstall, and manage GitCells system settings

  Error Logs
    View application errors and troubleshooting information

Use ↑/↓ or j/k to navigate, Enter to select, Ctrl+L for error logs, q to quit
```

### Navigation

- **Arrow Keys** (↑/↓) or **j/k**: Move between options
- **Enter**: Select an option
- **q**: Quit to main menu or exit
- **Ctrl+L**: Jump to error logs from anywhere
- **Esc**: Go back/cancel

## Setup Wizard

The Setup Wizard helps you configure GitCells for your project.

### Step 1: Welcome Screen
```
Welcome to GitCells Setup!

This wizard will help you:
• Initialize a Git repository (if needed)
• Configure GitCells for your Excel files
• Set up automatic tracking
• Start monitoring your files

Press Enter to continue or 'q' to quit
```

### Step 2: Directory Selection
```
Select Project Directory

Current: /Users/username/Documents/ExcelFiles

Options:
[Enter] Use current directory
[b] Browse for directory
[n] Enter new path
[q] Quit setup

Your choice:
```

### Step 3: Git Repository
```
Git Repository Setup

◯ Initialize new Git repository
◉ Use existing repository
◯ Skip Git setup (not recommended)

Repository status: ✓ Git repository found

[Space] to select, [Enter] to continue
```

### Step 4: Tracking Configuration
```
Configure File Tracking

Directories to watch:
✓ Current directory (.)
□ Subdirectory: reports/
□ Subdirectory: data/
□ Add custom directory...

File types to track:
✓ Excel files (.xlsx)
✓ Legacy Excel (.xls)
✓ Macro-enabled (.xlsm)
□ Excel binary (.xlsb)

[Space] to toggle, [a] to add directory, [Enter] to continue
```

### Step 5: Advanced Options
```
Advanced Configuration

Auto-commit changes: ◉ Yes ◯ No
Auto-push to remote: ◯ Yes ◉ No
Preserve formulas: ◉ Yes ◯ No
Preserve styles: ◉ Yes ◯ No
Debounce delay: 2s [+/-] to adjust

[Tab] to move between options, [Enter] to continue
```

### Step 6: Confirmation
```
Setup Summary

✓ Directory: /Users/username/Documents/ExcelFiles
✓ Git repository: Existing
✓ Watching: Current directory
✓ File types: .xlsx, .xls, .xlsm
✓ Auto-commit: Enabled
✓ Configuration saved to .gitcells.yaml

Start watching for changes now? [Y/n]:
```

## Status Dashboard

The Status Dashboard provides real-time monitoring of your Excel files.

### Overview Section
```
GitCells Status Dashboard

Repository: /Users/username/Documents/ExcelFiles
Status: ● Watching (2 watchers active)
Git Branch: main (clean)

Statistics:
Files Tracked: 24
Total Size: 15.3 MB
Last Change: 2 minutes ago
```

### Active Watchers
```
Active Watchers:
┌─────────────────────┬──────────┬─────────────┐
│ Directory           │ Status   │ Files       │
├─────────────────────┼──────────┼─────────────┤
│ ./reports           │ ● Active │ 12 tracked  │
│ ./data             │ ● Active │ 8 tracked   │
│ ./templates        │ ⚠ Error  │ 4 tracked   │
└─────────────────────┴──────────┴─────────────┘

[Enter] to view details, [r] to restart watcher
```

### Recent Activity
```
Recent Activity:
┌─────────────────────┬───────────┬──────────────┬─────────┐
│ File                │ Action    │ Time         │ Status  │
├─────────────────────┼───────────┼──────────────┼─────────┤
│ Budget2024.xlsx     │ Modified  │ 2 mins ago   │ ✓ Done  │
│ Report_Q4.xlsx      │ Created   │ 15 mins ago  │ ✓ Done  │
│ Data.xlsx           │ Modified  │ 1 hour ago   │ ✓ Done  │
│ Template.xlsx       │ Error     │ 2 hours ago  │ ⚠ Failed│
└─────────────────────┴───────────┴──────────────┴─────────┘

[Enter] to view details, [d] to show diff
```

### Quick Actions
```
Quick Actions:
[s] Sync all files    [w] Add watcher
[c] Check status      [g] Git operations
[r] Refresh          [?] Help
```

## Settings Menu

The Settings menu provides system-wide GitCells management.

### Main Settings
```
GitCells Settings

▶ Check for Updates
    Check if a newer version is available

  Update GitCells
    Download and install the latest version

  Configuration
    Edit GitCells configuration

  Clear Cache
    Remove temporary files and cache

  Uninstall
    Remove GitCells from your system

[Enter] to select, [q] to go back
```

### Update Screen
```
GitCells Update

Current Version: v0.3.0
Latest Version: v0.3.1 ✓ Update available!

Release Notes:
- Fixed Excel 2019 compatibility
- Improved performance for large files
- Added support for pivot tables

Update now? [Y/n]:

[●●●●●●●●○○] 80% Downloading...
```

### Configuration Editor
```
Configuration Editor

Current configuration file: .gitcells.yaml

1. Watcher Settings
   Debounce: 2s [+/-]
   Extensions: .xlsx, .xls, .xlsm [e]dit

2. Git Settings
   Auto-commit: ✓ Enabled [toggle]
   Auto-push: ✗ Disabled [toggle]
   
3. Converter Settings
   Preserve formulas: ✓ [toggle]
   Preserve styles: ✓ [toggle]
   Compact JSON: ✗ [toggle]

[s] Save changes, [r] Reset to defaults, [q] Cancel
```

## Error Logs

The Error Logs view helps troubleshoot issues.

### Log Viewer
```
GitCells Error Logs

Filter: [All Types ▼] [Last 24 hours ▼] 🔍 Search...

┌──────────────┬─────────┬────────────────────────────────┐
│ Time         │ Level   │ Message                        │
├──────────────┼─────────┼────────────────────────────────┤
│ 10:32:15     │ ERROR   │ Failed to convert Budget.xlsx  │
│ 10:32:15     │ INFO    │ File locked by another process │
│ 09:15:42     │ WARNING │ Large file detected (>50MB)    │
│ 08:45:21     │ ERROR   │ Git push failed: auth required │
└──────────────┴─────────┴────────────────────────────────┘

[↑/↓] Navigate, [Enter] View details, [c] Clear logs
[e] Export logs, [f] Filter, [r] Refresh
```

### Error Details
```
Error Details

Time: 2024-01-15 10:32:15
Level: ERROR
Component: Converter

Message: Failed to convert Budget.xlsx

Details:
The file appears to be locked by another process.
This usually happens when the file is open in Excel.

Suggested Actions:
1. Close the file in Excel
2. Check if another user has the file open
3. Restart the file watcher

Stack Trace: (Press 't' to toggle)

[b] Back to logs, [c] Copy details, [?] Get help
```

## Keyboard Shortcuts

### Global Shortcuts
- **Ctrl+L**: Jump to error logs
- **Ctrl+R**: Refresh current view
- **Ctrl+C**: Exit application
- **?**: Show context help

### Navigation
- **↑/↓** or **j/k**: Move up/down
- **←/→** or **h/l**: Move left/right (in tables)
- **Page Up/Down**: Scroll pages
- **Home/End**: Go to start/end
- **Tab**: Next field
- **Shift+Tab**: Previous field

### Actions
- **Enter**: Select/Confirm
- **Space**: Toggle checkbox
- **Esc**: Cancel/Go back
- **q**: Quit to previous screen

## Tips and Tricks

### Quick Status Check

Press `Ctrl+L` from any screen to quickly check for errors, then press `q` to return.

### Efficient Navigation

Use vim-style keys (h,j,k,l) for faster navigation if you're familiar with them.

### Monitoring Mode

In the Status Dashboard, the display auto-refreshes every 5 seconds. Press `p` to pause/resume auto-refresh.

### Batch Operations

In file lists, use:
- **a**: Select all
- **n**: Select none  
- **i**: Invert selection
- **Enter**: Perform action on selected

### Quick Filter

In any list view, start typing to filter items. Press `Esc` to clear the filter.

## Troubleshooting TUI Issues

### Display Problems

If the TUI looks garbled:
1. Ensure your terminal supports UTF-8
2. Try a different terminal emulator
3. Set environment variable: `export LANG=en_US.UTF-8`

### Performance Issues

For better performance:
1. Use a modern terminal (iTerm2, Windows Terminal, etc.)
2. Reduce terminal window size if very large
3. Disable transparency/blur effects

### Color Issues

If colors don't display correctly:
1. Check terminal color support: `echo $TERM`
2. Try: `export TERM=xterm-256color`
3. Adjust terminal color scheme

## Next Steps

- Explore [Configuration](configuration.md) options in detail
- Learn about [File Watching](watching.md) 
- Understand [Git Integration](git-integration.md)
- Check [Troubleshooting](troubleshooting.md) for common issues