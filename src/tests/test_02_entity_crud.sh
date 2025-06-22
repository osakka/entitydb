#!/bin/bash
# Test 2: Entity CRUD Operations
# EntityDB E2E Production Readiness Testing

echo "=== EntityDB Entity CRUD Tests ==="
echo ""

# Get admin token first
echo "Getting admin token..."
LOGIN_RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")
echo "Admin authenticated"
echo ""

# Test 1: Create single entity
echo "1. Creating single entity..."
CREATE_RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "tags": [
            "type:document",
            "name:Test Document",
            "status:draft",
            "priority:high"
        ],
        "content": "This is a test document for CRUD operations."
    }')

if echo "$CREATE_RESPONSE" | grep -q '"id"'; then
    ENTITY_ID=$(echo "$CREATE_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")
    echo "✓ Entity created successfully"
    echo "  ID: $ENTITY_ID"
else
    echo "✗ Failed to create entity: $CREATE_RESPONSE"
    exit 1
fi

# Test 2: Read entity by ID
echo ""
echo "2. Reading entity by ID..."
READ_RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
    -H "Authorization: Bearer $TOKEN")

if echo "$READ_RESPONSE" | grep -q "$ENTITY_ID"; then
    echo "✓ Entity retrieved successfully"
    # Check content
    if echo "$READ_RESPONSE" | grep -q "This is a test document"; then
        echo "✓ Content verified"
    else
        echo "✗ Content mismatch"
    fi
else
    echo "✗ Failed to read entity"
fi

# Test 3: Update entity
echo ""
echo "3. Updating entity..."
UPDATE_RESPONSE=$(curl -k -s -X PUT "https://localhost:8085/api/v1/entities/update" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"id\": \"$ENTITY_ID\",
        \"tags\": [
            \"type:document\",
            \"name:Updated Test Document\",
            \"status:published\",
            \"priority:low\",
            \"updated_at:$(date +%s)\"
        ],
        \"content\": \"This document has been updated with new content.\"
    }")

if echo "$UPDATE_RESPONSE" | grep -q '"success":true'; then
    echo "✓ Entity updated successfully"
else
    echo "✗ Update failed: $UPDATE_RESPONSE"
fi

# Test 4: List entities with filters
echo ""
echo "4. Listing entities with filters..."
LIST_RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/list?tag=type:document" \
    -H "Authorization: Bearer $TOKEN")

if echo "$LIST_RESPONSE" | grep -q "$ENTITY_ID"; then
    echo "✓ Entity found in filtered list"
else
    echo "✗ Entity not found in list"
fi

# Test 5: Query entities
echo ""
echo "5. Querying entities..."
QUERY_RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/query?tag=status:published&limit=10" \
    -H "Authorization: Bearer $TOKEN")

if echo "$QUERY_RESPONSE" | grep -q "$ENTITY_ID"; then
    echo "✓ Entity found in query results"
else
    echo "✗ Entity not found in query"
fi

# Test 6: Create large entity (>1MB)
echo ""
echo "6. Creating large entity (testing compression)..."
# Generate 1.5MB of content
LARGE_CONTENT=$(python3 -c "print('x' * 1572864)")

LARGE_RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"tags\": [
            \"type:large-file\",
            \"name:Large Test File\",
            \"size:1.5MB\"
        ],
        \"content\": \"$LARGE_CONTENT\"
    }" -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$LARGE_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "201" ] || [ "$STATUS" = "200" ]; then
    LARGE_ID=$(echo "$LARGE_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "✓ Large entity created successfully"
    echo "  ID: $LARGE_ID"
else
    echo "✗ Failed to create large entity (Status: $STATUS)"
fi

# Test 7: Batch entity creation
echo ""
echo "7. Batch creating entities..."
for i in {1..5}; do
    curl -k -s -X POST "https://localhost:8085/api/v1/entities/create" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"tags\": [
                \"type:batch-test\",
                \"batch:test-batch-1\",
                \"index:$i\",
                \"created_at:$(date +%s)\"
            ],
            \"content\": \"Batch entity number $i\"
        }" > /dev/null 2>&1 &
done
wait
echo "✓ Batch creation completed"

