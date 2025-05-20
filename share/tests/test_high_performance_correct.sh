#!/bin/bash
set -e

# Test to debug turbo repository issues

echo "Testing turbo repository issues..."

# Function to test admin user
test_turbo() {
    echo -e "\n=== Testing Turbo Repository ==="
    
    # Check database files
    echo "Database files:"
    ls -la /opt/entitydb/var/
    
    # Check entity.db file
    if [ -f /opt/entitydb/var/entities.ebf ]; then
        echo -e "\nentities.ebf size: $(stat -c%s /opt/entitydb/var/entities.ebf)"
        
        # Dump first 256 bytes
        echo -e "\nFirst 256 bytes of entities.ebf:"
        hexdump -C /opt/entitydb/var/entities.ebf | head -n 16
    else
        echo "ERROR: entities.ebf not found"
    fi
    
    # Check if the service is running
    echo -e "\nService status:"
    ps aux | grep entitydb | grep -v grep || true
    
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