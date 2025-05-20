#!/bin/bash
# Count total entities in the database

# Login as admin
SESSION=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

if [ -z "$SESSION" ]; then
    echo "Failed to login"
    exit 1
fi

echo "=== EntityDB Entity Count ==="

# Check file size first
echo -e "\nDatabase file size:"
ls -lh /opt/entitydb/var/entities.ebf

# Since we can't get a direct count, let's sample by entity types
echo -e "\nSampling entity counts by type:"

# Common entity types we might have
TYPES=("user" "test" "entity" "relationship" "issue" "project" "task" "server" "app" "db" "document" "feature" "bug" "stress")

total_sampled=0
for type in "${TYPES[@]}"; do
    # Query with limit to see how many we get
    result=$(curl -s -X GET "http://localhost:8085/api/v1/entities/query?filter=type:$type" \
        -H "Authorization: Bearer $SESSION")
    
    if [ $? -eq 0 ]; then
        count=$(echo "$result" | jq '.entities | length' 2>/dev/null || echo "0")
        if [ "$count" -gt 0 ]; then
            echo "  type:$type - $count entities"
            total_sampled=$((total_sampled + count))
        fi
    fi
done

echo -e "\nTotal sampled: $total_sampled entities"

# Also try to get some recent entities to see patterns
echo -e "\nRecent entity IDs:"
recent=$(curl -s -X GET "http://localhost:8085/api/v1/entities/query?limit=5&sort=id&order=desc" \
    -H "Authorization: Bearer $SESSION")

echo "$recent" | jq -r '.entities[].id' 2>/dev/null | head -5

# Check if there are bulk test entities
echo -e "\nChecking for bulk test patterns:"
for pattern in "bulk:test" "stress_" "test_" "batch:"; do
    result=$(curl -s -X GET "http://localhost:8085/api/v1/entities/query?filter=$pattern&limit=1" \
        -H "Authorization: Bearer $SESSION")
    
    count=$(echo "$result" | jq '.entities | length' 2>/dev/null || echo "0")
    if [ "$count" -gt 0 ]; then
        echo "  Pattern '$pattern*' found"
    fi
done

echo -e "\nNote: EntityDB doesn't provide a direct count endpoint."
echo "The 732MB file size suggests thousands of entities exist."
echo "Query results appear to be limited to ~1746 per query."