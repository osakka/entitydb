#!/bin/bash
# MetHub Agent - Simple metrics collector for EntityDB
# Minimal dependencies - only requires curl and standard Unix tools

# Configuration
ENTITYDB_URL="${ENTITYDB_URL:-https://localhost:8085}"
ENTITYDB_USER="${ENTITYDB_USER:-admin}"
ENTITYDB_PASS="${ENTITYDB_PASS:-admin}"
METHUB_INTERVAL="${METHUB_INTERVAL:-30}"
METHUB_HOSTNAME="${METHUB_HOSTNAME:-$(hostname)}"
METHUB_HUB="metrics"

# Get auth token
AUTH_TOKEN=""

# Function to authenticate and get token
authenticate() {
    local response=$(curl -sk -X POST "$ENTITYDB_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$ENTITYDB_USER\",\"password\":\"$ENTITYDB_PASS\"}" 2>/dev/null)
    
    AUTH_TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$AUTH_TOKEN" ]; then
        echo "Authentication failed" >&2
        return 1
    fi
    return 0
}

# Function to send metric to EntityDB
send_metric() {
    local metric_type="$1"
    local metric_name="$2"
    local metric_value="$3"
    local metric_unit="${4:-}"
    local additional_tags="${5:-}"
    
    local timestamp=$(date +%s%N)  # Nanosecond timestamp
    
    # Build tags array
    local tags="\"type:metric\", \"metric:$metric_type\", \"host:$METHUB_HOSTNAME\", \"name:$metric_name\""
    if [ -n "$metric_unit" ]; then
        tags="$tags, \"unit:$metric_unit\""
    fi
    if [ -n "$additional_tags" ]; then
        tags="$tags, $additional_tags"
    fi
    
    # Create metric entity
    local metric_data=$(cat <<EOF
{
    "self": {
        "name": "$metric_name",
        "value": $metric_value,
        "timestamp": "$timestamp",
        "host": "$METHUB_HOSTNAME"
    },
    "traits": {
        "metric_type": "$metric_type",
        "host": "$METHUB_HOSTNAME"
    },
    "tags": [$tags]
}
EOF
)
    
    # Send to EntityDB
    curl -sk -X POST "$ENTITYDB_URL/api/v1/hubs/$METHUB_HUB/entities" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$metric_data" >/dev/null 2>&1
}

# CPU metrics collector
collect_cpu() {
    # CPU usage percentage
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{print 100 - $1}')
    send_metric "cpu" "cpu_usage" "$cpu_usage" "percent"
    
    # Load average
    local load_1min=$(uptime | awk -F'load average:' '{print $2}' | awk -F, '{print $1}' | xargs)
    local load_5min=$(uptime | awk -F'load average:' '{print $2}' | awk -F, '{print $2}' | xargs)
    local load_15min=$(uptime | awk -F'load average:' '{print $2}' | awk -F, '{print $3}' | xargs)
    
    send_metric "cpu" "load_1min" "$load_1min" "load"
    send_metric "cpu" "load_5min" "$load_5min" "load"
    send_metric "cpu" "load_15min" "$load_15min" "load"
}

