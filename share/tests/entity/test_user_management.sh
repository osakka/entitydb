#!/bin/bash

# EntityDB User Management API Test Script
# Tests comprehensive CRUD operations for user management 

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SERVER_URL="http://localhost:8085"

# Test counter
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Tokens for different users
ADMIN_TOKEN=""
TEST_USER_TOKEN=""
TEST_READONLY_TOKEN=""
TEST_USER_ID=""

# ------------------------------------------------------------------
# Utility functions
# ------------------------------------------------------------------

log() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
  PASSED_TESTS=$((PASSED_TESTS+1))
  TOTAL_TESTS=$((TOTAL_TESTS+1))
}

warning() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
  echo -e "${RED}[ERROR]${NC} $1"
  FAILED_TESTS=$((FAILED_TESTS+1))
  TOTAL_TESTS=$((TOTAL_TESTS+1))
}

run_test() {
  TEST_NAME=$1
  TEST_FUNC=$2
  
  echo
  echo -e "${BLUE}======================================================${NC}"
  echo -e "${BLUE}Running test: ${TEST_NAME}${NC}"
  echo -e "${BLUE}======================================================${NC}"
  
  # Run the test function
  $TEST_FUNC
}

# ------------------------------------------------------------------
# Authenticate admin and get token
# ------------------------------------------------------------------

get_admin_token() {
  log "Getting admin token..."
  
  # Get admin token
  RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password"}' \
    "${SERVER_URL}/api/v1/auth/login")
  
  ADMIN_TOKEN=$(echo $RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
  
  if [ -z "$ADMIN_TOKEN" ]; then
    error "Failed to get admin token"
    echo "Response: $RESPONSE"
    exit 1
  else
    success "Successfully obtained admin token"
  fi
}

# ------------------------------------------------------------------
# User Management Tests
# ------------------------------------------------------------------

test_user_creation() {
  log "Testing user creation..."
  
  # Test 1: Create a regular user
  TIMESTAMP=$(date +%s)
  RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"username\":\"test_user_${TIMESTAMP}\",\"password\":\"password123\",\"roles\":[\"user\"]}" \
    "${SERVER_URL}/api/v1/users")
  
  if [[ "$RESPONSE" == *"created successfully"* ]]; then
    success "Admin can create regular users"
    TEST_USER_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    log "Created test user with ID: $TEST_USER_ID"
  else
    error "Failed to create regular user"
    echo "Response: $RESPONSE"
  fi
  
  # Test 2: Create a read-only user
  RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"username\":\"readonly_user_${TIMESTAMP}\",\"password\":\"password123\",\"roles\":[\"readonly\"]}" \
    "${SERVER_URL}/api/v1/users")
  
  if [[ "$RESPONSE" == *"created successfully"* ]]; then
    success "Admin can create read-only users"
    READONLY_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    log "Created read-only user with ID: $READONLY_ID"
  else
    error "Failed to create read-only user"
    echo "Response: $RESPONSE"
  fi
  
  # Test 3: Try to create user with invalid username
  RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"username\":\"test-user-${TIMESTAMP}\",\"password\":\"password123\",\"roles\":[\"user\"]}" \
    "${SERVER_URL}/api/v1/users")
  
  if [[ "$RESPONSE" == *"can only contain letters, numbers, and underscores"* ]]; then
    success "Properly rejects invalid username format"
  else
    error "Failed to validate username format"
    echo "Response: $RESPONSE"
  fi
  
  # Test 4: Try to create user with short password
  RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"username\":\"test_user2_${TIMESTAMP}\",\"password\":\"123\",\"roles\":[\"user\"]}" \
    "${SERVER_URL}/api/v1/users")
  
  if [[ "$RESPONSE" == *"at least 8 characters"* ]]; then
    success "Properly rejects short passwords"
  else
    error "Failed to validate password strength"
    echo "Response: $RESPONSE"
  fi
  
  # Test 5: Try to create user with invalid role
  RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"username\":\"test_user3_${TIMESTAMP}\",\"password\":\"password123\",\"roles\":[\"superadmin\"]}" \
    "${SERVER_URL}/api/v1/users")
  
  if [[ "$RESPONSE" == *"Invalid role"* ]]; then
    success "Properly rejects invalid roles"
  else
    error "Failed to validate roles"
    echo "Response: $RESPONSE"
  fi
  
  # Get tokens for test users
  log "Getting tokens for test users..."
  
  RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"username\":\"test_user_${TIMESTAMP}\",\"password\":\"password123\"}" \
    "${SERVER_URL}/api/v1/auth/login")
  
  TEST_USER_TOKEN=$(echo $RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
  
  if [ -z "$TEST_USER_TOKEN" ]; then
    warning "Could not get token for test user, some tests will be skipped"
  else
    log "Successfully obtained token for test user"
  fi
  
  RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"username\":\"readonly_user_${TIMESTAMP}\",\"password\":\"password123\"}" \
    "${SERVER_URL}/api/v1/auth/login")
  
  TEST_READONLY_TOKEN=$(echo $RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
  
  if [ -z "$TEST_READONLY_TOKEN" ]; then
    warning "Could not get token for read-only user, some tests will be skipped"
  else
    log "Successfully obtained token for read-only user"
  fi
}

