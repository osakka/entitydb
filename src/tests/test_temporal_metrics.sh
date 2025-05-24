#!/bin/bash
# Test temporal metrics collection in EntityDB

BASE_URL="https://localhost:8085/api/v1"
CURL_OPTS="-k -s"

echo "=== EntityDB Temporal Metrics Test ==="
echo "Demonstrating the power of temporal storage for metrics"
echo

# Login as admin
echo "1. Logging in..."
LOGIN_RESPONSE=$(curl $CURL_OPTS -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Failed to login"
    exit 1
fi
echo "✓ Logged in successfully"
echo

# Collect CPU metrics over time
echo "2. Simulating CPU metrics collection (one entity, multiple temporal values)..."
echo "   This creates ONE entity that accumulates values over time"
echo

for i in {1..5}; do
    # Simulate varying CPU usage
    CPU_VALUE=$(echo "scale=2; 20 + $RANDOM % 60" | bc)
    
    echo "   Collecting CPU usage: ${CPU_VALUE}%"
    curl $CURL_OPTS -X POST $BASE_URL/metrics/collect \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"metric_name\": \"cpu_usage\",
            \"value\": $CPU_VALUE,
            \"unit\": \"percent\",
            \"instance\": \"server1\",
            \"labels\": {
                \"datacenter\": \"us-west\",
                \"env\": \"production\"
            }
        }" | jq .
    
    sleep 2
done
echo

# Collect memory metrics
echo "3. Collecting memory metrics..."
for i in {1..3}; do
    MEMORY_GB=$(echo "scale=2; 4 + $RANDOM % 4" | bc)
    MEMORY_BYTES=$(echo "$MEMORY_GB * 1024 * 1024 * 1024" | bc)
    
    echo "   Collecting memory usage: ${MEMORY_GB}GB"
    curl $CURL_OPTS -X POST $BASE_URL/metrics/collect \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"metric_name\": \"memory_used\",
            \"value\": $MEMORY_BYTES,
            \"unit\": \"bytes\",
            \"instance\": \"server1\"
        }" | jq .
    
    sleep 1
done
echo

# Query current metrics
echo "4. Querying current metric values..."
curl $CURL_OPTS -X GET $BASE_URL/metrics/current \
    -H "Authorization: Bearer $TOKEN" | jq .
echo

# Get metric history
echo "5. Getting CPU usage history (time-series data from ONE entity)..."
curl $CURL_OPTS -X GET "$BASE_URL/metrics/history?metric=cpu_usage&instance=server1" \
    -H "Authorization: Bearer $TOKEN" | jq .
echo

# Show the actual entity to demonstrate temporal tags
echo "6. Examining the metric entity directly..."
# First get the entity ID
METRIC_ID="metric_cpu_usage_server1"
curl $CURL_OPTS -X GET "$BASE_URL/entities/get?id=$METRIC_ID&include_timestamps=true" \
    -H "Authorization: Bearer $TOKEN" | jq '.entity | {id, tags: .tags[:10], tag_count: (.tags | length), content}'
echo

# Demonstrate time-based queries
echo "7. Query metrics from specific time range..."
SINCE=$(date -u -d '5 minutes ago' +"%Y-%m-%dT%H:%M:%SZ")
UNTIL=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "   Time range: $SINCE to $UNTIL"
curl $CURL_OPTS -X GET "$BASE_URL/metrics/history?metric=cpu_usage&instance=server1&since=$SINCE&until=$UNTIL" \
    -H "Authorization: Bearer $TOKEN" | jq .
echo

echo "=== Key Insights ==="
echo "✓ ONE entity per metric (not one per data point)"
echo "✓ Each update adds a NEW temporal tag with timestamp"
echo "✓ Full history preserved in the entity's tags"
echo "✓ Can query by time range using temporal features"
echo "✓ No data loss, perfect for time-series analysis"
echo
echo "This is the POWER of EntityDB's temporal storage!"