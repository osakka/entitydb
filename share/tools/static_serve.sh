#!/bin/bash

# Simple script to serve the UI files for testing
# This uses Python's built-in HTTP server

# Default port
PORT=8086

# Check if port is provided as argument
if [ ! -z "$1" ]; then
  PORT=$1
fi

# Kill any existing static server
pkill -f "python3 -m http.server $PORT" 2>/dev/null

# Get the directory to serve
HTDOCS_DIR="/opt/entitydb/share/htdocs"

echo "Starting static server for EntityDB UI on port $PORT..."
echo "Serving directory: $HTDOCS_DIR"
echo "Access URL: http://localhost:$PORT"

# Start the server using Python's built-in HTTP server
cd $HTDOCS_DIR
nohup python3 -m http.server $PORT > /tmp/static_server.log 2>&1 &

# Make the script executable by default
chmod +x /opt/entitydb/share/tools/static_serve.sh

echo "Server started! Access your UI at: http://localhost:$PORT/"
echo "View logs at: /tmp/static_server.log"
echo "To stop the server, run: pkill -f \"python3 -m http.server $PORT\""