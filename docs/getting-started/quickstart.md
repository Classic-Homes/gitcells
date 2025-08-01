# Quick Start Guide

This guide will help you get started with GitCells in just a few minutes. No technical expertise required!

## What You'll Learn

- How to set up GitCells for your Excel files
- How to track changes to your Excel files automatically
- How to view what changed in your Excel files

## Step 1: Open Terminal (Command Line)

First, we need to open the terminal application:

- **Windows**: Press `Windows + R`, type `cmd`, and press Enter
- **Mac**: Press `Cmd + Space`, type `terminal`, and press Enter
- **Linux**: Press `Ctrl + Alt + T`

Don't worry if you've never used the terminal before - we'll guide you through each step!

## Step 2: Navigate to Your Excel Files

In the terminal, navigate to the folder containing your Excel files. For example:

```bash
# Windows example
cd C:\Users\YourName\Documents\ExcelFiles

# Mac example
cd /Users/YourName/Documents/ExcelFiles

# Linux example
cd /home/yourname/Documents/ExcelFiles
```

💡 **Tip**: You can also drag and drop a folder onto the terminal window to get its path!

## Step 3: Initialize GitCells

Now let's set up GitCells in your folder. Type this command and press Enter:

```bash
gitcells init
```

This creates a configuration file that tells GitCells how to handle your Excel files.

## Step 4: Start Tracking Your Excel Files

### Option A: Use the Interactive Interface (Recommended for Beginners)

Type this command to open GitCells' user-friendly interface:

```bash
gitcells tui
```

You'll see a menu with these options:
- **Setup Wizard** - Helps configure GitCells for your needs
- **Status Dashboard** - Shows which Excel files are being tracked
- **Settings** - Manage GitCells settings
- **Error Logs** - View any problems that occurred

Use the arrow keys to navigate and press Enter to select an option.

### Option B: Start Watching Automatically

If you prefer, you can start watching your Excel files directly:

```bash
gitcells watch .
```

This tells GitCells to watch the current folder (the `.` means "current folder") for any Excel file changes.

## Step 5: Make Changes to Your Excel Files

Now the magic happens! 

1. Open any Excel file in your folder
2. Make some changes (add data, modify formulas, etc.)
3. Save the file

GitCells will automatically:
- Detect that your Excel file changed
- Convert it to a trackable format
- Save a snapshot of the changes

You'll see messages in the terminal showing what GitCells is doing.

## Step 6: View Your File History

To see what files GitCells is tracking:

```bash
gitcells status
```

This shows:
- Which Excel files are being tracked
- When they were last modified
- If there are any pending changes

## What's Next?

### View Changes in Detail

To see what changed in a specific Excel file:

```bash
gitcells diff YourFile.xlsx
```

This shows exactly what cells, formulas, or data changed.

### Convert Files Manually

You can also convert files between Excel and JSON format:

```bash
# Convert Excel to JSON (to see the trackable format)
gitcells convert YourFile.xlsx

# Convert JSON back to Excel
gitcells convert YourFile.xlsx.json
```

### Keep GitCells Updated

Check for updates regularly:

```bash
gitcells update --check
```

If an update is available, install it with:

```bash
gitcells update
```

## Common Questions

**Q: Where are my Excel files?**  
A: GitCells doesn't move or modify your original Excel files. It creates companion JSON files that track the changes.

**Q: Can I still use Excel normally?**  
A: Yes! GitCells works in the background. Use Excel exactly as you always have.

**Q: What if I make a mistake?**  
A: GitCells tracks all changes, so you can always see previous versions of your files.

**Q: Do I need to keep the terminal open?**  
A: Yes, when using the `watch` command. The terminal needs to stay open for GitCells to monitor your files. You can minimize it.

## Getting Help

If you run into any issues:

1. Check the [Troubleshooting Guide](../user-guide/troubleshooting.md)
2. Use the TUI interface (`gitcells tui`) and check the Error Logs
3. Visit our [GitHub Issues page](https://github.com/Classic-Homes/gitcells/issues)

## Next Steps

- Learn more about [How GitCells Works](concepts.md)
- Explore [Configuration Options](../user-guide/configuration.md) to customize GitCells
- Set up [Automatic Git Integration](../user-guide/git-integration.md) for team collaboration