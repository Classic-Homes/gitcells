# GitCells TUI

The GitCells Terminal User Interface (TUI) provides an interactive way to manage Excel file tracking and conversion.

## Design Philosophy

The TUI is focused exclusively on GitCells-specific features:
- Excel file tracking configuration
- Monitoring conversion status

Users should use their preferred git tools (command line, SourceTree, GitHub Desktop, etc.) for all git operations including:
- Creating and switching branches
- Committing changes
- Pushing and pulling
- Merging branches
- Resolving conflicts

## Available Modes

### 1. Setup Wizard
Configure GitCells for your Excel tracking repository:
- Set directories to watch
- Configure file extensions
- Set conversion options

### 2. Status Dashboard
Monitor your Excel file tracking in real-time:
- View tracked Excel files
- Monitor conversion status
- See pending conversions
- Check auto-sync status

## Usage

```bash
gitcells tui
```

## Key Bindings

- `Tab` - Switch between tabs/sections
- `↑/↓` or `j/k` - Navigate lists
- `Enter` - Select/confirm
- `Esc` - Go back/cancel
- `q` - Quit
- `?` - Show help

## Integration with Git

GitCells creates a git repository to track your Excel files as JSON. Once set up, you can use any git workflow:

1. GitCells watches and converts Excel files to JSON
2. Use your git tools to commit, branch, merge, and collaborate
3. For Excel-aware conflict resolution, use `gitcells` CLI commands

This separation allows teams to use their existing git workflows while GitCells handles the Excel-specific conversion challenges.