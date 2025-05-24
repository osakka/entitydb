#!/bin/bash
# Performance benchmark after optimizations

echo "=== EntityDB Performance Benchmark - POST OPTIMIZATION ==="
echo "Time: $(date)"
echo

# Get token
TOKEN=$(curl -k -s -X POST https://localhost:8085/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}' | jq -r '.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "Failed to get auth token"
    exit 1
fi

echo "1. HEALTH CHECK (no auth, no DB query):"
echo "----------------------------------------"
for i in {1..5}; do
    time curl -k -s -X GET https://localhost:8085/health -o /dev/null
done
echo

echo "2. LIST ENTITIES (with caching):"
echo "--------------------------------"
for i in {1..5}; do
    time curl -k -s -X GET https://localhost:8085/api/v1/entities/list \
        -H "Authorization: Bearer $TOKEN" \
        -o /dev/null
done
echo

echo "3. QUERY ENTITIES (limit 10, cached):"
echo "-------------------------------------"
for i in {1..5}; do
    time curl -k -s -X GET "https://localhost:8085/api/v1/entities/query?limit=10" \
        -H "Authorization: Bearer $TOKEN" \
        -o /dev/null
done
echo

echo "4. CONCURRENT TEST (10 requests):"
echo "---------------------------------"
START=$(date +%s.%N)
for i in {1..10}; do
    curl -k -s -X GET https://localhost:8085/api/v1/entities/list \
        -H "Authorization: Bearer $TOKEN" \
        -o /dev/null &
done
wait
END=$(date +%s.%N)
TOTAL=$(echo "$END - $START" | bc)
echo "Total time for 10 concurrent requests: ${TOTAL}s"
AVG=$(echo "scale=3; $TOTAL / 10" | bc)
echo "Average per request: ${AVG}s"
echo

echo "5. STRESS TEST (50 concurrent):"
echo "-------------------------------"
START=$(date +%s.%N)
for i in {1..50}; do
    curl -k -s -X GET https://localhost:8085/api/v1/entities/list \
        -H "Authorization: Bearer $TOKEN" \
        -o /dev/null &
done
wait
END=$(date +%s.%N)
TOTAL=$(echo "$END - $START" | bc)
echo "Total time for 50 concurrent requests: ${TOTAL}s"
AVG=$(echo "scale=3; $TOTAL / 50" | bc)
echo "Average per request: ${AVG}s"
echo

# Get cache stats if available
echo "6. PERFORMANCE METRICS:"
echo "----------------------"
curl -k -s https://localhost:8085/api/v1/system/metrics \
    -H "Authorization: Bearer $TOKEN" | jq '.performance // empty'

echo
echo "=== OPTIMIZATION SUMMARY ==="
echo
echo "Target Goals:"
echo "- Single request: < 50ms ✓ Check above"
echo "- Health check: < 5ms ✓ Check above"
echo "- Concurrent scaling: Linear ✓ Check above"
echo
echo "Implemented Optimizations:"
echo "✓ Removed DB queries from health check"
echo "✓ Added request-level entity caching (5min TTL)"
echo "✓ Created optimized list/query handlers"
echo "✓ Implemented reader connection pooling"
echo "✓ Added performance metrics logging"