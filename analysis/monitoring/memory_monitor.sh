#!/bin/bash
# Memory Usage Monitor for 12-Hour Stability Test
# Logs system and EntityDB memory usage every 60 seconds

LOG_DIR="$1"
if [ -z "$LOG_DIR" ]; then
    echo "Usage: $0 <log_directory>"
    exit 1
fi

# Create CSV header
echo "timestamp,metric_type,total_mem_mb,used_mem_mb,available_mem_mb,entitydb_mem_mb,entitydb_mem_percent" > "$LOG_DIR/memory_metrics.csv"

echo "Starting memory monitoring - logging to $LOG_DIR/memory_metrics.csv"

while true; do
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Get system memory info (in MB)
    mem_info=$(free -m)
    total_mem=$(echo "$mem_info" | awk 'NR==2{print $2}')
    used_mem=$(echo "$mem_info" | awk 'NR==2{print $3}')
    available_mem=$(echo "$mem_info" | awk 'NR==2{print $7}')
    
    # Get EntityDB process memory usage
    entitydb_mem_mb=0
    entitydb_mem_percent=0
    
    if pgrep entitydb > /dev/null; then
        pid=$(pgrep entitydb)
        # Memory in MB (RSS)
        entitydb_mem_kb=$(ps -p $pid -o rss --no-headers 2>/dev/null || echo 0)
        entitydb_mem_mb=$((entitydb_mem_kb / 1024))
        # Memory percentage
        entitydb_mem_percent=$(ps -p $pid -o %mem --no-headers 2>/dev/null || echo 0)
    fi
    
    # Log the data
    echo "$timestamp,MEMORY,$total_mem,$used_mem,$available_mem,$entitydb_mem_mb,$entitydb_mem_percent" >> "$LOG_DIR/memory_metrics.csv"
    
    # Alert if EntityDB memory usage is excessive (>500MB)
    if (( entitydb_mem_mb > 500 )); then
        echo "[$timestamp] ALERT: High EntityDB memory usage: ${entitydb_mem_mb}MB" >> "$LOG_DIR/alerts.log"
    fi
    
    # Alert if system memory is low (<100MB available)
    if (( available_mem < 100 )); then
        echo "[$timestamp] ALERT: Low system memory: ${available_mem}MB available" >> "$LOG_DIR/alerts.log"
    fi
    
    sleep 60
done