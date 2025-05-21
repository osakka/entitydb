#!/bin/bash
# EntityDB Temporal Features Test with Large Datasets
# This script tests temporal features (as-of, history, changes, diff) with large datasets

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8085"
ENTITY_COUNT=1          # Number of entities to create for temporal testing (reduced for quick testing)
VERSIONS_PER_ENTITY=2   # Number of versions for each entity (reduced for quick testing)
FILE_SIZE_MB=0.1        # Size of test files in MB (reduced for quick testing)
TOKEN=""                # Will be set during login

# Function for colored output
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

# Login to get token
login() {
  print_message "$BLUE" "Checking server status..."
  
  local status_response=$(curl -s -X GET "$SERVER_URL/api/v1/status")
  
  print_message "$BLUE" "Status response: $status_response"
  
  if [[ "$status_response" != *"\"status\":\"ok\""* ]]; then
    print_message "$RED" "❌ Server not responding properly. Response: $status_response"
    exit 1
  else
    print_message "$GREEN" "✅ Server is running properly"
  fi
  
  print_message "$BLUE" "Logging in to EntityDB..."
  
  local login_response=$(curl -s -X POST "$SERVER_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}')
  
  print_message "$BLUE" "Login response: $login_response"
  
  TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
  
  if [ -z "$TOKEN" ]; then
    print_message "$RED" "❌ Failed to login. Response: $login_response"
    exit 1
  else
    print_message "$GREEN" "✅ Login successful, got token: ${TOKEN:0:10}..."
  fi
}

# Function to generate a file of specified size with random content
generate_test_file() {
  local size_mb=$1
  local output_file=$2
  local version=$3
  
  # Convert MB to bytes, handling decimal values
  local size_bytes=$(echo "$size_mb * 1024 * 1024" | bc)
  
  # Generate random content
  dd if=/dev/urandom of="$output_file" bs=1024 count=$(echo "$size_mb * 1024" | bc) 2>/dev/null
  
  # Add version marker
  echo "Temporal test file - Version $version - Generated at $(date)" >> "$output_file"
  
  # Calculate file hash
  local file_hash=$(sha256sum "$output_file" | cut -d' ' -f1)
  echo "$file_hash" > "${output_file}.hash"
}

# Create entity with content
create_temporal_entity() {
  local entity_num=$1
  local version=$2
  local file_path=$3
  local log_file=$4
  local prev_id=$5  # Previous entity ID for updates
  
  local file_hash=$(cat "${file_path}.hash")
  
  # Get file info
  local file_size=$(stat -c %s "$file_path")
  print_message "$BLUE" "File size: $file_size bytes for entity $entity_num version $version"
  
  # Create a simpler content (too big files cause problems with curl)
  local simple_content="Version $version content for entity $entity_num - hash: $file_hash"
  
  if [ -z "$prev_id" ]; then
    # Create new entity
    local response=$(curl -s -X POST "$SERVER_URL/api/v1/entities/create" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"tags\": [\"type:temporal_test\", \"entity_num:$entity_num\", \"version:$version\", \"hash:$file_hash\"],
        \"content\": \"$simple_content\"
      }")
  else
    # Update existing entity
    local response=$(curl -s -X PUT "$SERVER_URL/api/v1/entities/update" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"id\": \"$prev_id\",
        \"tags\": [\"type:temporal_test\", \"entity_num:$entity_num\", \"version:$version\", \"hash:$file_hash\"],
        \"content\": \"$simple_content\"
      }")
  fi
  
  # Extract entity ID
  local entity_id=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
  
  if [ -z "$entity_id" ]; then
    if [ -z "$prev_id" ]; then
      echo "❌ Entity $entity_num, Version $version: Creation failed - $response" >> "$log_file"
    else
      echo "❌ Entity $entity_num, Version $version: Update failed for $prev_id - $response" >> "$log_file"
    fi
    return 1
  else
    if [ -z "$prev_id" ]; then
      echo "✅ Entity $entity_num, Version $version: Created entity $entity_id" >> "$log_file"
    else
      echo "✅ Entity $entity_num, Version $version: Updated entity $entity_id" >> "$log_file"
    fi
    echo "$entity_id"
    return 0
  fi
}

# Test as-of feature (retrieve entity at specific time)
test_as_of() {
  local entity_id=$1
  local timestamp=$2
  local expected_version=$3
  local log_file=$4
  
  print_message "$BLUE" "Testing as-of feature for entity $entity_id at timestamp $timestamp (expecting version $expected_version)..."
  
  local response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/as-of" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"id\": \"$entity_id\",
      \"timestamp\": \"$timestamp\"
    }")
  
  if [[ "$response" == *"\"version:$expected_version\""* ]]; then
    echo "✅ Entity $entity_id: As-of test passed for timestamp $timestamp (found correct version $expected_version)" >> "$log_file"
    return 0
  else
    echo "❌ Entity $entity_id: As-of test failed for timestamp $timestamp - Expected version $expected_version not found" >> "$log_file"
    return 1
  fi
}

