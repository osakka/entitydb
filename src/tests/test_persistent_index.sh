#!/bin/bash

# Test persistent tag index functionality

set -e

echo "=== Testing Persistent Tag Index ==="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

# Test directory
TEST_DIR="var/test_persistent_index"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

# Start server with test data directory
echo -e "${YELLOW}Starting EntityDB with test data directory...${NC}"
../bin/entitydb --data="$TEST_DIR" --port=8186 &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Function to check if server is running
check_server() {
    curl -s -k "http://localhost:8186/health" >/dev/null 2>&1
}

if ! check_server; then
    echo -e "${RED}Server failed to start${NC}"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

echo -e "${GREEN}Server started successfully${NC}"

# Login
echo -e "${YELLOW}Logging in...${NC}"
TOKEN=$(curl -s -X POST http://localhost:8186/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}' | jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo -e "${RED}Failed to login${NC}"
    kill $SERVER_PID
    exit 1
fi

echo -e "${GREEN}Login successful${NC}"

# Create test entities
echo -e "${YELLOW}Creating test entities...${NC}"
for i in {1..10}; do
    curl -s -X POST http://localhost:8186/api/v1/entities/create \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"tags\": [\"test:persistent\", \"index:test\", \"item:$i\"],
            \"content\": \"Test entity $i for persistent index\"
        }" >/dev/null
done

echo -e "${GREEN}Created 10 test entities${NC}"

# Query to verify entities exist
COUNT=$(curl -s -X GET "http://localhost:8186/api/v1/entities/list?tags=test:persistent&matchAll=true" \
    -H "Authorization: Bearer $TOKEN" | jq '. | length')

echo -e "${GREEN}Found $COUNT entities with tag 'test:persistent'${NC}"

# Stop server gracefully to trigger index save
echo -e "${YELLOW}Stopping server to save index...${NC}"
kill -TERM $SERVER_PID
wait $SERVER_PID 2>/dev/null || true

# Check if index file was created
INDEX_FILE=$(find "$TEST_DIR" -name "*.idx" -type f | head -1)
if [ -z "$INDEX_FILE" ]; then
    echo -e "${RED}No index file created!${NC}"
    exit 1
fi

echo -e "${GREEN}Index file created: $INDEX_FILE${NC}"
ls -la "$INDEX_FILE"

# Start server again
echo -e "${YELLOW}Starting server again to test index loading...${NC}"
../bin/entitydb --data="$TEST_DIR" --port=8186 &
SERVER_PID=$!

# Wait for server to start
sleep 3

if ! check_server; then
    echo -e "${RED}Server failed to restart${NC}"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

# Login again
TOKEN=$(curl -s -X POST http://localhost:8186/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Query entities without creating new ones
echo -e "${YELLOW}Querying entities after restart...${NC}"
COUNT2=$(curl -s -X GET "http://localhost:8186/api/v1/entities/list?tags=test:persistent&matchAll=true" \
    -H "Authorization: Bearer $TOKEN" | jq '. | length')

echo -e "${GREEN}Found $COUNT2 entities after restart${NC}"

# Verify counts match
if [ "$COUNT" -eq "$COUNT2" ]; then
    echo -e "${GREEN}SUCCESS: Persistent index working correctly!${NC}"
    echo -e "${GREEN}Index was loaded from disk on startup${NC}"
else
    echo -e "${RED}FAILURE: Entity count mismatch (before: $COUNT, after: $COUNT2)${NC}"
    kill $SERVER_PID
    exit 1
fi

# Check server logs for index loading
echo -e "${YELLOW}Checking server logs for index loading...${NC}"
if grep -q "Loading tag index from persistent storage" "$TEST_DIR/../entitydb.log" 2>/dev/null || \
   grep -q "Loaded.*tags from persistent index" var/entitydb.log 2>/dev/null; then
    echo -e "${GREEN}Server logs confirm index was loaded from disk${NC}"
fi

# Cleanup
echo -e "${YELLOW}Cleaning up...${NC}"
kill $SERVER_PID 2>/dev/null || true
rm -rf "$TEST_DIR"

echo -e "${GREEN}=== Test Complete ===${NC}"