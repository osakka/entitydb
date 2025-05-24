#!/bin/bash
# Quick performance check

echo "=== Quick Performance Check ==="
echo "Time: $(date)"
echo

# Get token
TOKEN=$(curl -k -s -X POST https://localhost:8085/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Single request timing
echo "Single request timing:"
time curl -k -s -X GET https://localhost:8085/api/v1/entities/list \
    -H "Authorization: Bearer $TOKEN" \
    -o /dev/null

echo
echo "10 sequential requests:"
START=$(date +%s.%N)
for i in {1..10}; do
    curl -k -s -X GET https://localhost:8085/api/v1/entities/list \
        -H "Authorization: Bearer $TOKEN" \
        -o /dev/null
done
END=$(date +%s.%N)
TOTAL=$(echo "$END - $START" | bc)
AVG=$(echo "scale=3; $TOTAL / 10" | bc)
echo "Total time: ${TOTAL}s"
echo "Average per request: ${AVG}s"

echo
echo "10 concurrent requests:"
START=$(date +%s.%N)
for i in {1..10}; do
    curl -k -s -X GET https://localhost:8085/api/v1/entities/list \
        -H "Authorization: Bearer $TOKEN" \
        -o /dev/null &
done
wait
END=$(date +%s.%N)
TOTAL=$(echo "$END - $START" | bc)
echo "Total time for 10 concurrent: ${TOTAL}s"

# Check specific slow endpoints
echo
echo "Testing specific endpoints:"
echo -n "Health check: "
time curl -k -s -X GET https://localhost:8085/health -o /dev/null

echo -n "Query with limit: "
time curl -k -s -X GET "https://localhost:8085/api/v1/entities/query?limit=10" \
    -H "Authorization: Bearer $TOKEN" \
    -o /dev/null

echo -n "Dataspace list: "
time curl -k -s -X GET https://localhost:8085/api/v1/dataspaces \
    -H "Authorization: Bearer $TOKEN" \
    -o /dev/null