# Verify batch creation
BATCH_COUNT=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/query?tag=batch:test-batch-1" \
    -H "Authorization: Bearer $TOKEN" | python3 -c "import sys, json; data = json.load(sys.stdin); print(len(data))")

if [ "$BATCH_COUNT" -eq "5" ]; then
    echo "✓ All 5 batch entities created"
else
    echo "✗ Expected 5 batch entities, found $BATCH_COUNT"
fi

# Test 8: Wildcard search
echo ""
echo "8. Testing wildcard search..."
WILDCARD_RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/query?wildcard=test*" \
    -H "Authorization: Bearer $TOKEN")

COUNT=$(echo "$WILDCARD_RESPONSE" | python3 -c "import sys, json; data = json.load(sys.stdin) if sys.stdin.read().strip() else []; print(len(data))" 2>/dev/null || echo "0")
if [ "$COUNT" -gt "0" ]; then
    echo "✓ Wildcard search returned $COUNT results"
else
    echo "✗ Wildcard search returned no results"
fi

# Test 9: Tag operations
echo ""
echo "9. Testing tag operations..."
# Add tag
ADD_TAG_RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/entities/$ENTITY_ID/tags" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"tag": "new-tag:added"}' -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$ADD_TAG_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "200" ] || [ "$STATUS" = "201" ]; then
    echo "✓ Tag added successfully"
else
    echo "✗ Failed to add tag (Status: $STATUS)"
fi

# Test 10: Delete entity
echo ""
echo "10. Deleting entity..."
DELETE_RESPONSE=$(curl -k -s -X DELETE "https://localhost:8085/api/v1/entities/$ENTITY_ID" \
    -H "Authorization: Bearer $TOKEN" -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$DELETE_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "200" ] || [ "$STATUS" = "204" ]; then
    echo "✓ Entity deleted successfully"
    
    # Verify deletion
    VERIFY_RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
        -H "Authorization: Bearer $TOKEN" -w "\nHTTP_STATUS:%{http_code}")
    
    STATUS=$(echo "$VERIFY_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
    if [ "$STATUS" = "404" ]; then
        echo "✓ Entity confirmed deleted (404)"
    else
        echo "✗ Entity still exists after deletion"
    fi
else
    echo "✗ Failed to delete entity"
fi

# Test 11: Pagination
echo ""
echo "11. Testing pagination..."
PAGE1=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/list?limit=2&offset=0" \
    -H "Authorization: Bearer $TOKEN")

PAGE2=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/list?limit=2&offset=2" \
    -H "Authorization: Bearer $TOKEN")

COUNT1=$(echo "$PAGE1" | python3 -c "import sys, json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
COUNT2=$(echo "$PAGE2" | python3 -c "import sys, json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")

if [ "$COUNT1" -le "2" ] && [ "$COUNT2" -le "2" ]; then
    echo "✓ Pagination working (Page1: $COUNT1 items, Page2: $COUNT2 items)"
else
    echo "✗ Pagination not working correctly"
fi

# Test 12: Concurrent updates
echo ""
echo "12. Testing concurrent updates..."
# Create test entity
CREATE_RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:concurrent-test", "counter:0"],
        "content": "Concurrent update test"
    }')

CONCURRENT_ID=$(echo "$CREATE_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])")

# Launch 5 concurrent updates
for i in {1..5}; do
    (
        curl -k -s -X PUT "https://localhost:8085/api/v1/entities/update" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "{
                \"id\": \"$CONCURRENT_ID\",
                \"tags\": [\"type:concurrent-test\", \"counter:$i\", \"updater:thread-$i\"],
                \"content\": \"Updated by thread $i\"
            }" > /dev/null 2>&1
    ) &
done
wait
echo "✓ Concurrent updates completed"

# Check final state
FINAL_STATE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/get?id=$CONCURRENT_ID" \
    -H "Authorization: Bearer $TOKEN")

if echo "$FINAL_STATE" | grep -q "updater:thread"; then
    echo "✓ Entity survived concurrent updates"
else
    echo "✗ Entity corrupted by concurrent updates"
fi

echo ""
echo "=== Entity CRUD Tests Complete ===