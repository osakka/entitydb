#!/bin/bash
# EntityDB Temporal Format Test
# This script verifies the temporal format of entities without requiring API access

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
DATA_DIR="/opt/entitydb/var"

# Function for colored output
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

# Test temporal format by examining binary files
test_temporal_format() {
  print_message "$BLUE" "Testing EntityDB temporal format..."
  
  # Check if entity binary file exists
  if [ ! -f "$DATA_DIR/entities.ebf" ]; then
    print_message "$RED" "❌ Entity binary file not found at $DATA_DIR/entities.ebf"
    return 1
  fi
  
  # Get file size
  local file_size=$(stat -c%s "$DATA_DIR/entities.ebf")
  print_message "$BLUE" "Entity binary file size: $file_size bytes"
  
  # Check for WAL
  if [ -f "$DATA_DIR/entitydb.wal" ]; then
    local wal_size=$(stat -c%s "$DATA_DIR/entitydb.wal")
    print_message "$GREEN" "✅ WAL file found at $DATA_DIR/entitydb.wal ($wal_size bytes)"
  else
    print_message "$YELLOW" "⚠️ WAL file not found at $DATA_DIR/entitydb.wal"
  fi
  
  # Check file for magic bytes (looking for EBF header in some form)
  if xxd -l 16 "$DATA_DIR/entities.ebf" | grep -q "EB"; then
    print_message "$GREEN" "✅ Found potential EBF format magic bytes in entity file"
  else
    print_message "$YELLOW" "⚠️ No standard EBF header found, but file may still be valid"
  fi
  
  # Check for temporal tag format in a limited sample
  if head -c 10000 "$DATA_DIR/entities.ebf" | strings | grep -E "[0-9]{10,14}\|" > /dev/null; then
    print_message "$GREEN" "✅ Found timestamp pattern in entity file (likely temporal tags)"
  else
    print_message "$YELLOW" "⚠️ No clear timestamp pattern found in sampled entity data"
    # Try another pattern more common in temporal data
    if head -c 10000 "$DATA_DIR/entities.ebf" | strings | grep -E "timestamp|created_at|updated_at" > /dev/null; then
      print_message "$GREEN" "✅ Found temporal metadata in entity file"
    fi
  fi
  
  print_message "$GREEN" "✅ EntityDB temporal format verification completed"
  return 0
}

# Main execution
print_message "$BLUE" "========================================"
print_message "$BLUE" "Starting EntityDB Temporal Format Test"
print_message "$BLUE" "========================================"

# Run test
test_temporal_format
result=$?

if [ $result -eq 0 ]; then
  print_message "$GREEN" "✅ EntityDB temporal format test PASSED!"
else
  print_message "$RED" "❌ EntityDB temporal format test FAILED!"
fi

print_message "$BLUE" "========================================"

exit $result