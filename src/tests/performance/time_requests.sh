#!/bin/bash
# EntityDB Request Timing Analysis

BASE_URL="https://localhost:8085/api/v1"
CURL_OPTS="-k -s -w @-"

# Create curl timing format file
cat > /tmp/curl_timing_format.txt << 'EOF'
    time_namelookup:  %{time_namelookup}s\n
    time_connect:     %{time_connect}s\n
    time_appconnect:  %{time_appconnect}s\n
    time_pretransfer: %{time_pretransfer}s\n
    time_redirect:    %{time_redirect}s\n
    time_starttransfer: %{time_starttransfer}s\n
    ----------\n
    time_total:       %{time_total}s\n
EOF

echo "=== EntityDB Request Timing Analysis ==="
echo "Server: $BASE_URL"
echo "Date: $(date)"
echo

# 1. Time the login request
echo "1. LOGIN REQUEST TIMING:"
echo "------------------------"
LOGIN_START=$(date +%s.%N)
LOGIN_RESPONSE=$(curl $CURL_OPTS -o /tmp/login_response.json -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' \
  /tmp/curl_timing_format.txt)
LOGIN_END=$(date +%s.%N)
echo "$LOGIN_RESPONSE"
echo "Shell measured time: $(echo "$LOGIN_END - $LOGIN_START" | bc)s"
TOKEN=$(jq -r '.token' /tmp/login_response.json)
echo

# 2. Time entity list request
echo "2. LIST ENTITIES TIMING:"
echo "------------------------"
LIST_START=$(date +%s.%N)
LIST_RESPONSE=$(curl $CURL_OPTS -o /tmp/list_response.json -X GET $BASE_URL/entities/list \
  -H "Authorization: Bearer $TOKEN" \
  /tmp/curl_timing_format.txt)
LIST_END=$(date +%s.%N)
echo "$LIST_RESPONSE"
echo "Shell measured time: $(echo "$LIST_END - $LIST_START" | bc)s"
ENTITY_COUNT=$(jq '. | length' /tmp/list_response.json)
echo "Entities returned: $ENTITY_COUNT"
echo

# 3. Time dataset list request
echo "3. LIST DATASPACES TIMING:"
echo "--------------------------"
DS_START=$(date +%s.%N)
DS_RESPONSE=$(curl $CURL_OPTS -o /tmp/ds_response.json -X GET $BASE_URL/datasets \
  -H "Authorization: Bearer $TOKEN" \
  /tmp/curl_timing_format.txt)
DS_END=$(date +%s.%N)
echo "$DS_RESPONSE"
echo "Shell measured time: $(echo "$DS_END - $DS_START" | bc)s"
DS_COUNT=$(jq '. | length' /tmp/ds_response.json)
echo "Datasets returned: $DS_COUNT"
echo

# 4. Time create entity request
echo "4. CREATE ENTITY TIMING:"
echo "------------------------"
CREATE_START=$(date +%s.%N)
CREATE_RESPONSE=$(curl $CURL_OPTS -o /tmp/create_response.json -X POST $BASE_URL/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:test", "perf:timing"],
    "content": "Test entity for timing analysis"
  }' \
  /tmp/curl_timing_format.txt)
CREATE_END=$(date +%s.%N)
echo "$CREATE_RESPONSE"
echo "Shell measured time: $(echo "$CREATE_END - $CREATE_START" | bc)s"
ENTITY_ID=$(jq -r '.entity.id' /tmp/create_response.json 2>/dev/null || echo "N/A")
echo "Created entity ID: $ENTITY_ID"
echo

# 5. Time query with filtering
echo "5. QUERY ENTITIES TIMING (with filter):"
echo "---------------------------------------"
QUERY_START=$(date +%s.%N)
QUERY_RESPONSE=$(curl $CURL_OPTS -o /tmp/query_response.json -X GET "$BASE_URL/entities/query?tags=type:test" \
  -H "Authorization: Bearer $TOKEN" \
  /tmp/curl_timing_format.txt)
QUERY_END=$(date +%s.%N)
echo "$QUERY_RESPONSE"
echo "Shell measured time: $(echo "$QUERY_END - $QUERY_START" | bc)s"
QUERY_COUNT=$(jq '.entities | length' /tmp/query_response.json 2>/dev/null || echo "0")
echo "Entities matched: $QUERY_COUNT"
echo

# 6. Time health check (no auth)
echo "6. HEALTH CHECK TIMING (no auth):"
echo "---------------------------------"
HEALTH_START=$(date +%s.%N)
HEALTH_RESPONSE=$(curl $CURL_OPTS -o /tmp/health_response.json -X GET https://localhost:8085/health \
  /tmp/curl_timing_format.txt)
HEALTH_END=$(date +%s.%N)
echo "$HEALTH_RESPONSE"
echo "Shell measured time: $(echo "$HEALTH_END - $HEALTH_START" | bc)s"
echo

# Run multiple requests to get average
echo "7. BULK TIMING TEST (10 list requests):"
echo "---------------------------------------"
TOTAL_TIME=0
for i in {1..10}; do
    START=$(date +%s.%N)
    curl -k -s -X GET $BASE_URL/entities/list \
      -H "Authorization: Bearer $TOKEN" \
      -o /dev/null
    END=$(date +%s.%N)
    ELAPSED=$(echo "$END - $START" | bc)
    TOTAL_TIME=$(echo "$TOTAL_TIME + $ELAPSED" | bc)
    echo "Request $i: ${ELAPSED}s"
done
AVG_TIME=$(echo "scale=4; $TOTAL_TIME / 10" | bc)
echo "Average time: ${AVG_TIME}s"
echo

# Check server logs for slow queries
echo "8. CHECKING SERVER PERFORMANCE:"
echo "-------------------------------"
echo "Recent server log entries (last 20 lines):"
tail -20 /opt/entitydb/var/entitydb.log | grep -E "(slow|took|duration|timing)" || echo "No timing entries found"

# Cleanup
rm -f /tmp/curl_timing_format.txt /tmp/*_response.json

echo
echo "=== Analysis Complete ==="
echo
echo "Summary:"
echo "- Login time: Check time_starttransfer above"
echo "- Entity list time: Check time_total for list operation"
echo "- Create time: Check time_total for create operation"
echo "- Average list time (10 requests): ${AVG_TIME}s"
echo
echo "Key metrics to watch:"
echo "- time_connect: Time to establish TCP connection"
echo "- time_appconnect: Time to complete SSL/TLS handshake"
echo "- time_starttransfer: Time until first byte received"
echo "- time_total: Total request time"