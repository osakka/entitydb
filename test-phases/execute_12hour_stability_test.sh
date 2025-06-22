#!/bin/bash
# 12-Hour EntityDB Stability Test Executor
# Comprehensive automated test of intelligent recovery system

set -e

# Test configuration
TEST_START=$(date +%s)
TEST_DURATION=43200  # 12 hours in seconds
LOG_DIR="/opt/entitydb/test-results/$(date +%Y%m%d-%H%M%S)"

echo "=========================================="
echo "EntityDB 12-Hour Stability Test (v2.34.0)"
echo "=========================================="
echo "Testing intelligent recovery system performance"
echo "Start time: $(date)"
echo "Duration: 12 hours"
echo "Results directory: $LOG_DIR"
echo ""

# Verify EntityDB is running
if ! pgrep entitydb > /dev/null; then
    echo "ERROR: EntityDB is not running!"
    echo "Please start EntityDB with: ./bin/entitydbd.sh start"
    exit 1
fi

echo "✓ EntityDB is running (PID: $(pgrep entitydb))"

# Create test environment
mkdir -p "$LOG_DIR"
echo "✓ Created results directory: $LOG_DIR"

# Log test configuration
cat > "$LOG_DIR/test_config.txt" << EOF
EntityDB 12-Hour Stability Test Configuration
============================================
Test Version: v2.34.0 (Post-CPU Fix)
Start Time: $(date)
Duration: 12 hours (43200 seconds)
Target: Validate intelligent recovery system
EntityDB PID: $(pgrep entitydb)
Initial CPU Load: $(cat /proc/loadavg | cut -d' ' -f1)
Initial Memory: $(free -m | awk 'NR==2{printf "%.1f%% (%d/%d MB)", $3*100/$2, $3, $2}')

Success Criteria:
- CPU load < 1.5 average throughout test
- No infinite recovery loops  
- Memory usage stable
- API responses < 2000ms average
- No critical errors

Alert Thresholds:
- CPU load > 2.0 sustained (5+ minutes)
- Recovery attempts > 0 (intelligent system should prevent)
- Memory growth > 50MB/hour
- API failures > 5%
EOF

echo "✓ Test configuration logged"

# Start comprehensive monitoring
echo "Starting monitoring systems..."
chmod +x /opt/entitydb/monitoring/start_all_monitors.sh
/opt/entitydb/monitoring/start_all_monitors.sh "$LOG_DIR"

if [ $? -eq 0 ]; then
    echo "✓ All monitoring systems started"
else
    echo "ERROR: Failed to start monitoring systems"
    exit 1
fi

# Log initial baseline metrics
echo "Recording baseline metrics..."
echo "Baseline captured at: $(date)" > "$LOG_DIR/baseline_metrics.txt"
echo "CPU Load: $(cat /proc/loadavg)" >> "$LOG_DIR/baseline_metrics.txt"
echo "Memory: $(free -m | grep '^Mem')" >> "$LOG_DIR/baseline_metrics.txt"
echo "EntityDB Process: $(ps -p $(pgrep entitydb) -o pid,ppid,%cpu,%mem,cmd --no-headers)" >> "$LOG_DIR/baseline_metrics.txt"
echo "✓ Baseline metrics recorded"

# Start progress tracking
echo "Starting 12-hour test monitoring..."
cat > "$LOG_DIR/progress_tracker.sh" << 'EOF'
#!/bin/bash
LOG_DIR="$1"
START_TIME="$2"
DURATION="$3"

while true; do
    current_time=$(date +%s)
    elapsed=$((current_time - START_TIME))
    remaining=$((DURATION - elapsed))
    
    if [ $remaining -le 0 ]; then
        echo "Test completed at $(date)" >> "$LOG_DIR/progress.log"
        break
    fi
    
    hours_elapsed=$((elapsed / 3600))
    hours_remaining=$((remaining / 3600))
    percent_complete=$((elapsed * 100 / DURATION))
    
    echo "$(date): ${hours_elapsed}h elapsed, ${hours_remaining}h remaining (${percent_complete}% complete)" >> "$LOG_DIR/progress.log"
    
    # Check for critical alerts
    if [ -f "$LOG_DIR/alerts.log" ]; then
        alert_count=$(wc -l < "$LOG_DIR/alerts.log")
        if [ $alert_count -gt 0 ]; then
            echo "$(date): $alert_count alerts detected" >> "$LOG_DIR/progress.log"
        fi
    fi
    
    sleep 1800  # Update every 30 minutes
done
EOF

chmod +x "$LOG_DIR/progress_tracker.sh"
nohup "$LOG_DIR/progress_tracker.sh" "$LOG_DIR" "$TEST_START" "$TEST_DURATION" &
PROGRESS_PID=$!
echo $PROGRESS_PID > "$LOG_DIR/progress_tracker.pid"

echo "✓ Progress tracking started (PID: $PROGRESS_PID)"

# Create real-time status monitor
cat > "$LOG_DIR/show_status.sh" << 'EOF'
#!/bin/bash
LOG_DIR="$1"

echo "=========================================="
echo "EntityDB 12-Hour Test - Live Status"
echo "=========================================="

if [ -f "$LOG_DIR/progress.log" ]; then
    echo "Progress:"
    tail -n 3 "$LOG_DIR/progress.log"
    echo ""
fi