# Memory metrics collector
collect_memory() {
    # Parse /proc/meminfo for memory stats
    local mem_total=$(awk '/^MemTotal:/ {print $2}' /proc/meminfo)
    local mem_free=$(awk '/^MemFree:/ {print $2}' /proc/meminfo)
    local mem_available=$(awk '/^MemAvailable:/ {print $2}' /proc/meminfo)
    local mem_buffers=$(awk '/^Buffers:/ {print $2}' /proc/meminfo)
    local mem_cached=$(awk '/^Cached:/ {print $2}' /proc/meminfo)
    local swap_total=$(awk '/^SwapTotal:/ {print $2}' /proc/meminfo)
    local swap_free=$(awk '/^SwapFree:/ {print $2}' /proc/meminfo)
    
    # Calculate used memory
    local mem_used=$((mem_total - mem_available))
    local swap_used=$((swap_total - swap_free))
    
    # Send memory metrics (convert KB to MB)
    send_metric "memory" "mem_total" "$((mem_total / 1024))" "MB"
    send_metric "memory" "mem_used" "$((mem_used / 1024))" "MB"
    send_metric "memory" "mem_free" "$((mem_free / 1024))" "MB"
    send_metric "memory" "mem_available" "$((mem_available / 1024))" "MB"
    send_metric "memory" "mem_buffers" "$((mem_buffers / 1024))" "MB"
    send_metric "memory" "mem_cached" "$((mem_cached / 1024))" "MB"
    
    # Memory usage percentage
    local mem_percent=$((mem_used * 100 / mem_total))
    send_metric "memory" "mem_percent" "$mem_percent" "percent"
    
    # Swap metrics
    if [ "$swap_total" -gt 0 ]; then
        send_metric "memory" "swap_total" "$((swap_total / 1024))" "MB"
        send_metric "memory" "swap_used" "$((swap_used / 1024))" "MB"
        send_metric "memory" "swap_free" "$((swap_free / 1024))" "MB"
        local swap_percent=$((swap_used * 100 / swap_total))
        send_metric "memory" "swap_percent" "$swap_percent" "percent"
    fi
}

# Disk metrics collector
collect_disk() {
    # Get disk usage for all mounted filesystems
    df -BM | grep -E '^/dev/' | while read line; do
        local device=$(echo "$line" | awk '{print $1}')
        local mount=$(echo "$line" | awk '{print $6}')
        local size=$(echo "$line" | awk '{print $2}' | sed 's/M//')
        local used=$(echo "$line" | awk '{print $3}' | sed 's/M//')
        local avail=$(echo "$line" | awk '{print $4}' | sed 's/M//')
        local percent=$(echo "$line" | awk '{print $5}' | sed 's/%//')
        
        # Clean device name for tags
        local device_tag=$(echo "$device" | sed 's/\//_/g')
        
        send_metric "disk" "disk_total" "$size" "MB" "\"device:$device_tag\", \"mount:$mount\""
        send_metric "disk" "disk_used" "$used" "MB" "\"device:$device_tag\", \"mount:$mount\""
        send_metric "disk" "disk_free" "$avail" "MB" "\"device:$device_tag\", \"mount:$mount\""
        send_metric "disk" "disk_percent" "$percent" "percent" "\"device:$device_tag\", \"mount:$mount\""
    done
    
    # Disk I/O stats (if available)
    if [ -r /proc/diskstats ]; then
        # Simple I/O metrics for main block devices
        grep -E '(sda|vda|nvme0n1) ' /proc/diskstats | head -1 | while read line; do
            local device=$(echo "$line" | awk '{print $3}')
            local reads=$(echo "$line" | awk '{print $4}')
            local writes=$(echo "$line" | awk '{print $8}')
            
            send_metric "disk" "disk_reads" "$reads" "ops" "\"device:$device\""
            send_metric "disk" "disk_writes" "$writes" "ops" "\"device:$device\""
        done
    fi
}

# Custom metric sender (for user-defined metrics)
send_custom_metric() {
    local name="$1"
    local value="$2"
    local unit="${3:-}"
    local tags="${4:-}"
    
    send_metric "custom" "$name" "$value" "$unit" "$tags"
}

# Main collection loop
main() {
    echo "MetHub Agent starting..."
    echo "Host: $METHUB_HOSTNAME"
    echo "Interval: ${METHUB_INTERVAL}s"
    echo "Server: $ENTITYDB_URL"
    
    # Authenticate
    if ! authenticate; then
        echo "Failed to authenticate with EntityDB"
        exit 1
    fi
    echo "Authenticated successfully"
    
    # Main loop
    while true; do
        # Collect all metrics
        collect_cpu
        collect_memory
        collect_disk
        
        # Allow custom metrics via environment or file
        if [ -f /etc/methub/custom-metrics.sh ]; then
            source /etc/methub/custom-metrics.sh
        fi
        
        # Sleep until next collection
        sleep "$METHUB_INTERVAL"
    done
}

# Handle signals
trap 'echo "Shutting down..."; exit 0' SIGTERM SIGINT

# Run main function
main