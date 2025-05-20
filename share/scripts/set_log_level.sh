#!/bin/bash
# Set EntityDB log level

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Path configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENTITYDB_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Check if log level is provided
if [ $# -ne 1 ]; then
    echo -e "${YELLOW}Usage: $0 <log-level>${NC}"
    echo "Available log levels: debug, info, warn, error"
    exit 1
fi

LOG_LEVEL=$1

# Validate log level
case $LOG_LEVEL in
    debug|info|warn|error)
        ;;
    *)
        echo -e "${RED}Invalid log level: $LOG_LEVEL${NC}"
        echo "Available log levels: debug, info, warn, error"
        exit 1
        ;;
esac

echo -e "${YELLOW}Setting EntityDB log level to: $LOG_LEVEL${NC}"

# Stop the server
echo "Stopping EntityDB server..."
"$ENTITYDB_DIR/bin/entitydbd.sh" stop

sleep 2

# Start with new log level
echo "Starting EntityDB with log level: $LOG_LEVEL"
"$ENTITYDB_DIR/bin/entitydb" \
    --use-ssl \
    --ssl-cert=/etc/ssl/certs/server.pem \
    --ssl-key=/etc/ssl/private/server.key \
    --ssl-port=8443 \
    --log-level=$LOG_LEVEL \
    --data="$ENTITYDB_DIR/var" \
    --static-dir="$ENTITYDB_DIR/share/htdocs" \
    > "$ENTITYDB_DIR/var/entitydb.log" 2>&1 &

PID=$!
echo $PID > "$ENTITYDB_DIR/var/entitydb.pid"

echo -e "${GREEN}EntityDB started with $LOG_LEVEL log level (PID: $PID)${NC}"
echo "Log file: $ENTITYDB_DIR/var/entitydb.log"
echo
echo "To view logs: tail -f $ENTITYDB_DIR/var/entitydb.log"