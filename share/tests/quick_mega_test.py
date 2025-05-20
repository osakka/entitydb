#!/usr/bin/env python3
"""
Quick version of mega stress test
- 10,000 entities
- 30,000 relationships
"""

import requests
import time
import random
import json
import statistics
from datetime import datetime

BASE_URL = "http://localhost:8085"

# Entity types
ENTITY_TYPES = ["user", "project", "task", "document", "sensor"]
RELATIONSHIP_TYPES = ["owns", "manages", "assigned_to", "depends_on", "references"]

class QuickMegaTest:
    def __init__(self):
        self.session = requests.Session()
        self.entities = []
        self.stats = {
            "entity_creation_times": [],
            "relationship_creation_times": [],
            "query_times": []
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
        return False
    
    def create_entities(self, count=10000):
        """Create entities"""
        print(f"\n=== Creating {count} Entities ===")
        start_time = time.time()
        
        for i in range(count):
            entity_type = random.choice(ENTITY_TYPES)
            
            tags = [
                f"type:{entity_type}",
                f"status:{random.choice(['active', 'pending', 'completed'])}",
                f"priority:{random.choice(['high', 'medium', 'low'])}",
                f"department:{random.choice(['eng', 'sales', 'hr', 'it'])}",
                f"region:{random.choice(['us', 'eu', 'asia'])}",
                f"rbac:perm:{entity_type}:view",
                f"rbac:perm:{entity_type}:edit"
            ]
            
            # Add random tags
            for _ in range(random.randint(1, 3)):
                tags.append(f"custom{random.randint(1,10)}:value{random.randint(1,100)}")
            
            data = {
                "tags": tags,
                "content": [{
                    "type": "title",
                    "value": f"{entity_type}_{i}"
                }]
            }
            
            entity_start = time.time()
            response = self.session.post(f"{BASE_URL}/api/v1/entities/create", json=data)
            entity_end = time.time()
            
            if response.status_code == 201:
                entity = response.json()
                self.entities.append(entity["id"])
                self.stats["entity_creation_times"].append(entity_end - entity_start)
                
                if (i + 1) % 100 == 0:
                    avg_time = statistics.mean(self.stats["entity_creation_times"][-100:])
                    print(f"  Created {i + 1} entities... (avg: {avg_time*1000:.2f}ms)")
            else:
                print(f"  Failed to create entity: {response.status_code}")
        
        end_time = time.time()
        print(f"\n✓ Created {len(self.entities)} entities in {end_time - start_time:.2f} seconds")
        print(f"  Average: {statistics.mean(self.stats['entity_creation_times'])*1000:.2f}ms")
    
    def create_relationships(self, count=30000):
        """Create relationships"""
        print(f"\n=== Creating {count} Relationships ===")
        start_time = time.time()
        created = 0
        
        for i in range(count):
            source = random.choice(self.entities)
            target = random.choice(self.entities)
            
            while target == source:
                target = random.choice(self.entities)
            
            data = {
                "source_id": source,
                "relationship_type": random.choice(RELATIONSHIP_TYPES),
                "target_id": target
            }
            
            rel_start = time.time()
            response = self.session.post(f"{BASE_URL}/api/v1/entity-relationships", json=data)
            rel_end = time.time()
            
            if response.status_code in [200, 201]:
                created += 1
                self.stats["relationship_creation_times"].append(rel_end - rel_start)
                
                if created % 1000 == 0:
                    avg_time = statistics.mean(self.stats["relationship_creation_times"][-1000:])
                    print(f"  Created {created} relationships... (avg: {avg_time*1000:.2f}ms)")
        
        end_time = time.time()
        print(f"\n✓ Created {created} relationships in {end_time - start_time:.2f} seconds")
        print(f"  Average: {statistics.mean(self.stats['relationship_creation_times'])*1000:.2f}ms")
    
    def run_performance_tests(self):
        """Run various performance tests"""
        print("\n=== Performance Tests ===")
        
        tests = [
            ("List all entities", f"{BASE_URL}/api/v1/entities/list", {}),
            ("Query by type:user", f"{BASE_URL}/api/v1/entities/list?tag=type:user", {}),
            ("Wildcard type:*", f"{BASE_URL}/api/v1/entities/list?wildcard=type:*", {}),
            ("Namespace 'rbac'", f"{BASE_URL}/api/v1/entities/list?namespace=rbac", {}),
            ("Complex query", f"{BASE_URL}/api/v1/entities/query", {
                "filter": "tag:type",
                "operator": "eq", 
                "value": "user",
                "sort": "created_at",
                "order": "desc",
                "limit": "100"
            })
        ]
        
        for test_name, url, params in tests:
            start_time = time.time()
            if params:
                response = self.session.get(url, params=params)
            else:
                response = self.session.get(url)
            end_time = time.time()
            
            query_time = end_time - start_time
            self.stats["query_times"].append(query_time)
            
            result_count = len(response.json()) if response.status_code == 200 else 0
            print(f"{test_name}: {query_time*1000:.2f}ms ({result_count} results)")
    
    def print_summary(self):
        """Print summary statistics"""
        print("\n" + "="*50)
        print("=== PERFORMANCE SUMMARY ===")
        print("="*50)
        
        print(f"\nTotal Entities: {len(self.entities)}")
        print(f"Total Relationships: {len(self.stats['relationship_creation_times'])}")
        
        print("\nEntity Creation:")
        if self.stats['entity_creation_times']:
            print(f"  Average: {statistics.mean(self.stats['entity_creation_times'])*1000:.2f}ms")
            print(f"  Median: {statistics.median(self.stats['entity_creation_times'])*1000:.2f}ms")
            print(f"  Min: {min(self.stats['entity_creation_times'])*1000:.2f}ms") 
            print(f"  Max: {max(self.stats['entity_creation_times'])*1000:.2f}ms")
            
            # Calculate throughput
            total_time = sum(self.stats['entity_creation_times'])
            throughput = len(self.entities) / total_time if total_time > 0 else 0
            print(f"  Throughput: {throughput:.2f} entities/second")
        
        print("\nRelationship Creation:")
        if self.stats['relationship_creation_times']:
            print(f"  Average: {statistics.mean(self.stats['relationship_creation_times'])*1000:.2f}ms")
            print(f"  Median: {statistics.median(self.stats['relationship_creation_times'])*1000:.2f}ms")
            print(f"  Min: {min(self.stats['relationship_creation_times'])*1000:.2f}ms")
            print(f"  Max: {max(self.stats['relationship_creation_times'])*1000:.2f}ms")
            
            # Calculate throughput
            total_time = sum(self.stats['relationship_creation_times'])
            throughput = len(self.stats['relationship_creation_times']) / total_time if total_time > 0 else 0
            print(f"  Throughput: {throughput:.2f} relationships/second")
        
        print("\nQuery Performance:")
        if self.stats['query_times']:
            print(f"  Average: {statistics.mean(self.stats['query_times'])*1000:.2f}ms")
            print(f"  Median: {statistics.median(self.stats['query_times'])*1000:.2f}ms")
            print(f"  Min: {min(self.stats['query_times'])*1000:.2f}ms")
            print(f"  Max: {max(self.stats['query_times'])*1000:.2f}ms")
        
        print("\n✅ TEST COMPLETED!")
    
    def run(self):
        """Run the test"""
        print("=== EntityDB v2.10.0 Quick Mega Test ===")
        
        if not self.login_as_admin():
            print("Failed to login")
            return
        
        try:
            self.create_entities(10000)
            self.create_relationships(30000)
            self.run_performance_tests()
            self.print_summary()
        except KeyboardInterrupt:
            print("\nTest interrupted")
            self.print_summary()
        except Exception as e:
            print(f"\nError: {e}")
            import traceback
            traceback.print_exc()
            self.print_summary()

if __name__ == "__main__":
    test = QuickMegaTest()
    test.run()