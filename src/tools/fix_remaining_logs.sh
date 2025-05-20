#!/bin/bash
# Fix remaining log statements

cd /opt/entitydb/src

# Update all remaining log.Printf statements
for file in storage/binary/*.go api/*.go; do
    if [ -f "$file" ]; then
        # Replace WARNING logs
        sed -i 's/log\.Printf("WARNING: \([^"]*\)"/logger.Warn("\1"/g' "$file"
        
        # Replace ERROR logs
        sed -i 's/log\.Printf("Failed \([^"]*\)"/logger.Error("Failed \1"/g' "$file"
        
        # Replace generic logs to debug
        sed -i 's/log\.Printf("\([^"]*\)"/logger.Debug("\1"/g' "$file"
        
        # Replace log.Fatal statements
        sed -i 's/log\.Fatalf\("/logger.Error("/g' "$file"
        
        echo "Fixed: $file"
    fi
done

# Remove unused log imports
for file in storage/binary/*.go api/*.go; do
    if [ -f "$file" ]; then
        # Check if log is still used
        if ! grep -q "log\." "$file"; then
            # Remove log import
            sed -i '/"log"/d' "$file"
            echo "Removed log import from: $file"
        fi
    fi
done

echo "All log statements fixed"