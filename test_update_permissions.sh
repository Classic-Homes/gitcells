#!/bin/bash

echo "Testing GitCells update permission handling..."
echo "This script will demonstrate the new permission elevation feature"
echo ""

# First, let's check the current version
echo "Current version:"
./dist/gitcells version

echo ""
echo "To test the permission handling:"
echo "1. Install gitcells to a system directory (requires sudo):"
echo "   sudo cp ./dist/gitcells /usr/local/bin/gitcells"
echo ""
echo "2. Then run the update command as a regular user:"
echo "   gitcells update"
echo ""
echo "The update will fail with permission denied, and then offer to run with sudo."
echo ""
echo "The new implementation provides:"
echo "- Automatic detection of permission errors"
echo "- Interactive prompt to run with sudo (Unix/Mac)"
echo "- PowerShell elevation prompt (Windows)"
echo "- Manual fallback instructions if automatic elevation fails"