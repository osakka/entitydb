#!/bin/bash
# migrate_scripts_to_entity_api.sh
# Helper script to migrate bash scripts from legacy API to entity-based API
# This script scans a directory for bash scripts and suggests replacements for API calls

set -e

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <directory_to_scan>"
  echo "Example: $0 /path/to/scripts"
  exit 1
fi

SCAN_DIR="$1"

if [ ! -d "$SCAN_DIR" ]; then
  echo "Error: '$SCAN_DIR' is not a directory"
  exit 1
fi

echo "Scanning '$SCAN_DIR' for scripts that use the legacy API..."

# Define the migration mapping
declare -A API_REPLACEMENTS=(
  # Workspace API replacements
  ["/api/v1/workspaces/list"]="/api/v1/direct/workspace/list"
  ["/api/v1/workspaces/get"]="/api/v1/direct/workspace/get"
  ["/api/v1/workspaces/create"]="/api/v1/direct/workspace/create"
  ["/api/v1/workspaces/update"]="/api/v1/direct/workspace/create"
  ["/workspace/list"]="/api/v1/direct/workspace/list"
  ["/workspace/get"]="/api/v1/direct/workspace/get"
  
  # Issue API replacements
  ["/api/v1/issues/list"]="/api/v1/direct/issue/list"
  ["/api/v1/issues/get"]="/api/v1/direct/issue/get"
  ["/api/v1/issues/create"]="/api/v1/direct/issue/create"
  ["/api/v1/issues/update"]="/api/v1/direct/issue/create"
  ["/api/v1/issues/assign"]="/api/v1/direct/issue/assign"
  ["/api/v1/issues/start"]="/api/v1/direct/issue/status"
  ["/api/v1/issues/progress"]="/api/v1/direct/issue/status"
  ["/api/v1/issues/complete"]="/api/v1/direct/issue/status"
  ["/api/v1/issues/block"]="/api/v1/direct/issue/status"
  ["/api/v1/issues/unblock"]="/api/v1/direct/issue/status"
  ["/issue/list"]="/api/v1/direct/issue/list"
  ["/issue/get"]="/api/v1/direct/issue/get"
  ["/issue/create"]="/api/v1/direct/issue/create"
)

# Scan for script files
echo "Found the following scripts with API calls that need migration:"
echo "-----------------------------------------------------------------"

found_scripts=false

# Loop through all shell scripts
for script in $(find "$SCAN_DIR" -type f -name "*.sh" | sort); do
  needs_migration=false
  
  # Check if script contains any of the legacy API endpoints
  for old_api in "${!API_REPLACEMENTS[@]}"; do
    if grep -q "$old_api" "$script"; then
      needs_migration=true
      break
    fi
  done
  
  if $needs_migration; then
    found_scripts=true
    echo "* $script"
    
    # Show specific replacements needed
    echo "  Needed replacements:"
    for old_api in "${!API_REPLACEMENTS[@]}"; do
      if grep -q "$old_api" "$script"; then
        echo "    - $old_api -> ${API_REPLACEMENTS[$old_api]}"
      fi
    done
    echo ""
  fi
done

if ! $found_scripts; then
  echo "No scripts found that require migration."
fi

echo ""
echo "Migration guidance:"
echo "-----------------------------------------------------------------"
echo "1. Replace API endpoints according to the suggestions above"
echo ""
echo "2. Replace request bodies as needed:"
echo "   - For issue status updates, use the new status format:"
echo "     Old: POST /api/v1/issues/complete with issue_id"
echo "     New: POST /api/v1/direct/issue/status with {issue_id, status: 'completed'}"
echo ""
echo "3. See /opt/entitydb/docs/entity_api_usage_examples.md for complete examples"
echo ""
echo "4. For automated replacements, you can use:"
echo "   sed -i 's|/api/v1/workspaces/list|/api/v1/direct/workspace/list|g' script.sh"
echo ""
echo "For detailed migration instructions, see:"
echo "/opt/entitydb/docs/IMPORTANT_ARCHITECTURE_TRANSITION.md"
echo "/opt/entitydb/docs/entity_architecture_migration.md"
echo "/opt/entitydb/docs/entity_api_usage_examples.md"