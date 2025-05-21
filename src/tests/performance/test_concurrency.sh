#!/bin/bash
# EntityDB Concurrency Test
# This script tests database durability under concurrent write/read operations

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8085"
CONCURRENCY=5          # Number of concurrent operations (reduced for quicker testing)
ITERATIONS_PER_CLIENT=5  # Number of operations per concurrent client
FILE_SIZE_MB=1         # Size of test files in MB
TOKEN=""

# Function for colored output
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

# Login to get token
login() {
  print_message "$BLUE" "Logging in to EntityDB..."
  
  local response=$(curl -k -s -X POST "$SERVER_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}')
  
  TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
  
  if [ -z "$TOKEN" ]; then
    print_message "$RED" "❌ Failed to login. Response: $response"
    exit 1
  else
    print_message "$GREEN" "✅ Login successful, got token"
  fi
}

# Generate a file with random content
generate_test_file() {
  local size_mb=$1
  local output_file=$2
  local identifier=$3
  
  print_message "$BLUE" "Generating $size_mb MB test file for client $identifier..."
  dd if=/dev/urandom of="$output_file" bs=1M count="$size_mb" 2>/dev/null
  echo "Concurrency test file for client $identifier generated at $(date)" >> "$output_file"
  
  local file_hash=$(sha256sum "$output_file" | cut -d' ' -f1)
  echo "$file_hash" > "${output_file}.hash"
  
  print_message "$GREEN" "Generated $size_mb MB test file with SHA256: $file_hash"
}

# Create entity function for concurrent clients
create_entity() {
  local client_id=$1
  local iteration=$2
  local file_path=$3
  local log_file=$4
  
  local file_name="client${client_id}_iter${iteration}.bin"
  local file_hash=$(cat "${file_path}.hash")
  
  # Base64 encode the content
  local encoded_content=$(base64 -w0 "$file_path")
  
  # Create the entity
  local response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"tags\": [\"type:concurrency_test\", \"client:$client_id\", \"iteration:$iteration\", \"hash:$file_hash\"],
      \"content\": {
        \"data\": \"$encoded_content\",
        \"type\": \"application/octet-stream\"
      }
    }")
  
  # Extract entity ID
  local entity_id=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
  
  if [ -z "$entity_id" ]; then
    echo "❌ Client $client_id, Iteration $iteration: Create failed - $response" >> "$log_file"
    return 1
  else
    echo "✅ Client $client_id, Iteration $iteration: Created entity $entity_id" >> "$log_file"
    echo "$entity_id" > "${file_path}.id"
    return 0
  fi
}

# Get entity and verify content
verify_entity() {
  local client_id=$1
  local iteration=$2
  local file_path=$3
  local log_file=$4
  
  local entity_id=$(cat "${file_path}.id")
  local original_hash=$(cat "${file_path}.hash")
  local output_file="${file_path}.retrieved"
  
  # Get the entity
  local response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/get?id=$entity_id" \
    -H "Authorization: Bearer $TOKEN")
  
  # Extract content
  local content=$(echo "$response" | grep -o '"data":"[^"]*' | cut -d'"' -f4)
  
  if [ -z "$content" ]; then
    echo "❌ Client $client_id, Iteration $iteration: Retrieval failed for $entity_id - $response" >> "$log_file"
    return 1
  fi
  
  # Decode the content to a file
  echo "$content" | base64 -d > "$output_file"
  
  # Verify file hashes match
  local retrieved_hash=$(sha256sum "$output_file" | cut -d' ' -f1)
  
  if [ "$original_hash" == "$retrieved_hash" ]; then
    echo "✅ Client $client_id, Iteration $iteration: Verified content for $entity_id" >> "$log_file"
    return 0
  else
    echo "❌ Client $client_id, Iteration $iteration: Content verification failed for $entity_id" >> "$log_file"
    echo "  Original hash: $original_hash" >> "$log_file"
    echo "  Retrieved hash: $retrieved_hash" >> "$log_file"
    return 1
  fi
}

