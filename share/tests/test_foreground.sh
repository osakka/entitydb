#!/bin/bash

# Kill any running servers
cd /opt/entitydb/bin && ./entitydbd.sh stop

# Start server in foreground
echo "Starting server in foreground..."
/opt/entitydb/bin/entitydb &
SERVER_PID=$!

sleep 3

echo -e "\n=== Testing endpoints ==="
echo "1. Testing /api/v1/status:"
curl -v http://localhost:8085/api/v1/status

echo -e "\n\n2. Testing /api/v1/test/status:"
curl -v http://localhost:8085/api/v1/test/status

echo -e "\n\n3. Testing /debug/ping:"
curl -v http://localhost:8085/debug/ping

echo -e "\n\n=== Killing server ==="
kill $SERVER_PID