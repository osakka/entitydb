#!/bin/bash

BASE_URL="https://localhost:8085/api/v1"

echo "Testing multiple passwords for existing admin user..."

# Try login with 'admin'
echo -e "\nTesting login with password='admin'..."
LOGIN1=$(curl -sk -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin"
  }')

echo "Login response:"
echo "$LOGIN1" | jq '.' 2>/dev/null || echo "$LOGIN1"

# Try login with 'password123'
echo -e "\nTesting login with password='password123'..."
LOGIN2=$(curl -sk -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123"
  }')

echo "Login response:"
echo "$LOGIN2" | jq '.' 2>/dev/null || echo "$LOGIN2"

# Try login with 'adminadmin'
echo -e "\nTesting login with password='adminadmin'..."
LOGIN3=$(curl -sk -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "adminadmin"
  }')

echo "Login response:"
echo "$LOGIN3" | jq '.' 2>/dev/null || echo "$LOGIN3"

# Try login with admin2 again
echo -e "\nTesting login with admin2/admin..."
LOGIN4=$(curl -sk -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin2",
    "password": "admin"
  }')

echo "Login response:"
echo "$LOGIN4" | jq '.' 2>/dev/null || echo "$LOGIN4"

# Try with different content type header
echo -e "\nTesting login with different content type header..."
LOGIN5=$(curl -sk -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d 'username=admin&password=admin')

echo "Login response:"
echo "$LOGIN5" | jq '.' 2>/dev/null || echo "$LOGIN5"