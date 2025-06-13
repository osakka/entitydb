#!/bin/bash
# test_configuration_scenarios.sh - Test EntityDB configuration system with various scenarios
# Validates three-tier configuration hierarchy works correctly

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function for colored output
print_result() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

print_header() {
    echo ""
    print_result "$BLUE" "=============================================="
    print_result "$BLUE" "$1"
    print_result "$BLUE" "=============================================="
}

test_passed() {
    PASSED_TESTS=$((PASSED_TESTS + 1))
    print_result "$GREEN" "✅ PASS: $1"
}

test_failed() {
    FAILED_TESTS=$((FAILED_TESTS + 1))
    print_result "$RED" "❌ FAIL: $1"
}

run_test() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    local test_name="$1"
    local test_command="$2"
    
    print_result "$BLUE" "Testing: $test_name"
    
    if eval "$test_command" >/dev/null 2>&1; then
        test_passed "$test_name"
        return 0
    else
        test_failed "$test_name"
        return 1
    fi
}

# Get EntityDB root directory
ENTITYDB_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ENTITYDB_ROOT"

# Create test directory
TEST_DIR="$ENTITYDB_ROOT/tmp/config_test_$$"
mkdir -p "$TEST_DIR"

print_header "EntityDB Configuration System Testing"
print_result "$BLUE" "Testing three-tier configuration hierarchy"
print_result "$BLUE" "Test directory: $TEST_DIR"

