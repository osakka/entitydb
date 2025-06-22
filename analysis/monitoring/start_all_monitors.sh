#!/bin/bash
# Master script to start all monitoring systems for 12-hour stability test

LOG_DIR="$1"
if [ -z "$LOG_DIR" ]; then
    echo "Usage: $0 <log_directory>"
    exit 1
fi

# Ensure log directory exists
mkdir -p "$LOG_DIR"

echo "Starting comprehensive monitoring system..."
echo "Results directory: $LOG_DIR"
echo "Start time: $(date)"

# Make scripts executable
chmod +x /opt/entitydb/monitoring/*.sh

# Start CPU monitoring
echo "Starting CPU monitor..."
nohup /opt/entitydb/monitoring/cpu_monitor.sh "$LOG_DIR" > "$LOG_DIR/cpu_monitor.log" 2>&1 &
CPU_PID=$!
echo $CPU_PID > "$LOG_DIR/cpu_monitor.pid"

# Start memory monitoring  
echo "Starting memory monitor..."
nohup /opt/entitydb/monitoring/memory_monitor.sh "$LOG_DIR" > "$LOG_DIR/memory_monitor.log" 2>&1 &
MEM_PID=$!
echo $MEM_PID > "$LOG_DIR/memory_monitor.pid"

# Start EntityDB specific monitoring
echo "Starting EntityDB monitor..."
nohup /opt/entitydb/monitoring/entitydb_monitor.sh "$LOG_DIR" > "$LOG_DIR/entitydb_monitor.log" 2>&1 &
EDB_PID=$!
echo $EDB_PID > "$LOG_DIR/entitydb_monitor.pid"

# Start API health checking
echo "Starting API health monitor..."
nohup /opt/entitydb/monitoring/api_health_check.sh "$LOG_DIR" > "$LOG_DIR/api_health_monitor.log" 2>&1 &
API_PID=$!
echo $API_PID > "$LOG_DIR/api_health_monitor.pid"

# Create monitoring summary
echo "Monitor PIDs:" > "$LOG_DIR/monitor_summary.txt"
echo "CPU Monitor: $CPU_PID" >> "$LOG_DIR/monitor_summary.txt"
echo "Memory Monitor: $MEM_PID" >> "$LOG_DIR/monitor_summary.txt"
echo "EntityDB Monitor: $EDB_PID" >> "$LOG_DIR/monitor_summary.txt"
echo "API Health Monitor: $API_PID" >> "$LOG_DIR/monitor_summary.txt"
echo "Started at: $(date)" >> "$LOG_DIR/monitor_summary.txt"

echo "All monitoring systems started successfully!"
echo "Monitor processes:"
echo "  CPU Monitor PID: $CPU_PID"
echo "  Memory Monitor PID: $MEM_PID"  
echo "  EntityDB Monitor PID: $EDB_PID"
echo "  API Health Monitor PID: $API_PID"
echo ""
echo "To stop all monitors: kill $CPU_PID $MEM_PID $EDB_PID $API_PID"
echo "Monitor logs available in: $LOG_DIR"