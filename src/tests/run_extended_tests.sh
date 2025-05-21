#!/bin/bash
# Extended test script for EntityDB
# Uses the existing test framework but with increased iterations and larger content sizes

# Base directory
BASE_DIR="$(dirname "$0")"
TESTS_DIR="$BASE_DIR/src/tests"

# Import test framework
source "$TESTS_DIR/test_framework.sh"

# Configuration - adjust these for more intensive testing
TEST_ITERATIONS=100
BATCH_SIZE=10
LARGE_CONTENT_SIZE=5000000  # 5MB in bytes
CONCURRENCY=5

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to generate random data of specified size
generate_random_data() {
  local size=$1
  head -c $size /dev/urandom | base64 | tr -d '\n'
}

# Function to run tests repeatedly
run_repeated_tests() {
  local test_name=$1
  local iterations=$2
  local success=0
  local failure=0
  
  echo -e "${BLUE}Running $iterations iterations of $test_name test...${NC}"
  
  for ((i=1; i<=$iterations; i++)); do
    echo -e "${YELLOW}Iteration $i of $iterations${NC}"
    
    # Run the test using the framework
    run_test "$test_name"
    if [ $? -eq 0 ]; then
      ((success++))
    else
      ((failure++))
    fi
  done
  
  echo -e "${BLUE}Test $test_name completed: $success passed, $failure failed${NC}"
  
  if [ $failure -eq 0 ]; then
    return 0
  else
    return 1
  fi
}

# Run batch of tests concurrently 
run_concurrent_tests() {
  local test_name=$1
  local concurrency=$2
  local iterations_per_thread=$3
  local total_iterations=$((concurrency * iterations_per_thread))
  
  echo -e "${BLUE}Running $test_name test with $concurrency concurrent threads, $total_iterations total iterations...${NC}"
  
  # Create temp directory for results
  local tmp_dir="$TEMP_DIR/concurrent_$test_name"
  mkdir -p "$tmp_dir"
  
  # Launch concurrent test threads
  for ((thread=1; thread<=$concurrency; thread++)); do
    (
      for ((i=1; i<=$iterations_per_thread; i++)); do
        iteration=$((((thread-1) * iterations_per_thread) + i))
        echo -e "${YELLOW}Thread $thread, Iteration $i of $iterations_per_thread (overall: $iteration of $total_iterations)${NC}"
        
        # Run the test
        run_test "$test_name"
        if [ $? -eq 0 ]; then
          echo "success" >> "$tmp_dir/thread_${thread}_results.txt"
        else
          echo "failure" >> "$tmp_dir/thread_${thread}_results.txt"
        fi
      done
    ) &
    
    # Small delay to avoid exact same timestamp
    sleep 0.5
  done
  
  # Wait for all threads to complete
  wait
  
  # Gather results
  local total_success=0
  local total_failure=0
  
  for ((thread=1; thread<=$concurrency; thread++)); do
    if [ -f "$tmp_dir/thread_${thread}_results.txt" ]; then
      local success=$(grep -c "success" "$tmp_dir/thread_${thread}_results.txt")
      local failure=$(grep -c "failure" "$tmp_dir/thread_${thread}_results.txt")
      
      total_success=$((total_success + success))
      total_failure=$((total_failure + failure))
    fi
  done
  
  echo -e "${BLUE}Concurrent test $test_name completed: $total_success passed, $total_failure failed${NC}"
  
  if [ $total_failure -eq 0 ]; then
    return 0
  else
    return 1
  fi
}

