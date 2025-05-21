#!/bin/bash
# Apply temporal fixes to EntityDB

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Exit on any error
set -e

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB Temporal Features Fix${NC}"
echo -e "${BLUE}========================================${NC}"

# First, stop the running server
echo -e "${BLUE}Stopping EntityDB server...${NC}"
cd /opt/entitydb && ./bin/entitydbd.sh stop

# Edit main.go to add our patchTemporalEndpoints function call
echo -e "${BLUE}Patching main.go to use fixed temporal endpoints...${NC}"

grep -n "apiRouter.HandleFunc(\"/entities/as-of\"" /opt/entitydb/src/main.go
line_number=$(grep -n "apiRouter.HandleFunc(\"/entities/as-of\"" /opt/entitydb/src/main.go | cut -d: -f1)
if [ -z "$line_number" ]; then
  echo -e "${RED}Couldn't find line to patch in main.go${NC}"
  exit 1
fi

# Create a temporary file with the patched content
awk -v line="$line_number" 'NR==line {
  print "// Patched temporal endpoints with fixed implementations"
  print "\t// Original temporal endpoints commented out"
  print "\t/*"
  print "\tapiRouter.HandleFunc(\"/entities/as-of\", entityHandlerRBAC.GetEntityAsOfWithRBAC()).Methods(\"GET\")"
  print "\tapiRouter.HandleFunc(\"/entities/history\", entityHandlerRBAC.GetEntityHistoryWithRBAC()).Methods(\"GET\")"
  print "\tapiRouter.HandleFunc(\"/entities/changes\", entityHandlerRBAC.GetRecentChangesWithRBAC()).Methods(\"GET\")"
  print "\tapiRouter.HandleFunc(\"/entities/diff\", entityHandlerRBAC.GetEntityDiffWithRBAC()).Methods(\"GET\")"
  print "\t*/"
  print "\t// Fixed temporal endpoints"
  print "\tapiRouter.HandleFunc(\"/entities/as-of\", entityHandlerRBAC.GetEntityAsOfWithRBACFixed()).Methods(\"GET\")"
  print "\tapiRouter.HandleFunc(\"/entities/history\", entityHandlerRBAC.GetEntityHistoryWithRBACFixed()).Methods(\"GET\")"
  print "\tapiRouter.HandleFunc(\"/entities/changes\", entityHandlerRBAC.GetRecentChangesWithRBACFixed()).Methods(\"GET\")"
  print "\tapiRouter.HandleFunc(\"/entities/diff\", entityHandlerRBAC.GetEntityDiffWithRBACFixed()).Methods(\"GET\")"
  next
}
NR >= line && NR <= line+3 { next } # Skip the four original lines
{ print }' /opt/entitydb/src/main.go > /opt/entitydb/src/main.go.patched

# Verify the patch
if grep -q "GetEntityAsOfWithRBACFixed" /opt/entitydb/src/main.go.patched; then
  mv /opt/entitydb/src/main.go.patched /opt/entitydb/src/main.go
  echo -e "${GREEN}✅ main.go patched successfully${NC}"
else
  echo -e "${RED}❌ Patch creation failed${NC}"
  exit 1
fi

# Build the server
echo -e "${BLUE}Building EntityDB server with temporal fixes...${NC}"
cd /opt/entitydb/src && make

# Start the server
echo -e "${BLUE}Starting EntityDB server...${NC}"
cd /opt/entitydb && ./bin/entitydbd.sh start

# Wait for server to start
echo -e "${BLUE}Waiting for server to initialize...${NC}"
sleep 5

# Make test script executable
echo -e "${BLUE}Making test script executable...${NC}"
chmod +x /opt/entitydb/test_temporal_fix.sh

# Run the test
echo -e "${BLUE}Running temporal fix test...${NC}"
cd /opt/entitydb && ./test_temporal_fix.sh

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✅ Temporal fix applied and tested${NC}"
echo -e "${BLUE}========================================${NC}"