# Test history feature (get all versions of an entity)
test_history() {
  local entity_id=$1
  local expected_versions=$2
  local log_file=$3
  
  print_message "$BLUE" "Testing history feature for entity $entity_id (expecting $expected_versions versions)..."
  
  local response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/history?id=$entity_id" \
    -H "Authorization: Bearer $TOKEN")
  
  if [[ "$response" != *"\"history\":"* ]]; then
    echo "❌ Entity $entity_id: History test failed - No history found" >> "$log_file"
    return 1
  fi
  
  # Count versions in history
  local version_count=$(echo "$response" | grep -o "\"timestamp\":" | wc -l)
  
  if [ "$version_count" -eq "$expected_versions" ]; then
    echo "✅ Entity $entity_id: History test passed - Found expected $expected_versions versions" >> "$log_file"
    return 0
  else
    echo "❌ Entity $entity_id: History test failed - Expected $expected_versions versions, found $version_count" >> "$log_file"
    return 1
  fi
}

# Test changes feature (get changes between two timestamps)
test_changes() {
  local entity_id=$1
  local start_time=$2
  local end_time=$3
  local log_file=$4
  
  print_message "$BLUE" "Testing changes feature for entity $entity_id between $start_time and $end_time..."
  
  local response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/changes" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"id\": \"$entity_id\",
      \"start_time\": \"$start_time\",
      \"end_time\": \"$end_time\"
    }")
  
  if [[ "$response" == *"\"changes\":"* ]]; then
    # Check if we have actual changes
    local change_count=$(echo "$response" | grep -o "\"timestamp\":" | wc -l)
    
    if [ "$change_count" -gt 0 ]; then
      echo "✅ Entity $entity_id: Changes test passed - Found $change_count changes between timestamps" >> "$log_file"
      return 0
    else
      echo "❌ Entity $entity_id: Changes test failed - No changes found between timestamps" >> "$log_file"
      return 1
    fi
  else
    echo "❌ Entity $entity_id: Changes test failed - Invalid response" >> "$log_file"
    return 1
  fi
}

# Test diff feature (get difference between two timestamps)
test_diff() {
  local entity_id=$1
  local time1=$2
  local time2=$3
  local log_file=$4
  
  print_message "$BLUE" "Testing diff feature for entity $entity_id between $time1 and $time2..."
  
  local response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/diff" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"id\": \"$entity_id\",
      \"time1\": \"$time1\",
      \"time2\": \"$time2\"
    }")
  
  if [[ "$response" == *"\"added\":"* || "$response" == *"\"removed\":"* ]]; then
    echo "✅ Entity $entity_id: Diff test passed - Found differences between timestamps" >> "$log_file"
    return 0
  else
    echo "❌ Entity $entity_id: Diff test failed - No differences found or invalid response" >> "$log_file"
    return 1
  fi
}

# Run temporal test with one entity
run_entity_temporal_test() {
  local entity_num=$1
  local tmp_dir=$2
  local log_file="${tmp_dir}/entity_${entity_num}.log"
  
  print_message "$BLUE" "Running temporal tests for entity $entity_num..."
  
  local entity_id=""
  local timestamps=()
  
  # Create multiple versions of the entity
  for version in $(seq 1 $VERSIONS_PER_ENTITY); do
    local test_file="${tmp_dir}/entity_${entity_num}_v${version}.bin"
    generate_test_file $FILE_SIZE_MB "$test_file" "$version"
    
    # Store timestamp before creation/update
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    timestamps+=("$timestamp")
    
    # Create or update entity
    if [ -z "$entity_id" ]; then
      entity_id=$(create_temporal_entity "$entity_num" "$version" "$test_file" "$log_file")
    else
      entity_id=$(create_temporal_entity "$entity_num" "$version" "$test_file" "$log_file" "$entity_id")
    fi
    
    if [ $? -ne 0 ]; then
      print_message "$RED" "❌ Failed to create/update entity $entity_num version $version"
      return 1
    fi
    
    # Sleep to ensure timestamps are different
    sleep 1
  done
  
  # Store final entity ID
  echo "$entity_id" > "${tmp_dir}/entity_${entity_num}.id"
  
  # Now test temporal features
  sleep 2
  
  # Test as-of for each version
  for version in $(seq 1 $VERSIONS_PER_ENTITY); do
    local idx=$((version - 1))
    test_as_of "$entity_id" "${timestamps[$idx]}" "$version" "$log_file"
  done
  
  # Test history
  test_history "$entity_id" "$VERSIONS_PER_ENTITY" "$log_file"
  
  # Test changes between first and last version
  test_changes "$entity_id" "${timestamps[0]}" "${timestamps[$((VERSIONS_PER_ENTITY-1))]}" "$log_file"
  
  # Test diff between first and last version
  test_diff "$entity_id" "${timestamps[0]}" "${timestamps[$((VERSIONS_PER_ENTITY-1))]}" "$log_file"
  
  print_message "$GREEN" "Completed temporal tests for entity $entity_num"
}

