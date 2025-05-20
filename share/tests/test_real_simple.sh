#!/bin/bash

echo "Testing if server compiled correctly..."

# Start the server directly in foreground to see output
/opt/entitydb/bin/entitydb &
SERVER_PID=$!

sleep 3

echo "Trying simple status..."
curl -v http://localhost:8085/api/v1/status

echo -e "\n\nTrying test endpoint..."
curl -v http://localhost:8085/api/v1/test

echo -e "\n\nKilling server..."
kill $SERVER_PID