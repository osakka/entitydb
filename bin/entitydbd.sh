#!/bin/bash
# entitydbd.sh - Simplified Daemon Controller for EntityDB server
# Uses ConfigManager for all configuration logic - no duplication!

# Determine EntityDB directory
# EntityDB_DIR="$(dirname "$0")/.."
EntityDB_DIR="$(dirname "$(realpath "../$0")")"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function for colored output
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Load environment configurations (ConfigManager handles the rest)
load_environment() {
    # Load default environment configuration
    DEFAULT_ENV_FILE="$EntityDB_DIR/share/config/entitydb.env"
    if [ -f "$DEFAULT_ENV_FILE" ]; then
        print_message "$BLUE" "Loading defaults from $DEFAULT_ENV_FILE"
        source "$DEFAULT_ENV_FILE"
    fi
    
    # Override with instance-specific env file
    INSTANCE_ENV_FILE="$EntityDB_DIR/var/entitydb.env"
    if [ -f "$INSTANCE_ENV_FILE" ]; then
        print_message "$BLUE" "Loading instance config from $INSTANCE_ENV_FILE"
        source "$INSTANCE_ENV_FILE"
    fi
    
    # Export all ENTITYDB_ variables for the Go process
    # After sourcing the config files, we need to export the variables
    # so they are available to child processes
    for var in $(set | grep '^ENTITYDB_' | cut -d= -f1); do
        export "$var"
    done
}

# Get configuration values (using environment variables with fallbacks)
get_pid_file() {
    echo "${ENTITYDB_PID_FILE:-$EntityDB_DIR/var/entitydb.pid}"
}

get_log_file() {
    echo "${ENTITYDB_LOG_FILE:-$EntityDB_DIR/var/entitydb.log}"
}

get_server_url() {
    if [ "$ENTITYDB_USE_SSL" = "true" ]; then
        echo "https://localhost:${ENTITYDB_SSL_PORT:-8085}"
    else
        echo "http://localhost:${ENTITYDB_PORT:-8085}"
    fi
}

# Find the server binary
find_server_binary() {
    ENTITY_SERVER_BIN="$EntityDB_DIR/bin/entitydb"
    
    if [ -x "$ENTITY_SERVER_BIN" ]; then
        echo "$ENTITY_SERVER_BIN"
        return 0
    else
        print_message "$RED" "Server binary not found at $ENTITY_SERVER_BIN"
        print_message "$YELLOW" "Make sure to build the server using 'cd $EntityDB_DIR/src && make'"
        return 1
    fi
}

# Function to display usage
usage() {
    print_message "$BLUE" "EntityDB Server Daemon Controller (Simplified)"
    print_message "$BLUE" "Usage: $0 {start|stop|restart|status}"
    echo ""
    print_message "$BLUE" "Commands:"
    echo "  start    - Start the EntityDB server daemon"
    echo "  stop     - Stop the EntityDB server daemon"  
    echo "  restart  - Restart the EntityDB server daemon"
    echo "  status   - Check if the EntityDB server daemon is running"
    echo ""
    print_message "$YELLOW" "Note: All configuration is handled by ConfigManager in the Go binary."
    print_message "$YELLOW" "Set environment variables in share/config/entitydb.env or var/entitydb.env"
    echo ""
    exit 1
}

# Function to start the server
start_server() {
    print_message "$BLUE" "Starting EntityDB Server..."
    
    # Load environment
    load_environment
    
    # Get configuration values
    PID_FILE=$(get_pid_file)
    LOG_FILE=$(get_log_file)
    
    # Find server binary
    SERVER_BIN=$(find_server_binary)
    if [ $? -ne 0 ]; then
        return 1
    fi
    
    # Check if server is already running
    if [ -f "$PID_FILE" ] && ps -p "$(cat "$PID_FILE")" > /dev/null 2>&1; then
        print_message "$YELLOW" "EntityDB Server is already running (PID: $(cat "$PID_FILE"))"
        return 0
    fi
    
    # Create necessary directories
    mkdir -p "$(dirname "$PID_FILE")" "$(dirname "$LOG_FILE")"
    
    # Start the server - ConfigManager handles all configuration logic!
    # No need to build command line flags - environment variables are enough
    print_message "$BLUE" "Starting server with ConfigManager handling all configuration..."
    
    # Ensure environment variables are properly passed to the spawned process
    # by explicitly running with env to pass current environment
    env "$SERVER_BIN" > "$LOG_FILE" 2>&1 &
    
    # Save PID
    SERVER_PID=$!
    echo $SERVER_PID > "$PID_FILE"
    
    # Check if server started successfully
    sleep 2
    if [ -f "$PID_FILE" ] && ps -p "$(cat "$PID_FILE")" > /dev/null 2>&1; then
        print_message "$GREEN" "EntityDB Server started successfully (PID: $(cat "$PID_FILE"))"
        print_message "$GREEN" "Server URL: $(get_server_url)"
        print_message "$BLUE" "Dashboard: $(get_server_url)/"
        print_message "$BLUE" "API Status: $(get_server_url)/api/v1/status"
        print_message "$BLUE" "Logs: $LOG_FILE"
    else
        print_message "$RED" "Failed to start EntityDB Server. Check logs at $LOG_FILE"
        [ -f "$LOG_FILE" ] && tail -10 "$LOG_FILE"
        rm -f "$PID_FILE"
        return 1
    fi
}

