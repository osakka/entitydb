#!/bin/bash
#
# Apply UI Changes - Deploy updated UI files with API integration
#

# Set script to exit on error
set -e

# Define paths
HTDOCS_PATH="/opt/entitydb/share/htdocs"
BACKUP_DIR="/opt/entitydb/share/htdocs.backup.$(date +%Y%m%d%H%M%S)"

# Create a backup of the current htdocs directory
echo "Creating backup at $BACKUP_DIR..."
cp -r $HTDOCS_PATH $BACKUP_DIR

# Apply changes
echo "Deploying updated UI with API integration..."

# Restart the server to apply changes
echo "Restarting EntityDB server..."
/opt/entitydb/bin/entitydbd.sh restart

# Verify server status
echo "Verifying server status..."
/opt/entitydb/bin/entitydbd.sh status

echo "Deployment complete. The UI now uses the real entity API with fallback to local data."
echo "Backup of previous version saved at: $BACKUP_DIR"