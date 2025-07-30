#!/bin/bash

# SheetSync Uninstallation Script
# This script removes SheetSync from your system

set -e

# Configuration
BINARY_NAME="sheetsync"
INSTALL_PATHS=(
    "/usr/local/bin/${BINARY_NAME}"
    "/usr/bin/${BINARY_NAME}"
    "$HOME/.local/bin/${BINARY_NAME}"
    "$HOME/bin/${BINARY_NAME}"
)
CONFIG_DIRS=(
    "$HOME/.config/sheetsync"
    "$HOME/.sheetsync"
)

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

# Find installed binary
find_binary() {
    local found_paths=()
    
    for path in "${INSTALL_PATHS[@]}"; do
        if [ -f "$path" ]; then
            found_paths+=("$path")
        fi
    done
    
    # Also check PATH
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local path_binary
        path_binary=$(command -v "$BINARY_NAME")
        # Add to found_paths if not already there
        local already_found=false
        for existing in "${found_paths[@]}"; do
            if [ "$existing" = "$path_binary" ]; then
                already_found=true
                break
            fi
        done
        if [ "$already_found" = false ]; then
            found_paths+=("$path_binary")
        fi
    fi
    
    printf '%s\n' "${found_paths[@]}"
}

# Remove binary
remove_binary() {
    local binary_path="$1"
    local needs_sudo=false
    
    # Check if we need sudo
    local dir
    dir=$(dirname "$binary_path")
    if [ ! -w "$dir" ] && [ "$(id -u)" -ne 0 ]; then
        needs_sudo=true
    fi
    
    log "Removing binary: $binary_path"
    
    if [ "$needs_sudo" = true ]; then
        if command -v sudo >/dev/null 2>&1; then
            sudo rm -f "$binary_path" || warn "Failed to remove $binary_path"
        else
            warn "Cannot remove $binary_path (no write permission and sudo not available)"
            return 1
        fi
    else
        rm -f "$binary_path" || warn "Failed to remove $binary_path"
    fi
    
    if [ ! -f "$binary_path" ]; then
        success "Removed: $binary_path"
        return 0
    else
        warn "Failed to remove: $binary_path"
        return 1
    fi
}

# Remove configuration
remove_config() {
    local removed=false
    
    for config_dir in "${CONFIG_DIRS[@]}"; do
        if [ -d "$config_dir" ]; then
            log "Removing configuration directory: $config_dir"
            rm -rf "$config_dir"
            success "Removed configuration: $config_dir"
            removed=true
        fi
    done
    
    if [ "$removed" = false ]; then
        log "No configuration directories found"
    fi
}

# Show help
show_help() {
    echo "SheetSync Uninstallation Script"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --keep-config    Keep configuration files"
    echo "  --dry-run        Show what would be removed without actually removing"
    echo "  -h, --help       Show this help message"
    echo ""
    echo "This script will remove:"
    echo "  - SheetSync binary from common installation locations"
    echo "  - Configuration files (unless --keep-config is specified)"
    echo ""
}

# Dry run mode
dry_run() {
    echo "DRY RUN: The following items would be removed:"
    echo ""
    
    # Find binaries
    local binaries
    mapfile -t binaries < <(find_binary)
    
    if [ ${#binaries[@]} -gt 0 ]; then
        echo "Binaries:"
        for binary in "${binaries[@]}"; do
            echo "  - $binary"
        done
    else
        echo "  - No binaries found"
    fi
    
    echo ""
    
    # Check config directories
    echo "Configuration directories:"
    local found_config=false
    for config_dir in "${CONFIG_DIRS[@]}"; do
        if [ -d "$config_dir" ]; then
            echo "  - $config_dir"
            found_config=true
        fi
    done
    
    if [ "$found_config" = false ]; then
        echo "  - No configuration directories found"
    fi
    
    echo ""
    echo "To proceed with removal, run without --dry-run"
}

# Main uninstallation function
main() {
    local keep_config=false
    local dry_run_mode=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --keep-config)
                keep_config=true
                shift
                ;;
            --dry-run)
                dry_run_mode=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                error "Unknown option: $1. Use --help for usage information."
                ;;
        esac
    done
    
    log "Starting SheetSync uninstallation..."
    
    if [ "$dry_run_mode" = true ]; then
        dry_run
        exit 0
    fi
    
    # Find and remove binaries
    local binaries removed_count=0
    mapfile -t binaries < <(find_binary)
    
    if [ ${#binaries[@]} -eq 0 ]; then
        log "No SheetSync binaries found"
    else
        log "Found ${#binaries[@]} SheetSync installation(s)"
        for binary in "${binaries[@]}"; do
            if remove_binary "$binary"; then
                ((removed_count++))
            fi
        done
    fi
    
    # Remove configuration if requested
    if [ "$keep_config" = false ]; then
        remove_config
    else
        log "Keeping configuration files (--keep-config specified)"
    fi
    
    # Summary
    echo ""
    if [ $removed_count -gt 0 ]; then
        success "SheetSync uninstallation completed"
        echo "Removed $removed_count binary file(s)"
        
        # Check if still in PATH
        if command -v "$BINARY_NAME" >/dev/null 2>&1; then
            warn "SheetSync is still available in PATH. You may need to restart your shell or update your PATH."
        fi
    else
        warn "No SheetSync installations were removed"
        echo "SheetSync may not have been installed or may be in a location not checked by this script."
    fi
    
    if [ "$keep_config" = false ]; then
        echo "Configuration files removed"
    fi
    
    echo ""
    echo "Thank you for using SheetSync!"
}

# Run main function
main "$@"