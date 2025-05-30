#!/bin/bash

# Simple optimization test for EntityDB

set -e

BASE_URL="https://localhost:8085"
ADMIN_USER="admin"
ADMIN_PASS="admin"
NUM_ENTITIES=100

echo "EntityDB Simple Optimization Test"
echo "================================="

# Login
echo "Logging in..."
TOKEN=$(curl -k -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}" | \
    jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "Failed to login"
    exit 1
fi

echo "Login successful"

# Get initial metrics
echo -e "\nInitial metrics:"
curl -k -s -X GET "$BASE_URL/api/v1/system/metrics" \
    -H "Authorization: Bearer $TOKEN" | \
    jq '{
        memory_alloc: .memory.alloc_bytes,
        heap_in_use: .memory.heap_alloc_bytes,
        gc_runs: .performance.gc_runs,
        entity_count: .database.total_entities
    }'

# Create test entities
echo -e "\nCreating $NUM_ENTITIES test entities..."
START=$(date +%s%N)

for i in $(seq 1 $NUM_ENTITIES); do
    # Create entities with repeating tags to test string interning
    curl -k -s -X POST "$BASE_URL/api/v1/entities/create" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "tags": ["type:document", "status:active", "priority:high"],
            "content": "Test entity '$i' content"
        }' > /dev/null
done

END=$(date +%s%N)
DURATION=$(( (END - START) / 1000000 ))
echo "Created $NUM_ENTITIES entities in ${DURATION}ms"
echo "Average: $((DURATION / NUM_ENTITIES))ms per entity"

# Get final metrics
echo -e "\nFinal metrics:"
curl -k -s -X GET "$BASE_URL/api/v1/system/metrics" \
    -H "Authorization: Bearer $TOKEN" | \
    jq '{
        memory_alloc: .memory.alloc_bytes,
        heap_in_use: .memory.heap_alloc_bytes,
        gc_runs: .performance.gc_runs,
        entity_count: .database.total_entities,
        unique_tags: .database.tags_unique
    }'

echo -e "\nTest completed!"