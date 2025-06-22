#!/bin/bash
# CPU and System Load Monitor for 12-Hour Stability Test
# Logs CPU load averages and utilization every 30 seconds

LOG_DIR="$1"
if [ -z "$LOG_DIR" ]; then
    echo "Usage: $0 <log_directory>"
    exit 1
fi

# Create CSV header
echo "timestamp,metric_type,load_1min,load_5min,load_15min,cpu_user_percent" > "$LOG_DIR/cpu_metrics.csv"

echo "Starting CPU monitoring - logging to $LOG_DIR/cpu_metrics.csv"

while true; do
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Get load averages
    load_avg=$(cat /proc/loadavg | cut -d' ' -f1-3)
    load_1min=$(echo $load_avg | cut -d' ' -f1)
    load_5min=$(echo $load_avg | cut -d' ' -f2)  
    load_15min=$(echo $load_avg | cut -d' ' -f3)
    
    # Get CPU utilization (user percentage)
    cpu_user=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//')
    
    # Log the data
    echo "$timestamp,CPU,$load_1min,$load_5min,$load_15min,$cpu_user" >> "$LOG_DIR/cpu_metrics.csv"
    
    # Alert if load is too high (legendary efficiency threshold)
    if (( $(echo "$load_1min > 2.0" | bc -l) )); then
        echo "[$timestamp] ALERT: High CPU load detected: $load_1min" >> "$LOG_DIR/alerts.log"
    fi
    
    sleep 30
done