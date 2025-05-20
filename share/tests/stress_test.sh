#!/bin/bash
# Stress test script for EntityDB - creates 100k entities with relationships

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

echo "Creating users..."
# Create 10 users
for i in {1..10}; do
    curl -s -X POST $BASE_URL/users/create \
        -H "Authorization: Bearer $SESSION" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"user$i\",
            \"password\": \"password$i\",
            \"roles\": [\"user\"],
            \"permissions\": [\"entity:view\", \"entity:create\", \"entity:update\"]
        }" > /dev/null
    echo -n "."
done
echo " Users created!"

echo "Creating entities..."
# Create entities in batches
ENTITY_TYPES=("project" "task" "document" "milestone" "comment" "issue" "feature" "bug")
STATUS_TYPES=("active" "pending" "completed" "archived" "draft")
PRIORITIES=("high" "medium" "low" "critical")

created_entities=()

# Function to create entities
create_entities() {
    local start=$1
    local end=$2
    
    for i in $(seq $start $end); do
        # Random entity type and status
        type_idx=$((i % ${#ENTITY_TYPES[@]}))
        status_idx=$((i % ${#STATUS_TYPES[@]}))
        priority_idx=$((i % ${#PRIORITIES[@]}))
        
        tags="[\"type:${ENTITY_TYPES[$type_idx]}\","
        tags+="\"status:${STATUS_TYPES[$status_idx]}\","
        tags+="\"priority:${PRIORITIES[$priority_idx]}\","
        tags+="\"label:entity_$i\","
        tags+="\"category:test\","
        tags+="\"index:$i\"]"
        
        response=$(curl -s -X POST $BASE_URL/entities/create \
            -H "Authorization: Bearer $SESSION" \
            -H "Content-Type: application/json" \
            -d "{\"tags\": $tags}")
        
        # Store some entity IDs for relationships
        if [ $((i % 1000)) -eq 0 ]; then
            entity_id=$(echo $response | jq -r '.entity.id')
            created_entities+=($entity_id)
            echo -n "."
        fi
    done
}

# Create entities in parallel batches
for batch in {0..9}; do
    start=$((batch * 10000 + 1))
    end=$((start + 9999))
    echo "Creating batch $((batch + 1))/10 (entities $start-$end)..."
    create_entities $start $end &
done

# Wait for all batches to complete
wait
echo -e "\nAll entities created!"

echo "Creating relationships..."
# Create relationships between random entities
relationship_types=("depends_on" "blocks" "relates_to" "parent_of" "linked_to")

for i in {1..1000}; do
    from_idx=$((RANDOM % ${#created_entities[@]}))
    to_idx=$((RANDOM % ${#created_entities[@]}))
    rel_idx=$((RANDOM % ${#relationship_types[@]}))
    
    # Skip self-relationships
    if [ $from_idx -eq $to_idx ]; then
        continue
    fi
    
    curl -s -X POST $BASE_URL/entity-relationships \
        -H "Authorization: Bearer $SESSION" \
        -H "Content-Type: application/json" \
        -d "{
            \"from_entity_id\": \"${created_entities[$from_idx]}\",
            \"to_entity_id\": \"${created_entities[$to_idx]}\",
            \"relationship_type\": \"${relationship_types[$rel_idx]}\",
            \"metadata\": {\"created_by\": \"stress_test\", \"index\": $i}
        }" > /dev/null
    
    if [ $((i % 100)) -eq 0 ]; then
        echo -n "."
    fi
done
echo -e "\nRelationships created!"

echo "Testing queries..."
# Test some queries
echo "- Counting entities by type..."
for type in "${ENTITY_TYPES[@]}"; do
    count=$(curl -s -X GET "$BASE_URL/entities/query?filter=type:$type" \
        -H "Authorization: Bearer $SESSION" | jq '.entities | length')
    echo "  $type: $count entities"
done

echo "- Testing temporal query..."
timestamp=$(date -u +%s%N)
history_count=$(curl -s -X GET "$BASE_URL/entities/history?entity_id=${created_entities[0]}" \
    -H "Authorization: Bearer $SESSION" | jq '.history | length')
echo "  Entity ${created_entities[0]} has $history_count historical states"

echo -e "\nStress test complete!"
echo "Created:"
echo "- 10 users"
echo "- 100,000 entities"
echo "- 1,000 relationships"