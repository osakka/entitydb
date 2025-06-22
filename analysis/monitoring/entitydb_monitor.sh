#!/bin/bash
# EntityDB Specific Monitor for 12-Hour Stability Test
# Monitors EntityDB health, recovery attempts, and key metrics every 2 minutes

LOG_DIR="$1"
if [ -z "$LOG_DIR" ]; then
    echo "Usage: $0 <log_directory>"
    exit 1
fi

# Create CSV header
echo "timestamp,status,cpu_percent,mem_percent,recovery_attempts,wal_size_mb,entities_count,log_errors" > "$LOG_DIR/entitydb_metrics.csv"

echo "Starting EntityDB monitoring - logging to $LOG_DIR/entitydb_metrics.csv"

while true; do
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Check if EntityDB is running
    if pgrep entitydb > /dev/null; then
        status="running"
        pid=$(pgrep entitydb)
        
        # Get EntityDB process stats
        cpu_percent=$(ps -p $pid -o %cpu --no-headers 2>/dev/null | xargs || echo 0)
        mem_percent=$(ps -p $pid -o %mem --no-headers 2>/dev/null | xargs || echo 0)
        
        # Check log for recovery attempts (last 2 minutes) - THIS IS THE CRITICAL METRIC
        two_minutes_ago=$(date -d '2 minutes ago' '+%Y/%m/%d %H:%M')
        recovery_attempts=$(tail -n 200 /opt/entitydb/var/entitydb.log 2>/dev/null | \
                          grep "$two_minutes_ago" | \
                          grep -c "attempting to recover corrupted entity" || echo 0)
        
        # Get WAL file size in MB if exists
        wal_size_mb=0
        if [ -f /opt/entitydb/var/entities.edb ]; then
            wal_size_bytes=$(stat -c%s /opt/entitydb/var/entities.edb 2>/dev/null || echo 0)
            wal_size_mb=$((wal_size_bytes / 1024 / 1024))
        fi
        
        # Try to get entity count via API (with timeout)
        entities_count=$(timeout 5 curl -k -s https://localhost:8085/api/v1/system/metrics 2>/dev/null | \
                        grep -o '"entity_count_total":[0-9]*' | cut -d':' -f2 || echo "unknown")
        
        # Count recent errors in log (last 2 minutes)
        log_errors=$(tail -n 200 /opt/entitydb/var/entitydb.log 2>/dev/null | \
                    grep "$two_minutes_ago" | \
                    grep -c "\[ERROR\]" || echo 0)
        
    else
        status="stopped"
        cpu_percent=0
        mem_percent=0
        recovery_attempts=0
        wal_size_mb=0
        entities_count=0
        log_errors=0
        
        echo "[$timestamp] CRITICAL: EntityDB process stopped!" >> "$LOG_DIR/alerts.log"
    fi
    
    # Log the data
    echo "$timestamp,$status,$cpu_percent,$mem_percent,$recovery_attempts,$wal_size_mb,$entities_count,$log_errors" >> "$LOG_DIR/entitydb_metrics.csv"
    
    # CRITICAL ALERT: Recovery attempts detected (intelligent recovery should prevent this)
    if (( recovery_attempts > 0 )); then
        echo "[$timestamp] CRITICAL: Recovery attempts detected: $recovery_attempts (intelligent recovery may be failing!)" >> "$LOG_DIR/alerts.log"
        
        # Capture the actual recovery log entries for analysis
        tail -n 200 /opt/entitydb/var/entitydb.log | grep "$two_minutes_ago" | \
        grep "attempting to recover corrupted entity" >> "$LOG_DIR/recovery_attempts.log"
    fi
    
    # Alert for excessive errors
    if (( log_errors > 10 )); then
        echo "[$timestamp] WARNING: High error rate: $log_errors errors in 2 minutes" >> "$LOG_DIR/alerts.log"
    fi
    
    # Alert for large WAL file (>100MB)
    if (( wal_size_mb > 100 )); then
        echo "[$timestamp] WARNING: Large WAL file: ${wal_size_mb}MB" >> "$LOG_DIR/alerts.log"
    fi
    
    sleep 120  # Every 2 minutes
done