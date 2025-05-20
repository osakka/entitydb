#!/bin/bash
# Check how many entities are in the database

# Login as admin
SESSION=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

if [ -z "$SESSION" ]; then
    echo "Failed to login"
    exit 1
fi

# Get total count by querying all entities
echo "Checking database size..."

# Since EntityDB doesn't have a direct count endpoint, we'll use the query API
# with a unique tag that all entities should have (like type:*)
# Or we can check the file size

# Method 1: Check file sizes
echo -e "\n=== File Sizes ==="
ls -lh /opt/entitydb/var/entities.ebf
ls -lh /opt/entitydb/var/entitydb.wal

# Method 2: Query different types
echo -e "\n=== Entity Counts by Type ==="
TYPES=("user" "entity" "test" "project" "task" "server" "app" "db")

total=0
for type in "${TYPES[@]}"; do
    count=$(curl -s -X GET "http://localhost:8085/api/v1/entities/query?filter=type:$type" \
        -H "Authorization: Bearer $SESSION" | jq '.entities | length')
    if [ "$count" -gt 0 ]; then
        echo "$type: $count entities"
        total=$((total + count))
    fi
done

# Method 3: Get dashboard stats
echo -e "\n=== Dashboard Stats ==="
curl -s -X GET http://localhost:8085/api/v1/dashboard/stats \
    -H "Authorization: Bearer $SESSION" | jq

echo -e "\nNote: Can't get exact total count without iterating through all entities"
echo "Estimated total from types found: $total+"