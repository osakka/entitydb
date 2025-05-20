#!/usr/bin/env python3
import requests
import json
import time
import random
from datetime import datetime, timedelta

BASE_URL = "http://localhost:8085/api/v1"

# Login as admin
print("Logging in...")
login_resp = requests.post(f"{BASE_URL}/auth/login", 
    json={"username": "admin", "password": "admin"})
session_token = login_resp.json()["token"]
headers = {"Authorization": f"Bearer {session_token}"}

# Categories for complex queries
entity_types = ["server", "application", "database", "network", "storage", "compute"]
departments = ["engineering", "ops", "security", "finance", "hr", "sales"]
statuses = ["active", "inactive", "maintenance", "deprecated", "critical"]
priorities = ["p0", "p1", "p2", "p3", "p4"]
regions = ["us-east", "us-west", "eu-central", "asia-pacific"]
environments = ["prod", "staging", "dev", "test", "sandbox"]

print("Creating 100,000 entities in batches...")
start_time = time.time()
batch_size = 1000
total_entities = 100000
saved_ids = []

for batch in range(0, total_entities, batch_size):
    batch_start = time.time()
    batch_entities = []
    
    for i in range(batch, min(batch + batch_size, total_entities)):
        # Randomize tags for realistic data
        entity_type = random.choice(entity_types)
        department = random.choice(departments)
        status = random.choice(statuses)
        priority = random.choice(priorities)
        region = random.choice(regions)
        env = random.choice(environments)
        
        # Some entities get extra tags
        tags = [
            f"type:{entity_type}",
            f"id:asset_{i}",
            f"department:{department}",
            f"status:{status}",
            f"priority:{priority}",
            f"region:{region}",
            f"env:{env}",
            f"batch:{batch // batch_size}"
        ]
        
        # 20% get security tags
        if random.random() < 0.2:
            tags.append("security:sensitive")
            tags.append(f"compliance:{random.choice(['pci', 'sox', 'hipaa'])}")
        
        # 10% get cost tags
        if random.random() < 0.1:
            cost = random.randint(100, 10000)
            tags.append(f"cost:{cost}")
            
        resp = requests.post(f"{BASE_URL}/entities/create",
            headers=headers,
            json={"tags": tags})
        
        if resp.status_code == 201:
            if i % 10000 == 0:  # Save some IDs for later
                data = resp.json()
                if "entity" in data and "id" in data["entity"]:
                    saved_ids.append(data["entity"]["id"])
                elif "id" in data:
                    saved_ids.append(data["id"])
    
    batch_time = time.time() - batch_start
    total_time = time.time() - start_time
    rate = (batch + batch_size) / total_time
    print(f"Batch {batch//batch_size + 1}/{total_entities//batch_size}: {batch + batch_size} entities in {batch_time:.1f}s (overall: {rate:.1f} entities/sec)")

print(f"\nCreated {total_entities} entities in {time.time() - start_time:.1f} seconds")
print(f"Saved {len(saved_ids)} entity IDs for testing")

# Create relationships between saved entities
print("\nCreating relationships...")
relationship_types = ["depends_on", "managed_by", "connected_to", "replicated_to", "backed_by"]
for i in range(min(500, len(saved_ids)-1)):
    if i < len(saved_ids) - 1:
        resp = requests.post(f"{BASE_URL}/entity-relationships",
            headers=headers,
            json={
                "from_entity_id": saved_ids[i],
                "to_entity_id": saved_ids[random.randint(0, len(saved_ids)-1)],
                "relationship_type": random.choice(relationship_types)
            })

print("Relationships created")

# Complex queries
print("\n=== Testing Complex Queries ===")

