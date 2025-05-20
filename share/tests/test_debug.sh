#!/bin/bash

# Debug test script

BASE_URL="http://localhost:8085"

echo "Debug test..."

# Test simple endpoints
echo -e "\n=== Test API Status ==="
curl -v "$BASE_URL/api/v1/status"

echo -e "\n\n=== Test root ==="
curl -v "$BASE_URL/"

echo -e "\n\n=== Test auth endpoint ==="
curl -v "$BASE_URL/api/v1/auth/login"

echo -e "\n\n=== Test if server is running ==="
curl -v "$BASE_URL/api/v1/test/entities/create" \
  -H "Content-Type: application/json" \
  -d '{}'

echo -e "\n\n=== Server logs ==="
tail -20 /opt/entitydb/var/log/entitydb.log 2>/dev/null || echo "No log file found"