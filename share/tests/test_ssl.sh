#!/bin/bash
# Test SSL support for EntityDB

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="https://localhost:8443"
BASE_URL_HTTP="http://localhost:8085"

echo -e "${YELLOW}Testing EntityDB SSL Support${NC}"
echo "================================"

# Test 1: Check if HTTPS server is running
echo -n "Test 1: Check HTTPS server status... "
if curl -k -f -s ${BASE_URL}/api/v1/status > /dev/null; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    echo "  Error: Could not connect to HTTPS server at ${BASE_URL}"
fi

# Test 2: Check if HTTP redirects to HTTPS
echo -n "Test 2: Check HTTP to HTTPS redirect... "
REDIRECT_URL=$(curl -s -o /dev/null -w %{redirect_url} ${BASE_URL_HTTP}/api/v1/status)
if [[ "${REDIRECT_URL}" == "https://"* ]]; then
    echo -e "${GREEN}PASS${NC}"
    echo "  Redirects to: ${REDIRECT_URL}"
else
    echo -e "${RED}FAIL${NC}"
    echo "  Error: No redirect or incorrect redirect URL: ${REDIRECT_URL}"
fi

# Test 3: Test login via HTTPS
echo -n "Test 3: Test login via HTTPS... "
LOGIN_RESPONSE=$(curl -k -s -X POST ${BASE_URL}/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}')

if echo "${LOGIN_RESPONSE}" | grep -q "token"; then
    echo -e "${GREEN}PASS${NC}"
    TOKEN=$(echo "${LOGIN_RESPONSE}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
else
    echo -e "${RED}FAIL${NC}"
    echo "  Error: Could not login via HTTPS"
    echo "  Response: ${LOGIN_RESPONSE}"
    exit 1
fi

# Test 4: Test authenticated API call via HTTPS
echo -n "Test 4: Test authenticated API call... "
ENTITIES_RESPONSE=$(curl -k -s ${BASE_URL}/api/v1/entities/list \
    -H "Authorization: Bearer ${TOKEN}")

if echo "${ENTITIES_RESPONSE}" | grep -q '\['; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    echo "  Error: Could not fetch entities via HTTPS"
    echo "  Response: ${ENTITIES_RESPONSE}"
fi

# Test 5: Check certificate details
echo -n "Test 5: Check certificate details... "
CERT_INFO=$(echo | openssl s_client -connect localhost:8443 2>/dev/null | openssl x509 -noout -subject -issuer -dates 2>/dev/null)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}PASS${NC}"
    echo "  Certificate info:"
    echo "${CERT_INFO}" | sed 's/^/    /'
else
    echo -e "${YELLOW}SKIP${NC} (certificate details not available)"
fi

echo
echo -e "${GREEN}SSL testing complete!${NC}"