test_user_retrieval() {
  log "Testing user retrieval..."
  
  # Skip if we don't have test user ID
  if [ -z "$TEST_USER_ID" ]; then
    warning "Test user ID not available, skipping user retrieval tests"
    return
  fi
  
  # Test 1: Admin can list all users
  RESPONSE=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
    "${SERVER_URL}/api/v1/users")
  
  if [[ "$RESPONSE" == *"\"status\":\"ok\""* ]]; then
    success "Admin can list all users"
  else
    error "Admin cannot list users"
    echo "Response: $RESPONSE"
  fi
  
  # Test 2: Admin can get specific user by ID
  RESPONSE=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
    "${SERVER_URL}/api/v1/users/$TEST_USER_ID")
  
  if [[ "$RESPONSE" == *"\"status\":\"ok\""* ]]; then
    success "Admin can retrieve specific user by ID"
  else
    error "Admin cannot retrieve specific user"
    echo "Response: $RESPONSE"
  fi
  
  # Test 3: Regular user can get their own details
  if [ -n "$TEST_USER_TOKEN" ]; then
    RESPONSE=$(curl -s -H "Authorization: Bearer $TEST_USER_TOKEN" \
      "${SERVER_URL}/api/v1/users/$TEST_USER_ID")
    
    if [[ "$RESPONSE" == *"\"status\":\"ok\""* ]]; then
      success "User can retrieve their own details"
    else
      error "User cannot retrieve their own details"
      echo "Response: $RESPONSE"
    fi
  fi
  
  # Test 4: Regular user cannot list all users
  if [ -n "$TEST_USER_TOKEN" ]; then
    RESPONSE=$(curl -s -H "Authorization: Bearer $TEST_USER_TOKEN" \
      "${SERVER_URL}/api/v1/users")
    
    if [[ "$RESPONSE" == *"Admin privileges required"* ]]; then
      success "Regular users cannot list all users"
    else
      error "Regular user was able to list all users"
      echo "Response: $RESPONSE"
    fi
  fi
}

