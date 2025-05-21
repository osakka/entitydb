#!/bin/bash
# Benchmark script for testing the chunking improvements with high performance mode

# ANSI color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB Chunking Benchmark${NC}"
echo -e "${BLUE}========================================${NC}"

# File sizes to test (in KB)
FILE_SIZES=(100 500 1000)
ITERATIONS=3

# Create temp directory for test files
mkdir -p /tmp/entitydb_benchmark

# Function to create a test file of specified size
create_test_file() {
    local size=$1
    local filename="/tmp/entitydb_benchmark/test_${size}KB.bin"
    
    # Create file with predictable content for verification
    echo "START_MARKER_${size}KB" > "$filename"
    dd if=/dev/urandom bs=1K count=$size 2>/dev/null >> "$filename"
    echo "END_MARKER_${size}KB" >> "$filename"
    
    echo "$filename"
}

# Function to create an entity with a test file
create_entity() {
    local filename=$1
    local size=$2
    
    # Convert the file to base64
    local base64_data=$(base64 -w 0 < "$filename")
    
    # Create a small JSON file for the request
    local json_file="/tmp/entitydb_benchmark/request_${size}KB.json"
    cat > "$json_file" << EOF
{
  "tags": ["type:benchmark", "size:${size}KB", "test:chunking"],
  "content": "$base64_data"
}
EOF
    
    # Create entity and extract ID
    local response=$(curl -s -X POST "http://localhost:8085/api/v1/test/entities/create" \
        -H "Content-Type: application/json" \
        -d @"$json_file")
    
    # Extract entity ID
    local entity_id=$(echo "$response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    
    # Clean up request file
    rm -f "$json_file"
    
    echo "$entity_id"
}

# Function to retrieve and verify entity
retrieve_entity() {
    local entity_id=$1
    local expected_size=$2
    local output_file="/tmp/entitydb_benchmark/retrieved_${expected_size}KB.bin"
    
    # Retrieve entity with raw content
    curl -s "http://localhost:8085/api/v1/entities/get?id=${entity_id}&include_content=true&raw=true" \
        -o "$output_file"
    
    # Get file size in bytes
    local actual_size=$(stat -c%s "$output_file")
    
    # Check file size (approximate due to markers)
    local expected_bytes=$((expected_size * 1024 + 50))  # Add some bytes for markers
    local size_diff=$((expected_bytes - actual_size))
    local abs_diff=${size_diff#-}  # Absolute value
    
    # Check if size is within 5% tolerance
    local tolerance=$((expected_bytes / 20))
    
    if [ $abs_diff -le $tolerance ]; then
        # Check for markers
        if grep -q "START_MARKER_${expected_size}KB" "$output_file" && \
           grep -q "END_MARKER_${expected_size}KB" "$output_file"; then
            echo "success"
        else
            echo "marker_missing"
        fi
    else
        echo "size_mismatch:${actual_size}/${expected_bytes}"
    fi
}

echo -e "${BLUE}Running benchmark on ${#FILE_SIZES[@]} file sizes...${NC}"
echo -e "${BLUE}----------------------------------------${NC}"

# Table header
printf "| %-10s | %-12s | %-18s | %-12s | %-15s |\n" "Size (KB)" "Create (sec)" "Retrieve (sec)" "Status" "Throughput (KB/s)"
printf "|------------|--------------|-------------------|--------------|----------------|\n"

# Run benchmarks for each file size
for size in "${FILE_SIZES[@]}"; do
    echo -e "${BLUE}Testing with ${size}KB file...${NC}"
    
    # Create test file
    test_file=$(create_test_file $size)
    
    # Track times and results
    total_create_time=0
    total_retrieve_time=0
    success_count=0
    
    for i in $(seq 1 $ITERATIONS); do
        echo -e "${BLUE}  Iteration $i/${ITERATIONS}${NC}"
        
        # Time entity creation
        start_time=$(date +%s.%N)
        entity_id=$(create_entity "$test_file" "$size")
        end_time=$(date +%s.%N)
        create_time=$(echo "$end_time - $start_time" | bc)
        total_create_time=$(echo "$total_create_time + $create_time" | bc)
        
        if [ -z "$entity_id" ]; then
            echo -e "${RED}  Failed to create entity!${NC}"
            continue
        fi
        
        echo -e "${GREEN}  Created entity: $entity_id in ${create_time}s${NC}"
        
        # Time entity retrieval
        start_time=$(date +%s.%N)
        result=$(retrieve_entity "$entity_id" "$size")
        end_time=$(date +%s.%N)
        retrieve_time=$(echo "$end_time - $start_time" | bc)
        total_retrieve_time=$(echo "$total_retrieve_time + $retrieve_time" | bc)
        
        if [ "$result" = "success" ]; then
            echo -e "${GREEN}  Successfully retrieved and verified content in ${retrieve_time}s${NC}"
            ((success_count++))
        else
            echo -e "${RED}  Failed to verify content: $result${NC}"
        fi
    done
    
    # Calculate averages
    avg_create_time=$(echo "scale=3; $total_create_time / $ITERATIONS" | bc)
    avg_retrieve_time=$(echo "scale=3; $total_retrieve_time / $ITERATIONS" | bc)
    
    # Calculate throughput for retrieval (KB/s)
    throughput=$(echo "scale=2; $size / $avg_retrieve_time" | bc)
    
    # Status
    if [ $success_count -eq $ITERATIONS ]; then
        status="${GREEN}PASS${NC}"
    else
        status="${RED}FAIL ($success_count/$ITERATIONS)${NC}"
    fi
    
    # Print result row
    printf "| %-10d | %-12.3f | %-18.3f | %-12s | %-15.2f |\n" "$size" "$avg_create_time" "$avg_retrieve_time" "$status" "$throughput"
done

echo -e "${BLUE}----------------------------------------${NC}"
echo -e "${GREEN}Benchmark completed!${NC}"

# Clean up
rm -rf /tmp/entitydb_benchmark

exit 0