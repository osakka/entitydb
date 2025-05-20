#!/bin/bash

# Inspect database data directly

BASE_URL="https://localhost:8085"
echo "Inspecting database data..."

# Use the test endpoint to list entities
echo -e "\n=== Listing all entities via test endpoint ==="
ENTITIES=$(curl -sk "$BASE_URL/api/v1/test/entities/list")

echo "Total entities found:"
echo "$ENTITIES" | jq '. | length'

echo -e "\n=== User entities ==="
echo "$ENTITIES" | jq '.[] | select(.tags[] | contains("type:user")) | {id, tags}'

echo -e "\n=== Admin user details ==="
ADMIN=$(echo "$ENTITIES" | jq '.[] | select(.tags[] | contains("id:username:admin"))')
echo "$ADMIN" | jq .

if [ ! -z "$ADMIN" ]; then
    echo -e "\n=== Admin content (base64) ==="
    echo "$ADMIN" | jq -r .content
    
    echo -e "\n=== Admin content (decoded) ==="
    CONTENT=$(echo "$ADMIN" | jq -r .content)
    if [ ! -z "$CONTENT" ] && [ "$CONTENT" != "null" ]; then
        echo "$CONTENT" | base64 -d
        echo
    else
        echo "No content found"
    fi
fi

echo -e "\n✅ Inspection complete"