#!/usr/bin/env python3
"""
Mega stress test for EntityDB v2.10.0
- 100,000 entities
- 300,000 relationships
- Multiple tag types
- RBAC permissions
- Performance measurements
"""

import requests
import time
import random
import string
import json
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime
import statistics

BASE_URL = "http://localhost:8085"

# Entity types and their tag patterns
ENTITY_TYPES = {
    "user": {
        "tags": ["type:user", "status:active"],
        "permissions": ["rbac:role:user", "rbac:perm:entity:view"],
        "content": ["name", "email", "department"]
    },
    "project": {
        "tags": ["type:project", "status:in_progress"],
        "permissions": ["rbac:perm:project:view", "rbac:perm:project:edit"],
        "content": ["title", "description", "deadline"]
    },
    "task": {
        "tags": ["type:task", "priority:high", "status:pending"],
        "permissions": ["rbac:perm:task:view", "rbac:perm:task:update"],
        "content": ["title", "description", "estimate"]
    },
    "document": {
        "tags": ["type:document", "visibility:public"],
        "permissions": ["rbac:perm:document:read"],
        "content": ["title", "content", "version"]
    },
    "sensor": {
        "tags": ["type:sensor", "location:building-a", "status:online"],
        "permissions": ["rbac:perm:sensor:read"],
        "content": ["model", "data"]
    }
}

# Relationship types
RELATIONSHIP_TYPES = [
    "owns",
    "manages",
    "assigned_to",
    "depends_on",
    "references",
    "monitors",
    "approves",
    "collaborates_with"
]

# Tag namespaces for variety
TAG_NAMESPACES = ["category", "region", "department", "team", "phase", "version", "env"]
TAG_VALUES = ["alpha", "beta", "prod", "dev", "test", "qa", "staging", "research", "sales", "engineering"]

