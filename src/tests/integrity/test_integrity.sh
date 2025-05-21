#!/bin/bash
# EntityDB Integrity Test Suite
# This script tests database integrity with large files and many iterations

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="https://localhost:8085"
TEST_ITERATIONS=20
MAX_FILE_SIZE_MB=20   # Maximum file size to test (in MB)
CONCURRENCY=5         # Number of concurrent operations
START_SIZE_MB=1       # Starting size in MB
TOKEN=""

# Function for colored output
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

# Function to generate a file of specified size with random content
generate_test_file() {
  local size_mb=$1
  local output_file=$2
  
  print_message "$BLUE" "Generating $size_mb MB test file at $output_file..."
  dd if=/dev/urandom of="$output_file" bs=1M count="$size_mb" 2>/dev/null
  echo "Test file generated at $(date)" >> "$output_file"
  
  print_message "$GREEN" "Generated $size_mb MB test file with SHA256: $(sha256sum "$output_file" | cut -d' ' -f1)"
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

# Create entity with content
create_entity_with_content() {
  local file_path=$1
  local size_mb=$2
  local file_name=$(basename "$file_path")
  
  print_message "$BLUE" "Creating entity with $size_mb MB file: $file_name..."
  
  # Base64 encode the content (streaming to avoid memory issues)
  local encoded_content=$(base64 -w0 "$file_path")
  
  # Calculate original file hash for verification
  local original_hash=$(sha256sum "$file_path" | cut -d' ' -f1)
  
  # Create the entity
  local response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"tags\": [\"type:test_file\", \"name:$file_name\", \"size:${size_mb}MB\", \"hash:$original_hash\"],
      \"content\": {
        \"data\": \"$encoded_content\",
        \"type\": \"application/octet-stream\"
      }
    }")
  
  # Extract entity ID
  local entity_id=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
  
  if [ -z "$entity_id" ]; then
    print_message "$RED" "❌ Failed to create entity. Response: $response"
    return 1
  else
    print_message "$GREEN" "✅ Created entity with ID: $entity_id"
    echo "$entity_id"
    return 0
  fi
}

# Get entity and verify its content
verify_entity_content() {
  local entity_id=$1
  local original_file=$2
  local output_file="tmp_output_${entity_id}.bin"
  
  print_message "$BLUE" "Verifying entity content for ID: $entity_id..."
  
  # Get the entity
  local response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/get?id=$entity_id" \
    -H "Authorization: Bearer $TOKEN")
  
  # Extract content
  local content=$(echo "$response" | grep -o '"data":"[^"]*' | cut -d'"' -f4)
  
  if [ -z "$content" ]; then
    print_message "$RED" "❌ Failed to get entity content. Response: $response"
    return 1
  fi
  
  # Decode the content to a file
  echo "$content" | base64 -d > "$output_file"
  
  # Verify file hashes match
  local original_hash=$(sha256sum "$original_file" | cut -d'"' -f1)
  local retrieved_hash=$(sha256sum "$output_file" | cut -d'"' -f1)
  
  if [ "$original_hash" == "$retrieved_hash" ]; then
    print_message "$GREEN" "✅ Verified entity content integrity"
    rm "$output_file"
    return 0
  else
    print_message "$RED" "❌ Content verification failed!"
    print_message "$RED" "Original hash: $original_hash"
    print_message "$RED" "Retrieved hash: $retrieved_hash"
    return 1
  fi
}

# Test entity history
test_entity_history() {
  local entity_id=$1
  
  print_message "$BLUE" "Testing entity history for ID: $entity_id..."
  
  # Get entity history
  local response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/history?id=$entity_id" \
    -H "Authorization: Bearer $TOKEN")
  
  if [[ "$response" == *"\"history\":"* ]]; then
    print_message "$GREEN" "✅ Entity history retrieved successfully"
    return 0
  else
    print_message "$RED" "❌ Failed to get entity history. Response: $response"
    return 1
  fi
}

