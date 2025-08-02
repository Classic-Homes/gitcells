# Installation

GitCells can be installed in several ways depending on your operating system and preferences. Choose the method that works best for you.

## System Requirements

- **Operating System**: macOS, Linux, or Windows
- **Git**: Version 2.0 or higher (for Git integration features)
- **Disk Space**: Approximately 30MB for the binary

## Installation Methods

### Using Pre-built Binaries (Recommended)

The easiest way to install GitCells is to download a pre-built binary for your platform.

#### macOS

1. Download the latest release:
```bash
# For Intel Macs
curl -L https://github.com/Classic-Homes/gitcells/releases/latest/download/gitcells-macos-intel.tar.gz -o gitcells-macos-intel.tar.gz
tar -xzf gitcells-macos-intel.tar.gz

# For Apple Silicon Macs
curl -L https://github.com/Classic-Homes/gitcells/releases/latest/download/gitcells-macos-apple-silicon.tar.gz -o gitcells-macos-apple-silicon.tar.gz
tar -xzf gitcells-macos-apple-silicon.tar.gz
```

2. Make executable and move to your PATH:
```bash
# Make executable
chmod +x gitcells-macos-*

# For Intel Macs
sudo mv gitcells-macos-intel /usr/local/bin/gitcells

# For Apple Silicon Macs
sudo mv gitcells-macos-apple-silicon /usr/local/bin/gitcells
```

3. Verify installation:
```bash
gitcells version
```

#### Linux

1. Download the latest release:
```bash
# For x86_64 Linux
curl -L https://github.com/Classic-Homes/gitcells/releases/latest/download/gitcells-linux.tar.gz -o gitcells-linux.tar.gz
tar -xzf gitcells-linux.tar.gz

# For ARM64 Linux
curl -L https://github.com/Classic-Homes/gitcells/releases/latest/download/gitcells-linux-arm64.tar.gz -o gitcells-linux-arm64.tar.gz
tar -xzf gitcells-linux-arm64.tar.gz
```

2. Make executable and move to your PATH:
```bash
# Make executable
chmod +x gitcells-linux*

# For x86_64 Linux
sudo mv gitcells-linux /usr/local/bin/gitcells

# For ARM64 Linux
sudo mv gitcells-linux-arm64 /usr/local/bin/gitcells
```

3. Verify installation:
```bash
gitcells version
```

#### Windows

1. Download the latest release:
```powershell
# Download the executable directly
curl -L https://github.com/Classic-Homes/gitcells/releases/latest/download/gitcells-windows.exe -o gitcells.exe

# Or download the zip archive
curl -L https://github.com/Classic-Homes/gitcells/releases/latest/download/gitcells-windows.zip -o gitcells-windows.zip
Expand-Archive gitcells-windows.zip .
```

2. Add to your PATH:
   - Move `gitcells.exe` to a directory in your PATH (e.g., `C:\Windows\System32`)
   - Or add the current directory to your PATH environment variable

3. Verify installation:
```cmd
gitcells version
```

### Building from Source

If you prefer to build GitCells from source, you'll need Go 1.21 or higher installed.

1. Clone the repository:
```bash
git clone https://github.com/Classic-Homes/gitcells.git
cd gitcells
```

2. Install dependencies:
```bash
go mod download
```

3. Build the binary:
```bash
make build
```

This will create a binary for your current platform in the `dist/` directory.

4. Install locally:
```bash
# macOS/Linux
sudo cp dist/gitcells /usr/local/bin/gitcells
sudo chmod +x /usr/local/bin/gitcells

# Windows (run as Administrator)
copy dist\gitcells.exe C:\Windows\System32\gitcells.exe
```

### Using Docker

GitCells is also available as a Docker image:

```bash
docker pull ghcr.io/classic-homes/gitcells:latest
```

To use GitCells with Docker:
```bash
docker run -v $(pwd):/workspace ghcr.io/classic-homes/gitcells:latest [command]
```

## Post-Installation Setup

After installation, you may want to:

1. **Enable Auto-completion** (bash/zsh):
```bash
gitcells completion bash > /etc/bash_completion.d/gitcells
# or for zsh
gitcells completion zsh > "${fpath[1]}/_gitcells"
```

2. **Check for Updates**:
```bash
gitcells update --check
```

3. **Initialize GitCells** in your project:
```bash
gitcells init
```

## Updating GitCells

GitCells includes a self-update feature:

```bash
# Check for updates
gitcells update --check

# Update to latest stable version
gitcells update

# Update to latest version (including pre-releases)
gitcells update --prerelease
```

## Uninstallation

To uninstall GitCells:

### macOS/Linux
```bash
sudo rm /usr/local/bin/gitcells
```

### Windows
Remove `gitcells.exe` from your PATH directory.

### Docker
```bash
docker rmi ghcr.io/classic-homes/gitcells:latest
```

## Troubleshooting Installation

### Permission Denied
If you get a permission denied error, ensure you're using `sudo` for system directories or install to a user directory in your PATH.

### Command Not Found
Make sure the GitCells binary is in a directory listed in your PATH environment variable.

### Version Compatibility
If you encounter issues, ensure you're using a compatible version:
```bash
gitcells version --check-update
```

## Next Steps

- Read the [Quick Start Guide](quickstart.md) to begin using GitCells
- Learn about [Basic Concepts](concepts.md) to understand how GitCells works
- Configure GitCells for your project with the [Configuration Guide](../user-guide/configuration.md)