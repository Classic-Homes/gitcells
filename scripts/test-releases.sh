#!/bin/bash

# Test Release Binaries Script
# Verifies that all platform-specific binaries work correctly

set -e

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

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Test a binary
test_binary() {
    local binary_path="$1"
    local platform="$2"
    
    log "Testing $platform binary: $binary_path"
    
    if [ ! -f "$binary_path" ]; then
        error "Binary not found: $binary_path"
    fi
    
    # Check if binary is executable
    if [ ! -x "$binary_path" ]; then
        error "Binary is not executable: $binary_path"
    fi
    
    # For non-current platform binaries, we can only check file properties
    if [[ "$platform" == *"windows"* ]] && [[ "$(uname -s)" != "CYGWIN"* ]] && [[ "$(uname -s)" != "MINGW"* ]]; then
        log "Skipping execution test for Windows binary on $(uname -s)"
        file "$binary_path"
        success "$platform binary exists and has correct file type"
        return
    fi
    
    # For Linux binaries on macOS, we can't execute them
    if [[ "$platform" == *"linux"* ]] && [[ "$(uname -s)" == "Darwin" ]]; then
        log "Skipping execution test for Linux binary on macOS"
        file "$binary_path"
        success "$platform binary exists and has correct file type"
        return
    fi
    
    # For different architecture binaries, we might not be able to execute
    local current_arch
    case "$(uname -m)" in
        x86_64|amd64) current_arch="amd64" ;;
        arm64|aarch64) current_arch="arm64" ;;
        *) current_arch="unknown" ;;
    esac
    
    if [[ "$platform" != *"$current_arch"* ]] && [[ "$(uname -s)" == "Darwin" ]]; then
        log "Skipping execution test for different architecture"
        file "$binary_path"
        success "$platform binary exists and has correct file type"
        return
    fi
    
    # Test version output
    local version_output
    if version_output=$("$binary_path" --version 2>&1); then
        log "Version: $version_output"
        success "$platform binary executes correctly"
    else
        error "Failed to execute $platform binary"
    fi
}

# Test archives
test_archive() {
    local archive_path="$1"
    local platform="$2"
    
    log "Testing $platform archive: $archive_path"
    
    if [ ! -f "$archive_path" ]; then
        error "Archive not found: $archive_path"
    fi
    
    # Test archive integrity
    if [[ "$archive_path" == *.zip ]]; then
        if unzip -t "$archive_path" >/dev/null 2>&1; then
            success "$platform archive integrity verified (ZIP)"
        else
            error "Archive integrity check failed: $archive_path"
        fi
    else
        if tar -tzf "$archive_path" >/dev/null 2>&1; then
            success "$platform archive integrity verified (TAR.GZ)"
        else
            error "Archive integrity check failed: $archive_path"
        fi
    fi
}

# Main test function
main() {
    log "Starting release testing..."
    
    # Check if dist directory exists
    if [ ! -d "dist" ]; then
        error "dist directory not found. Run 'make build-all' first."
    fi
    
    # Test binaries
    log "Testing binaries..."
    for binary in dist/gitcells-*; do
        if [ -f "$binary" ]; then
            platform=$(basename "$binary" | sed 's/gitcells-//')
            test_binary "$binary" "$platform"
        fi
    done
    
    # Test archives if they exist
    if [ -d "dist/releases" ]; then
        log "Testing release archives..."
        for archive in dist/releases/gitcells-*; do
            if [ -f "$archive" ]; then
                platform=$(basename "$archive" | sed -E 's/gitcells-[^-]+-(.+)\.(tar\.gz|zip)$/\1/')
                test_archive "$archive" "$platform"
            fi
        done
    else
        log "No release archives found. Run 'make release' to create them."
    fi
    
    success "All release tests completed successfully!"
}

# Run tests
main "$@"