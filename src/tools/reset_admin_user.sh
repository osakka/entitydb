#\!/bin/bash

echo "=== FINAL ADMIN FIX ==="
echo

# Stop the server
echo "1. Stopping server..."
/opt/entitydb/bin/entitydbd.sh stop

# Run the create working admin tool
echo -e "\n2. Creating working admin user..."
cd /opt/entitydb/src
go run tools/create_working_admin.go 2>&1  < /dev/null |  grep -v "^2025"

# Start the server
echo -e "\n3. Starting server..."
/opt/entitydb/bin/entitydbd.sh start

# Wait for server to be ready
echo -e "\n4. Waiting for server to be ready..."
sleep 3

# Test authentication
echo -e "\n5. Testing authentication..."
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' 2>&1

echo -e "\n\n6. Testing RBAC metrics endpoint..."
# Get the token if login works
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ \! -z "$TOKEN" ]; then
  echo "Got token: $TOKEN"
  echo "Testing RBAC metrics..."
  curl -H "Authorization: Bearer $TOKEN" http://localhost:8085/api/v1/rbac/metrics 2>&1 | head -5
else
  echo "No token received - authentication failed"
fi
