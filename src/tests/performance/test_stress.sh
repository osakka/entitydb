#!/bin/bash
# EntityDB Stress Test
# This script runs a stress test to verify system stability with many entities

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8085"
ENTITY_COUNT=100         # Total number of entities to create (reduced for quicker testing)
BATCH_SIZE=10            # Number of entities per batch
QUERY_INTERVAL=20        # Run queries after every N entities
CONTENT_SIZE_KB=5        # Size of content for each entity
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

# Generate random string of specified size
generate_random_string() {
  local size_kb=$1
  local bytes=$((size_kb * 1024))
  
  # Create random alphanumeric content
  cat /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c $bytes
}

# Create a batch of entities
create_entity_batch() {
  local batch_number=$1
  local batch_size=$2
  local content_size_kb=$3
  local result_file=$4
  
  print_message "$BLUE" "Creating batch $batch_number ($batch_size entities)..."
  
  local batch_start=$((batch_number * batch_size))
  local success_count=0
  local content=""
  
  for i in $(seq 1 $batch_size); do
    local entity_number=$((batch_start + i))
    
    # Generate random content
    content=$(generate_random_string $content_size_kb)
    
    # Create the entity
    local response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/create" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"tags\": [\"type:stress_test\", \"batch:$batch_number\", \"number:$entity_number\"],
        \"content\": \"$content\"
      }")
    
    # Extract entity ID
    local entity_id=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    
    if [ -n "$entity_id" ]; then
      echo "$entity_id" >> "$result_file"
      ((success_count++))
    fi
  done
  
  print_message "$GREEN" "✅ Created $success_count/$batch_size entities in batch $batch_number"
  echo "$success_count"
}

# Run a query to retrieve entities by tag
run_query() {
  local query_tag=$1
  
  print_message "$BLUE" "Running query for tag '$query_tag'..."
  
  local start_time=$(date +%s.%N)
  
  local response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/list?tag=$query_tag" \
    -H "Authorization: Bearer $TOKEN")
  
  local end_time=$(date +%s.%N)
  local query_time=$(echo "$end_time - $start_time" | bc)
  
  local entity_count=$(echo "$response" | grep -o "\"id\":" | wc -l)
  
  print_message "$GREEN" "✅ Query returned $entity_count entities in $query_time seconds"
  
  echo "$entity_count:$query_time"
}

# Verify database consistency
verify_database() {
  local id_file=$1
  
  print_message "$BLUE" "Verifying database consistency..."
  
  local total_ids=$(wc -l < "$id_file")
  local sample_size=50
  local success_count=0
  
  # Select random sample of IDs
  local sample_ids=($(shuf -n $sample_size "$id_file"))
  
  for id in "${sample_ids[@]}"; do
    local response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/get?id=$id" \
      -H "Authorization: Bearer $TOKEN")
    
    if [[ "$response" == *"\"id\":\"$id\""* ]]; then
      ((success_count++))
    fi
  done
  
  local success_rate=$(awk "BEGIN {printf \"%.2f\", ($success_count / $sample_size) * 100}")
  
  print_message "$BLUE" "Verification: $success_count/$sample_size entities retrievable ($success_rate%)"
  
  if [ $(echo "$success_rate >= 95" | bc -l) -eq 1 ]; then
    print_message "$GREEN" "✅ Database consistency check passed"
    return 0
  else
    print_message "$RED" "❌ Database consistency check failed"
    return 1
  fi
}

