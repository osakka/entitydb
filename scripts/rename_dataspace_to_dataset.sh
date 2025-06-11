#!/bin/bash
# Script to rename 'dataset' to 'dataset' throughout the codebase
# This script will:
# 1. Show all files that will be affected
# 2. Create a backup of files before changes
# 3. Perform the rename operation

set -e

echo "=== EntityDB: Dataset to Dataset Rename Script ==="
echo "This script will rename all occurrences of 'dataset' to 'dataset'"
echo "in the EntityDB codebase."
echo ""

# Check if we're in the right directory
if [ ! -f "src/main.go" ]; then
    echo "ERROR: This script must be run from the EntityDB root directory"
    exit 1
fi

# Function to create backup
backup_file() {
    local file=$1
    if [ ! -f "${file}.dataset_backup" ]; then
        cp "$file" "${file}.dataset_backup"
    fi
}

# Function to perform case-sensitive replacements
rename_in_file() {
    local file=$1
    echo "Processing: $file"
    
    # Create backup
    backup_file "$file"
    
    # Perform replacements (case-sensitive)
    # API endpoints
    sed -i 's|/datasets/|/datasets/|g' "$file"
    sed -i 's|/datasets"|/datasets"|g' "$file"
    sed -i 's|/datasets'\''|/datasets'\''|g' "$file"
    
    # Go code - types and functions
    sed -i 's/Dataset/Dataset/g' "$file"
    sed -i 's/dataset/dataset/g' "$file"
    
    # Environment variables
    sed -i 's/ENTITYDB_ENABLE_DATASET/ENTITYDB_ENABLE_DATASET/g' "$file"
    sed -i 's/ENTITYDB_DATASET_/ENTITYDB_DATASET_/g' "$file"
    
    # Special cases for proper nouns in comments/docs
    sed -i 's/Dataset isolation/Dataset isolation/g' "$file"
    sed -i 's/dataset-architecture/dataset-architecture/g' "$file"
}

# Step 1: Show affected files
echo "=== Step 1: Files that will be modified ==="
echo ""

echo "Go source files:"
find ./src -name "*.go" -type f | xargs grep -l "dataset" | sort || true
echo ""

echo "Documentation files:"
find ./docs -name "*.md" -type f | xargs grep -l "dataset" | sort || true
echo ""

echo "Frontend files:"
find ./share/htdocs -name "*.html" -o -name "*.js" | xargs grep -l "dataset" | grep -v ".bak" | sort || true
echo ""

echo "Configuration files:"
find . -name "*.yaml" -o -name "*.json" -o -name "*.env" | grep -v "./trash" | grep -v "./var" | xargs grep -l "dataset" 2>/dev/null | sort || true
echo ""

# Ask for confirmation
read -p "Do you want to proceed with the rename operation? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Operation cancelled."
    exit 0
fi

# Step 2: Rename Go files first
echo ""
echo "=== Step 2: Renaming Go source files ==="

# Rename Go files that have 'dataset' in their name
cd src
for file in $(find . -name "*dataset*.go" -type f); do
    newfile=$(echo "$file" | sed 's/dataset/dataset/g')
    echo "Renaming file: $file -> $newfile"
    git mv "$file" "$newfile" 2>/dev/null || mv "$file" "$newfile"
done
cd ..

# Step 3: Update Go source code
echo ""
echo "=== Step 3: Updating Go source code ==="

find ./src -name "*.go" -type f | while read -r file; do
    if grep -q "dataset" "$file"; then
        rename_in_file "$file"
    fi
done

# Step 4: Update documentation
echo ""
echo "=== Step 4: Updating documentation ==="

find ./docs -name "*.md" -type f | while read -r file; do
    if grep -q "dataset" "$file"; then
        rename_in_file "$file"
    fi
done

# Also update root README and CLAUDE.md
for file in README.md CLAUDE.md; do
    if [ -f "$file" ] && grep -q "dataset" "$file"; then
        rename_in_file "$file"
    fi
done

# Step 5: Update frontend files
echo ""
echo "=== Step 5: Updating frontend files ==="

find ./share/htdocs -type f \( -name "*.html" -o -name "*.js" \) | grep -v ".bak" | while read -r file; do
    if grep -q "dataset" "$file"; then
        rename_in_file "$file"
    fi
done

# Step 6: Update configuration files
echo ""
echo "=== Step 6: Updating configuration files ==="

for file in $(find . -name "*.yaml" -o -name "*.json" -o -name "*.env" | grep -v "./trash" | grep -v "./var" | grep -v ".git"); do
    if [ -f "$file" ] && grep -q "dataset" "$file"; then
        rename_in_file "$file"
    fi
done

# Step 7: Update scripts
echo ""
echo "=== Step 7: Updating shell scripts ==="

find . -name "*.sh" -type f | grep -v "./trash" | grep -v ".dataset_backup" | while read -r file; do
    if grep -q "dataset" "$file"; then
        rename_in_file "$file"
    fi
done

echo ""
echo "=== Rename operation completed! ==="
echo ""
echo "Summary of changes:"
echo "1. All API endpoints changed from /datasets to /datasets"
echo "2. All Go types/functions renamed (Dataset -> Dataset)"
echo "3. All documentation updated"
echo "4. All frontend references updated"
echo "5. Environment variables updated (ENTITYDB_ENABLE_DATASET -> ENTITYDB_ENABLE_DATASET)"
echo ""
echo "Backup files created with .dataset_backup extension"
echo ""
echo "Next steps:"
echo "1. Review the changes: git diff"
echo "2. Run 'make' to ensure the code compiles"
echo "3. Run tests to ensure functionality is intact"
echo "4. Commit the changes when satisfied"
echo ""
echo "To restore backups if needed:"
echo "find . -name '*.dataset_backup' | while read f; do mv \"\$f\" \"\${f%.dataset_backup}\"; done"