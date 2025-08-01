#!/bin/bash

# Version management script for GitCells
# This script provides utilities for managing the application version consistently

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VERSION_FILE="$PROJECT_ROOT/VERSION"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print usage information
usage() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  get                     Get current version"
    echo "  set <version>          Set new version"
    echo "  bump <major|minor|patch> Bump version component"
    echo "  validate               Validate version format"
    echo "  sync                   Sync version across all files"
    echo "  check                  Check version consistency"
    echo ""
    echo "Examples:"
    echo "  $0 get"
    echo "  $0 set 1.2.3"
    echo "  $0 bump minor"
    echo "  $0 validate"
    echo "  $0 sync"
    echo "  $0 check"
}

# Get current version
get_version() {
    if [[ -f "$VERSION_FILE" ]]; then
        cat "$VERSION_FILE"
    else
        echo -e "${RED}ERROR: VERSION file not found at $VERSION_FILE${NC}" >&2
        exit 1
    fi
}

# Validate version format (semantic versioning)
validate_version() {
    local version="$1"
    if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$ ]]; then
        echo -e "${RED}ERROR: Invalid version format '$version'. Must be semantic version (e.g., 1.2.3)${NC}" >&2
        return 1
    fi
    return 0
}

# Set new version
set_version() {
    local new_version="$1"
    
    if ! validate_version "$new_version"; then
        exit 1
    fi
    
    echo "$new_version" > "$VERSION_FILE"
    echo -e "${GREEN}✅ Version set to $new_version${NC}"
    
    # Automatically sync to other files
    sync_version
}

# Bump version component
bump_version() {
    local component="$1"
    local current_version
    current_version=$(get_version)
    
    # Parse current version
    if [[ "$current_version" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)(-.*)?(\+.*)?$ ]]; then
        local major="${BASH_REMATCH[1]}"
        local minor="${BASH_REMATCH[2]}"
        local patch="${BASH_REMATCH[3]}"
        local prerelease="${BASH_REMATCH[4]}"
        local build="${BASH_REMATCH[5]}"
    else
        echo -e "${RED}ERROR: Cannot parse current version '$current_version'${NC}" >&2
        exit 1
    fi
    
    # Bump the specified component
    case "$component" in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            prerelease=""
            build=""
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            prerelease=""
            build=""
            ;;
        patch)
            patch=$((patch + 1))
            prerelease=""
            build=""
            ;;
        *)
            echo -e "${RED}ERROR: Invalid component '$component'. Must be major, minor, or patch${NC}" >&2
            exit 1
            ;;
    esac
    
    local new_version="$major.$minor.$patch$prerelease$build"
    echo -e "${BLUE}Bumping $component: $current_version → $new_version${NC}"
    set_version "$new_version"
}

# Sync version to other files
sync_version() {
    local version
    version=$(get_version)
    
    echo -e "${BLUE}🔄 Syncing version $version to all files...${NC}"
    
    # Update go.mod if it exists and has a different version
    if [[ -f "$PROJECT_ROOT/go.mod" ]]; then
        echo -e "${YELLOW}  📦 go.mod - version managed by Go modules${NC}"
    fi
    
    # Update internal/constants/version.go default value
    if [[ -f "$PROJECT_ROOT/internal/constants/version.go" ]]; then
        sed -i.bak "s/Version = \"[^\"]*\"/Version = \"$version\"/" "$PROJECT_ROOT/internal/constants/version.go"
        rm -f "$PROJECT_ROOT/internal/constants/version.go.bak"
        echo -e "${GREEN}  ✅ internal/constants/version.go${NC}"
    fi
    
    echo -e "${GREEN}✅ Version sync complete${NC}"
}

# Check version consistency across files
check_version() {
    local version
    version=$(get_version)
    local inconsistent=false
    
    echo -e "${BLUE}🔍 Checking version consistency...${NC}"
    echo -e "${BLUE}Current version: $version${NC}"
    echo ""
    
    # Check internal/constants/version.go
    if [[ -f "$PROJECT_ROOT/internal/constants/version.go" ]]; then
        local go_version
        go_version=$(grep -o 'Version = "[^"]*"' "$PROJECT_ROOT/internal/constants/version.go" | sed 's/Version = "\([^"]*\)"/\1/')
        if [[ "$go_version" == "$version" ]] || [[ "$go_version" == "dev" ]]; then
            echo -e "${GREEN}  ✅ internal/constants/version.go: $go_version${NC}"
        else
            echo -e "${RED}  ❌ internal/constants/version.go: $go_version (expected: $version)${NC}"
            inconsistent=true
        fi
    fi
    
    # Check Makefile (it should use the VERSION file now)
    if [[ -f "$PROJECT_ROOT/Makefile" ]]; then
        if grep -q "shell cat VERSION" "$PROJECT_ROOT/Makefile"; then
            echo -e "${GREEN}  ✅ Makefile: uses VERSION file${NC}"
        else
            echo -e "${YELLOW}  ⚠️  Makefile: may not be using VERSION file${NC}"
        fi
    fi
    
    # Check GitHub Actions
    if [[ -f "$PROJECT_ROOT/.github/workflows/release.yml" ]]; then
        echo -e "${GREEN}  ✅ GitHub Actions: uses git tags${NC}"
    fi
    
    # Check Dockerfile
    if [[ -f "$PROJECT_ROOT/Dockerfile" ]]; then
        echo -e "${GREEN}  ✅ Dockerfile: uses build args${NC}"
    fi
    
    echo ""
    if [[ "$inconsistent" == true ]]; then
        echo -e "${RED}❌ Version inconsistencies found. Run '$0 sync' to fix.${NC}"
        return 1
    else
        echo -e "${GREEN}✅ All versions are consistent${NC}"
        return 0
    fi
}

# Main script logic
main() {
    if [[ $# -eq 0 ]]; then
        usage
        exit 1
    fi
    
    case "$1" in
        get)
            get_version
            ;;
        set)
            if [[ $# -ne 2 ]]; then
                echo -e "${RED}ERROR: 'set' command requires a version argument${NC}" >&2
                usage
                exit 1
            fi
            set_version "$2"
            ;;
        bump)
            if [[ $# -ne 2 ]]; then
                echo -e "${RED}ERROR: 'bump' command requires a component argument (major|minor|patch)${NC}" >&2
                usage
                exit 1
            fi
            bump_version "$2"
            ;;
        validate)
            if [[ $# -eq 1 ]]; then
                # Validate current version
                version=$(get_version)
            else
                version="$2"
            fi
            if validate_version "$version"; then
                echo -e "${GREEN}✅ Version '$version' is valid${NC}"
            else
                exit 1
            fi
            ;;
        sync)
            sync_version
            ;;
        check)
            check_version
            ;;
        -h|--help|help)
            usage
            ;;
        *)
            echo -e "${RED}ERROR: Unknown command '$1'${NC}" >&2
            usage
            exit 1
            ;;
    esac
}

main "$@"