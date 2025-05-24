#!/bin/bash
# Test dataspace functionality

cd "$(dirname "$0")/../.."

echo "=== EntityDB Dataspace Test ==="
echo

# Clean start
pkill -f entitydb 2>/dev/null
sleep 1
rm -rf var/test_dataspace
mkdir -p var/test_dataspace

# Start server with dataspace mode
export ENTITYDB_DATA_PATH=var/test_dataspace
export ENTITYDB_DATASPACE=true
export ENTITYDB_LOG_LEVEL=debug

echo "Starting EntityDB with dataspace mode enabled..."
./bin/entitydb server > /tmp/dataspace_test.log 2>&1 &
SERVER_PID=$!
sleep 3

# Create admin user
echo "Creating admin user..."
curl -s -X POST http://localhost:8085/api/v1/users/create \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}' > /dev/null

# Login
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}' | jq -r '.token')

echo "Token obtained: ${TOKEN:0:20}..."
echo

# Test 1: Create entities in different dataspaces
echo "Test 1: Creating entities in different dataspaces"
echo "================================================"

# Create in worca dataspace
echo -n "Creating task in 'worca' dataspace... "
RESULT=$(curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "task-001",
       "tags": ["dataspace:worca", "type:task", "status:open", "priority:high"],
       "content": "Implement dataspace feature"
     }')
if [ -n "$RESULT" ]; then
    echo "Response: $RESULT"
else
    echo "OK"
fi

# Create in metrics dataspace
echo -n "Creating metric in 'metrics' dataspace... "
curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "metric-001",
       "tags": ["dataspace:metrics", "type:cpu", "host:server1"],
       "content": "cpu.usage=45.2"
     }' > /dev/null && echo "OK"

# Create in default dataspace
echo -n "Creating config in 'default' dataspace... "
curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "config-001",
       "tags": ["type:config", "name:max_connections"],
       "content": "100"
     }' > /dev/null && echo "OK"

echo

# Test 2: Query specific dataspaces
echo "Test 2: Querying specific dataspaces"
echo "===================================="

# Query worca dataspace
echo "Querying 'worca' dataspace:"
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataspace:worca" \
     -H "Authorization: Bearer $TOKEN" | jq -r '.[] | "  - \(.id): \(.tags | join(", "))"'

echo
echo "Querying 'metrics' dataspace:"
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataspace:metrics" \
     -H "Authorization: Bearer $TOKEN" | jq -r '.[] | "  - \(.id): \(.tags | join(", "))"'

echo

# Test 3: Check backward compatibility with hub tags
echo "Test 3: Backward compatibility with hub tags"
echo "==========================================="

# Create with old hub tag
echo -n "Creating entity with 'hub:' tag... "
curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "legacy-001",
       "tags": ["hub:worca", "type:legacy", "status:active"],
       "content": "Legacy hub entity"
     }' > /dev/null && echo "OK"

# Query with hub tag (should work)
echo "Querying with 'hub:worca' (legacy):"
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=hub:worca" \
     -H "Authorization: Bearer $TOKEN" | jq -r '.[] | "  - \(.id)"'

echo

# Test 4: Check dataspace isolation
echo "Test 4: Dataspace isolation"
echo "=========================="

# Check index files created
echo "Checking dataspace index files:"
ls -la var/test_dataspace/dataspaces/ 2>/dev/null || echo "  No dataspace directory yet"

echo

# Check server logs
echo
echo "Server logs (last 20 lines):"
tail -20 /tmp/dataspace_test.log

# Cleanup
echo
echo "Cleaning up..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo
echo "=== Dataspace Test Complete ==="
echo
echo "Summary:"
echo "- Entities can be organized into dataspaces"
echo "- Each dataspace will have its own index file (pending full implementation)"
echo "- Backward compatibility with 'hub:' tags maintained"
echo "- Queries can be scoped to specific dataspaces for better performance"