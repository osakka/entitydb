#!/bin/bash
# entitydbd.sh - Daemon Controller for the EntityDB server (Consolidated Entity-based Architecture)

# Configuration
EntityDB_DIR="$(dirname "$0")/.."

# Load default environment configuration from share/config
DEFAULT_ENV_FILE="$EntityDB_DIR/share/config/entitydb_server.env"
if [ -f "$DEFAULT_ENV_FILE" ]; then
    echo "Loading defaults from $DEFAULT_ENV_FILE"
    source "$DEFAULT_ENV_FILE"
fi

# Override with instance-specific env file in var directory
INSTANCE_ENV_FILE="$EntityDB_DIR/var/entitydb.env"
if [ -f "$INSTANCE_ENV_FILE" ]; then
    echo "Loading instance config from $INSTANCE_ENV_FILE"
    source "$INSTANCE_ENV_FILE"
fi

# Export all ENTITYDB_ variables so they're available to the server
for var in $(env | grep ^ENTITYDB_ | cut -d= -f1); do
    export "$var"
done

# Set defaults (these can be overridden by environment variables)
PID_FILE="${ENTITYDB_PID_FILE:-$EntityDB_DIR/var/entitydb.pid}"
LOG_FILE="${ENTITYDB_LOG_FILE:-$EntityDB_DIR/var/entitydb.log}"
SSL_PORT="${ENTITYDB_SSL_PORT:-8085}"
HOST="${ENTITYDB_HOST:-0.0.0.0}"
DB_PATH="${ENTITYDB_DATA_PATH:-$EntityDB_DIR/var}"
STATIC_DIR="${ENTITYDB_STATIC_DIR:-$EntityDB_DIR/share/htdocs}"
SSL_CERT="${ENTITYDB_SSL_CERT:-/etc/ssl/certs/server.pem}"
SSL_KEY="${ENTITYDB_SSL_KEY:-/etc/ssl/private/server.key}"

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

# Function to initialize database with default admin user
init_database() {
    print_message "$BLUE" "Checking database initialization..."
    
    # Determine the correct URL based on SSL setting
    if [ "$ENTITYDB_USE_SSL" = "true" ]; then
        LOGIN_URL="https://localhost:$SSL_PORT/api/v1/auth/login"
        TEST_URL="https://localhost:$SSL_PORT/api/v1/test/entities/create"
        CURL_OPTS="-k"
    else
        LOGIN_URL="http://localhost:${ENTITYDB_PORT:-8085}/api/v1/auth/login"
        TEST_URL="http://localhost:${ENTITYDB_PORT:-8085}/api/v1/test/entities/create"
        CURL_OPTS=""
    fi
    
    # Try to login with default credentials
    LOGIN=$(curl $CURL_OPTS -s -X POST "$LOGIN_URL" \
      -H "Content-Type: application/json" \
      -d '{"username":"admin","password":"admin"}')
    
    TOKEN=$(echo "$LOGIN" | jq -r '.token' 2>/dev/null)
    
    if [ "$TOKEN" != "null" ] && [ ! -z "$TOKEN" ]; then
        print_message "$GREEN" "Default admin user already exists and is working."
        return 0
    fi
    
    print_message "$BLUE" "Default admin user not found or not working. Creating..."
    
    # Generate bcrypt hash for "admin" password
    cd /opt/entitydb/src
    cat > generate_admin_hash.go << 'EOF'
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
    fmt.Println(string(hash))
}
EOF
    
    HASH=$(go run generate_admin_hash.go)
    rm generate_admin_hash.go
    cd - > /dev/null
    
    # Create admin user using test endpoint (without fixed ID)
    # Using the new entity format with single content field
    USER_DATA="{\"username\":\"admin\",\"password_hash\":\"$HASH\",\"display_name\":\"Administrator\"}"
    ENCODED_DATA=$(echo -n "$USER_DATA" | base64 -w0)
    
    RESPONSE=$(curl $CURL_OPTS -s -X POST "$TEST_URL" \
      -H "Content-Type: application/json" \
      -d "{
        \"tags\": [
          \"type:user\",
          \"id:username:admin\",
          \"rbac:role:admin\",
          \"rbac:perm:*\",
          \"status:active\"
        ],
        \"content\": {
          \"data\": \"$ENCODED_DATA\",
          \"type\": \"application/json\"
        }
      }")
    
    # Verify creation
    LOGIN=$(curl $CURL_OPTS -s -X POST "$LOGIN_URL" \
      -H "Content-Type: application/json" \
      -d '{"username":"admin","password":"admin"}')
    
    TOKEN=$(echo "$LOGIN" | jq -r '.token' 2>/dev/null)
    
    if [ "$TOKEN" != "null" ] && [ ! -z "$TOKEN" ]; then
        print_message "$GREEN" "✅ Default admin user created successfully!"
        print_message "$GREEN" "   Username: admin"
        print_message "$GREEN" "   Password: admin"
    else
        print_message "$RED" "❌ Failed to create admin user"
        return 1
    fi
}

