#!/usr/bin/env python3
"""
EntityDB v2.10.0 Turbo Performance Numbers
Shows real performance metrics
"""

import requests
import time
import random
import statistics
import json

BASE_URL = "http://localhost:8085"

class PerformanceTest:
    def __init__(self):
        self.session = requests.Session()
        self.entities = []
        
    def login(self):
        response = self.session.post(f"{BASE_URL}/api/v1/auth/login",
                                   json={"username": "admin", "password": "admin"})
        if response.status_code == 200:
            token = response.json()["token"]
            self.session.headers.update({"Authorization": f"Bearer {token}"})
            return True
        return False
    
    def test_entity_creation(self, count=1000):
        """Test entity creation performance"""
        print(f"\n=== Testing Entity Creation ({count} entities) ===")
        
        times = []
        batch_times = []
        batch_start = time.time()
        
        for i in range(count):
            data = {
                "tags": [
                    f"type:performance-test",
                    f"index:{i}",
                    f"status:{random.choice(['active', 'pending', 'completed'])}",
                    f"priority:{random.choice(['high', 'medium', 'low'])}",
                    f"department:{random.choice(['eng', 'sales', 'hr', 'it', 'finance'])}",
                    f"region:{random.choice(['us-east', 'us-west', 'eu', 'asia'])}",
                    f"rbac:perm:view",
                    f"rbac:perm:edit",
                    f"custom1:value{random.randint(1,100)}",
                    f"custom2:value{random.randint(1,100)}"
                ],
                "content": [{
                    "type": "title",
                    "value": f"Performance Test Entity {i}"
                }, {
                    "type": "description",
                    "value": f"This is a test entity created at {time.time()}"
                }]
            }
            
            start = time.time()
            response = self.session.post(f"{BASE_URL}/api/v1/entities/create", json=data)
            end = time.time()
            
            if response.status_code == 201:
                self.entities.append(response.json()["id"])
                times.append(end - start)
                
                if (i + 1) % 100 == 0:
                    batch_end = time.time()
                    batch_time = batch_end - batch_start
                    batch_times.append(batch_time)
                    
                    avg = statistics.mean(times[-100:]) * 1000
                    print(f"  {i + 1} entities: Last 100 avg: {avg:.2f}ms, Batch time: {batch_time:.2f}s")
                    batch_start = time.time()
        
        # Calculate statistics
        total_time = sum(times)
        avg_time = statistics.mean(times) * 1000
        median_time = statistics.median(times) * 1000
        p95_time = sorted(times)[int(len(times) * 0.95)] * 1000
        p99_time = sorted(times)[int(len(times) * 0.99)] * 1000
        min_time = min(times) * 1000
        max_time = max(times) * 1000
        throughput = len(times) / total_time
        
        print(f"\n🔥 Entity Creation Performance:")
        print(f"  Total created: {len(self.entities)}")
        print(f"  Average: {avg_time:.2f}ms")
        print(f"  Median: {median_time:.2f}ms")
        print(f"  P95: {p95_time:.2f}ms")
        print(f"  P99: {p99_time:.2f}ms")
        print(f"  Min: {min_time:.2f}ms")
        print(f"  Max: {max_time:.2f}ms")
        print(f"  Throughput: {throughput:.2f} entities/second")
        print(f"  Total time: {total_time:.2f}s")
    
    def test_relationships(self, count=3000):
        """Test relationship creation performance"""
        print(f"\n=== Testing Relationship Creation ({count} relationships) ===")
        
        if len(self.entities) < 100:
            print("Not enough entities, skipping relationship test")
            return
        
        times = []
        created = 0
        
        for i in range(count):
            source = random.choice(self.entities)
            target = random.choice(self.entities)
            
            while target == source:
                target = random.choice(self.entities)
            
            data = {
                "source_id": source,
                "relationship_type": random.choice([
                    "owns", "manages", "assigned_to", "depends_on", 
                    "references", "collaborates_with", "approves"
                ]),
                "target_id": target
            }
            
            start = time.time()
            response = self.session.post(f"{BASE_URL}/api/v1/entity-relationships", json=data)
            end = time.time()
            
            if response.status_code in [200, 201]:
                created += 1
                times.append(end - start)
                
                if created % 500 == 0:
                    avg = statistics.mean(times[-500:]) * 1000
                    print(f"  {created} relationships: Last 500 avg: {avg:.2f}ms")
        
        if times:
            avg_time = statistics.mean(times) * 1000
            median_time = statistics.median(times) * 1000
            p95_time = sorted(times)[int(len(times) * 0.95)] * 1000
            p99_time = sorted(times)[int(len(times) * 0.99)] * 1000
            throughput = len(times) / sum(times)
            
            print(f"\n🔥 Relationship Creation Performance:")
            print(f"  Total created: {created}")
            print(f"  Average: {avg_time:.2f}ms")
            print(f"  Median: {median_time:.2f}ms")
            print(f"  P95: {p95_time:.2f}ms")
            print(f"  P99: {p99_time:.2f}ms")
            print(f"  Throughput: {throughput:.2f} relationships/second")
    
    def test_queries(self):
        """Test query performance"""
        print("\n=== Testing Query Performance ===")
        
        test_queries = [
            ("List all entities", f"{BASE_URL}/api/v1/entities/list", None, 1),
            ("List with limit", f"{BASE_URL}/api/v1/entities/list?limit=100", None, 5),
            ("Query by type", f"{BASE_URL}/api/v1/entities/list?tag=type:performance-test", None, 5),
            ("Query by status", f"{BASE_URL}/api/v1/entities/list?tag=status:active", None, 5),
            ("Wildcard type:*", f"{BASE_URL}/api/v1/entities/list?wildcard=type:*", None, 5),
            ("Wildcard status:*", f"{BASE_URL}/api/v1/entities/list?wildcard=status:*", None, 5),
            ("Namespace 'rbac'", f"{BASE_URL}/api/v1/entities/list?namespace=rbac", None, 5),
            ("Namespace 'department'", f"{BASE_URL}/api/v1/entities/list?namespace=department", None, 5),
            ("Complex query", f"{BASE_URL}/api/v1/entities/query", {
                "filter": "tag:type",
                "operator": "eq",
                "value": "performance-test",
                "sort": "created_at",
                "order": "desc",
                "limit": "100"
            }, 5)
        ]
        
        for name, url, params, runs in test_queries:
            times = []
            result_counts = []
            
            for _ in range(runs):
                start = time.time()
                if params:
                    response = self.session.get(url, params=params)
                else:
                    response = self.session.get(url)
                end = time.time()
                
                if response.status_code == 200:
                    times.append(end - start)
                    result_counts.append(len(response.json()))
            
            if times:
                avg_time = statistics.mean(times) * 1000
                result_count = result_counts[0]  # Should be consistent
                
                print(f"\n{name}:")
                print(f"  Average: {avg_time:.2f}ms ({runs} runs)")
                print(f"  Results: {result_count} entities")
    
    def test_temporal_queries(self):
        """Test temporal query performance"""
        print("\n=== Testing Temporal Query Performance ===")
        
        if not self.entities:
            print("No entities to test temporal queries")
            return
        
        # Pick a few entities for testing
        test_entities = random.sample(self.entities, min(10, len(self.entities)))
        
        # Test history queries
        print("\nEntity History:")
        history_times = []
        for entity_id in test_entities[:5]:
            start = time.time()
            response = self.session.get(f"{BASE_URL}/api/v1/entities/history?id={entity_id}")
            end = time.time()
            
            if response.status_code == 200:
                history_times.append(end - start)
        
        if history_times:
            avg_time = statistics.mean(history_times) * 1000
            print(f"  Average: {avg_time:.2f}ms ({len(history_times)} entities)")
        
        # Test recent changes
        print("\nRecent Changes:")
        changes_times = []
        for _ in range(5):
            start = time.time()
            response = self.session.get(f"{BASE_URL}/api/v1/entities/changes")
            end = time.time()
            
            if response.status_code == 200:
                changes_times.append(end - start)
        
        if changes_times:
            avg_time = statistics.mean(changes_times) * 1000
            print(f"  Average: {avg_time:.2f}ms ({len(changes_times)} runs)")
    
    def run(self):
        """Run all performance tests"""
        print("🚀 EntityDB v2.10.0 Performance Test")
        print("Repository: Temporal Turbo with B-tree indexes")
        print("Features: Memory-mapped files, Skip-lists, Bloom filters")
        
        if not self.login():
            print("Failed to login")
            return
        
        try:
            # Entity creation test
            self.test_entity_creation(1000)
            
            # Relationship creation test
            self.test_relationships(3000)
            
            # Query performance test
            self.test_queries()
            
            # Temporal query test
            self.test_temporal_queries()
            
            print("\n" + "="*50)
            print("✅ Performance Test Complete!")
            print("="*50)
            
        except Exception as e:
            print(f"\nError: {e}")
            import traceback
            traceback.print_exc()

if __name__ == "__main__":
    test = PerformanceTest()
    test.run()