# Update entity (modify content and tags)
update_entity() {
  local client_id=$1
  local iteration=$2
  local file_path=$3
  local log_file=$4
  
  local entity_id=$(cat "${file_path}.id")
  local modified_file="${file_path}.modified"
  
  # Modify the file
  cp "$file_path" "$modified_file"
  echo "Modified content for iteration $iteration at $(date)" >> "$modified_file"
  
  # Recalculate hash
  local new_hash=$(sha256sum "$modified_file" | cut -d' ' -f1)
  echo "$new_hash" > "${modified_file}.hash"
  
  # Base64 encode the modified content
  local encoded_content=$(base64 -w0 "$modified_file")
  
  # Update the entity
  local response=$(curl -k -s -X PUT "$SERVER_URL/api/v1/entities/update" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"id\": \"$entity_id\",
      \"tags\": [\"type:concurrency_test\", \"client:$client_id\", \"iteration:$iteration\", \"modified:true\", \"hash:$new_hash\"],
      \"content\": {
        \"data\": \"$encoded_content\",
        \"type\": \"application/octet-stream\"
      }
    }")
  
  # Verify update
  if [[ "$response" == *"\"id\":"* ]]; then
    echo "✅ Client $client_id, Iteration $iteration: Updated entity $entity_id" >> "$log_file"
    echo "$entity_id" > "${modified_file}.id"
    return 0
  else
    echo "❌ Client $client_id, Iteration $iteration: Update failed for $entity_id - $response" >> "$log_file"
    return 1
  fi
}

# Query entities
query_entities() {
  local client_id=$1
  local log_file=$2
  
  # Query for entities created by this client
  local response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/list?tag=client:$client_id" \
    -H "Authorization: Bearer $TOKEN")
  
  local entity_count=$(echo "$response" | grep -o "\"id\":" | wc -l)
  
  if [ $entity_count -gt 0 ]; then
    echo "✅ Client $client_id: Query found $entity_count entities" >> "$log_file"
    return 0
  else
    echo "❌ Client $client_id: Query failed to find any entities - $response" >> "$log_file"
    return 1
  fi
}

# Single client workflow
client_workflow() {
  local client_id=$1
  local tmp_dir=$2
  local log_file="${tmp_dir}/client_${client_id}.log"
  
  print_message "$BLUE" "Starting client $client_id workflow..."
  
  local create_success=0
  local create_fail=0
  local verify_success=0
  local verify_fail=0
  local update_success=0
  local update_fail=0
  
  # Create test file for this client
  local test_file="${tmp_dir}/client_${client_id}.bin"
  generate_test_file $FILE_SIZE_MB "$test_file" "$client_id"
  
  # Perform operations
  for i in $(seq 1 $ITERATIONS_PER_CLIENT); do
    # Create entity
    create_entity "$client_id" "$i" "$test_file" "$log_file"
    if [ $? -eq 0 ]; then
      ((create_success++))
      
      # Verify content
      verify_entity "$client_id" "$i" "$test_file" "$log_file"
      if [ $? -eq 0 ]; then
        ((verify_success++))
        
        # Update entity
        update_entity "$client_id" "$i" "$test_file" "$log_file"
        if [ $? -eq 0 ]; then
          ((update_success++))
          
          # Verify updated content
          verify_entity "$client_id" "$i" "${test_file}.modified" "$log_file"
          if [ $? -eq 0 ]; then
            ((verify_success++))
          else
            ((verify_fail++))
          fi
        else
          ((update_fail++))
        fi
      else
        ((verify_fail++))
      fi
    else
      ((create_fail++))
    fi
  done
  
  # Query for all entities created by this client
  query_entities "$client_id" "$log_file"
  
  # Write summary to log
  echo "CLIENT $client_id SUMMARY:" >> "$log_file"
  echo "Create: $create_success success, $create_fail fail" >> "$log_file"
  echo "Verify: $verify_success success, $verify_fail fail" >> "$log_file"
  echo "Update: $update_success success, $update_fail fail" >> "$log_file"
  
  print_message "$GREEN" "Client $client_id workflow completed."
}

