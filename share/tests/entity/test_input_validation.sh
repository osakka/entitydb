#!/bin/bash
#
# Test script for input validation functionality
#

# Set environment
HOST="localhost:8085"
ADMIN_TOKEN=""
TEST_PREFIX="validation_test_$(date +%s)"

echo "Testing Input Validation Functionality"
echo "======================================"

# Get admin token
echo "Getting admin token..."
ADMIN_TOKEN=$(curl -s -X POST -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password"}' \
    "http://${HOST}/api/v1/auth/login" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ADMIN_TOKEN" ]; then
    echo "Failed to get admin token, aborting test"
    exit 1
fi

echo "Admin token obtained"

# Test 1: Valid entity creation
echo -e "\nTest 1: Valid entity creation"
VALID_ENTITY=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"type\":\"issue\",\"title\":\"${TEST_PREFIX}_valid\",\"description\":\"Valid test entity\",\"status\":\"active\",\"tags\":[\"test\",\"validation\"],\"properties\":{\"priority\":\"high\"}}" \
    "http://${HOST}/api/v1/entities")

if [[ "$VALID_ENTITY" == *"created successfully"* ]]; then
    echo "PASS: Valid entity created successfully"
    VALID_ENTITY_ID=$(echo "$VALID_ENTITY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "Created entity ID: $VALID_ENTITY_ID"
else
    echo "FAIL: Valid entity creation failed"
    echo "Response: $VALID_ENTITY"
fi

# Test 2: Invalid entity creation (missing required fields)
echo -e "\nTest 2: Invalid entity creation (missing required fields)"
INVALID_ENTITY_1=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"description\":\"Invalid test entity\"}" \
    "http://${HOST}/api/v1/entities")

if [[ "$INVALID_ENTITY_1" == *"required fields"* ]]; then
    echo "PASS: Validation correctly rejected entity without required fields"
else
    echo "FAIL: Validation did not reject entity without required fields"
    echo "Response: $INVALID_ENTITY_1"
fi

# Test 3: Invalid entity creation (invalid field values)
echo -e "\nTest 3: Invalid entity creation (invalid field values)"
INVALID_ENTITY_2=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"type\":\"issue\",\"title\":\"${TEST_PREFIX}_invalid\",\"description\":\"Invalid test entity\",\"status\":\"active\",\"tags\":[\"invalid tag with spaces\"],\"properties\":{\"priority\":\"high\"}}" \
    "http://${HOST}/api/v1/entities")

if [[ "$INVALID_ENTITY_2" == *"Validation failed"* ]] && [[ "$INVALID_ENTITY_2" == *"pattern"* ]]; then
    echo "PASS: Validation correctly rejected entity with invalid tag format"
else
    echo "FAIL: Validation did not reject entity with invalid tag format"
    echo "Response: $INVALID_ENTITY_2"
fi

# Test 4: Valid relationship creation
echo -e "\nTest 4: Valid relationship creation"
if [ ! -z "$VALID_ENTITY_ID" ]; then
    VALID_REL=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "{\"source_id\":\"$VALID_ENTITY_ID\",\"target_id\":\"$VALID_ENTITY_ID\",\"type\":\"self_reference\",\"properties\":{\"created_date\":\"$(date -Iseconds)\"}}" \
        "http://${HOST}/api/v1/entity-relationships")

    if [[ "$VALID_REL" == *"created successfully"* ]]; then
        echo "PASS: Valid relationship created successfully"
        VALID_REL_ID=$(echo "$VALID_REL" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        echo "Created relationship ID: $VALID_REL_ID"
    else
        echo "FAIL: Valid relationship creation failed"
        echo "Response: $VALID_REL"
    fi
else
    echo "SKIP: No valid entity created, skipping relationship test"
fi

# Test 5: Invalid relationship creation (missing required fields)
echo -e "\nTest 5: Invalid relationship creation (missing required fields)"
INVALID_REL_1=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"source_id\":\"entity_001\",\"properties\":{\"test\":\"value\"}}" \
    "http://${HOST}/api/v1/entity-relationships")

if [[ "$INVALID_REL_1" == *"Validation failed"* ]] && [[ "$INVALID_REL_1" == *"required"* ]]; then
    echo "PASS: Validation correctly rejected relationship without required fields"
else
    echo "FAIL: Validation did not reject relationship without required fields"
    echo "Response: $INVALID_REL_1"
fi

