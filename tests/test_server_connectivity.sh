#!/bin/bash
# Test EntityDB server connectivity

echo "Testing EntityDB Server Connectivity"
echo "===================================="

# Check if server is running
echo -e "\n1. Checking if EntityDB process is running..."
if pgrep -f "entitydb" > /dev/null; then
    echo "✓ EntityDB process found"
    pgrep -af "entitydb" | grep -v grep
else
    echo "✗ EntityDB process not found"
    echo "  Start it with: cd /opt/entitydb && ./bin/entitydbd.sh start"
fi

# Check port 8085
echo -e "\n2. Checking if port 8085 is listening..."
if netstat -tuln 2>/dev/null | grep -q ":8085"; then
    echo "✓ Port 8085 is listening"
    netstat -tuln 2>/dev/null | grep ":8085"
else
    echo "✗ Port 8085 is not listening"
fi

# Test health endpoint
echo -e "\n3. Testing health endpoint..."
HEALTH_RESPONSE=$(curl -sk -w "\nHTTP_STATUS:%{http_code}" https://localhost:8085/health 2>/dev/null)
HTTP_STATUS=$(echo "$HEALTH_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)

if [ "$HTTP_STATUS" = "200" ]; then
    echo "✓ Health endpoint returned 200 OK"
    echo "$HEALTH_RESPONSE" | grep -v "HTTP_STATUS:" | jq . 2>/dev/null || echo "$HEALTH_RESPONSE" | grep -v "HTTP_STATUS:"
else
    echo "✗ Health endpoint returned: $HTTP_STATUS"
    echo "$HEALTH_RESPONSE" | grep -v "HTTP_STATUS:"
fi

# Test API endpoint (will likely need auth)
echo -e "\n4. Testing API endpoint (no auth)..."
API_RESPONSE=$(curl -sk -w "\nHTTP_STATUS:%{http_code}" https://localhost:8085/api/v1/entities/list 2>/dev/null)
HTTP_STATUS=$(echo "$API_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)

if [ "$HTTP_STATUS" = "401" ]; then
    echo "✓ API endpoint returned 401 (auth required - this is expected)"
elif [ "$HTTP_STATUS" = "200" ]; then
    echo "✓ API endpoint returned 200 OK (no auth required)"
else
    echo "✗ API endpoint returned: $HTTP_STATUS"
    echo "$API_RESPONSE" | grep -v "HTTP_STATUS:"
fi

# Check SSL certificate
echo -e "\n5. Checking SSL certificate..."
CERT_INFO=$(echo | openssl s_client -connect localhost:8085 2>/dev/null | openssl x509 -noout -subject -dates 2>/dev/null)
if [ -n "$CERT_INFO" ]; then
    echo "✓ SSL certificate found:"
    echo "$CERT_INFO"
else
    echo "✗ Could not retrieve SSL certificate"
fi

# Check browser URL
echo -e "\n6. Browser access information:"
echo "  Local URL: https://localhost:8085/"
echo "  Remote URL: https://claude-code.uk.home.arpa:8085/"
echo ""
echo "Note: If you see certificate warnings in the browser, you need to:"
echo "  1. Navigate to https://localhost:8085/ in your browser"
echo "  2. Click 'Advanced' or 'Show Details'"
echo "  3. Click 'Proceed to localhost (unsafe)' or similar"
echo "  4. This accepts the self-signed certificate"

echo -e "\n===================================="
echo "Connectivity test complete"