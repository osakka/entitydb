#!/bin/bash
# Test script for EntityDB server login with security components

# Kill any existing server instances
echo "Killing any existing server instances..."
pkill -f "entitydb_server_entity" || true
sleep 1

# Ensure audit log directory exists
mkdir -p /opt/entitydb/var/log/audit

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

# Test successful login
echo -e "\nTesting login with correct credentials..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8087/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}')

echo "$LOGIN_RESPONSE" | jq .

# Extract token from response
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')
echo "Token: $TOKEN"

# Test login with invalid password
echo -e "\nTesting login with invalid password..."
curl -s -X POST http://localhost:8087/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"wrongpassword"}' | jq .

# Test login with invalid username
echo -e "\nTesting login with invalid username..."
curl -s -X POST http://localhost:8087/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"nonexistent","password":"password"}' | jq .

# Test login with missing password
echo -e "\nTesting login with missing password..."
curl -s -X POST http://localhost:8087/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin"}' | jq .

# Check audit log
echo -e "\nChecking audit log..."
ls -la /opt/entitydb/var/log/audit/
echo "Latest audit log entries:"
find /opt/entitydb/var/log/audit/ -name "audit_*.log" -type f -exec ls -t {} \; | head -1 | xargs tail -n 10

# Stop the entity server
echo -e "\nStopping EntityDB entity server..."
kill $SERVER_PID || true
wait $SERVER_PID 2>/dev/null || true

echo "Test completed successfully."