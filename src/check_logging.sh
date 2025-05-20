#!/bin/bash

# Check for log. usage in Go files
echo "Files still using log.* instead of logger.*:"
find . -name "*.go" -not -path "./logger/*" -exec grep -l "log\.\(Print\|Fatal\|Panic\)" {} \; 2>/dev/null | while read file; do
    echo -e "\n$file:"
    grep -n "log\.\(Print\|Fatal\|Panic\)" "$file"
done

echo -e "\n\nFiles already using logger.*:"
find . -name "*.go" -not -path "./logger/*" -exec grep -l "logger\." {} \;

echo -e "\n\nFiles without any logging:"
find . -name "*.go" -not -path "./logger/*" | while read file; do
    grep -q "log\.\|logger\." "$file" || echo "$file"
done