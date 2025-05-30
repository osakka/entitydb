#!/bin/bash

# Test string interning effectiveness

set -e

BASE_URL="https://localhost:8085"
ADMIN_USER="admin"
ADMIN_PASS="admin"

echo "Testing String Interning Effectiveness"
echo "====================================="

# Login
TOKEN=$(curl -k -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}" | \
    jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "Failed to login"
    exit 1
fi

# Get initial state
echo -e "\nInitial state:"
INITIAL=$(curl -k -s -X GET "$BASE_URL/api/v1/system/metrics" \
    -H "Authorization: Bearer $TOKEN")
    
INITIAL_ENTITIES=$(echo "$INITIAL" | jq -r '.database.total_entities')
INITIAL_TAGS=$(echo "$INITIAL" | jq -r '.database.tags_unique')
INITIAL_MEM=$(echo "$INITIAL" | jq -r '.memory.alloc_bytes')

echo "Entities: $INITIAL_ENTITIES"
echo "Unique tags: $INITIAL_TAGS"  
echo "Memory: $(numfmt --to=iec $INITIAL_MEM)"

# Create 10 entities with the EXACT same tags
echo -e "\nCreating 10 entities with identical tags..."

for i in $(seq 1 10); do
    curl -k -s -X POST "$BASE_URL/api/v1/entities/create" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "tags": ["test:interning", "type:test", "status:active"],
            "content": "Testing string interning"
        }' > /dev/null
done

# Get final state
echo -e "\nFinal state:"
FINAL=$(curl -k -s -X GET "$BASE_URL/api/v1/system/metrics" \
    -H "Authorization: Bearer $TOKEN")
    
FINAL_ENTITIES=$(echo "$FINAL" | jq -r '.database.total_entities')
FINAL_TAGS=$(echo "$FINAL" | jq -r '.database.tags_unique')
FINAL_MEM=$(echo "$FINAL" | jq -r '.memory.alloc_bytes')

echo "Entities: $FINAL_ENTITIES"
echo "Unique tags: $FINAL_TAGS"
echo "Memory: $(numfmt --to=iec $FINAL_MEM)"

# Calculate differences
ENTITY_DIFF=$((FINAL_ENTITIES - INITIAL_ENTITIES))
TAG_DIFF=$((FINAL_TAGS - INITIAL_TAGS))
MEM_DIFF=$((FINAL_MEM - INITIAL_MEM))

echo -e "\nDifferences:"
echo "============"
echo "New entities: $ENTITY_DIFF"
echo "New unique tags: $TAG_DIFF (should be minimal if interning works)"
echo "Memory growth: $(numfmt --to=iec $MEM_DIFF)"
echo "Memory per entity: $(numfmt --to=iec $((MEM_DIFF / ENTITY_DIFF)))"