if [ -f "$LOG_DIR/alerts.log" ]; then
    alert_count=$(wc -l < "$LOG_DIR/alerts.log")
    echo "Alerts: $alert_count total"
    if [ $alert_count -gt 0 ]; then
        echo "Recent alerts:"
        tail -n 5 "$LOG_DIR/alerts.log"
    fi
    echo ""
fi

if [ -f "$LOG_DIR/cpu_metrics.csv" ]; then
    echo "Current CPU Load:"
    tail -n 1 "$LOG_DIR/cpu_metrics.csv"
    echo ""
fi

if [ -f "$LOG_DIR/entitydb_metrics.csv" ]; then
    echo "EntityDB Status:"
    tail -n 1 "$LOG_DIR/entitydb_metrics.csv"
    echo ""
fi

echo "To monitor continuously: watch -n 30 $LOG_DIR/show_status.sh"
echo "To view all alerts: tail -f $LOG_DIR/alerts.log"
echo "To stop test: kill \$(cat $LOG_DIR/*.pid)"
EOF

chmod +x "$LOG_DIR/show_status.sh"

echo ""
echo "=========================================="
echo "12-HOUR STABILITY TEST IS NOW RUNNING"
echo "=========================================="
echo ""
echo "Test will complete at: $(date -d '+12 hours')"
echo ""
echo "Monitoring commands:"
echo "  Status: $LOG_DIR/show_status.sh $LOG_DIR"
echo "  Alerts: tail -f $LOG_DIR/alerts.log"
echo "  Progress: tail -f $LOG_DIR/progress.log"
echo ""
echo "The test will automatically:"
echo "  - Monitor CPU performance (target: <1.5 load)"
echo "  - Detect recovery attempts (target: 0)"
echo "  - Track memory usage and API response times"
echo "  - Generate comprehensive report at completion"
echo ""
echo "EntityDB Intelligent Recovery System Test in progress..."
echo "Validating legendary efficiency over 12 hours! 🚀"

# Wait for test completion (12 hours)
sleep $TEST_DURATION

echo ""
echo "=========================================="
echo "12-HOUR TEST COMPLETED!"
echo "=========================================="
echo "Completion time: $(date)"

# Stop all monitors
echo "Stopping monitoring systems..."
if [ -f "$LOG_DIR/cpu_monitor.pid" ]; then kill $(cat "$LOG_DIR/cpu_monitor.pid") 2>/dev/null || true; fi
if [ -f "$LOG_DIR/memory_monitor.pid" ]; then kill $(cat "$LOG_DIR/memory_monitor.pid") 2>/dev/null || true; fi
if [ -f "$LOG_DIR/entitydb_monitor.pid" ]; then kill $(cat "$LOG_DIR/entitydb_monitor.pid") 2>/dev/null || true; fi
if [ -f "$LOG_DIR/api_health_monitor.pid" ]; then kill $(cat "$LOG_DIR/api_health_monitor.pid") 2>/dev/null || true; fi
if [ -f "$LOG_DIR/progress_tracker.pid" ]; then kill $(cat "$LOG_DIR/progress_tracker.pid") 2>/dev/null || true; fi

echo "✓ All monitors stopped"

# Generate final summary
echo "Generating test summary..."
cat > "$LOG_DIR/test_summary.txt" << EOF
EntityDB 12-Hour Stability Test Summary
======================================
Test Version: v2.34.0 (Intelligent Recovery System)
Start Time: $(cat "$LOG_DIR/test_config.txt" | grep "Start Time" | cut -d': ' -f2-)
End Time: $(date)
Duration: 12 hours

Results Overview:
EOF

# Add basic statistics
if [ -f "$LOG_DIR/alerts.log" ]; then
    alert_count=$(wc -l < "$LOG_DIR/alerts.log")
    echo "Total Alerts: $alert_count" >> "$LOG_DIR/test_summary.txt"
else
    echo "Total Alerts: 0" >> "$LOG_DIR/test_summary.txt"
fi

if [ -f "$LOG_DIR/cpu_metrics.csv" ]; then
    avg_load=$(tail -n +2 "$LOG_DIR/cpu_metrics.csv" | awk -F',' '{sum+=$3; count++} END {printf "%.2f", sum/count}')
    max_load=$(tail -n +2 "$LOG_DIR/cpu_metrics.csv" | awk -F',' 'BEGIN{max=0} {if($3>max) max=$3} END {print max}')
    echo "Average CPU Load: $avg_load" >> "$LOG_DIR/test_summary.txt"
    echo "Maximum CPU Load: $max_load" >> "$LOG_DIR/test_summary.txt"
fi

if [ -f "$LOG_DIR/entitydb_metrics.csv" ]; then
    total_recovery_attempts=$(tail -n +2 "$LOG_DIR/entitydb_metrics.csv" | awk -F',' '{sum+=$5} END {print sum+0}')
    echo "Total Recovery Attempts: $total_recovery_attempts" >> "$LOG_DIR/test_summary.txt"
fi

echo "" >> "$LOG_DIR/test_summary.txt"
echo "Test Status: COMPLETED" >> "$LOG_DIR/test_summary.txt"
echo "Full results available in: $LOG_DIR" >> "$LOG_DIR/test_summary.txt"

echo "✓ Test summary generated"
echo ""
echo "Test completed successfully!"
echo "Results saved to: $LOG_DIR"
echo "Summary: cat $LOG_DIR/test_summary.txt"