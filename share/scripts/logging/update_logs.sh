#!/bin/bash
# Script to update log statements to use the new logger

cd /opt/entitydb/src

# Add logger import to all Go files that have log statements
for file in $(find . -name "*.go" -exec grep -l "log\.Printf.*DEBUG\|log\.Printf.*ERROR\|log\.Printf.*INFO" {} \;); do
    # Skip the logger package itself
    if [[ "$file" == "./logger/logger.go" ]]; then
        continue
    fi
    
    # Check if logger is already imported
    if ! grep -q "entitydb/logger" "$file"; then
        # Add logger import after the last import
        sed -i '/^import (/,/^)/ {
            /^)/ i\	"entitydb/logger"
        }' "$file"
    fi
    
    # Replace log statements
    sed -i 's/log\.Printf("DEBUG: \([^"]*\)"/logger.Debug("\1"/g' "$file"
    sed -i 's/log\.Printf("ERROR: \([^"]*\)"/logger.Error("\1"/g' "$file"
    sed -i 's/log\.Printf("INFO: \([^"]*\)"/logger.Info("\1"/g' "$file"
    sed -i 's/log\.Printf("WARN: \([^"]*\)"/logger.Warn("\1"/g' "$file"
    
    # Replace log.Printf with format markers
    sed -i 's/log\.Printf("\[EntityDB\] .*DEBUG: \([^"]*\)"/logger.Debug("\1"/g' "$file"
    sed -i 's/log\.Printf("\[EntityDB\] .*ERROR: \([^"]*\)"/logger.Error("\1"/g' "$file"
    sed -i 's/log\.Printf("\[EntityDB\] .*INFO: \([^"]*\)"/logger.Info("\1"/g' "$file"
    
    echo "Updated: $file"
done

echo "Log statements updated to use the new logger"