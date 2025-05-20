#!/usr/bin/env python3
import requests
import json
import time

BASE_URL = "http://localhost:8085/api/v1"

# Login as admin
print("Logging in...")
login_resp = requests.post(f"{BASE_URL}/auth/login", 
    json={"username": "admin", "password": "admin"})
session_token = login_resp.json()["token"]
headers = {"Authorization": f"Bearer {session_token}"}

# Create entities
print("Creating entities...")
entity_types = ["project", "task", "issue", "document", "feature"]
statuses = ["active", "pending", "completed"]
created_ids = []

start_time = time.time()
success_count = 0
error_count = 0

for i in range(100):  # Start with 100 to be safe
    entity_type = entity_types[i % len(entity_types)]
    status = statuses[i % len(statuses)]
    
    resp = requests.post(f"{BASE_URL}/entities/create",
        headers=headers,
        json={
            "tags": [
                f"type:{entity_type}",
                f"name:bulk_test_{i}",
                f"status:{status}",
                f"index:{i}",
                "bulk:test"
            ]
        })
    
    if resp.status_code == 201:
        success_count += 1
        data = resp.json()
        # Handle different response formats
        if "entity" in data and "id" in data["entity"]:
            entity_id = data["entity"]["id"]
        elif "id" in data:
            entity_id = data["id"]
        else:
            entity_id = None
        
        if entity_id and i % 10 == 0:
            created_ids.append(entity_id)
    else:
        error_count += 1
        print(f"Error creating entity {i}: {resp.status_code}")
    
    if i % 20 == 0:
        elapsed = time.time() - start_time
        rate = success_count / elapsed if elapsed > 0 else 0
        print(f"Progress: {i+1}/100, Success: {success_count}, Errors: {error_count} ({rate:.1f} entities/sec)")

print(f"\nCreated {success_count} entities successfully")
print(f"Stored {len(created_ids)} entity IDs for relationships")

# Create a few relationships
if len(created_ids) >= 2:
    print("\nCreating relationships...")
    rel_count = 0
    for i in range(min(10, len(created_ids)-1)):
        resp = requests.post(f"{BASE_URL}/entity-relationships",
            headers=headers,
            json={
                "from_entity_id": created_ids[i],
                "to_entity_id": created_ids[i+1],
                "relationship_type": "relates_to"
            })
        if resp.status_code == 201:
            rel_count += 1
        else:
            print(f"Error creating relationship: {resp.status_code}")
    print(f"Created {rel_count} relationships")

# Test query
print("\nTesting query...")
query_resp = requests.get(f"{BASE_URL}/entities/query?filter=bulk:test&limit=5",
    headers=headers)
if query_resp.status_code == 200:
    entities = query_resp.json().get("entities", [])
    print(f"Query returned {len(entities)} entities")
else:
    print(f"Query error: {query_resp.status_code}")

total_time = time.time() - start_time
print(f"\nTest completed in {total_time:.1f} seconds")
print(f"Average rate: {success_count/total_time:.1f} entities/sec")