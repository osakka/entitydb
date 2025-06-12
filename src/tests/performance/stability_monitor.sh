#!/bin/bash
# Monitor EntityDB stability for 5 minutes

HOST="https://localhost:8085"
DURATION=300  # 5 minutes
INTERVAL=30   # Check every 30 seconds

echo "🔍 EntityDB Stability Monitor"
echo "============================="
echo "Duration: 5 minutes"
echo "Interval: 30 seconds"
echo ""

start_time=$(date +%s)
iteration=0

while [ $(($(date +%s) - start_time)) -lt $DURATION ]; do
    iteration=$((iteration + 1))
    echo "Check #$iteration - $(date '+%Y-%m-%d %H:%M:%S')"
    
    # Get health metrics
    HEALTH=$(curl -k -s "$HOST/health")
    GOROUTINES=$(echo "$HEALTH" | jq -r '.memory.goroutines // 0')
    MEMORY_MB=$(echo "$HEALTH" | jq -r '.memory.alloc_mb // 0')
    
    # Get system metrics
    METRICS=$(curl -k -s "$HOST/api/v1/system/metrics")
    ERRORS=$(echo "$METRICS" | jq -r '.activity.errors_last_hour // 0')
    WAL_SIZE=$(echo "$METRICS" | jq -r '.database.wal_size_mb // 0')
    
    # Test a simple query
    QUERY_START=$(date +%s%N)
    curl -k -s "$HOST/api/v1/entities/list?limit=1" > /dev/null
    QUERY_END=$(date +%s%N)
    QUERY_MS=$(( (QUERY_END - QUERY_START) / 1000000 ))
    
    # Display results
    printf "  Goroutines: %3d | Memory: %6.1fMB | Errors: %2d | WAL: %4.1fMB | Query: %3dms\n" \
        "$GOROUTINES" "$MEMORY_MB" "$ERRORS" "$WAL_SIZE" "$QUERY_MS"
    
    # Alert on issues
    if [ $GOROUTINES -gt 100 ]; then
        echo "  ⚠️  WARNING: High goroutine count!"
    fi
    
    if [ $QUERY_MS -gt 100 ]; then
        echo "  ⚠️  WARNING: Slow query response!"
    fi
    
    sleep $INTERVAL
done

echo ""
echo "============================="
echo "✅ Stability monitoring complete"