#!/bin/bash

# Direct login test
echo "Testing direct entity access..."

# Check if admin user exists
curl -X GET "http://localhost:8085/api/v1/entities/test/create" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Direct Login",
    "tags": ["test:direct"]
  }' | jq .

echo -e "\nChecking admin user in logs..."
grep -i "admin" /opt/entitydb/var/entitydb.log | tail -5