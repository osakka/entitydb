#!/bin/bash

# Test login with debug logging

echo "Testing login..."
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# Check the logs
echo -e "\n\nChecking recent debug logs..."
tail -20 /opt/entitydb/var/entitydb.log | grep -i "debug\|error"