#!/bin/bash
# Debug dataspace queries

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

# Create one entity in worcha
curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "worcha-task",
       "tags": ["dataspace:worcha", "type:task"],
       "content": "A worcha task"
     }' > /dev/null

# Create one entity in metrics
curl -s -X POST http://localhost:8085/api/v1/entities/create \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "id": "metrics-cpu",
       "tags": ["dataspace:metrics", "type:cpu"],
       "content": "cpu usage"
     }' > /dev/null

echo "Testing dataspace query..."
echo

# Query worcha dataspace
echo "Querying dataspace:worcha..."
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataspace:worcha" \
     -H "Authorization: Bearer $TOKEN" | jq -r '.[] | .id'

echo
echo "Relevant logs:"
grep -E "(Dataspace query|Found.*entities.*dataspace)" /tmp/query_debug.log | tail -10

kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null