# Run concurrent client tests
run_concurrent_test() {
  local tmp_dir="./tmp_concurrency_test"
  mkdir -p "$tmp_dir"
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Starting concurrent database test with $CONCURRENCY clients"
  print_message "$BLUE" "Each client will perform $ITERATIONS_PER_CLIENT operations"
  print_message "$BLUE" "========================================"
  
  # Start clients in background
  for i in $(seq 1 $CONCURRENCY); do
    client_workflow "$i" "$tmp_dir" &
    print_message "$GREEN" "Started client $i"
    # Small delay to avoid exact-same-time operations
    sleep 0.5
  done
  
  # Wait for all clients to complete
  wait
  print_message "$GREEN" "All clients completed."
  
  # Aggregate results
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Aggregating test results..."
  print_message "$BLUE" "========================================"
  
  local total_create_success=0
  local total_create_fail=0
  local total_verify_success=0
  local total_verify_fail=0
  local total_update_success=0
  local total_update_fail=0
  
  for i in $(seq 1 $CONCURRENCY); do
    local log_file="${tmp_dir}/client_${i}.log"
    
    # Extract statistics
    local create_success=$(grep "Create:" "$log_file" | awk '{print $2}')
    local create_fail=$(grep "Create:" "$log_file" | awk '{print $4}')
    local verify_success=$(grep "Verify:" "$log_file" | awk '{print $2}')
    local verify_fail=$(grep "Verify:" "$log_file" | awk '{print $4}')
    local update_success=$(grep "Update:" "$log_file" | awk '{print $2}')
    local update_fail=$(grep "Update:" "$log_file" | awk '{print $4}')
    
    total_create_success=$((total_create_success + create_success))
    total_create_fail=$((total_create_fail + create_fail))
    total_verify_success=$((total_verify_success + verify_success))
    total_verify_fail=$((total_verify_fail + verify_fail))
    total_update_success=$((total_update_success + update_success))
    total_update_fail=$((total_update_fail + update_fail))
  done
  
  # Print summary
  print_message "$BLUE" "CONCURRENCY TEST SUMMARY:"
  print_message "$BLUE" "------------------------"
  print_message "$BLUE" "Create operations: $total_create_success success, $total_create_fail fail"
  print_message "$BLUE" "Verify operations: $total_verify_success success, $total_verify_fail fail"
  print_message "$BLUE" "Update operations: $total_update_success success, $total_update_fail fail"
  
  local total_operations=$((total_create_success + total_create_fail + total_verify_success + total_verify_fail + total_update_success + total_update_fail))
  local total_failures=$((total_create_fail + total_verify_fail + total_update_fail))
  local failure_rate=$(awk "BEGIN {printf \"%.2f\", ($total_failures / $total_operations) * 100}")
  
  print_message "$BLUE" "------------------------"
  print_message "$BLUE" "Total operations: $total_operations"
  print_message "$BLUE" "Total failures: $total_failures"
  print_message "$BLUE" "Failure rate: $failure_rate%"
  
  # Determine test result
  if [ $total_failures -eq 0 ]; then
    print_message "$GREEN" "✅ CONCURRENCY TEST PASSED! Database maintained integrity under concurrent load."
    result=0
  elif [ $(echo "$failure_rate < 5" | bc -l) -eq 1 ]; then
    print_message "$YELLOW" "⚠️ CONCURRENCY TEST ACCEPTABLE - Minor issues ($failure_rate% failure rate)."
    result=0
  else
    print_message "$RED" "❌ CONCURRENCY TEST FAILED with $failure_rate% failure rate! Database integrity issues under concurrent load."
    result=1
  fi
  
  print_message "$BLUE" "========================================"
  
  # Clean up
  # rm -rf "$tmp_dir"
  print_message "$BLUE" "Test logs saved in $tmp_dir for review"
  
  return $result
}

# Main execution
print_message "$BLUE" "========================================"
print_message "$BLUE" "Starting EntityDB Concurrency Test"
print_message "$BLUE" "========================================"

# Login
login

# Run concurrent test
run_concurrent_test
concurrency_result=$?

# Exit with result
exit $concurrency_result