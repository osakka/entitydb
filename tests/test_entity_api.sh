#!/bin/bash
# Test EntityDB Entity API

echo "Testing EntityDB Entity API"
echo "============================"

# Test login first
echo -e "\n1. Testing login..."
LOGIN_RESPONSE=$(curl -sk -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}' \
    https://localhost:8085/api/v1/auth/login)

echo "Login response: $LOGIN_RESPONSE"

# Extract token (assuming JSON response)
TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$TOKEN" ]; then
    echo "✓ Login successful, token: ${TOKEN:0:20}..."
    
    echo -e "\n2. Testing entity list..."
    ENTITY_RESPONSE=$(curl -sk \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        "https://localhost:8085/api/v1/entities/list")
    
    echo "Entity list response:"
    echo "$ENTITY_RESPONSE" | jq . 2>/dev/null || echo "$ENTITY_RESPONSE"
    
    # Count entities
    ENTITY_COUNT=$(echo "$ENTITY_RESPONSE" | jq 'length' 2>/dev/null || echo "Could not parse")
    echo -e "\nEntity count: $ENTITY_COUNT"
    
    echo -e "\n3. Testing with tag filter..."
    USER_RESPONSE=$(curl -sk \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        "https://localhost:8085/api/v1/entities/list?tags=type:user")
    
    echo "User entities response:"
    echo "$USER_RESPONSE" | jq . 2>/dev/null || echo "$USER_RESPONSE"
    
    echo -e "\n4. Testing health endpoint for comparison..."
    HEALTH_RESPONSE=$(curl -sk https://localhost:8085/health)
    echo "Health response:"
    echo "$HEALTH_RESPONSE" | jq . 2>/dev/null || echo "$HEALTH_RESPONSE"
    
else
    echo "✗ Login failed"
    echo "Response was: $LOGIN_RESPONSE"
fi

echo -e "\n============================"
echo "API test complete"