# Function to stop the server
stop_server() {
    print_message "$BLUE" "Stopping EntityDB Server..."
    
    # Load environment to get PID file location
    load_environment
    PID_FILE=$(get_pid_file)
    
    # Check if PID file exists
    if [ ! -f "$PID_FILE" ]; then
        print_message "$YELLOW" "PID file not found, server is not running or was not properly started"
        return 0
    fi
    
    # Get PID from file
    PID=$(cat "$PID_FILE")
    
    # Check if process is running
    if ! ps -p "$PID" > /dev/null 2>&1; then
        print_message "$YELLOW" "Process with PID $PID is not running"
        rm -f "$PID_FILE"
        return 0
    fi
    
    # Send SIGTERM to process
    print_message "$BLUE" "Sending SIGTERM to PID $PID..."
    kill "$PID"
    
    # Wait for process to terminate gracefully
    print_message "$BLUE" "Waiting for server to stop..."
    for i in {1..10}; do
        if ! ps -p "$PID" > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    
    # Force kill if necessary
    if ps -p "$PID" > /dev/null 2>&1; then
        print_message "$YELLOW" "Server did not stop gracefully, forcing termination..."
        kill -9 "$PID"
        sleep 1
    fi
    
    # Remove PID file
    rm -f "$PID_FILE"
    print_message "$GREEN" "EntityDB Server stopped"
}

# Function to check server status
check_status() {
    # Load environment to get configuration
    load_environment
    PID_FILE=$(get_pid_file)
    SERVER_URL=$(get_server_url)
    
    if [ -f "$PID_FILE" ] && ps -p "$(cat "$PID_FILE")" > /dev/null 2>&1; then
        print_message "$GREEN" "EntityDB Server is running (PID: $(cat "$PID_FILE"))"
        print_message "$BLUE" "Server URL: $SERVER_URL"
        print_message "$BLUE" "Dashboard: $SERVER_URL/"
        print_message "$BLUE" "API Status: $SERVER_URL/api/v1/status"
        
        # Check API status using curl if available
        if command -v curl >/dev/null 2>&1; then
            CURL_OPTS=""
            [ "$ENTITYDB_USE_SSL" = "true" ] && CURL_OPTS="-k"
            
            status_response=$(curl $CURL_OPTS -s "$SERVER_URL/api/v1/status" 2>/dev/null)
            if [[ $status_response == *"\"status\":\"ok\""* ]]; then
                print_message "$GREEN" "✅ Server API is responding normally"
            else
                print_message "$YELLOW" "⚠️  Server API response may have issues"
            fi
        fi
        
        return 0
    else
        print_message "$RED" "EntityDB Server is not running"
        # Remove stale PID file if it exists
        [ -f "$PID_FILE" ] && rm -f "$PID_FILE"
        return 1
    fi
}

# Check command
if [ $# -eq 0 ]; then
    usage
fi

# Process command
case "$1" in
    start)
        print_message "$BLUE" "=== EntityDB Server Daemon ==="
        print_message "$BLUE" "Using ConfigManager for all configuration"
        print_message "$BLUE" "=============================="
        start_server
        ;;
    stop)
        stop_server
        ;;
    restart)
        print_message "$BLUE" "=== EntityDB Server Restart ==="
        stop_server
        sleep 2
        start_server
        ;;
    status)
        print_message "$BLUE" "=== EntityDB Server Status ==="
        check_status
        ;;
    *)
        usage
        ;;
esac

exit $?