# Find the server binary
ENTITY_SERVER_BIN="$EntityDB_DIR/bin/entitydb"

# Check if the binary exists
if [ -x "$ENTITY_SERVER_BIN" ]; then
    ACTIVE_SERVER="$ENTITY_SERVER_BIN"
    SERVER_TYPE="consolidated_entity"
    print_message "$GREEN" "Using consolidated entity server with integrated static file support: $ACTIVE_SERVER"
else
    print_message "$RED" "Server binary not found at $ENTITY_SERVER_BIN"
    print_message "$YELLOW" "Make sure to build the server using 'cd $EntityDB_DIR/src && make server'"
    exit 1
fi

# Create necessary directories
mkdir -p "$(dirname "$PID_FILE")" "$(dirname "$LOG_FILE")" "$(dirname "$DB_PATH")"

# Function to display usage
usage() {
    print_message "$BLUE" "EntityDB Server Daemon Controller"
    print_message "$BLUE" "Usage: $0 {start|stop|restart|status}"
    echo ""
    print_message "$BLUE" "Commands:"
    echo "  start    - Start the EntityDB server daemon"
    echo "  stop     - Stop the EntityDB server daemon"
    echo "  restart  - Restart the EntityDB server daemon"
    echo "  status   - Check if the EntityDB server daemon is running"
    echo ""
    print_message "$YELLOW" "Note: This script only controls the server daemon."
    print_message "$YELLOW" "To build the server, use: cd $(dirname "$0")/../src && make server"
    echo ""
    exit 1
}

# Function to start the server
start_server() {
    print_message "$BLUE" "Starting EntityDB Consolidated Entity Server..."

    # Check if server is already running
    if [ -f "$PID_FILE" ] && ps -p "$(cat "$PID_FILE")" > /dev/null; then
        print_message "$YELLOW" "EntityDB Consolidated Entity Server is already running (PID: $(cat "$PID_FILE"))"
        return 0
    fi

    # Verify server binary exists
    if [ ! -x "$ACTIVE_SERVER" ]; then
        print_message "$RED" "Server binary not found or not executable"
        print_message "$YELLOW" "Please build the server using: cd $EntityDB_DIR/src && make server"
        return 1
    fi
    
    # Only verify SSL certificates if SSL is enabled
    if [ "$ENTITYDB_USE_SSL" = "true" ]; then
        if [ ! -f "$SSL_CERT" ]; then
            print_message "$RED" "SSL certificate not found: $SSL_CERT"
            print_message "$YELLOW" "Please install certificate or generate a self-signed one"
            return 1
        fi
        
        if [ ! -f "$SSL_KEY" ]; then
            print_message "$RED" "SSL private key not found: $SSL_KEY"
            print_message "$YELLOW" "Please install private key or generate a self-signed one"
            return 1
        fi
    fi

    # Build command line arguments from environment
    CMD_ARGS=""
    [ -n "$ENTITYDB_USE_SSL" ] && [ "$ENTITYDB_USE_SSL" = "true" ] && CMD_ARGS="$CMD_ARGS --use-ssl"
    [ -n "$SSL_CERT" ] && CMD_ARGS="$CMD_ARGS --ssl-cert $SSL_CERT"
    [ -n "$SSL_KEY" ] && CMD_ARGS="$CMD_ARGS --ssl-key $SSL_KEY"
    [ -n "$SSL_PORT" ] && CMD_ARGS="$CMD_ARGS --ssl-port $SSL_PORT"
    [ -n "$DB_PATH" ] && CMD_ARGS="$CMD_ARGS -data $DB_PATH"
    [ -n "$STATIC_DIR" ] && CMD_ARGS="$CMD_ARGS -static-dir $STATIC_DIR"
    [ -n "$ENTITYDB_PORT" ] && CMD_ARGS="$CMD_ARGS -port $ENTITYDB_PORT"
    [ -n "$ENTITYDB_LOG_LEVEL" ] && CMD_ARGS="$CMD_ARGS -log-level $ENTITYDB_LOG_LEVEL"
    [ -n "$ENTITYDB_TOKEN_SECRET" ] && CMD_ARGS="$CMD_ARGS -token-secret $ENTITYDB_TOKEN_SECRET"
    
    # Start the server
    "$ACTIVE_SERVER" $CMD_ARGS > "$LOG_FILE" 2>&1 &

    # Save PID
    echo $! > "$PID_FILE"

    # Check if server started successfully
    sleep 2
    if [ -f "$PID_FILE" ] && ps -p "$(cat "$PID_FILE")" > /dev/null; then
        print_message "$GREEN" "EntityDB Consolidated Entity Server started successfully (PID: $(cat "$PID_FILE"))"
        
        # Display the correct URL based on SSL setting
        if [ "$ENTITYDB_USE_SSL" = "true" ]; then
            print_message "$GREEN" "Server is running at https://$HOST:$SSL_PORT"
            print_message "$BLUE" "API Documentation: https://$HOST:$SSL_PORT/api/v1/status"
            print_message "$BLUE" "Dashboard: https://$HOST:$SSL_PORT/"
            
            # Display SSL certificate info
            if command -v openssl >/dev/null 2>&1; then
                CERT_INFO=$(openssl x509 -in "$SSL_CERT" -noout -subject -enddate 2>/dev/null)
                if [ $? -eq 0 ]; then
                    print_message "$BLUE" "SSL Certificate Info:"
                    echo "$CERT_INFO" | sed 's/^/  /'
                fi
            fi
        else
            print_message "$GREEN" "Server is running at http://$HOST:${ENTITYDB_PORT:-8085}"
            print_message "$BLUE" "API Documentation: http://$HOST:${ENTITYDB_PORT:-8085}/api/v1/status"
            print_message "$BLUE" "Dashboard: http://$HOST:${ENTITYDB_PORT:-8085}/"
        fi
        
        # Initialize database if needed
        init_database
    else
        print_message "$RED" "Failed to start EntityDB Consolidated Entity Server. Check logs at $LOG_FILE"
        cat "$LOG_FILE"
        rm -f "$PID_FILE"
        return 1
    fi
}

