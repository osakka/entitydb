#!/bin/bash

# This script attempts to debug the login 500 error

# First, check that the server is running
echo "Checking server status..."
curl -sk https://localhost:8085/api/v1/status

# List all user entities that have admin in the username
echo -e "\n\nListing all user entities with 'admin' username tag..."
curl -sk "https://localhost:8085/api/v1/test/entities/list?tag=id:username:admin"

# Try logging in with verbose curl output
echo -e "\n\nAttempting login with verbose output..."
curl -vsk -X POST "https://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin"
  }'

# Let's try a different password in case it's hardcoded differently
echo -e "\n\nAttempting login with alternate password..."
curl -sk -X POST "https://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin", 
    "password": "password123"
  }'

# Try the osakka user
echo -e "\n\nAttempting login with osakka user..."
curl -sk -X POST "https://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "osakka",
    "password": "mypassword"
  }'