# Create a large entity and test it
test_large_entity() {
  local size=$1
  local success=0
  local failure=0
  
  echo -e "${BLUE}Testing with large entity ($((size / 1000000)) MB)...${NC}"
  
  # Generate large random content
  local content=$(generate_random_data $size)
  
  # Create entity with large content
  echo -e "${YELLOW}Creating large entity...${NC}"
  local response=$(create_entity "[\"type:large_test\",\"size:$size\"]" "$content")
  local entity_id=$(echo "$response" | grep -o '"id":"[^"]*' | head -1 | sed 's/"id":"//')
  
  if [[ -z "$entity_id" ]]; then
    echo -e "${RED}Failed to extract entity ID. Response: $response${NC}"
    return 1
  fi
  
  echo -e "${GREEN}Created large entity with ID: $entity_id${NC}"
  
  # Run temporal tests on the large entity
  echo -e "${YELLOW}Testing temporal features with large entity...${NC}"
  
  # Test history
  run_test_with_entity "entity_history" "$entity_id"
  if [ $? -eq 0 ]; then
    ((success++))
  else
    ((failure++))
  fi
  
  # Test as-of
  run_test_with_entity "entity_as_of" "$entity_id"
  if [ $? -eq 0 ]; then
    ((success++))
  else
    ((failure++))
  fi
  
  # Create more versions of this entity to test diff
  local versions=3
  for ((v=1; v<=$versions; v++)); do
    # Update the content slightly
    local updated_content="${content}$(generate_random_data 1000)v$v"
    
    # Update entity
    echo -e "${YELLOW}Creating version $v of large entity...${NC}"
    local update_response=$(update_entity "$entity_id" "[\"type:large_test\",\"size:$size\",\"version:$v\"]" "$updated_content")
    
    if [[ "$update_response" != *"\"id\":"* ]]; then
      echo -e "${RED}Failed to update entity. Response: $update_response${NC}"
      ((failure++))
    else
      ((success++))
    fi
    
    # Small delay between versions
    sleep 1
  done
  
  # Test changes
  run_test_with_entity "entity_changes" "$entity_id"
  if [ $? -eq 0 ]; then
    ((success++))
  else
    ((failure++))
  fi
  
  # Test diff
  run_test_with_entity "entity_diff" "$entity_id"
  if [ $? -eq 0 ]; then
    ((success++))
  else
    ((failure++))
  fi
  
  echo -e "${BLUE}Large entity tests completed: $success passed, $failure failed${NC}"
  
  if [ $failure -eq 0 ]; then
    return 0
  else
    return 1
  fi
}

