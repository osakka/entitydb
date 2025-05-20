#!/bin/bash

# EntityDB Ticketing System Demo

BASE_URL="http://localhost:8085/api/v1"

# Login and get token
echo "=== Logging in ==="
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' | jq -r '.token')

echo "Token: $TOKEN"
echo

# Create a ticket
echo "=== Creating a new ticket ==="
TICKET_RESPONSE=$(curl -s -X POST "$BASE_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:ticket",
      "id:ticket:DEMO-1",
      "project:DEMO",
      "status:open",
      "priority:high",
      "label:bug",
      "created_by:admin"
    ],
    "content": [
      {
        "type": "title",
        "value": "Demo ticket for testing"
      },
      {
        "type": "description",
        "value": "This is a demo ticket to show the ticketing system capabilities."
      }
    ]
  }')

echo "$TICKET_RESPONSE" | jq
echo

# Query tickets
echo "=== Listing all tickets ==="
curl -s -X GET "$BASE_URL/entities/list?tag=type:ticket" \
  -H "Authorization: Bearer $TOKEN" | jq

echo
echo "=== Ticketing System Structure ==="
echo "1. Project entities (type:project) - Organize tickets by project"
echo "2. Ticket entities (type:ticket) - The actual tickets with status, priority, etc."
echo "3. Comment entities (type:comment) - Comments linked to tickets"
echo "4. Label entities (type:label) - Categories and priorities"
echo
echo "Tags used:"
echo "- type:ticket - Identifies ticket entities"
echo "- status:open/closed/in_progress - Ticket status"
echo "- priority:high/medium/low - Ticket priority"
echo "- project:PROJECT_CODE - Associates ticket with project"
echo "- assigned_to:username - Assignment"
echo "- label:bug/feature/enhancement - Categorization"