#!/bin/bash
# Detailed EntityDB Performance Analysis

BASE_URL="https://localhost:8085/api/v1"

echo "=== EntityDB Detailed Performance Analysis ==="
echo "Time: $(date)"
echo

# Function to time a request with detailed breakdown
time_request() {
    local method=$1
    local url=$2
    local data=$3
    local auth=$4
    
    echo "Request: $method $url"
    
    if [ -n "$auth" ]; then
        if [ -n "$data" ]; then
            curl -k -s -o /dev/null -X "$method" "$url" \
                -H "Authorization: Bearer $auth" \
                -H "Content-Type: application/json" \
                -d "$data" \
                -w "Connect: %{time_connect}s\nTLS: %{time_appconnect}s\nServer Processing: %{time_starttransfer}s\nTotal: %{time_total}s\nHTTP Code: %{http_code}\n\n"
        else
            curl -k -s -o /dev/null -X "$method" "$url" \
                -H "Authorization: Bearer $auth" \
                -w "Connect: %{time_connect}s\nTLS: %{time_appconnect}s\nServer Processing: %{time_starttransfer}s\nTotal: %{time_total}s\nHTTP Code: %{http_code}\n\n"
        fi
    else
        curl -k -s -o /dev/null -X "$method" "$url" \
            -w "Connect: %{time_connect}s\nTLS: %{time_appconnect}s\nServer Processing: %{time_starttransfer}s\nTotal: %{time_total}s\nHTTP Code: %{http_code}\n\n"
    fi
}

# Get auth token
echo "Getting auth token..."
TOKEN=$(curl -k -s -X POST $BASE_URL/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}' | jq -r '.token')
echo

echo "=== Individual Request Timings ==="
echo

# Test various endpoints
time_request "GET" "https://localhost:8085/health" "" ""
time_request "POST" "$BASE_URL/auth/login" '{"username":"admin","password":"admin"}' ""
time_request "GET" "$BASE_URL/entities/list" "" "$TOKEN"
time_request "GET" "$BASE_URL/datasets" "" "$TOKEN"
time_request "GET" "$BASE_URL/entities/query?limit=10" "" "$TOKEN"

# Test server under load
echo "=== Load Test (50 concurrent requests) ==="
echo

# Create a temporary file for storing times
TMPFILE=$(mktemp)

# Run 50 requests in parallel
for i in {1..50}; do
    (
        START=$(date +%s.%N)
        curl -k -s -X GET $BASE_URL/entities/list \
            -H "Authorization: Bearer $TOKEN" \
            -o /dev/null
        END=$(date +%s.%N)
        echo "$(echo "$END - $START" | bc)" >> $TMPFILE
    ) &
done

# Wait for all background jobs
wait

# Calculate statistics
echo "Response times for 50 concurrent requests:"
sort -n $TMPFILE | awk '
    BEGIN { count = 0; total = 0; }
    {
        times[count++] = $1;
        total += $1;
    }
    END {
        avg = total / count;
        if (count % 2 == 0) {
            median = (times[count/2 - 1] + times[count/2]) / 2;
        } else {
            median = times[(count-1)/2];
        }
        print "Min: " times[0] "s";
        print "Max: " times[count-1] "s";
        print "Median: " median "s";
        print "Average: " avg "s";
        print "Total requests: " count;
    }
'

rm -f $TMPFILE

# Check system resources
echo
echo "=== System Resource Check ==="
echo "CPU Load:"
uptime
echo
echo "Memory Usage:"
free -h | grep -E "^(Mem|Swap):"
echo
echo "EntityDB Process Info:"
ps aux | grep entitydb | grep -v grep | head -1

# Check if running with performance flags
echo
echo "=== Configuration Check ==="
grep -E "(WAL_ONLY|HIGH_PERFORMANCE|DATASPACE)" /opt/entitydb/var/entitydb.env

# Database size check
echo
echo "=== Database Statistics ==="
DB_SIZE=$(du -sh /opt/entitydb/var/*.bin 2>/dev/null | awk '{print $1}' | head -1)
echo "Database file size: ${DB_SIZE:-N/A}"
ENTITY_COUNT=$(curl -k -s -X GET $BASE_URL/entities/list -H "Authorization: Bearer $TOKEN" | jq '. | length')
echo "Total entities: $ENTITY_COUNT"

echo
echo "=== Performance Recommendations ==="
echo
echo "Based on the analysis:"
echo "1. If TLS time is high: Consider using HTTP for internal services"
echo "2. If Server Processing is high: Check if WAL_ONLY=true is enabled"
echo "3. If Total time is high with many entities: Check if HIGH_PERFORMANCE=true"
echo "4. For better concurrency: Ensure ENTITYDB_DATASET_ISOLATION=true is enabled"