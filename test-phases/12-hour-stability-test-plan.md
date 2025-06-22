# 12-Hour EntityDB Stability Test Plan
**Test Version**: v2.34.0 (Post-CPU Fix)  
**Date**: 2025-06-21  
**Objective**: Validate long-term stability of intelligent recovery system

## Test Overview

### Goal
Verify that EntityDB maintains legendary efficiency over 12 hours of continuous operation, with the intelligent recovery system preventing CPU performance degradation.

### Success Criteria
- CPU load remains below 1.5 average throughout test
- No infinite recovery loops detected  
- Memory usage stable (no leaks)
- All critical functionality operational
- EntityDB responsive to API requests
- Metrics collection functioning without performance impact

## Test Architecture

### Monitoring Infrastructure
```bash
# CPU and system monitoring (every 30 seconds)
/opt/entitydb/monitoring/cpu_monitor.sh

# Memory monitoring (every 60 seconds)  
/opt/entitydb/monitoring/memory_monitor.sh

# EntityDB specific monitoring (every 2 minutes)
/opt/entitydb/monitoring/entitydb_monitor.sh

# API health checking (every 5 minutes)
/opt/entitydb/monitoring/api_health_check.sh

# Recovery attempt detection (continuous log analysis)
/opt/entitydb/monitoring/recovery_monitor.sh
```

### Data Collection Points
1. **System Metrics**: CPU, memory, disk I/O, network
2. **EntityDB Metrics**: Entity count, tag operations, WAL size
3. **Recovery Metrics**: Recovery attempts, patterns, success rates
4. **API Performance**: Response times, error rates, throughput
5. **Log Analysis**: Error patterns, warning trends, recovery messages

## Test Phases

### Phase 1: Baseline Establishment (0-30 minutes)
- **Objective**: Establish stable baseline metrics
- **Activities**: 
  - Start monitoring systems
  - Perform basic CRUD operations  
  - Verify metrics collection functioning
  - Confirm no recovery loops present

### Phase 2: Normal Operations (30 minutes - 6 hours)
- **Objective**: Validate performance under typical load
- **Activities**:
  - Background metrics collection (every 1 minute)
  - Periodic API health checks
  - Simulated user operations (entity creation/updates)
  - Monitor for any performance degradation patterns

### Phase 3: Stress Operations (6-8 hours)
- **Objective**: Test system under higher load
- **Activities**:
  - Increased API request frequency
  - Bulk entity operations
  - Multiple concurrent sessions
  - Verify intelligent recovery handles load spikes

### Phase 4: Extended Soak (8-12 hours)
- **Objective**: Long-term stability validation
- **Activities**:
  - Return to normal operation levels
  - Monitor for memory leaks
  - Verify WAL management functioning
  - Confirm sustained performance

## Automated Test Execution

### Master Control Script
```bash
#!/bin/bash
# /opt/entitydb/execute_12hour_stability_test.sh

# Test configuration
TEST_START=$(date +%s)
TEST_DURATION=43200  # 12 hours in seconds
LOG_DIR="/opt/entitydb/test-results/$(date +%Y%m%d-%H%M%S)"

# Create test environment
mkdir -p "$LOG_DIR"
echo "Starting 12-hour stability test at $(date)"
echo "Results will be logged to: $LOG_DIR"

# Start monitoring systems
./monitoring/start_all_monitors.sh "$LOG_DIR"

# Execute test phases
./test-phases/phase1_baseline.sh "$LOG_DIR" &
./test-phases/phase2_normal_ops.sh "$LOG_DIR" &  
./test-phases/phase3_stress_ops.sh "$LOG_DIR" &
./test-phases/phase4_extended_soak.sh "$LOG_DIR" &

# Wait for test completion
sleep $TEST_DURATION

# Generate final report
./generate_stability_report.sh "$LOG_DIR"
echo "12-hour stability test completed at $(date)"
```

### Monitoring Scripts Implementation

#### `/opt/entitydb/monitoring/cpu_monitor.sh`
```bash
#!/bin/bash
LOG_DIR="$1"
while true; do
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    load_avg=$(cat /proc/loadavg | cut -d' ' -f1-3)
    cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
    echo "$timestamp,CPU,$load_avg,$cpu_usage" >> "$LOG_DIR/cpu_metrics.csv"
    sleep 30
done
```

#### `/opt/entitydb/monitoring/entitydb_monitor.sh`
```bash
#!/bin/bash
LOG_DIR="$1"
while true; do
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Check if EntityDB is running
    if pgrep entitydb > /dev/null; then
        # Get EntityDB process stats
        pid=$(pgrep entitydb)
        cpu_percent=$(ps -p $pid -o %cpu --no-headers)
        mem_percent=$(ps -p $pid -o %mem --no-headers)
        
        # Check log for recovery attempts (last 2 minutes)
        recovery_attempts=$(tail -n 100 /opt/entitydb/var/entitydb.log | \
                          grep "$(date -d '2 minutes ago' '+%Y/%m/%d %H:%M')" | \
                          grep -c "attempting to recover corrupted entity" || echo 0)
        
        # Get WAL file size if exists
        wal_size=0
        if [ -f /opt/entitydb/var/entities.edb ]; then
            wal_size=$(stat -c%s /opt/entitydb/var/entities.edb)
        fi
        
        echo "$timestamp,ENTITYDB,running,$cpu_percent,$mem_percent,$recovery_attempts,$wal_size" >> "$LOG_DIR/entitydb_metrics.csv"
    else
        echo "$timestamp,ENTITYDB,stopped,0,0,0,0" >> "$LOG_DIR/entitydb_metrics.csv"
    fi
    sleep 120
done
```

