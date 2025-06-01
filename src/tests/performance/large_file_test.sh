#!/bin/bash
# Simple EntityDB Large File Test

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8085"
START_SIZE_MB=1          # Starting file size in MB
MAX_SIZE_MB=8            # Maximum file size to test in MB
SIZE_INCREMENT_MB=3      # Size increment in MB
TEMP_DIR="/tmp/entitydb_simple_large_tests"
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
  
  local response=$(curl -s -X POST "$SERVER_URL/api/v1/auth/login" \
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
  
  local file_size=$(du -m "$output_file" | cut -f1)
  print_message "$GREEN" "✅ Generated $file_size MB test file"
}

# Create entity with file content - using base64 file
create_entity_with_file() {
  local file_path=$1
  local size_mb=$2
  
  print_message "$BLUE" "Creating entity with $size_mb MB file..."
  
  # Base64 encode the file to a temporary file to avoid command line limits
  local b64_file="${file_path}.b64"
  base64 "$file_path" > "$b64_file"
  
  # Create a temporary JSON request file
  local json_file="${file_path}.json"
  echo "{\"tags\": [\"type:large_file_test\", \"size:${size_mb}MB\"], \"content\": {\"data\": \"$(cat $b64_file)\", \"type\": \"application/octet-stream\"}}" > "$json_file"
  
  # Create the entity using the JSON file
  local response=$(curl -s -X POST "$SERVER_URL/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d @"$json_file")
  
  # Extract entity ID
  local entity_id=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
  
  # Clean up temporary files
  rm -f "$b64_file" "$json_file"
  
  if [ -z "$entity_id" ]; then
    print_message "$RED" "❌ Failed to create entity with $size_mb MB file."
    print_message "$RED" "Response: $response"
    return 1
  else
    print_message "$GREEN" "✅ Created entity with ID: $entity_id"
    echo "$entity_id" > "${file_path}.id"
    return 0
  fi
}

# Retrieve entity to verify
verify_entity() {
  local file_path=$1
  local size_mb=$2
  local entity_id=$(cat "${file_path}.id")
  
  print_message "$BLUE" "Verifying $size_mb MB entity (ID: $entity_id)..."
  
  # Get the entity
  local response=$(curl -s -X GET "$SERVER_URL/api/v1/entities/get?id=$entity_id" \
    -H "Authorization: Bearer $TOKEN")
  
  # Check if we got the entity
  if [[ "$response" == *"\"id\":"* ]]; then
    print_message "$GREEN" "✅ Successfully retrieved entity of $size_mb MB"
    return 0
  else
    print_message "$RED" "❌ Failed to retrieve entity"
    print_message "$RED" "Response: $response"
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
    
    # Verify entity
    verify_entity "$test_file" "$size_mb"
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

# Main function
main() {
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "EntityDB Simple Large File Test"
  print_message "$BLUE" "========================================"
  
  # Setup
  cleanup
  login
  
  # Run incremental size tests
  test_incremental_sizes
  size_test_result=$?
  
  # Final result
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "FINAL RESULT:"
  
  if [ $size_test_result -eq 0 ]; then
    print_message "$GREEN" "✅ ALL TESTS PASSED! EntityDB handles large files correctly."
  else
    print_message "$RED" "❌ SOME TESTS FAILED! EntityDB has issues with large files."
  fi
  
  print_message "$BLUE" "========================================"
  
  return $size_test_result
}

# Run the tests
main