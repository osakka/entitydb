#!/bin/bash
# Final test for dataspace isolation

cd "$(dirname "$0")/../.."

echo "=== Dataspace Isolation Final Test ==="
echo

# Clean start
pkill -f entitydb 2>/dev/null
sleep 1
rm -rf var/test_final
mkdir -p var/test_final

# Start server
export ENTITYDB_DATA_PATH=var/test_final
export ENTITYDB_DATASPACE=true
export ENTITYDB_LOG_LEVEL=info

./bin/entitydb server > /tmp/final_test.log 2>&1 &
SERVER_PID=$!
sleep 3

# Setup
curl -s -X POST http://localhost:8085/api/v1/users/create \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}' > /dev/null

TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}' | jq -r '.token')

echo "Creating test entities in different dataspaces..."
echo

# Create entities
for i in 1 2 3; do
    curl -s -X POST http://localhost:8085/api/v1/entities/create \
         -H "Authorization: Bearer $TOKEN" \
         -H "Content-Type: application/json" \
         -d "{
           \"id\": \"worcha-$i\",
           \"tags\": [\"dataspace:worcha\", \"type:task\", \"priority:p$i\"],
           \"content\": \"Worcha task $i\"
         }" > /dev/null
    echo -n "W"
done

for i in 1 2; do
    curl -s -X POST http://localhost:8085/api/v1/entities/create \
         -H "Authorization: Bearer $TOKEN" \
         -H "Content-Type: application/json" \
         -d "{
           \"id\": \"metrics-$i\",
           \"tags\": [\"dataspace:metrics\", \"type:metric\", \"name:metric$i\"],
           \"content\": \"Metric value $i\"
         }" > /dev/null
    echo -n "M"
done

for i in 1 2; do
    curl -s -X POST http://localhost:8085/api/v1/entities/create \
         -H "Authorization: Bearer $TOKEN" \
         -H "Content-Type: application/json" \
         -d "{
           \"id\": \"config-$i\",
           \"tags\": [\"type:config\", \"name:setting$i\"],
           \"content\": \"Config value $i\"
         }" > /dev/null
    echo -n "C"
done

echo
echo

# Test queries
echo "Testing dataspace isolation..."
echo

echo "1. Worcha dataspace (should have 3 tasks):"
WORCHA_COUNT=$(curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataspace:worcha" \
     -H "Authorization: Bearer $TOKEN" | jq -r 'length')
echo "   Found: $WORCHA_COUNT entities"
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataspace:worcha" \
     -H "Authorization: Bearer $TOKEN" | jq -r '.[] | "   - " + .id + " (" + (.tags | join(", ")) + ")"'

echo
echo "2. Metrics dataspace (should have 2 metrics):"
METRICS_COUNT=$(curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataspace:metrics" \
     -H "Authorization: Bearer $TOKEN" | jq -r 'length')
echo "   Found: $METRICS_COUNT entities"
curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataspace:metrics" \
     -H "Authorization: Bearer $TOKEN" | jq -r '.[] | "   - " + .id'

echo
echo "3. Default dataspace query by type:config (should have 2 configs + admin):"
CONFIG_COUNT=$(curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=type:config" \
     -H "Authorization: Bearer $TOKEN" | jq -r 'length')
echo "   Found: $CONFIG_COUNT entities"

echo
echo "4. Global query (should have 8 total: 3+2+2+1admin):"
TOTAL_COUNT=$(curl -s -X GET "http://localhost:8085/api/v1/entities/list" \
     -H "Authorization: Bearer $TOKEN" | jq -r 'length')
echo "   Found: $TOTAL_COUNT entities"

echo
echo "5. Cross-filter: worcha + type:task (should have 3):"
CROSS_COUNT=$(curl -s -X GET "http://localhost:8085/api/v1/entities/list?tags=dataspace:worcha,type:task&matchAll=true" \
     -H "Authorization: Bearer $TOKEN" | jq -r 'length')
echo "   Found: $CROSS_COUNT entities"

echo
echo "Dataspace files created:"
ls -la var/test_final/dataspaces/

echo
echo "Test Summary:"
echo "============"
if [ "$WORCHA_COUNT" -eq 3 ] && [ "$METRICS_COUNT" -eq 2 ]; then
    echo "✅ PASS: Dataspace isolation is working correctly!"
else
    echo "❌ FAIL: Dataspace isolation not working"
    echo "   Expected: worcha=3, metrics=2"
    echo "   Actual: worcha=$WORCHA_COUNT, metrics=$METRICS_COUNT"
fi

# Cleanup
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo