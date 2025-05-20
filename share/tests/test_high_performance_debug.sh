#!/bin/bash
set -e

# Test to debug turbo repository issues

echo "Testing turbo repository issues..."

# Function to test admin user
test_turbo() {
    echo -e "\n=== Testing Turbo Repository ==="
    
    # Check database files
    echo "Database files:"
    ls -la /opt/entitydb/var/db/binary/
    
    # Check entity.db file
    if [ -f /opt/entitydb/var/db/binary/entity.db ]; then
        echo "Entity.db size: $(stat -c%s /opt/entitydb/var/db/binary/entity.db)"
        
        # Dump first 256 bytes
        echo "First 256 bytes of entity.db:"
        hexdump -C /opt/entitydb/var/db/binary/entity.db | head -n 16
    else
        echo "ERROR: entity.db not found"
    fi
    
    # Check if the service is running
    echo -e "\nService status:"
    ps aux | grep entitydb || true
    
    # Try to log in
    echo -e "\nTrying to login with admin/admin..."
    # Use a timeout to prevent hanging
    timeout 5 curl -X POST http://localhost:8080/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"username": "admin", "password": "admin"}' \
        -v 2>&1 | grep -E "(< HTTP|token|error|ERROR|message)" || true
}

# Main test
test_turbo