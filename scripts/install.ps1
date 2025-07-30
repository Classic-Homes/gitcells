# SheetSync PowerShell Installation Script
# This script downloads and installs the latest release of SheetSync on Windows

param(
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\SheetSync",
    [string]$Version = "latest",
    [switch]$AddToPath = $true,
    [switch]$Help
)

# Configuration
$Repo = "Classic-Homes/sheetsync"
$BinaryName = "sheetsync.exe"
$ConfigDir = "$env:APPDATA\sheetsync"

# Helper functions
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $color = switch ($Level) {
        "INFO" { "Blue" }
        "WARN" { "Yellow" }
        "ERROR" { "Red" }
        "SUCCESS" { "Green" }
        default { "White" }
    }
    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] [$Level] $Message" -ForegroundColor $color
}

function Write-Error-And-Exit {
    param([string]$Message)
    Write-Log $Message "ERROR"
    exit 1
}

function Show-Help {
    Write-Host @"
SheetSync PowerShell Installation Script

Usage: .\install.ps1 [options]

Options:
  -InstallDir DIR     Installation directory (default: $env:LOCALAPPDATA\Programs\SheetSync)
  -Version VERSION    Specific version to install (default: latest)
  -AddToPath          Add installation directory to PATH (default: true)
  -Help              Show this help message

Examples:
  .\install.ps1                                    # Install latest version
  .\install.ps1 -InstallDir "C:\Tools\SheetSync"  # Install to custom directory
  .\install.ps1 -Version "v1.0.0"                 # Install specific version
  .\install.ps1 -AddToPath:`$false                # Don't add to PATH

For more information, visit: https://github.com/$Repo
"@
}

function Get-LatestVersion {
    Write-Log "Fetching latest release information..."
    
    try {
        $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"
        $response = Invoke-RestMethod -Uri $apiUrl -Method Get
        return $response.tag_name
    }
    catch {
        Write-Error-And-Exit "Failed to fetch latest version: $($_.Exception.Message)"
    }
}

function Download-Binary {
    param([string]$Version)
    
    $tempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
    Write-Log "Created temporary directory: $tempDir"
    
    # Construct download URL
    $filename = "sheetsync-$Version-windows-amd64.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$filename"
    $tempFile = Join-Path $tempDir $filename
    
    Write-Log "Downloading from: $downloadUrl"
    
    try {
        # Download with progress
        $progressPreference = 'Continue'
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
        Write-Log "Download completed"
    }
    catch {
        Write-Error-And-Exit "Download failed: $($_.Exception.Message)"
    }
    
    # Extract archive
    try {
        Write-Log "Extracting archive..."
        Expand-Archive -Path $tempFile -DestinationPath $tempDir -Force
        
        # Find the binary
        $binaryPath = Join-Path $tempDir "sheetsync-windows-amd64.exe"
        if (-not (Test-Path $binaryPath)) {
            Write-Error-And-Exit "Binary not found in archive: $binaryPath"
        }
        
        return @{
            BinaryPath = $binaryPath
            TempDir = $tempDir
        }
    }
    catch {
        Write-Error-And-Exit "Failed to extract archive: $($_.Exception.Message)"
    }
}

function Install-Binary {
    param([string]$BinaryPath, [string]$InstallPath)
    
    Write-Log "Installing to $InstallPath..."
    
    # Create install directory
    $installDir = Split-Path $InstallPath -Parent
    if (-not (Test-Path $installDir)) {
        try {
            New-Item -ItemType Directory -Path $installDir -Force | Out-Null
            Write-Log "Created install directory: $installDir"
        }
        catch {
            Write-Error-And-Exit "Failed to create install directory: $($_.Exception.Message)"
        }
    }
    
    # Copy binary
    try {
        Copy-Item $BinaryPath $InstallPath -Force
        Write-Log "Binary installed to $InstallPath" "SUCCESS"
    }
    catch {
        Write-Error-And-Exit "Failed to copy binary: $($_.Exception.Message)"
    }
}

function Add-ToPath {
    param([string]$Directory)
    
    Write-Log "Adding $Directory to PATH..."
    
    try {
        # Get current user PATH
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        
        # Check if directory is already in PATH
        $pathEntries = $currentPath -split ";"
        if ($pathEntries -contains $Directory) {
            Write-Log "Directory already in PATH"
            return
        }
        
        # Add to PATH
        $newPath = if ($currentPath) { "$currentPath;$Directory" } else { $Directory }
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        
        # Update current session PATH
        $env:PATH = "$env:PATH;$Directory"
        
        Write-Log "Added to PATH successfully" "SUCCESS"
    }
    catch {
        Write-Log "Failed to add to PATH: $($_.Exception.Message)" "WARN"
        Write-Log "You can manually add $Directory to your PATH"
    }
}

function New-DefaultConfig {
    Write-Log "Creating default configuration..."
    
    if (-not (Test-Path $ConfigDir)) {
        New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
    }
    
    $configFile = Join-Path $ConfigDir "config.yaml"
    if (-not (Test-Path $configFile)) {
        $configContent = @"
version: 1.0

git:
  branch: main
  auto_push: false
  auto_pull: true
  user_name: "SheetSync"
  user_email: "sheetsync@localhost"
  commit_template: "SheetSync: {action} {filename} at {timestamp}"

watcher:
  directories: []
  ignore_patterns:
    - "~`$*"
    - "*.tmp"
    - ".~lock.*"
  debounce_delay: 2s
  file_extensions:
    - ".xlsx"
    - ".xls"
    - ".xlsm"

converter:
  preserve_formulas: true
  preserve_styles: true
  preserve_comments: true
  compact_json: false
  ignore_empty_cells: true
  max_cells_per_sheet: 1000000
"@
        
        try {
            $configContent | Out-File -FilePath $configFile -Encoding UTF8
            Write-Log "Created default configuration: $configFile" "SUCCESS"
        }
        catch {
            Write-Log "Failed to create configuration: $($_.Exception.Message)" "WARN"
        }
    }
    else {
        Write-Log "Configuration already exists: $configFile"
    }
}

