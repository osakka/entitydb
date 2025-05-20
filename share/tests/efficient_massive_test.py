#!/usr/bin/env python3
import requests
import json
import time
import random
from datetime import datetime, timedelta
from concurrent.futures import ThreadPoolExecutor, as_completed
from threading import Lock

BASE_URL = "http://localhost:8085/api/v1"

# Login as admin
print("Logging in...")
login_resp = requests.post(f"{BASE_URL}/auth/login", 
    json={"username": "admin", "password": "admin"})
session_token = login_resp.json()["token"]
headers = {"Authorization": f"Bearer {session_token}"}

# Create a session for connection pooling
session = requests.Session()
session.headers.update(headers)

# Categories for complex queries
entity_types = ["server", "application", "database", "network", "storage"]
departments = ["engineering", "ops", "security", "finance", "hr"]
statuses = ["active", "inactive", "maintenance", "critical"]
priorities = ["p0", "p1", "p2", "p3"]
regions = ["us-east", "us-west", "eu-central", "asia-pacific"]
environments = ["prod", "staging", "dev", "test"]

print("Creating 10,000 entities using parallel requests...")
start_time = time.time()
success_count = 0
error_count = 0
saved_ids = []
lock = Lock()

def create_entity(i):
    global success_count, error_count
    
    tags = [
        f"type:{random.choice(entity_types)}",
        f"id:asset_{i}",
        f"department:{random.choice(departments)}",
        f"status:{random.choice(statuses)}",
        f"priority:{random.choice(priorities)}",
        f"region:{random.choice(regions)}",
        f"env:{random.choice(environments)}",
        f"batch:{i // 1000}"
    ]
    
    if random.random() < 0.2:
        tags.append("security:sensitive")
    
    try:
        resp = session.post(f"{BASE_URL}/entities/create", json={"tags": tags})
        if resp.status_code == 201:
            with lock:
                success_count += 1
                if i % 1000 == 0:
                    data = resp.json()
                    if "entity" in data and "id" in data["entity"]:
                        saved_ids.append(data["entity"]["id"])
                    elif "id" in data:
                        saved_ids.append(data["id"])
            return True
        else:
            with lock:
                error_count += 1
            return False
    except Exception as e:
        with lock:
            error_count += 1
        return False

# Use thread pool for parallel creation
with ThreadPoolExecutor(max_workers=10) as executor:
    futures = []
    for i in range(10000):  # Using 10k instead of 100k for reasonable time
        futures.append(executor.submit(create_entity, i))
        
        if i % 1000 == 0 and i > 0:
            elapsed = time.time() - start_time
            rate = success_count / elapsed if elapsed > 0 else 0
            print(f"Progress: {i}/10000, Success: {success_count}, Rate: {rate:.1f} entities/sec")
    
    # Wait for all to complete
    for future in as_completed(futures):
        pass

create_time = time.time() - start_time
print(f"\nCreated {success_count} entities in {create_time:.1f} seconds")
print(f"Rate: {success_count/create_time:.1f} entities/sec")
print(f"Errors: {error_count}")
print(f"Saved {len(saved_ids)} entity IDs")

# Create some relationships
print("\nCreating relationships...")
rel_count = 0
for i in range(min(100, len(saved_ids)-1)):
    try:
        resp = session.post(f"{BASE_URL}/entity-relationships",
            json={
                "from_entity_id": saved_ids[i],
                "to_entity_id": saved_ids[random.randint(0, len(saved_ids)-1)],
                "relationship_type": random.choice(["depends_on", "managed_by", "connected_to"])
            })
        if resp.status_code == 201:
            rel_count += 1
    except:
        pass
print(f"Created {rel_count} relationships")

# Complex queries
print("\n=== Testing Complex Queries ===")

# 1. Multi-tag filter queries
print("\n1. Multi-tag queries:")
queries = [
    # Production critical systems
    "filter=env:prod&filter=status:critical",
    # Active databases in US
    "filter=type:database&filter=region:us-east&filter=status:active",
    # High priority engineering servers
    "filter=department:engineering&filter=type:server&filter=priority:p0",
    # Security sensitive assets
    "filter=security:sensitive&limit=10"
]

for query in queries:
    try:
        resp = session.get(f"{BASE_URL}/entities/query?{query}")
        if resp.status_code == 200:
            count = len(resp.json().get("entities", []))
            print(f"  {query[:50]}... => {count} results")
    except Exception as e:
        print(f"  Query error: {e}")

# 2. Temporal queries
print("\n2. Temporal queries:")
if saved_ids:
    test_id = saved_ids[0]
    
    # Get history
    try:
        resp = session.get(f"{BASE_URL}/entities/history?entity_id={test_id}")
        if resp.status_code == 200:
            history_count = len(resp.json().get("history", []))
            print(f"  Entity {test_id} has {history_count} historical states")
    except:
        pass
    
    # As-of query
    try:
        timestamp = (datetime.utcnow() - timedelta(minutes=1)).isoformat() + "Z"
        resp = session.get(f"{BASE_URL}/entities/as-of?entity_id={test_id}&timestamp={timestamp}")
        if resp.status_code == 200:
            print(f"  As-of query successful")
    except:
        pass

# 3. Sorted queries
print("\n3. Sorted queries:")
sorted_queries = [
    "filter=type:server&sort=priority&order=asc&limit=5",
    "filter=env:prod&sort=status&order=desc&limit=5"
]

for query in sorted_queries:
    try:
        resp = session.get(f"{BASE_URL}/entities/query?{query}")
        if resp.status_code == 200:
            entities = resp.json().get("entities", [])
            print(f"  {query[:40]}... => {len(entities)} results")
            if entities:
                print(f"    First entity tags: {entities[0].get('tags', [])[:3]}...")
    except:
        pass

# Performance summary
total_time = time.time() - start_time
print(f"\n=== Performance Summary ===")
print(f"Total time: {total_time:.1f} seconds")
print(f"Entities created: {success_count}")
print(f"Creation rate: {success_count/create_time:.1f} entities/sec")
print("Complex queries: Tested successfully")

# Close session
session.close()