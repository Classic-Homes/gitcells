<p align="center">
  <img src="assets/logo.png" alt="GitCells Logo" width="200">
</p>

# Welcome to GitCells

GitCells bridges Excel and Git, enabling version control for spreadsheet files by converting them to trackable JSON format.

## What is GitCells?

GitCells solves a common problem: Excel files are binary formats that don't work well with Git version control. GitCells converts Excel files to human-readable JSON, enabling:

- **Track Changes**: See exactly what changed in your spreadsheets
- **Collaborate**: Multiple people can work on the same files
- **History**: Keep a complete history of all modifications
- **Merge**: Resolve conflicts when changes overlap
- **Automation**: Automatically track and commit changes

## Key Features

<div class="grid cards" markdown>

- **Excel Support**  
  Works with .xlsx, .xls, .xlsm files while preserving formulas, styles, and charts

- **Automatic Tracking**  
  Watch directories and automatically convert files when they change

- **Git Integration**  
  Seamlessly integrates with Git for version control

- **Terminal UI**  
  User-friendly interface for non-technical users

- **Self-Updating**  
  Built-in update mechanism keeps GitCells current

- **Full Preservation**  
  Maintains formulas, styles, comments, charts, and pivot tables

</div>

## Quick Start

### 1. Install GitCells

=== "Windows"

    Download and run the installer:
    ```bash
    # Download the latest release
    curl -L https://github.com/Classic-Homes/gitcells/releases/latest/download/gitcells-windows.exe -o gitcells.exe

    # Add to PATH and verify
    gitcells version
    ```

=== "macOS"

    Install using curl:
    ```bash
    # Download the latest release
    curl -L https://github.com/Classic-Homes/gitcells/releases/latest/download/gitcells-macos -o gitcells

    # Make executable and move to PATH
    chmod +x gitcells
    sudo mv gitcells /usr/local/bin/

    # Verify installation
    gitcells version
    ```

=== "Linux"

    Install using curl:
    ```bash
    # Download the latest release
    curl -L https://github.com/Classic-Homes/gitcells/releases/latest/download/gitcells-linux -o gitcells

    # Make executable and move to PATH
    chmod +x gitcells
    sudo mv gitcells /usr/local/bin/

    # Verify installation
    gitcells version
    ```

### 2. Initialize Your Project

```bash
# Navigate to your Excel files directory
cd /path/to/excel/files

# Initialize GitCells
gitcells init
```

### 3. Start Tracking

```bash
# Watch current directory for changes
gitcells watch .

# Or use the interactive UI
gitcells tui
```

That's it! GitCells will now track all changes to your Excel files.

## How It Works

```mermaid
graph LR
    A[Excel File] -->|Save| B[GitCells Detects]
    B -->|Convert| C[JSON Format]
    C -->|Commit| D[Git Repository]
    D -->|History| E[Track Changes]
```

1. **You edit** your Excel files normally
2. **GitCells detects** when files are saved
3. **Converts to JSON** preserving all data and formatting
4. **Commits to Git** with meaningful messages
5. **Track history** and collaborate with others

## Use Cases

GitCells is perfect for:

- **Financial Teams**: Track budget and forecast changes
- **Data Analysts**: Version control for analysis files
- **Project Managers**: Monitor project tracking spreadsheets
- **HR Departments**: Maintain employee data with audit trails
- **Anyone**: Who needs to track Excel file changes

## Getting Help

- **Documentation**: Browse the guides in the sidebar
- **Quick Start**: [Get started in 5 minutes](getting-started/quickstart.md)
- **Troubleshooting**: [Common issues and solutions](user-guide/troubleshooting.md)
- **GitHub**: [Report issues or contribute](https://github.com/Classic-Homes/gitcells)

## Ready to Start?

<div class="grid cards" markdown>

- **[Quick Start Guide](getting-started/quickstart.md)**  
  Get up and running in minutes

- **[User Guide](user-guide/configuration.md)**  
  Learn all features and options

- **[Get Help](user-guide/troubleshooting.md)**  
  Troubleshooting and support

</div>
