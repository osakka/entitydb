#!/usr/bin/env python3
import requests
import json
import time
import sys

# Base configuration
BASE_URL = "http://localhost:8085/api/v1"
BATCH_SIZE = 100
TOTAL_ENTITIES = 1000  # Start smaller

# Login
print("Logging in...")
login_resp = requests.post(f"{BASE_URL}/auth/login", 
    json={"username": "admin", "password": "admin"})

if login_resp.status_code != 200:
    print(f"Login failed: {login_resp.status_code}")
    sys.exit(1)

token = login_resp.json()["token"]
headers = {"Authorization": f"Bearer {token}"}
print(f"Logged in successfully")

# Create entities in batches
print(f"\nCreating {TOTAL_ENTITIES} entities in batches of {BATCH_SIZE}...")
start_time = time.time()
created_count = 0
saved_ids = []

for batch_num in range(0, TOTAL_ENTITIES, BATCH_SIZE):
    batch_start = time.time()
    
    for i in range(batch_num, min(batch_num + BATCH_SIZE, TOTAL_ENTITIES)):
        # Simple tags for speed
        tags = [
            f"type:entity",
            f"id:test_{i}",
            f"batch:{batch_num // BATCH_SIZE}",
            f"index:{i}"
        ]
        
        try:
            resp = requests.post(f"{BASE_URL}/entities/create",
                headers=headers,
                json={"tags": tags},
                timeout=5)
            
            if resp.status_code == 201:
                created_count += 1
                if i % 100 == 0:
                    data = resp.json()
                    entity_id = data.get("id") or data.get("entity", {}).get("id")
                    if entity_id:
                        saved_ids.append(entity_id)
        except requests.Timeout:
            print(f"Timeout at entity {i}")
        except Exception as e:
            print(f"Error at entity {i}: {e}")
    
    batch_time = time.time() - batch_start
    elapsed = time.time() - start_time
    rate = created_count / elapsed if elapsed > 0 else 0
    print(f"Batch {batch_num//BATCH_SIZE + 1}: Created {created_count} total in {elapsed:.1f}s ({rate:.1f}/sec)")

print(f"\nSuccessfully created {created_count} entities")
print(f"Saved {len(saved_ids)} IDs for testing")

# Test some queries
print("\n=== Testing Queries ===")

# Simple query test
try:
    resp = requests.get(f"{BASE_URL}/entities/query?filter=type:entity&limit=5",
        headers=headers, timeout=5)
    if resp.status_code == 200:
        count = len(resp.json().get("entities", []))
        print(f"Simple query returned {count} entities")
except Exception as e:
    print(f"Query error: {e}")

# Batch query test
try:
    resp = requests.get(f"{BASE_URL}/entities/query?filter=batch:0&limit=10",
        headers=headers, timeout=5)
    if resp.status_code == 200:
        count = len(resp.json().get("entities", []))
        print(f"Batch query returned {count} entities")
except Exception as e:
    print(f"Query error: {e}")

# Summary
total_time = time.time() - start_time
print(f"\n=== Summary ===")
print(f"Total time: {total_time:.1f} seconds")
print(f"Entities created: {created_count}")
print(f"Average rate: {created_count/total_time:.1f} entities/sec")
print("Status: Complete")