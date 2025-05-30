#!/bin/bash

# Final optimization test for EntityDB

set -e

BASE_URL="https://localhost:8085"
ADMIN_USER="admin"
ADMIN_PASS="admin"

echo "EntityDB Final Optimization Test"
echo "================================"
echo "Testing all optimizations:"
echo "- Buffer pooling"
echo "- String interning"
echo "- Improved sorting algorithms"
echo ""

# Login
TOKEN=$(curl -k -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}" | \
    jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "Failed to login"
    exit 1
fi

# Get initial state
INITIAL=$(curl -k -s -X GET "$BASE_URL/api/v1/system/metrics" \
    -H "Authorization: Bearer $TOKEN")

echo "Initial State:"
echo "============="
echo "$INITIAL" | jq '{
    entities: .database.total_entities,
    memory_mb: (.memory.alloc_bytes / 1048576),
    gc_runs: .performance.gc_runs
}'

# Test 1: Entity Creation Performance
echo -e "\nTest 1: Entity Creation (100 entities)"
echo "======================================"
START=$(date +%s%N)

for i in $(seq 1 100); do
    curl -k -s -X POST "$BASE_URL/api/v1/entities/create" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "tags": ["type:performance", "test:optimization", "iteration:'$i'"],
            "content": "Performance test content"
        }' > /dev/null
done

END=$(date +%s%N)
CREATE_TIME=$(( (END - START) / 1000000 ))
echo "Time: ${CREATE_TIME}ms ($(( CREATE_TIME / 100 ))ms per entity)"

# Test 2: Query Performance
echo -e "\nTest 2: Query Performance"
echo "========================"

# Simple tag query
START=$(date +%s%N)
curl -k -s -X GET "$BASE_URL/api/v1/entities/list?tags=type:performance&limit=50" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
END=$(date +%s%N)
QUERY_TIME=$(( (END - START) / 1000000 ))
echo "Tag query (50 entities): ${QUERY_TIME}ms"

# Advanced query
START=$(date +%s%N)
curl -k -s -X GET "$BASE_URL/api/v1/entities/query?tags=test:optimization&sort_by=created_at&sort_order=desc&limit=20" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
END=$(date +%s%N)
ADV_QUERY_TIME=$(( (END - START) / 1000000 ))
echo "Advanced query (20 entities): ${ADV_QUERY_TIME}ms"

# Test 3: Temporal Query Performance
echo -e "\nTest 3: Temporal Performance"
echo "==========================="

# Get an entity ID for temporal testing
ENTITY_ID=$(curl -k -s -X GET "$BASE_URL/api/v1/entities/list?tags=type:performance&limit=1" \
    -H "Authorization: Bearer $TOKEN" | jq -r '.entities[0].id')

if [ "$ENTITY_ID" != "null" ] && [ -n "$ENTITY_ID" ]; then
    # History query
    START=$(date +%s%N)
    curl -k -s -X GET "$BASE_URL/api/v1/entities/history?id=${ENTITY_ID}&limit=10" \
        -H "Authorization: Bearer $TOKEN" > /dev/null
    END=$(date +%s%N)
    HISTORY_TIME=$(( (END - START) / 1000000 ))
    echo "History query: ${HISTORY_TIME}ms"
    
    # As-of query
    AS_OF=$(date -d "1 minute ago" -u +"%Y-%m-%dT%H:%M:%SZ")
    START=$(date +%s%N)
    curl -k -s -X GET "$BASE_URL/api/v1/entities/as-of?id=${ENTITY_ID}&timestamp=${AS_OF}" \
        -H "Authorization: Bearer $TOKEN" > /dev/null 2>&1
    END=$(date +%s%N)
    ASOF_TIME=$(( (END - START) / 1000000 ))
    echo "As-of query: ${ASOF_TIME}ms"
fi

# Final metrics
echo -e "\nFinal State:"
echo "============"
FINAL=$(curl -k -s -X GET "$BASE_URL/api/v1/system/metrics" \
    -H "Authorization: Bearer $TOKEN")

echo "$FINAL" | jq '{
    entities: .database.total_entities,
    memory_mb: (.memory.alloc_bytes / 1048576),
    gc_runs: .performance.gc_runs,
    unique_tags: .database.tags_unique
}'

# Calculate differences
INITIAL_MEM=$(echo "$INITIAL" | jq -r '.memory.alloc_bytes')
FINAL_MEM=$(echo "$FINAL" | jq -r '.memory.alloc_bytes')
INITIAL_GC=$(echo "$INITIAL" | jq -r '.performance.gc_runs')
FINAL_GC=$(echo "$FINAL" | jq -r '.performance.gc_runs')

MEM_DIFF=$((FINAL_MEM - INITIAL_MEM))
GC_DIFF=$((FINAL_GC - INITIAL_GC))

echo -e "\nOptimization Summary:"
echo "===================="
echo "Memory growth: $(numfmt --to=iec $MEM_DIFF)"
echo "GC runs: $GC_DIFF"
echo "Create performance: $(( CREATE_TIME / 100 ))ms per entity"
echo "Query performance: ${QUERY_TIME}ms"
echo -e "\nTest completed!"