# Run tests with increasing file sizes
run_size_tests() {
  local tmp_dir="./tmp_test_files"
  mkdir -p "$tmp_dir"
  
  local success_count=0
  local fail_count=0
  
  for size_mb in $(seq $START_SIZE_MB $MAX_FILE_SIZE_MB); do
    print_message "$BLUE" "========================================"
    print_message "$BLUE" "Testing with file size: $size_mb MB"
    print_message "$BLUE" "========================================"
    
    local test_file="$tmp_dir/test_${size_mb}MB.bin"
    generate_test_file "$size_mb" "$test_file"
    
    local entity_id=$(create_entity_with_content "$test_file" "$size_mb")
    
    if [ $? -eq 0 ]; then
      sleep 1 # Allow time for entity to be fully processed
      
      # Verify content
      verify_entity_content "$entity_id" "$test_file"
      local verify_result=$?
      
      # Test history
      test_entity_history "$entity_id"
      local history_result=$?
      
      if [ $verify_result -eq 0 ] && [ $history_result -eq 0 ]; then
        print_message "$GREEN" "✅ All tests passed for $size_mb MB file"
        ((success_count++))
      else
        print_message "$RED" "❌ Tests failed for $size_mb MB file"
        ((fail_count++))
      fi
    else
      print_message "$RED" "❌ Entity creation failed for $size_mb MB file"
      ((fail_count++))
    fi
    
    echo ""
  done
  
  # Clean up
  rm -rf "$tmp_dir"
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Size Test Results:"
  print_message "$GREEN" "✅ Successful tests: $success_count"
  print_message "$RED" "❌ Failed tests: $fail_count"
  print_message "$BLUE" "========================================"
  
  return $fail_count
}

# Test multiple iterations of creation and verification
run_iteration_tests() {
  local size_mb=5 # Fixed size for iteration tests
  local tmp_dir="./tmp_iteration_files"
  mkdir -p "$tmp_dir"
  
  local success_count=0
  local fail_count=0
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Running $TEST_ITERATIONS iterations of $size_mb MB files"
  print_message "$BLUE" "========================================"
  
  for i in $(seq 1 $TEST_ITERATIONS); do
    print_message "$BLUE" "Iteration $i of $TEST_ITERATIONS"
    
    local test_file="$tmp_dir/iter_${i}_${size_mb}MB.bin"
    generate_test_file "$size_mb" "$test_file"
    
    local entity_id=$(create_entity_with_content "$test_file" "$size_mb")
    
    if [ $? -eq 0 ]; then
      sleep 1 # Allow time for entity to be fully processed
      
      # Verify content
      verify_entity_content "$entity_id" "$test_file"
      if [ $? -eq 0 ]; then
        print_message "$GREEN" "✅ Iteration $i passed"
        ((success_count++))
      else
        print_message "$RED" "❌ Iteration $i failed during verification"
        ((fail_count++))
      fi
    else
      print_message "$RED" "❌ Iteration $i failed during creation"
      ((fail_count++))
    fi
    
    echo ""
  done
  
  # Clean up
  rm -rf "$tmp_dir"
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Iteration Test Results:"
  print_message "$GREEN" "✅ Successful iterations: $success_count"
  print_message "$RED" "❌ Failed iterations: $fail_count"
  print_message "$BLUE" "========================================"
  
  return $fail_count
}

