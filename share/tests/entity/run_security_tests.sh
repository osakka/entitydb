#!/bin/bash
#
# Run all security tests for the EntityDB platform
#

echo "============================================================"
echo "EntityDB Security Implementation Test Suite"
echo "============================================================"
echo

# Create necessary directories
echo "Setting up test environment..."
mkdir -p /opt/entitydb/var/log/audit
echo "Created audit log directory"
echo

# Check if server is running
echo "Checking if EntityDB server is running..."
SERVER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8085/api/v1/status)

if [ "$SERVER_STATUS" != "200" ]; then
    echo "ERROR: EntityDB server is not running on localhost:8085"
    echo "Please start the server with: ./bin/entitydbd.sh start"
    exit 1
fi
echo "Server is running"
echo

# Run each test with summary reporting
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=4
TEST_RESULTS=()

echo "============================================================"
echo "RUNNING: Secure Password Test"
echo "============================================================"
if /opt/entitydb/share/tests/entity/test_secure_password.sh; then
    TEST_RESULTS+=("✅ Secure Password Test: PASSED")
    TESTS_PASSED=$((TESTS_PASSED+1))
else
    TEST_RESULTS+=("❌ Secure Password Test: FAILED")
    TESTS_FAILED=$((TESTS_FAILED+1))
fi
echo

echo "============================================================"
echo "RUNNING: RBAC Test"
echo "============================================================"
if /opt/entitydb/share/tests/entity/test_rbac_entity.sh; then
    TEST_RESULTS+=("✅ RBAC Test: PASSED")
    TESTS_PASSED=$((TESTS_PASSED+1))
else
    TEST_RESULTS+=("❌ RBAC Test: FAILED")
    TESTS_FAILED=$((TESTS_FAILED+1))
fi
echo

echo "============================================================"
echo "RUNNING: Audit Logging Test"
echo "============================================================"
if /opt/entitydb/share/tests/entity/test_audit_logging.sh; then
    TEST_RESULTS+=("✅ Audit Logging Test: PASSED")
    TESTS_PASSED=$((TESTS_PASSED+1))
else
    TEST_RESULTS+=("❌ Audit Logging Test: FAILED")
    TESTS_FAILED=$((TESTS_FAILED+1))
fi
echo

echo "============================================================"
echo "RUNNING: Input Validation Test"
echo "============================================================"
if /opt/entitydb/share/tests/entity/test_input_validation.sh; then
    TEST_RESULTS+=("✅ Input Validation Test: PASSED")
    TESTS_PASSED=$((TESTS_PASSED+1))
else
    TEST_RESULTS+=("❌ Input Validation Test: FAILED")
    TESTS_FAILED=$((TESTS_FAILED+1))
fi
echo

# Output summary
echo "============================================================"
echo "SECURITY TESTS SUMMARY"
echo "============================================================"
echo "Tests passed: $TESTS_PASSED/$TESTS_TOTAL"
echo "Tests failed: $TESTS_FAILED/$TESTS_TOTAL"
echo

echo "Detailed Results:"
for result in "${TEST_RESULTS[@]}"; do
    echo "$result"
done
echo

if [ "$TESTS_FAILED" -eq 0 ]; then
    echo "All security tests passed!"
    echo "The EntityDB security implementation is functioning correctly."
    exit 0
else
    echo "Some security tests failed!"
    echo "Please check the test output for details."
    exit 1
fi