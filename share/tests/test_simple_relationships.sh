#!/bin/bash

# Start the server in foreground
echo "Starting server..."
/opt/entitydb/bin/entitydb &
SERVER_PID=$!

sleep 3

# Try to create entities without auth (will fail but let's see)
echo -e "\n=== Creating entities directly ==="
ENTITY1_RESP=$(curl -s -X POST "http://localhost:8085/api/v1/entities/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token_admin_12345" \
  -d '{
    "tags": ["type:task", "title:Task 1"]
  }')

echo "Response: $ENTITY1_RESP"

# Try the test endpoint
echo -e "\n=== Using test endpoint ==="
TEST_RESP=$(curl -s "http://localhost:8085/api/v1/test")
echo "Test response: $TEST_RESP"

echo -e "\n=== Checking logs ==="
tail -30 /opt/entitydb/var/log/entitydb.log 2>/dev/null || echo "No log file"

echo -e "\n=== Stopping server ==="
kill $SERVER_PID 2>/dev/null