#!/usr/bin/env python3
import requests
import time
import json
import uuid
import random
from datetime import datetime, timedelta

# Configuration
BASE_URL = "http://localhost:8085"
USERNAME = "admin"
PASSWORD = "admin"

class TemporalPerformanceTester:
    def __init__(self):
        self.session = requests.Session()
        self.token = None
        self.entities = []
        
    def login(self):
        """Login and get session token"""
        response = self.session.post(f"{BASE_URL}/api/v1/auth/login", 
                                   json={"username": USERNAME, "password": PASSWORD})
        response.raise_for_status()
        data = response.json()
        self.token = data["token"]
        self.session.headers.update({"Authorization": f"Bearer {self.token}"})
        print(f"Logged in successfully")
        
    def create_temporal_entity(self, timestamp):
        """Create an entity with specific timestamp"""
        # Convert datetime to nanosecond timestamp
        nano_timestamp = int(timestamp.timestamp() * 1e9)
        
        entity_data = {
            "tags": [
                f"{nano_timestamp}|type:temporal_test",
                f"{nano_timestamp}|event:data_point",
                f"{nano_timestamp}|value:{random.randint(1, 100)}"
            ],
            "content": []
        }
        
        response = self.session.post(f"{BASE_URL}/api/v1/entities/create",
                                   json=entity_data)
        if response.status_code == 201:
            return response.json()["id"]
        return None
        
    def run_temporal_performance_test(self):
        """Test temporal query performance"""
        print("Starting temporal performance test...")
        
        # Phase 1: Create entities across time
        print("\nPhase 1: Creating temporal entities...")
        start_time = datetime.now() - timedelta(days=30)
        
        for i in range(100):
            # Create entities spread across 30 days
            timestamp = start_time + timedelta(hours=i*7.2)  # ~5 per day
            entity_id = self.create_temporal_entity(timestamp)
            if entity_id:
                self.entities.append((entity_id, timestamp))
            
            if (i + 1) % 10 == 0:
                print(f"  Created {i + 1} temporal entities...")
        
        print(f"Created {len(self.entities)} temporal entities")
        
        # Phase 2: Test temporal queries
        print("\nPhase 2: Testing temporal query performance...")
        
        # Test 1: Point-in-time queries
        print("\nTest 1: Point-in-time queries (finding entity state at specific times)")
        query_times = []
        
        for i in range(20):
            # Pick a random time in our range
            random_time = start_time + timedelta(days=random.randint(0, 30))
            
            start = time.time()
            # This would use as-of query in real implementation
            response = self.session.get(f"{BASE_URL}/api/v1/entities/list",
                                      params={"tag": "type:temporal_test", 
                                             "include_timestamps": "true"})
            query_time = (time.time() - start) * 1000
            query_times.append(query_time)
            
            if response.status_code == 200:
                entities = response.json()
                # In a real temporal query, we'd filter by timestamp
                print(f"  Query {i+1}: {query_time:.2f}ms - Found {len(entities)} entities")
        
        avg_query_time = sum(query_times) / len(query_times) if query_times else 0
        print(f"\nAverage point-in-time query: {avg_query_time:.2f}ms")
        
        # Test 2: Range queries
        print("\nTest 2: Range queries (finding changes within time periods)")
        range_times = []
        
        for i in range(10):
            # Pick a random 7-day range
            range_start = start_time + timedelta(days=random.randint(0, 23))
            range_end = range_start + timedelta(days=7)
            
            start = time.time()
            # This would use temporal range query in real implementation
            response = self.session.get(f"{BASE_URL}/api/v1/entities/list",
                                      params={"tag": "type:temporal_test",
                                             "include_timestamps": "true"})
            query_time = (time.time() - start) * 1000
            range_times.append(query_time)
            
            if response.status_code == 200:
                entities = response.json()
                print(f"  Range query {i+1}: {query_time:.2f}ms - Found {len(entities)} entities")
        
        avg_range_time = sum(range_times) / len(range_times) if range_times else 0
        print(f"\nAverage range query: {avg_range_time:.2f}ms")
        
        # Test 3: Timeline navigation
        print("\nTest 3: Timeline navigation (finding next/previous events)")
        nav_times = []
        
        for i in range(10):
            # Pick a random entity and time
            entity_id, base_time = random.choice(self.entities)
            
            start = time.time()
            # This would use temporal navigation in real implementation
            response = self.session.get(f"{BASE_URL}/api/v1/entities/get",
                                      params={"id": entity_id,
                                             "include_timestamps": "true"})
            query_time = (time.time() - start) * 1000
            nav_times.append(query_time)
            
            if response.status_code == 200:
                print(f"  Navigation {i+1}: {query_time:.2f}ms")
        
        avg_nav_time = sum(nav_times) / len(nav_times) if nav_times else 0
        print(f"\nAverage timeline navigation: {avg_nav_time:.2f}ms")
        
        # Summary
        print("\n=== TEMPORAL PERFORMANCE SUMMARY ===")
        print(f"Entities created: {len(self.entities)}")
        print(f"Point-in-time queries: {avg_query_time:.2f}ms avg")
        print(f"Range queries: {avg_range_time:.2f}ms avg")
        print(f"Timeline navigation: {avg_nav_time:.2f}ms avg")
        
        # With temporal optimization, these should be significantly faster
        print("\nPotential optimizations:")
        print("1. Temporal indexes could reduce point-in-time queries to <1ms")
        print("2. Time-bucketed indexes could make range queries 10x faster")
        print("3. Sorted temporal data structures could enable instant navigation")

def main():
    tester = TemporalPerformanceTester()
    
    try:
        tester.login()
        tester.run_temporal_performance_test()
    except Exception as e:
        print(f"Error: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    main()