# Run all tests in batches
run_all_tests_in_batches() {
  local test_files=("create_entity" "list_entities" "entity_history" "entity_as_of" "entity_changes" "entity_diff")
  local iterations=$1
  local batch_size=$2
  local batches=$((iterations / batch_size))
  local success=0
  local failure=0
  
  echo -e "${BLUE}Running $iterations total iterations in $batches batches of $batch_size...${NC}"
  
  for ((batch=1; batch<=$batches; batch++)); do
    echo -e "${BLUE}==== Batch $batch of $batches ====${NC}"
    
    # Run each test type in this batch
    for test_name in "${test_files[@]}"; do
      echo -e "${YELLOW}Running $batch_size iterations of $test_name in batch $batch${NC}"
      
      # Run batch_size iterations of this test
      for ((i=1; i<=$batch_size; i++)); do
        iteration=$(( ((batch-1) * batch_size * ${#test_files[@]}) + (i * ${#test_files[@]}) ))
        total_iterations=$((iterations * ${#test_files[@]}))
        
        echo -e "${YELLOW}Test $test_name, Iteration $iteration of $total_iterations${NC}"
        
        run_test "$test_name"
        if [ $? -eq 0 ]; then
          ((success++))
        else
          ((failure++))
        fi
      done
    done
    
    # Print batch summary
    echo -e "${BLUE}==== Batch $batch completed: Current stats - $success passed, $failure failed ====${NC}"
  done
  
  echo -e "${BLUE}All batched tests completed: $success passed, $failure failed${NC}"
  
  if [ $failure -eq 0 ]; then
    return 0
  else
    return 1
  fi
}

# Main test sequence
main() {
  echo -e "${BLUE}======================================${NC}"
  echo -e "${BLUE}EntityDB Extended Testing${NC}"
  echo -e "${BLUE}======================================${NC}"
  
  # Initialize with clean database
  initialize "clean"
  
  # Login first
  login
  if [ $? -ne 0 ]; then
    echo -e "${RED}Login failed. Cannot continue with tests.${NC}"
    exit 1
  fi
  
  echo -e "${BLUE}======================================${NC}"
  echo -e "${BLUE}Running Basic Tests${NC}"
  echo -e "${BLUE}======================================${NC}"
  
  # Run repeated tests for basic functionality
  run_repeated_tests "login_admin" $((TEST_ITERATIONS / 10))
  login_result=$?
  
  run_repeated_tests "create_entity" $((TEST_ITERATIONS / 5))
  create_result=$?
  
  run_repeated_tests "list_entities" $((TEST_ITERATIONS / 5))
  list_result=$?
  
  echo -e "${BLUE}======================================${NC}"
  echo -e "${BLUE}Running Concurrent Tests${NC}"
  echo -e "${BLUE}======================================${NC}"
  
  # Run concurrent tests
  run_concurrent_tests "create_entity" $CONCURRENCY $((TEST_ITERATIONS / CONCURRENCY))
  concurrent_create_result=$?
  
  echo -e "${BLUE}======================================${NC}"
  echo -e "${BLUE}Testing with Large Entities${NC}"
  echo -e "${BLUE}======================================${NC}"
  
  # Test with increasingly large entities
  for size in 1000000 2000000 5000000 10000000; do
    test_large_entity $size
    large_entity_result=$?
    
    if [ $large_entity_result -ne 0 ]; then
      echo -e "${RED}Large entity test failed at size $((size / 1000000)) MB${NC}"
      break
    fi
  done
  
  echo -e "${BLUE}======================================${NC}"
  echo -e "${BLUE}Running Batch Tests${NC}"
  echo -e "${BLUE}======================================${NC}"
  
  # Run all tests in batches
  run_all_tests_in_batches $TEST_ITERATIONS $BATCH_SIZE
  batch_result=$?
  
  echo -e "${BLUE}======================================${NC}"
  echo -e "${BLUE}Test Summary${NC}"
  echo -e "${BLUE}======================================${NC}"
  
  if [ $login_result -eq 0 ]; then
    echo -e "${GREEN}✓ Basic Login Tests: Passed${NC}"
  else
    echo -e "${RED}✗ Basic Login Tests: Failed${NC}"
  fi
  
  if [ $create_result -eq 0 ]; then
    echo -e "${GREEN}✓ Entity Creation Tests: Passed${NC}"
  else
    echo -e "${RED}✗ Entity Creation Tests: Failed${NC}"
  fi
  
  if [ $list_result -eq 0 ]; then
    echo -e "${GREEN}✓ Entity Listing Tests: Passed${NC}"
  else
    echo -e "${RED}✗ Entity Listing Tests: Failed${NC}"
  fi
  
  if [ $concurrent_create_result -eq 0 ]; then
    echo -e "${GREEN}✓ Concurrent Creation Tests: Passed${NC}"
  else
    echo -e "${RED}✗ Concurrent Creation Tests: Failed${NC}"
  fi
  
  if [ $large_entity_result -eq 0 ]; then
    echo -e "${GREEN}✓ Large Entity Tests: Passed${NC}"
  else
    echo -e "${RED}✗ Large Entity Tests: Failed${NC}"
  fi
  
  if [ $batch_result -eq 0 ]; then
    echo -e "${GREEN}✓ Batch Tests: Passed${NC}"
  else
    echo -e "${RED}✗ Batch Tests: Failed${NC}"
  fi
  
  # Overall result
  if [ $login_result -eq 0 ] && [ $create_result -eq 0 ] && [ $list_result -eq 0 ] && 
     [ $concurrent_create_result -eq 0 ] && [ $large_entity_result -eq 0 ] && [ $batch_result -eq 0 ]; then
    echo -e "${GREEN}✓ ALL EXTENDED TESTS PASSED!${NC}"
    echo -e "${GREEN}Database integrity verified with large datasets and many iterations.${NC}"
    exit 0
  else
    echo -e "${RED}✗ SOME EXTENDED TESTS FAILED!${NC}"
    echo -e "${RED}Database may have integrity issues under load or with large files.${NC}"
    exit 1
  fi
}

# Start testing
main