# Test 6: Invalid relationship creation (invalid field values)
echo -e "\nTest 6: Invalid relationship creation (invalid field values)"
INVALID_REL_2=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"source_id\":\"invalid_id\",\"target_id\":\"invalid_id\",\"type\":\"test\",\"properties\":{}}" \
    "http://${HOST}/api/v1/entity-relationships")

if [[ "$INVALID_REL_2" == *"Validation failed"* ]] && [[ "$INVALID_REL_2" == *"pattern"* ]]; then
    echo "PASS: Validation correctly rejected relationship with invalid ID format"
else
    echo "FAIL: Validation did not reject relationship with invalid ID format"
    echo "Response: $INVALID_REL_2"
fi

# Test 7: Invalid login (missing fields)
echo -e "\nTest 7: Invalid login (missing fields)"
INVALID_LOGIN_1=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"username\":\"admin\"}" \
    "http://${HOST}/api/v1/auth/login")

if [[ "$INVALID_LOGIN_1" == *"Validation failed"* ]] && [[ "$INVALID_LOGIN_1" == *"required"* ]]; then
    echo "PASS: Validation correctly rejected login without password"
else
    echo "FAIL: Validation did not reject login without password"
    echo "Response: $INVALID_LOGIN_1"
fi

# Test 8: Invalid login (invalid username format)
echo -e "\nTest 8: Invalid login (invalid username format)"
INVALID_LOGIN_2=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"username\":\"admin@example.com\",\"password\":\"password\"}" \
    "http://${HOST}/api/v1/auth/login")

if [[ "$INVALID_LOGIN_2" == *"Validation failed"* ]] && [[ "$INVALID_LOGIN_2" == *"pattern"* ]]; then
    echo "PASS: Validation correctly rejected login with invalid username format"
else
    echo "FAIL: Validation did not reject login with invalid username format"
    echo "Response: $INVALID_LOGIN_2"
fi

# Test 9: Invalid user creation (invalid roles)
echo -e "\nTest 9: Invalid user creation (invalid roles)"
INVALID_USER=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"type\":\"user\",\"title\":\"${TEST_PREFIX}_user\",\"description\":\"Test user\",\"status\":\"active\",\"tags\":[\"user\"],\"properties\":{\"username\":\"${TEST_PREFIX}_user\",\"roles\":[\"invalid_role\"],\"password_hash\":\"securepassword\"}}" \
    "http://${HOST}/api/v1/entities")

if [[ "$INVALID_USER" == *"Validation failed"* ]] && [[ "$INVALID_USER" == *"role"* ]]; then
    echo "PASS: Validation correctly rejected user with invalid role"
else
    echo "FAIL: Validation did not reject user with invalid role"
    echo "Response: $INVALID_USER"
fi

# Test 10: Valid user creation
echo -e "\nTest 10: Valid user creation"
VALID_USER=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"type\":\"user\",\"title\":\"${TEST_PREFIX}_valid_user\",\"description\":\"Valid test user\",\"status\":\"active\",\"tags\":[\"user\"],\"properties\":{\"username\":\"${TEST_PREFIX}_valid_user\",\"roles\":[\"user\"],\"password_hash\":\"securepassword\"}}" \
    "http://${HOST}/api/v1/entities")

if [[ "$VALID_USER" == *"created successfully"* ]]; then
    echo "PASS: Valid user created successfully"
    VALID_USER_ID=$(echo "$VALID_USER" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "Created user ID: $VALID_USER_ID"
else
    echo "FAIL: Valid user creation failed"
    echo "Response: $VALID_USER"
fi

# Clean up test entities
echo -e "\nCleaning up..."
if [ ! -z "$VALID_ENTITY_ID" ]; then
    curl -s -X DELETE -H "Authorization: Bearer $ADMIN_TOKEN" \
        "http://${HOST}/api/v1/entities/$VALID_ENTITY_ID" > /dev/null
    echo "Deleted test entity"
fi

if [ ! -z "$VALID_USER_ID" ]; then
    curl -s -X DELETE -H "Authorization: Bearer $ADMIN_TOKEN" \
        "http://${HOST}/api/v1/entities/$VALID_USER_ID" > /dev/null
    echo "Deleted test user"
fi

echo -e "\nInput Validation Test Summary:"
echo "======================================"
echo "- Verified required field validation"
echo "- Verified field format validation"
echo "- Verified pattern validation for entity fields"
echo "- Verified pattern validation for relationship fields"
echo "- Verified pattern validation for authentication fields"
echo "- Verified user-specific validation rules"
echo "======================================"