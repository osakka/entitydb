#!/usr/bin/env python3
import requests
import json

# Configuration
BASE_URL = "http://localhost:8085"

session = requests.Session()

# Login
response = session.post(f"{BASE_URL}/api/v1/auth/login", 
                       json={"username": "admin", "password": "admin"})
token = response.json()["token"]
session.headers.update({"Authorization": f"Bearer {token}"})
print("Logged in")

# Create entity
print("\nCreating entity...")
create_response = session.post(f"{BASE_URL}/api/v1/entities/create",
                              json={
                                  "tags": ["type:debug", "test:simple"],
                                  "content": []
                              })

print(f"Create status: {create_response.status_code}")
print(f"Create response: {create_response.text}")

if create_response.status_code == 201:
    entity = create_response.json()
    entity_id = entity["id"]
    
    # Try to get it immediately
    print(f"\nGetting entity {entity_id}...")
    get_response = session.get(f"{BASE_URL}/api/v1/entities/get",
                              params={"id": entity_id})
    
    print(f"Get status: {get_response.status_code}")
    print(f"Get response: {get_response.text}")
    
    # Also try listing
    print("\nListing all entities with type:debug...")
    list_response = session.get(f"{BASE_URL}/api/v1/entities/list",
                               params={"tag": "type:debug"})
    
    print(f"List status: {list_response.status_code}")
    print(f"List response: {list_response.text}")
    
    # Check with timestamps
    print("\nGetting with timestamps...")
    ts_response = session.get(f"{BASE_URL}/api/v1/entities/get",
                             params={"id": entity_id, "include_timestamps": "true"})
    
    print(f"Timestamps status: {ts_response.status_code}")
    print(f"Timestamps response: {ts_response.text}")