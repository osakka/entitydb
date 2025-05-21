#!/bin/bash
# EntityDB Large File Test
# This script tests the database's ability to handle large files

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8085"
START_SIZE_MB=1          # Starting file size in MB
MAX_SIZE_MB=8            # Maximum file size to test in MB (reduced for quicker testing)
SIZE_INCREMENT_MB=3      # Size increment in MB
TEMP_DIR="/tmp/entitydb_large_tests"
TOKEN=""

# Function for colored output
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

# Clean up from previous runs and create temp directory
cleanup() {
  rm -rf "$TEMP_DIR"
  mkdir -p "$TEMP_DIR"
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

# Generate a test file with random content
generate_test_file() {
  local size_mb=$1
  local output_file=$2
  
  print_message "$BLUE" "Generating $size_mb MB test file..."
  dd if=/dev/urandom of="$output_file" bs=1M count="$size_mb" 2>/dev/null
  
  # Add a marker at the end of the file for verification
  echo "Test file generated at $(date) with size ${size_mb}MB" >> "$output_file"
  
  local file_hash=$(sha256sum "$output_file" | cut -d' ' -f1)
  print_message "$GREEN" "✅ Generated $size_mb MB test file with hash: $file_hash"
  
  # Save hash for verification
  echo "$file_hash" > "${output_file}.hash"
}

# Create entity with file content - using file-based approach to avoid argument list too long
create_entity_with_file() {
  local file_path=$1
  local size_mb=$2
  local output_json="$TEMP_DIR/create_request.json"
  local response_file="$TEMP_DIR/create_response.json"
  
  # Calculate file hash for verification
  local file_hash=$(cat "${file_path}.hash")
  
  # Create JSON payload file
  cat > "$output_json" << EOL
{
  "tags": ["type:large_file_test", "size:${size_mb}MB", "hash:${file_hash}"],
  "content": {
    "data": "$(base64 -w0 "$file_path")",
    "type": "application/octet-stream"
  }
}
EOL
  
  print_message "$BLUE" "Creating entity with $size_mb MB file..."
  
  # Use the JSON file for the request
  curl -k -s -X POST "$SERVER_URL/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d @"$output_json" > "$response_file"
  
  # Extract entity ID
  local entity_id=$(grep -o '"id":"[^"]*' "$response_file" | head -1 | cut -d'"' -f4)
  
  # Clean up
  rm -f "$output_json"
  
  if [ -z "$entity_id" ]; then
    print_message "$RED" "❌ Failed to create entity with $size_mb MB file."
    cat "$response_file"
    return 1
  else
    print_message "$GREEN" "✅ Created entity with ID: $entity_id"
    echo "$entity_id" > "${file_path}.id"
    return 0
  fi
}

# Verify entity content
verify_entity_content() {
  local file_path=$1
  local size_mb=$2
  local entity_id=$(cat "${file_path}.id")
  local original_hash=$(cat "${file_path}.hash")
  local output_file="${file_path}.retrieved"
  local response_file="$TEMP_DIR/get_response.json"
  
  print_message "$BLUE" "Verifying content for $size_mb MB entity (ID: $entity_id)..."
  
  # Get the entity
  curl -k -s -X GET "$SERVER_URL/api/v1/entities/get?id=$entity_id" \
    -H "Authorization: Bearer $TOKEN" > "$response_file"
  
  # Extract and decode content
  local content=$(grep -o '"data":"[^"]*' "$response_file" | head -1 | cut -d'"' -f4)
  
  if [ -z "$content" ]; then
    print_message "$RED" "❌ Failed to get content for entity $entity_id."
    cat "$response_file"
    return 1
  fi
  
  # Decode the content
  echo "$content" | base64 -d > "$output_file"
  
  # Verify file hash
  local retrieved_hash=$(sha256sum "$output_file" | cut -d' ' -f1)
  
  if [ "$original_hash" == "$retrieved_hash" ]; then
    print_message "$GREEN" "✅ Content verification passed! Hashes match."
    return 0
  else
    print_message "$RED" "❌ Content verification failed!"
    print_message "$RED" "Original hash: $original_hash"
    print_message "$RED" "Retrieved hash: $retrieved_hash"
    return 1
  fi
}

# Test with gradually increasing file sizes
test_incremental_sizes() {
  local success_count=0
  local fail_count=0
  
  for size_mb in $(seq $START_SIZE_MB $SIZE_INCREMENT_MB $MAX_SIZE_MB); do
    print_message "$BLUE" "========================================"
    print_message "$BLUE" "Testing with file size: $size_mb MB"
    print_message "$BLUE" "========================================"
    
    local test_file="$TEMP_DIR/test_${size_mb}MB.bin"
    
    # Generate test file
    generate_test_file "$size_mb" "$test_file"
    
    # Create entity
    create_entity_with_file "$test_file" "$size_mb"
    if [ $? -ne 0 ]; then
      print_message "$RED" "❌ Failed to create entity with $size_mb MB file."
      ((fail_count++))
      continue
    fi
    
    # Verify content
    verify_entity_content "$test_file" "$size_mb"
    if [ $? -eq 0 ]; then
      print_message "$GREEN" "✅ Test passed for $size_mb MB file."
      ((success_count++))
    else
      print_message "$RED" "❌ Test failed for $size_mb MB file."
      ((fail_count++))
    fi
    
    echo ""
  done
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Test Summary:"
  print_message "$GREEN" "✅ Successful tests: $success_count"
  print_message "$RED" "❌ Failed tests: $fail_count"
  print_message "$BLUE" "========================================"
  
  if [ $fail_count -eq 0 ]; then
    print_message "$GREEN" "✅ ALL TESTS PASSED!"
    return 0
  else
    print_message "$RED" "❌ SOME TESTS FAILED!"
    return 1
  fi
}

# Update entity to test temporal features
update_entity() {
  local file_path=$1
  local version=$2
  local entity_id=$(cat "${file_path}.id")
  local output_json="$TEMP_DIR/update_request.json"
  local response_file="$TEMP_DIR/update_response.json"
  
  # Modify the file for this version
  local modified_file="${file_path}.v${version}"
  cp "$file_path" "$modified_file"
  echo "Version $version update at $(date)" >> "$modified_file"
  
  # Calculate new hash
  local file_hash=$(sha256sum "$modified_file" | cut -d' ' -f1)
  echo "$file_hash" > "${modified_file}.hash"
  
  # Create JSON payload file
  cat > "$output_json" << EOL
{
  "id": "${entity_id}",
  "tags": ["type:large_file_test", "version:${version}", "hash:${file_hash}"],
  "content": {
    "data": "$(base64 -w0 "$modified_file")",
    "type": "application/octet-stream"
  }
}
EOL
  
  print_message "$BLUE" "Updating entity with version $version..."
  
  # Use the JSON file for the request
  curl -k -s -X PUT "$SERVER_URL/api/v1/entities/update" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d @"$output_json" > "$response_file"
  
  # Verify update
  if grep -q '"id"' "$response_file"; then
    print_message "$GREEN" "✅ Updated entity $entity_id to version $version"
    return 0
  else
    print_message "$RED" "❌ Failed to update entity to version $version"
    cat "$response_file"
    return 1
  fi
}

# Test entity history
test_entity_history() {
  local file_path=$1
  local entity_id=$(cat "${file_path}.id")
  local response_file="$TEMP_DIR/history_response.json"
  
  print_message "$BLUE" "Testing history for entity $entity_id..."
  
  # Get history
  curl -k -s -X GET "$SERVER_URL/api/v1/entities/history?id=$entity_id" \
    -H "Authorization: Bearer $TOKEN" > "$response_file"
  
  # Check for history data
  if grep -q '"history"' "$response_file"; then
    local version_count=$(grep -c '"timestamp"' "$response_file")
    print_message "$GREEN" "✅ History test passed - found $version_count versions"
    return 0
  else
    print_message "$RED" "❌ History test failed - no history data found"
    cat "$response_file"
    return 1
  fi
}

# Test entity as-of feature
test_entity_as_of() {
  local file_path=$1
  local entity_id=$(cat "${file_path}.id")
  local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  local response_file="$TEMP_DIR/as_of_response.json"
  
  print_message "$BLUE" "Testing as-of feature for entity $entity_id at timestamp $timestamp..."
  
  # Create JSON payload
  local payload="{\"id\":\"${entity_id}\",\"timestamp\":\"${timestamp}\"}"
  
  # Get entity as of timestamp
  curl -k -s -X POST "$SERVER_URL/api/v1/entities/as-of" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$payload" > "$response_file"
  
  # Check response
  if grep -q '"id"' "$response_file"; then
    print_message "$GREEN" "✅ As-of test passed - entity retrieved at timestamp"
    return 0
  else
    print_message "$RED" "❌ As-of test failed - could not retrieve entity at timestamp"
    cat "$response_file"
    return 1
  fi
}

# Test temporal features with large entities
test_temporal_features() {
  local size_mb=$1
  local test_file="$TEMP_DIR/temporal_${size_mb}MB.bin"
  local success_count=0
  local fail_count=0
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Testing temporal features with $size_mb MB file"
  print_message "$BLUE" "========================================"
  
  # Generate test file
  generate_test_file "$size_mb" "$test_file"
  
  # Create initial entity
  create_entity_with_file "$test_file" "$size_mb"
  if [ $? -ne 0 ]; then
    print_message "$RED" "❌ Failed to create initial entity."
    return 1
  fi
  
  # Record timestamp after creation
  local timestamp1=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  sleep 2
  
  # Create multiple versions
  for version in {1..3}; do
    update_entity "$test_file" "$version"
    if [ $? -eq 0 ]; then
      ((success_count++))
    else
      ((fail_count++))
    fi
    sleep 2
  done
  
  # Record timestamp after updates
  local timestamp2=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  
  # Test history
  test_entity_history "$test_file"
  if [ $? -eq 0 ]; then
    ((success_count++))
  else
    ((fail_count++))
  fi
  
  # Test as-of
  test_entity_as_of "$test_file"
  if [ $? -eq 0 ]; then
    ((success_count++))
  else
    ((fail_count++))
  fi
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Temporal Test Summary:"
  print_message "$GREEN" "✅ Successful operations: $success_count"
  print_message "$RED" "❌ Failed operations: $fail_count"
  print_message "$BLUE" "========================================"
  
  if [ $fail_count -eq 0 ]; then
    print_message "$GREEN" "✅ TEMPORAL TESTS PASSED!"
    return 0
  else
    print_message "$RED" "❌ SOME TEMPORAL TESTS FAILED!"
    return 1
  fi
}

# Main function
main() {
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "EntityDB Large File Test"
  print_message "$BLUE" "========================================"
  
  # Setup
  cleanup
  login
  
  # Run incremental size tests
  test_incremental_sizes
  size_test_result=$?
  
  # Run temporal tests with a large file
  test_temporal_features 10
  temporal_test_result=$?
  
  # Final result
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "FINAL RESULT:"
  
  if [ $size_test_result -eq 0 ] && [ $temporal_test_result -eq 0 ]; then
    print_message "$GREEN" "✅ ALL TESTS PASSED! EntityDB handles large files correctly."
  else
    print_message "$RED" "❌ SOME TESTS FAILED! EntityDB has issues with large files."
  fi
  
  print_message "$BLUE" "========================================"
}

# Run the tests
main