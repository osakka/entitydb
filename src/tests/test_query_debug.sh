#!/bin/bash
# Debug dataset queries

cd "$(dirname "$0")/../.."

# Start fresh server
pkill -f entitydb 2>/dev/null
sleep 1
rm -rf var/debug_query
mkdir -p var/debug_query

export ENTITYDB_DATA_PATH=var/debug_query
export ENTITYDB_DATASPACE=true
export ENTITYDB_LOG_LEVEL=debug

./bin/entitydb server > /tmp/query_debug.log 2>&1 &
SERVER_PID=$!
sleep 3

# Create admin and login
curl -s -X POST http://localhost:8085/api/v1/users/create \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}' > /dev/null

TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Create one entity in worca
curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "worca-task",
       "tags": ["dataset:worca", "type:task"],
       "content": "A worca task"
     }' > /dev/null

# Create one entity in metrics
curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "metrics-cpu",
       "tags": ["dataset:metrics", "type:cpu"],
       "content": "cpu usage"
     }' > /dev/null

echo "Testing dataset query..."
echo

# Query worca dataset
echo "Querying dataset:worca..."
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataset:worca" \
     -H "Authorization: Bearer $TOKEN" | jq -r '.[] | .id'

echo
echo "Relevant logs:"
grep -E "(Dataset query|Found.*entities.*dataset)" /tmp/query_debug.log | tail -10

kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null