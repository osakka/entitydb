#!/usr/bin/env python3
"""
EntityDB API Client Example

Demonstrates how to use the EntityDB API client programmatically
"""

from entitydb_client import EntityDBClient

# Create client
client = EntityDBClient(server="http://localhost:8085")

# Print line separator
def print_separator():
    print("\n" + "=" * 50 + "\n")

def main():
    # Login
    print("Logging in...\n")
    login_result = client.login("admin", "password")
    print(f"Login result: {login_result}")
    print_separator()

    # List entities
    print("Listing entities...\n")
    entities = client.list_entities(limit=5)
    print(f"Entities: {len(entities.get('data', []))}")
    for entity in entities.get('data', []):
        print(f"  - {entity.get('id')}: {entity.get('title')} ({entity.get('type')})")
    print_separator()

    # Create an entity
    print("Creating a new entity...\n")
    new_entity = client.create_entity(
        entity_type="issue",
        title="Test Issue",
        description="This is a test issue created via the Python client",
        tags=["priority:medium", "status:pending"],
        properties={"assignee": "bot", "points": 3}
    )
    print(f"New entity: {new_entity}")
    
    # Extract entity ID from response
    entity_id = None
    if new_entity.get('status') == 'ok' and 'entity' in new_entity:
        entity_id = new_entity['entity'].get('id')
    elif new_entity.get('status') == 'ok' and 'data' in new_entity:
        entity_id = new_entity['data'].get('id')
    
    if entity_id:
        print(f"Created entity with ID: {entity_id}")
        print_separator()

        # Get entity details
        print(f"Getting entity details for {entity_id}...\n")
        entity_details = client.get_entity(entity_id)
        print(f"Entity details: {entity_details}")
        print_separator()

        # Update entity
        print(f"Updating entity {entity_id}...\n")
        update_result = client.update_entity(
            entity_id=entity_id,
            title="Updated Test Issue",
            tags=["priority:high", "status:in_progress"]
        )
        print(f"Update result: {update_result}")
        print_separator()

        # Create a relationship
        print("Creating a relationship...\n")
        
        # First, let's find another entity to create a relationship with
        other_entities = client.list_entities(
            entity_type="issue", 
            limit=1
        )
        
        if other_entities.get('data') and len(other_entities['data']) > 0:
            other_entity = other_entities['data'][0]
            other_id = other_entity.get('id')
            
            if other_id and other_id != entity_id:
                print(f"Found other entity: {other_id}")
                
                # Create dependency relationship
                relationship = client.create_relationship(
                    source_id=entity_id,
                    target_id=other_id,
                    relationship_type="depends_on",
                    properties={"critical": True}
                )
                print(f"Relationship created: {relationship}")
                print_separator()
                
                # List relationships
                print("Listing relationships...\n")
                relationships = client.list_relationships(source_id=entity_id)
                print(f"Relationships: {relationships}")
                print_separator()
    
    # List all entity types
    print("Listing entity types...\n")
    entity_types = {
        "issue": "Issues",
        "agent": "Agents", 
        "workspace": "Workspaces",
        "session": "Sessions"
    }
    
    for type_key, type_name in entity_types.items():
        entities = client.list_entities(entity_type=type_key, limit=3)
        count = len(entities.get('data', []))
        print(f"{type_name}: {count} found")
        for entity in entities.get('data', [])[:3]:  # Show up to 3
            print(f"  - {entity.get('id')}: {entity.get('title')}")
    
    print_separator()
    print("Example completed successfully!")

if __name__ == "__main__":
    main()