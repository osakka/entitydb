#!/bin/bash

echo "Testing User Creation with Password Hashing"
echo "=========================================="

# Get admin token
echo "1. Getting admin token..."
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin"}' | \
  python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")

echo "Token: $TOKEN"

echo -e "\n2. Creating new user 'testuser'..."
curl -X POST http://localhost:8085/api/v1/users/create \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "email": "test@example.com",
    "full_name": "Test User",
    "role": "user"
  }' | python3 -m json.tool

echo -e "\n3. Checking created user entity..."
curl -s "http://localhost:8085/api/v1/entities/list?tag=id:username:testuser" | python3 -m json.tool | grep -A 5 password_hash

echo -e "\n4. Testing login with new user..."
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","password":"testpass123"}' | python3 -m json.tool

echo -e "\n5. Testing login with wrong password..."
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","password":"wrongpass"}' | python3 -m json.tool