# Run concurrent operations test
run_concurrent_tests() {
  local size_mb=3 # Fixed size for concurrent tests
  local tmp_dir="./tmp_concurrent_files"
  mkdir -p "$tmp_dir"
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Running concurrent tests with $CONCURRENCY parallel operations"
  print_message "$BLUE" "========================================"
  
  # Array to store entity IDs and file paths
  declare -a entity_ids
  declare -a file_paths
  
  # Generate test files
  for i in $(seq 1 $CONCURRENCY); do
    local test_file="$tmp_dir/conc_${i}_${size_mb}MB.bin"
    generate_test_file "$size_mb" "$test_file"
    file_paths[$i]="$test_file"
  done
  
  # Create entities concurrently
  print_message "$BLUE" "Creating $CONCURRENCY entities concurrently..."
  
  for i in $(seq 1 $CONCURRENCY); do
    (
      local entity_id=$(create_entity_with_content "${file_paths[$i]}" "$size_mb")
      if [ $? -eq 0 ]; then
        echo "$i:$entity_id" > "$tmp_dir/entity_id_$i.txt"
      fi
    ) &
  done
  
  # Wait for all background jobs to complete
  wait
  
  # Collect entity IDs
  for i in $(seq 1 $CONCURRENCY); do
    if [ -f "$tmp_dir/entity_id_$i.txt" ]; then
      local id_data=$(cat "$tmp_dir/entity_id_$i.txt")
      local id=$(echo "$id_data" | cut -d':' -f2)
      entity_ids[$i]="$id"
    fi
  done
  
  sleep 2 # Allow time for all entities to be fully processed
  
  # Verify entities concurrently
  print_message "$BLUE" "Verifying $CONCURRENCY entities concurrently..."
  
  local success_count=0
  local fail_count=0
  
  for i in $(seq 1 $CONCURRENCY); do
    if [ -n "${entity_ids[$i]}" ]; then
      (
        verify_entity_content "${entity_ids[$i]}" "${file_paths[$i]}"
        if [ $? -eq 0 ]; then
          echo "success" > "$tmp_dir/result_$i.txt"
        else
          echo "fail" > "$tmp_dir/result_$i.txt"
        fi
      ) &
    else
      ((fail_count++))
    fi
  done
  
  # Wait for all background jobs to complete
  wait
  
  # Collect results
  for i in $(seq 1 $CONCURRENCY); do
    if [ -f "$tmp_dir/result_$i.txt" ]; then
      local result=$(cat "$tmp_dir/result_$i.txt")
      if [ "$result" = "success" ]; then
        ((success_count++))
      else
        ((fail_count++))
      fi
    fi
  done
  
  # Clean up
  rm -rf "$tmp_dir"
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Concurrent Test Results:"
  print_message "$GREEN" "✅ Successful operations: $success_count"
  print_message "$RED" "❌ Failed operations: $fail_count"
  print_message "$BLUE" "========================================"
  
  return $fail_count
}

# Test temporal features (as-of, changes, diff)
test_temporal_features() {
  local size_mb=2 # Fixed size for temporal tests
  local tmp_dir="./tmp_temporal_files"
  mkdir -p "$tmp_dir"
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Testing temporal features"
  print_message "$BLUE" "========================================"
  
  # Create initial version
  local test_file_v1="$tmp_dir/temporal_v1_${size_mb}MB.bin"
  generate_test_file "$size_mb" "$test_file_v1"
  echo "Version 1" >> "$test_file_v1"
  
  local entity_id=$(create_entity_with_content "$test_file_v1" "$size_mb")
  if [ $? -ne 0 ]; then
    print_message "$RED" "❌ Failed to create initial entity for temporal test"
    rm -rf "$tmp_dir"
    return 1
  fi
  
  # Store the timestamp of version 1
  local timestamp_v1=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  sleep 2
  
  # Create updated version
  local test_file_v2="$tmp_dir/temporal_v2_${size_mb}MB.bin"
  cp "$test_file_v1" "$test_file_v2"
  echo "Version 2" >> "$test_file_v2"
  
  # Update entity
  print_message "$BLUE" "Updating entity for temporal test..."
  
  local encoded_content_v2=$(base64 -w0 "$test_file_v2")
  
  local update_response=$(curl -k -s -X PUT "$SERVER_URL/api/v1/entities/update" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"id\": \"$entity_id\",
      \"tags\": [\"type:test_file\", \"name:temporal_test\", \"size:${size_mb}MB\", \"version:2\"],
      \"content\": {
        \"data\": \"$encoded_content_v2\",
        \"type\": \"application/octet-stream\"
      }
    }")
  
  if [[ "$update_response" != *"\"id\":"* ]]; then
    print_message "$RED" "❌ Failed to update entity for temporal test. Response: $update_response"
    rm -rf "$tmp_dir"
    return 1
  fi
  
  print_message "$GREEN" "✅ Entity updated successfully"
  
  # Store the timestamp of version 2
  local timestamp_v2=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  sleep 2
  
  # Test as-of feature with first version
  print_message "$BLUE" "Testing as-of feature with first version timestamp..."
  
  local as_of_response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/as-of" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"id\": \"$entity_id\",
      \"timestamp\": \"$timestamp_v1\"
    }")
  
  if [[ "$as_of_response" == *"\"version:2\""* ]]; then
    print_message "$RED" "❌ As-of test failed - got version 2 when requested version 1"
    rm -rf "$tmp_dir"
    return 1
  elif [[ "$as_of_response" != *"\"type:test_file\""* ]]; then
    print_message "$RED" "❌ As-of test failed - couldn't retrieve any version. Response: $as_of_response"
    rm -rf "$tmp_dir"
    return 1
  else
    print_message "$GREEN" "✅ As-of test passed"
  fi
  
  # Test history feature
  print_message "$BLUE" "Testing history feature..."
  
  local history_response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/history?id=$entity_id" \
    -H "Authorization: Bearer $TOKEN")
  
  if [[ "$history_response" != *"\"history\":"* ]]; then
    print_message "$RED" "❌ History test failed. Response: $history_response"
    rm -rf "$tmp_dir"
    return 1
  else
    print_message "$GREEN" "✅ History test passed"
  fi
  
  # Test changes feature
  print_message "$BLUE" "Testing changes feature..."
  
  local changes_response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/changes" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"id\": \"$entity_id\",
      \"start_time\": \"$timestamp_v1\",
      \"end_time\": \"$timestamp_v2\"
    }")
  
  if [[ "$changes_response" != *"\"changes\":"* ]]; then
    print_message "$RED" "❌ Changes test failed. Response: $changes_response"
    rm -rf "$tmp_dir"
    return 1
  else
    print_message "$GREEN" "✅ Changes test passed"
  fi
  
  # Test diff feature
  print_message "$BLUE" "Testing diff feature..."
  
  local diff_response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/diff" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"id\": \"$entity_id\",
      \"time1\": \"$timestamp_v1\",
      \"time2\": \"$timestamp_v2\"
    }")
  
  if [[ "$diff_response" != *"\"added\":"* || "$diff_response" != *"\"removed\":"* ]]; then
    print_message "$RED" "❌ Diff test failed. Response: $diff_response"
    rm -rf "$tmp_dir"
    return 1
  else
    print_message "$GREEN" "✅ Diff test passed"
  fi
  
  # Clean up
  rm -rf "$tmp_dir"
  
  print_message "$BLUE" "========================================"
  print_message "$GREEN" "✅ All temporal tests passed"
  print_message "$BLUE" "========================================"
  
  return 0
}