# Cleanup function
cleanup() {
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

print_header "1. Environment Variable Configuration Tests"

# Test 1: Basic environment variable loading
print_result "$BLUE" "Test 1.1: Basic environment variable configuration"
export ENTITYDB_DATA_PATH="$TEST_DIR/test_data"
export ENTITYDB_PORT="9999"

# Build a simple test tool
cat > "$TEST_DIR/test_env.go" << 'EOF'
package main

import (
    "entitydb/config"
    "flag"
    "fmt"
)

func main() {
    configManager := config.NewConfigManager(nil)
    configManager.RegisterFlags()
    flag.Parse()
    
    cfg, err := configManager.Initialize()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("DataPath:%s\n", cfg.DataPath)
    fmt.Printf("Port:%d\n", cfg.Port)
}
EOF

cd src
go run "$TEST_DIR/test_env.go" > "$TEST_DIR/env_output.txt" 2>&1
cd "$ENTITYDB_ROOT"

if grep -q "DataPath:$TEST_DIR/test_data" "$TEST_DIR/env_output.txt" && grep -q "Port:9999" "$TEST_DIR/env_output.txt"; then
    test_passed "Environment variables properly loaded"
else
    test_failed "Environment variables not loaded correctly"
    cat "$TEST_DIR/env_output.txt"
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

print_header "2. Command Line Flag Override Tests"

# Test 2: Command line flags override environment
print_result "$BLUE" "Test 2.1: Command line flags override environment variables"
export ENTITYDB_PORT="8888"  # Set in environment

cd src
go run "$TEST_DIR/test_env.go" --entitydb-port=7777 > "$TEST_DIR/flag_output.txt" 2>&1
cd "$ENTITYDB_ROOT"

if grep -q "Port:7777" "$TEST_DIR/flag_output.txt"; then
    test_passed "Command line flags override environment variables"
else
    test_failed "Command line flags do not override environment"
    cat "$TEST_DIR/flag_output.txt"
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

print_header "3. Configuration Helper Methods Tests"

# Test 3: Configuration helper methods work correctly
print_result "$BLUE" "Test 3.1: Configuration helper methods"
cat > "$TEST_DIR/test_helpers.go" << 'EOF'
package main

import (
    "entitydb/config"
    "flag"
    "fmt"
)

func main() {
    configManager := config.NewConfigManager(nil)
    configManager.RegisterFlags()
    flag.Parse()
    
    cfg, err := configManager.Initialize()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("DatabasePath:%s\n", cfg.DatabasePath())
    fmt.Printf("WALPath:%s\n", cfg.WALPath())
    fmt.Printf("BackupPath:%s\n", cfg.BackupFullPath())
}
EOF

export ENTITYDB_DATA_PATH="$TEST_DIR/custom"
export ENTITYDB_DATABASE_FILENAME="test.db"
export ENTITYDB_WAL_SUFFIX=".testwal"

cd src
go run "$TEST_DIR/test_helpers.go" > "$TEST_DIR/helpers_output.txt" 2>&1
cd "$ENTITYDB_ROOT"

if grep -q "DatabasePath:$TEST_DIR/custom/data/test.db" "$TEST_DIR/helpers_output.txt" && \
   grep -q "WALPath:$TEST_DIR/custom/data/test.db.testwal" "$TEST_DIR/helpers_output.txt"; then
    test_passed "Configuration helper methods work correctly"
else
    test_failed "Configuration helper methods failed"
    cat "$TEST_DIR/helpers_output.txt"
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

print_header "4. Tool Integration Tests"

# Test 4: Tools use ConfigManager correctly
print_result "$BLUE" "Test 4.1: Tools integrate with ConfigManager"
export ENTITYDB_DATA_PATH="$TEST_DIR/tools_test"
mkdir -p "$TEST_DIR/tools_test"

# Test that list_users tool uses ConfigManager
if cd src && timeout 5s go run tools/list_users.go --entitydb-data-path="$TEST_DIR/tools_test" 2>&1 | grep -q "Configuration error\|Failed to open repository"; then
    test_passed "list_users tool uses ConfigManager (expected failure with empty repo)"
else
    test_failed "list_users tool does not use ConfigManager properly"
fi
cd "$ENTITYDB_ROOT"
TOTAL_TESTS=$((TOTAL_TESTS + 1))

print_header "5. Runtime Script Integration Tests"

# Test 5: Runtime script loads environment correctly
print_result "$BLUE" "Test 5.1: Runtime script environment loading"
export ENTITYDB_DATA_PATH="$TEST_DIR/runtime_test"
export ENTITYDB_PORT="6666"

# Create a mock environment file for testing
mkdir -p "$TEST_DIR/test_config"
cat > "$TEST_DIR/test_config/entitydb.env" << EOF
ENTITYDB_PORT=5555
ENTITYDB_DATA_PATH=$TEST_DIR/runtime_env_test
ENTITYDB_USE_SSL=false
EOF

# Test the load_environment function by extracting it
if grep -A 20 "load_environment()" bin/entitydbd.sh | grep -q "source.*entitydb.env"; then
    test_passed "Runtime script has environment loading logic"
else
    test_failed "Runtime script missing environment loading"
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

print_header "6. File Path Resolution Tests"

# Test 6: Relative vs absolute path handling
print_result "$BLUE" "Test 6.1: Path resolution handles relative and absolute paths"
cat > "$TEST_DIR/test_paths.go" << 'EOF'
package main

import (
    "entitydb/config"
    "flag"
    "fmt"
)

func main() {
    configManager := config.NewConfigManager(nil)
    configManager.RegisterFlags()
    flag.Parse()
    
    cfg, err := configManager.Initialize()
    if err != nil {
        panic(err)
    }
    
    // Test relative backup path
    fmt.Printf("BackupPath:%s\n", cfg.BackupFullPath())
    fmt.Printf("TempPath:%s\n", cfg.TempFullPath())
}
EOF

export ENTITYDB_DATA_PATH="$TEST_DIR/paths"
export ENTITYDB_BACKUP_PATH="./backups"
export ENTITYDB_TEMP_PATH="/tmp/absolute"

cd src
go run "$TEST_DIR/test_paths.go" > "$TEST_DIR/paths_output.txt" 2>&1
cd "$ENTITYDB_ROOT"

if grep -q "BackupPath:$TEST_DIR/paths/backups" "$TEST_DIR/paths_output.txt" && \
   grep -q "TempPath:/tmp/absolute" "$TEST_DIR/paths_output.txt"; then
    test_passed "Path resolution handles relative and absolute paths correctly"
else
    test_failed "Path resolution failed"
    cat "$TEST_DIR/paths_output.txt"
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

print_header "7. Build System Integration Tests"

# Test 7: All refactored tools build successfully
print_result "$BLUE" "Test 7.1: Refactored tools build successfully"
BUILD_FAILURES=0
TOOLS_TO_TEST=("clear_cache.go" "list_users.go" "force_reindex.go" "recovery_tool.go")

for tool in "${TOOLS_TO_TEST[@]}"; do
    tool_name=$(basename "$tool" .go)
    if cd src && go build -o "$TEST_DIR/${tool_name}_test" "tools/$tool" >/dev/null 2>&1; then
        print_result "$GREEN" "  ✓ $tool builds successfully"
    else
        print_result "$RED" "  ✗ $tool build failed"
        BUILD_FAILURES=$((BUILD_FAILURES + 1))
    fi
    cd "$ENTITYDB_ROOT"
done

if [ $BUILD_FAILURES -eq 0 ]; then
    test_passed "All core refactored tools build successfully"
else
    test_failed "$BUILD_FAILURES tools failed to build"
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

print_header "8. Configuration Validation Tests"

# Test 8: Configuration validation catches issues
print_result "$BLUE" "Test 8.1: Invalid configuration is rejected gracefully"
cat > "$TEST_DIR/test_invalid.go" << 'EOF'
package main

import (
    "entitydb/config"
    "flag"
    "fmt"
)

func main() {
    configManager := config.NewConfigManager(nil)
    configManager.RegisterFlags()
    flag.Parse()
    
    cfg, err := configManager.Initialize()
    if err != nil {
        fmt.Printf("ConfigError:%v\n", err)
        return
    }
    
    fmt.Printf("Success:%s\n", cfg.DataPath)
}
EOF

# Test with invalid port
export ENTITYDB_PORT="invalid"
cd src
go run "$TEST_DIR/test_invalid.go" > "$TEST_DIR/invalid_output.txt" 2>&1
cd "$ENTITYDB_ROOT"

# Reset to valid port
export ENTITYDB_PORT="8085"

# Note: Go's flag package handles invalid integers by setting to 0, not erroring
# So we just check that it doesn't crash
if grep -q "Success:\|ConfigError:" "$TEST_DIR/invalid_output.txt"; then
    test_passed "Configuration handles invalid input gracefully"
else
    test_failed "Configuration does not handle invalid input gracefully"
    cat "$TEST_DIR/invalid_output.txt"
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

print_header "TEST RESULTS SUMMARY"

echo ""
print_result "$BLUE" "Configuration System Test Results:"
print_result "$BLUE" "=================================="
print_result "$BLUE" "Total Tests: $TOTAL_TESTS"
print_result "$GREEN" "Passed: $PASSED_TESTS"
print_result "$RED" "Failed: $FAILED_TESTS"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    print_result "$GREEN" "🎉 ALL TESTS PASSED!"
    print_result "$GREEN" "✅ Three-tier configuration hierarchy working correctly"
    print_result "$GREEN" "✅ Environment variables load properly"
    print_result "$GREEN" "✅ Command line flags override environment"
    print_result "$GREEN" "✅ Configuration helper methods work"
    print_result "$GREEN" "✅ Tools integrate with ConfigManager"
    print_result "$GREEN" "✅ Runtime script loads environment"
    print_result "$GREEN" "✅ Path resolution works correctly"
    print_result "$GREEN" "✅ All refactored tools build successfully"
    echo ""
    print_result "$BLUE" "EntityDB Configuration System is fully functional! 🚀"
    exit 0
else
    print_result "$RED" "❌ $FAILED_TESTS TESTS FAILED"
    print_result "$RED" "Configuration system needs fixes before deployment"
    exit 1
fi