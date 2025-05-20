#!/bin/bash

# Test concurrent access to EntityDB with granular locking

API_URL="http://localhost:8085/api/v1/entities"
AUTH_HEADER="Authorization: Bearer test-token"

echo "Testing concurrent entity creation..."

# Function to create an entity
create_entity() {
    local id=$1
    local type=$2
    
    curl -s -X POST $API_URL \
        -H "Content-Type: application/json" \
        -H "$AUTH_HEADER" \
        -d "{
            \"tags\": [\"type:$type\", \"test:concurrent\", \"worker:$id\"],
            \"content\": [
                {\"type\": \"title\", \"value\": \"Worker $id Entity\"},
                {\"type\": \"description\", \"value\": \"Created by worker $id at $(date)\"}
            ]
        }" > /dev/null
    
    echo "Worker $id: Created $type entity"
}

# Start multiple workers in parallel
for i in {1..10}; do
    (
        for j in {1..5}; do
            create_entity "$i" "issue"
            sleep 0.1
        done
    ) &
done

# Wait for all workers to complete
wait

echo "All workers completed. Verifying results..."

# Count entities
TOTAL=$(curl -s -H "$AUTH_HEADER" "$API_URL/list" | jq length)
echo "Total entities in database: $TOTAL"

# Count test entities
TEST_COUNT=$(curl -s -H "$AUTH_HEADER" "$API_URL/list?tag=test:concurrent" | jq length)
echo "Test entities created: $TEST_COUNT"

echo "Test completed!"