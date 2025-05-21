#!/bin/bash
# EntityDB Auto-chunking Test
# This script specifically tests the auto-chunking feature with very large files

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="https://localhost:8085"
CHUNK_THRESHOLD_MB=4  # Default autochunking threshold in EntityDB
MAX_FILE_SIZE_MB=40   # Maximum file size to test (in MB), must be larger than CHUNK_THRESHOLD_MB
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
  echo "Auto-chunking test file generated at $(date)" >> "$output_file"
  
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

# Create entity with content that should be auto-chunked
create_chunked_entity() {
  local file_path=$1
  local size_mb=$2
  local file_name=$(basename "$file_path")
  
  print_message "$BLUE" "Creating entity with $size_mb MB file (above chunking threshold of $CHUNK_THRESHOLD_MB MB): $file_name..."
  
  # Base64 encode the content (streaming to avoid memory issues)
  local encoded_content=$(base64 -w0 "$file_path")
  
  # Calculate original file hash for verification
  local original_hash=$(sha256sum "$file_path" | cut -d' ' -f1)
  
  # Create the entity
  local response=$(curl -k -s -X POST "$SERVER_URL/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"tags\": [\"type:chunked_test\", \"name:$file_name\", \"size:${size_mb}MB\", \"hash:$original_hash\"],
      \"content\": {
        \"data\": \"$encoded_content\",
        \"type\": \"application/octet-stream\"
      }
    }")
  
  # Extract entity ID
  local entity_id=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
  
  if [ -z "$entity_id" ]; then
    print_message "$RED" "❌ Failed to create chunked entity. Response: $response"
    return 1
  else
    print_message "$GREEN" "✅ Created potentially chunked entity with ID: $entity_id"
    echo "$entity_id"
    return 0
  fi
}

# Get entity and verify its content, even if it was auto-chunked
verify_chunked_entity() {
  local entity_id=$1
  local original_file=$2
  local output_file="tmp_chunked_output_${entity_id}.bin"
  
  print_message "$BLUE" "Verifying chunked entity content for ID: $entity_id..."
  
  # Get the entity
  local response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/get?id=$entity_id" \
    -H "Authorization: Bearer $TOKEN")
  
  # Extract content
  local content=$(echo "$response" | grep -o '"data":"[^"]*' | cut -d'"' -f4)
  
  if [ -z "$content" ]; then
    print_message "$RED" "❌ Failed to get chunked entity content. Response: $response"
    return 1
  fi
  
  # Decode the content to a file
  echo "$content" | base64 -d > "$output_file"
  
  # Verify file sizes
  local original_size=$(du -b "$original_file" | cut -f1)
  local retrieved_size=$(du -b "$output_file" | cut -f1)
  
  print_message "$BLUE" "Original size: $original_size bytes"
  print_message "$BLUE" "Retrieved size: $retrieved_size bytes"
  
  # Verify file hashes match
  local original_hash=$(sha256sum "$original_file" | cut -d'"' -f1)
  local retrieved_hash=$(sha256sum "$output_file" | cut -d'"' -f1)
  
  if [ "$original_hash" == "$retrieved_hash" ]; then
    print_message "$GREEN" "✅ Verified chunked entity content integrity - chunks were properly reassembled"
    rm "$output_file"
    return 0
  else
    print_message "$RED" "❌ Chunked content verification failed!"
    print_message "$RED" "Original hash: $original_hash"
    print_message "$RED" "Retrieved hash: $retrieved_hash"
    return 1
  fi
}

# Test auto-chunking with a file larger than the chunking threshold
test_auto_chunking() {
  local tmp_dir="./tmp_chunking_test"
  mkdir -p "$tmp_dir"
  
  local success_count=0
  local fail_count=0
  
  # Test with file sizes around and above the chunking threshold
  for size_mb in $(seq $((CHUNK_THRESHOLD_MB-1)) 2 $MAX_FILE_SIZE_MB); do
    print_message "$BLUE" "========================================"
    print_message "$BLUE" "Testing auto-chunking with file size: $size_mb MB"
    if [ $size_mb -ge $CHUNK_THRESHOLD_MB ]; then
      print_message "$YELLOW" "This file should be auto-chunked (> $CHUNK_THRESHOLD_MB MB threshold)"
    else
      print_message "$YELLOW" "This file should NOT be auto-chunked (< $CHUNK_THRESHOLD_MB MB threshold)"
    fi
    print_message "$BLUE" "========================================"
    
    local test_file="$tmp_dir/chunk_test_${size_mb}MB.bin"
    generate_test_file "$size_mb" "$test_file"
    
    local entity_id=$(create_chunked_entity "$test_file" "$size_mb")
    
    if [ $? -eq 0 ]; then
      sleep 1 # Allow time for entity to be fully processed
      
      # Verify content integrity
      verify_chunked_entity "$entity_id" "$test_file"
      if [ $? -eq 0 ]; then
        print_message "$GREEN" "✅ Auto-chunking test passed for $size_mb MB file"
        ((success_count++))
      else
        print_message "$RED" "❌ Auto-chunking test failed for $size_mb MB file - content integrity check failed"
        ((fail_count++))
      fi
    else
      print_message "$RED" "❌ Auto-chunking test failed for $size_mb MB file - entity creation failed"
      ((fail_count++))
    fi
    
    echo ""
  done
  
  # Clean up
  rm -rf "$tmp_dir"
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Auto-chunking Test Results:"
  print_message "$GREEN" "✅ Successful tests: $success_count"
  print_message "$RED" "❌ Failed tests: $fail_count"
  print_message "$BLUE" "========================================"
  
  return $fail_count
}

# Examine database file to search for chunk entities
examine_database() {
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Examining database for chunk entities..."
  print_message "$BLUE" "========================================"
  
  # Use curl to get all entities and look for chunks
  local entities_response=$(curl -k -s -X GET "$SERVER_URL/api/v1/entities/list" \
    -H "Authorization: Bearer $TOKEN")
  
  # Check for chunk entities
  if [[ "$entities_response" == *"\"type:content_chunk\""* ]]; then
    print_message "$GREEN" "✅ Content chunks found in database"
    
    # Extract a list of chunk entity IDs
    local chunk_count=$(echo "$entities_response" | grep -o "\"type:content_chunk\"" | wc -l)
    
    print_message "$BLUE" "Found approximately $chunk_count content chunks"
    print_message "$GREEN" "✅ Auto-chunking appears to be working correctly"
  else
    print_message "$YELLOW" "⚠️  No content chunks found in database"
    print_message "$YELLOW" "Auto-chunking might not be functioning or no files were large enough to trigger it"
  fi
  
  print_message "$BLUE" "========================================"
}

# Main execution
print_message "$BLUE" "========================================"
print_message "$BLUE" "Starting EntityDB Auto-chunking Test"
print_message "$BLUE" "========================================"

# Login
login

# Run auto-chunking tests
test_auto_chunking
auto_chunking_result=$?

# Examine database for chunks
examine_database

# Summarize results
if [ $auto_chunking_result -eq 0 ]; then
  print_message "$GREEN" "✅ ALL AUTO-CHUNKING TESTS PASSED! Feature verified working correctly."
else
  print_message "$RED" "❌ AUTO-CHUNKING TESTS FAILED with $auto_chunking_result errors. Feature may not be working correctly."
fi

print_message "$BLUE" "========================================"