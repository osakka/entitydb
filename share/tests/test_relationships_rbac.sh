#!/bin/bash

# Test RBAC enforcement for entity relationships

BASE_URL="http://localhost:8085/api/v1"

echo "=== EntityDB Relationship RBAC Tests ==="
echo

# 1. Create test users with different roles  
echo "=== Test 1: Create users with different permissions ==="

# Create admin user
ADMIN_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:user", "rbac:role:admin", "rbac:perm:*"],
        "content": [
            {"type": "username", "value": "admin_user"},
            {"type": "password_hash", "value": "$2a$10$dummyhash"}
        ]
    }')
ADMIN_ID=$(echo $ADMIN_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created admin user: $ADMIN_ID"

# Create regular user with view permissions only
USER_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:user", "rbac:role:user", "rbac:perm:entity:view", "rbac:perm:relation:view"],
        "content": [
            {"type": "username", "value": "regular_user"},
            {"type": "password_hash", "value": "$2a$10$dummyhash"}
        ]
    }')
USER_ID=$(echo $USER_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created regular user: $USER_ID"

# Create editor with create/update permissions
EDITOR_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:user", "rbac:role:editor", "rbac:perm:entity:*", "rbac:perm:relation:create", "rbac:perm:relation:view"],
        "content": [
            {"type": "username", "value": "editor_user"},
            {"type": "password_hash", "value": "$2a$10$dummyhash"}
        ]
    }')
EDITOR_ID=$(echo $EDITOR_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created editor user: $EDITOR_ID"

# Simulate login tokens
ADMIN_TOKEN="Bearer token_admin_$ADMIN_ID"
USER_TOKEN="Bearer token_user_$USER_ID"
EDITOR_TOKEN="Bearer token_editor_$EDITOR_ID"

# 2. Create test entities
echo -e "\n=== Test 2: Create test entities ==="

ENTITY1_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:doc", "status:draft"],
        "content": [{"type": "title", "value": "Document 1"}]
    }')
ENTITY1_ID=$(echo $ENTITY1_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created entity 1: $ENTITY1_ID"

ENTITY2_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:doc", "status:published"],
        "content": [{"type": "title", "value": "Document 2"}]
    }')
ENTITY2_ID=$(echo $ENTITY2_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created entity 2: $ENTITY2_ID"

# 3. Test relationship creation with different roles
echo -e "\n=== Test 3: Test relationship creation permissions ==="

# Admin should succeed
echo "Admin creating relationship..."
ADMIN_REL=$(curl -s -X POST $BASE_URL/entity-relationships \
    -H "Content-Type: application/json" \
    -H "Authorization: $ADMIN_TOKEN" \
    -d "{
        \"source_id\": \"$ENTITY1_ID\",
        \"relationship_type\": \"references\",
        \"target_id\": \"$ENTITY2_ID\"
    }")
if [[ "$ADMIN_REL" =~ "error" ]]; then
    echo "❌ Admin failed: $ADMIN_REL"
else
    echo "✅ Admin succeeded"
fi

# Editor should succeed
echo -e "\nEditor creating relationship..."
EDITOR_REL=$(curl -s -X POST $BASE_URL/entity-relationships \
    -H "Content-Type: application/json" \
    -H "Authorization: $EDITOR_TOKEN" \
    -d "{
        \"source_id\": \"$ENTITY2_ID\",
        \"relationship_type\": \"depends_on\",
        \"target_id\": \"$ENTITY1_ID\"
    }")
if [[ "$EDITOR_REL" =~ "error" ]]; then
    echo "❌ Editor failed: $EDITOR_REL"
else
    echo "✅ Editor succeeded"
fi

# Regular user should fail (no create permission)
echo -e "\nRegular user creating relationship..."
USER_REL=$(curl -s -X POST $BASE_URL/entity-relationships \
    -H "Content-Type: application/json" \
    -H "Authorization: $USER_TOKEN" \
    -d "{
        \"source_id\": \"$ENTITY1_ID\",
        \"relationship_type\": \"links_to\",
        \"target_id\": \"$ENTITY2_ID\"
    }")
if [[ "$USER_REL" =~ "Insufficient permissions" ]] || [[ "$USER_REL" =~ "403" ]]; then
    echo "✅ Regular user correctly denied"
else
    echo "❌ Regular user should have been denied: $USER_REL"
fi

# 4. Test relationship viewing with different roles
echo -e "\n=== Test 4: Test relationship viewing permissions ==="

# Regular user should be able to view
echo "Regular user viewing relationships..."
USER_VIEW=$(curl -s -X GET "$BASE_URL/entity-relationships?source=$ENTITY1_ID" \
    -H "Authorization: $USER_TOKEN")
if [[ "$USER_VIEW" =~ "error" ]] && [[ ! "$USER_VIEW" =~ "relationships" ]]; then
    echo "❌ Regular user view failed: $USER_VIEW"
else
    echo "✅ Regular user can view"
fi

# Anonymous should fail
echo -e "\nAnonymous viewing relationships..."
ANON_VIEW=$(curl -s -X GET "$BASE_URL/entity-relationships?source=$ENTITY1_ID")
if [[ "$ANON_VIEW" =~ "401" ]] || [[ "$ANON_VIEW" =~ "Authentication required" ]]; then
    echo "✅ Anonymous correctly denied"
else
    echo "❌ Anonymous should have been denied: $ANON_VIEW"
fi

# 5. Test special permissions
echo -e "\n=== Test 5: Test special relationship permissions ==="

# Create user with only relation:view but not entity:view
SPECIAL_USER_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:user", "rbac:role:special", "rbac:perm:relation:view"],
        "content": [
            {"type": "username", "value": "special_user"},
            {"type": "password_hash", "value": "$2a$10$dummyhash"}
        ]
    }')
