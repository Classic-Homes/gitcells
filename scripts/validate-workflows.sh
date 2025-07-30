#!/bin/bash

# Validate GitHub Actions Workflows
# Tests workflow compatibility with our build system

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

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Test build system compatibility
test_build_system() {
    log "Testing build system compatibility..."
    
    # Test that all Makefile targets work
    local targets=("clean" "deps" "build" "build-all" "release" "test-releases")
    
    for target in "${targets[@]}"; do
        if make -n "$target" >/dev/null 2>&1; then
            success "Makefile target '$target' is valid"
        else
            error "Makefile target '$target' is invalid or missing"
        fi
    done
    
    # Test with version variable
    log "Testing with VERSION environment variable..."
    if VERSION="test-1.0.0" make -n build-all >/dev/null 2>&1; then
        success "Build system accepts VERSION environment variable"
    else
        error "Build system doesn't handle VERSION environment variable correctly"
    fi
}

# Test workflow file syntax
test_workflow_syntax() {
    log "Testing workflow file syntax..."
    
    local workflow_files=(".github/workflows/ci.yml" ".github/workflows/release.yml")
    
    for workflow in "${workflow_files[@]}"; do
        if [ ! -f "$workflow" ]; then
            error "Workflow file not found: $workflow"
        fi
        
        # Basic YAML structure check
        if grep -q "^name:" "$workflow" && grep -q "^on:" "$workflow" && grep -q "^jobs:" "$workflow"; then
            success "Workflow file has valid structure: $workflow"
        else
            error "Workflow file has invalid structure: $workflow"
        fi
        
        # Check for required actions versions
        if grep -q "actions/checkout@v4" "$workflow"; then
            success "Uses current checkout action: $workflow"
        else
            warn "May be using outdated checkout action: $workflow"
        fi
        
        if grep -q "actions/setup-go@v5" "$workflow"; then
            success "Uses current Go setup action: $workflow"
        else
            warn "May be using outdated Go setup action: $workflow"
        fi
        
        # Check Go version
        if grep -q "go-version: '1.23'" "$workflow"; then
            success "Uses correct Go version (1.23): $workflow"
        else
            warn "May be using incorrect Go version: $workflow"
        fi
    done
}

# Test package configuration
test_package_config() {
    log "Testing package configuration..."
    
    # Test nfpm config
    if [ -f "build/package/nfpm.yaml" ]; then
        if grep -q "name: gitcells" "build/package/nfpm.yaml"; then
            success "nfpm config has correct package name"
        else
            error "nfpm config has incorrect package name"
        fi
        
        if grep -q "version: \${VERSION}" "build/package/nfpm.yaml"; then
            success "nfpm config uses VERSION variable"
        else
            error "nfpm config doesn't use VERSION variable"
        fi
        
        if grep -q "Classic-Homes/gitcells" "build/package/nfpm.yaml"; then
            success "nfpm config has correct repository URL"
        else
            warn "nfpm config may have incorrect repository URL"
        fi
    else
        error "nfpm configuration file not found"
    fi
}

# Test install script compatibility
test_install_script() {
    log "Testing install script compatibility..."
    
    if [ -f "scripts/install.sh" ]; then
        # Check if install script expects correct archive format
        if grep -q "BINARY_NAME.*version.*platform\|gitcells.*version.*platform" "scripts/install.sh"; then
            success "Install script expects correct archive naming format"
        else
            error "Install script archive naming doesn't match our build system"
        fi
        
        # Check repository reference
        if grep -q "Classic-Homes/gitcells" "scripts/install.sh"; then
            success "Install script references correct repository"
        else
            warn "Install script may reference incorrect repository"
        fi
    else
        error "Install script not found"
    fi
}

# Test release artifacts format
test_release_format() {
    log "Testing release artifacts format..."
    
    # Test that our build creates the expected format
    local test_version="workflow-test"
    export VERSION="$test_version"
    
    log "Building test release with version: $test_version"
    make clean >/dev/null 2>&1
    make release >/dev/null 2>&1
    
    # Check that archives have correct naming
    local expected_files=(
        "dist/releases/gitcells-${test_version}-darwin-amd64.tar.gz"
        "dist/releases/gitcells-${test_version}-darwin-arm64.tar.gz"
        "dist/releases/gitcells-${test_version}-linux-amd64.tar.gz"
        "dist/releases/gitcells-${test_version}-linux-arm64.tar.gz"
        "dist/releases/gitcells-${test_version}-windows-amd64.zip"
    )
    
    for file in "${expected_files[@]}"; do
        if [ -f "$file" ]; then
            success "Release archive created with correct name: $(basename "$file")"
        else
            error "Expected release archive not found: $(basename "$file")"
        fi
    done
    
    # Test that binary has correct version
    local test_binary="dist/gitcells-darwin-arm64"
    if [ -f "$test_binary" ] && [ -x "$test_binary" ]; then
        local version_output
        if version_output=$("$test_binary" --version 2>&1); then
            if echo "$version_output" | grep -q "$test_version"; then
                success "Binary contains correct version: $test_version"
            else
                error "Binary version doesn't match expected: $test_version"
            fi
        else
            warn "Could not test binary version (architecture mismatch?)"
        fi
    fi
    
    # Clean up test artifacts
    make clean >/dev/null 2>&1
}

# Main validation function
main() {
    log "Starting GitHub Actions workflow validation..."
    log "================================================"
    
    test_build_system
    echo ""
    
    test_workflow_syntax
    echo ""
    
    test_package_config
    echo ""
    
    test_install_script
    echo ""
    
    test_release_format
    echo ""
    
    success "All workflow validation tests passed!"
    log "GitHub Actions workflows are ready for use"
}

# Run validation
main "$@"