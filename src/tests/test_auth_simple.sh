#!/bin/bash
# Simple Authentication Test

echo "=== EntityDB Authentication Tests ==="
echo ""

# Test 1: Admin login
echo "1. Testing admin login..."
RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}')

if echo "$RESPONSE" | grep -q '"token"'; then
    TOKEN=$(echo "$RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")
    echo "✓ Admin login successful"
    echo "  Token: ${TOKEN:0:20}..."
else
    echo "✗ Admin login failed: $RESPONSE"
    exit 1
fi

# Test 2: Invalid password
echo ""
echo "2. Testing invalid password..."
RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"wrong"}' \
    -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "401" ]; then
    echo "✓ Invalid password rejected (401)"
else
    echo "✗ Expected 401, got $STATUS"
fi

# Test 3: Access with valid token
echo ""
echo "3. Testing protected endpoint with valid token..."
RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/dashboard/stats" \
    -H "Authorization: Bearer $TOKEN" \
    -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "200" ]; then
    echo "✓ Valid token accepted (200)"
else
    echo "✗ Expected 200, got $STATUS"
fi

# Test 4: Access with invalid token
echo ""
echo "4. Testing protected endpoint with invalid token..."
RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/dashboard/stats" \
    -H "Authorization: Bearer invalid-token" \
    -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "401" ]; then
    echo "✓ Invalid token rejected (401)"
else
    echo "✗ Expected 401, got $STATUS"
fi

# Test 5: Create test user
echo ""
echo "5. Creating test user with limited permissions..."
RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/users/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "testuser",
        "password": "testpass123",
        "email": "test@example.com",
        "tags": [
            "rbac:role:user",
            "rbac:perm:entity:view"
        ]
    }')

if echo "$RESPONSE" | grep -q '"id"'; then
    echo "✓ Test user created"
else
    echo "✗ Failed to create test user: $RESPONSE"
fi

# Test 6: Login as test user
echo ""
echo "6. Testing test user login..."
RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser","password":"testpass123"}')

if echo "$RESPONSE" | grep -q '"token"'; then
    TEST_TOKEN=$(echo "$RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")
    echo "✓ Test user login successful"
else
    echo "✗ Test user login failed"
fi

# Test 7: Test user permissions
echo ""
echo "7. Testing test user permissions..."

# Can view entities (has permission)
RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/list" \
    -H "Authorization: Bearer $TEST_TOKEN" \
    -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "200" ]; then
    echo "✓ Test user can view entities (200)"
else
    echo "✗ Expected 200, got $STATUS"
fi

# Cannot access admin endpoints (no permission)
RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/rbac/metrics" \
    -H "Authorization: Bearer $TEST_TOKEN" \
    -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "403" ]; then
    echo "✓ Test user blocked from admin endpoint (403)"
else
    echo "✗ Expected 403, got $STATUS"
fi

# Test 8: Logout
echo ""
echo "8. Testing logout..."
RESPONSE=$(curl -k -s -X POST "https://localhost:8085/api/v1/auth/logout" \
    -H "Authorization: Bearer $TEST_TOKEN" \
    -w "\nHTTP_STATUS:%{http_code}")

# Try to use token after logout
RESPONSE=$(curl -k -s -X GET "https://localhost:8085/api/v1/entities/list" \
    -H "Authorization: Bearer $TEST_TOKEN" \
    -w "\nHTTP_STATUS:%{http_code}")

STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
if [ "$STATUS" = "401" ]; then
    echo "✓ Token invalidated after logout (401)"
else
    echo "✗ Expected 401, got $STATUS"
fi

echo ""
echo "=== Authentication Tests Complete ==="