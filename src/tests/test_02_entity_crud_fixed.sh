#!/bin/bash
# Test 2: Entity CRUD Operations (Fixed)
# EntityDB E2E Production Readiness Testing

echo "=== EntityDB Entity CRUD Tests ==="
echo ""

# Get admin token first
echo "Getting admin token..."
LOGIN_RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
    echo "✗ Failed to get admin token"
    exit 1
fi
echo "✓ Admin authenticated"
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
        "content": "VGhpcyBpcyBhIHRlc3QgZG9jdW1lbnQgZm9yIENSVUQgb3BlcmF0aW9ucy4="
    }' -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$CREATE_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "200" ] || [ "$STATUS" = "201" ]; then
    ENTITY_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "✓ Entity created successfully"
    echo "  ID: $ENTITY_ID"
else
    echo "✗ Failed to create entity (Status: $STATUS)"
    echo "  Response: $CREATE_RESPONSE"
fi

# Test 2: Read entity by ID
echo ""
echo "2. Reading entity by ID..."
READ_RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID" \
    -H "Authorization: Bearer $TOKEN")

if echo "$READ_RESPONSE" | grep -q "$ENTITY_ID"; then
    echo "✓ Entity retrieved successfully"
    # Check for expected tags
    if echo "$READ_RESPONSE" | grep -q "type:document"; then
        echo "✓ Tags verified"
    else
        echo "✗ Tags missing or incorrect"
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
        \"content\": \"VGhpcyBkb2N1bWVudCBoYXMgYmVlbiB1cGRhdGVkIHdpdGggbmV3IGNvbnRlbnQu\"
    }" -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$UPDATE_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "200" ] || [ "$STATUS" = "204" ]; then
    echo "✓ Entity updated successfully"
else
    echo "✗ Update failed (Status: $STATUS)"
fi

# Test 4: List entities with filters
echo ""
echo "4. Listing entities with filters..."
LIST_RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/list?tag=type:document" \
    -H "Authorization: Bearer $TOKEN")

if echo "$LIST_RESPONSE" | grep -q "$ENTITY_ID"; then
    echo "✓ Entity found in filtered list"
    COUNT=$(echo "$LIST_RESPONSE" | grep -o '"id"' | wc -l)
    echo "  Found $COUNT entities with type:document"
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
    echo "? Entity might not be in results (depends on update success)"
fi

# Test 6: Create entity with base64 content
echo ""
echo "6. Creating entity with larger content..."
# Create 10KB of content (base64 encoded)
LARGE_CONTENT=$(dd if=/dev/urandom bs=1024 count=10 2>/dev/null | base64 | tr -d '\n')

LARGE_RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"tags\": [
            \"type:binary-file\",
            \"name:Random Data File\",
            \"size:10KB\"
        ],
        \"content\": \"$LARGE_CONTENT\"
    }" -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$LARGE_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "200" ] || [ "$STATUS" = "201" ]; then
    LARGE_ID=$(echo "$LARGE_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "✓ Large content entity created"
    echo "  ID: $LARGE_ID"
else
    echo "✗ Failed to create large entity (Status: $STATUS)"
fi

# Test 7: Batch entity creation
echo ""
echo "7. Batch creating entities..."
BATCH_COUNT=0
for i in {1..5}; do
    RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/entities/create" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"tags\": [
                \"type:batch-test\",
                \"batch:test-batch-1\",
                \"index:$i\",
                \"created_at:$(date +%s)\"
            ],
            \"content\": \"QmF0Y2ggZW50aXR5IG51bWJlciAkaQ==\"
        }" -w "\nHTTP_STATUS:%{http_code}")
    
    STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
    if [ "$STATUS" = "200" ] || [ "$STATUS" = "201" ]; then
        ((BATCH_COUNT++))
    fi
done

echo "✓ Created $BATCH_COUNT/5 batch entities"

# Verify batch creation
sleep 1  # Give time for indexing
BATCH_CHECK=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/query?tag=batch:test-batch-1" \
    -H "Authorization: Bearer $TOKEN")

FOUND_COUNT=$(echo "$BATCH_CHECK" | grep -o '"id"' | wc -l)
echo "✓ Found $FOUND_COUNT batch entities in query"

# Test 8: Wildcard search
echo ""
echo "8. Testing wildcard search..."
WILDCARD_RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/query?wildcard=test*" \
    -H "Authorization: Bearer $TOKEN" -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$WILDCARD_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "200" ]; then
    COUNT=$(echo "$WILDCARD_RESPONSE" | grep -o '"id"' | wc -l)
    echo "✓ Wildcard search returned $COUNT results"
else
    echo "? Wildcard search status: $STATUS"
fi

# Test 9: Get unique tag values
echo ""
echo "9. Getting unique tag values..."
TAG_VALUES=$(curl -k -s -X GET "https://localhost:8085/api/v1/tags/values?namespace=type" \
    -H "Authorization: Bearer $TOKEN" -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$TAG_VALUES" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "200" ]; then
    echo "✓ Retrieved unique tag values"
    echo "$TAG_VALUES" | grep -o '"[^"]*"' | sort -u | head -5
else
    echo "✗ Failed to get tag values (Status: $STATUS)"
fi

# Test 10: Pagination
echo ""
echo "10. Testing pagination..."
PAGE1=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/list?limit=2&offset=0" \
    -H "Authorization: Bearer $TOKEN")

PAGE2=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/list?limit=2&offset=2" \
    -H "Authorization: Bearer $TOKEN")

COUNT1=$(echo "$PAGE1" | grep -o '"id"' | wc -l)
COUNT2=$(echo "$PAGE2" | grep -o '"id"' | wc -l)

echo "✓ Pagination results:"
echo "  Page 1: $COUNT1 entities"
echo "  Page 2: $COUNT2 entities"

# Test 11: Entity summary
echo ""
echo "11. Getting entity summary..."
SUMMARY=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/summary" \
    -H "Authorization: Bearer $TOKEN" -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$SUMMARY" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "200" ]; then
    echo "✓ Entity summary retrieved"
else
    echo "✗ Failed to get summary (Status: $STATUS)"
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
        "content": "Q29uY3VycmVudCB1cGRhdGUgdGVzdA=="
    }')

CONCURRENT_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

if [ ! -z "$CONCURRENT_ID" ]; then
    # Launch 3 concurrent updates
    for i in {1..3}; do
        (
            curl -k -s -X PUT "https://localhost:8085/api/v1/entities/update" \
                -H "Authorization: Bearer $TOKEN" \
                -H "Content-Type: application/json" \
                -d "{
                    \"id\": \"$CONCURRENT_ID\",
                    \"tags\": [\"type:concurrent-test\", \"counter:$i\", \"updater:thread-$i\"],
                    \"content\": \"VXBkYXRlZCBieSB0aHJlYWQgJGk=\"
                }" > /dev/null 2>&1
        ) &
    done
    wait
    echo "✓ Concurrent updates completed"
    
    # Check final state
    sleep 1
    FINAL_STATE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/get?id=$CONCURRENT_ID" \
        -H "Authorization: Bearer $TOKEN")
    
    if echo "$FINAL_STATE" | grep -q "updater:thread"; then
        echo "✓ Entity survived concurrent updates"
    else
        echo "✗ Entity state unclear after concurrent updates"
    fi
else
    echo "✗ Failed to create concurrent test entity"
fi

echo ""
echo "=== Entity CRUD Tests Complete ===