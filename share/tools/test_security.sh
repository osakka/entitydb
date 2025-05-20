#!/bin/bash
# Test script for EntityDB server with security components

# Kill any existing server instances
echo "Killing any existing server instances..."
pkill -f "entitydb_server_entity" || true
pkill -f "simple_server" || true
sleep 1

# Build the simplified server with security components
echo "Building simplified server with security components..."
cd /opt/entitydb/src
go build -o simple_server simplified_server.go security_manager.go security_types.go simple_security.go security_bridge.go security_input_audit.go

# Ensure audit log directory exists
mkdir -p /opt/entitydb/var/log/audit

# Start the server in the background
echo "Starting simplified server on port 8086..."
./simple_server -port 8086 &
SERVER_PID=$!

# Wait for server to start
sleep 2
echo "Server started with PID: $SERVER_PID"

# Test basic status endpoint
echo -e "\nTesting status endpoint..."
curl -s http://localhost:8086/ | jq .

# Check if audit logging is working
echo -e "\nChecking audit log..."
ls -la /opt/entitydb/var/log/audit/
echo "Latest audit log entries:"
find /opt/entitydb/var/log/audit/ -name "audit_*.log" -exec tail -n 5 {} \; 2>/dev/null || echo "No audit log entries yet"

# Stop the simplified server
echo -e "\nStopping simplified server..."
kill $SERVER_PID || true
wait $SERVER_PID 2>/dev/null || true
sleep 1

echo -e "\n=== Now testing the entity server with security components ==="

# Build the entity server with security components
echo "Building EntityDB entity server with security components..."
cd /opt/entitydb/src
go build -o entitydb_server_entity server_db.go security_manager.go security_types.go simple_security.go security_bridge.go security_input_audit.go

# Start the entity server in the background
echo "Starting EntityDB entity server on port 8087..."
./entitydb_server_entity -port 8087 &
SERVER_PID=$!

# Wait for server to start
sleep 2
echo "Server started with PID: $SERVER_PID"

# Test basic status endpoint
echo -e "\nTesting status endpoint..."
curl -s http://localhost:8087/api/v1/status | jq .

# Stop the entity server
echo -e "\nStopping EntityDB entity server..."
kill $SERVER_PID || true
wait $SERVER_PID 2>/dev/null || true

echo "Test completed successfully."