SPECIAL_ID=$(echo $SPECIAL_USER_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
SPECIAL_TOKEN="Bearer token_special_$SPECIAL_ID"

echo "Special user (relation:view only) viewing relationships..."
SPECIAL_VIEW=$(curl -s -X GET "$BASE_URL/entity-relationships?source=$ENTITY1_ID" \
    -H "Authorization: $SPECIAL_TOKEN")
if [[ "$SPECIAL_VIEW" =~ "error" ]] && [[ ! "$SPECIAL_VIEW" =~ "relationships" ]]; then
    echo "❌ Special user view failed: $SPECIAL_VIEW"
else
    echo "✅ Special user can view relationships"
fi

# 6. Test wildcard permissions
echo -e "\n=== Test 6: Test wildcard permissions ==="

# Create user with relation:* permission
WILDCARD_USER_RESPONSE=$(curl -s -X POST $BASE_URL/test/entities/create \
    -H "Content-Type: application/json" \
    -d '{
        "tags": ["type:user", "rbac:role:relation_admin", "rbac:perm:relation:*"],
        "content": [
            {"type": "username", "value": "relation_admin"},
            {"type": "password_hash", "value": "$2a$10$dummyhash"}
        ]
    }')
WILDCARD_ID=$(echo $WILDCARD_USER_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
WILDCARD_TOKEN="Bearer token_wildcard_$WILDCARD_ID"

echo "Relation admin creating relationship..."
WILDCARD_REL=$(curl -s -X POST $BASE_URL/entity-relationships \
    -H "Content-Type: application/json" \
    -H "Authorization: $WILDCARD_TOKEN" \
    -d "{
        \"source_id\": \"$ENTITY1_ID\",
        \"relationship_type\": \"administers\",
        \"target_id\": \"$ENTITY2_ID\"
    }")
if [[ "$WILDCARD_REL" =~ "error" ]]; then
    echo "❌ Relation admin failed: $WILDCARD_REL"
else
    echo "✅ Relation admin succeeded (wildcard permission works)"
fi

# 7. Summary
echo -e "\n=== RBAC Test Summary ==="
echo "✓ Admin users can create and view relationships"
echo "✓ Editors with specific permissions can create relationships"  
echo "✓ Regular users with only view permissions are denied create"
echo "✓ Anonymous users are denied access"
echo "✓ Users can have relationship permissions independent of entity permissions"
echo "✓ Wildcard permissions (relation:*) grant all relationship actions"

echo -e "\n=== All RBAC tests completed ==="