# Run all tests
run_all_tests() {
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Starting EntityDB Integrity Test Suite"
  print_message "$BLUE" "========================================"
  
  # Login
  login
  
  # Run size tests
  run_size_tests
  local size_tests_result=$?
  
  # Run iteration tests
  run_iteration_tests
  local iteration_tests_result=$?
  
  # Run concurrent tests
  run_concurrent_tests
  local concurrent_tests_result=$?
  
  # Run temporal tests
  test_temporal_features
  local temporal_tests_result=$?
  
  # Summarize all results
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "EntityDB Integrity Test Summary"
  print_message "$BLUE" "========================================"
  
  if [ $size_tests_result -eq 0 ]; then
    print_message "$GREEN" "✅ Size tests passed"
  else
    print_message "$RED" "❌ Size tests failed with $size_tests_result errors"
  fi
  
  if [ $iteration_tests_result -eq 0 ]; then
    print_message "$GREEN" "✅ Iteration tests passed"
  else
    print_message "$RED" "❌ Iteration tests failed with $iteration_tests_result errors"
  fi
  
  if [ $concurrent_tests_result -eq 0 ]; then
    print_message "$GREEN" "✅ Concurrent tests passed"
  else
    print_message "$RED" "❌ Concurrent tests failed with $concurrent_tests_result errors"
  fi
  
  if [ $temporal_tests_result -eq 0 ]; then
    print_message "$GREEN" "✅ Temporal tests passed"
  else
    print_message "$RED" "❌ Temporal tests failed"
  fi
  
  local total_errors=$((size_tests_result + iteration_tests_result + concurrent_tests_result + temporal_tests_result))
  
  if [ $total_errors -eq 0 ]; then
    print_message "$GREEN" "✅ ALL TESTS PASSED! Database integrity verified."
  else
    print_message "$RED" "❌ TESTS FAILED with $total_errors total errors. Database integrity issues detected."
  fi
  
  print_message "$BLUE" "========================================"
}

# Main execution
run_all_tests