#!/bin/bash

# Simple Ticketing System using EntityDB

BASE_URL="http://localhost:8085/api/v1"

# Login
echo "=== EntityDB Ticketing System Demo ==="
echo "1. Logging in..."
TOKEN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}' | jq -r '.token')

# Create a project
echo "2. Creating project..."
curl -s -X POST "$BASE_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:project",
      "id:code:HELPDESK",
      "name:Help Desk",
      "status:active"
    ],
    "content": [
      {"type": "description", "value": "IT Help Desk ticketing"}
    ]
  }' | jq -c '.tags[]' | grep -E '(type:|id:)' 

# Create tickets
echo "3. Creating tickets..."
echo "   - Creating critical ticket..."
curl -s -X POST "$BASE_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:ticket",
      "id:ticket:HD-001",
      "project:HELPDESK",
      "status:open",
      "priority:critical",
      "category:network",
      "created_by:user123"
    ],
    "content": [
      {"type": "title", "value": "Network outage in Building A"},
      {"type": "description", "value": "Complete network connectivity loss affecting all users in Building A since 10:30 AM."}
    ]
  }' > /dev/null && echo "   ✓ HD-001: Network outage"

echo "   - Creating high priority ticket..."
curl -s -X POST "$BASE_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:ticket",
      "id:ticket:HD-002",
      "project:HELPDESK",  
      "status:open",
      "priority:high",
      "category:hardware",
      "created_by:user456"
    ],
    "content": [
      {"type": "title", "value": "Printer not working in HR department"},
      {"type": "description", "value": "Main printer in HR showing error code E-102. Unable to print payroll documents."}
    ]
  }' > /dev/null && echo "   ✓ HD-002: Printer issue"

echo "   - Creating medium priority ticket..."
curl -s -X POST "$BASE_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:ticket",
      "id:ticket:HD-003",
      "project:HELPDESK",
      "status:open",
      "priority:medium",
      "category:software",
      "created_by:user789"
    ],
    "content": [
      {"type": "title", "value": "Cannot access email on mobile"},
      {"type": "description", "value": "Email app on company phone stopped syncing after latest update."}
    ]
  }' > /dev/null && echo "   ✓ HD-003: Email access issue"

# Add a comment
echo "4. Adding comment to ticket..."
curl -s -X POST "$BASE_URL/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:comment",
      "ticket:HD-001",
      "author:tech_john"
    ],
    "content": [
      {"type": "text", "value": "Investigating issue. Found that main switch in Building A has failed. ETA for replacement: 2 hours."}
    ]
  }' > /dev/null && echo "   ✓ Comment added to HD-001"

# Query tickets
echo ""
echo "=== Ticket Summary ==="
echo "All tickets:"
TICKETS=$(curl -s -X GET "$BASE_URL/entities/list?tag=type:ticket" \
  -H "Authorization: Bearer $TOKEN")

echo "$TICKETS" | jq -r '.[] | 
  (.tags | map(select(startswith("id:ticket:"))) | .[0] | split(":")[2]) as $id |
  (.tags | map(select(startswith("priority:"))) | .[0] | split(":")[1]) as $priority |
  (.tags | map(select(startswith("status:"))) | .[0] | split(":")[1]) as $status |
  (.content | map(select(.type == "title")) | .[0].value) as $title |
  "\($id): \($title) [\($status)] Priority: \($priority)"'

echo ""
echo "=== Ticketing System Structure ==="
echo "Entities created:"
echo "• Project: type:project (organizes tickets)"
echo "• Tickets: type:ticket (issues to be resolved)"  
echo "• Comments: type:comment (updates on tickets)"
echo ""
echo "Key tags used:"
echo "• status:open/closed/in_progress"
echo "• priority:critical/high/medium/low"
echo "• category:network/hardware/software"
echo "• assigned_to:username"
echo "• created_by:username"
echo ""
echo "Temporal benefits:"
echo "• All changes are automatically timestamped"
echo "• Can query ticket state at any point in time"
echo "• Full audit trail of all modifications"
echo "• Can see ticket history and progression"