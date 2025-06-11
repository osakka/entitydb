#!/bin/bash
# Debug dataset functionality

cd "$(dirname "$0")/../.."

echo "=== Dataset Debug Test ==="
echo

# Clean start
pkill -f entitydb 2>/dev/null
sleep 1
rm -rf var/debug_dataset
mkdir -p var/debug_dataset
# Create initial empty file so repository can start
touch var/debug_dataset/entities.ebf

# Start server with dataset mode
export ENTITYDB_DATA_PATH=var/debug_dataset
export ENTITYDB_DATASPACE=true
export ENTITYDB_LOG_LEVEL=debug

echo "Starting EntityDB with dataset mode..."
./bin/entitydb server &
SERVER_PID=$!
sleep 3

# Create admin user
echo "Creating admin user..."
RESULT=$(curl -s -X POST http://localhost:8085/api/v1/users/create \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}')
echo "Admin creation result: $RESULT"

# Login
echo "Logging in..."
LOGIN_RESULT=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}')
echo "Login result: $LOGIN_RESULT"

TOKEN=$(echo "$LOGIN_RESULT" | jq -r '.token')
echo "Token: ${TOKEN:0:20}..."
echo

# Create a simple entity
echo "Creating test entity in worca dataset..."
CREATE_RESULT=$(curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "test-123",
       "tags": ["dataset:worca", "type:test"],
       "content": "Test entity"
     }')
echo "Create result: $CREATE_RESULT"
echo

# Try different queries
echo "Query 1: List all entities..."
curl -s -X GET "http://localhost:8085/api/v1/entities/list" \
     -H "Authorization: Bearer $TOKEN" | jq '.'

echo
echo "Query 2: List with dataset:worca tag..."
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataset:worca" \
     -H "Authorization: Bearer $TOKEN" | jq '.'

echo
echo "Query 3: Get specific entity..."
curl -s -X GET "http://localhost:8085/api/v1/entities/get?id=test-123" \
     -H "Authorization: Bearer $TOKEN" | jq '.'

echo
echo "Checking dataset files..."
ls -la var/debug_dataset/
ls -la var/debug_dataset/datasets/ 2>/dev/null || echo "No datasets directory"

echo
echo "Checking server logs..."
tail -20 /tmp/dataset_test.log 2>/dev/null || echo "No log file"

# Cleanup
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo
echo "=== Debug Test Complete ==="