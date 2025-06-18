#!/usr/bin/env python3
"""
Controlled Load Test
===================

Test our temporal retention fix with controlled, measured load to isolate
the issue between retention and high-frequency entity updates.
"""

import requests
import time
import json

class ControlledTester:
    def __init__(self):
        self.session = requests.Session()
        self.session.verify = False
        self.authenticate()
    
    def authenticate(self):
        response = self.session.post(
            "https://localhost:8085/api/v1/auth/login",
            json={"username": "admin", "password": "admin"}
        )
        if response.status_code == 200:
            token = response.json()["token"]
            self.session.headers.update({"Authorization": f"Bearer {token}"})
            print("✅ Authenticated")
        else:
            raise Exception(f"Auth failed: {response.text}")
    
    def create_test_entity(self):
        """Create a single test entity"""
        response = self.session.post(
            "https://localhost:8085/api/v1/entities/create",
            json={
                "tags": ["type:test", "dataset:load-test", "name:controlled-test"],
                "content": "Controlled load test entity"
            }
        )
        if response.status_code == 201:
            entity_id = response.json()["id"]
            print(f"✅ Created test entity: {entity_id}")
            return entity_id
        else:
            raise Exception(f"Failed to create entity: {response.text}")
    
    def add_single_metric(self, entity_id, value):
        """Add a single metric value via entity update"""
        # Get current entity
        response = self.session.get(f"https://localhost:8085/api/v1/entities/get?id={entity_id}")
        if response.status_code != 200:
            print(f"❌ Failed to get entity: {response.text}")
            return False
        
        entity = response.json()
        current_tags = entity.get("tags", [])
        
        # Add new value tag
        new_tag = f"value:{value}"
        updated_tags = current_tags + [new_tag]
        
        # Update entity
        update_data = {
            "id": entity_id,
            "tags": updated_tags,
            "content": entity["content"]
        }
        
        response = self.session.put("https://localhost:8085/api/v1/entities/update", json=update_data)
        return response.status_code == 200
    
    def test_low_frequency_load(self):
        """Test with low frequency updates (1 per second)"""
        print("\n🧪 Testing LOW frequency load (1 update/second)...")
        
        entity_id = self.create_test_entity()
        
        for i in range(10):  # 10 updates over 10 seconds
            success = self.add_single_metric(entity_id, i)
            print(f"  Update {i+1}: {'✅' if success else '❌'}")
            time.sleep(1)  # 1 second between updates
        
        print("✅ Low frequency test complete")
    
    def test_medium_frequency_load(self):
        """Test with medium frequency updates (1 per 100ms)"""
        print("\n🧪 Testing MEDIUM frequency load (10 updates/second)...")
        
        entity_id = self.create_test_entity()
        
        for i in range(30):  # 30 updates over 3 seconds
            success = self.add_single_metric(entity_id, i + 100)
            if i % 10 == 0:
                print(f"  Update {i+1}: {'✅' if success else '❌'}")
            time.sleep(0.1)  # 100ms between updates
        
        print("✅ Medium frequency test complete")
    
    def test_high_frequency_load(self):
        """Test with high frequency updates (1 per 10ms)"""
        print("\n🧪 Testing HIGH frequency load (100 updates/second)...")
        
        entity_id = self.create_test_entity()
        
        failed_count = 0
        for i in range(50):  # 50 updates over 0.5 seconds
            success = self.add_single_metric(entity_id, i + 200)
            if not success:
                failed_count += 1
            if i % 25 == 0:
                print(f"  Update {i+1}: {'✅' if success else '❌'}")
            time.sleep(0.01)  # 10ms between updates
        
        print(f"✅ High frequency test complete - {failed_count} failures")
        return failed_count

def main():
    print("🎯 Controlled Load Test - Isolating Temporal Retention Issues")
    print("=" * 60)
    
    tester = ControlledTester()
    
    # Test progression from low to high frequency
    tester.test_low_frequency_load()
    time.sleep(2)
    
    tester.test_medium_frequency_load()
    time.sleep(2)
    
    failures = tester.test_high_frequency_load()
    
    print(f"\n📊 Results:")
    print(f"   Low frequency (1/sec): Expected to work perfectly")
    print(f"   Medium frequency (10/sec): Expected to work well")
    print(f"   High frequency (100/sec): {failures} failures - this reveals the threshold")

if __name__ == "__main__":
    main()