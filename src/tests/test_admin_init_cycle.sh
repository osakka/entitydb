#!/bin/bash

# Test admin initialization cycle
# This script deletes the DB and tests if admin user is created correctly

set -e

echo "=== Admin Initialization Test Cycle ==="
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if server is running
is_server_running() {
    pgrep -f "entitydb" > /dev/null 2>&1
}

# Function to wait for server to be ready
wait_for_server() {
    echo -n "Waiting for server to be ready..."
    for i in {1..30}; do
        if curl -k -s https://localhost:8085/health > /dev/null 2>&1; then
            echo -e " ${GREEN}Ready!${NC}"
            return 0
        fi
        echo -n "."
        sleep 1
    done
    echo -e " ${RED}Timeout!${NC}"
    return 1
}

# Step 1: Stop server if running
echo "Step 1: Stopping server if running..."
if is_server_running; then
    echo "Server is running, stopping it..."
    cd /opt/entitydb/bin
    ./entitydbd.sh stop
    sleep 2
else
    echo "Server is not running"
fi

# Step 2: Delete database files
echo -e "\nStep 2: Deleting database files..."
cd /opt/entitydb/var
rm -f entities.ebf entitydb.wal entities.idx
echo -e "${GREEN}Database files deleted${NC}"

# Step 3: Start server (should create admin user)
echo -e "\nStep 3: Starting server (should auto-create admin user)..."
cd /opt/entitydb/bin
./entitydbd.sh start

# Wait for server to be ready
if ! wait_for_server; then
    echo -e "${RED}Server failed to start!${NC}"
    exit 1
fi

# Step 4: Test admin login
echo -e "\nStep 4: Testing admin login..."
RESPONSE=$(curl -k -s -X POST https://localhost:8085/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}' \
    -w "\nHTTP_STATUS:%{http_code}" 2>/dev/null)

HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | grep -v "HTTP_STATUS:")

echo "Response status: $HTTP_STATUS"
echo "Response body: $BODY"

if [ "$HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Admin login successful!${NC}"
else
    echo -e "${RED}✗ Admin login failed!${NC}"
fi

# Step 5: Check what entities were created
echo -e "\nStep 5: Checking created entities..."

# Get all users
echo -e "\n${YELLOW}Users:${NC}"
curl -k -s https://localhost:8085/api/v1/entities/listbytag?tag=type:user | jq -r '.[] | "- \(.id) (tags: \(.tags | join(", ")))"' 2>/dev/null || echo "Failed to list users"

# Get all credentials
echo -e "\n${YELLOW}Credentials:${NC}"
curl -k -s https://localhost:8085/api/v1/entities/listbytag?tag=type:credential | jq -r '.[] | "- \(.id)"' 2>/dev/null || echo "Failed to list credentials"

# Get all relationships
echo -e "\n${YELLOW}User-Credential Relationships:${NC}"
curl -k -s https://localhost:8085/api/v1/entities/listbytag?tag=type:relationship | jq -r 'map(select(.tags[] | contains("has_credential"))) | .[] | "- \(.id): \(.tags | map(select(startswith("source:") or startswith("target:"))) | join(" -> "))"' 2>/dev/null || echo "Failed to list relationships"

# Get admin user specifically
echo -e "\n${YELLOW}Admin user details:${NC}"
ADMIN_USERS=$(curl -k -s https://localhost:8085/api/v1/entities/listbytag?tag=type:user | jq -r '.[] | select(.tags[] | contains("username:admin")) | .id' 2>/dev/null)

if [ -n "$ADMIN_USERS" ]; then
    for USER_ID in $ADMIN_USERS; do
        echo -e "\nUser: $USER_ID"
        USER_DETAILS=$(curl -k -s "https://localhost:8085/api/v1/entities/get?id=$USER_ID")
        echo "$USER_DETAILS" | jq -r '.tags[]' 2>/dev/null | sed 's/^/  - /'
        
        # Check if this user has credentials
        HAS_CRED=$(curl -k -s "https://localhost:8085/api/v1/entities/listbytag?tag=source:$USER_ID" | jq -r '.[] | select(.tags[] | contains("has_credential"))' 2>/dev/null)
        if [ -n "$HAS_CRED" ]; then
            echo -e "  ${GREEN}✓ Has credential relationship${NC}"
        else
            echo -e "  ${RED}✗ No credential relationship${NC}"
        fi
    done
else
    echo -e "${RED}No admin users found!${NC}"
fi

# Summary
echo -e "\n=== Summary ==="
if [ "$HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Admin initialization successful${NC}"
    exit 0
else
    echo -e "${RED}✗ Admin initialization failed${NC}"
    echo -e "\nPossible issues:"
    echo "- Admin user not created"
    echo "- Credential not created"
    echo "- User-credential relationship not created"
    echo "- Password hashing issue"
    exit 1
fi