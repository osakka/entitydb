#!/bin/bash
# API Health Monitor for 12-Hour Stability Test
# Tests EntityDB API responsiveness every 5 minutes

LOG_DIR="$1"
if [ -z "$LOG_DIR" ]; then
    echo "Usage: $0 <log_directory>"
    exit 1
fi

# Create CSV header
echo "timestamp,endpoint,status,response_time_ms,http_code" > "$LOG_DIR/api_health.csv"

echo "Starting API health monitoring - logging to $LOG_DIR/api_health.csv"

# Function to test an endpoint
test_endpoint() {
    local endpoint="$1"
    local url="https://localhost:8085$endpoint"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Measure response time
    start_time=$(date +%s%3N)
    
    # Perform the API call with timeout
    response=$(curl -k -s -w "%{http_code}" -o /dev/null --max-time 10 "$url" 2>/dev/null)
    http_code=$?
    
    end_time=$(date +%s%3N)
    response_time=$((end_time - start_time))
    
    if [ $http_code -eq 0 ] && [ "$response" = "200" ]; then
        status="success"
        echo "$timestamp,$endpoint,$status,$response_time,$response" >> "$LOG_DIR/api_health.csv"
    else
        status="failed"
        echo "$timestamp,$endpoint,$status,$response_time,$response" >> "$LOG_DIR/api_health.csv"
        echo "[$timestamp] API ERROR: $endpoint failed (HTTP: $response, curl: $http_code)" >> "$LOG_DIR/alerts.log"
    fi
    
    # Alert for slow responses (>2000ms)
    if (( response_time > 2000 )); then
        echo "[$timestamp] WARNING: Slow API response: $endpoint took ${response_time}ms" >> "$LOG_DIR/alerts.log"
    fi
}

while true; do
    # Test critical endpoints
    test_endpoint "/health"
    sleep 30
    test_endpoint "/api/v1/system/metrics"
    sleep 30
    test_endpoint "/metrics"
    sleep 30
    
    # Test a simple entity operation (if admin credentials exist)
    test_endpoint "/api/v1/entities/list"
    
    # Wait remainder of 5 minutes (240 seconds already used)
    sleep 210
done