function Test-Installation {
    param([string]$InstallPath)
    
    Write-Log "Verifying installation..."
    
    if (Test-Path $InstallPath) {
        try {
            $version = & $InstallPath --version 2>$null | Select-Object -First 1
            Write-Log "SheetSync installed successfully: $version" "SUCCESS"
            
            Write-Host ""
            Write-Host "To get started:" -ForegroundColor Green
            Write-Host "  sheetsync --help"
            Write-Host "  sheetsync init"
            Write-Host ""
            Write-Host "Documentation: https://github.com/$Repo#readme"
            
            return $true
        }
        catch {
            Write-Log "Installation verification failed: $($_.Exception.Message)" "WARN"
            return $false
        }
    }
    else {
        Write-Log "Binary not found at $InstallPath" "ERROR"
        return $false
    }
}

function Remove-TempDirectory {
    param([string]$TempDir)
    
    if ($TempDir -and (Test-Path $TempDir)) {
        try {
            Remove-Item $TempDir -Recurse -Force
            Write-Log "Cleaned up temporary files"
        }
        catch {
            Write-Log "Failed to clean up temporary files: $($_.Exception.Message)" "WARN"
        }
    }
}

# Main installation function
function Install-SheetSync {
    Write-Log "Starting SheetSync installation..."
    
    # Resolve version
    $targetVersion = if ($Version -eq "latest") { Get-LatestVersion } else { $Version }
    Write-Log "Target version: $targetVersion"
    
    # Download and extract
    $download = Download-Binary $targetVersion
    
    try {
        # Install binary
        $installPath = Join-Path $InstallDir $BinaryName
        Install-Binary $download.BinaryPath $installPath
        
        # Add to PATH if requested
        if ($AddToPath) {
            Add-ToPath $InstallDir
        }
        
        # Create default configuration
        New-DefaultConfig
        
        # Verify installation
        $success = Test-Installation $installPath
        
        if ($success) {
            Write-Log "Installation completed successfully!" "SUCCESS"
        }
        else {
            Write-Error-And-Exit "Installation verification failed"
        }
    }
    finally {
        # Cleanup
        Remove-TempDirectory $download.TempDir
    }
}

# Main script execution
if ($Help) {
    Show-Help
    exit 0
}

# Check PowerShell version
if ($PSVersionTable.PSVersion.Major -lt 5) {
    Write-Error-And-Exit "PowerShell 5.0 or later is required"
}

# Check if running as administrator for system-wide installation
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")

if ($InstallDir.StartsWith($env:ProgramFiles) -and -not $isAdmin) {
    Write-Error-And-Exit "Administrator privileges required for installation to $InstallDir. Run PowerShell as Administrator or choose a different install directory."
}

# Run installation
try {
    Install-SheetSync
}
catch {
    Write-Error-And-Exit "Installation failed: $($_.Exception.Message)"
}