class MegaStressTest:
    def __init__(self):
        self.session = requests.Session()
        self.entities = []
        self.relationships = []
        self.stats = {
            "entity_creation_times": [],
            "relationship_creation_times": [],
            "query_times": [],
            "temporal_query_times": []
        }
        
    def login_as_admin(self):
        """Login and get admin token"""
        print("Logging in as admin...")
        response = self.session.post(f"{BASE_URL}/api/v1/auth/login",
                                   json={"username": "admin", "password": "admin"})
        if response.status_code == 200:
            token = response.json()["token"]
            self.session.headers.update({"Authorization": f"Bearer {token}"})
            print("✓ Logged in successfully")
            return True
        print(f"✗ Login failed: {response.status_code}")
        return False
    
    def generate_entity_data(self, entity_type):
        """Generate random entity data based on type"""
        type_info = ENTITY_TYPES[entity_type]
        
        # Base tags from type
        tags = type_info["tags"].copy()
        
        # Add random tags
        for _ in range(random.randint(2, 5)):
            namespace = random.choice(TAG_NAMESPACES)
            value = random.choice(TAG_VALUES)
            tags.append(f"{namespace}:{value}")
        
        # Add permissions
        tags.extend(type_info["permissions"])
        
        # Add some random permissions
        if random.random() > 0.7:
            tags.append(f"rbac:perm:custom:{random.choice(['read', 'write', 'delete'])}")
        
        # Generate content
        content = []
        for content_type in type_info["content"]:
            content.append({
                "type": content_type,
                "value": f"{content_type}_{random.randint(1000, 9999)}"
            })
        
        return {
            "tags": tags,
            "content": content
        }
    
    def create_entities_batch(self, start_idx, count, entity_type):
        """Create a batch of entities"""
        results = []
        for i in range(count):
            data = self.generate_entity_data(entity_type)
            
            start_time = time.time()
            response = self.session.post(f"{BASE_URL}/api/v1/entities/create", json=data)
            end_time = time.time()
            
            if response.status_code == 201:
                entity = response.json()
                results.append(entity["id"])
                self.stats["entity_creation_times"].append(end_time - start_time)
                
                if (start_idx + i + 1) % 1000 == 0:
                    print(f"  Created {start_idx + i + 1} entities...")
            else:
                print(f"  Failed to create entity: {response.status_code}")
                
        return results
    
    def create_all_entities(self):
        """Create 100k entities across all types"""
        print("\n=== Creating 100,000 Entities ===")
        start_time = time.time()
        
        entities_per_type = 20000  # 5 types * 20k = 100k
        
        with ThreadPoolExecutor(max_workers=10) as executor:
            futures = []
            
            for entity_type in ENTITY_TYPES:
                print(f"Creating {entities_per_type} {entity_type} entities...")
                
                # Create in batches of 1000
                for i in range(0, entities_per_type, 1000):
                    future = executor.submit(
                        self.create_entities_batch,
                        i, min(1000, entities_per_type - i), entity_type
                    )
                    futures.append(future)
            
            # Collect results
            for future in as_completed(futures):
                self.entities.extend(future.result())
        
        end_time = time.time()
        print(f"\n✓ Created {len(self.entities)} entities in {end_time - start_time:.2f} seconds")
        print(f"  Average creation time: {statistics.mean(self.stats['entity_creation_times'])*1000:.2f}ms")
    
    def create_relationships_batch(self, count):
        """Create a batch of relationships"""
        results = []
        
        for _ in range(count):
            source = random.choice(self.entities)
            target = random.choice(self.entities)
            
            # Avoid self-relationships
            while target == source:
                target = random.choice(self.entities)
            
            relationship_type = random.choice(RELATIONSHIP_TYPES)
            
            data = {
                "source_id": source,
                "relationship_type": relationship_type,
                "target_id": target
            }
            
            start_time = time.time()
            response = self.session.post(f"{BASE_URL}/api/v1/entity-relationships", json=data)
            end_time = time.time()
            
            if response.status_code in [200, 201]:
                results.append(response.json())
                self.stats["relationship_creation_times"].append(end_time - start_time)
            
        return len(results)
    
    def create_all_relationships(self):
        """Create 300k relationships"""
        print("\n=== Creating 300,000 Relationships ===")
        start_time = time.time()
        
        total_created = 0
        batch_size = 1000
        total_relationships = 300000
        
        with ThreadPoolExecutor(max_workers=20) as executor:
            futures = []
            
            for i in range(0, total_relationships, batch_size):
                future = executor.submit(
                    self.create_relationships_batch,
                    min(batch_size, total_relationships - i)
                )
                futures.append(future)
                
                if (i + batch_size) % 10000 == 0:
                    print(f"  Queued {i + batch_size} relationships...")
            
            # Collect results
            for future in as_completed(futures):
                total_created += future.result()
                if total_created % 10000 == 0:
                    print(f"  Created {total_created} relationships...")
        
        end_time = time.time()
        print(f"\n✓ Created {total_created} relationships in {end_time - start_time:.2f} seconds")
        print(f"  Average creation time: {statistics.mean(self.stats['relationship_creation_times'])*1000:.2f}ms")
    
    def run_query_tests(self):
        """Run various query performance tests"""
        print("\n=== Running Query Performance Tests ===")
        
        # Test 1: List all entities
        print("\n1. List all entities:")
        start_time = time.time()
        response = self.session.get(f"{BASE_URL}/api/v1/entities/list")
        end_time = time.time()
        query_time = end_time - start_time
        self.stats["query_times"].append(query_time)
        print(f"  Time: {query_time*1000:.2f}ms")
        print(f"  Count: {len(response.json())}")
        
        # Test 2: Query by tag
        print("\n2. Query by specific tag:")
        for tag in ["type:user", "status:active", "priority:high"]:
            start_time = time.time()
            response = self.session.get(f"{BASE_URL}/api/v1/entities/list?tag={tag}")
            end_time = time.time()
            query_time = end_time - start_time
            self.stats["query_times"].append(query_time)
            print(f"  Tag '{tag}': {query_time*1000:.2f}ms ({len(response.json())} results)")
        
        # Test 3: Wildcard queries
        print("\n3. Wildcard queries:")
        for pattern in ["type:*", "rbac:perm:*", "status:*"]:
            start_time = time.time()
            response = self.session.get(f"{BASE_URL}/api/v1/entities/list?wildcard={pattern}")
            end_time = time.time()
            query_time = end_time - start_time
            self.stats["query_times"].append(query_time)
            print(f"  Pattern '{pattern}': {query_time*1000:.2f}ms ({len(response.json())} results)")
        
        # Test 4: Namespace queries
        print("\n4. Namespace queries:")
        for namespace in ["type", "rbac", "status", "category"]:
            start_time = time.time()
            response = self.session.get(f"{BASE_URL}/api/v1/entities/list?namespace={namespace}")
            end_time = time.time()
            query_time = end_time - start_time
            self.stats["query_times"].append(query_time)
            print(f"  Namespace '{namespace}': {query_time*1000:.2f}ms ({len(response.json())} results)")
        
        # Test 5: Complex queries with sorting
        print("\n5. Complex queries with sorting:")
        params = {
            "filter": "tag:type",
            "operator": "eq",
            "value": "user",
            "sort": "created_at",
            "order": "desc",
            "limit": "100"
        }
        start_time = time.time()
        response = self.session.get(f"{BASE_URL}/api/v1/entities/query", params=params)
        end_time = time.time()
        query_time = end_time - start_time
        self.stats["query_times"].append(query_time)
        print(f"  Complex query: {query_time*1000:.2f}ms ({len(response.json())} results)")
    
    def run_temporal_tests(self):
        """Run temporal query tests"""
        print("\n=== Running Temporal Query Tests ===")
        
        # Pick a random entity for temporal tests
        test_entity = random.choice(self.entities[:1000])  # From early entities
        
        # Test 1: Entity history
        print(f"\n1. Entity history for {test_entity}:")
        start_time = time.time()
        response = self.session.get(f"{BASE_URL}/api/v1/entities/history?id={test_entity}")
        end_time = time.time()
        query_time = end_time - start_time
        self.stats["temporal_query_times"].append(query_time)
        print(f"  Time: {query_time*1000:.2f}ms")
        
        # Test 2: Recent changes
        print("\n2. Recent changes (last hour):")
        start_time = time.time()
        response = self.session.get(f"{BASE_URL}/api/v1/entities/changes")
        end_time = time.time()
        query_time = end_time - start_time
        self.stats["temporal_query_times"].append(query_time)
        print(f"  Time: {query_time*1000:.2f}ms")
        
        # Test 3: As-of queries
        print("\n3. As-of query (5 minutes ago):")
        as_of = datetime.utcnow().replace(minute=datetime.utcnow().minute - 5).isoformat() + "Z"
        start_time = time.time()
        response = self.session.get(f"{BASE_URL}/api/v1/entities/as-of?id={test_entity}&as_of={as_of}")
        end_time = time.time()
        query_time = end_time - start_time
        self.stats["temporal_query_times"].append(query_time)
        print(f"  Time: {query_time*1000:.2f}ms")
    
    def run_relationship_tests(self):
        """Run relationship query tests"""
        print("\n=== Running Relationship Query Tests ===")
        
        # Pick some entities for relationship tests
        test_entities = random.sample(self.entities[:1000], 10)
        
        print("\n1. Get relationships by source:")
        rel_times = []
        for entity in test_entities:
            start_time = time.time()
            response = self.session.get(f"{BASE_URL}/api/v1/entity-relationships?source={entity}")
            end_time = time.time()
            query_time = end_time - start_time
            rel_times.append(query_time)
            print(f"  Entity {entity}: {query_time*1000:.2f}ms ({len(response.json())} relationships)")
        
        print(f"  Average: {statistics.mean(rel_times)*1000:.2f}ms")
        
        print("\n2. Get relationships by target:")
        rel_times = []
        for entity in test_entities:
            start_time = time.time()
            response = self.session.get(f"{BASE_URL}/api/v1/entity-relationships?target={entity}")
            end_time = time.time()
            query_time = end_time - start_time
            rel_times.append(query_time)
            print(f"  Entity {entity}: {query_time*1000:.2f}ms ({len(response.json())} relationships)")
        
        print(f"  Average: {statistics.mean(rel_times)*1000:.2f}ms")
    
    def print_summary(self):
        """Print performance summary"""
        print("\n" + "="*50)
        print("=== MEGA STRESS TEST SUMMARY ===")
        print("="*50)
        
        print(f"\nTotal Entities: {len(self.entities)}")
        print(f"Total Relationships: {len(self.stats['relationship_creation_times'])}")
        
        print("\nEntity Creation Performance:")
        print(f"  Average: {statistics.mean(self.stats['entity_creation_times'])*1000:.2f}ms")
        print(f"  Median: {statistics.median(self.stats['entity_creation_times'])*1000:.2f}ms")
        print(f"  Min: {min(self.stats['entity_creation_times'])*1000:.2f}ms")
        print(f"  Max: {max(self.stats['entity_creation_times'])*1000:.2f}ms")
        
        print("\nRelationship Creation Performance:")
        print(f"  Average: {statistics.mean(self.stats['relationship_creation_times'])*1000:.2f}ms")
        print(f"  Median: {statistics.median(self.stats['relationship_creation_times'])*1000:.2f}ms")
        print(f"  Min: {min(self.stats['relationship_creation_times'])*1000:.2f}ms")
        print(f"  Max: {max(self.stats['relationship_creation_times'])*1000:.2f}ms")
        
        print("\nQuery Performance:")
        print(f"  Average: {statistics.mean(self.stats['query_times'])*1000:.2f}ms")
        print(f"  Median: {statistics.median(self.stats['query_times'])*1000:.2f}ms")
        print(f"  Min: {min(self.stats['query_times'])*1000:.2f}ms")
        print(f"  Max: {max(self.stats['query_times'])*1000:.2f}ms")
        
        if self.stats['temporal_query_times']:
            print("\nTemporal Query Performance:")
            print(f"  Average: {statistics.mean(self.stats['temporal_query_times'])*1000:.2f}ms")
            print(f"  Median: {statistics.median(self.stats['temporal_query_times'])*1000:.2f}ms")
        
        # Calculate throughput
        total_ops = len(self.entities) + len(self.stats['relationship_creation_times'])
        total_time = sum(self.stats['entity_creation_times']) + sum(self.stats['relationship_creation_times'])
        throughput = total_ops / total_time if total_time > 0 else 0
        
        print(f"\nOverall Throughput: {throughput:.2f} operations/second")
        
        print("\n✅ MEGA STRESS TEST COMPLETED!")
    
    def run(self):
        """Run the complete mega stress test"""
        print("=== EntityDB v2.10.0 Mega Stress Test ===")
        print("Target: 100k entities, 300k relationships")
        print("Features: Multiple tag types, RBAC permissions, temporal queries")
        
        if not self.login_as_admin():
            print("Failed to login, aborting test")
            return
        
        try:
            self.create_all_entities()
            self.create_all_relationships()
            self.run_query_tests()
            self.run_temporal_tests()
            self.run_relationship_tests()
            self.print_summary()
            
        except KeyboardInterrupt:
            print("\n\nTest interrupted by user")
            self.print_summary()
        except Exception as e:
            print(f"\n\nError during test: {e}")
            import traceback
            traceback.print_exc()
            self.print_summary()


if __name__ == "__main__":
    test = MegaStressTest()
    test.run()