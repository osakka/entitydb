#!/usr/bin/env python3
import requests
import json
import time

BASE_URL = "http://localhost:8085/api/v1"

# Login as admin
print("Logging in...")
login_resp = requests.post(f"{BASE_URL}/auth/login", 
    json={"username": "admin", "password": "admin"})
session_token = login_resp.json()["session_token"]
headers = {"Authorization": f"Bearer {session_token}"}

# Create a few users
print("Creating 3 test users...")
for i in range(1, 4):
    requests.post(f"{BASE_URL}/users/create", 
        headers=headers,
        json={
            "username": f"test{i}",
            "password": f"test{i}",
            "roles": ["user"],
            "permissions": ["entity:view", "entity:create"]
        })
print("Users created")

# Create entities
print("Creating 100 test entities...")
entity_types = ["project", "task", "issue", "document"]
entity_ids = []

for i in range(100):
    entity_type = entity_types[i % len(entity_types)]
    resp = requests.post(f"{BASE_URL}/entities/create",
        headers=headers,
        json={
            "tags": [
                f"type:{entity_type}",
                f"name:test_{i}",
                "status:active",
                "test:stress"
            ]
        })
    if i % 25 == 0:
        entity_ids.append(resp.json()["entity"]["id"])
        print(f"Created {i+1} entities...")

# Create some relationships
print("\nCreating 10 relationships...")
for i in range(10):
    if len(entity_ids) >= 2:
        requests.post(f"{BASE_URL}/entity-relationships",
            headers=headers,
            json={
                "from_entity_id": entity_ids[i % len(entity_ids)],
                "to_entity_id": entity_ids[(i+1) % len(entity_ids)],
                "relationship_type": "relates_to"
            })

# Test query
print("\nTesting query...")
query_resp = requests.get(f"{BASE_URL}/entities/query?filter=test:stress",
    headers=headers)
count = len(query_resp.json()["entities"])
print(f"Found {count} test entities")

print("\nStress test complete!")