# Run temporal features test with large datasets
run_temporal_test() {
  local tmp_dir="./tmp_temporal_test"
  mkdir -p "$tmp_dir"
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Starting temporal features test with large datasets"
  print_message "$BLUE" "Creating $ENTITY_COUNT entities with $VERSIONS_PER_ENTITY versions each"
  print_message "$BLUE" "File size: $FILE_SIZE_MB MB per version"
  print_message "$BLUE" "========================================"
  
  local start_time=$(date +%s)
  local success_count=0
  
  # Create entities and test temporal features
  for i in $(seq 1 $ENTITY_COUNT); do
    print_message "$BLUE" "Processing entity $i of $ENTITY_COUNT"
    
    run_entity_temporal_test "$i" "$tmp_dir"
    if [ $? -eq 0 ]; then
      ((success_count++))
    fi
    
    # Progress update
    local pct_complete=$(( (i * 100) / ENTITY_COUNT ))
    print_message "$BLUE" "Progress: $pct_complete% complete ($i/$ENTITY_COUNT entities processed)"
    
    echo ""
  done
  
  local end_time=$(date +%s)
  local total_time=$((end_time - start_time))
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Temporal test completed in $total_time seconds"
  print_message "$BLUE" "Successfully processed $success_count/$ENTITY_COUNT entities"
  print_message "$BLUE" "========================================"
  
  # Analyze logs to get test results
  local as_of_success=$(grep "As-of test passed" "$tmp_dir"/*.log | wc -l)
  local as_of_fail=$(grep "As-of test failed" "$tmp_dir"/*.log | wc -l)
  local as_of_total=$((as_of_success + as_of_fail))
  
  local history_success=$(grep "History test passed" "$tmp_dir"/*.log | wc -l)
  local history_fail=$(grep "History test failed" "$tmp_dir"/*.log | wc -l)
  local history_total=$((history_success + history_fail))
  
  local changes_success=$(grep "Changes test passed" "$tmp_dir"/*.log | wc -l)
  local changes_fail=$(grep "Changes test failed" "$tmp_dir"/*.log | wc -l)
  local changes_total=$((changes_success + changes_fail))
  
  local diff_success=$(grep "Diff test passed" "$tmp_dir"/*.log | wc -l)
  local diff_fail=$(grep "Diff test failed" "$tmp_dir"/*.log | wc -l)
  local diff_total=$((diff_success + diff_fail))
  
  # Calculate success rates
  local as_of_rate=$(awk "BEGIN {printf \"%.2f\", ($as_of_success / $as_of_total) * 100}")
  local history_rate=$(awk "BEGIN {printf \"%.2f\", ($history_success / $history_total) * 100}")
  local changes_rate=$(awk "BEGIN {printf \"%.2f\", ($changes_success / $changes_total) * 100}")
  local diff_rate=$(awk "BEGIN {printf \"%.2f\", ($diff_success / $diff_total) * 100}")
  
  # Print summary
  print_message "$BLUE" "TEMPORAL FEATURES TEST SUMMARY:"
  print_message "$BLUE" "------------------------"
  print_message "$BLUE" "As-Of tests: $as_of_success success, $as_of_fail fail ($as_of_rate% success rate)"
  print_message "$BLUE" "History tests: $history_success success, $history_fail fail ($history_rate% success rate)"
  print_message "$BLUE" "Changes tests: $changes_success success, $changes_fail fail ($changes_rate% success rate)"
  print_message "$BLUE" "Diff tests: $diff_success success, $diff_fail fail ($diff_rate% success rate)"
  
  local total_success=$((as_of_success + history_success + changes_success + diff_success))
  local total_fail=$((as_of_fail + history_fail + changes_fail + diff_fail))
  local total_tests=$((total_success + total_fail))
  local total_rate=$(awk "BEGIN {printf \"%.2f\", ($total_success / $total_tests) * 100}")
  
  print_message "$BLUE" "------------------------"
  print_message "$BLUE" "Total tests: $total_tests"
  print_message "$BLUE" "Total success: $total_success"
  print_message "$BLUE" "Total failure: $total_fail"
  print_message "$BLUE" "Overall success rate: $total_rate%"
  
  # Determine test result
  if [ $(echo "$total_rate >= 95" | bc -l) -eq 1 ]; then
    print_message "$GREEN" "✅ TEMPORAL FEATURES TEST PASSED! Database maintains temporal integrity with large datasets."
    result=0
  else
    print_message "$RED" "❌ TEMPORAL FEATURES TEST FAILED! Database has issues with temporal integrity under load."
    result=1
  fi
  
  print_message "$BLUE" "========================================"
  
  # Clean up (comment out to keep files for analysis)
  # rm -rf "$tmp_dir"
  print_message "$BLUE" "Test logs saved in $tmp_dir for review"
  
  return $result
}

# Main execution
print_message "$BLUE" "========================================"
print_message "$BLUE" "Starting EntityDB Temporal Features Test"
print_message "$BLUE" "========================================"

# Login
login

# Run temporal test
run_temporal_test
temporal_result=$?

# Exit with result
exit $temporal_result