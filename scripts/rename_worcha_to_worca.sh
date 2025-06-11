#!/bin/bash
# Rename all occurrences of worcha to worca

echo "=== Renaming worcha to worca ==="
echo

# Files to update
FILES_TO_UPDATE=(
    # Test files
    "src/tests/test_query_debug.sh"
    "src/tests/test_dataset_final.sh"
    "src/tests/debug_dataset.sh"
    "src/tests/test_dataset.sh"
    "src/tests/test_dataset_isolation.sh"
    "src/tests/test_dataset_performance.sh"
    
    # Documentation
    "docs/performance/PERFORMANCE_SOLUTIONS_SUMMARY.md"
    "docs/implementation/MULTI_HUB_ARCHITECTURE.md"
    "docs/implementation/DATASPACE_IMPLEMENTATION_STATUS.md"
    "docs/implementation/DATASPACE_COMPLETE_SUMMARY.md"
    "docs/architecture/DATASPACE_ARCHITECTURE_VISION.md"
)

# Perform replacements
for file in "${FILES_TO_UPDATE[@]}"; do
    if [ -f "$file" ]; then
        echo "Updating $file..."
        # Replace worcha with worca (case-sensitive)
        sed -i 's/worcha/worca/g' "$file"
        # Replace Worcha with Worca (capitalized)
        sed -i 's/Worcha/Worca/g' "$file"
        # Replace WORCHA with WORCA (uppercase)
        sed -i 's/WORCHA/WORCA/g' "$file"
    else
        echo "Warning: $file not found"
    fi
done

echo
echo "Checking for any remaining occurrences..."
echo

# Check for any remaining occurrences
echo "In src/tests/:"
grep -r "worcha" src/tests/ 2>/dev/null | grep -v Binary | grep -v ".git" || echo "No occurrences found"

echo
echo "In docs/:"
grep -r "worcha" docs/ 2>/dev/null | grep -v Binary | grep -v ".git" || echo "No occurrences found"

echo
echo "=== Rename complete ==="