test_user_update() {
  log "Testing user updates..."
  
  # Skip if we don't have test user ID or token
  if [ -z "$TEST_USER_ID" ] || [ -z "$TEST_USER_TOKEN" ]; then
    warning "Test user ID or token not available, skipping user update tests"
    return
  fi
  
  # Test 1: Admin can update user roles
  RESPONSE=$(curl -s -X PUT -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"roles\":[\"user\",\"readonly\"]}" \
    "${SERVER_URL}/api/v1/users/$TEST_USER_ID")
  
  if [[ "$RESPONSE" == *"updated successfully"* ]]; then
    success "Admin can update user roles"
  else
    error "Admin failed to update user roles"
    echo "Response: $RESPONSE"
  fi
  
  # Test 2: User can update their own password
  RESPONSE=$(curl -s -X PUT -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TEST_USER_TOKEN" \
    -d "{\"password\":\"newpassword123\"}" \
    "${SERVER_URL}/api/v1/users/$TEST_USER_ID")
  
  if [[ "$RESPONSE" == *"updated successfully"* ]]; then
    success "User can update their own password"
  else
    error "User failed to update their own password"
    echo "Response: $RESPONSE"
  fi
  
  # Test 3: User cannot update their roles
  RESPONSE=$(curl -s -X PUT -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TEST_USER_TOKEN" \
    -d "{\"roles\":[\"admin\"]}" \
    "${SERVER_URL}/api/v1/users/$TEST_USER_ID")
  
  if [[ "$RESPONSE" == *"Only admins can update user roles"* ]]; then
    success "Users cannot update their own roles"
  else
    error "User was able to update their roles"
    echo "Response: $RESPONSE"
  fi
  
  # Test 4: Reject invalid password update
  RESPONSE=$(curl -s -X PUT -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TEST_USER_TOKEN" \
    -d "{\"password\":\"pass\"}" \
    "${SERVER_URL}/api/v1/users/$TEST_USER_ID")
  
  if [[ "$RESPONSE" == *"at least 8 characters"* ]]; then
    success "Properly rejects short passwords in update"
  else
    error "Failed to validate password strength in update"
    echo "Response: $RESPONSE"
  fi
}

test_user_deletion() {
  log "Testing user deletion..."
  
  # Skip if we don't have test user ID
  if [ -z "$TEST_USER_ID" ]; then
    warning "Test user ID not available, skipping user deletion tests"
    return
  fi
  
  # Create a temporary test user for deletion
  TIMESTAMP=$(date +%s)
  RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -d "{\"username\":\"delete_user_${TIMESTAMP}\",\"password\":\"password123\",\"roles\":[\"user\"]}" \
    "${SERVER_URL}/api/v1/users")
  
  DELETE_USER_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  
  if [ -z "$DELETE_USER_ID" ]; then
    warning "Could not create temporary user for deletion test, skipping"
    return
  else
    log "Created temporary user for deletion with ID: $DELETE_USER_ID"
  fi
  
  # Test 1: Admin can delete users
  RESPONSE=$(curl -s -X DELETE -H "Authorization: Bearer $ADMIN_TOKEN" \
    "${SERVER_URL}/api/v1/users/$DELETE_USER_ID")
  
  if [[ "$RESPONSE" == *"deleted successfully"* ]]; then
    success "Admin can delete users"
  else
    error "Admin failed to delete user"
    echo "Response: $RESPONSE"
  fi
  
  # Test 2: Cannot delete non-existent user
  RESPONSE=$(curl -s -X DELETE -H "Authorization: Bearer $ADMIN_TOKEN" \
    "${SERVER_URL}/api/v1/users/$DELETE_USER_ID")
  
  if [[ "$RESPONSE" == *"not found"* ]]; then
    success "Properly handles deletion of non-existent users"
  else
    error "Improper handling of non-existent user deletion"
    echo "Response: $RESPONSE"
  fi
  
  # Test 3: User cannot delete other users
  if [ -n "$TEST_USER_TOKEN" ]; then
    # Create another test user
    TIMESTAMP=$(date +%s)
    RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -d "{\"username\":\"other_user_${TIMESTAMP}\",\"password\":\"password123\",\"roles\":[\"user\"]}" \
      "${SERVER_URL}/api/v1/users")
    
    OTHER_USER_ID=$(echo $RESPONSE | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    
    if [ -n "$OTHER_USER_ID" ]; then
      RESPONSE=$(curl -s -X DELETE -H "Authorization: Bearer $TEST_USER_TOKEN" \
        "${SERVER_URL}/api/v1/users/$OTHER_USER_ID")
      
      if [[ "$RESPONSE" == *"Admin privileges required"* ]]; then
        success "Regular users cannot delete other users"
      else
        error "Regular user was able to delete another user"
        echo "Response: $RESPONSE"
      fi
      
      # Clean up: Admin deletes the other test user
      curl -s -X DELETE -H "Authorization: Bearer $ADMIN_TOKEN" \
        "${SERVER_URL}/api/v1/users/$OTHER_USER_ID" > /dev/null
    else
      warning "Could not create second test user, skipping related test"
    fi
  fi
  
  # Test 4: Admin cannot delete their own account
  RESPONSE=$(curl -s -X DELETE -H "Authorization: Bearer $ADMIN_TOKEN" \
    "${SERVER_URL}/api/v1/users/usr_admin")
  
  if [[ "$RESPONSE" == *"cannot delete your own user account"* ]]; then
    success "Admin cannot delete their own account"
  else
    error "Improper handling of self-deletion"
    echo "Response: $RESPONSE"
  fi
  
  # Clean up test users
  log "Cleaning up test users..."
  
  if [ -n "$TEST_USER_ID" ]; then
    curl -s -X DELETE -H "Authorization: Bearer $ADMIN_TOKEN" \
      "${SERVER_URL}/api/v1/users/$TEST_USER_ID" > /dev/null
    log "Deleted test user"
  fi
}

# ------------------------------------------------------------------
# Main execution
# ------------------------------------------------------------------

main() {
  echo -e "${BLUE}=========================================================${NC}"
  echo -e "${BLUE}       EntityDB User Management API Test                     ${NC}"
  echo -e "${BLUE}=========================================================${NC}"
  echo
  
  # Check if server is running
  if ! curl -s "${SERVER_URL}/api/v1/status" > /dev/null; then
    error "EntityDB server is not running at ${SERVER_URL}"
    echo "Please start the server with: /opt/entitydb/bin/entitydbd.sh start"
    exit 1
  fi
  
  log "EntityDB server is running at ${SERVER_URL}"
  
  # Run tests
  run_test "Get Admin Token" get_admin_token
  run_test "User Creation" test_user_creation
  run_test "User Retrieval" test_user_retrieval
  run_test "User Update" test_user_update
  run_test "User Deletion" test_user_deletion
  
  # Print summary
  echo
  echo -e "${BLUE}=========================================================${NC}"
  echo -e "${BLUE}                     Test Summary                        ${NC}"
  echo -e "${BLUE}=========================================================${NC}"
  echo
  echo -e "Total tests run: ${TOTAL_TESTS}"
  echo -e "Passed: ${GREEN}${PASSED_TESTS}${NC}"
  echo -e "Failed: ${RED}${FAILED_TESTS}${NC}"
  echo
  
  if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
  else
    echo -e "${RED}Some tests failed. Please check the logs above.${NC}"
    exit 1
  fi
}

# Run main function
main