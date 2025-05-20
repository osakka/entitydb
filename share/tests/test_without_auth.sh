#!/bin/bash

# Test without authentication

BASE_URL="https://localhost:8085"
echo "Testing basic functionality..."

# Check status endpoint
echo -e "\n=== Status Check ==="
curl -sk "$BASE_URL/api/v1/status"
echo

# Try to login with admin credentials
echo -e "\n=== Login Attempt ==="
LOGIN_RESPONSE=$(curl -sk -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' -v 2>&1)
  
echo "Login response:"
echo "$LOGIN_RESPONSE" | grep -E "HTTP|token|error"

# Check database files
echo -e "\n=== Database Files ==="
ls -la /opt/entitydb/var/*.ebf /opt/entitydb/var/*.wal 2>/dev/null

echo -e "\n✅ Test complete"