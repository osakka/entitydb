#!/bin/bash
# Add logger import to files that use it

cd /opt/entitydb/src

# Add logger import to files that need it
files=("storage/binary/entity_repository.go" "api/dashboard_handler.go" "api/entity_relationship_handler.go" "api/router.go")

for file in "${files[@]}"; do
    # Check if logger is already imported
    if ! grep -q "entitydb/logger" "$file"; then
        # Add logger import after the first import statement
        sed -i '0,/^import (/,/^)/ {
            /^)/ i\	"entitydb/logger"
        }' "$file"
        echo "Added logger import to: $file"
    fi
done

echo "Logger imports added"