#!/bin/bash
# Test dataspace isolation in detail

cd "$(dirname "$0")/../.."

echo "=== Dataspace Isolation Test ==="
echo

# Clean start
pkill -f entitydb 2>/dev/null
sleep 1
rm -rf var/test_isolation
mkdir -p var/test_isolation

# Start server
export ENTITYDB_DATA_PATH=var/test_isolation
export ENTITYDB_DATASPACE=true
export ENTITYDB_LOG_LEVEL=debug

./bin/entitydb server > /tmp/isolation_test.log 2>&1 &
SERVER_PID=$!
sleep 3

# Create admin
curl -s -X POST http://localhost:8085/api/v1/users/create \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}' > /dev/null

# Login
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}' | jq -r '.token')

echo "Creating entities in different dataspaces..."
echo

# Create in worca dataspace
echo "Creating in worca dataspace:"
curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "worca-1",
       "tags": ["dataspace:worca", "type:task", "priority:high"],
       "content": "Worca task 1"
     }' | jq -r '.id'

curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "worca-2",
       "tags": ["dataspace:worca", "type:task", "priority:low"],
       "content": "Worca task 2"
     }' | jq -r '.id'

# Create in metrics dataspace
echo
echo "Creating in metrics dataspace:"
curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "metric-1",
       "tags": ["dataspace:metrics", "type:cpu", "host:server1"],
       "content": "cpu.usage=45"
     }' | jq -r '.id'

curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "metric-2",
       "tags": ["dataspace:metrics", "type:memory", "host:server1"],
       "content": "memory.usage=80"
     }' | jq -r '.id'

# Create without dataspace (should go to default)
echo
echo "Creating without dataspace:"
curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "config-1",
       "tags": ["type:config", "name:timeout"],
       "content": "30"
     }' | jq -r '.id'

echo
echo "Checking dataspace indexes created:"
ls -la var/test_isolation/dataspaces/

echo
echo "Testing queries:"
echo

echo "1. Query worca dataspace (should return 2 entities):"
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataspace:worca" \
     -H "Authorization: Bearer $TOKEN" | jq -r 'length'

echo
echo "2. Query metrics dataspace (should return 2 entities):"
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataspace:metrics" \
     -H "Authorization: Bearer $TOKEN" | jq -r 'length'

echo
echo "3. Query default dataspace (should return admin + config):"
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=type:config" \
     -H "Authorization: Bearer $TOKEN" | jq -r 'length'

echo
echo "4. Query all entities (global):"
curl -s -X GET "http://localhost:8085/api/v1/entities/list" \
     -H "Authorization: Bearer $TOKEN" | jq -r 'length'

echo
echo "Checking debug logs for dataspace operations:"
grep -i "dataspace" /tmp/isolation_test.log | grep -E "(Creating|Added|query)" | tail -10

# Cleanup
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo
echo "=== Test Complete ==="