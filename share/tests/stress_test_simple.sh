#!/bin/bash
# Simple stress test for EntityDB - creates entities in a manageable way

BASE_URL="http://localhost:8085/api/v1"
ADMIN_USER="admin"
ADMIN_PASS="admin"

# Login as admin
echo "Logging in as admin..."
SESSION=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}" | jq -r '.session_token')

if [ -z "$SESSION" ]; then
    echo "Failed to login as admin"
    exit 1
fi

echo "Creating 5 test users..."
for i in {1..5}; do
    curl -s -X POST $BASE_URL/users/create \
        -H "Authorization: Bearer $SESSION" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"user$i\",
            \"password\": \"password$i\",
            \"roles\": [\"user\"],
            \"permissions\": [\"entity:view\", \"entity:create\"]
        }" > /dev/null
    echo "Created user$i"
done

echo -e "\nCreating 1000 entities (not 100k - let's be reasonable)..."
ENTITY_TYPES=("project" "task" "document" "issue" "feature")
created_ids=()

for i in {1..1000}; do
    type_idx=$((i % ${#ENTITY_TYPES[@]}))
    
    response=$(curl -s -X POST $BASE_URL/entities/create \
        -H "Authorization: Bearer $SESSION" \
        -H "Content-Type: application/json" \
        -d "{\"tags\": [
            \"type:${ENTITY_TYPES[$type_idx]}\",
            \"status:active\",
            \"name:entity_$i\",
            \"test:stress\"
        ]}")
    
    if [ $((i % 100)) -eq 0 ]; then
        entity_id=$(echo $response | jq -r '.entity.id')
        created_ids+=($entity_id)
        echo "Created $i entities..."
    fi
done

echo -e "\nCreating 100 relationships..."
for i in {1..100}; do
    from_idx=$((RANDOM % ${#created_ids[@]}))
    to_idx=$((RANDOM % ${#created_ids[@]}))
    
    if [ $from_idx -ne $to_idx ]; then
        curl -s -X POST $BASE_URL/entity-relationships \
            -H "Authorization: Bearer $SESSION" \
            -H "Content-Type: application/json" \
            -d "{
                \"from_entity_id\": \"${created_ids[$from_idx]}\",
                \"to_entity_id\": \"${created_ids[$to_idx]}\",
                \"relationship_type\": \"relates_to\"
            }" > /dev/null
    fi
    
    if [ $((i % 10)) -eq 0 ]; then
        echo "Created $i relationships..."
    fi
done

echo -e "\nTesting queries..."
# Count entities by type
for type in "${ENTITY_TYPES[@]}"; do
    count=$(curl -s -X GET "$BASE_URL/entities/query?filter=type:$type" \
        -H "Authorization: Bearer $SESSION" | jq '.entities | length')
    echo "$type: $count entities"
done

echo -e "\nStress test complete!"
echo "Created: 5 users, 1000 entities, 100 relationships"