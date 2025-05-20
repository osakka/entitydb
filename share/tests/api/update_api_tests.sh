#!/bin/bash
# Update API test script with improved version

# Source and destination paths
SRC="/opt/entitydb/src/tools/run_api_tests.sh.improved"
DEST="/opt/entitydb/src/tools/run_api_tests.sh"

# Ensure the improved script is executable
chmod +x "$SRC"

# Create a backup of the original script
cp -f "$DEST" "${DEST}.bak"

# Copy the improved script to the destination
cp -f "$SRC" "$DEST"

# Make the destination script executable
chmod +x "$DEST"

echo "Updated API test script with improved version."
echo "Original script backed up to ${DEST}.bak"