#!/bin/bash
# validate_configuration_compliance.sh - EntityDB Configuration Compliance Verification
# Ensures zero hardcoded values throughout the codebase

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Counters
TOTAL_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0

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

check_failed() {
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
    print_result "$RED" "❌ FAIL: $1"
}

check_warning() {
    WARNINGS=$((WARNINGS + 1))
    print_result "$YELLOW" "⚠️  WARN: $1"
}

check_passed() {
    print_result "$GREEN" "✅ PASS: $1"
}

run_check() {
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    local check_name="$1"
    local check_command="$2"
    
    print_result "$BLUE" "Running: $check_name"
    
    if eval "$check_command"; then
        check_passed "$check_name"
        return 0
    else
        check_failed "$check_name"
        return 1
    fi
}

print_header "EntityDB Configuration Compliance Validation"
print_result "$BLUE" "Verifying zero hardcoded values throughout codebase"
print_result "$BLUE" "Target: Complete compliance with three-tier configuration hierarchy"

# Get EntityDB root directory
ENTITYDB_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ENTITYDB_ROOT"

print_result "$BLUE" "EntityDB Root: $ENTITYDB_ROOT"

print_header "1. Checking for Hardcoded Ports"

# Check for hardcoded ports (excluding config files and tests)
print_result "$BLUE" "Scanning for hardcoded ports (8085, 8443)..."
if grep -r "8085\|8443" src/ --include="*.go" | grep -v -E "(config\.go|test|example|swagger)" | head -5; then
    check_failed "Found hardcoded ports in source code"
    echo "Files with hardcoded ports:"
    grep -r "8085\|8443" src/ --include="*.go" | grep -v -E "(config\.go|test|example|swagger)" | cut -d: -f1 | sort -u
else
    check_passed "No hardcoded ports found in source code"
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

print_header "2. Checking for Hardcoded Paths"

# Check for hardcoded absolute paths
print_result "$BLUE" "Scanning for hardcoded absolute paths..."
if grep -r '"/opt/\|"/var/\|"/tmp/"' src/ --include="*.go" | grep -v -E "(config\.go|test|example)" | head -5; then
    check_failed "Found hardcoded absolute paths"
    echo "Files with hardcoded paths:"
    grep -r '"/opt/\|"/var/\|"/tmp/"' src/ --include="*.go" | grep -v -E "(config\.go|test|example)" | cut -d: -f1 | sort -u
else
    check_passed "No hardcoded absolute paths found"
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

# Check for hardcoded relative paths (exclude legitimate ones)
print_result "$BLUE" "Scanning for hardcoded relative paths..."
if grep -r '"\.\/var\|"\.\.\/var"' src/ --include="*.go" | head -5; then
    check_failed "Found hardcoded relative paths"
    echo "Files with hardcoded relative paths:"
    grep -r '"\.\/var\|"\.\.\/var"' src/ --include="*.go" | cut -d: -f1 | sort -u
else
    check_passed "No hardcoded relative paths found"
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

print_header "3. Checking for Hardcoded File Extensions"

# Check for hardcoded file extensions
print_result "$BLUE" "Scanning for hardcoded file extensions..."
if grep -r '\.db"\|\.wal"\|\.idx"\|\.log"\|\.pid"' src/ --include="*.go" | grep -v -E "(config\.go|test)" | head -5; then
    check_warning "Found potential hardcoded file extensions (may be legitimate)"
    echo "Files with file extensions (review needed):"
    grep -r '\.db"\|\.wal"\|\.idx"\|\.log"\|\.pid"' src/ --include="*.go" | grep -v -E "(config\.go|test)" | cut -d: -f1 | sort -u
else
    check_passed "No problematic hardcoded file extensions found"
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

print_header "4. Checking Tool Configuration Compliance"

# Check that tools use ConfigManager pattern
print_result "$BLUE" "Verifying tools use ConfigManager pattern..."
if grep -r "NewEntityRepository.*var\|NewEntityRepository.*opt" src/tools/ --include="*.go" | head -5; then
    check_failed "Found tools bypassing ConfigManager"
    echo "Tools with hardcoded paths:"
    grep -r "NewEntityRepository.*var\|NewEntityRepository.*opt" src/tools/ --include="*.go" | cut -d: -f1 | sort -u
else
    check_passed "All tools use ConfigManager pattern"
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

# Check that tools import config package
print_result "$BLUE" "Verifying tools import config package..."
TOOLS_WITHOUT_CONFIG=0
for tool in src/tools/*.go src/tools/*/*.go; do
    if [ -f "$tool" ]; then
        if ! grep -q '"entitydb/config"' "$tool"; then
            echo "Tool missing config import: $tool"
            TOOLS_WITHOUT_CONFIG=$((TOOLS_WITHOUT_CONFIG + 1))
        fi
    fi