# Function to stop the server
stop_server() {
    print_message "$BLUE" "Stopping EntityDB Consolidated Entity Server..."

    # Check if PID file exists
    if [ ! -f "$PID_FILE" ]; then
        print_message "$YELLOW" "PID file not found, server is not running or was not properly started"
        return 0
    fi

    # Get PID from file
    PID=$(cat "$PID_FILE")

    # Check if process is running
    if ! ps -p "$PID" > /dev/null; then
        print_message "$YELLOW" "Process with PID $PID is not running"
        rm -f "$PID_FILE"
        return 0
    fi

    # Send SIGTERM to process
    kill "$PID"

    # Wait for process to terminate
    print_message "$BLUE" "Waiting for server to stop..."
    for i in {1..10}; do
        if ! ps -p "$PID" > /dev/null; then
            break
        fi
        sleep 1
    done

    # Force kill if necessary
    if ps -p "$PID" > /dev/null; then
        print_message "$YELLOW" "Server did not stop gracefully, forcing termination..."
        kill -9 "$PID"
        sleep 1
    fi

    # Remove PID file
    rm -f "$PID_FILE"

    print_message "$GREEN" "EntityDB Consolidated Entity Server stopped"
}

# Function to check server status
check_status() {
    if [ -f "$PID_FILE" ] && ps -p "$(cat "$PID_FILE")" > /dev/null; then
        print_message "$GREEN" "EntityDB Consolidated Entity Server is running (PID: $(cat "$PID_FILE"))"
        print_message "$BLUE" "Server address: https://$HOST:$SSL_PORT"
        print_message "$BLUE" "API Status: https://$HOST:$SSL_PORT/api/v1/status"
        print_message "$BLUE" "Dashboard: https://$HOST:$SSL_PORT/"

        # Check API status using curl
        if command -v curl >/dev/null 2>&1; then
            status_response=$(curl -k -s "https://$HOST:$SSL_PORT/api/v1/status")
            if [[ $status_response == *"\"status\":\"ok\""* ]]; then
                print_message "$GREEN" "Server API is responding normally"
                print_message "$BLUE" "API Mode: Consolidated Entity-based Architecture"
            else
                print_message "$YELLOW" "Server API response may have issues:"
                echo "$status_response"
            fi
        fi

        return 0
    else
        print_message "$RED" "EntityDB Consolidated Entity Server is not running"
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
        print_message "$BLUE" "=== EntityDB Consolidated Entity Server ==="
        print_message "$BLUE" "Starting server daemon"
        print_message "$BLUE" "==========================================="
        start_server
        ;;
    stop)
        stop_server
        ;;
    restart)
        print_message "$BLUE" "=== EntityDB Consolidated Entity Server ==="
        print_message "$BLUE" "Restarting server daemon"
        print_message "$BLUE" "==========================================="
        stop_server
        sleep 2
        start_server
        ;;
    status)
        print_message "$BLUE" "=== EntityDB Consolidated Entity Server Status ==="
        check_status
        ;;
    *)
        usage
        ;;
esac

exit 0
