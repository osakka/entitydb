#!/bin/bash
# Monitor EntityDB database size growth

echo "=== EntityDB Size Monitoring ==="
echo "Start time: $(date)"
echo "Initial size: $(ls -lh /opt/entitydb/var/entities.edb | awk '{print $5}')"
echo ""

while true; do
    if [ -f /opt/entitydb/var/entities.edb ]; then
        SIZE=$(ls -lh /opt/entitydb/var/entities.edb | awk '{print $5}')
        ACTUAL=$(du -h /opt/entitydb/var/entities.edb | awk '{print $1}')
        APPARENT=$(du -h --apparent-size /opt/entitydb/var/entities.edb | awk '{print $1}')
        TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
        
        echo "[$TIMESTAMP] Size: $SIZE | Actual: $ACTUAL | Apparent: $APPARENT"
        
        # Check for sparse file (apparent >> actual)
        if [ "$APPARENT" != "$ACTUAL" ]; then
            echo "  WARNING: Sparse file detected!"
        fi
    else
        echo "[$(date +"%Y-%m-%d %H:%M:%S")] Database file not found!"
    fi
    
    sleep 300  # Check every 5 minutes
done