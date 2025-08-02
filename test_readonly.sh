#!/bin/bash

echo "Testing GitCells TUI with read-only directory..."
echo "1. Try to setup in ~/gitcells_readonly_test (read-only)"
echo "2. Navigate to directory step and enter the path"
echo "3. Should see permission error"
echo ""
echo "Press any key to start test..."
read -n 1

# Run the TUI
./dist/gitcells tui

# Cleanup
chmod 755 ~/gitcells_readonly_test
rm -rf ~/gitcells_readonly_test
echo "Test directory cleaned up."