#!/usr/bin/env python3
import requests
import json

BASE_URL = "http://localhost:8085/api/v1"

# Login as admin
print("Logging in...")
login_resp = requests.post(f"{BASE_URL}/auth/login", 
    json={"username": "admin", "password": "admin"})

# Check response
if login_resp.status_code != 200:
    print(f"Login failed: {login_resp.text}")
    exit(1)

resp_data = login_resp.json()
if "error" in resp_data:
    print(f"Login error: {resp_data['error']}")
    exit(1)

# Get session token
if "token" in resp_data:
    session_token = resp_data["token"]
else:
    print(f"Unexpected response: {resp_data}")
    exit(1)

headers = {"Authorization": f"Bearer {session_token}"}
print(f"Logged in with session: {session_token[:10]}...")

# Create one user
print("Creating test user...")
user_resp = requests.post(f"{BASE_URL}/users/create", 
    headers=headers,
    json={
        "username": "testuser1",
        "password": "testpass1",
        "roles": ["user"],
        "permissions": ["entity:view"]
    })
print(f"User creation response: {user_resp.status_code}")

# Create a few entities
print("Creating 10 test entities...")
for i in range(10):
    resp = requests.post(f"{BASE_URL}/entities/create",
        headers=headers,
        json={
            "tags": [
                "type:test",
                f"name:test_{i}",
                "status:active"
            ]
        })
    print(f"Created entity {i+1}: {resp.status_code}")

print("Test complete!")