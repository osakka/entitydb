#!/bin/bash

# EntityDB Memory Guardian
# Monitors server memory usage and kills process if it exceeds threshold

PID_FILE="/opt/entitydb/var/entitydb.pid"
LOG_FILE="/opt/entitydb/var/memory-guardian.log"
MEMORY_THRESHOLD=80  # Kill if memory usage exceeds 80%
CHECK_INTERVAL=5     # Check every 5 seconds

log() {
    echo "$(date '+%Y/%m/%d %H:%M:%S') [GUARDIAN] $1" | tee -a "$LOG_FILE"
}

get_memory_usage() {
    local pid=$1
    if [ -z "$pid" ] || ! kill -0 "$pid" 2>/dev/null; then
        echo "0"
        return
    fi
    
    # Get memory usage percentage for the process
    ps -p "$pid" -o %mem --no-headers | awk '{print int($1)}'
}

kill_server_safely() {
    local pid=$1
    log "CRITICAL: Memory usage exceeded ${MEMORY_THRESHOLD}%, killing server (PID: $pid)"
    
    # Try graceful shutdown first
    if kill -TERM "$pid" 2>/dev/null; then
        log "Sent SIGTERM to process $pid"
        sleep 3
        
        # Check if process is still running
        if kill -0 "$pid" 2>/dev/null; then
            log "Process still running, sending SIGKILL"
            kill -KILL "$pid" 2>/dev/null
        fi
    else
        log "Failed to send SIGTERM, trying SIGKILL"
        kill -KILL "$pid" 2>/dev/null
    fi
    
    # Clean up PID file
    rm -f "$PID_FILE"
    log "Server killed and PID file cleaned up"
}

main() {
    log "Memory Guardian started (threshold: ${MEMORY_THRESHOLD}%, interval: ${CHECK_INTERVAL}s)"
    
    while true; do
        if [ -f "$PID_FILE" ]; then
            PID=$(cat "$PID_FILE" 2>/dev/null)
            
            if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then
                MEMORY_USAGE=$(get_memory_usage "$PID")
                
                if [ "$MEMORY_USAGE" -gt "$MEMORY_THRESHOLD" ]; then
                    kill_server_safely "$PID"
                    break
                elif [ "$MEMORY_USAGE" -gt 50 ]; then
                    # Log warning at 50%+ usage
                    log "Warning: Memory usage at ${MEMORY_USAGE}% for PID $PID"
                fi
            else
                log "Server process not found, guardian exiting"
                break
            fi
        else
            log "PID file not found, guardian exiting"
            break
        fi
        
        sleep "$CHECK_INTERVAL"
    done
    
    log "Memory Guardian stopped"
}

# Handle signals
trap 'log "Guardian received signal, stopping"; exit 0' SIGTERM SIGINT

main "$@"