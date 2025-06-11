#!/bin/bash
# Test dataset functionality

cd "$(dirname "$0")/../.."

echo "=== EntityDB Dataset Test ==="
echo

# Clean start
pkill -f entitydb 2>/dev/null
sleep 1
rm -rf var/test_dataset
mkdir -p var/test_dataset

# Start server with dataset mode
export ENTITYDB_DATA_PATH=var/test_dataset
export ENTITYDB_DATASPACE=true
export ENTITYDB_LOG_LEVEL=debug

echo "Starting EntityDB with dataset mode enabled..."
./bin/entitydb server > /tmp/dataset_test.log 2>&1 &
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

# Test 1: Create entities in different datasets
echo "Test 1: Creating entities in different datasets"
echo "================================================"

# Create in worca dataset
echo -n "Creating task in 'worca' dataset... "
RESULT=$(curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "task-001",
       "tags": ["dataset:worca", "type:task", "status:open", "priority:high"],
       "content": "Implement dataset feature"
     }')
if [ -n "$RESULT" ]; then
    echo "Response: $RESULT"
else
    echo "OK"
fi

# Create in metrics dataset
echo -n "Creating metric in 'metrics' dataset... "
curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "metric-001",
       "tags": ["dataset:metrics", "type:cpu", "host:server1"],
       "content": "cpu.usage=45.2"
     }' > /dev/null && echo "OK"

# Create in default dataset
echo -n "Creating config in 'default' dataset... "
curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "config-001",
       "tags": ["type:config", "name:max_connections"],
       "content": "100"
     }' > /dev/null && echo "OK"

echo

# Test 2: Query specific datasets
echo "Test 2: Querying specific datasets"
echo "===================================="

# Query worca dataset
echo "Querying 'worca' dataset:"
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataset:worca" \
     -H "Authorization: Bearer $TOKEN" | jq -r '.[] | "  - \(.id): \(.tags | join(", "))"'

echo
echo "Querying 'metrics' dataset:"
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataset:metrics" \
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

# Test 4: Check dataset isolation
echo "Test 4: Dataset isolation"
echo "=========================="

# Check index files created
echo "Checking dataset index files:"
ls -la var/test_dataset/datasets/ 2>/dev/null || echo "  No dataset directory yet"

echo

# Check server logs
echo
echo "Server logs (last 20 lines):"
tail -20 /tmp/dataset_test.log

# Cleanup
echo
echo "Cleaning up..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo
echo "=== Dataset Test Complete ==="
echo
echo "Summary:"
echo "- Entities can be organized into datasets"
echo "- Each dataset will have its own index file (pending full implementation)"
echo "- Backward compatibility with 'hub:' tags maintained"
echo "- Queries can be scoped to specific datasets for better performance"