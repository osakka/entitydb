#!/bin/bash

# Monitor the stress test in real-time

echo "=== EntityDB Stress Test Monitor ==="
echo "Press Ctrl+C to stop monitoring"

while true; do
    clear
    echo "=== EntityDB Stress Test Monitor ==="
    echo "Time: $(date)"
    echo
    
    # Get entity count
    ENTITY_COUNT=$(curl -s http://localhost:8085/api/v1/entities/list | jq length 2>/dev/null || echo "0")
    echo "Total Entities: $ENTITY_COUNT"
    
    # Get memory usage of entitydb process
    PID=$(pgrep -f "entitydb" | head -1)
    if [ -n "$PID" ]; then
        MEM=$(ps -o rss= -p $PID | awk '{print $1/1024 " MB"}')
        CPU=$(ps -o %cpu= -p $PID)
        echo "Server PID: $PID"
        echo "Memory Usage: $MEM"
        echo "CPU Usage: $CPU%"
    fi
    
    # Check disk usage
    DISK_USAGE=$(du -sh /opt/entitydb/var/entitydb.ebf 2>/dev/null || echo "N/A")
    echo "Database Size: $DISK_USAGE"
    
    # Get latest log entries
    echo -e "\nLatest Log Entries:"
    if [ -f /opt/entitydb/var/entitydb.log ]; then
        tail -5 /opt/entitydb/var/entitydb.log | grep -v "DEBUG"
    else
        echo "(No log file found)"
    fi
    
    sleep 2
done
