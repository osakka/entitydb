#!/bin/bash

# secure_entities.sh - Script to convert user entities to use secure password hashes

# Exit on any error
set -e

# Set directory variables
BASE_DIR="/opt/entitydb"
SERVER_DIR="$BASE_DIR/src"
TOOLS_DIR="$BASE_DIR/share/tools"

echo "===== EntityDB Security Implementation Tool ====="
echo "This tool will update user entities to use secure password hashing"

# Check if server is running
echo "Checking if server is running..."
if ! pgrep -f "entitydb_server_entity" > /dev/null; then
    echo "Starting server for entity operations..."
    cd $SERVER_DIR
    go build -o entitydb_server_entity server_db.go
    nohup ./entitydb_server_entity > /dev/null 2>&1 &
    # Wait for server to start
    sleep 2
    echo "Server started"
else
    echo "Server is already running"
fi

# Get admin token
echo "Getting admin token..."
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ADMIN_TOKEN" ]; then
    echo "Failed to get admin token, aborting"
    exit 1
fi
echo "Admin token obtained: ${ADMIN_TOKEN:0:10}..."

# Get list of all user entities
echo "Getting list of user entities..."
USER_ENTITIES=$(curl -s -X GET "http://localhost:8085/api/v1/entities/list?type=user" \
    -H "Authorization: Bearer $ADMIN_TOKEN")

# Extract entity IDs using grep and cut
ENTITY_IDS=$(echo "$USER_ENTITIES" | grep -o '"id":"entity_user_[^"]*"' | cut -d'"' -f4)

# Count entities
ENTITY_COUNT=$(echo "$ENTITY_IDS" | wc -l)
echo "Found $ENTITY_COUNT user entities"

# Process each entity
if [ $ENTITY_COUNT -gt 0 ]; then
    echo "Processing user entities to secure passwords..."
    
    for ENTITY_ID in $ENTITY_IDS; do
        echo "Processing $ENTITY_ID..."
        
        # Get entity details
        ENTITY=$(curl -s -X GET "http://localhost:8085/api/v1/entities/$ENTITY_ID" \
            -H "Authorization: Bearer $ADMIN_TOKEN")
        
        # Extract username and current password
        USERNAME=$(echo "$ENTITY" | grep -o '"username":"[^"]*"' | head -1 | cut -d'"' -f4)
        PASSWORD=$(echo "$ENTITY" | grep -o '"password_hash":"[^"]*"' | cut -d'"' -f4)
        
        if [ -z "$USERNAME" ] || [ -z "$PASSWORD" ]; then
            echo "Could not extract username or password for $ENTITY_ID, skipping"
            continue
        fi
        
        echo "Found user: $USERNAME"
        
        # Check if password is already hashed (bcrypt hashes start with $2a$)
        if [[ "$PASSWORD" == '$2a$'* ]]; then
            echo "Password is already securely hashed, skipping"
            continue
        fi
        
        # Hash the password using our Go password hasher
        HASHED_PASSWORD=$(cd $SERVER_DIR && go run -mod=mod <<EOF
package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "$PASSWORD"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		fmt.Println("ERROR: Failed to hash password")
	} else {
		fmt.Println(string(hash))
	}
}
EOF
)
        
        if [[ "$HASHED_PASSWORD" == "ERROR:"* ]]; then
            echo "Failed to hash password for $USERNAME, skipping"
            continue
        fi
        
        echo "Updating entity with secure password hash..."
        
        # Update the entity
        UPDATE_RESULT=$(curl -s -X PUT "http://localhost:8085/api/v1/entities/$ENTITY_ID" \
            -H "Authorization: Bearer $ADMIN_TOKEN" \
            -H "Content-Type: application/json" \
            -d '{
                "properties": {
                    "username": "'$USERNAME'",
                    "password_hash": "'$HASHED_PASSWORD'"
                }
            }')
        
        if echo "$UPDATE_RESULT" | grep -q '"status":"ok"'; then
            echo "Successfully updated password for $USERNAME"
        else
            echo "Failed to update password for $USERNAME"
            echo "$UPDATE_RESULT"
        fi
    done
    
    echo "Security update completed"
else
    echo "No user entities found, nothing to process"
fi

echo "===== Security Implementation Complete ====="