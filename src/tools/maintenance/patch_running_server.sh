#!/bin/bash
# Script to apply the temporal tag patch to a running EntityDB server

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Show a formatted message
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

print_message "$BLUE" "========================================"
print_message "$BLUE" "EntityDB Temporal Tag Patch Utility"
print_message "$BLUE" "========================================"

# Check if server is running
if ! pgrep -f "entitydb" > /dev/null; then
  print_message "$YELLOW" "⚠️ EntityDB server doesn't appear to be running."
  print_message "$YELLOW" "Starting server..."
  
  if [ -f "/opt/entitydb/bin/entitydbd.sh" ]; then
    /opt/entitydb/bin/entitydbd.sh start
    sleep 5
  else
    print_message "$RED" "❌ Server startup script not found."
    exit 1
  fi
fi

# Create a simple test endpoint to apply the patch
print_message "$BLUE" "Creating patch endpoint..."

# Create the patch handler file
cat > /opt/entitydb/src/api/patch_handler.go << 'EOF'
package api

import (
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
	"net/http"
)

// PatchHandler applies runtime patches to the server
type PatchHandler struct {
	repo models.EntityRepository
}

// NewPatchHandler creates a new patch handler
func NewPatchHandler(repo models.EntityRepository) *PatchHandler {
	return &PatchHandler{
		repo: repo,
	}
}

// ApplyTemporalTagPatch applies the temporal tag patch
func (h *PatchHandler) ApplyTemporalTagPatch(w http.ResponseWriter, r *http.Request) {
	logger.Info("Applying temporal tag patch...")
	
	err := binary.PatchListByTag(h.repo)
	if err != nil {
		logger.Error("Failed to apply patch: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to apply patch: " + err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, map[string]string{
		"status": "success",
		"message": "Temporal tag patch applied successfully",
	})
}
EOF

# Recompile the code
print_message "$BLUE" "Compiling patched code..."
cd /opt/entitydb/src
go build -o /opt/entitydb/bin/entitydb_patched

if [ $? -ne 0 ]; then
  print_message "$RED" "❌ Failed to compile patched code."
  exit 1
fi

# Stop the current server
print_message "$BLUE" "Stopping current server..."
if [ -f "/opt/entitydb/bin/entitydbd.sh" ]; then
  /opt/entitydb/bin/entitydbd.sh stop
  sleep 2
else
  pkill -f "entitydb"
  sleep 2
fi

# Start with patched version
print_message "$BLUE" "Starting patched server..."
if [ -f "/opt/entitydb/bin/entitydbd.sh" ]; then
  /opt/entitydb/bin/entitydbd.sh start
  sleep 5
else
  nohup /opt/entitydb/bin/entitydb_patched > /dev/null 2>&1 &
  sleep 5
fi

# Test the patch
print_message "$BLUE" "Testing temporal tag fix..."
cd /opt/entitydb
./improved_temporal_fix.sh

print_message "$GREEN" "✅ Patch process completed!"
print_message "$BLUE" "========================================="