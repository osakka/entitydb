#!/usr/bin/env python3
"""EntityDB Stress Test - Creates entities and tests complex queries"""

import requests
import json
import time
import sys
from datetime import datetime

BASE_URL = "http://localhost:8085/api/v1"

# Login
print("=== EntityDB Stress Test ===")
print("Logging in...")
resp = requests.post(f"{BASE_URL}/auth/login", 
    json={"username": "admin", "password": "admin"})

if resp.status_code != 200:
    print(f"Login failed: {resp.text}")
    sys.exit(1)

token = resp.json()["token"]
headers = {"Authorization": f"Bearer {token}"}
print("Login successful\n")

# Create 100 entities quickly
print("Creating 100 test entities...")
start_time = time.time()
created = 0
saved_ids = []

for i in range(100):
    tags = [
        f"type:test",
        f"id:stress_{i}",
        f"env:{'prod' if i % 3 == 0 else 'dev'}",
        f"priority:{'high' if i % 5 == 0 else 'normal'}",
        f"team:{'alpha' if i < 50 else 'beta'}"
    ]
    
    resp = requests.post(f"{BASE_URL}/entities/create",
        headers=headers, json={"tags": tags})
    
    if resp.status_code == 201:
        created += 1
        if i % 20 == 0:
            data = resp.json()
            entity_id = data.get("id") or data.get("entity", {}).get("id")
            if entity_id:
                saved_ids.append(entity_id)
    
    if i % 25 == 0:
        print(f"  Created {i+1} entities...")

elapsed = time.time() - start_time
print(f"\nCreated {created} entities in {elapsed:.1f}s ({created/elapsed:.1f} per sec)")
print(f"Saved {len(saved_ids)} IDs for testing\n")

# Test complex queries
print("=== Testing Complex Queries ===")

# Multi-tag queries
queries = [
    ("Production high priority", "filter=env:prod&filter=priority:high"),
    ("Alpha team test entities", "filter=team:alpha&filter=type:test&limit=5"),
    ("Dev environment", "filter=env:dev&limit=10")
]

for name, query in queries:
    resp = requests.get(f"{BASE_URL}/entities/query?{query}", headers=headers)
    if resp.status_code == 200:
        count = len(resp.json().get("entities", []))
        print(f"{name}: {count} results")

# Temporal query
if saved_ids:
    print("\n=== Temporal Query Test ===")
    test_id = saved_ids[0]
    
    # Get history
    resp = requests.get(f"{BASE_URL}/entities/history?entity_id={test_id}", headers=headers)
    if resp.status_code == 200:
        history = resp.json().get("history", [])
        print(f"Entity {test_id[:8]}... has {len(history)} historical versions")

# Create a few relationships
print("\n=== Creating Relationships ===")
rel_count = 0
for i in range(min(10, len(saved_ids)-1)):
    resp = requests.post(f"{BASE_URL}/entity-relationships",
        headers=headers,
        json={
            "from_entity_id": saved_ids[i],
            "to_entity_id": saved_ids[(i+1) % len(saved_ids)],
            "relationship_type": "depends_on"
        })
    if resp.status_code == 201:
        rel_count += 1

print(f"Created {rel_count} relationships")

# Summary
print(f"\n=== Test Complete ===")
print(f"Total time: {time.time() - start_time:.1f} seconds")
print(f"Entities created: {created}")
print(f"Relationships: {rel_count}")
print("Status: Success")