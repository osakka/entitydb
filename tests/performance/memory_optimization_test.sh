#!/bin/bash

# Memory optimization test for EntityDB

set -e

BASE_URL="https://localhost:8085"
ADMIN_USER="admin"
ADMIN_PASS="admin"
NUM_ENTITIES=1000
BATCH_SIZE=100

echo "EntityDB Memory Optimization Test"
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
INITIAL_METRICS=$(curl -k -s -X GET "$BASE_URL/api/v1/system/metrics" \
    -H "Authorization: Bearer $TOKEN")
    
echo "$INITIAL_METRICS" | jq '{
    memory_alloc: .memory.alloc_bytes,
    heap_in_use: .memory.heap_alloc_bytes,
    heap_sys: .memory.heap_sys_bytes,
    gc_runs: .performance.gc_runs,
    goroutines: .performance.goroutines,
    entity_count: .database.total_entities
}'

INITIAL_ALLOC=$(echo "$INITIAL_METRICS" | jq -r '.memory.alloc_bytes')
INITIAL_GC=$(echo "$INITIAL_METRICS" | jq -r '.performance.gc_runs')

# Create test entities with repeated tags to test string interning
echo -e "\nCreating $NUM_ENTITIES test entities in batches of $BATCH_SIZE..."
START=$(date +%s%N)

# Common tags that will be repeated
COMMON_TAGS=(
    "type:document"
    "status:active"
    "priority:high"
    "department:engineering"
    "project:optimization"
    "version:1.0"
    "team:backend"
    "category:test"
)

for batch in $(seq 0 $BATCH_SIZE $((NUM_ENTITIES-1))); do
    echo -n "."
    for i in $(seq $batch $((batch+BATCH_SIZE-1))); do
        if [ $i -ge $NUM_ENTITIES ]; then
            break
        fi
        
        # Create entities with repeated tags
        TAG_LIST=""
        for tag in "${COMMON_TAGS[@]}"; do
            TAG_LIST="$TAG_LIST\"$tag\","
        done
        TAG_LIST="${TAG_LIST%,}"
        
        curl -k -s -X POST "$BASE_URL/api/v1/entities/create" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "{
                \"tags\": [$TAG_LIST, \"id:test-$i\"],
                \"content\": \"Test entity $i with repeated content patterns for optimization testing\"
            }" > /dev/null
    done
done

echo ""
END=$(date +%s%N)
DURATION=$(( (END - START) / 1000000 ))
echo "Created $NUM_ENTITIES entities in ${DURATION}ms"
echo "Average: $((DURATION / NUM_ENTITIES))ms per entity"

# Get final metrics
echo -e "\nFinal metrics:"
FINAL_METRICS=$(curl -k -s -X GET "$BASE_URL/api/v1/system/metrics" \
    -H "Authorization: Bearer $TOKEN")
    
echo "$FINAL_METRICS" | jq '{
    memory_alloc: .memory.alloc_bytes,
    heap_in_use: .memory.heap_alloc_bytes,
    heap_sys: .memory.heap_sys_bytes,
    gc_runs: .performance.gc_runs,
    goroutines: .performance.goroutines,
    entity_count: .database.total_entities,
    unique_tags: .database.tags_unique
}'

FINAL_ALLOC=$(echo "$FINAL_METRICS" | jq -r '.memory.alloc_bytes')
FINAL_GC=$(echo "$FINAL_METRICS" | jq -r '.performance.gc_runs')

# Calculate improvements
ALLOC_DIFF=$((FINAL_ALLOC - INITIAL_ALLOC))
GC_DIFF=$((FINAL_GC - INITIAL_GC))
ALLOC_PER_ENTITY=$((ALLOC_DIFF / NUM_ENTITIES))

echo -e "\nOptimization Results:"
echo "===================="
echo "Memory allocated: $(numfmt --to=iec $ALLOC_DIFF) total"
echo "Memory per entity: $(numfmt --to=iec $ALLOC_PER_ENTITY)"
echo "GC runs during test: $GC_DIFF"
echo "Entities created: $NUM_ENTITIES"

# Test query performance
echo -e "\nTesting query performance..."
QUERY_START=$(date +%s%N)

# Query all test entities
curl -k -s -X GET "$BASE_URL/api/v1/entities/query?tags=type:document&limit=100" \
    -H "Authorization: Bearer $TOKEN" > /dev/null

QUERY_END=$(date +%s%N)
QUERY_DURATION=$(( (QUERY_END - QUERY_START) / 1000000 ))
echo "Query 100 entities: ${QUERY_DURATION}ms"

# Test list by tags
LIST_START=$(date +%s%N)

curl -k -s -X GET "$BASE_URL/api/v1/entities/list?tags=status:active,priority:high&limit=50" \
    -H "Authorization: Bearer $TOKEN" > /dev/null

LIST_END=$(date +%s%N)
LIST_DURATION=$(( (LIST_END - LIST_START) / 1000000 ))
echo "List by tags (50 entities): ${LIST_DURATION}ms"

echo -e "\nTest completed!"