#\!/bin/bash

HOST="https://localhost:8085"

echo "1. Testing login..."
RESPONSE=$(curl -k -s -X POST "$HOST/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}' \
    -w "\nHTTP_STATUS:%{http_code}")

BODY=$(echo "$RESPONSE"  < /dev/null |  sed -n '1,/HTTP_STATUS:/p' | sed '$d')
STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)

echo "Login Status: $STATUS"
echo "Login Response: $BODY"

TOKEN=$(echo "$BODY" | jq -r '.token // empty')
echo "Token: $TOKEN"

if [ -z "$TOKEN" ]; then
    echo "ERROR: No token received"
    exit 1
fi

echo -e "\n2. Testing whoami with token..."
WHOAMI_RESPONSE=$(curl -k -s -X GET "$HOST/api/v1/auth/whoami" \
    -H "Authorization: Bearer $TOKEN" \
    -w "\nHTTP_STATUS:%{http_code}")

WHOAMI_BODY=$(echo "$WHOAMI_RESPONSE" | sed -n '1,/HTTP_STATUS:/p' | sed '$d')
WHOAMI_STATUS=$(echo "$WHOAMI_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)

echo "Whoami Status: $WHOAMI_STATUS"
echo "Whoami Response: $WHOAMI_BODY"

echo -e "\n3. Testing entity list with token..."
LIST_RESPONSE=$(curl -k -s -X GET "$HOST/api/v1/entities/list" \
    -H "Authorization: Bearer $TOKEN" \
    -w "\nHTTP_STATUS:%{http_code}")

LIST_STATUS=$(echo "$LIST_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
echo "List Status: $LIST_STATUS"

