#!/usr/bin/env python3
import requests
import json
import time
import random

BASE_URL = "http://localhost:8085/api/v1"

# Login
print("Logging in...")
login_resp = requests.post(f"{BASE_URL}/auth/login", 
    json={"username": "admin", "password": "admin"})
session_token = login_resp.json()["token"]
headers = {"Authorization": f"Bearer {session_token}"}

# Data categories
entity_types = ["server", "app", "db", "network", "storage"]
departments = ["eng", "ops", "sec", "finance"]
statuses = ["active", "inactive", "critical"]
envs = ["prod", "dev", "test"]

print("\nCreating 1000 entities quickly...")
start = time.time()
created_ids = []

for i in range(1000):
    tags = [
        f"type:{random.choice(entity_types)}",
        f"id:quick_{i}",
        f"dept:{random.choice(departments)}",
        f"status:{random.choice(statuses)}",
        f"env:{random.choice(envs)}",
        f"batch:{i//100}"
    ]
    
    if i % 5 == 0:
        tags.append("priority:high")
    if i % 10 == 0:
        tags.append("security:sensitive")
    
    resp = requests.post(f"{BASE_URL}/entities/create",
        headers=headers, json={"tags": tags})
    
    if resp.status_code == 201 and i % 100 == 0:
        data = resp.json()
        entity_id = data.get("id") or data.get("entity", {}).get("id")
        if entity_id:
            created_ids.append(entity_id)
        print(f"Created {i+1} entities...")

create_time = time.time() - start
print(f"\nCreated 1000 entities in {create_time:.1f}s ({1000/create_time:.1f} per sec)")

# Complex queries
print("\n=== Complex Query Tests ===")

# Multi-tag queries
queries = [
    ("Prod servers", "filter=env:prod&filter=type:server"),
    ("Critical DBs", "filter=type:db&filter=status:critical"),
    ("High priority", "filter=priority:high&limit=5"),
    ("Security sensitive", "filter=security:sensitive&filter=dept:eng"),
    ("Batch 3", "filter=batch:3&limit=10")
]

for name, query in queries:
    resp = requests.get(f"{BASE_URL}/entities/query?{query}", headers=headers)
    if resp.status_code == 200:
        count = len(resp.json().get("entities", []))
        print(f"{name}: {count} results")

# Temporal queries
if created_ids:
    print("\n=== Temporal Queries ===")
    test_id = created_ids[0]
    
    # History
    resp = requests.get(f"{BASE_URL}/entities/history?entity_id={test_id}", 
        headers=headers)
    if resp.status_code == 200:
        history = resp.json().get("history", [])
        print(f"Entity history: {len(history)} states")
    
    # As-of query
    from datetime import datetime, timedelta
    past_time = (datetime.utcnow() - timedelta(seconds=30)).isoformat() + "Z"
    resp = requests.get(
        f"{BASE_URL}/entities/as-of?entity_id={test_id}&timestamp={past_time}",
        headers=headers)
    print(f"As-of query: {resp.status_code}")

# Relationships
print("\n=== Relationship Tests ===")
rel_count = 0
for i in range(min(50, len(created_ids)-1)):
    resp = requests.post(f"{BASE_URL}/entity-relationships",
        headers=headers,
        json={
            "from_entity_id": created_ids[i],
            "to_entity_id": created_ids[(i+1) % len(created_ids)],
            "relationship_type": "depends_on"
        })
    if resp.status_code == 201:
        rel_count += 1

print(f"Created {rel_count} relationships")

# Final stats
print(f"\n=== Summary ===")
print(f"Total time: {time.time() - start:.1f}s")
print(f"Entities: 1000")
print(f"Relationships: {rel_count}")
print("Complex queries: Tested")
print("Temporal queries: Tested")
print("\nEntityDB stress test complete!")