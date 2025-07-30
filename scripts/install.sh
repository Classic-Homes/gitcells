#!/bin/bash

# GitCells Installation Script
# This script downloads and installs the latest release of GitCells

set -e

# Configuration
REPO="Classic-Homes/gitcells"
BINARY_NAME="gitcells"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/gitcells"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          error "Unsupported operating system: $(uname -s)" ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)          error "Unsupported architecture: $(uname -m)" ;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release version from GitHub
get_latest_version() {
    log "Fetching latest release information..."
    
    local latest_url="https://api.github.com/repos/${REPO}/releases/latest"
    local version
    
    if command -v curl >/dev/null 2>&1; then
        version=$(curl -s "$latest_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        version=$(wget -qO- "$latest_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget is available. Please install one of them."
    fi
    
    if [ -z "$version" ]; then
        error "Failed to fetch latest version"
    fi
    
    echo "$version"
}

# Download and extract binary
download_binary() {
    local version="$1"
    local platform="$2"
    local temp_dir
    
    temp_dir=$(mktemp -d)
    log "Created temporary directory: $temp_dir"
    
    # Construct download URL
    local filename="${BINARY_NAME}-${version}-${platform}"
    if [[ "$platform" == *"windows"* ]]; then
        filename="${filename}.zip"
    else
        filename="${filename}.tar.gz"
    fi
    
    local download_url="https://github.com/${REPO}/releases/download/${version}/${filename}"
    log "Downloading from: $download_url"
    
    # Download file
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "${temp_dir}/${filename}" "$download_url" || error "Download failed"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "${temp_dir}/${filename}" "$download_url" || error "Download failed"
    else
        error "Neither curl nor wget is available"
    fi
    
    # Extract archive
    cd "$temp_dir"
    if [[ "$filename" == *.zip ]]; then
        unzip -q "$filename" || error "Failed to extract zip file"
    else
        tar -xzf "$filename" || error "Failed to extract tar.gz file"
    fi
    
    # Find the binary
    local binary_path
    if [[ "$platform" == *"windows"* ]]; then
        binary_path="${BINARY_NAME}-${platform}.exe"
    else
        binary_path="${BINARY_NAME}-${platform}"
    fi
    
    if [ ! -f "$binary_path" ]; then
        error "Binary not found in archive: $binary_path"
    fi
    
    echo "${temp_dir}/${binary_path}"
}

# Install binary
install_binary() {
    local binary_path="$1"
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"
    
    log "Installing to $install_path..."
    
    # Check if install directory exists and is writable
    if [ ! -d "$INSTALL_DIR" ]; then
        warn "Install directory $INSTALL_DIR does not exist"
        if [ "$(id -u)" -eq 0 ]; then
            mkdir -p "$INSTALL_DIR"
        else
            error "Please run with sudo or create $INSTALL_DIR manually"
        fi
    fi
    
    if [ ! -w "$INSTALL_DIR" ]; then
        if [ "$(id -u)" -ne 0 ]; then
            error "No write permission to $INSTALL_DIR. Please run with sudo."
        fi
    fi
    
    # Copy binary
    cp "$binary_path" "$install_path" || error "Failed to copy binary"
    chmod +x "$install_path" || error "Failed to make binary executable"
    
    success "Binary installed to $install_path"
}

# Create default configuration
create_config() {
    log "Creating default configuration..."
    
    mkdir -p "$CONFIG_DIR"
    
    local config_file="${CONFIG_DIR}/config.yaml"
    if [ ! -f "$config_file" ]; then
        cat > "$config_file" << 'EOF'
version: 1.0

git:
  branch: main
  auto_push: false
  auto_pull: true
  user_name: "GitCells"
  user_email: "gitcells@localhost"
  commit_template: "GitCells: {action} {filename} at {timestamp}"

watcher:
  directories: []
  ignore_patterns:
    - "~$*"
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
EOF
        success "Created default configuration: $config_file"
    else
        log "Configuration already exists: $config_file"
    fi
}

# Verify installation
verify_installation() {
    log "Verifying installation..."
    
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local version
        version=$("$BINARY_NAME" --version 2>/dev/null | head -n1 || echo "unknown")
        success "GitCells installed successfully: $version"
        
        echo ""
        echo "To get started:"
        echo "  gitcells --help"
        echo "  gitcells init"
        echo ""
        echo "Documentation: https://github.com/${REPO}#readme"
    else
        warn "GitCells was installed but is not in PATH"
        echo "Add $INSTALL_DIR to your PATH or run:"
        echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
}

# Cleanup temporary files
cleanup() {
    if [ -n "$temp_dir" ] && [ -d "$temp_dir" ]; then
        rm -rf "$temp_dir"
        log "Cleaned up temporary files"
    fi
}

# Main installation function
main() {
    log "Starting GitCells installation..."
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    # Check prerequisites
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        error "Neither curl nor wget is available. Please install one of them."
    fi
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    log "Detected platform: $platform"
    
    # Get latest version
    local version
    version=$(get_latest_version)
    log "Latest version: $version"
    
    # Download binary
    local binary_path
    binary_path=$(download_binary "$version" "$platform")
    
    # Install binary
    install_binary "$binary_path"
    
    # Create configuration
    create_config
    
    # Verify installation
    verify_installation
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --version)
            VERSION="$2"
            shift 2
            ;;
        -h|--help)
            echo "GitCells Installation Script"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --install-dir DIR    Installation directory (default: /usr/local/bin)"
            echo "  --version VERSION    Specific version to install (default: latest)"
            echo "  -h, --help          Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Install latest version"
            echo "  $0 --install-dir ~/bin               # Install to ~/bin"
            echo "  $0 --version v1.0.0                  # Install specific version"
            echo ""
            echo "For more information, visit: https://github.com/${REPO}"
            exit 0
            ;;
        *)
            error "Unknown option: $1. Use --help for usage information."
            ;;
    esac
done

# Run main installation
main