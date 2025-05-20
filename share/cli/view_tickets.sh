#!/bin/bash

# View tickets in EntityDB

BASE_URL="http://localhost:8085/api/v1"

# Login
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' | jq -r '.token')

echo "=== Current Tickets in System ==="

# Get all entities and filter for tickets
echo "Fetching all entities..."
ALL_ENTITIES=$(curl -s -X GET "$BASE_URL/entities/list" \
  -H "Authorization: Bearer $TOKEN")

# Count different entity types
echo ""
echo "Entity Types Summary:"
echo "$ALL_ENTITIES" | jq -r '.[].tags[]' | grep "^type:" | sort | uniq -c

echo ""
echo "Tickets in system:"
echo "$ALL_ENTITIES" | jq -r '.[] | 
  select(.tags[] | contains("type:ticket")) | 
  (.tags | map(select(startswith("id:ticket:"))) | .[0] | split(":")[2]) as $id |
  (.tags | map(select(startswith("priority:"))) | .[0] | split(":")[1] // "none") as $priority |
  (.tags | map(select(startswith("status:"))) | .[0] | split(":")[1] // "unknown") as $status |
  (.content | map(select(.type == "title")) | .[0].value // "No title") as $title |
  "• \($id): \($title) [Status: \($status), Priority: \($priority)]"'

echo ""
echo "Projects in system:"
echo "$ALL_ENTITIES" | jq -r '.[] | 
  select(.tags[] | contains("type:project")) | 
  (.tags | map(select(startswith("id:code:"))) | .[0] | split(":")[2]) as $code |
  (.tags | map(select(startswith("name:"))) | .[0] | split(":")[1]) as $name |
  "• \($code): \($name)"'

echo ""
echo "Recent comments:"
echo "$ALL_ENTITIES" | jq -r '.[] | 
  select(.tags[] | contains("type:comment")) | 
  (.tags | map(select(startswith("ticket:"))) | .[0] | split(":")[1]) as $ticket |
  (.tags | map(select(startswith("author:"))) | .[0] | split(":")[1]) as $author |
  (.content | map(select(.type == "text")) | .[0].value) as $text |
  "• On \($ticket) by \($author): \($text)"' | head -5

# Show a specific ticket with history
echo ""
echo "=== Ticket HD-001 Timeline (if exists) ==="
HD001_ID=$(echo "$ALL_ENTITIES" | jq -r '.[] | 
  select((.tags[] | contains("type:ticket")) and (.tags[] | contains("id:ticket:HD-001"))) | .id')

if [ ! -z "$HD001_ID" ]; then
  # Get history of this ticket
  curl -s -X GET "$BASE_URL/entities/history?id=$HD001_ID" \
    -H "Authorization: Bearer $TOKEN" | jq -r '.[] | 
    "\(.timestamp): \(.tags | join(", "))"' | head -5
else
  echo "Ticket HD-001 not found"
fi