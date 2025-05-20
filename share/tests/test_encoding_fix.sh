#!/bin/bash

echo "=== Testing EntityDB Content Encoding Fix ==="
echo "Building and restarting server with fixes..."

# Build server
cd /opt/entitydb/src && make

# Restart to apply changes
cd /opt/entitydb && ENTITYDB_LOG_LEVEL=debug ./bin/entitydbd.sh restart

echo "Waiting for server to start..."
sleep 2

# Test entity creation and retrieval
echo "Testing entity creation and retrieval..."

# Get auth token
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

echo "Got auth token: ${TOKEN:0:20}..."

# Step 1: Delete entity database to start clean
echo -e "\n===== Step 1: Reinitializing database ====="
echo "Stopping server..."
./bin/entitydbd.sh stop
echo "Removing entity database files..."
rm -f var/entity_*
echo "Restarting server..."
ENTITYDB_LOG_LEVEL=debug ./bin/entitydbd.sh start
sleep 2

# Get fresh token
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Step 2: Create string content entity
echo -e "\n===== Step 2: Creating string content entity ====="
STR_ENTITY_ID=$(curl -s -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:test", "encoding_fix:v2.13.0", "content_type:string"],
    "content": "This is a plain text content string that should be stored directly."
  }' | jq -r '.id')

echo "Created string entity: $STR_ENTITY_ID"

# Step 3: Create JSON content entity
echo -e "\n===== Step 3: Creating JSON content entity ====="
JSON_ENTITY_ID=$(curl -s -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:test", "encoding_fix:v2.13.0", "content_type:json"],
    "content": {"title": "Test JSON", "count": 42, "active": true}
  }' | jq -r '.id')

echo "Created JSON entity: $JSON_ENTITY_ID"

# Step 4: Test string content retrieval
echo -e "\n===== Step 4: Testing string content retrieval ====="
STR_ENTITY=$(curl -s -X GET "http://localhost:8085/api/v1/entities/get?id=$STR_ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Retrieved string entity (summary):" 
echo "$STR_ENTITY" | jq '{id: .id, tags: .tags, content_length: (.content | length)}'

# Extract and decode content
STR_CONTENT_B64=$(echo "$STR_ENTITY" | jq -r '.content')
echo -e "\nExtracted base64 content (first 30 chars): ${STR_CONTENT_B64:0:30}..."

# Decode the content
if [ -n "$STR_CONTENT_B64" ]; then
    STR_DECODED=$(echo "$STR_CONTENT_B64" | base64 -d 2>/dev/null)
    echo -e "\nDecoded string content: $STR_DECODED"
    
    # Check if this is wrapped in JSON
    if echo "$STR_DECODED" | jq -e . >/dev/null 2>&1; then
        # Try to detect if it's just valid JSON but not wrapped
        if [[ "$STR_DECODED" == '{'*'"application/octet-stream"'*'}' ]]; then
            echo "WARNING: Content appears to be JSON-wrapped!"
            echo "JSON content: $(echo "$STR_DECODED" | jq .)"
            echo "FAILED: String content is wrapped in JSON"
        else
            echo "NOTE: Content is valid JSON but appears to be intended JSON"
            echo "SUCCESS: This is expected for JSON content"
        fi
    else
        echo "SUCCESS: Content is properly stored as plain text!"
    fi
else
    echo "ERROR: No content returned!"
fi

# Step 5: Test JSON content retrieval
echo -e "\n===== Step 5: Testing JSON content retrieval ====="
JSON_ENTITY=$(curl -s -X GET "http://localhost:8085/api/v1/entities/get?id=$JSON_ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN")

echo "Retrieved JSON entity (summary):" 
echo "$JSON_ENTITY" | jq '{id: .id, tags: .tags, content_length: (.content | length)}'

# Extract and decode content
JSON_CONTENT_B64=$(echo "$JSON_ENTITY" | jq -r '.content')
echo -e "\nExtracted base64 content (first 30 chars): ${JSON_CONTENT_B64:0:30}..."

# Decode the content
if [ -n "$JSON_CONTENT_B64" ]; then
    JSON_DECODED=$(echo "$JSON_CONTENT_B64" | base64 -d 2>/dev/null)
    echo -e "\nDecoded JSON content: $JSON_DECODED"
    
    # Check if this is valid JSON
    if echo "$JSON_DECODED" | jq -e . >/dev/null 2>&1; then
        echo "SUCCESS: JSON content is valid JSON"
        echo "JSON content: $(echo "$JSON_DECODED" | jq .)"
        
        # Check if it's wrapped in application/octet-stream
        if [[ "$JSON_DECODED" == '{'*'"application/octet-stream"'*'}' ]]; then
            echo "WARNING: JSON content is wrapped in application/octet-stream"
            echo "FAILED: JSON content should not be wrapped"
        else
            echo "SUCCESS: JSON content is not wrapped in application/octet-stream"
        fi
    else
        echo "WARNING: JSON content is not valid JSON"
        echo "FAILED: JSON content should be valid JSON"
    fi
else
    echo "ERROR: No content returned!"
fi

echo -e "\nTest completed!"

echo -e "\n===== Entity Database Status ====="
ls -la var/entity_*