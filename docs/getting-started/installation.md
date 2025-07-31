# Installation Guide

## Quick Install

### macOS/Linux

Run this command in your terminal:

```bash
curl -sSL https://raw.githubusercontent.com/Classic-Homes/gitcells/main/scripts/install.sh | bash
```

### Windows

Run this command in PowerShell as Administrator:

```powershell
iwr -useb https://raw.githubusercontent.com/Classic-Homes/gitcells/main/scripts/install.ps1 | iex
```

## Manual Installation

### 1. Download the Binary

Download the appropriate version for your system from the [releases page](https://github.com/Classic-Homes/gitcells/releases):

- **macOS**: `gitcells-darwin-amd64` (Intel) or `gitcells-darwin-arm64` (Apple Silicon)
- **Windows**: `gitcells-windows-amd64.exe`
- **Linux**: `gitcells-linux-amd64`

### 2. Install the Binary

#### macOS/Linux

```bash
# Make it executable
chmod +x gitcells-*

# Move to a directory in your PATH
sudo mv gitcells-* /usr/local/bin/gitcells

# Verify installation
gitcells --version
```

#### Windows

1. Rename the downloaded file to `gitcells.exe`
2. Move it to `C:\Program Files\GitCells\`
3. Add `C:\Program Files\GitCells\` to your PATH environment variable
4. Open a new Command Prompt and verify: `gitcells --version`

## Building from Source

### Prerequisites

- Go 1.19 or later
- Git
- Make (optional)

### Build Steps

```bash
# Clone the repository
git clone https://github.com/Classic-Homes/gitcells.git
cd gitcells

# Build
go build -o gitcells cmd/gitcells/main.go

# Or use make
make build

# Install
sudo mv gitcells /usr/local/bin/
```

## Verifying Installation

Run these commands to verify GitCells is properly installed:

```bash
# Check version
gitcells --version

# View help
gitcells --help

# Check Git integration
git --version
```

## Uninstalling

### macOS/Linux

```bash
curl -sSL https://raw.githubusercontent.com/Classic-Homes/gitcells/main/scripts/uninstall.sh | bash
```

### Windows

```powershell
# Remove from Program Files
Remove-Item -Path "C:\Program Files\GitCells" -Recurse -Force

# Remove from PATH (manually through System Properties)
```

## Next Steps

- Follow the [Quick Start Guide](quickstart.md) to begin using GitCells
- Learn about [Basic Concepts](concepts.md) 
- Configure GitCells with a [configuration file](../reference/configuration.md)