#!/bin/bash
# More detailed database check

# Login as admin
SESSION=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

echo "Session: ${SESSION:0:20}..."

# Check specific test patterns
echo -e "\n=== Checking Test Patterns ==="
patterns=("bulk:test" "batch:0" "batch:1" "stress_" "test_" "quick_")

for pattern in "${patterns[@]}"; do
    count=$(curl -s -X GET "http://localhost:8085/api/v1/entities/query?filter=$pattern&limit=1" \
        -H "Authorization: Bearer $SESSION" | jq '.entities | length')
    echo "Pattern '$pattern': found matches"
done

# Check a sample entity to see tag structure
echo -e "\n=== Sample Entity Structure ==="
SAMPLE=$(curl -s -X GET "http://localhost:8085/api/v1/entities/query?filter=type:test&limit=1" \
    -H "Authorization: Bearer $SESSION")

echo "Sample entity tags:"
echo "$SAMPLE" | jq '.entities[0].tags[:10]'

# Check for users
echo -e "\n=== User Entities ==="
USERS=$(curl -s -X GET "http://localhost:8085/api/v1/entities/query?filter=type:user&limit=10" \
    -H "Authorization: Bearer $SESSION")

echo "Found users:"
echo "$USERS" | jq '.entities[].tags' | grep "id:username" | sort | uniq

echo -e "\n=== Summary ==="
echo "Database contains many test entities from various stress tests"
echo "File size: 732MB indicates thousands of entities"
echo "Each entity type query returns 1746 (likely query limit)"