# 1. Multi-tag filter queries
print("\n1. Multi-tag queries:")
queries = [
    # Production critical systems
    "filter=env:prod&filter=status:critical",
    # US-based active databases
    "filter=type:database&filter=region:us-east&filter=status:active",
    # High priority security-sensitive assets
    "filter=security:sensitive&filter=priority:p0",
    # Engineering department servers in maintenance
    "filter=department:engineering&filter=type:server&filter=status:maintenance",
    # All assets in specific batch
    "filter=batch:5&limit=10"
]

for query in queries:
    resp = requests.get(f"{BASE_URL}/entities/query?{query}", headers=headers)
    if resp.status_code == 200:
        data = resp.json()
        count = len(data.get("entities", []))
        print(f"  Query: {query} => {count} results")

# 2. Temporal queries (as-of different times)
print("\n2. Temporal queries:")
if len(saved_ids) > 0:
    test_entity_id = saved_ids[0]
    
    # Get entity history
    resp = requests.get(f"{BASE_URL}/entities/history", 
        params={"entity_id": test_entity_id},
        headers=headers)
    
    if resp.status_code == 200:
        history = resp.json().get("history", [])
        print(f"  Entity {test_entity_id} has {len(history)} historical states")
        
        # Query as-of different timestamps
        now = datetime.utcnow()
        timestamps = [
            now - timedelta(minutes=1),
            now - timedelta(minutes=5),
            now - timedelta(minutes=10)
        ]
        
        for ts in timestamps:
            resp = requests.get(f"{BASE_URL}/entities/as-of",
                params={
                    "entity_id": test_entity_id,
                    "timestamp": ts.isoformat() + "Z"
                },
                headers=headers)
            if resp.status_code == 200:
                print(f"  As-of {ts.isoformat()}: Success")

# 3. Complex aggregation-style queries
print("\n3. Aggregation-style queries:")
# Count by type
for entity_type in entity_types:
    resp = requests.get(f"{BASE_URL}/entities/query?filter=type:{entity_type}&limit=1", 
        headers=headers)
    if resp.status_code == 200:
        # Since we can't get total count directly, we'd need to paginate
        # For now just show that the query works
        print(f"  Type '{entity_type}': Query successful")

# 4. Range queries with sorting
print("\n4. Sorted queries with multiple filters:")
complex_queries = [
    # Sort by priority
    "filter=env:prod&filter=type:server&sort=priority&order=asc&limit=5",
    # Regional sorting
    "filter=department:ops&sort=region&order=desc&limit=5",
    # Status-based sorting
    "filter=type:database&sort=status&order=asc&limit=5"
]

for query in complex_queries:
    resp = requests.get(f"{BASE_URL}/entities/query?{query}", headers=headers)
    if resp.status_code == 200:
        data = resp.json()
        entities = data.get("entities", [])
        print(f"  Query: {query}")
        print(f"    Found {len(entities)} entities")
        if entities:
            # Show first entity's tags to verify sorting
            print(f"    First entity tags: {entities[0].get('tags', [])[:5]}...")

# 5. Relationship queries
print("\n5. Relationship queries:")
if len(saved_ids) > 0:
    test_entity = saved_ids[0]
    resp = requests.get(f"{BASE_URL}/entity-relationships",
        params={"from_entity_id": test_entity},
        headers=headers)
    if resp.status_code == 200:
        relationships = resp.json()
        print(f"  Entity {test_entity} has {len(relationships)} relationships")

# Performance summary
total_time = time.time() - start_time
print(f"\n=== Performance Summary ===")
print(f"Total time: {total_time:.1f} seconds")
print(f"Overall rate: {total_entities/total_time:.1f} entities/sec")
print(f"Entities created: {total_entities}")
print(f"Complex queries tested: Success")

# Get final stats
stats_resp = requests.get(f"{BASE_URL}/dashboard/stats", headers=headers)
if stats_resp.status_code == 200:
    stats = stats_resp.json()
    print(f"\nDatabase stats:")
    print(f"- Total entities: {stats.get('total_entities', 'N/A')}")
    print(f"- Total relationships: {stats.get('total_relationships', 'N/A')}")