# Run query performance tests
run_query_performance_test() {
  local log_dir=$1
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Running query performance tests..."
  print_message "$BLUE" "========================================"
  
  local performance_log="${log_dir}/query_performance.log"
  
  # Run different types of queries
  local query_types=("type:stress_test" "batch:1" "number:50" "type:stress_test&batch:5")
  
  for query in "${query_types[@]}"; do
    # Run the query multiple times to check consistency
    local total_time=0
    local iterations=3
    local total_count=0
    
    for i in $(seq 1 $iterations); do
      local result=$(run_query "$query")
      local count=$(echo "$result" | cut -d':' -f1)
      local time=$(echo "$result" | cut -d':' -f2)
      
      total_count=$((total_count + count))
      total_time=$(echo "$total_time + $time" | bc)
      
      sleep 1
    done
    
    local avg_count=$((total_count / iterations))
    local avg_time=$(echo "scale=3; $total_time / $iterations" | bc)
    
    echo "Query: $query, Avg Count: $avg_count, Avg Time: ${avg_time}s" >> "$performance_log"
    print_message "$BLUE" "Query '$query': $avg_count entities, ${avg_time}s average"
  done
  
  print_message "$GREEN" "✅ Query performance tests completed. See $performance_log for details."
}

# Run the stress test
run_stress_test() {
  local tmp_dir="./tmp_stress_test"
  mkdir -p "$tmp_dir"
  
  local id_file="${tmp_dir}/entity_ids.txt"
  local perf_file="${tmp_dir}/performance.log"
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Starting stress test with $ENTITY_COUNT entities"
  print_message "$BLUE" "Content size: $CONTENT_SIZE_KB KB per entity"
  print_message "$BLUE" "========================================"
  
  local start_time=$(date +%s)
  local total_success=0
  local query_times=()
  
  # Create entities in batches
  local batch_count=$((ENTITY_COUNT / BATCH_SIZE))
  
  for i in $(seq 0 $((batch_count - 1))); do
    local batch_success=$(create_entity_batch "$i" "$BATCH_SIZE" "$CONTENT_SIZE_KB" "$id_file")
    total_success=$((total_success + batch_success))
    
    # Run queries periodically
    if [ $((i % (QUERY_INTERVAL / BATCH_SIZE))) -eq 0 ] && [ $i -gt 0 ]; then
      local query_result=$(run_query "type:stress_test")
      local query_time=$(echo "$query_result" | cut -d':' -f2)
      query_times+=("$query_time")
      
      echo "After $((i * BATCH_SIZE)) entities, query time: $query_time" >> "$perf_file"
    fi
  done
  
  local end_time=$(date +%s)
  local total_time=$((end_time - start_time))
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Stress test completed in $total_time seconds"
  print_message "$BLUE" "Successfully created $total_success/$ENTITY_COUNT entities"
  print_message "$BLUE" "========================================"
  
  # Calculate average query time
  local total_query_time=0
  local query_count=${#query_times[@]}
  
  for time in "${query_times[@]}"; do
    total_query_time=$(echo "$total_query_time + $time" | bc)
  done
  
  local avg_query_time=0
  if [ $query_count -gt 0 ]; then
    avg_query_time=$(echo "scale=3; $total_query_time / $query_count" | bc)
  fi
  
  print_message "$BLUE" "Average query time: ${avg_query_time}s"
  
  # Verify database consistency
  verify_database "$id_file"
  local verify_result=$?
  
  # Run additional query performance tests
  run_query_performance_test "$tmp_dir"
  
  # Clean up (comment out to keep files for analysis)
  # rm -rf "$tmp_dir"
  print_message "$BLUE" "Test data saved in $tmp_dir for analysis"
  
  if [ $verify_result -eq 0 ] && [ $total_success -ge $((ENTITY_COUNT * 95 / 100)) ]; then
    print_message "$GREEN" "✅ STRESS TEST PASSED! System maintained stability with $ENTITY_COUNT entities."
    return 0
  else
    print_message "$RED" "❌ STRESS TEST FAILED! System may have stability issues with large entity counts."
    return 1
  fi
}

# Main execution
print_message "$BLUE" "========================================"
print_message "$BLUE" "Starting EntityDB Stress Test"
print_message "$BLUE" "========================================"

# Login
login

# Run stress test
run_stress_test
stress_result=$?

exit $stress_result