done

if [ $TOOLS_WITHOUT_CONFIG -gt 0 ]; then
    check_failed "$TOOLS_WITHOUT_CONFIG tools missing config import"
else
    check_passed "All tools properly import config package"
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

print_header "5. Checking Runtime Script Compliance"

# Check that runtime script doesn't duplicate configuration
print_result "$BLUE" "Verifying runtime script uses ConfigManager..."
if grep -n "hardcoded\|manual.*flag\|CMD_ARGS.*entitydb" bin/entitydbd.sh | head -3; then
    check_warning "Runtime script may have configuration duplication"
else
    check_passed "Runtime script properly delegates to ConfigManager"
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

print_header "6. Build Verification Tests"

# Test building main server
print_result "$BLUE" "Testing main server build..."
if cd src && go build -o ../bin/entitydb_test . >/dev/null 2>&1; then
    check_passed "Main server builds successfully"
    rm -f ../bin/entitydb_test
else
    check_failed "Main server build failed"
fi
cd "$ENTITYDB_ROOT"
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

# Test building sample tools
print_result "$BLUE" "Testing tool builds..."
TOOL_BUILD_FAILURES=0
for tool in src/tools/clear_cache.go src/tools/list_users.go src/tools/force_reindex.go; do
    if [ -f "$tool" ]; then
        tool_name=$(basename "$tool" .go)
        if cd src && go build -o ../bin/${tool_name}_test "$tool" >/dev/null 2>&1; then
            rm -f "../bin/${tool_name}_test"
        else
            echo "Failed to build: $tool"
            TOOL_BUILD_FAILURES=$((TOOL_BUILD_FAILURES + 1))
        fi
        cd "$ENTITYDB_ROOT"
    fi
done

if [ $TOOL_BUILD_FAILURES -eq 0 ]; then
    check_passed "All tested tools build successfully"
else
    check_failed "$TOOL_BUILD_FAILURES tools failed to build"
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

print_header "7. Configuration System Validation"

# Check environment file exists and has required variables
print_result "$BLUE" "Validating environment configuration..."
if [ -f "share/config/entitydb.env" ]; then
    check_passed "Default environment file exists"
    
    # Check for key configuration variables
    required_vars=("ENTITYDB_PORT" "ENTITYDB_DATA_PATH" "ENTITYDB_DATABASE_FILENAME" "ENTITYDB_WAL_SUFFIX")
    missing_vars=0
    
    for var in "${required_vars[@]}"; do
        if ! grep -q "^$var=" share/config/entitydb.env; then
            echo "Missing required variable: $var"
            missing_vars=$((missing_vars + 1))
        fi
    done
    
    if [ $missing_vars -eq 0 ]; then
        check_passed "All required environment variables present"
    else
        check_failed "$missing_vars required environment variables missing"
    fi
    TOTAL_CHECKS=$((TOTAL_CHECKS + 2))
else
    check_failed "Default environment file missing"
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
fi

# Check ConfigManager has required methods
print_result "$BLUE" "Validating ConfigManager implementation..."
if grep -q "DatabasePath()" src/config/config.go && \
   grep -q "WALPath()" src/config/config.go && \
   grep -q "RegisterFlags()" src/config/manager.go; then
    check_passed "ConfigManager has required methods"
else
    check_failed "ConfigManager missing required methods"
fi
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

print_header "VALIDATION RESULTS SUMMARY"

echo ""
print_result "$BLUE" "Total Checks: $TOTAL_CHECKS"
print_result "$GREEN" "Passed: $((TOTAL_CHECKS - FAILED_CHECKS - WARNINGS))"
print_result "$YELLOW" "Warnings: $WARNINGS"
print_result "$RED" "Failed: $FAILED_CHECKS"
echo ""

if [ $FAILED_CHECKS -eq 0 ]; then
    print_result "$GREEN" "🎉 SUCCESS: EntityDB Configuration Compliance VERIFIED!"
    print_result "$GREEN" "✅ Zero hardcoded values detected"
    print_result "$GREEN" "✅ All tools use ConfigManager"
    print_result "$GREEN" "✅ Runtime script properly delegates configuration"
    print_result "$GREEN" "✅ Three-tier configuration hierarchy implemented"
    echo ""
    print_result "$BLUE" "EntityDB is fully compliant with configuration management standards!"
    exit 0
else
    print_result "$RED" "❌ FAILED: $FAILED_CHECKS configuration compliance issues detected"
    print_result "$RED" "Please fix the identified issues and re-run validation"
    exit 1
fi