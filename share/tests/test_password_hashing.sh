#!/bin/bash

echo "Testing Password Hashing Implementation"
echo "======================================"

# First get a token as admin (if possible)
echo "1. Trying to get entities list without auth..."
curl -s http://localhost:8085/api/v1/entities/list | python3 -m json.tool

echo -e "\n2. Checking if admin entity exists..."
curl -s "http://localhost:8085/api/v1/entities/list?tag=type:user" | python3 -m json.tool

echo -e "\n3. Checking specific admin tag format..."
curl -s "http://localhost:8085/api/v1/entities/list?tag=id:username:admin" | python3 -m json.tool