#### `/opt/entitydb/monitoring/api_health_check.sh`
```bash
#!/bin/bash
LOG_DIR="$1"
while true; do
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Test health endpoint
    start_time=$(date +%s%3N)
    if curl -k -s https://localhost:8085/health > /dev/null; then
        end_time=$(date +%s%3N)
        response_time=$((end_time - start_time))
        echo "$timestamp,HEALTH,success,$response_time" >> "$LOG_DIR/api_health.csv"
    else
        echo "$timestamp,HEALTH,failed,0" >> "$LOG_DIR/api_health.csv"
    fi
    
    sleep 300  # Every 5 minutes
done
```

### Test Phase Scripts

#### `/opt/entitydb/test-phases/phase1_baseline.sh`
```bash
#!/bin/bash
LOG_DIR="$1"
echo "Phase 1: Baseline establishment (30 minutes)" >> "$LOG_DIR/test_phases.log"

# Wait for monitors to establish baseline
sleep 1800  # 30 minutes

echo "Phase 1 completed" >> "$LOG_DIR/test_phases.log"
```

#### `/opt/entitydb/test-phases/phase2_normal_ops.sh`
```bash
#!/bin/bash
LOG_DIR="$1"
sleep 1800  # Wait for phase 1

echo "Phase 2: Normal operations (5.5 hours)" >> "$LOG_DIR/test_phases.log"

# Simulate normal API usage
end_time=$(($(date +%s) + 19800))  # 5.5 hours
while [ $(date +%s) -lt $end_time ]; do
    # Create test entity every 10 minutes
    curl -k -s -X POST https://localhost:8085/api/v1/entities/create \
         -H "Content-Type: application/json" \
         -d '{"tags":["type:test","phase:2"],"content":"dGVzdCBkYXRh"}' >> "$LOG_DIR/api_operations.log" 2>&1
    sleep 600
done

echo "Phase 2 completed" >> "$LOG_DIR/test_phases.log"
```

### Alert Thresholds
- **Critical**: CPU load > 3.0 sustained for 5+ minutes
- **Warning**: CPU load > 1.5 sustained for 10+ minutes  
- **Critical**: Recovery attempts > 10 in any 5-minute window
- **Warning**: API response time > 5000ms
- **Critical**: EntityDB process stopped unexpectedly

## Expected Results

### CPU Performance
- **Target**: Load average < 1.5 throughout test
- **Alert Threshold**: Load > 3.0 for >5 minutes
- **Pattern**: Stable performance without spikes during metrics collection

### Memory Usage  
- **Target**: Stable memory usage, no leaks
- **Alert Threshold**: Memory growth >50MB/hour sustained
- **Pattern**: Minor fluctuations normal, major trends concerning

### Recovery System
- **Target**: 0 recovery attempts for metric entities
- **Alert Threshold**: >5 recovery attempts/hour
- **Pattern**: Only legitimate entities should trigger recovery

### API Responsiveness
- **Target**: <500ms average response time
- **Alert Threshold**: >2000ms average for 10+ minutes
- **Pattern**: Consistent performance throughout test

## Failure Scenarios & Responses

### Scenario 1: CPU Spike Detection
```bash
if load_avg > 3.0 for 5+ minutes:
    - Capture detailed process list
    - Analyze EntityDB logs for recovery patterns
    - Check metrics collection timing
    - Consider test termination if sustained
```

### Scenario 2: Memory Leak Detection  
```bash
if memory_growth > 50MB/hour sustained:
    - Capture memory profile
    - Analyze garbage collection patterns
    - Monitor file handle usage
    - Document leak progression
```

### Scenario 3: Recovery Loop Return
```bash
if recovery_attempts > threshold:
    - Immediate log analysis 
    - Identify entity patterns causing recovery
    - Verify intelligent recovery logic functioning
    - Potential emergency stop
```

## Test Execution Command

To execute this comprehensive 12-hour test with full automation:

```bash
# Ensure EntityDB is running and stable
cd /opt/entitydb

# Create monitoring infrastructure  
mkdir -p monitoring test-phases
# [Scripts will be created during setup phase]

# Execute the comprehensive test
nohup ./execute_12hour_stability_test.sh > 12hour_test.log 2>&1 &

# Monitor progress
tail -f 12hour_test.log
```

## Post-Test Analysis

The test will automatically generate:
1. **Performance Report**: CPU, memory, and API metrics summary
2. **Recovery Analysis**: Pattern analysis of any recovery attempts
3. **Trend Analysis**: Performance trends over the 12-hour period
4. **Stability Score**: Overall system stability rating (0-100)
5. **Recommendations**: Any identified optimization opportunities

This plan ensures comprehensive validation of the intelligent recovery system under real-world conditions while providing detailed data for performance analysis.