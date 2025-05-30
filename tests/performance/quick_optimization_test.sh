#!/bin/bash

# Quick optimization test for EntityDB

set -e

BASE_URL="https://localhost:8085"
ADMIN_USER="admin"
ADMIN_PASS="admin"
NUM_ENTITIES=50

echo "EntityDB Quick Optimization Test"
echo "================================"

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

# Get before metrics
echo -e "\nBefore creating entities:"
BEFORE=$(curl -k -s -X GET "$BASE_URL/api/v1/system/metrics" \
    -H "Authorization: Bearer $TOKEN")
    
BEFORE_ALLOC=$(echo "$BEFORE" | jq -r '.memory.alloc_bytes')
BEFORE_GC=$(echo "$BEFORE" | jq -r '.performance.gc_runs')
echo "Memory allocated: $(numfmt --to=iec $BEFORE_ALLOC)"
echo "GC runs: $BEFORE_GC"

# Create entities with repeated tags (testing string interning)
echo -e "\nCreating $NUM_ENTITIES entities..."
START=$(date +%s%N)

for i in $(seq 1 $NUM_ENTITIES); do
    # Use repeated tags to test string interning
    curl -k -s -X POST "$BASE_URL/api/v1/entities/create" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "tags": [
                "type:document", 
                "status:active", 
                "priority:high",
                "team:backend",
                "version:1.0"
            ],
            "content": "Test content for optimization"
        }' > /dev/null
done

END=$(date +%s%N)
DURATION=$(( (END - START) / 1000000 ))

# Get after metrics
echo -e "\nAfter creating entities:"
AFTER=$(curl -k -s -X GET "$BASE_URL/api/v1/system/metrics" \
    -H "Authorization: Bearer $TOKEN")
    
AFTER_ALLOC=$(echo "$AFTER" | jq -r '.memory.alloc_bytes')
AFTER_GC=$(echo "$AFTER" | jq -r '.performance.gc_runs')
UNIQUE_TAGS=$(echo "$AFTER" | jq -r '.database.tags_unique')

echo "Memory allocated: $(numfmt --to=iec $AFTER_ALLOC)"
echo "GC runs: $AFTER_GC"
echo "Unique tags: $UNIQUE_TAGS"

# Calculate results
ALLOC_DIFF=$((AFTER_ALLOC - BEFORE_ALLOC))
GC_DIFF=$((AFTER_GC - BEFORE_GC))
ALLOC_PER_ENTITY=$((ALLOC_DIFF / NUM_ENTITIES))

echo -e "\nResults:"
echo "========"
echo "Time: ${DURATION}ms total, $((DURATION / NUM_ENTITIES))ms per entity"
echo "Memory growth: $(numfmt --to=iec $ALLOC_DIFF) total, $(numfmt --to=iec $ALLOC_PER_ENTITY) per entity"
echo "GC runs: $GC_DIFF"

# Quick query test
echo -e "\nQuery performance:"
QUERY_START=$(date +%s%N)
curl -k -s -X GET "$BASE_URL/api/v1/entities/query?tags=type:document&limit=20" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
QUERY_END=$(date +%s%N)
QUERY_TIME=$(( (QUERY_END - QUERY_START) / 1000000 ))
echo "Query 20 entities: ${QUERY_TIME}ms"

echo -e "\nTest completed!"