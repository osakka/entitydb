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

# Create entities in bulk
print("Creating 1000 entities...")
entity_types = ["project", "task", "issue", "document", "feature", "bug"]
statuses = ["active", "pending", "completed", "archived"]
created_ids = []

start_time = time.time()

for i in range(1000):
    entity_type = entity_types[i % len(entity_types)]
    status = statuses[i % len(statuses)]
    
    resp = requests.post(f"{BASE_URL}/entities/create",
        headers=headers,
        json={
            "tags": [
                f"type:{entity_type}",
                f"name:entity_{i}",
                f"status:{status}",
                f"index:{i}",
                "bulk:test"
            ]
        })
    
    if i % 100 == 0:
        created_ids.append(resp.json()["entity"]["id"])
        elapsed = time.time() - start_time
        rate = (i+1) / elapsed if elapsed > 0 else 0
        print(f"Created {i+1} entities... ({rate:.1f} entities/sec)")

# Create relationships
print("\nCreating 100 relationships...")
rel_types = ["depends_on", "blocks", "related_to", "parent_of"]
rel_count = 0

for i in range(len(created_ids)):
    for j in range(i+1, min(i+4, len(created_ids))):
        rel_type = rel_types[rel_count % len(rel_types)]
        resp = requests.post(f"{BASE_URL}/entity-relationships",
            headers=headers,
            json={
                "from_entity_id": created_ids[i],
                "to_entity_id": created_ids[j],
                "relationship_type": rel_type
            })
        rel_count += 1
        if rel_count >= 100:
            break
    if rel_count >= 100:
        break

print(f"Created {rel_count} relationships")

# Test queries
print("\nTesting queries...")
query_resp = requests.get(f"{BASE_URL}/entities/query?filter=bulk:test&limit=10",
    headers=headers)
count = len(query_resp.json()["entities"])
print(f"Query returned {count} entities")

# Get stats
stats_resp = requests.get(f"{BASE_URL}/dashboard/stats", headers=headers)
if stats_resp.status_code == 200:
    stats = stats_resp.json()
    print(f"\nDatabase stats:")
    print(f"- Total entities: {stats.get('total_entities', 'N/A')}")
    print(f"- Total relationships: {stats.get('total_relationships', 'N/A')}")

total_time = time.time() - start_time
print(f"\nTest completed in {total_time:.1f} seconds")
print(f"Average rate: {1000/total_time:.1f} entities/sec")