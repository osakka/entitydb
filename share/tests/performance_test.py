#!/usr/bin/env python3
import requests
import time
import json
import uuid
import random
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed

# Configuration
BASE_URL = "http://localhost:8085"
USERNAME = "admin"
PASSWORD = "admin"

# Number of entities to create
NUM_ENTITIES = 1000
# Number of queries to perform
NUM_QUERIES = 10000
# Number of concurrent threads
NUM_THREADS = 10

class PerformanceTester:
    def __init__(self):
        self.session = requests.Session()
        self.created_entities = []
        self.token = None
        
    def login(self):
        """Login and get session token"""
        response = self.session.post(f"{BASE_URL}/api/v1/auth/login", 
                                   json={"username": USERNAME, "password": PASSWORD})
        response.raise_for_status()
        data = response.json()
        self.token = data["token"]
        self.session.headers.update({"Authorization": f"Bearer {self.token}"})
        print(f"Logged in successfully")
        
    def create_entity(self, index):
        """Create a test entity"""
        entity_data = {
            "tags": [
                f"1732018871547277568|type:test",
                f"1732018871547277568|test:entity",
                f"1732018871547277568|index:{index}",
                f"1732018871547277568|batch:{index // 100}",
            ],
            "content": [
                {
                    "type": "test_data",
                    "value": f"Test entity {index} with random data {uuid.uuid4()}",
                    "timestamp": "2025-01-01T00:00:00Z"
                }
            ]
        }
        
        response = self.session.post(f"{BASE_URL}/api/v1/entities/create",
                                   json=entity_data)
        if response.status_code == 201:
            entity = response.json()
            return entity["id"]
        return None
        
    def query_by_id(self, entity_id):
        """Query entity by ID"""
        start_time = time.time()
        response = self.session.get(f"{BASE_URL}/api/v1/entities/get",
                                  params={"id": entity_id})
        query_time = (time.time() - start_time) * 1000  # Convert to ms
        return query_time if response.status_code == 200 else None
        
    def query_by_tag(self, tag):
        """Query entities by tag"""
        start_time = time.time()
        response = self.session.get(f"{BASE_URL}/api/v1/entities/list",
                                  params={"tag": tag})
        query_time = (time.time() - start_time) * 1000  # Convert to ms
        return query_time if response.status_code == 200 else None
        
    def run_performance_test(self):
        """Run the complete performance test"""
        print(f"Starting performance test...")
        
        # Phase 1: Create entities
        print(f"\nPhase 1: Creating {NUM_ENTITIES} entities...")
        start_time = time.time()
        
        with ThreadPoolExecutor(max_workers=NUM_THREADS) as executor:
            futures = []
            for i in range(NUM_ENTITIES):
                future = executor.submit(self.create_entity, i)
                futures.append(future)
            
            for future in as_completed(futures):
                entity_id = future.result()
                if entity_id:
                    self.created_entities.append(entity_id)
        
        creation_time = time.time() - start_time
        print(f"Created {len(self.created_entities)} entities in {creation_time:.2f} seconds")
        print(f"Average creation time: {creation_time/len(self.created_entities)*1000:.2f} ms per entity")
        
        # Phase 2: Query by ID
        print(f"\nPhase 2: Performing {NUM_QUERIES} random ID queries...")
        query_times = []
        start_time = time.time()
        
        with ThreadPoolExecutor(max_workers=NUM_THREADS) as executor:
            futures = []
            for _ in range(NUM_QUERIES):
                entity_id = random.choice(self.created_entities)
                future = executor.submit(self.query_by_id, entity_id)
                futures.append(future)
            
            for future in as_completed(futures):
                query_time = future.result()
                if query_time:
                    query_times.append(query_time)
        
        query_phase_time = time.time() - start_time
        
        if query_times:
            avg_query_time = sum(query_times) / len(query_times)
            min_query_time = min(query_times)
            max_query_time = max(query_times)
            
            print(f"Completed {len(query_times)} queries in {query_phase_time:.2f} seconds")
            print(f"Query performance:")
            print(f"  Average: {avg_query_time:.2f} ms")
            print(f"  Min: {min_query_time:.2f} ms")
            print(f"  Max: {max_query_time:.2f} ms")
            print(f"  Queries per second: {len(query_times)/query_phase_time:.2f}")
        
        # Phase 3: Query by tag
        print(f"\nPhase 3: Performing tag-based queries...")
        tag_query_times = []
        start_time = time.time()
        
        # Query for different batch tags
        for batch in range(10):
            query_time = self.query_by_tag(f"batch:{batch}")
            if query_time:
                tag_query_times.append(query_time)
        
        tag_phase_time = time.time() - start_time
        
        if tag_query_times:
            avg_tag_query_time = sum(tag_query_times) / len(tag_query_times)
            print(f"Completed {len(tag_query_times)} tag queries in {tag_phase_time:.2f} seconds")
            print(f"Average tag query time: {avg_tag_query_time:.2f} ms")
            
        # Summary
        print("\n=== PERFORMANCE SUMMARY ===")
        print(f"Entities created: {len(self.created_entities)}")
        print(f"Total queries performed: {len(query_times)}")
        print(f"Average query latency: {avg_query_time:.2f} ms")
        
        # Check if we achieved 100x improvement (target < 2ms from ~189ms baseline)
        if avg_query_time < 2:
            print(f"\n✅ TURBO MODE SUCCESS: Achieved {189/avg_query_time:.1f}x performance improvement!")
        else:
            print(f"\n⚠️  Performance improvement: {189/avg_query_time:.1f}x (target was 100x)")

def main():
    tester = PerformanceTester()
    
    try:
        tester.login()
        tester.run_performance_test()
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()