#!/bin/bash

echo "=== Simple Login Test ==="

# Set debug mode
export ENTITYDB_LOG_LEVEL=DEBUG

# Stop server
cd /opt/entitydb
./bin/entitydbd.sh stop

# Clean database
rm -f var/*.ebf var/*.wal var/*.log

# Start fresh
./bin/entitydbd.sh start

sleep 3

echo "Testing login..."
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

echo -e "\n\nChecking logs..."
tail -20 var/entitydb.log